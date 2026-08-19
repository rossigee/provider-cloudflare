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
	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
)

// MockLoadBalancerClient acts as a testable representation of the Cloudflare Load Balancer API.
type MockLoadBalancerClient struct {
	MockCreateLoadBalancer func(ctx context.Context, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockGetLoadBalancer    func(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockUpdateLoadBalancer func(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	MockDeleteLoadBalancer func(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) error
}

// CreateLoadBalancer mocks the CreateLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) CreateLoadBalancer(ctx context.Context, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockCreateLoadBalancer != nil {
		return m.MockCreateLoadBalancer(ctx, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// GetLoadBalancer mocks the GetLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) GetLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockGetLoadBalancer != nil {
		return m.MockGetLoadBalancer(ctx, lbID, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// UpdateLoadBalancer mocks the UpdateLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	if m.MockUpdateLoadBalancer != nil {
		return m.MockUpdateLoadBalancer(ctx, lbID, params)
	}
	return &cloudflare.LoadBalancer{}, nil
}

// DeleteLoadBalancer mocks the DeleteLoadBalancer method of the Cloudflare API.
func (m MockLoadBalancerClient) DeleteLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) error {
	if m.MockDeleteLoadBalancer != nil {
		return m.MockDeleteLoadBalancer(ctx, lbID, params)
	}
	return nil
}

// MockMonitorClient acts as a testable representation of the Cloudflare Load Balancer Monitor API.
type MockMonitorClient struct {
	MockCreateMonitor func(ctx context.Context, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockGetMonitor    func(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockUpdateMonitor func(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockDeleteMonitor func(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) error
}

// CreateMonitor mocks the CreateMonitor method of the Cloudflare API.
func (m MockMonitorClient) CreateMonitor(ctx context.Context, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockCreateMonitor != nil {
		return m.MockCreateMonitor(ctx, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// GetMonitor mocks the GetMonitor method of the Cloudflare API.
func (m MockMonitorClient) GetMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockGetMonitor != nil {
		return m.MockGetMonitor(ctx, monitorID, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// UpdateMonitor mocks the UpdateMonitor method of the Cloudflare API.
func (m MockMonitorClient) UpdateMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockUpdateMonitor != nil {
		return m.MockUpdateMonitor(ctx, monitorID, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// DeleteMonitor mocks the DeleteMonitor method of the Cloudflare API.
func (m MockMonitorClient) DeleteMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) error {
	if m.MockDeleteMonitor != nil {
		return m.MockDeleteMonitor(ctx, monitorID, params)
	}
	return nil
}

// MockPoolClient acts as a testable representation of the Cloudflare Load Balancer Pool API.
type MockPoolClient struct {
	MockCreatePool func(ctx context.Context, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockGetPool    func(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockUpdatePool func(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	MockDeletePool func(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) error
}

// CreatePool mocks the CreatePool method of the Cloudflare API.
func (m MockPoolClient) CreatePool(ctx context.Context, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockCreatePool != nil {
		return m.MockCreatePool(ctx, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// GetPool mocks the GetPool method of the Cloudflare API.
func (m MockPoolClient) GetPool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockGetPool != nil {
		return m.MockGetPool(ctx, poolID, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// UpdatePool mocks the UpdatePool method of the Cloudflare API.
func (m MockPoolClient) UpdatePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	if m.MockUpdatePool != nil {
		return m.MockUpdatePool(ctx, poolID, params)
	}
	return &cloudflare.LoadBalancerPool{}, nil
}

// DeletePool mocks the DeletePool method of the Cloudflare API.
func (m MockPoolClient) DeletePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) error {
	if m.MockDeletePool != nil {
		return m.MockDeletePool(ctx, poolID, params)
	}
	return nil
}
