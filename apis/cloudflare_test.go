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

package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestVerifySchemeRegistration(t *testing.T) {
	// CRITICAL TEST: Verifies all API types are properly registered with the scheme
	// This catches scheme registration issues that would cause panics during provider startup
	// FAILURE INDICATOR: If this test fails, the provider will panic on startup

	err := VerifySchemeRegistration()
	if err != nil {
		t.Fatalf("CRITICAL: Scheme registration verification failed - TESTING FAILURE: %vnThis indicates that some API types are not properly registered with the scheme.nCheck that all API packages have init() functions that call SchemeBuilder.Register()", err)
	}
}

func TestAddToScheme(t *testing.T) {
	// Test that AddToScheme works without panicking
	// This would catch issues where scheme registration fails

	scheme := runtime.NewScheme()
	err := AddToScheme(scheme)
	if err != nil {
		t.Errorf("AddToScheme failed: %v", err)
	}
}
