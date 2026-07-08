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

package workers

import (
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/controller/testutils"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestV1Beta1ScriptCreation tests basic Worker Script v1beta1 creation
func TestV1Beta1ScriptCreation(t *testing.T) {
	script := &v1beta1.Script{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-worker-script",
			Namespace: "worker-scripts",
		},
		Spec: v1beta1.ScriptSpec{
			ForProvider: v1beta1.ScriptParameters{
				ScriptName:        "my-worker",
				Script:            "addEventListener('fetch', event => { event.respondWith(new Response('Hello World!')) })",
				Module:            testutils.BoolPtr(false),
				CompatibilityDate: testutils.StringPtr("2023-01-01"),
				Bindings: []v1beta1.WorkerBinding{
					{
						Type:        "kv_namespace",
						Name:        "MY_KV",
						NamespaceID: testutils.StringPtr("namespace-123"),
					},
				},
				Tags: []string{"production", "api"},
			},
		},
	}

	// Test that the Script implements the managed resource interface
	if script.GetCondition("Ready").Status == "" {
		// This validates the managed resource interface is implemented
		t.Log("Script v1beta1 successfully implements managed resource interface")
	}

	// Test basic field access
	if script.Spec.ForProvider.ScriptName != "my-worker" {
		t.Errorf("Expected script name 'my-worker', got %s", script.Spec.ForProvider.ScriptName)
	}

	if script.Spec.ForProvider.Script == "" {
		t.Error("Expected script content to be set")
	}

	if script.Spec.ForProvider.Module == nil || *script.Spec.ForProvider.Module != false {
		t.Errorf("Expected module to be false, got %v", script.Spec.ForProvider.Module)
	}

	if script.Spec.ForProvider.CompatibilityDate == nil || *script.Spec.ForProvider.CompatibilityDate != "2023-01-01" {
		t.Errorf("Expected compatibility date '2023-01-01', got %v", script.Spec.ForProvider.CompatibilityDate)
	}

	// Test bindings
	if len(script.Spec.ForProvider.Bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(script.Spec.ForProvider.Bindings))
	}

	binding := script.Spec.ForProvider.Bindings[0]
	if binding.Type != "kv_namespace" {
		t.Errorf("Expected binding type 'kv_namespace', got %s", binding.Type)
	}

	if binding.Name != "MY_KV" {
		t.Errorf("Expected binding name 'MY_KV', got %s", binding.Name)
	}

	if binding.NamespaceID == nil || *binding.NamespaceID != "namespace-123" {
		t.Errorf("Expected namespace ID 'namespace-123', got %v", binding.NamespaceID)
	}

	// Test tags
	if len(script.Spec.ForProvider.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(script.Spec.ForProvider.Tags))
	}

	// Test namespace scope
	if script.Namespace != "worker-scripts" {
		t.Errorf("Expected namespace 'worker-scripts', got %s", script.Namespace)
	}

	t.Log("v1beta1 Worker Script creation and field access tests passed")
}

// TestV1Beta1ScriptAdvancedFeatures tests advanced v1beta1 Worker Script features
func TestV1Beta1ScriptAdvancedFeatures(t *testing.T) {
	placementMode := v1beta1.PlacementModeSmart
	script := &v1beta1.Script{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "advanced-worker-script",
			Namespace: "production-workers",
		},
		Spec: v1beta1.ScriptSpec{
			ForProvider: v1beta1.ScriptParameters{
				ScriptName: "advanced-worker",
				Script: `export default {
					async fetch(request, env, ctx) {
						return new Response('Advanced Worker Response');
					}
				}`,
				Module:             testutils.BoolPtr(true),
				CompatibilityDate:  testutils.StringPtr("2024-01-01"),
				CompatibilityFlags: []string{"nodejs_compat", "durable_objects"},
				UsageModel:         testutils.StringPtr("bundled"),
				Bindings: []v1beta1.WorkerBinding{
					{
						Type: "text_blob",
						Name: "CONFIG",
						Text: testutils.StringPtr("production-config"),
					},
					{
						Type: "json_data",
						Name: "SETTINGS",
						JSON: testutils.StringPtr("{\"debug\": false, \"version\": \"1.2.3\"}"),
					},
				},
				Placement: &placementMode,
				TailConsumers: []v1beta1.TailConsumer{
					{
						Service:     "log-consumer",
						Environment: testutils.StringPtr("production"),
					},
				},
				LogPush: testutils.BoolPtr(true),
				Tags:    []string{"production", "advanced", "es-module"},
			},
		},
	}

	// Test ES module
	if script.Spec.ForProvider.Module == nil || *script.Spec.ForProvider.Module != true {
		t.Errorf("Expected module to be true, got %v", script.Spec.ForProvider.Module)
	}

	// Test usage model
	if script.Spec.ForProvider.UsageModel == nil || *script.Spec.ForProvider.UsageModel != "bundled" {
		t.Errorf("Expected usage model 'bundled', got %v", script.Spec.ForProvider.UsageModel)
	}

	// Test compatibility flags
	if len(script.Spec.ForProvider.CompatibilityFlags) != 2 {
		t.Errorf("Expected 2 compatibility flags, got %d", len(script.Spec.ForProvider.CompatibilityFlags))
	}

	// Test multiple bindings
	if len(script.Spec.ForProvider.Bindings) != 2 {
		t.Errorf("Expected 2 bindings, got %d", len(script.Spec.ForProvider.Bindings))
	}

	// Test text blob binding
	textBinding := script.Spec.ForProvider.Bindings[0]
	if textBinding.Type != "text_blob" {
		t.Errorf("Expected binding type 'text_blob', got %s", textBinding.Type)
	}
	if textBinding.Text == nil || *textBinding.Text != "production-config" {
		t.Errorf("Expected text 'production-config', got %v", textBinding.Text)
	}

	// Test JSON binding
	jsonBinding := script.Spec.ForProvider.Bindings[1]
	if jsonBinding.Type != "json_data" {
		t.Errorf("Expected binding type 'json_data', got %s", jsonBinding.Type)
	}
	if jsonBinding.JSON == nil || *jsonBinding.JSON != "{\"debug\": false, \"version\": \"1.2.3\"}" {
		t.Errorf("Expected JSON data, got %v", jsonBinding.JSON)
	}

	// Test placement
	if script.Spec.ForProvider.Placement == nil || *script.Spec.ForProvider.Placement != v1beta1.PlacementModeSmart {
		t.Errorf("Expected placement 'smart', got %v", script.Spec.ForProvider.Placement)
	}

	// Test tail consumers
	if len(script.Spec.ForProvider.TailConsumers) != 1 {
		t.Errorf("Expected 1 tail consumer, got %d", len(script.Spec.ForProvider.TailConsumers))
	}

	consumer := script.Spec.ForProvider.TailConsumers[0]
	if consumer.Service != "log-consumer" {
		t.Errorf("Expected consumer service 'log-consumer', got %s", consumer.Service)
	}
	if consumer.Environment == nil || *consumer.Environment != "production" {
		t.Errorf("Expected consumer environment 'production', got %v", consumer.Environment)
	}

	// Test log push
	if script.Spec.ForProvider.LogPush == nil || *script.Spec.ForProvider.LogPush != true {
		t.Errorf("Expected log push true, got %v", script.Spec.ForProvider.LogPush)
	}

	// Test tags
	if len(script.Spec.ForProvider.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(script.Spec.ForProvider.Tags))
	}

	t.Log("v1beta1 Worker Script advanced features tests passed")
}
