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

package tunnel

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/tunnel/v1beta1"
)

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params v1beta1.TunnelParameters
		obs    v1beta1.TunnelObservation
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
				params: v1beta1.TunnelParameters{
					AccountID: "test-account-id",
					Name:      "test-tunnel",
					Secret:    "dGVzdC1zZWNyZXQ=",
					ConfigSrc: &v1beta1.TunnelConfigSrc{},
				},
				obs: v1beta1.TunnelObservation{
					Name: "test-tunnel",
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"OutOfDateName": {
			reason: "Should return false when name differs",
			args: args{
				params: v1beta1.TunnelParameters{
					AccountID: "test-account-id",
					Name:      "updated-tunnel",
					Secret:    "dGVzdC1zZWNyZXQ=",
				},
				obs: v1beta1.TunnelObservation{
					Name: "original-tunnel",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &CloudflareTunnelClient{}
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
		"StringNotFoundDirect": {
			reason: "Should return true for string containing 'not found'",
			args: args{
				err: errors.New("tunnel not found"),
			},
			want: want{
				notFound: true,
			},
		},
		"StringResourceNotFound": {
			reason: "Should return true for string containing 'resource not found'",
			args: args{
				err: errors.New("resource not found"),
			},
			want: want{
				notFound: true,
			},
		},
		"StringDoesNotExist": {
			reason: "Should return true for string containing 'does not exist'",
			args: args{
				err: errors.New("tunnel does not exist"),
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
