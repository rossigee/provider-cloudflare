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
	"github.com/rossigee/provider-cloudflare/apis/transform/v1alpha1"
)

// MockClient provides a mock implementation of the Transform Rule client interface
type MockClient struct {
	MockCreateTransformRule func(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error)
	MockUpdateTransformRule func(ctx context.Context, zoneID string, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error)
	MockGetTransformRule    func(ctx context.Context, zoneID string, ruleID string, phase string) (cloudflare.RulesetRule, error)
	MockDeleteTransformRule func(ctx context.Context, zoneID string, ruleID string, phase string) error
	MockListTransformRules  func(ctx context.Context, zoneID string, phase string) ([]cloudflare.RulesetRule, error)
}

// CreateTransformRule mocks creating a transform rule
func (m *MockClient) CreateTransformRule(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
	return m.MockCreateTransformRule(ctx, zoneID, spec)
}

// UpdateTransformRule mocks updating a transform rule
func (m *MockClient) UpdateTransformRule(ctx context.Context, zoneID string, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
	return m.MockUpdateTransformRule(ctx, zoneID, ruleID, spec)
}

// GetTransformRule mocks getting a transform rule
func (m *MockClient) GetTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) (cloudflare.RulesetRule, error) {
	return m.MockGetTransformRule(ctx, zoneID, ruleID, phase)
}

// DeleteTransformRule mocks deleting a transform rule
func (m *MockClient) DeleteTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) error {
	return m.MockDeleteTransformRule(ctx, zoneID, ruleID, phase)
}

// ListTransformRules mocks listing transform rules
func (m *MockClient) ListTransformRules(ctx context.Context, zoneID string, phase string) ([]cloudflare.RulesetRule, error) {
	return m.MockListTransformRules(ctx, zoneID, phase)
}