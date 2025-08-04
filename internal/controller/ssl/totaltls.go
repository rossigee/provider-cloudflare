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

package ssl

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	rtv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/ssl/totaltls"
)

const (
	errNotTotalTLS      = "managed resource is not a Total TLS custom resource"
	errTrackPCUsageTLS  = "cannot track ProviderConfig usage"
	errGetPCTLS         = "cannot get ProviderConfig"
	errGetCredsTLS      = "cannot get credentials"
	errNewClientTLS     = "cannot create new Service"
)

// SetupTotalTLSController adds a controller that reconciles Total TLS managed resources.
func SetupTotalTLSController(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.TotalTLSKind)

	o := controller.Options{
		RateLimiter: rl,
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.TotalTLSGroupVersionKind),
		managed.WithExternalConnecter(&totalTLSConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (*cloudflare.API, error) {
				return clients.NewClient(cfg, nil)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.TotalTLS{}).
		Complete(r)
}

// A totalTLSConnector is expected to produce an ExternalClient when its Connect method
// is called.
type totalTLSConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
}

// Connect typically produces an ExternalClient by:
// 1. Getting the managed resource's ProviderConfig.
// 2. Getting the credentials specified by the ProviderConfig.
// 3. Using the credentials to form a client.
func (c *totalTLSConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.TotalTLS)
	if !ok {
		return nil, errors.New(errNotTotalTLS)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCredsTLS)
	}

	cloudflareClient, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, errors.Wrap(err, errNewClientTLS)
	}

	service := totaltls.NewClient(cloudflareClient)

	return &totalTLSExternal{service: service}, nil
}

// An totalTLSExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type totalTLSExternal struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *totaltls.CloudflareTotalTLSClient
}

func (c *totalTLSExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.TotalTLS)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTotalTLS)
	}

	// Total TLS settings always exist for a zone, so we never create them
	// We only observe and update the configuration
	observation, err := c.service.Get(ctx, cr.Spec.ForProvider.Zone)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get Total TLS settings")
	}

	cr.Status.AtProvider = *observation

	// Total TLS settings always exist, so we consider the resource to exist
	// Check if the current state matches desired state
	upToDate, err := c.service.IsUpToDate(ctx, cr.Spec.ForProvider, *observation)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to check if Total TLS is up to date")
	}

	cr.Status.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *totalTLSExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.TotalTLS)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotTotalTLS)
	}

	// Total TLS settings always exist for a zone, so we treat "create" as "update"
	cr.Status.SetConditions(rtv1.Creating())

	observation, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to update Total TLS settings")
	}

	cr.Status.AtProvider = *observation

	return managed.ExternalCreation{}, nil
}

func (c *totalTLSExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.TotalTLS)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotTotalTLS)
	}

	observation, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update Total TLS settings")
	}

	cr.Status.AtProvider = *observation

	return managed.ExternalUpdate{}, nil
}

func (c *totalTLSExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.TotalTLS)
	if !ok {
		return errors.New(errNotTotalTLS)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	// Total TLS settings cannot be deleted, only disabled
	// We set enabled to false when the resource is being deleted
	params := cr.Spec.ForProvider
	enabled := false
	params.Enabled = &enabled

	_, err := c.service.Update(ctx, params)
	if err != nil {
		return errors.Wrap(err, "failed to disable Total TLS settings")
	}

	return nil
}