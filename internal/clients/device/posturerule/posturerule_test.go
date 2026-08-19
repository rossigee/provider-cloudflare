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
)

func stringPtr(s string) *string {
	return &s
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
