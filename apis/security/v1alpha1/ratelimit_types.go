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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	rtv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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

// RateLimitTrafficMatcher contains the rules that will be used to apply a rate limit to traffic.
type RateLimitTrafficMatcher struct {
	// Request contains matching rules for requests.
	// +required
	Request RateLimitRequestMatcher `json:"request"`

	// Response contains matching rules for responses.
	// +optional
	Response *RateLimitResponseMatcher `json:"response,omitempty"`
}

// RateLimitRequestMatcher contains the matching rules pertaining to requests.
type RateLimitRequestMatcher struct {
	// Methods is a list of HTTP methods to match.
	// +optional
	// +kubebuilder:validation:Enum=GET;POST;PUT;DELETE;PATCH;HEAD;OPTIONS
	Methods []string `json:"methods,omitempty"`

	// Schemes is a list of schemes to match.
	// +optional
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	Schemes []string `json:"schemes,omitempty"`

	// URLPattern is a URL pattern to match.
	// +optional
	URLPattern *string `json:"urlPattern,omitempty"`
}

// RateLimitResponseMatcher contains the matching rules pertaining to responses.
type RateLimitResponseMatcher struct {
	// Statuses is a list of HTTP status codes to match.
	// +optional
	Statuses []int `json:"statuses,omitempty"`

	// OriginTraffic indicates if origin traffic should be considered.
	// +optional
	OriginTraffic *bool `json:"originTraffic,omitempty"`

	// Headers is a list of headers to match.
	// +optional
	Headers []RateLimitResponseMatcherHeader `json:"headers,omitempty"`
}

// RateLimitResponseMatcherHeader contains the structure of the origin
// HTTP headers used in request matcher checks.
type RateLimitResponseMatcherHeader struct {
	// Name is the header name.
	// +required
	Name string `json:"name"`

	// Op is the operation to perform.
	// +required
	// +kubebuilder:validation:Enum=eq;ne;contains;startswith;endswith
	Op string `json:"op"`

	// Value is the header value to compare against.
	// +required
	Value string `json:"value"`
}

// RateLimitKeyValue is a key-value pair used in bypass rules.
type RateLimitKeyValue struct {
	// Name is the characteristic name.
	// +required
	Name string `json:"name"`

	// Value is the characteristic value.
	// +required
	Value string `json:"value"`
}

// RateLimitAction defines the action to take when a rate limit is exceeded.
type RateLimitAction struct {
	// Mode defines the action mode.
	// +required
	// +kubebuilder:validation:Enum=simulate;ban;challenge;js_challenge;managed_challenge
	Mode string `json:"mode"`

	// Timeout is the timeout for the action in seconds.
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=86400
	Timeout *int `json:"timeout,omitempty"`

	// Response is a custom response to return.
	// +optional
	Response *RateLimitActionResponse `json:"response,omitempty"`
}

// RateLimitActionResponse defines a custom response for rate limit actions.
type RateLimitActionResponse struct {
	// ContentType is the content type of the response.
	// +required
	ContentType string `json:"contentType"`

	// Body is the response body.
	// +required
	Body string `json:"body"`
}

// RateLimitCorrelate defines how requests are correlated for rate limiting.
type RateLimitCorrelate struct {
	// By defines the correlation method.
	// +required
	// +kubebuilder:validation:Enum=nat;cf-connecting-ip
	By string `json:"by"`
}

// RateLimitSpec defines the desired state of a Rate Limit.
type RateLimitSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       RateLimitParameters `json:"forProvider"`
}

// RateLimitStatus defines the observed state of a Rate Limit.
type RateLimitStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          RateLimitObservation `json:"atProvider,omitempty"`
}

// A RateLimit is a managed resource that represents a Cloudflare Rate Limit rule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="THRESHOLD",type="integer",JSONPath=".status.atProvider.threshold"
// +kubebuilder:printcolumn:name="PERIOD",type="integer",JSONPath=".status.atProvider.period"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type RateLimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RateLimitSpec   `json:"spec"`
	Status            RateLimitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// RateLimitList contains a list of Rate Limit objects.
type RateLimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RateLimit `json:"items"`
}

// GetCondition of this RateLimit.
func (mg *RateLimit) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this RateLimit.
func (mg *RateLimit) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this RateLimit.
func (mg *RateLimit) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this RateLimit.
func (mg *RateLimit) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this RateLimit.
func (mg *RateLimit) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this RateLimit.
func (mg *RateLimit) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this RateLimit.
func (mg *RateLimit) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this RateLimit.
func (mg *RateLimit) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this RateLimit.
func (mg *RateLimit) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this RateLimit.
func (mg *RateLimit) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this RateLimit.
func (mg *RateLimit) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this RateLimit.
func (mg *RateLimit) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for RateLimit.
func (mg *RateLimit) GetGroupVersionKind() schema.GroupVersionKind {
	return RateLimitGroupVersionKind
}