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

package posturerule

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	"github.com/pkg/errors"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

// mockDevicePostureRuleAPI mocks the DevicePostureRuleAPI interface
type mockDevicePostureRuleAPI struct {
	MockDevicePostureRules    func(ctx context.Context, accountID string) ([]cloudflare.DevicePostureRule, cloudflare.ResultInfo, error)
	MockDevicePostureRule     func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error)
	MockCreateDevicePostureRule func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	MockUpdateDevicePostureRule func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	MockDeleteDevicePostureRule func(ctx context.Context, accountID, ruleID string) error
}

func (m *mockDevicePostureRuleAPI) DevicePostureRules(ctx context.Context, accountID string) ([]cloudflare.DevicePostureRule, cloudflare.ResultInfo, error) {
	return m.MockDevicePostureRules(ctx, accountID)
}

func (m *mockDevicePostureRuleAPI) DevicePostureRule(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
	return m.MockDevicePostureRule(ctx, accountID, ruleID)
}

func (m *mockDevicePostureRuleAPI) CreateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
	return m.MockCreateDevicePostureRule(ctx, accountID, rule)
}

func (m *mockDevicePostureRuleAPI) UpdateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
	return m.MockUpdateDevicePostureRule(ctx, accountID, rule)
}

func (m *mockDevicePostureRuleAPI) DeleteDevicePostureRule(ctx context.Context, accountID, ruleID string) error {
	return m.MockDeleteDevicePostureRule(ctx, accountID, ruleID)
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params v1beta1.DevicePostureRuleParameters
		obs    v1beta1.DevicePostureRuleObservation
	}

	type want struct {
		upToDate bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"UpToDateIdentical": {
			reason: "Should return true when all fields match",
			args: args{
				params: v1beta1.DevicePostureRuleParameters{
					AccountID:   "test-account-id",
					Name:        "test-rule",
					Type:        "os_version",
					Description: stringPtr("Test rule"),
					Schedule:    stringPtr("5m"),
					Expiration:  stringPtr("24h"),
				},
				obs: v1beta1.DevicePostureRuleObservation{
					Name:        "test-rule",
					Type:        "os_version",
					Description: "Test rule",
					Schedule:    "5m",
					Expiration:  "24h",
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"OutOfDateName": {
			reason: "Should return false when name differs",
			args: args{
				params: v1beta1.DevicePostureRuleParameters{
					AccountID: "test-account-id",
					Name:      "updated-rule",
					Type:      "os_version",
				},
				obs: v1beta1.DevicePostureRuleObservation{
					Name: "original-rule",
					Type: "os_version",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &CloudflareDevicePostureRuleClient{}
			got, err := client.IsUpToDate(context.Background(), tc.args.params, tc.args.obs)
			if err != nil {
				t.Errorf("IsUpToDate() error = %v", err)
				return
			}
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("%s\nIsUpToDate(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestConvertParametersToDevicePostureRule(t *testing.T) {
	type args struct {
		params v1beta1.DevicePostureRuleParameters
	}

	type want struct {
		rule cloudflare.DevicePostureRule
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"BasicRule": {
			reason: "Should convert basic parameters correctly",
			args: args{
				params: v1beta1.DevicePostureRuleParameters{
					AccountID:   "test-account-id",
					Name:        "test-rule",
					Type:        "os_version",
					Description: stringPtr("Test device posture rule"),
					Schedule:    stringPtr("5m"),
					Expiration:  stringPtr("24h"),
					Input: &v1beta1.DevicePostureRuleInput{
						OS:       stringPtr("windows"),
						Operator: stringPtr(">="),
						Version:  stringPtr("10.0.19043"),
					},
					Match: []v1beta1.DevicePostureRuleMatch{
						{
							Platform: stringPtr("windows"),
						},
					},
				},
			},
			want: want{
				rule: cloudflare.DevicePostureRule{
					Name:        "test-rule",
					Type:        "os_version",
					Description: "Test device posture rule",
					Schedule:    "5m",
					Expiration:  "24h",
				},
			},
		},
		"MinimalRule": {
			reason: "Should convert minimal parameters correctly",
			args: args{
				params: v1beta1.DevicePostureRuleParameters{
					AccountID: "test-account-id",
					Name:      "test-rule",
					Type:      "os_version",
				},
			},
			want: want{
				rule: cloudflare.DevicePostureRule{
					Name: "test-rule",
					Type: "os_version",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToDevicePostureRule(tc.args.params)
			if got.Name != tc.want.rule.Name || got.Type != tc.want.rule.Type {
				t.Errorf("%s\nconvertParametersToDevicePostureRule(...): name=%s (expected %s), type=%s (expected %s)",
					tc.reason, got.Name, tc.want.rule.Name, got.Type, tc.want.rule.Type)
			}
		})
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	accountID := "account-123"
	ruleID := "rule-456"

	cases := map[string]struct {
		reason  string
		mock    func() DevicePostureRuleAPI
		wantErr bool
		verify  func(t *testing.T, obs *v1beta1.DevicePostureRuleObservation)
	}{
		"SuccessGet": {
			reason: "Should successfully get a rule",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:          ruleID,
							Name:        "test-rule",
							Type:        "os_version",
							Description: "Test rule",
							Schedule:    "5m",
							Expiration:  "24h",
						}, nil
					},
				}
			},
			wantErr: false,
			verify: func(t *testing.T, obs *v1beta1.DevicePostureRuleObservation) {
				if obs == nil || obs.Name != "test-rule" {
					t.Errorf("Expected rule name 'test-rule'")
				}
			},
		},
		"NotFound": {
			reason: "Should return error when rule not found",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("device posture rule not found")
					},
				}
			},
			wantErr: true,
		},
		"APIError": {
			reason: "Should return error from API",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("API error")
					},
				}
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mock())
			obs, err := client.Get(ctx, accountID, ruleID)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
			if !tc.wantErr && tc.verify != nil {
				tc.verify(t, obs)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	accountID := "account-123"

	cases := map[string]struct {
		reason  string
		mock    func() DevicePostureRuleAPI
		params  v1beta1.DevicePostureRuleParameters
		wantErr bool
	}{
		"SuccessCreate": {
			reason: "Should successfully create a rule",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockCreateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   "new-rule-id",
							Name: rule.Name,
							Type: rule.Type,
						}, nil
					},
				}
			},
			params: v1beta1.DevicePostureRuleParameters{
				AccountID: accountID,
				Name:      "new-rule",
				Type:      "os_version",
			},
			wantErr: false,
		},
		"APIError": {
			reason: "Should return error from API",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockCreateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("API error")
					},
				}
			},
			params: v1beta1.DevicePostureRuleParameters{
				AccountID: accountID,
				Name:      "new-rule",
				Type:      "os_version",
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mock())
			_, err := client.Create(ctx, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	accountID := "account-123"
	ruleID := "rule-456"

	cases := map[string]struct {
		reason  string
		mock    func() DevicePostureRuleAPI
		params  v1beta1.DevicePostureRuleParameters
		wantErr bool
	}{
		"SuccessUpdate": {
			reason: "Should successfully update a rule",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockUpdateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   rule.ID,
							Name: rule.Name,
							Type: rule.Type,
						}, nil
					},
				}
			},
			params: v1beta1.DevicePostureRuleParameters{
				AccountID: accountID,
				Name:      "updated-rule",
				Type:      "os_version",
			},
			wantErr: false,
		},
		"APIError": {
			reason: "Should return error from API",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockUpdateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("API error")
					},
				}
			},
			params: v1beta1.DevicePostureRuleParameters{
				AccountID: accountID,
				Name:      "updated-rule",
				Type:      "os_version",
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mock())
			_, err := client.Update(ctx, ruleID, tc.params)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	accountID := "account-123"
	ruleID := "rule-456"

	cases := map[string]struct {
		reason  string
		mock    func() DevicePostureRuleAPI
		wantErr bool
	}{
		"SuccessDelete": {
			reason: "Should successfully delete a rule",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDeleteDevicePostureRule: func(ctx context.Context, accountID, ruleID string) error {
						return nil
					},
				}
			},
			wantErr: false,
		},
		"NotFoundIgnored": {
			reason: "Should not error when rule already deleted",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDeleteDevicePostureRule: func(ctx context.Context, accountID, ruleID string) error {
						return errors.New("rule not found")
					},
				}
			},
			wantErr: false,
		},
		"APIError": {
			reason: "Should return error for non-not-found errors",
			mock: func() DevicePostureRuleAPI {
				return &mockDevicePostureRuleAPI{
					MockDeleteDevicePostureRule: func(ctx context.Context, accountID, ruleID string) error {
						return errors.New("API error")
					},
				}
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mock())
			err := client.Delete(ctx, accountID, ruleID)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
		})
	}
}

