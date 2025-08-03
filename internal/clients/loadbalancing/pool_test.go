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

func TestGeneratePoolObservation(t *testing.T) {
	createdOn := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	modifiedOn := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	pool := &cloudflare.LoadBalancerPool{
		ID:         "test-pool-id",
		CreatedOn:  &createdOn,
		ModifiedOn: &modifiedOn,
		Healthy:    boolPtr(true),
	}

	expected := v1alpha1.LoadBalancerPoolObservation{
		ID:         "test-pool-id",
		CreatedOn:  stringPtr("2025-01-01 00:00:00 +0000 UTC"),
		ModifiedOn: stringPtr("2025-01-02 00:00:00 +0000 UTC"),
		Healthy:    boolPtr(true),
	}

	result := GeneratePoolObservation(pool)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("GeneratePoolObservation(...): -want, +got:\n%s", diff)
	}
}

func TestIsPoolUpToDate(t *testing.T) {
	type args struct {
		params *v1alpha1.LoadBalancerPoolParameters
		pool   *cloudflare.LoadBalancerPool
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
				params: &v1alpha1.LoadBalancerPoolParameters{
					Name:        "test-pool",
					Description: stringPtr("Test pool"),
					Enabled:     boolPtr(true),
					Origins: []v1alpha1.LoadBalancerOrigin{
						{
							Name:    "origin1",
							Address: "192.168.1.1",
							Enabled: boolPtr(true),
							Weight:  stringPtr("1.0"),
						},
					},
				},
				pool: &cloudflare.LoadBalancerPool{
					Name:        "test-pool",
					Description: "Test pool",
					Enabled:     true,
					Origins: []cloudflare.LoadBalancerOrigin{
						{
							Name:    "origin1",
							Address: "192.168.1.1",
							Enabled: true,
							Weight:  1.0,
						},
					},
				},
			},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateDifferentName": {
			reason: "Should return false when names differ",
			args: args{
				params: &v1alpha1.LoadBalancerPoolParameters{
					Name: "test-pool-1",
				},
				pool: &cloudflare.LoadBalancerPool{
					Name: "test-pool-2",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentDescription": {
			reason: "Should return false when descriptions differ",
			args: args{
				params: &v1alpha1.LoadBalancerPoolParameters{
					Name:        "test-pool",
					Description: stringPtr("Description 1"),
				},
				pool: &cloudflare.LoadBalancerPool{
					Name:        "test-pool",
					Description: "Description 2",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateNilDescription": {
			reason: "Should return false when param description is nil but pool has description",
			args: args{
				params: &v1alpha1.LoadBalancerPoolParameters{
					Name:        "test-pool",
					Description: nil,
				},
				pool: &cloudflare.LoadBalancerPool{
					Name:        "test-pool",
					Description: "Some description",
				},
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentOriginCount": {
			reason: "Should return false when origin counts differ",
			args: args{
				params: &v1alpha1.LoadBalancerPoolParameters{
					Name: "test-pool",
					Origins: []v1alpha1.LoadBalancerOrigin{
						{Name: "origin1", Address: "192.168.1.1"},
						{Name: "origin2", Address: "192.168.1.2"},
					},
				},
				pool: &cloudflare.LoadBalancerPool{
					Name: "test-pool",
					Origins: []cloudflare.LoadBalancerOrigin{
						{Name: "origin1", Address: "192.168.1.1"},
					},
				},
			},
			want: want{
				upToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsPoolUpToDate(tc.args.params, tc.args.pool)
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsPoolUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertOriginsToCloudflare(t *testing.T) {
	origins := []v1alpha1.LoadBalancerOrigin{
		{
			Name:              "origin1",
			Address:           "192.168.1.1",
			Enabled:           boolPtr(true),
			Weight:            stringPtr("0.5"),
			Header:            map[string][]string{"Host": {"example.com"}},
			VirtualNetworkID:  stringPtr("vnet-123"),
		},
		{
			Name:    "origin2",
			Address: "192.168.1.2",
			// Enabled is nil, should default to true
		},
	}

	expected := []cloudflare.LoadBalancerOrigin{
		{
			Name:             "origin1",
			Address:          "192.168.1.1",
			Enabled:          true,
			Weight:           0.5,
			Header:           map[string][]string{"Host": {"example.com"}},
			VirtualNetworkID: "vnet-123",
		},
		{
			Name:    "origin2",
			Address: "192.168.1.2",
			Enabled: true, // defaulted
		},
	}

	result := convertOriginsToCloudflare(origins)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertOriginsToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertLoadSheddingToCloudflare(t *testing.T) {
	loadShedding := v1alpha1.LoadBalancerLoadShedding{
		DefaultPercent:  stringPtr("0.1"),
		DefaultPolicy:   stringPtr("random"),
		SessionPercent:  stringPtr("0.2"),
		SessionPolicy:   stringPtr("hash"),
	}

	expected := &cloudflare.LoadBalancerLoadShedding{
		DefaultPercent:  0.1,
		DefaultPolicy:   "random",
		SessionPercent:  0.2,
		SessionPolicy:   "hash",
	}

	result := convertLoadSheddingToCloudflare(loadShedding)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertLoadSheddingToCloudflare(...): -want, +got:\n%s", diff)
	}
}

func TestConvertOriginSteeringToCloudflare(t *testing.T) {
	steering := v1alpha1.LoadBalancerOriginSteering{
		Policy: stringPtr("random"),
	}

	expected := &cloudflare.LoadBalancerOriginSteering{
		Policy: "random",
	}

	result := convertOriginSteeringToCloudflare(steering)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("convertOriginSteeringToCloudflare(...): -want, +got:\n%s", diff)
	}
}