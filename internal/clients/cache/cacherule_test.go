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

package cache

import (
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"

	"github.com/rossigee/provider-cloudflare/apis/cache/v1alpha1"
)

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func uintPtr(u uint) *uint {
	return &u
}

func TestGenerateCacheRuleObservation(t *testing.T) {
	lastUpdated := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	version := "1"
	
	rule := &cloudflare.RulesetRule{
		ID:          "test-rule-id",
		Version:     &version,
		LastUpdated: &lastUpdated,
	}

	ruleset := &cloudflare.Ruleset{
		ID:          "test-ruleset-id",
		Version:     &version,
		LastUpdated: &lastUpdated,
	}

	expected := v1alpha1.CacheRuleObservation{
		ID:          "test-rule-id",
		RulesetID:   "test-ruleset-id",
		Version:     "1",
		LastUpdated: stringPtr("2025-01-01 00:00:00 +0000 UTC"),
		ModifiedOn:  stringPtr("2025-01-01 00:00:00 +0000 UTC"),
	}

	result := GenerateCacheRuleObservation(rule, ruleset)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("GenerateCacheRuleObservation(...): -want, +got:\n%s", diff)
	}
}

func TestIsCacheRuleUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.CacheRuleParameters
		rule   *cloudflare.RulesetRule
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
				params: &v1alpha1.CacheRuleParameters{
					Zone:        "test-zone-id",
					Name:        "test-cache-rule",
					Description: stringPtr("Test cache rule"),
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
					Priority:    intPtr(1000),
					Cache:       boolPtr(true),
				},
				rule: &cloudflare.RulesetRule{
					Description: "Test cache rule",
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
					Action:      "set_cache_settings",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						Cache: boolPtr(true),
					},
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"OutOfDateDescription": {
			reason: "Should return false when description differs",
			args: args{
				params: &v1alpha1.CacheRuleParameters{
					Zone:        "test-zone-id",
					Name:        "test-cache-rule",
					Description: stringPtr("Updated description"),
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
				},
				rule: &cloudflare.RulesetRule{
					Description: "Original description",
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
					Action:      "set_cache_settings",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"OutOfDateExpression": {
			reason: "Should return false when expression differs",
			args: args{
				params: &v1alpha1.CacheRuleParameters{
					Zone:        "test-zone-id",
					Name:        "test-cache-rule",
					Expression:  "(http.request.uri.path contains \"/css/\")",
					Enabled:     boolPtr(true),
				},
				rule: &cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Enabled:    boolPtr(true),
					Action:     "set_cache_settings",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"OutOfDateEnabled": {
			reason: "Should return false when enabled status differs",
			args: args{
				params: &v1alpha1.CacheRuleParameters{
					Zone:       "test-zone-id",
					Name:       "test-cache-rule",
					Expression: "(http.request.uri.path contains \"/images/\")",
					Enabled:    boolPtr(false),
				},
				rule: &cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Enabled:    boolPtr(true),
					Action:     "set_cache_settings",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"NilEnabled": {
			reason: "Should handle nil enabled values",
			args: args{
				params: &v1alpha1.CacheRuleParameters{
					Zone:       "test-zone-id",
					Name:       "test-cache-rule",
					Expression: "(http.request.uri.path contains \"/images/\")",
					Enabled:    nil,
				},
				rule: &cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Enabled:    nil,
					Action:     "set_cache_settings",
				},
			},
			want: want{
				upToDate: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsCacheRuleUpToDate(tc.args.params, tc.args.rule)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("%s\nIsCacheRuleUpToDate(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestConvertCacheRuleParametersToCloudflare(t *testing.T) {
	type args struct {
		params v1alpha1.CacheRuleParameters
	}

	type want struct {
		rule cloudflare.RulesetRule
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"BasicCacheRule": {
			reason: "Should convert basic cache rule parameters correctly",
			args: args{
				params: v1alpha1.CacheRuleParameters{
					Zone:        "test-zone-id",
					Name:        "test-cache-rule",
					Description: stringPtr("Test cache rule"),
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
					Priority:    intPtr(1000),
					Cache:       boolPtr(true),
				},
			},
			want: want{
				rule: cloudflare.RulesetRule{
					Description: "Test cache rule",
					Expression:  "(http.request.uri.path contains \"/images/\")",
					Enabled:     boolPtr(true),
					Action:      "set_cache_settings",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						Cache: boolPtr(true),
					},
				},
			},
		},
		"ComplexCacheRuleWithTTL": {
			reason: "Should convert cache rule with TTL settings correctly",
			args: args{
				params: v1alpha1.CacheRuleParameters{
					Zone:       "test-zone-id",
					Name:       "test-cache-rule",
					Expression: "(http.request.uri.path contains \"/images/\")",
					EdgeTTL: &v1alpha1.EdgeTTL{
						Mode:    "override_origin",
						Default: intPtr(3600),
					},
					BrowserTTL: &v1alpha1.BrowserTTL{
						Mode:    "override_origin",
						Default: intPtr(1800),
					},
				},
			},
			want: want{
				rule: cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Action:     "set_cache_settings",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						EdgeTTL: &cloudflare.RulesetRuleActionParametersEdgeTTL{
							Mode:    "override_origin",
							Default: uintPtr(3600),
						},
						BrowserTTL: &cloudflare.RulesetRuleActionParametersBrowserTTL{
							Mode:    "override_origin",
							Default: uintPtr(1800),
						},
					},
				},
			},
		},
		"CacheRuleWithServeStale": {
			reason: "Should convert cache rule with serve stale settings correctly",
			args: args{
				params: v1alpha1.CacheRuleParameters{
					Zone:       "test-zone-id",
					Name:       "test-cache-rule",
					Expression: "(http.request.uri.path contains \"/images/\")",
					ServeStale: &v1alpha1.ServeStale{
						DisableStaleWhileUpdating: boolPtr(true),
					},
				},
			},
			want: want{
				rule: cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Action:     "set_cache_settings",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						ServeStale: &cloudflare.RulesetRuleActionParametersServeStale{
							DisableStaleWhileUpdating: boolPtr(true),
						},
					},
				},
			},
		},
		"CacheRuleWithCacheKey": {
			reason: "Should convert cache rule with cache key settings correctly",
			args: args{
				params: v1alpha1.CacheRuleParameters{
					Zone:       "test-zone-id",
					Name:       "test-cache-rule",
					Expression: "(http.request.uri.path contains \"/images/\")",
					CacheKey: &v1alpha1.CacheKey{
						CacheByDeviceType:       boolPtr(true),
						IgnoreQueryStringsOrder: boolPtr(false),
						CustomKey: &v1alpha1.CustomKey{
							Query: &v1alpha1.CustomKeyQuery{
								Include: []string{"param1", "param2"},
							},
						},
					},
				},
			},
			want: want{
				rule: cloudflare.RulesetRule{
					Expression: "(http.request.uri.path contains \"/images/\")",
					Action:     "set_cache_settings",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						CacheKey: &cloudflare.RulesetRuleActionParametersCacheKey{
							CacheByDeviceType:       boolPtr(true),
							IgnoreQueryStringsOrder: boolPtr(false),
							CustomKey: &cloudflare.RulesetRuleActionParametersCustomKey{
								Query: &cloudflare.RulesetRuleActionParametersCustomKeyQuery{
									Include: &cloudflare.RulesetRuleActionParametersCustomKeyList{
										List: []string{"param1", "param2"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertCacheRuleParametersToCloudflare(tc.args.params)
			if diff := cmp.Diff(tc.want.rule, got); diff != "" {
				t.Errorf("%s\nconvertCacheRuleParametersToCloudflare(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestIsCacheRuleNotFound(t *testing.T) {
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
			reason: "Should return true for 404 error",
			args: args{
				err: &cloudflare.Error{
					StatusCode: 404,
					ErrorCodes: []int{10006},
				},
			},
			want: want{
				notFound: true,
			},
		},
		"OtherError": {
			reason: "Should return false for other errors",
			args: args{
				err: &cloudflare.Error{
					StatusCode: 500,
					ErrorCodes: []int{10001},
				},
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
			got := IsCacheRuleNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.notFound, got); diff != "" {
				t.Errorf("%s\nIsCacheRuleNotFound(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}