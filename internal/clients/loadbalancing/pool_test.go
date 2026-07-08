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
	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
	"testing"
	"time"
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

	expected := v1beta1.LoadBalancerPoolObservation{
		ID:         "test-pool-id",
		CreatedOn:  stringPtr("2025-01-01 00:00:00 +0000 UTC"),
		ModifiedOn: stringPtr("2025-01-02 00:00:00 +0000 UTC"),
		Healthy:    nil, // Healthy status might not be directly available
	}

	result := GeneratePoolObservation(pool)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("GeneratePoolObservation(...): -want, +got:\n%s", diff)
	}
}
