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

package security

import (
	"github.com/rossigee/provider-cloudflare/apis/security/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/controller/testutils"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestV1Beta1RateLimitCreation tests basic RateLimit v1beta1 creation
func TestV1Beta1RateLimitCreation(t *testing.T) {
	rateLimit := &v1beta1.RateLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-rate-limit",
			Namespace: "security-policies",
		},
		Spec: v1beta1.RateLimitSpec{
			ForProvider: v1beta1.RateLimitParameters{
				Zone:        "test-zone-id",
				Description: testutils.StringPtr("Test v1beta1 rate limit"),
				Threshold:   10,
				Period:      60,
				Action: v1beta1.RateLimitAction{
					Mode: "challenge",
				},
				Match: v1beta1.RateLimitTrafficMatcher{
					Request: &v1beta1.RateLimitMatchRequest{
						URL:     testutils.StringPtr("/api"),
						Methods: []string{"GET", "POST"},
					},
				},
			},
		},
	}

	// Test that the RateLimit implements the managed resource interface
	if rateLimit.GetCondition("Ready").Status == "" {
		// This validates the managed resource interface is implemented
		t.Log("RateLimit v1beta1 successfully implements managed resource interface")
	}

	// Test basic field access
	if rateLimit.Spec.ForProvider.Zone != "test-zone-id" {
		t.Errorf("Expected zone 'test-zone-id', got %s", rateLimit.Spec.ForProvider.Zone)
	}

	if rateLimit.Spec.ForProvider.Action.Mode != "challenge" {
		t.Errorf("Expected action mode 'challenge', got %s", rateLimit.Spec.ForProvider.Action.Mode)
	}

	if rateLimit.Spec.ForProvider.Threshold != 10 {
		t.Errorf("Expected threshold 10, got %d", rateLimit.Spec.ForProvider.Threshold)
	}

	if rateLimit.Spec.ForProvider.Period != 60 {
		t.Errorf("Expected period 60, got %d", rateLimit.Spec.ForProvider.Period)
	}

	// Test URL match
	if rateLimit.Spec.ForProvider.Match.Request == nil || rateLimit.Spec.ForProvider.Match.Request.URL == nil {
		t.Error("Expected URL match to be set")
	} else if *rateLimit.Spec.ForProvider.Match.Request.URL != "/api" {
		t.Errorf("Expected URL '/api', got %s", *rateLimit.Spec.ForProvider.Match.Request.URL)
	}

	// Test namespace scope
	if rateLimit.Namespace != "security-policies" {
		t.Errorf("Expected namespace 'security-policies', got %s", rateLimit.Namespace)
	}

	t.Log("v1beta1 RateLimit creation and field access tests passed")
}

// TestV1Beta1RateLimitAdvancedFeatures tests advanced v1beta1 features
func TestV1Beta1RateLimitAdvancedFeatures(t *testing.T) {
	rateLimit := &v1beta1.RateLimit{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "advanced-rate-limit",
			Namespace: "production-security",
		},
		Spec: v1beta1.RateLimitSpec{
			ForProvider: v1beta1.RateLimitParameters{
				Zone:        "production-zone-id",
				Description: testutils.StringPtr("Advanced production rate limit"),
				Threshold:   100,
				Period:      300,
				Action: v1beta1.RateLimitAction{
					Mode: "ban",
				},
				Match: v1beta1.RateLimitTrafficMatcher{
					Request: &v1beta1.RateLimitMatchRequest{
						URL:     testutils.StringPtr("/api/v2/"),
						Methods: []string{"POST", "PUT", "DELETE"},
					},
				},
			},
		},
	}

	// Test action mode
	if rateLimit.Spec.ForProvider.Action.Mode != "ban" {
		t.Errorf("Expected action mode 'ban', got %s", rateLimit.Spec.ForProvider.Action.Mode)
	}

	// Test URL match
	if rateLimit.Spec.ForProvider.Match.Request == nil || rateLimit.Spec.ForProvider.Match.Request.URL == nil {
		t.Error("Expected URL match to be set")
	} else if *rateLimit.Spec.ForProvider.Match.Request.URL != "/api/v2/" {
		t.Errorf("Expected URL '/api/v2/', got %s", *rateLimit.Spec.ForProvider.Match.Request.URL)
	}

	// Test methods
	if len(rateLimit.Spec.ForProvider.Match.Request.Methods) != 3 {
		t.Errorf("Expected 3 methods, got %d", len(rateLimit.Spec.ForProvider.Match.Request.Methods))
	}

	t.Log("v1beta1 RateLimit advanced features tests passed")
}
