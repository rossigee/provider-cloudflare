/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RateLimitParameters define the desired state of a Cloudflare Rate Limit rule.
type RateLimitParameters struct {
	// Zone is the zone ID where this rate limit will be applied.
	// +required
	Zone string `json:"zone"`

	// Disabled indicates if the rate limit is disabled.
	// +optional
	Disabled *bool `json:"disabled,omitempty"`

	// Description is a human-readable description of the rate limit.
	// +optional
	Description *string `json:"description,omitempty"`

	// Match defines the traffic matching rules for this rate limit.
	// +required
	Match RateLimitTrafficMatcher `json:"match"`

	// Bypass is a list of characteristics that will bypass this rate limit.
	// +optional
	Bypass []RateLimitKeyValue `json:"bypass,omitempty"`

	// Threshold is the number of requests allowed within the specified period.
	// +required
	// +kubebuilder:validation:Minimum=1
	Threshold int `json:"threshold"`

	// Period is the time period in seconds during which the threshold applies.
	// +required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=86400
	Period int `json:"period"`

	// Action defines what action to take when the rate limit is exceeded.
	// +required
	Action RateLimitAction `json:"action"`

	// Correlate defines how requests are correlated for rate limiting.
	// +optional
	Correlate *RateLimitCorrelate `json:"correlate,omitempty"`
}

// RateLimitTrafficMatcher defines the traffic matching rules.
type RateLimitTrafficMatcher struct {
	// Request defines request matching criteria.
	// +optional
	Request *RateLimitMatchRequest `json:"request,omitempty"`

	// Response defines response matching criteria.
	// +optional
	Response *RateLimitMatchResponse `json:"response,omitempty"`
}

// RateLimitMatchRequest defines request matching criteria.
type RateLimitMatchRequest struct {
	// Methods is a list of HTTP methods to match.
	// +optional
	Methods []string `json:"methods,omitempty"`

	// Schemes is a list of URI schemes to match.
	// +optional
	Schemes []string `json:"schemes,omitempty"`

	// URL is the URL pattern to match.
	// +optional
	URL *string `json:"url,omitempty"`
}

// RateLimitMatchResponse defines response matching criteria.
type RateLimitMatchResponse struct {
	// Statuses is a list of HTTP status codes to match.
	// +optional
	Statuses []int `json:"statuses,omitempty"`

	// OriginTraffic indicates whether to count origin traffic.
	// +optional
	OriginTraffic *bool `json:"originTraffic,omitempty"`

	// Headers defines header matching criteria.
	// +optional
	Headers []RateLimitKeyValue `json:"headers,omitempty"`
}

// RateLimitKeyValue represents a key-value pair for matching.
type RateLimitKeyValue struct {
	// Name is the name of the key.
	// +required
	Name string `json:"name"`

	// Value is the value to match.
	// +required
	Value string `json:"value"`
}

// RateLimitAction defines the action to take when rate limit is exceeded.
type RateLimitAction struct {
	// Mode defines the action mode.
	// +kubebuilder:validation:Enum=simulate;ban;challenge;js_challenge;managed_challenge
	// +required
	Mode string `json:"mode"`

	// Timeout is the timeout duration in seconds for ban actions.
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=86400
	Timeout *int `json:"timeout,omitempty"`

	// Response defines the response to send for challenge actions.
	// +optional
	Response *RateLimitActionResponse `json:"response,omitempty"`
}

// RateLimitActionResponse defines the response for challenge actions.
type RateLimitActionResponse struct {
	// ContentType is the content type of the response.
	// +optional
	ContentType *string `json:"contentType,omitempty"`

	// Body is the response body content.
	// +optional
	Body *string `json:"body,omitempty"`
}

// RateLimitCorrelate defines how requests are correlated.
type RateLimitCorrelate struct {
	// By is the correlation method.
	// +kubebuilder:validation:Enum=nat
	// +optional
	By *string `json:"by,omitempty"`
}

// RateLimitObservation are the observable fields of a Rate Limit.
type RateLimitObservation struct {
	// ID is the unique identifier of the rate limit.
	ID string `json:"id,omitempty"`

	// Disabled indicates if the rate limit is disabled.
	Disabled bool `json:"disabled,omitempty"`

	// Description is a human-readable description of the rate limit.
	Description string `json:"description,omitempty"`

	// Match defines the traffic matching rules for this rate limit.
	Match RateLimitTrafficMatcher `json:"match,omitempty"`

	// Bypass is a list of characteristics that will bypass this rate limit.
	Bypass []RateLimitKeyValue `json:"bypass,omitempty"`

	// Threshold is the number of requests allowed within the specified period.
	Threshold int `json:"threshold,omitempty"`

	// Period is the time period in seconds during which the threshold applies.
	Period int `json:"period,omitempty"`

	// Action defines what action to take when the rate limit is exceeded.
	Action RateLimitAction `json:"action,omitempty"`

	// Correlate defines how requests are correlated for rate limiting.
	Correlate *RateLimitCorrelate `json:"correlate,omitempty"`
}

// A RateLimitSpec defines the desired state of a Rate Limit.
type RateLimitSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     RateLimitParameters `json:"forProvider"`
}

// A RateLimitStatus represents the observed state of a Rate Limit.
type RateLimitStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 RateLimitObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RateLimit represents a Cloudflare Rate Limit rule for traffic control.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type RateLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RateLimitSpec   `json:"spec"`
	Status RateLimitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RateLimitList contains a list of Rate Limit objects.
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimit `json:"items"`
}
