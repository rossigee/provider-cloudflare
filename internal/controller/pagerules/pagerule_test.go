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

package pagerules

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/pagerules/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/pagerules"
)

type mockPageRuleClient struct {
	MockCreatePageRule func(ctx context.Context, zoneID string, rule cloudflare.PageRule) (*cloudflare.PageRule, error)
	MockGetPageRule    func(ctx context.Context, zoneID, ruleID string) (cloudflare.PageRule, error)
	MockUpdatePageRule func(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error
	MockDeletePageRule func(ctx context.Context, zoneID, ruleID string) error
}

func (m *mockPageRuleClient) CreatePageRule(ctx context.Context, zoneID string, rule cloudflare.PageRule) (*cloudflare.PageRule, error) {
	if m.MockCreatePageRule != nil {
		return m.MockCreatePageRule(ctx, zoneID, rule)
	}
	return &cloudflare.PageRule{ID: "created-123"}, nil
}

func (m *mockPageRuleClient) GetPageRule(ctx context.Context, zoneID, ruleID string) (cloudflare.PageRule, error) {
	if m.MockGetPageRule != nil {
		return m.MockGetPageRule(ctx, zoneID, ruleID)
	}
	return cloudflare.PageRule{ID: ruleID, Priority: 1, Status: "active"}, nil
}

func (m *mockPageRuleClient) UpdatePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error {
	if m.MockUpdatePageRule != nil {
		return m.MockUpdatePageRule(ctx, zoneID, ruleID, rule)
	}
	return nil
}

func (m *mockPageRuleClient) DeletePageRule(ctx context.Context, zoneID, ruleID string) error {
	if m.MockDeletePageRule != nil {
		return m.MockDeletePageRule(ctx, zoneID, ruleID)
	}
	return nil
}

func (m *mockPageRuleClient) ListPageRules(ctx context.Context, zoneID string) ([]cloudflare.PageRule, error) {
	return nil, nil
}

func (m *mockPageRuleClient) ChangePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error {
	return nil
}

func pageRule(mods ...func(*v1beta1.PageRule)) *v1beta1.PageRule {
	pr := &v1beta1.PageRule{
		Spec: v1beta1.PageRuleSpec{
			ForProvider: v1beta1.PageRuleParameters{
				ZoneID: "zone-123",
				Targets: []v1beta1.Target{{
					Target:     "url",
					Constraint: v1beta1.TargetConstraint{Operator: "matches", Value: "*"},
				}},
				Actions: []v1beta1.Action{{ID: "forwarding_url", Value: "{}"}},
			},
		},
	}
	for _, m := range mods {
		m(pr)
	}
	return pr
}

func TestPageRuleConnect(t *testing.T) {
	// Connect requires valid ProviderConfig + secret setup for full happy path.
	// Main logic (Observe/Create etc) is tested by injecting the service directly.
	t.Skip("Connect happy path requires full config secret mocking - covered indirectly via integration")
}

func TestPageRuleObserve(t *testing.T) {
	cases := map[string]struct {
		service pagerules.CloudflarePageRuleClient
		mg      resource.Managed
		want    managed.ExternalObservation
		wantErr error
	}{
		"NoExternalName": {
			mg:   pageRule(),
			want: managed.ExternalObservation{ResourceExists: false},
		},
		"Success": {
			service: &mockPageRuleClient{},
			mg:      pageRule(func(pr *v1beta1.PageRule) { meta.SetExternalName(pr, "rule-123") }),
			want: managed.ExternalObservation{
				ResourceExists:    true,
				ResourceUpToDate:  true,
				ConnectionDetails: managed.ConnectionDetails{},
			},
		},
		"NotFound": {
			service: &mockPageRuleClient{
				MockGetPageRule: func(ctx context.Context, zone, id string) (cloudflare.PageRule, error) {
					return cloudflare.PageRule{}, errors.New("not found")
				},
			},
			mg:   pageRule(func(pr *v1beta1.PageRule) { meta.SetExternalName(pr, "rule-123") }),
			want: managed.ExternalObservation{ResourceExists: false},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pageRuleExternal{service: tc.service}
			got, err := e.Observe(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s Observe err: -want +got\n%s", name, diff)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s Observe: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPageRuleCreate(t *testing.T) {
	cases := map[string]struct {
		service pagerules.CloudflarePageRuleClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPageRuleClient{},
			mg:      pageRule(),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pageRuleExternal{service: tc.service}
			_, err := e.Create(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s Create: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPageRuleUpdate(t *testing.T) {
	cases := map[string]struct {
		service pagerules.CloudflarePageRuleClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPageRuleClient{},
			mg:      pageRule(func(pr *v1beta1.PageRule) { meta.SetExternalName(pr, "rule-123") }),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pageRuleExternal{service: tc.service}
			_, err := e.Update(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s Update: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPageRuleDelete(t *testing.T) {
	cases := map[string]struct {
		service pagerules.CloudflarePageRuleClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPageRuleClient{},
			mg:      pageRule(func(pr *v1beta1.PageRule) { meta.SetExternalName(pr, "rule-123") }),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pageRuleExternal{service: tc.service}
			_, err := e.Delete(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s Delete: -want +got\n%s", name, diff)
			}
		})
	}
}
