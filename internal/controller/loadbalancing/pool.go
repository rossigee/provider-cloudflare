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
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	errNotPool          = "managed resource is not a LoadBalancerPool custom resource"
	errTrackPoolPCUsage = "cannot track ProviderConfig usage"
	errGetPoolPC        = "cannot get ProviderConfig"
	errGetPoolCreds     = "cannot get credentials"
	errNewPoolClient    = "cannot create new Service"
)

// SetupPool adds a controller that reconciles LoadBalancerPool managed resources.
func SetupPool(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.LoadBalancerPoolGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.LoadBalancerPoolGroupVersionKind),
		managed.WithExternalConnecter(&poolConnector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1alpha1.ProviderConfigUsage{}),
			newServiceFn: loadbalancing.NewPoolClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.LoadBalancerPool{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

// A poolConnector is expected to produce an ExternalClient when its Connect method
// is called.
type poolConnector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(cfg clients.Config, httpClient *http.Client) (loadbalancing.PoolClient, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *poolConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerPool)
	if !ok {
		return nil, errors.New(errNotPool)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPoolPCUsage)
	}

	pc := &apisv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPoolPC)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetPoolCreds)
	}

	svc, err := c.newServiceFn(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewPoolClient)
	}

	return &poolExternal{service: svc, kube: c.kube}, nil
}

// A poolExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type poolExternal struct {
	service loadbalancing.PoolClient
	kube    client.Client
}

func (c *poolExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerPool)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPool)
	}

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalObservation{}, err
	}

	pool, err := c.service.GetPool(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		if loadbalancing.IsPoolNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get load balancer pool from Cloudflare API")
	}

	cr.Status.AtProvider = loadbalancing.GeneratePoolObservation(pool)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        loadbalancing.IsPoolUpToDate(&cr.Spec.ForProvider, pool),
		ResourceLateInitialized: c.lateInitialize(&cr.Spec.ForProvider, pool),
		ConnectionDetails:       managed.ConnectionDetails{},
	}, nil
}

func (c *poolExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerPool)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPool)
	}

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalCreation{}, err
	}

	pool, err := c.service.CreatePool(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create load balancer pool in Cloudflare API")
	}

	cr.Status.AtProvider = loadbalancing.GeneratePoolObservation(pool)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *poolExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerPool)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPool)
	}

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	_, err := c.service.UpdatePool(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update load balancer pool in Cloudflare API")
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *poolExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.LoadBalancerPool)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPool)
	}

	err := c.service.DeletePool(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil && !loadbalancing.IsPoolNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete load balancer pool from Cloudflare API")
	}

	return managed.ExternalDelete{}, nil
}

func (c *poolExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

func (c *poolExternal) lateInitialize(spec *v1alpha1.LoadBalancerPoolParameters, pool *cloudflare.LoadBalancerPool) bool {
	li := false

	if spec.Description == nil && pool.Description != "" {
		spec.Description = &pool.Description
		li = true
	}

	if spec.Enabled == nil {
		spec.Enabled = &pool.Enabled
		li = true
	}

	if spec.MinimumOrigins == nil && pool.MinimumOrigins != nil {
		spec.MinimumOrigins = pool.MinimumOrigins
		li = true
	}

	if spec.NotificationEmail == nil && pool.NotificationEmail != "" {
		spec.NotificationEmail = &pool.NotificationEmail
		li = true
	}

	return li
}

func (c *poolExternal) resolveReferences(ctx context.Context, cr *v1alpha1.LoadBalancerPool) error {
	// Resolve MonitorRef
	if cr.Spec.ForProvider.MonitorRef != nil {
		r := cr.Spec.ForProvider.MonitorRef
		monitor := &v1alpha1.LoadBalancerMonitor{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: r.Name}, monitor); err != nil {
			return errors.Wrap(err, "cannot get referenced monitor")
		}
		if monitor.Status.AtProvider.ID == "" {
			return errors.New("referenced monitor does not have an ID yet")
		}
		cr.Spec.ForProvider.Monitor = &monitor.Status.AtProvider.ID
	}

	// Resolve MonitorSelector
	if cr.Spec.ForProvider.MonitorSelector != nil {
		monitors := &v1alpha1.LoadBalancerMonitorList{}
		if err := c.kube.List(ctx, monitors, client.MatchingLabels(cr.Spec.ForProvider.MonitorSelector.MatchLabels)); err != nil {
			return errors.Wrap(err, "cannot list monitors for monitor selector")
		}
		if len(monitors.Items) > 0 && monitors.Items[0].Status.AtProvider.ID != "" {
			cr.Spec.ForProvider.Monitor = &monitors.Items[0].Status.AtProvider.ID
		}
	}

	return nil
}