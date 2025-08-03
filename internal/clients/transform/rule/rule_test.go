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
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"

	"github.com/rossigee/provider-cloudflare/apis/transform/v1alpha1"
)

func TestUpToDate(t *testing.T) {
	type args struct {
		spec *v1alpha1.RuleParameters
		rule cloudflare.RulesetRule
	}

	type want struct {
		upToDate bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"UpToDateSpecNil": {
			reason: "UpToDate should return true when not passed a spec",
			args:   args{},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateExpressionDifferent": {
			reason: "UpToDate should return false if the expression differs",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression: `http.request.uri.path eq "/old"`,
					Action:     "rewrite",
				},
				rule: cloudflare.RulesetRule{
					Expression: `http.request.uri.path eq "/new"`,
					Action:     "rewrite",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateActionDifferent": {
			reason: "UpToDate should return false if the action differs",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
				},
				rule: cloudflare.RulesetRule{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "redirect",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDescriptionDifferent": {
			reason: "UpToDate should return false if the description differs",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression:  `http.request.uri.path eq "/test"`,
					Action:      "rewrite",
					Description: ptr.To("Old description"),
				},
				rule: cloudflare.RulesetRule{
					Expression:  `http.request.uri.path eq "/test"`,
					Action:      "rewrite",
					Description: "New description",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateEnabledDifferent": {
			reason: "UpToDate should return false if the enabled status differs",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
					Enabled:    ptr.To(true),
				},
				rule: cloudflare.RulesetRule{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
					Enabled:    ptr.To(false),
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateIdentical": {
			reason: "UpToDate should return true if all fields match",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression:  `http.request.uri.path eq "/test"`,
					Action:      "rewrite",
					Description: ptr.To("Test rule"),
					Enabled:     ptr.To(true),
				},
				rule: cloudflare.RulesetRule{
					Expression:  `http.request.uri.path eq "/test"`,
					Action:      "rewrite",
					Description: "Test rule",
					Enabled:     ptr.To(true),
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateActionParametersNilSpec": {
			reason: "UpToDate should return false if spec has no action parameters but rule does",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
				},
				rule: cloudflare.RulesetRule{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
					ActionParameters: &cloudflare.RulesetRuleActionParameters{
						URI: &cloudflare.RulesetRuleActionParametersURI{},
					},
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateActionParametersNilRule": {
			reason: "UpToDate should return false if rule has no action parameters but spec does",
			args: args{
				spec: &v1alpha1.RuleParameters{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
					ActionParameters: &v1alpha1.RuleActionParameters{
						URI: &v1alpha1.URITransform{},
					},
				},
				rule: cloudflare.RulesetRule{
					Expression: `http.request.uri.path eq "/test"`,
					Action:     "rewrite",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := UpToDate(tc.args.spec, tc.args.rule)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	rule := cloudflare.RulesetRule{
		ID:          "test-rule-id",
		Version:     ptr.To("1"),
		Expression:  `http.request.uri.path eq "/test"`,
		Action:      "rewrite",
		Description: "Test rule",
	}

	rulesetID := "test-ruleset-id"

	obs := GenerateObservation(rule, rulesetID)

	expectedObs := v1alpha1.RuleObservation{
		ID:        "test-rule-id",
		RulesetID: "test-ruleset-id",
		Version:   "1",
	}

	if diff := cmp.Diff(expectedObs.ID, obs.ID); diff != "" {
		t.Errorf("GenerateObservation() ID: -want, +got:\n%s\n", diff)
	}
	if diff := cmp.Diff(expectedObs.RulesetID, obs.RulesetID); diff != "" {
		t.Errorf("GenerateObservation() RulesetID: -want, +got:\n%s\n", diff)
	}
	if diff := cmp.Diff(expectedObs.Version, obs.Version); diff != "" {
		t.Errorf("GenerateObservation() Version: -want, +got:\n%s\n", diff)
	}
}

func TestIsRulesetNotFound(t *testing.T) {
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
		"NotFoundError": {
			reason: "IsRulesetNotFound should return true for ruleset not found errors",
			args: args{
				err: errors.New("API Error 10007: Ruleset not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "IsRulesetNotFound should return false for other errors",
			args: args{
				err: errors.New("API Error 10001: Invalid request"),
			},
			want: want{
				isNotFound: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRulesetNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.isNotFound, got); diff != "" {
				t.Errorf("\n%s\nIsRulesetNotFound(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"NotFoundError": {
			reason: "IsRuleNotFound should return true for rule not found errors",
			args: args{
				err: errors.New("API Error 10014: Rule not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "IsRuleNotFound should return false for other errors",
			args: args{
				err: errors.New("API Error 10001: Invalid request"),
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