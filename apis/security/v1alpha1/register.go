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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	CRDGroup   = "security.cloudflare.crossplane.io"
	CRDVersion = "v1alpha1"
)

var (
	// CRDGroupVersion is the API Group Version used to register the objects
	CRDGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: CRDGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// RateLimit type metadata.
var (
	RateLimitKind             = reflect.TypeOf(RateLimit{}).Name()
	RateLimitGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: RateLimitKind}
	RateLimitKindAPIVersion   = RateLimitKind + "." + CRDGroupVersion.String()
	RateLimitGroupVersionKind = CRDGroupVersion.WithKind(RateLimitKind)
)

// BotManagement type metadata.
var (
	BotManagementKind             = reflect.TypeOf(BotManagement{}).Name()
	BotManagementGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: BotManagementKind}
	BotManagementKindAPIVersion   = BotManagementKind + "." + CRDGroupVersion.String()
	BotManagementGroupVersionKind = CRDGroupVersion.WithKind(BotManagementKind)
)

// Turnstile type metadata.
var (
	TurnstileKind             = reflect.TypeOf(Turnstile{}).Name()
	TurnstileGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TurnstileKind}
	TurnstileKindAPIVersion   = TurnstileKind + "." + CRDGroupVersion.String()
	TurnstileGroupVersionKind = CRDGroupVersion.WithKind(TurnstileKind)
)

func init() {
	SchemeBuilder.Register(&RateLimit{}, &RateLimitList{}, &BotManagement{}, &BotManagementList{}, &Turnstile{}, &TurnstileList{})
}