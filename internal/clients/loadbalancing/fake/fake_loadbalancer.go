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

// MockLoadBalancerClient acts as a testable representation of the Cloudflare Load Balancer API.
type MockLoadBalancerClient struct {
	MockCreateLoadBalancer func(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockGetLoadBalancer    func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockUpdateLoadBalancer func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockDeleteLoadBalancer func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error
}

// CreateLoadBalancer mocks the CreateLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockCreateLoadBalancer != nil {
		return m.MockCreateLoadBalancer(ctx, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// GetLoadBalancer mocks the GetLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockGetLoadBalancer != nil {
		return m.MockGetLoadBalancer(ctx, lbID, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// UpdateLoadBalancer mocks the UpdateLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockUpdateLoadBalancer != nil {
		return m.MockUpdateLoadBalancer(ctx, lbID, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// DeleteLoadBalancer mocks the DeleteLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error {
	if m.MockDeleteLoadBalancer != nil {
		return m.MockDeleteLoadBalancer(ctx, lbID, params)
	}
	return nil
}