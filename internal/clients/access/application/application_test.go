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

package application

import (
	"context"
	"errors"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/access/v1beta1"
)

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params v1beta1.AccessApplicationParameters
		obs    v1beta1.AccessApplicationObservation
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
				params: v1beta1.AccessApplicationParameters{
					AccountID:        "test-account-id",
					Name:             "test-app",
					Domain:           "test.example.com",
					Type:             "self_hosted",
					SessionDuration:  stringPtr("24h"),
					SkipInterstitial: boolPtr(true),
				},
				obs: v1beta1.AccessApplicationObservation{
					Name:             "test-app",
					Domain:           "test.example.com",
					Type:             "self_hosted",
					SessionDuration:  "24h",
					SkipInterstitial: true,
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"OutOfDateName": {
			reason: "Should return false when name differs",
			args: args{
				params: v1beta1.AccessApplicationParameters{
					AccountID: "test-account-id",
					Name:      "updated-app",
					Domain:    "test.example.com",
					Type:      "self_hosted",
				},
				obs: v1beta1.AccessApplicationObservation{
					Name:   "original-app",
					Domain: "test.example.com",
					Type:   "self_hosted",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &CloudflareAccessApplicationClient{}
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

func TestConvertParametersToCreateAccessApplication(t *testing.T) {
	type args struct {
		params v1beta1.AccessApplicationParameters
	}

	cases := map[string]struct {
		reason string
		args   args
	}{
		"BasicApplication": {
			reason: "Should convert basic parameters correctly",
			args: args{
				params: v1beta1.AccessApplicationParameters{
					AccountID:        "test-account-id",
					Name:             "test-app",
					Domain:           "test.example.com",
					Type:             "self_hosted",
					SessionDuration:  stringPtr("24h"),
					AllowedIdps:      []string{"idp-1", "idp-2"},
					SkipInterstitial: boolPtr(true),
				},
			},
		},
		"MinimalApplication": {
			reason: "Should convert minimal parameters correctly",
			args: args{
				params: v1beta1.AccessApplicationParameters{
					AccountID: "test-account-id",
					Name:      "test-app",
					Domain:    "test.example.com",
					Type:      "self_hosted",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToCreateAccessApplication(tc.args.params)
			if got.Name != tc.args.params.Name || got.Domain != tc.args.params.Domain {
				t.Errorf("%s\nconvertParametersToCreateAccessApplication(...): failed to convert basic fields", tc.reason)
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
				err: errors.New("access application not found"),
			},
			want: want{
				notFound: true,
			},
		},
		"StringNotFound": {
			reason: "Should return true for string containing 'not found'",
			args: args{
				err: errors.New("application not found"),
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

func TestGetResourceContainer(t *testing.T) {
	type args struct {
		params v1beta1.AccessApplicationParameters
	}

	type want struct {
		rc *cloudflare.ResourceContainer
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"AccountLevel": {
			reason: "Should create account level resource container",
			args: args{
				params: v1beta1.AccessApplicationParameters{
					AccountID: "test-account-id",
				},
			},
			want: want{
				rc: &cloudflare.ResourceContainer{
					Level:      cloudflare.AccountRouteLevel,
					Identifier: "test-account-id",
				},
			},
		},
		"ZoneLevel": {
			reason: "Should create zone level resource container when zone ID is provided",
			args: args{
				params: v1beta1.AccessApplicationParameters{
					AccountID: "test-account-id",
					ZoneID:    stringPtr("test-zone-id"),
				},
			},
			want: want{
				rc: &cloudflare.ResourceContainer{
					Level:      cloudflare.ZoneRouteLevel,
					Identifier: "test-zone-id",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := getResourceContainer(tc.args.params)
			if diff := cmp.Diff(tc.want.rc, got); diff != "" {
				t.Errorf("%s\ngetResourceContainer(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
