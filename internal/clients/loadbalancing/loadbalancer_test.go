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

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func TestGenerateLoadBalancerObservation(t *testing.T) {
	createdOn := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedOn := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	lb := &cloudflare.LoadBalancer{
		ID:         "test-lb-id",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
	}

	expected := v1alpha1.LoadBalancerObservation{
		ID:         "test-lb-id",
		CreatedOn:  stringPtr("2025-01-01 00:00:00 +0000 UTC"),
		ModifiedOn: stringPtr("2025-01-02 00:00:00 +0000 UTC"),
	}

	result := GenerateLoadBalancerObservation(lb)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("GenerateLoadBalancerObservation(...): -want, +got:\n%s", diff)
	}
}

func TestIsLoadBalancerUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.LoadBalancerParameters
		lb     *cloudflare.LoadBalancer
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
				params: &v1alpha1.LoadBalancerParameters{
					Zone:            "example.com",
					Name:            stringPtr("test-lb"),
					Description:     stringPtr("Test load balancer"),
					TTL:             intPtr(300),
					Proxied:         boolPtr(true),
					Enabled:         boolPtr(true),
					SteeringPolicy:  stringPtr("geo"),
					DefaultPools:    []string{"pool1", "pool2"},
					SessionAffinity: stringPtr("cookie"),
				},
				lb: &cloudflare.LoadBalancer{
					Name:         "test-lb",
					Description:  "Test load balancer",
					TTL:          300,
					Proxied:      true,
					Enabled:      boolPtr(true),
					Persistence:  "cookie",
					DefaultPools: []string{"pool1", "pool2"},
					SteeringPolicy: "geo",
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateDifferentName": {
			reason: "Should return false when names differ",
			args: args{
				params: &v1alpha1.LoadBalancerParameters{
					Zone: "example.com",
					Name: stringPtr("test-lb-1"),
				},
				lb: &cloudflare.LoadBalancer{
					Name: "test-lb-2",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentDescription": {
			reason: "Should return false when descriptions differ",
			args: args{
				params: &v1alpha1.LoadBalancerParameters{
					Zone:        "example.com",
					Description: stringPtr("Description 1"),
				},
				lb: &cloudflare.LoadBalancer{
					Description: "Description 2",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentDefaultPools": {
			reason: "Should return false when default pools count differs",
			args: args{
				params: &v1alpha1.LoadBalancerParameters{
					Zone:         "example.com",
					DefaultPools: []string{"pool1", "pool2"},
				},
				lb: &cloudflare.LoadBalancer{
					DefaultPools: []string{"pool1"},
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateNilDescription": {
			reason: "Should return false when param description is nil but LB has description",
			args: args{
				params: &v1alpha1.LoadBalancerParameters{
					Zone:        "example.com",
					Description: nil,
				},
				lb: &cloudflare.LoadBalancer{
					Description: "Some description",
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsLoadBalancerUpToDate(tc.args.params, tc.args.lb)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsLoadBalancerUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertSessionAffinityAttributesToCloudflare(t *testing.T) {
	attrs := v1alpha1.SessionAffinityAttributes{
		SameSite:              stringPtr("lax"),
		Secure:                stringPtr("auto"),
		DrainDuration:         intPtr(100),
		ZeroDowntimeFailover:  stringPtr("temporary"),
		Headers:               []string{"CF-IPCountry"},
		RequireAllHeaders:     boolPtr(true),
	}

	expected := &cloudflare.SessionAffinityAttributes{
		SameSite:              "lax",
		Secure:                "auto",
		DrainDuration:         100,
		ZeroDowntimeFailover:  "temporary",
		Headers:               []string{"CF-IPCountry"},
		RequireAllHeaders:     true,
	}

	result := convertSessionAffinityAttributesToCloudflare(attrs)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertSessionAffinityAttributesToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertRandomSteeringToCloudflare(t *testing.T) {
	steering := v1alpha1.RandomSteering{
		DefaultWeight: stringPtr("0.5"),
		PoolWeights: map[string]string{
			"pool1": "0.3",
			"pool2": "0.7",
		},
	}

	expected := &cloudflare.RandomSteering{
		DefaultWeight: 0.5,
		PoolWeights: map[string]float64{
			"pool1": 0.3,
			"pool2": 0.7,
		},
	}

	result := convertRandomSteeringToCloudflare(steering)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertRandomSteeringToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertAdaptiveRoutingToCloudflare(t *testing.T) {
	routing := v1alpha1.AdaptiveRouting{
		FailoverAcrossPools: boolPtr(true),
	}

	expected := &cloudflare.AdaptiveRouting{
		FailoverAcrossPools: boolPtr(true),
	}

	result := convertAdaptiveRoutingToCloudflare(routing)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertAdaptiveRoutingToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertLocationStrategyToCloudflare(t *testing.T) {
	strategy := v1alpha1.LocationStrategy{
		Mode:             stringPtr("resolver_ip"),
		PreferECSRegion:  stringPtr("closest"),
	}

	expected := &cloudflare.LocationStrategy{
		Mode:      "resolver_ip",
		PreferECS: "closest",
	}

	result := convertLocationStrategyToCloudflare(strategy)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertLocationStrategyToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertRulesToCloudflare(t *testing.T) {
	rules := []v1alpha1.LoadBalancerRule{
		{
			Name:       "test-rule",
			Condition:  "http.request.uri.path contains \"/api\"",
			Priority:   1,
			Disabled:   boolPtr(false),
			Terminates: boolPtr(true),
			FixedResponse: &v1alpha1.LoadBalancerFixedResponse{
				MessageBody: stringPtr("API unavailable"),
				StatusCode:  intPtr(503),
				ContentType: stringPtr("text/plain"),
				Location:    stringPtr("https://example.com/maintenance"),
			},
		},
	}

	expected := []*cloudflare.LoadBalancerRule{
		{
			Name:       "test-rule",
			Condition:  "http.request.uri.path contains \"/api\"",
			Priority:   1,
			Disabled:   false,
			Terminates: true,
			FixedResponse: &cloudflare.LoadBalancerFixedResponseData{
				MessageBody: "API unavailable",
				StatusCode:  503,
				ContentType: "text/plain",
				Location:    "https://example.com/maintenance",
			},
		},
	}

	result := convertRulesToCloudflare(rules)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertRulesToCloudflare(...): -want, +got:\n%s", diff)
	}
}