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

package cache

import (
	"testing"

	"github.com/rossigee/provider-cloudflare/apis/cache/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/controller/testutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestV1Beta1CacheRuleCreation tests basic CacheRule v1beta1 creation
func TestV1Beta1CacheRuleCreation(t *testing.T) {
	cacheRule := &v1beta1.CacheRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cache-rule",
			Namespace: "test-namespace",
		},
		Spec: v1beta1.CacheRuleSpec{
			ForProvider: v1beta1.CacheRuleParameters{
				Zone:        "test-zone-id",
				Name:        "test-cache-rule",
				Description: testutils.StringPtr("Test v1beta1 cache rule"),
				Expression:  "http.request.uri.path contains \"/api/\"",
				Enabled:     testutils.BoolPtr(true),
				ActionParameters: &v1beta1.CacheRuleActionParameters{
					Cache: testutils.BoolPtr(true),
					EdgeTTL: &v1beta1.EdgeTTL{
						Mode:    testutils.StringPtr("override_origin"),
						Default: testutils.Int64Ptr(3600),
					},
					BrowserTTL: &v1beta1.BrowserTTL{
						Mode:    testutils.StringPtr("override_origin"),
						Default: testutils.Int64Ptr(1800),
					},
				},
			},
		},
	}

	// Test that the CacheRule implements the managed resource interface
	if cacheRule.GetCondition("Ready").Status == "" {
		// This validates the managed resource interface is implemented
		t.Log("CacheRule v1beta1 successfully implements managed resource interface")
	}

	// Test basic field access
	if cacheRule.Spec.ForProvider.Zone != "test-zone-id" {
		t.Errorf("Expected zone 'test-zone-id', got %s", cacheRule.Spec.ForProvider.Zone)
	}

	if *cacheRule.Spec.ForProvider.ActionParameters.EdgeTTL.Mode != "override_origin" {
		t.Errorf("Expected EdgeTTL mode 'override_origin', got %s", *cacheRule.Spec.ForProvider.ActionParameters.EdgeTTL.Mode)
	}

	if *cacheRule.Spec.ForProvider.ActionParameters.EdgeTTL.Default != int64(3600) {
		t.Errorf("Expected EdgeTTL default 3600, got %d", *cacheRule.Spec.ForProvider.ActionParameters.EdgeTTL.Default)
	}

	// Test namespace scope
	if cacheRule.Namespace != "test-namespace" {
		t.Errorf("Expected namespace 'test-namespace', got %s", cacheRule.Namespace)
	}

	t.Log("v1beta1 CacheRule creation and field access tests passed")
}

// TestV1Beta1CacheRuleAdvancedFeatures tests advanced v1beta1 features
func TestV1Beta1CacheRuleAdvancedFeatures(t *testing.T) {
	cacheRule := &v1beta1.CacheRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "advanced-cache-rule",
			Namespace: "production",
		},
		Spec: v1beta1.CacheRuleSpec{
			ForProvider: v1beta1.CacheRuleParameters{
				Zone:        "advanced-zone-id",
				Name:        "advanced-cache-rule",
				Description: testutils.StringPtr("Advanced cache rule with complex expression"),
				Expression:  "http.request.uri.path matches \"^/api/v2/\" and http.request.method eq \"GET\"",
				Enabled:     testutils.BoolPtr(true),
				Action:      testutils.StringPtr("set_cache_settings"),
				ActionParameters: &v1beta1.CacheRuleActionParameters{
					Cache: testutils.BoolPtr(true),
					EdgeTTL: &v1beta1.EdgeTTL{
						Mode:    testutils.StringPtr("override_origin"),
						Default: testutils.Int64Ptr(7200),
					},
				},
			},
		},
	}

	// Test advanced expression
	if !testutils.ContainsString(cacheRule.Spec.ForProvider.Expression, "http.request.method eq \"GET\"") {
		t.Error("Expected expression to include method check")
	}

	// Test action parameters
	if cacheRule.Spec.ForProvider.ActionParameters == nil {
		t.Error("Expected ActionParameters to be set")
	}

	if !*cacheRule.Spec.ForProvider.ActionParameters.Cache {
		t.Error("Expected Cache to be true")
	}

	// Test namespace scope
	if cacheRule.Namespace != "production" {
		t.Errorf("Expected namespace 'production', got %s", cacheRule.Namespace)
	}

	t.Log("v1beta1 CacheRule advanced features tests passed")
}
