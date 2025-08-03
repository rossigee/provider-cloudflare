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

package fake

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
)

// MockClient provides a mock implementation of the Worker Route client interface
type MockClient struct {
	MockCreateWorkerRoute func(ctx context.Context, zoneID string, spec *v1alpha1.RouteParameters) (cloudflare.WorkerRouteResponse, error)
	MockUpdateWorkerRoute func(ctx context.Context, zoneID string, routeID string, spec *v1alpha1.RouteParameters) error
	MockWorkerRoute       func(ctx context.Context, zoneID string, routeID string) (cloudflare.WorkerRoute, error)
	MockDeleteWorkerRoute func(ctx context.Context, zoneID string, routeID string) error
}

// CreateWorkerRoute mocks creating a worker route
func (m *MockClient) CreateWorkerRoute(ctx context.Context, zoneID string, spec *v1alpha1.RouteParameters) (cloudflare.WorkerRouteResponse, error) {
	return m.MockCreateWorkerRoute(ctx, zoneID, spec)
}

// UpdateWorkerRoute mocks updating a worker route
func (m *MockClient) UpdateWorkerRoute(ctx context.Context, zoneID string, routeID string, spec *v1alpha1.RouteParameters) error {
	return m.MockUpdateWorkerRoute(ctx, zoneID, routeID, spec)
}

// WorkerRoute mocks getting a worker route
func (m *MockClient) WorkerRoute(ctx context.Context, zoneID string, routeID string) (cloudflare.WorkerRoute, error) {
	return m.MockWorkerRoute(ctx, zoneID, routeID)
}

// DeleteWorkerRoute mocks deleting a worker route
func (m *MockClient) DeleteWorkerRoute(ctx context.Context, zoneID string, routeID string) error {
	return m.MockDeleteWorkerRoute(ctx, zoneID, routeID)
}