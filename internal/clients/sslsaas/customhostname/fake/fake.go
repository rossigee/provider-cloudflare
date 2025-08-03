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
)

// MockClient is a fake implementation of the CustomHostname client for testing
type MockClient struct {
	MockCustomHostname       func(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error)
	MockCreateCustomHostname func(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	MockUpdateCustomHostname func(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	MockDeleteCustomHostname func(ctx context.Context, zoneID, hostnameID string) error
}

// CustomHostname calls the MockCustomHostname function
func (m *MockClient) CustomHostname(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
	if m.MockCustomHostname != nil {
		return m.MockCustomHostname(ctx, zoneID, hostnameID)
	}
	return cloudflare.CustomHostname{}, nil
}

// CreateCustomHostname calls the MockCreateCustomHostname function
func (m *MockClient) CreateCustomHostname(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
	if m.MockCreateCustomHostname != nil {
		return m.MockCreateCustomHostname(ctx, zoneID, hostname)
	}
	return &cloudflare.CustomHostnameResponse{}, nil
}

// UpdateCustomHostname calls the MockUpdateCustomHostname function
func (m *MockClient) UpdateCustomHostname(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
	if m.MockUpdateCustomHostname != nil {
		return m.MockUpdateCustomHostname(ctx, zoneID, hostnameID, hostname)
	}
	return &cloudflare.CustomHostnameResponse{}, nil
}

// DeleteCustomHostname calls the MockDeleteCustomHostname function
func (m *MockClient) DeleteCustomHostname(ctx context.Context, zoneID, hostnameID string) error {
	if m.MockDeleteCustomHostname != nil {
		return m.MockDeleteCustomHostname(ctx, zoneID, hostnameID)
	}
	return nil
}