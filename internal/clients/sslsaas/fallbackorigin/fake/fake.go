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

// MockClient is a fake implementation of the FallbackOrigin client for testing
type MockClient struct {
	MockFallbackOrigin       func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error)
	MockUpdateFallbackOrigin func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error)
	MockDeleteFallbackOrigin func(ctx context.Context, zoneID string) error
}

// FallbackOrigin calls the MockFallbackOrigin function
func (m *MockClient) FallbackOrigin(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
	if m.MockFallbackOrigin != nil {
		return m.MockFallbackOrigin(ctx, zoneID)
	}
	return cloudflare.CustomHostnameFallbackOrigin{}, nil
}

// UpdateFallbackOrigin calls the MockUpdateFallbackOrigin function
func (m *MockClient) UpdateFallbackOrigin(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
	if m.MockUpdateFallbackOrigin != nil {
		return m.MockUpdateFallbackOrigin(ctx, zoneID, origin)
	}
	return &cloudflare.CustomHostnameFallbackOriginResponse{}, nil
}

// DeleteFallbackOrigin calls the MockDeleteFallbackOrigin function
func (m *MockClient) DeleteFallbackOrigin(ctx context.Context, zoneID string) error {
	if m.MockDeleteFallbackOrigin != nil {
		return m.MockDeleteFallbackOrigin(ctx, zoneID)
	}
	return nil
}