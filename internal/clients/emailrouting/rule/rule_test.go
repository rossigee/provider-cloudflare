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

package rule

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	"github.com/rossigee/provider-cloudflare/apis/emailrouting/v1alpha1"
)

// MockEmailRoutingRuleAPI implements the EmailRoutingRuleAPI interface for testing
type MockEmailRoutingRuleAPI struct {
	MockCreateEmailRoutingRule func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error)
	MockGetEmailRoutingRule    func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error)
	MockUpdateEmailRoutingRule func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error)
	MockDeleteEmailRoutingRule func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error)
	MockListEmailRoutingRules  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error)
}

func (m *MockEmailRoutingRuleAPI) CreateEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
	if m.MockCreateEmailRoutingRule != nil {
		return m.MockCreateEmailRoutingRule(ctx, rc, params)
	}
	return cloudflare.EmailRoutingRule{}, nil
}

func (m *MockEmailRoutingRuleAPI) GetEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
	if m.MockGetEmailRoutingRule != nil {
		return m.MockGetEmailRoutingRule(ctx, rc, ruleTag)
	}
	return cloudflare.EmailRoutingRule{}, nil
}

func (m *MockEmailRoutingRuleAPI) UpdateEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
	if m.MockUpdateEmailRoutingRule != nil {
		return m.MockUpdateEmailRoutingRule(ctx, rc, params)
	}
	return cloudflare.EmailRoutingRule{}, nil
}

func (m *MockEmailRoutingRuleAPI) DeleteEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
	if m.MockDeleteEmailRoutingRule != nil {
		return m.MockDeleteEmailRoutingRule(ctx, rc, ruleTag)
	}
	return cloudflare.EmailRoutingRule{}, nil
}

