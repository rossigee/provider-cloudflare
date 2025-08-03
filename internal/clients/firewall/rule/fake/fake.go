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

// MockClient is a fake implementation of the Rule client for testing
type MockClient struct {
	MockFirewallRule       func(ctx context.Context, zoneID, ruleID string) (cloudflare.FirewallRule, error)
	MockCreateFirewallRule func(ctx context.Context, zoneID string, rule cloudflare.FirewallRule) (*cloudflare.FirewallRule, error)
	MockUpdateFirewallRule func(ctx context.Context, zoneID, ruleID string, rule cloudflare.FirewallRule) error
	MockDeleteFirewallRule func(ctx context.Context, zoneID, ruleID string) error
}

// FirewallRule calls the MockFirewallRule function
func (m *MockClient) FirewallRule(ctx context.Context, zoneID, ruleID string) (cloudflare.FirewallRule, error) {
	if m.MockFirewallRule != nil {
		return m.MockFirewallRule(ctx, zoneID, ruleID)
	}
	return cloudflare.FirewallRule{}, nil
}

// CreateFirewallRule calls the MockCreateFirewallRule function
func (m *MockClient) CreateFirewallRule(ctx context.Context, zoneID string, rule cloudflare.FirewallRule) (*cloudflare.FirewallRule, error) {
	if m.MockCreateFirewallRule != nil {
		return m.MockCreateFirewallRule(ctx, zoneID, rule)
	}
	return &cloudflare.FirewallRule{}, nil
}

// UpdateFirewallRule calls the MockUpdateFirewallRule function
func (m *MockClient) UpdateFirewallRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.FirewallRule) error {
	if m.MockUpdateFirewallRule != nil {
		return m.MockUpdateFirewallRule(ctx, zoneID, ruleID, rule)
	}
	return nil
}

// DeleteFirewallRule calls the MockDeleteFirewallRule function
func (m *MockClient) DeleteFirewallRule(ctx context.Context, zoneID, ruleID string) error {
	if m.MockDeleteFirewallRule != nil {
		return m.MockDeleteFirewallRule(ctx, zoneID, ruleID)
	}
	return nil
}