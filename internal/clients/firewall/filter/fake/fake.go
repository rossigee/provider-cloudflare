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

// MockClient is a fake implementation of the Filter client for testing
type MockClient struct {
	MockFilter       func(ctx context.Context, zoneID, filterID string) (cloudflare.Filter, error)
	MockCreateFilter func(ctx context.Context, zoneID string, filter cloudflare.Filter) (*cloudflare.Filter, error)
	MockUpdateFilter func(ctx context.Context, zoneID, filterID string, filter cloudflare.Filter) error
	MockDeleteFilter func(ctx context.Context, zoneID, filterID string) error
}

// Filter calls the MockFilter function
func (m *MockClient) Filter(ctx context.Context, zoneID, filterID string) (cloudflare.Filter, error) {
	if m.MockFilter != nil {
		return m.MockFilter(ctx, zoneID, filterID)
	}
	return cloudflare.Filter{}, nil
}

// CreateFilter calls the MockCreateFilter function
func (m *MockClient) CreateFilter(ctx context.Context, zoneID string, filter cloudflare.Filter) (*cloudflare.Filter, error) {
	if m.MockCreateFilter != nil {
		return m.MockCreateFilter(ctx, zoneID, filter)
	}
	return &cloudflare.Filter{}, nil
}

// UpdateFilter calls the MockUpdateFilter function
func (m *MockClient) UpdateFilter(ctx context.Context, zoneID, filterID string, filter cloudflare.Filter) error {
	if m.MockUpdateFilter != nil {
		return m.MockUpdateFilter(ctx, zoneID, filterID, filter)
	}
	return nil
}

// DeleteFilter calls the MockDeleteFilter function
func (m *MockClient) DeleteFilter(ctx context.Context, zoneID, filterID string) error {
	if m.MockDeleteFilter != nil {
		return m.MockDeleteFilter(ctx, zoneID, filterID)
	}
	return nil
}