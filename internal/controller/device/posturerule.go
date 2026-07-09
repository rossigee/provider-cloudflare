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

package device

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	devicev1beta1 "github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/device/posturerule"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"
)

const (
	errNotDevicePostureRule = "managed resource is not a DevicePostureRule custom resource"
	errGetCreds             = "cannot get credentials"
	errNewDeviceRuleClient  = "cannot create new DevicePostureRule client"
)

// SetupDevicePostureRule adds a controller that reconciles DevicePostureRule managed resources.
func SetupDevicePostureRule(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(devicev1beta1.DevicePostureRuleKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(devicev1beta1.DevicePostureRuleGroupVersionKind),
		managed.WithExternalConnector(&devicePostureRuleConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) *posturerule.CloudflareDevicePostureRuleClient {
				return posturerule.NewClientFromAPI(api)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&devicev1beta1.DevicePostureRule{}).
		Complete(r)
}

// A devicePostureRuleConnector is expected to produce an ExternalClient when its Connect method
// is called.
type devicePostureRuleConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) *posturerule.CloudflareDevicePostureRuleClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *devicePostureRuleConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*devicev1beta1.DevicePostureRule)
	if !ok {
		return nil, errors.New(errNotDevicePostureRule)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewDeviceRuleClient)
	}

	// Create the device posture rule client
	return &devicePostureRuleExternal{service: c.newServiceFn(client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type devicePostureRuleExternal struct {
	service *posturerule.CloudflareDevicePostureRuleClient
}

func (c *devicePostureRuleExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*devicev1beta1.DevicePostureRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDevicePostureRule)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(clients.IsNotFound, err), "cannot get external resource")
	}

	cr.Status.AtProvider = *obs

	cr.Status.SetConditions(xpv1.Available())

	upToDate, err := c.service.IsUpToDate(ctx, cr.Spec.ForProvider, *obs)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot determine if resource is up to date")
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *devicePostureRuleExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*devicev1beta1.DevicePostureRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDevicePostureRule)
	}

	cr.Status.SetConditions(xpv1.Creating())

	obs, err := c.service.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create external resource")
	}

	cr.Status.AtProvider = *obs
	meta.SetExternalName(cr, obs.ID)

	return managed.ExternalCreation{}, nil
}

func (c *devicePostureRuleExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*devicev1beta1.DevicePostureRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDevicePostureRule)
	}

	obs, err := c.service.Update(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *devicePostureRuleExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*devicev1beta1.DevicePostureRule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotDevicePostureRule)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	return managed.ExternalDelete{}, c.service.Delete(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr))
}

func (c *devicePostureRuleExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}
