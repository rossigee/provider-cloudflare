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

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RateLimitTrafficMatcher type metadata.
var (
	RateLimitTrafficMatcherKind             = reflect.TypeOf(RateLimitTrafficMatcher{}).Name()
	RateLimitTrafficMatcherGroupKind        = schema.GroupKind{Group: Group, Kind: RateLimitTrafficMatcherKind}
	RateLimitTrafficMatcherKindAPIVersion   = RateLimitTrafficMatcherKind + "." + SchemeGroupVersion.String()
	RateLimitTrafficMatcherGroupVersionKind = SchemeGroupVersion.WithKind(RateLimitTrafficMatcherKind)

	RateLimitKind             = reflect.TypeOf(RateLimit{}).Name()
	RateLimitGroupKind        = schema.GroupKind{Group: Group, Kind: RateLimitKind}
	RateLimitKindAPIVersion   = RateLimitKind + "." + SchemeGroupVersion.String()
	RateLimitGroupVersionKind = SchemeGroupVersion.WithKind(RateLimitKind)

	BotManagementKind             = reflect.TypeOf(BotManagement{}).Name()
	BotManagementGroupKind        = schema.GroupKind{Group: Group, Kind: BotManagementKind}
	BotManagementKindAPIVersion   = BotManagementKind + "." + SchemeGroupVersion.String()
	BotManagementGroupVersionKind = SchemeGroupVersion.WithKind(BotManagementKind)

	TurnstileKind             = reflect.TypeOf(Turnstile{}).Name()
	TurnstileGroupKind        = schema.GroupKind{Group: Group, Kind: TurnstileKind}
	TurnstileKindAPIVersion   = TurnstileKind + "." + SchemeGroupVersion.String()
	TurnstileGroupVersionKind = SchemeGroupVersion.WithKind(TurnstileKind)
)
