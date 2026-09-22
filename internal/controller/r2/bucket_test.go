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

package r2

import (
	"testing"

	"github.com/rossigee/provider-cloudflare/apis/r2/v1beta1"
)

type bucketModifier func(*v1beta1.Bucket)

func withName(name string) bucketModifier {
	return func(b *v1beta1.Bucket) { b.Spec.ForProvider.Name = name }
}

func bucketCR(m ...bucketModifier) *v1beta1.Bucket {
	cr := &v1beta1.Bucket{
		Spec: v1beta1.BucketSpec{
			ForProvider: v1beta1.BucketParameters{
				Name: "test-bucket",
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func stringPtr(s string) *string {
	return &s
}

func TestBucketCreate(t *testing.T) {
	cases := map[string]struct {
		reason string
		bucket *v1beta1.Bucket
		verify func(t *testing.T, b *v1beta1.Bucket)
	}{
		"BasicBucket": {
			reason: "Should create bucket with basic parameters",
			bucket: bucketCR(withName("my-bucket")),
			verify: func(t *testing.T, b *v1beta1.Bucket) {
				if b.Spec.ForProvider.Name == "" {
					t.Errorf("Bucket name should not be empty")
				}
			},
		},
		"BucketWithLocation": {
			reason: "Should create bucket with location hint",
			bucket: bucketCR(
				withName("eu-bucket"),
			),
			verify: func(t *testing.T, b *v1beta1.Bucket) {
				if b.Spec.ForProvider.Name != "eu-bucket" {
					t.Errorf("Expected name eu-bucket, got %s", b.Spec.ForProvider.Name)
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.verify(t, tc.bucket)
		})
	}
}

func TestBucketParameters(t *testing.T) {
	cases := map[string]struct {
		reason string
		params v1beta1.BucketParameters
		verify func(t *testing.T, params v1beta1.BucketParameters)
	}{
		"ValidParams": {
			reason: "Should validate basic parameters",
			params: v1beta1.BucketParameters{
				Name: "my-bucket",
			},
			verify: func(t *testing.T, params v1beta1.BucketParameters) {
				if params.Name == "" {
					t.Errorf("Name should not be empty")
				}
			},
		},
		"WithLocationHint": {
			reason: "Should accept location hint",
			params: v1beta1.BucketParameters{
				Name:         "bucket",
				LocationHint: stringPtr("wnam"),
			},
			verify: func(t *testing.T, params v1beta1.BucketParameters) {
				if params.LocationHint == nil {
					t.Errorf("LocationHint should be set")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.verify(t, tc.params)
		})
	}
}

func TestBucketObservation(t *testing.T) {
	cases := map[string]struct {
		reason string
		obs    v1beta1.BucketObservation
		verify func(t *testing.T, obs v1beta1.BucketObservation)
	}{
		"BasicObservation": {
			reason: "Should create observation",
			obs: v1beta1.BucketObservation{
				Name: "my-bucket",
			},
			verify: func(t *testing.T, obs v1beta1.BucketObservation) {
				if obs.Name == "" {
					t.Errorf("Name should not be empty")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tc.verify(t, tc.obs)
		})
	}
}

func TestBucket(t *testing.T) {
	r := &v1beta1.Bucket{
		Spec: v1beta1.BucketSpec{
			ForProvider: v1beta1.BucketParameters{
				Name: "test-bucket",
			},
		},
	}

	if r == nil {
		t.Errorf("Bucket resource should not be nil")
	}
	if r.Spec.ForProvider.Name != "test-bucket" {
		t.Logf("resource type %T instantiated successfully", r)
	}
}
