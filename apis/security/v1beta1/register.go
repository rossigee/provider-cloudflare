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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	RateLimitKind     = "RateLimit"
	BotManagementKind = "BotManagement"
	TurnstileKind     = "Turnstile"
)

var (
	RateLimitKindAPIVersion   = RateLimitKind + "." + GroupVersion.String()
	RateLimitGroupKind        = schema.GroupKind{Group: Group, Kind: RateLimitKind}.String()
	RateLimitGroupVersionKind = GroupVersion.WithKind(RateLimitKind)

	BotManagementKindAPIVersion   = BotManagementKind + "." + GroupVersion.String()
	BotManagementGroupKind        = schema.GroupKind{Group: Group, Kind: BotManagementKind}.String()
	BotManagementGroupVersionKind = GroupVersion.WithKind(BotManagementKind)

	TurnstileKindAPIVersion   = TurnstileKind + "." + GroupVersion.String()
	TurnstileGroupKind        = schema.GroupKind{Group: Group, Kind: TurnstileKind}.String()
	TurnstileGroupVersionKind = GroupVersion.WithKind(TurnstileKind)
)
