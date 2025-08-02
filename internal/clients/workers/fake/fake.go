/*
Copyright 2021 The Crossplane Authors.

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

// MockClient is a fake implementation of the Workers client for testing
type MockClient struct {
	MockWorkerRoute       func(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error)
	MockCreateWorkerRoute func(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error)
	MockUpdateWorkerRoute func(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error
	MockDeleteWorkerRoute func(ctx context.Context, zoneID, routeID string) error
}

// WorkerRoute calls the MockWorkerRoute function
func (m *MockClient) WorkerRoute(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
	if m.MockWorkerRoute != nil {
		return m.MockWorkerRoute(ctx, zoneID, routeID)
	}
	return cloudflare.WorkerRoute{}, nil
}

// CreateWorkerRoute calls the MockCreateWorkerRoute function
func (m *MockClient) CreateWorkerRoute(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error) {
	if m.MockCreateWorkerRoute != nil {
		return m.MockCreateWorkerRoute(ctx, zoneID, params)
	}
	return cloudflare.WorkerRoute{}, nil
}

// UpdateWorkerRoute calls the MockUpdateWorkerRoute function
func (m *MockClient) UpdateWorkerRoute(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error {
	if m.MockUpdateWorkerRoute != nil {
		return m.MockUpdateWorkerRoute(ctx, zoneID, routeID, params)
	}
	return nil
}

// DeleteWorkerRoute calls the MockDeleteWorkerRoute function
func (m *MockClient) DeleteWorkerRoute(ctx context.Context, zoneID, routeID string) error {
	if m.MockDeleteWorkerRoute != nil {
		return m.MockDeleteWorkerRoute(ctx, zoneID, routeID)
	}
	return nil
}