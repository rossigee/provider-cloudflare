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

package loadbalancing

import (
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
)

func TestGenerateMonitorObservation(t *testing.T) {
	createdOn := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedOn := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	monitor := &cloudflare.LoadBalancerMonitor{
		ID:         "test-monitor-id",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}

	expected := v1alpha1.LoadBalancerMonitorObservation{
		ID:         "test-monitor-id",
		CreatedOn:  stringPtr("2025-01-01 00:00:00 +0000 UTC"),
		ModifiedOn: stringPtr("2025-01-02 00:00:00 +0000 UTC"),
	}

	result := GenerateMonitorObservation(monitor)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("GenerateMonitorObservation(...): -want, +got:\n%s", diff)
	}
}

func TestIsMonitorUpToDate(t *testing.T) {
	type args struct {
		params  *v1alpha1.LoadBalancerMonitorParameters
		monitor *cloudflare.LoadBalancerMonitor
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
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:             "http",
					Description:      stringPtr("Test monitor"),
					Method:           stringPtr("GET"),
					Path:             stringPtr("/health"),
					Timeout:          intPtr(10),
					Retries:          intPtr(3),
					Interval:         intPtr(60),
					ConsecutiveUp:    intPtr(2),
					ConsecutiveDown:  intPtr(3),
					Port:             intPtr(80),
					ExpectedBody:     stringPtr("OK"),
					ExpectedCodes:    stringPtr("200"),
					FollowRedirects:  boolPtr(false),
					AllowInsecure:    boolPtr(false),
					ProbeZone:        stringPtr("example.com"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:            "http",
					Description:     "Test monitor",
					Method:          "GET",
					Path:            "/health",
					Timeout:         10,
					Retries:         3,
					Interval:        60,
					ConsecutiveUp:   2,
					ConsecutiveDown: 3,
					Port:            80,
					ExpectedBody:    "OK",
					ExpectedCodes:   "200",
					FollowRedirects: false,
					AllowInsecure:   false,
					ProbeZone:       "example.com",
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateDifferentType": {
			reason: "Should return false when types differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type: "http",
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type: "tcp",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentDescription": {
			reason: "Should return false when descriptions differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:        "http",
					Description: stringPtr("Description 1"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:        "http",
					Description: "Description 2",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateNilDescription": {
			reason: "Should return false when param description is nil but monitor has description",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:        "http",
					Description: nil,
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:        "http",
					Description: "Some description",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentMethod": {
			reason: "Should return false when methods differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:   "http",
					Method: stringPtr("GET"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:   "http",
					Method: "POST",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentPath": {
			reason: "Should return false when paths differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type: "http",
					Path: stringPtr("/health"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type: "http",
					Path: "/status",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentTimeout": {
			reason: "Should return false when timeouts differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:    "http",
					Timeout: intPtr(10),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:    "http",
					Timeout: 20,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentRetries": {
			reason: "Should return false when retries differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:    "http",
					Retries: intPtr(3),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:    "http",
					Retries: 5,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentInterval": {
			reason: "Should return false when intervals differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:     "http",
					Interval: intPtr(60),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:     "http",
					Interval: 120,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentConsecutiveUp": {
			reason: "Should return false when consecutive up values differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:          "http",
					ConsecutiveUp: intPtr(2),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:          "http",
					ConsecutiveUp: 3,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentConsecutiveDown": {
			reason: "Should return false when consecutive down values differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:            "http",
					ConsecutiveDown: intPtr(3),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:            "http",
					ConsecutiveDown: 5,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentPort": {
			reason: "Should return false when ports differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type: "http",
					Port: intPtr(80),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type: "http",
					Port: 443,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentExpectedBody": {
			reason: "Should return false when expected bodies differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:         "http",
					ExpectedBody: stringPtr("OK"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:         "http",
					ExpectedBody: "Healthy",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentExpectedCodes": {
			reason: "Should return false when expected codes differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:          "http",
					ExpectedCodes: stringPtr("200"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:          "http",
					ExpectedCodes: "200,201",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentFollowRedirects": {
			reason: "Should return false when follow redirects differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:            "http",
					FollowRedirects: boolPtr(true),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:            "http",
					FollowRedirects: false,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentAllowInsecure": {
			reason: "Should return false when allow insecure differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:          "http",
					AllowInsecure: boolPtr(true),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:          "http",
					AllowInsecure: false,
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentProbeZone": {
			reason: "Should return false when probe zones differ",
			args: args{
				params: &v1alpha1.LoadBalancerMonitorParameters{
					Type:      "http",
					ProbeZone: stringPtr("example.com"),
				},
				monitor: &cloudflare.LoadBalancerMonitor{
					Type:      "http",
					ProbeZone: "test.com",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsMonitorUpToDate(tc.args.params, tc.args.monitor)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsMonitorUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}