func TestConvertDevicePostureRuleToObservation(t *testing.T) {
	cases := map[string]struct {
		reason string
		rule   cloudflare.DevicePostureRule
		verify func(t *testing.T, obs *v1beta1.DevicePostureRuleObservation)
	}{
		"WithAllFields": {
			reason: "Should convert rule with all fields",
			rule: cloudflare.DevicePostureRule{
				ID:          "rule-123",
				Name:        "test-rule",
				Type:        "os_version",
				Description: "Test description",
				Schedule:    "5m",
				Expiration:  "24h",
				Match: []cloudflare.DevicePostureRuleMatch{
					{Platform: "windows"},
				},
				Input: cloudflare.DevicePostureRuleInput{
					ID:       "input-123",
					OS:       "windows",
					Version:  "10.0.19043",
					Operator: ">=",
				},
			},
			verify: func(t *testing.T, obs *v1beta1.DevicePostureRuleObservation) {
				if obs == nil || obs.ID != "rule-123" || obs.Name != "test-rule" {
					t.Errorf("Observation mismatch")
				}
				if obs.Match == nil || len(obs.Match) != 1 {
					t.Errorf("Match conversion failed")
				}
				if obs.Input == nil {
					t.Errorf("Input conversion failed")
				}
			},
		},
		"WithMinimalFields": {
			reason: "Should convert rule with minimal fields",
			rule: cloudflare.DevicePostureRule{
				ID:   "rule-456",
				Name: "minimal-rule",
				Type: "os_version",
			},
			verify: func(t *testing.T, obs *v1beta1.DevicePostureRuleObservation) {
				if obs == nil || obs.ID != "rule-456" {
					t.Errorf("ID should be set")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			obs := convertDevicePostureRuleToObservation(tc.rule)
			tc.verify(t, obs)
		})
	}
}

func TestConvertMatchFromCloudflare(t *testing.T) {
	cases := map[string]struct {
		reason  string
		matches []cloudflare.DevicePostureRuleMatch
		verify  func(t *testing.T, result []v1beta1.DevicePostureRuleMatch)
	}{
		"SingleMatch": {
			reason: "Should convert single match",
			matches: []cloudflare.DevicePostureRuleMatch{
				{Platform: "windows"},
			},
			verify: func(t *testing.T, result []v1beta1.DevicePostureRuleMatch) {
				if len(result) != 1 || result[0].Platform == nil || *result[0].Platform != "windows" {
					t.Errorf("Match conversion failed")
				}
			},
		},
		"MultipleMatches": {
			reason: "Should convert multiple matches",
			matches: []cloudflare.DevicePostureRuleMatch{
				{Platform: "windows"},
				{Platform: "macos"},
			},
			verify: func(t *testing.T, result []v1beta1.DevicePostureRuleMatch) {
				if len(result) != 2 {
					t.Errorf("Expected 2 matches, got %d", len(result))
				}
			},
		},
		"EmptyMatches": {
			reason: "Should handle empty matches",
			matches: []cloudflare.DevicePostureRuleMatch{},
			verify: func(t *testing.T, result []v1beta1.DevicePostureRuleMatch) {
				if len(result) != 0 {
					t.Errorf("Expected empty matches")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := convertMatchFromCloudflare(tc.matches)
			tc.verify(t, result)
		})
	}
}

func TestConvertInputFromCloudflare(t *testing.T) {
	cases := map[string]struct {
		reason string
		input  *cloudflare.DevicePostureRuleInput
		verify func(t *testing.T, result *v1beta1.DevicePostureRuleInput)
	}{
		"NilInput": {
			reason: "Should handle nil input",
			input:  nil,
			verify: func(t *testing.T, result *v1beta1.DevicePostureRuleInput) {
				if result != nil {
					t.Errorf("Expected nil result")
				}
			},
		},
		"WithAllFields": {
			reason: "Should convert input with all fields",
			input: &cloudflare.DevicePostureRuleInput{
				ID:              "input-123",
				Path:            "/path/to/file",
				OS:              "windows",
				Version:         "10.0.19043",
				Operator:        ">=",
				Domain:          "example.com",
				Thumbprint:      "thumb123",
				Sha256:          "hash123",
				Running:         boolPtr(true),
				RequireAll:      boolPtr(false),
				Enabled:         boolPtr(true),
				CheckDisks:      []string{"/disk1", "/disk2"},
				CheckPrivateKey: boolPtr(true),
				ComplianceStatus: "compliant",
				ConnectionID:     "conn-123",
				SensorConfig:     "active",
				CertificateID:    "cert-123",
				CommonName:       "example.com",
				Exists:           boolPtr(true),
				Overall:          "ok",
				State:            "enabled",
			},
			verify: func(t *testing.T, result *v1beta1.DevicePostureRuleInput) {
				if result == nil {
					t.Errorf("Expected non-nil result")
					return
				}
				if result.ID == nil || *result.ID != "input-123" {
					t.Errorf("ID conversion failed")
				}
				if result.OS == nil || *result.OS != "windows" {
					t.Errorf("OS conversion failed")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			result := convertInputFromCloudflare(tc.input)
			tc.verify(t, result)
		})
	}
}

func TestIsUpToDateEnhanced(t *testing.T) {
	cases := map[string]struct {
		reason  string
		params  v1beta1.DevicePostureRuleParameters
		obs     v1beta1.DevicePostureRuleObservation
		want    bool
	}{
		"UpToDateWithAllOptionals": {
			reason: "Should return true when all optional fields match",
			params: v1beta1.DevicePostureRuleParameters{
				Name:        "rule",
				Type:        "os_version",
				Description: stringPtr("desc"),
				Schedule:    stringPtr("5m"),
				Expiration:  stringPtr("24h"),
			},
			obs: v1beta1.DevicePostureRuleObservation{
				Name:        "rule",
				Type:        "os_version",
				Description: "desc",
				Schedule:    "5m",
				Expiration:  "24h",
			},
			want: true,
		},
		"OutOfDateDescription": {
			reason: "Should return false when description differs",
			params: v1beta1.DevicePostureRuleParameters{
				Name:        "rule",
				Type:        "os_version",
				Description: stringPtr("new desc"),
			},
			obs: v1beta1.DevicePostureRuleObservation{
				Name:        "rule",
				Type:        "os_version",
				Description: "old desc",
			},
			want: false,
		},
		"OutOfDateSchedule": {
			reason: "Should return false when schedule differs",
			params: v1beta1.DevicePostureRuleParameters{
				Name:     "rule",
				Type:     "os_version",
				Schedule: stringPtr("10m"),
			},
			obs: v1beta1.DevicePostureRuleObservation{
				Name:     "rule",
				Type:     "os_version",
				Schedule: "5m",
			},
			want: false,
		},
		"OutOfDateExpiration": {
			reason: "Should return false when expiration differs",
			params: v1beta1.DevicePostureRuleParameters{
				Name:       "rule",
				Type:       "os_version",
				Expiration: stringPtr("48h"),
			},
			obs: v1beta1.DevicePostureRuleObservation{
				Name:       "rule",
				Type:       "os_version",
				Expiration: "24h",
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &CloudflareDevicePostureRuleClient{}
			got, err := client.IsUpToDate(context.Background(), tc.params, tc.obs)
			if err != nil {
				t.Errorf("%s: unexpected error %v", tc.reason, err)
			}
			if got != tc.want {
				t.Errorf("%s: expected %v, got %v", tc.reason, tc.want, got)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	type args struct {
		err error
	}

	type want struct {
		notFound bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NotFoundError": {
			reason: "Should return true for not found error",
			args: args{
				err: errors.New("device posture rule not found"),
			},
			want: want{
				notFound: true,
			},
		},
		"StringNotFound": {
			reason: "Should return true for string containing 'not found'",
			args: args{
				err: errors.New("device posture rule not found"),
			},
			want: want{
				notFound: true,
			},
		},
		"OtherError": {
			reason: "Should return false for other errors",
			args: args{
				err: errors.New("internal server error"),
			},
			want: want{
				notFound: false,
			},
		},
		"NilError": {
			reason: "Should return false for nil error",
			args: args{
				err: nil,
			},
			want: want{
				notFound: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.notFound, got); diff != "" {
				t.Errorf("%s\nisNotFound(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
