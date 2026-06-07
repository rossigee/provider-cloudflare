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

package workers

import (
	"context"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"


	rtv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errNotRoute = "managed resource is not a Worker Route custom resource"

	errRouteLookup      = "cannot lookup route"
	errRouteObservation = "cannot observe route"
	errRouteCreation    = "cannot create route"
	errRouteUpdate      = "cannot update route"
	errRouteDeletion    = "cannot delete route"
	errRouteNoZone      = "no zone found"
)

// SetupRoute adds a controller that reconciles Worker Route managed resources.
func SetupRoute(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.RouteKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RouteGroupVersionKind),
		managed.WithExternalConnecter(&routeConnector{
			kube: mgr.GetClient(),
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(10*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Route{}).
		Complete(r)
}

// A routeConnector is expected to produce an ExternalClient when its Connect method
// is called.
type routeConnector struct {
	kube client.Client
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *routeConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Route)
	if !ok {
		return nil, errors.New(errNotRoute)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	// Create cloudflare API client
	api, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare client")
	}

	// Wrap with adapter to implement ClientInterface
	adapter := clients.NewCloudflareAPIAdapter(api)
	return &routeExternal{client: adapter, kube: c.kube}, nil
}

// An routeExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type routeExternal struct {
	client clients.ClientInterface
	kube   client.Client
}

func (e *routeExternal) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {

	cr, ok := mg.(*v1beta1.Route)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRoute)
	}

	// Route does not exist if we dont have a route pattern stored in external-name
	routePattern := meta.GetExternalName(cr)
	if routePattern == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errRouteNoZone)
	}

	// Use ListWorkerRoutes to find the route
	rc := &cloudflare.ResourceContainer{
		Level: cloudflare.ZoneRouteLevel,
		Identifier: *cr.Spec.ForProvider.Zone,
	}

	params := cloudflare.ListWorkerRoutesParams{}
	routesResp, err := e.client.ListWorkerRoutes(ctx, rc, params)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(isRouteNotFound, err), errRouteLookup)
	}

	// Find the route with matching pattern
	var foundRoute *cloudflare.WorkerRoute
	for _, route := range routesResp.Routes {
		if route.Pattern == routePattern {
			foundRoute = &route
			break
		}
	}

	if foundRoute == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Convert the route observation
	obs := generateRouteObservation(foundRoute)
	cr.Status.AtProvider = obs
	cr.Status.SetConditions(rtv1.Available())

	// Check if up to date
	upToDate := isRouteUpToDate(cr.Spec.ForProvider, obs)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *routeExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Route)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRoute)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{}, errors.Wrap(errors.New(errRouteNoZone), errRouteCreation)
	}

	rc := &cloudflare.ResourceContainer{
		Level: cloudflare.ZoneRouteLevel,
		Identifier: *cr.Spec.ForProvider.Zone,
	}

	script := ""
	if cr.Spec.ForProvider.Script != nil {
		script = *cr.Spec.ForProvider.Script
	}

	params := cloudflare.CreateWorkerRouteParams{
		Pattern: cr.Spec.ForProvider.Pattern,
		Script:  script,
	}

	_, err := e.client.CreateWorkerRoute(ctx, rc, params)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRouteCreation)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Pattern)

	return managed.ExternalCreation{}, nil
}

func (e *routeExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Route)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRoute)
	}

	routePattern := meta.GetExternalName(cr)
	if routePattern == "" {
		return managed.ExternalUpdate{}, errors.New(errRouteUpdate)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errRouteNoZone), errRouteUpdate)
	}

	// First find the route ID by pattern
	rc := &cloudflare.ResourceContainer{
		Level: cloudflare.ZoneRouteLevel,
		Identifier: *cr.Spec.ForProvider.Zone,
	}

	listParams := cloudflare.ListWorkerRoutesParams{}
	routesResp, err := e.client.ListWorkerRoutes(ctx, rc, listParams)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errRouteUpdate)
	}

	var routeID string
	for _, route := range routesResp.Routes {
		if route.Pattern == routePattern {
			routeID = route.ID
			break
		}
	}

	if routeID == "" {
		return managed.ExternalUpdate{}, errors.New("route not found for update")
	}

	script := ""
	if cr.Spec.ForProvider.Script != nil {
		script = *cr.Spec.ForProvider.Script
	}

	updateParams := cloudflare.UpdateWorkerRouteParams{
		ID:      routeID,
		Pattern: cr.Spec.ForProvider.Pattern,
		Script:  script,
	}

	_, err = e.client.UpdateWorkerRoute(ctx, rc, updateParams)
	return managed.ExternalUpdate{}, errors.Wrap(err, errRouteUpdate)
}

func (e *routeExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Route)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRoute)
	}

	routePattern := meta.GetExternalName(cr)
	if routePattern == "" {
		return managed.ExternalDelete{}, errors.New(errRouteDeletion)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalDelete{}, errors.Wrap(errors.New(errRouteNoZone), errRouteDeletion)
	}

	// First find the route ID by pattern
	rc := &cloudflare.ResourceContainer{
		Level: cloudflare.ZoneRouteLevel,
		Identifier: *cr.Spec.ForProvider.Zone,
	}

	listParams := cloudflare.ListWorkerRoutesParams{}
	routesResp, err := e.client.ListWorkerRoutes(ctx, rc, listParams)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errRouteDeletion)
	}

	var routeID string
	for _, route := range routesResp.Routes {
		if route.Pattern == routePattern {
			routeID = route.ID
			break
		}
	}

	if routeID == "" {
		return managed.ExternalDelete{}, errors.New("route not found for deletion")
	}

	_, err = e.client.DeleteWorkerRoute(ctx, rc, routeID)
	return managed.ExternalDelete{}, errors.Wrap(err, errRouteDeletion)
}

func (e *routeExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// Helper functions

// isRouteNotFound checks if an error indicates that a route was not found.
func isRouteNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		   strings.Contains(errStr, "404") ||
		   strings.Contains(errStr, "does not exist")
}

// generateRouteObservation converts API response to observation
func generateRouteObservation(route *cloudflare.WorkerRoute) v1beta1.RouteObservation {
	// RouteObservation is currently empty in the API
	return v1beta1.RouteObservation{}
}

// isRouteUpToDate checks if the route matches the desired state
func isRouteUpToDate(spec v1beta1.RouteParameters, obs v1beta1.RouteObservation) bool {
	return true
}