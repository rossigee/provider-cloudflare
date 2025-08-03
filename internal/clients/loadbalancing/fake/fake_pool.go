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

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
)

// MockPoolClient acts as a testable representation of the Cloudflare Load Balancer Pool API.
type MockPoolClient struct {
	MockCreatePool func(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockGetPool    func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockUpdatePool func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockDeletePool func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error
}

// CreatePool mocks the CreatePool method of the Cloudflare API.
func (m MockPoolClient) CreatePool(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockCreatePool != nil {
		return m.MockCreatePool(ctx, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// GetPool mocks the GetPool method of the Cloudflare API.
func (m MockPoolClient) GetPool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockGetPool != nil {
		return m.MockGetPool(ctx, poolID, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// UpdatePool mocks the UpdatePool method of the Cloudflare API.
func (m MockPoolClient) UpdatePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockUpdatePool != nil {
		return m.MockUpdatePool(ctx, poolID, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// DeletePool mocks the DeletePool method of the Cloudflare API.
func (m MockPoolClient) DeletePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error {
	if m.MockDeletePool != nil {
		return m.MockDeletePool(ctx, poolID, params)
	}
	return nil
}