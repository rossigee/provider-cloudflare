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

// MockMonitorClient acts as a testable representation of the Cloudflare Load Balancer Monitor API.
type MockMonitorClient struct {
	MockCreateMonitor func(ctx context.Context, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockGetMonitor    func(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockUpdateMonitor func(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	MockDeleteMonitor func(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) error
}

// CreateMonitor mocks the CreateMonitor method of the Cloudflare API.
func (m MockMonitorClient) CreateMonitor(ctx context.Context, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockCreateMonitor != nil {
		return m.MockCreateMonitor(ctx, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// GetMonitor mocks the GetMonitor method of the Cloudflare API.
func (m MockMonitorClient) GetMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockGetMonitor != nil {
		return m.MockGetMonitor(ctx, monitorID, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// UpdateMonitor mocks the UpdateMonitor method of the Cloudflare API.
func (m MockMonitorClient) UpdateMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	if m.MockUpdateMonitor != nil {
		return m.MockUpdateMonitor(ctx, monitorID, params)
	}
	return &cloudflare.LoadBalancerMonitor{}, nil
}

// DeleteMonitor mocks the DeleteMonitor method of the Cloudflare API.
func (m MockMonitorClient) DeleteMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) error {
	if m.MockDeleteMonitor != nil {
		return m.MockDeleteMonitor(ctx, monitorID, params)
	}
	return nil
}