func (m *MockEmailRoutingRuleAPI) ListEmailRoutingRules(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error) {
	if m.MockListEmailRoutingRules != nil {
		return m.MockListEmailRoutingRules(ctx, rc, params)
	}
	return []cloudflare.EmailRoutingRule{}, &cloudflare.ResultInfo{}, nil
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.RuleParameters
	}

	type want struct {
		obs *v1alpha1.RuleObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateEmailRoutingRuleSuccess": {
			reason: "Create should create Email Routing Rule when API call succeeds",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockCreateEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong resource type")
						}
						if params.Name != "Test Email Rule" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong name")
						}
						if params.Priority != 100 {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong priority")
						}
						return cloudflare.EmailRoutingRule{
							Tag:      "test-rule-tag",
							Name:     params.Name,
							Priority: params.Priority,
							Enabled:  params.Enabled,
							Matchers: []cloudflare.EmailRoutingRuleMatcher{
								{
									Type:  "literal",
									Field: "to",
									Value: "test@example.com",
								},
							},
							Actions: []cloudflare.EmailRoutingRuleAction{
								{
									Type:  "forward",
									Value: []string{"user@domain.com"},
								},
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Email Rule",
					Priority: 100,
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "test@example.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"user@domain.com"},
						},
					},
				},
			},
			want: want{
				obs: &v1alpha1.RuleObservation{
					Tag:      "test-rule-tag",
					Name:     "Test Email Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "test@example.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"user@domain.com"},
						},
					},
				},
				err: nil,
			},
		},
		"CreateEmailRoutingRuleMinimal": {
			reason: "Create should create Email Routing Rule with minimal parameters",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockCreateEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{
							Tag:      "minimal-rule-tag",
							Name:     params.Name,
							Priority: params.Priority,
							Enabled:  params.Enabled,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Minimal Rule",
					Priority: 1,
					Enabled:  ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.RuleObservation{
					Tag:      "minimal-rule-tag",
					Name:     "Minimal Rule",
					Priority: ptr.To(1),
					Enabled:  ptr.To(false),
				},
				err: nil,
			},
		},
		"CreateEmailRoutingRuleAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockCreateEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 10,
					Enabled:  ptr.To(true),
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errCreateRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Create(tc.args.ctx, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	ruleTag := "test-rule-tag"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx     context.Context
		zoneID  string
		ruleTag string
	}

	type want struct {
		obs *v1alpha1.RuleObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetEmailRoutingRuleSuccess": {
			reason: "Get should return Email Routing Rule when API call succeeds",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockGetEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong resource type")
						}
						if ruleTag != "test-rule-tag" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong rule tag")
						}
						return cloudflare.EmailRoutingRule{
							Tag:      ruleTag,
							Name:     "Retrieved Rule",
							Priority: 50,
							Enabled:  ptr.To(true),
							Matchers: []cloudflare.EmailRoutingRuleMatcher{
								{
									Type:  "contains",
									Field: "subject",
									Value: "urgent",
								},
							},
							Actions: []cloudflare.EmailRoutingRuleAction{
								{
									Type:  "forward",
									Value: []string{"admin@example.com"},
								},
								{
									Type:  "worker",
									Value: []string{"email-processor"},
								},
							},
						}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				zoneID:  zoneID,
				ruleTag: ruleTag,
			},
			want: want{
				obs: &v1alpha1.RuleObservation{
					Tag:      "test-rule-tag",
					Name:     "Retrieved Rule",
					Priority: ptr.To(50),
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "contains",
							Field: "subject",
							Value: "urgent",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"admin@example.com"},
						},
						{
							Type:  "worker",
							Value: []string{"email-processor"},
						},
					},
				},
				err: nil,
			},
		},
		"GetEmailRoutingRuleAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockGetEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{}, errBoom
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				zoneID:  zoneID,
				ruleTag: ruleTag,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errGetRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.zoneID, tc.args.ruleTag)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nGet(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nGet(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	ruleTag := "test-rule-tag"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx     context.Context
		ruleTag string
		params  v1alpha1.RuleParameters
	}

	type want struct {
		obs *v1alpha1.RuleObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateEmailRoutingRuleSuccess": {
			reason: "Update should update Email Routing Rule when API call succeeds",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockUpdateEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong resource type")
						}
						if params.Name != "Updated Rule" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong name")
						}
						return cloudflare.EmailRoutingRule{
							Tag:      "test-rule-tag",
							Name:     params.Name,
							Priority: params.Priority,
							Enabled:  params.Enabled,
							Matchers: params.Matchers,
							Actions:  params.Actions,
						}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				ruleTag: ruleTag,
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Updated Rule",
					Priority: 75,
					Enabled:  ptr.To(false),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "regex",
							Field: "from",
							Value: ".*@partner\\.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "drop",
							Value: []string{},
						},
					},
				},
			},
			want: want{
				obs: &v1alpha1.RuleObservation{
					Tag:      "test-rule-tag",
					Name:     "Updated Rule",
					Priority: ptr.To(75),
					Enabled:  ptr.To(false),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "regex",
							Field: "from",
							Value: ".*@partner\\.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "drop",
							Value: []string{},
						},
					},
				},
				err: nil,
			},
		},
		"UpdateEmailRoutingRuleAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockUpdateEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{}, errBoom
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				ruleTag: ruleTag,
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 10,
					Enabled:  ptr.To(true),
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errUpdateRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.ruleTag, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	ruleTag := "test-rule-tag"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx     context.Context
		zoneID  string
		ruleTag string
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"DeleteEmailRoutingRuleSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockDeleteEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong resource type")
						}
						if ruleTag != "test-rule-tag" {
							return cloudflare.EmailRoutingRule{}, errors.New("wrong rule tag")
						}
						return cloudflare.EmailRoutingRule{}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				zoneID:  zoneID,
				ruleTag: ruleTag,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteEmailRoutingRuleNotFound": {
			reason: "Delete should succeed when Email Routing Rule is not found (already deleted)",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockDeleteEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{}, errors.New("rule not found")
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				zoneID:  zoneID,
				ruleTag: ruleTag,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteEmailRoutingRuleAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockDeleteEmailRoutingRule: func(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error) {
						return cloudflare.EmailRoutingRule{}, errBoom
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				zoneID:  zoneID,
				ruleTag: ruleTag,
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.zoneID, tc.args.ruleTag)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestList(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx    context.Context
		zoneID string
	}

	type want struct {
		obs []v1alpha1.RuleObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ListEmailRoutingRulesSuccess": {
			reason: "List should return Email Routing Rules when API call succeeds",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockListEmailRoutingRules: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error) {
						if rc.Identifier != "test-zone-id" {
							return nil, &cloudflare.ResultInfo{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return nil, &cloudflare.ResultInfo{}, errors.New("wrong resource type")
						}
						return []cloudflare.EmailRoutingRule{
							{
								Tag:      "rule1",
								Name:     "First Rule",
								Priority: 100,
								Enabled:  ptr.To(true),
							},
							{
								Tag:      "rule2",
								Name:     "Second Rule",
								Priority: 200,
								Enabled:  ptr.To(false),
							},
						}, &cloudflare.ResultInfo{Count: 2}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: []v1alpha1.RuleObservation{
					{
						Tag:      "rule1",
						Name:     "First Rule",
						Priority: ptr.To(100),
						Enabled:  ptr.To(true),
					},
					{
						Tag:      "rule2",
						Name:     "Second Rule",
						Priority: ptr.To(200),
						Enabled:  ptr.To(false),
					},
				},
				err: nil,
			},
		},
		"ListEmailRoutingRulesEmpty": {
			reason: "List should return empty slice when no rules exist",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockListEmailRoutingRules: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error) {
						return []cloudflare.EmailRoutingRule{}, &cloudflare.ResultInfo{Count: 0}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: []v1alpha1.RuleObservation{},
				err: nil,
			},
		},
		"ListEmailRoutingRulesAPIError": {
			reason: "List should return wrapped error when API call fails",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{
					MockListEmailRoutingRules: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error) {
						return nil, &cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errListRules),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.List(tc.args.ctx, tc.args.zoneID)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nList(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nList(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	zoneID := "test-zone-id"

	type fields struct {
		client *MockEmailRoutingRuleAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.RuleParameters
		obs    v1alpha1.RuleObservation
	}

	type want struct {
		upToDate bool
		err      error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"IsUpToDateTrue": {
			reason: "IsUpToDate should return true when all settings match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 100,
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "test@example.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"user@domain.com"},
						},
					},
				},
				obs: v1alpha1.RuleObservation{
					Tag:      "rule-tag",
					Name:     "Test Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "test@example.com",
						},
					},
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"user@domain.com"},
						},
					},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseName": {
			reason: "IsUpToDate should return false when name doesn't match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Updated Rule",
					Priority: 100,
					Enabled:  ptr.To(true),
				},
				obs: v1alpha1.RuleObservation{
					Name:     "Original Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalsePriority": {
			reason: "IsUpToDate should return false when priority doesn't match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 200,
					Enabled:  ptr.To(true),
				},
				obs: v1alpha1.RuleObservation{
					Name:     "Test Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseEnabled": {
			reason: "IsUpToDate should return false when enabled doesn't match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 100,
					Enabled:  ptr.To(false),
				},
				obs: v1alpha1.RuleObservation{
					Name:     "Test Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseMatchers": {
			reason: "IsUpToDate should return false when matchers don't match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 100,
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "new@example.com",
						},
					},
				},
				obs: v1alpha1.RuleObservation{
					Name:     "Test Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
					Matchers: []v1alpha1.RuleMatcher{
						{
							Type:  "literal",
							Field: "to",
							Value: "old@example.com",
						},
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseActions": {
			reason: "IsUpToDate should return false when actions don't match",
			fields: fields{
				client: &MockEmailRoutingRuleAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RuleParameters{
					ZoneID:   zoneID,
					Name:     "Test Rule",
					Priority: 100,
					Enabled:  ptr.To(true),
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"new@domain.com"},
						},
					},
				},
				obs: v1alpha1.RuleObservation{
					Name:     "Test Rule",
					Priority: ptr.To(100),
					Enabled:  ptr.To(true),
					Actions: []v1alpha1.RuleAction{
						{
							Type:  "forward",
							Value: []string{"old@domain.com"},
						},
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.IsUpToDate(tc.args.ctx, tc.args.params, tc.args.obs)
			
			if diff := cmp.Diff(tc.want.err, err, cmp.Comparer(func(x, y error) bool {
				if x == nil && y == nil {
					return true
				}
				if x == nil || y == nil {
					return false
				}
				return x.Error() == y.Error()
			})); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsRuleNotFound(t *testing.T) {
	type args struct {
		err error
	}

	type want struct {
		isNotFound bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NilError": {
			reason: "IsRuleNotFound should return false for nil error",
			args: args{
				err: nil,
			},
			want: want{
				isNotFound: false,
			},
		},
		"RuleNotFoundError": {
			reason: "IsRuleNotFound should return true for 'rule not found' error",
			args: args{
				err: errors.New("rule not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"NotFoundError404": {
			reason: "IsRuleNotFound should return true for '404' error",
			args: args{
				err: errors.New("404"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"NotFoundError": {
			reason: "IsRuleNotFound should return true for 'Not found' error",
			args: args{
				err: errors.New("Not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "IsRuleNotFound should return false for other errors",
			args: args{
				err: errors.New("some other error"),
			},
			want: want{
				isNotFound: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRuleNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.isNotFound, got); diff != "" {
				t.Errorf("\n%s\nIsRuleNotFound(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}