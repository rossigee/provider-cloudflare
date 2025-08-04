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

package route

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateRoute = "cannot create worker route"
	errUpdateRoute = "cannot update worker route"
	errGetRoute    = "cannot get worker route"
	errDeleteRoute = "cannot delete worker route"
	errListRoutes  = "cannot list worker routes"
	errRouteNotFound = "Worker Route not found"
)

// RouteClient provides operations for Worker Routes.
type RouteClient struct {
	client clients.ClientInterface
}

// NewClient creates a new Worker Route client.
func NewClient(client clients.ClientInterface) *RouteClient {
	return &RouteClient{
		client: client,
	}
}

// convertToObservation converts cloudflare-go worker route to Crossplane observation.
func convertToObservation(route cloudflare.WorkerRoute) v1alpha1.RouteObservation {
	// Currently RouteObservation is empty, but ready for future fields
	return v1alpha1.RouteObservation{}
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.RouteParameters) cloudflare.CreateWorkerRouteParams {
	cfParams := cloudflare.CreateWorkerRouteParams{
		Pattern: params.Pattern,
	}

	if params.Script != nil {
		cfParams.Script = *params.Script
	}

	return cfParams
}

// Create creates a new Worker Route.
func (c *RouteClient) Create(ctx context.Context, zoneID string, params v1alpha1.RouteParameters) (*v1alpha1.RouteObservation, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	createParams := convertToCloudflareParams(params)
	
	resp, err := c.client.CreateWorkerRoute(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateRoute)
	}

	obs := convertToObservation(resp.WorkerRoute)
	return &obs, nil
}

// Get retrieves a Worker Route.
func (c *RouteClient) Get(ctx context.Context, zoneID, routeID string) (*v1alpha1.RouteObservation, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	response, err := c.client.ListWorkerRoutes(ctx, rc, cloudflare.ListWorkerRoutesParams{})
	if err != nil {
		return nil, errors.Wrap(err, errGetRoute)
	}

	for _, route := range response.Routes {
		if route.ID == routeID {
			obs := convertToObservation(route)
			return &obs, nil
		}
	}

	return nil, clients.NewNotFoundError("worker route not found")
}

// Update updates an existing Worker Route.
func (c *RouteClient) Update(ctx context.Context, zoneID, routeID string, params v1alpha1.RouteParameters) (*v1alpha1.RouteObservation, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	updateParams := cloudflare.UpdateWorkerRouteParams{
		ID:      routeID,
		Pattern: params.Pattern,
	}

	if params.Script != nil {
		updateParams.Script = *params.Script
	}

	_, err := c.client.UpdateWorkerRoute(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateRoute)
	}

	// Get the updated route to return the observation
	return c.Get(ctx, zoneID, routeID)
}

// Delete removes a Worker Route.
func (c *RouteClient) Delete(ctx context.Context, zoneID, routeID string) error {
	rc := cloudflare.ZoneIdentifier(zoneID)

	_, err := c.client.DeleteWorkerRoute(ctx, rc, routeID)
	if err != nil && !IsRouteNotFound(err) {
		return errors.Wrap(err, errDeleteRoute)
	}

	return nil
}

// IsUpToDate checks if the Worker Route is up to date.
func (c *RouteClient) IsUpToDate(ctx context.Context, zoneID, routeID string, params v1alpha1.RouteParameters) (bool, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	response, err := c.client.ListWorkerRoutes(ctx, rc, cloudflare.ListWorkerRoutesParams{})
	if err != nil {
		return false, errors.Wrap(err, errListRoutes)
	}

	for _, route := range response.Routes {
		if route.ID == routeID {
			// Check pattern
			if params.Pattern != route.Pattern {
				return false, nil
			}

			// Check script
			if params.Script != nil && *params.Script != route.ScriptName {
				return false, nil
			}

			return true, nil
		}
	}

	return false, clients.NewNotFoundError("worker route not found")
}

// IsRouteNotFound returns true if the error indicates the route was not found
func IsRouteNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errRouteNotFound ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}