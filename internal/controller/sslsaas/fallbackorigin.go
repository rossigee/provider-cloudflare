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

package sslsaas

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	rtv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	fallbackorigin "github.com/rossigee/provider-cloudflare/internal/clients/sslsaas/fallbackorigin"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotFallbackOrigin = "managed resource is not a Fallback Origin custom resource"

	errFallbackOriginLookup   = "cannot lookup fallback origin"
	errFallbackOriginCreation = "cannot create fallback origin"
	errFallbackOriginUpdate   = "cannot update fallback origin"
	errFallbackOriginDeletion = "cannot delete fallback origin"
	errFallbackOriginNoZone   = "cannot create fallback origin no zone found"
)


// SetupFallbackOrigin adds a controller that reconciles FallbackOrigin managed resources.
func SetupFallbackOrigin(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1alpha1.FallbackOriginGroupKind)

	o := controller.Options{
		RateLimiter: nil, // Use default rate limiter
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.FallbackOriginGroupVersionKind),
		managed.WithExternalConnecter(&fallbackOriginConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (fallbackorigin.Client, error) {
				return fallbackorigin.NewClient(cfg, hc)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(5*time.Minute),
		// Do not initialize external-name field.
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.FallbackOrigin{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type fallbackOriginConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (fallbackorigin.Client, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *fallbackOriginConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.FallbackOrigin)
	if !ok {
		return nil, errors.New(errNotFallbackOrigin)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	client, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, err
	}

	return &fallbackOriginExternal{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type fallbackOriginExternal struct {
	client fallbackorigin.Client
}

func (e *fallbackOriginExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.FallbackOrigin)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFallbackOrigin)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errFallbackOriginNoZone)
	}

	// FallbackOrigin uses the zone ID as the external name since it's zone-scoped
	zoneID := *cr.Spec.ForProvider.Zone
	meta.SetExternalName(cr, zoneID)

	origin, err := e.client.FallbackOrigin(ctx, zoneID)
	if err != nil {
		if fallbackorigin.IsFallbackOriginNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errFallbackOriginLookup)
	}

	cr.Status.AtProvider = fallbackorigin.GenerateObservation(origin)
	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: fallbackorigin.UpToDate(&cr.Spec.ForProvider, origin),
	}, nil
}

func (e *fallbackOriginExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.FallbackOrigin)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFallbackOrigin)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{}, errors.New(errFallbackOriginNoZone)
	}

	cr.SetConditions(rtv1.Creating())

	zoneID := *cr.Spec.ForProvider.Zone
	origin := fallbackorigin.ParametersToFallbackOrigin(cr.Spec.ForProvider)

	result, err := e.client.UpdateFallbackOrigin(ctx, zoneID, origin)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFallbackOriginCreation)
	}

	cr.Status.AtProvider = fallbackorigin.GenerateObservation(result.Result)
	meta.SetExternalName(cr, zoneID)

	return managed.ExternalCreation{}, nil
}

func (e *fallbackOriginExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.FallbackOrigin)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFallbackOrigin)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.New(errFallbackOriginNoZone)
	}

	zoneID := *cr.Spec.ForProvider.Zone
	origin := fallbackorigin.ParametersToFallbackOrigin(cr.Spec.ForProvider)

	_, err := e.client.UpdateFallbackOrigin(ctx, zoneID, origin)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errFallbackOriginUpdate)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *fallbackOriginExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.FallbackOrigin)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotFallbackOrigin)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalDelete{}, errors.New(errFallbackOriginNoZone)
	}

	zoneID := *cr.Spec.ForProvider.Zone

	// Delete by calling the delete API
	err := e.client.DeleteFallbackOrigin(ctx, zoneID)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errFallbackOriginDeletion)
	}
	return managed.ExternalDelete{}, nil
}

func (e *fallbackOriginExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}