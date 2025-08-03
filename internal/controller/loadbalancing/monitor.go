/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package loadbalancing

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
	apisv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/loadbalancing"
)

const (
	errNotMonitor          = "managed resource is not a LoadBalancerMonitor custom resource"
	errTrackMonitorPCUsage = "cannot track ProviderConfig usage"
	errGetMonitorPC        = "cannot get ProviderConfig"
	errGetMonitorCreds     = "cannot get credentials"
	errNewMonitorClient    = "cannot create new Service"
)

// SetupMonitor adds a controller that reconciles LoadBalancerMonitor managed resources.
func SetupMonitor(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.LoadBalancerMonitorGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.LoadBalancerMonitorGroupVersionKind),
		managed.WithExternalConnecter(&monitorConnector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: loadbalancing.NewMonitorClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.LoadBalancerMonitor{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A monitorConnector is expected to produce an ExternalClient when its Connect method
// is called.
type monitorConnector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(cfg clients.Config, httpClient *clients.HTTPClient) (loadbalancing.MonitorClient, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *monitorConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerMonitor)
	if !ok {
		return nil, errors.New(errNotMonitor)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackMonitorPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetMonitorPC)
	}

	cd := pc.Spec.Credentials
	data, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetMonitorCreds)
	}

	cfg := clients.NewConfig(data)

	svc, err := c.newServiceFn(cfg, clients.NewHTTPClient())
	if err != nil {
		return nil, errors.Wrap(err, errNewMonitorClient)
	}

	return &monitorExternal{service: svc}, nil
}

// A monitorExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type monitorExternal struct {
	service loadbalancing.MonitorClient
}

func (c *monitorExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerMonitor)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMonitor)
	}

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	monitor, err := c.service.GetMonitor(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		if loadbalancing.IsMonitorNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get monitor")
	}

	cr.Status.AtProvider = loadbalancing.GenerateMonitorObservation(monitor)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        loadbalancing.IsMonitorUpToDate(&cr.Spec.ForProvider, monitor),
		ResourceLateInitialized: c.lateInitialize(&cr.Spec.ForProvider, monitor),
		ConnectionDetails:       managed.ConnectionDetails{},
	}, nil
}

func (c *monitorExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerMonitor)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMonitor)
	}

	monitor, err := c.service.CreateMonitor(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create monitor")
	}

	cr.Status.AtProvider = loadbalancing.GenerateMonitorObservation(monitor)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *monitorExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerMonitor)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMonitor)
	}

	_, err := c.service.UpdateMonitor(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update monitor")
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *monitorExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.LoadBalancerMonitor)
	if !ok {
		return errors.New(errNotMonitor)
	}

	err := c.service.DeleteMonitor(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil && !loadbalancing.IsMonitorNotFound(err) {
		return errors.Wrap(err, "failed to delete monitor")
	}

	return nil
}

func (c *monitorExternal) lateInitialize(spec *v1alpha1.LoadBalancerMonitorParameters, monitor *cloudflare.LoadBalancerMonitor) bool {
	li := false

	if spec.Description == nil && monitor.Description != "" {
		spec.Description = &monitor.Description
		li = true
	}

	if spec.Method == nil && monitor.Method != "" {
		spec.Method = &monitor.Method
		li = true
	}

	if spec.Path == nil && monitor.Path != "" {
		spec.Path = &monitor.Path
		li = true
	}

	if spec.Timeout == nil && monitor.Timeout != 0 {
		spec.Timeout = &monitor.Timeout
		li = true
	}

	if spec.Retries == nil && monitor.Retries != 0 {
		spec.Retries = &monitor.Retries
		li = true
	}

	if spec.Interval == nil && monitor.Interval != 0 {
		spec.Interval = &monitor.Interval
		li = true
	}

	if spec.ConsecutiveUp == nil && monitor.ConsecutiveUp != 0 {
		spec.ConsecutiveUp = &monitor.ConsecutiveUp
		li = true
	}

	if spec.ConsecutiveDown == nil && monitor.ConsecutiveDown != 0 {
		spec.ConsecutiveDown = &monitor.ConsecutiveDown
		li = true
	}

	if spec.Port == nil && monitor.Port != 0 {
		port := int(monitor.Port)
		spec.Port = &port
		li = true
	}

	if spec.ExpectedBody == nil && monitor.ExpectedBody != "" {
		spec.ExpectedBody = &monitor.ExpectedBody
		li = true
	}

	if spec.ExpectedCodes == nil && monitor.ExpectedCodes != "" {
		spec.ExpectedCodes = &monitor.ExpectedCodes
		li = true
	}

	if spec.FollowRedirects == nil {
		spec.FollowRedirects = &monitor.FollowRedirects
		li = true
	}

	if spec.AllowInsecure == nil {
		spec.AllowInsecure = &monitor.AllowInsecure
		li = true
	}

	if spec.ProbeZone == nil && monitor.ProbeZone != "" {
		spec.ProbeZone = &monitor.ProbeZone
		li = true
	}

	return li
}