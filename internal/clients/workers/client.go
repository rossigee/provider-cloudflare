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
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errRouteNotFound = "Worker Route not found"
)

// Client is a Cloudflare Workers API client
type Client interface {
	WorkerRoute(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error)
	CreateWorkerRoute(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error)
	UpdateWorkerRoute(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error
	DeleteWorkerRoute(ctx context.Context, zoneID, routeID string) error
}

type client struct {
	cf *cloudflare.API
}

// NewClient returns a new Workers client
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &client{cf: cf}, nil
}

// WorkerRoute retrieves a Worker Route
func (c *client) WorkerRoute(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
	// Worker Routes use zone-level API, but need proper ResourceContainer
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	response, err := c.cf.ListWorkerRoutes(ctx, rc, cloudflare.ListWorkerRoutesParams{})
	if err != nil {
		return cloudflare.WorkerRoute{}, err
	}

	for _, route := range response.Routes {
		if route.ID == routeID {
			return route, nil
		}
	}

	return cloudflare.WorkerRoute{}, errors.New(errRouteNotFound)
}

// CreateWorkerRoute creates a new Worker Route
func (c *client) CreateWorkerRoute(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error) {
	// Worker Routes use zone-level API, but need proper ResourceContainer
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	createParams := cloudflare.CreateWorkerRouteParams{
		Pattern: params.Pattern,
	}

	if params.Script != nil {
		createParams.Script = *params.Script
	}

	resp, err := c.cf.CreateWorkerRoute(ctx, rc, createParams)
	if err != nil {
		return cloudflare.WorkerRoute{}, err
	}

	return resp.WorkerRoute, nil
}

// UpdateWorkerRoute updates an existing Worker Route
func (c *client) UpdateWorkerRoute(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error {
	// Worker Routes use zone-level API, but need proper ResourceContainer
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	updateParams := cloudflare.UpdateWorkerRouteParams{
		ID:      routeID,
		Pattern: params.Pattern,
	}

	if params.Script != nil {
		updateParams.Script = *params.Script
	}

	_, err := c.cf.UpdateWorkerRoute(ctx, rc, updateParams)
	return err
}

// DeleteWorkerRoute deletes a Worker Route
func (c *client) DeleteWorkerRoute(ctx context.Context, zoneID, routeID string) error {
	// Worker Routes use zone-level API, but need proper ResourceContainer
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	_, err := c.cf.DeleteWorkerRoute(ctx, rc, routeID)
	return err
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

// GenerateObservation creates observation data from a Worker Route
func GenerateObservation(route cloudflare.WorkerRoute) v1alpha1.RouteObservation {
	return v1alpha1.RouteObservation{}
}

// LateInitialize fills in any missing fields in the spec from the observed route
func LateInitialize(spec *v1alpha1.RouteParameters, route cloudflare.WorkerRoute) bool {
	// No late initialization needed for Worker routes currently
	return false
}

// UpToDate checks if the spec is up to date with the observed route
func UpToDate(spec *v1alpha1.RouteParameters, route cloudflare.WorkerRoute) bool {
	// Check pattern
	if spec.Pattern != route.Pattern {
		return false
	}

	// Check script
	if spec.Script != nil && *spec.Script != route.ScriptName {
		return false
	}

	return true
}