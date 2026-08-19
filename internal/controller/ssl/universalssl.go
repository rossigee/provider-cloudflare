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
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/ssl/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/ssl/universalssl"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	errNotUniversalSSL = "managed resource is not a Universal SSL custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errNewClient       = "cannot create new Service"
)

// SetupUniversalSSLController adds a controller that reconciles Universal SSL managed resources.
func SetupUniversalSSLController(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.UniversalSSLKind)

	o := controller.Options{
		RateLimiter: nil, // Use default rate limiter
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.UniversalSSLGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (*cloudflare.API, error) {
				return clients.NewClient(cfg, nil)
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
		For(&v1beta1.UniversalSSL{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
}

// Connect typically produces an ExternalClient by:
// 1. Getting the managed resource's ProviderConfig.
// 2. Getting the credentials specified by the ProviderConfig.
// 3. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.UniversalSSL)
	if !ok {
		return nil, errors.New(errNotUniversalSSL)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	cloudflareClient, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	service := universalssl.NewClient(cloudflareClient)

	return &external{service: service}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *universalssl.CloudflareUniversalSSLClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.UniversalSSL)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotUniversalSSL)
	}
	_, span := tracing.StartSpan(ctx, "universalssl.observe",
		tracing.SpanAttrs("universalssl", cr.GetName(), "observe")...)
	defer span.End()

	// Universal SSL settings always exist for a zone, so we never create them
	// We only observe and update the configuration
	observation, err := c.service.Get(ctx, cr.Spec.ForProvider.Zone)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get Universal SSL settings")
	}

	cr.Status.AtProvider = *observation

	// Universal SSL settings always exist, so we consider the resource to exist
	// Check if the current state matches desired state
	upToDate, err := c.service.IsUpToDate(ctx, cr.Spec.ForProvider, *observation)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to check if Universal SSL is up to date")
	}

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.UniversalSSL)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotUniversalSSL)
	}
	_, span := tracing.StartSpan(ctx, "universalssl.create",
		tracing.SpanAttrs("universalssl", cr.GetName(), "create")...)
	defer span.End()

	// Universal SSL settings always exist for a zone, so we treat "create" as "update"
	cr.Status.SetConditions(xpv1.Creating())

	observation, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to update Universal SSL settings")
	}

	cr.Status.AtProvider = *observation

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.UniversalSSL)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotUniversalSSL)
	}
	_, span := tracing.StartSpan(ctx, "universalssl.update",
		tracing.SpanAttrs("universalssl", cr.GetName(), "update")...)
	defer span.End()

	observation, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update Universal SSL settings")
	}

	cr.Status.AtProvider = *observation

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.UniversalSSL)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotUniversalSSL)
	}
	_, span := tracing.StartSpan(ctx, "universalssl.delete",
		tracing.SpanAttrs("universalssl", cr.GetName(), "delete")...)
	defer span.End()

	cr.Status.SetConditions(xpv1.Deleting())

	// Universal SSL settings cannot be deleted, only disabled
	// We set enabled to false when the resource is being deleted
	params := cr.Spec.ForProvider
	params.Enabled = false

	_, err := c.service.Update(ctx, params)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to disable Universal SSL settings")
	}

	return managed.ExternalDelete{}, nil
}
func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}
