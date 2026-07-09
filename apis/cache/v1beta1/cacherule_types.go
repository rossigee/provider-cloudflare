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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)


// CacheRuleParameters define the desired state of a Cloudflare Cache Rule
type CacheRuleParameters struct {
	// Zone is the zone ID where this cache rule will be applied.
	// Cache rules are zone-scoped resources.
	// +required
	Zone string `json:"zone"`

	// Name is the name of the cache rule.
	// +required
	Name string `json:"name"`

	// Description is a description of the cache rule.
	// +optional
	Description *string `json:"description,omitempty"`

	// Expression is the Cloudflare expression that determines when this cache rule applies.
	// Examples:
	// - "(http.request.uri.path contains "/images/")"
	// - "(http.request.uri.path.extension eq "jpg")"
	// - "(http.host eq "example.com" and http.request.uri.path.extension in {"css" "js"})"
	// +required
	Expression string `json:"expression"`

	// Enabled specifies whether the cache rule is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Action is the cache action to take.
	// +kubebuilder:validation:Enum=set_cache_settings;bypass_cache
	// +kubebuilder:default=set_cache_settings
	// +optional
	Action *string `json:"action,omitempty"`

	// ActionParameters specifies the action parameters for the cache rule.
	// +optional
	ActionParameters *CacheRuleActionParameters `json:"actionParameters,omitempty"`
}

// CacheRuleActionParameters define the action parameters for a cache rule
type CacheRuleActionParameters struct {
	// Cache indicates whether to cache or not cache.
	// +optional
	Cache *bool `json:"cache,omitempty"`

	// CacheKey defines how the cache key is constructed.
	// +optional
	CacheKey *CacheKey `json:"cacheKey,omitempty"`

	// EdgeTTL defines the edge cache TTL settings.
	// +optional
	EdgeTTL *EdgeTTL `json:"edgeTtl,omitempty"`

	// BrowserTTL defines the browser cache TTL settings.
	// +optional
	BrowserTTL *BrowserTTL `json:"browserTtl,omitempty"`

	// ServeStale defines the serve stale settings.
	// +optional
	ServeStale *ServeStale `json:"serveStale,omitempty"`

	// RespectOrigin indicates whether to respect origin cache headers.
	// +optional
	RespectOrigin *bool `json:"respectOrigin,omitempty"`

	// OriginErrorPagePassThru indicates whether to pass through origin error pages.
	// +optional
	OriginErrorPagePassThru *bool `json:"originErrorPagePassThru,omitempty"`
}

// CacheKey defines how the cache key is constructed
type CacheKey struct {
	// IgnoreQueryStringsOrder indicates whether to ignore query string order.
	// +optional
	IgnoreQueryStringsOrder *bool `json:"ignoreQueryStringsOrder,omitempty"`

	// CacheDeceptionArmor indicates whether to enable cache deception armor.
	// +optional
	CacheDeceptionArmor *bool `json:"cacheDeceptionArmor,omitempty"`

	// CustomKey defines custom cache key settings.
	// +optional
	CustomKey *CustomKey `json:"customKey,omitempty"`
}

// CustomKey defines custom cache key settings
type CustomKey struct {
	// Query defines custom query string settings.
	// +optional
	Query *QueryKey `json:"query,omitempty"`

	// Header defines custom header settings.
	// +optional
	Header *HeaderKey `json:"header,omitempty"`

	// Cookie defines custom cookie settings.
	// +optional
	Cookie *CookieKey `json:"cookie,omitempty"`

	// User defines custom user settings.
	// +optional
	User *UserKey `json:"user,omitempty"`

	// Host defines custom host settings.
	// +optional
	Host *HostKey `json:"host,omitempty"`
}

// QueryKey defines query string cache key settings
type QueryKey struct {
	// Exclude is a list of query string parameters to exclude.
	// +optional
	Exclude []string `json:"exclude,omitempty"`

	// Include is a list of query string parameters to include.
	// +optional
	Include []string `json:"include,omitempty"`
}

// HeaderKey defines header cache key settings
type HeaderKey struct {
	// CheckPresence is a list of headers to check for presence.
	// +optional
	CheckPresence []string `json:"checkPresence,omitempty"`

	// Exclude is a list of headers to exclude.
	// +optional
	Exclude []string `json:"exclude,omitempty"`

	// Include is a list of headers to include.
	// +optional
	Include []string `json:"include,omitempty"`
}

// CookieKey defines cookie cache key settings
type CookieKey struct {
	// CheckPresence is a list of cookies to check for presence.
	// +optional
	CheckPresence []string `json:"checkPresence,omitempty"`

	// Include is a list of cookies to include.
	// +optional
	Include []string `json:"include,omitempty"`
}

// UserKey defines user cache key settings
type UserKey struct {
	// DeviceType indicates whether to vary by device type.
	// +optional
	DeviceType *bool `json:"deviceType,omitempty"`

	// Geo indicates whether to vary by geo.
	// +optional
	Geo *bool `json:"geo,omitempty"`

	// Lang indicates whether to vary by language.
	// +optional
	Lang *bool `json:"lang,omitempty"`
}

// HostKey defines host cache key settings
type HostKey struct {
	// Resolved indicates whether to use resolved host.
	// +optional
	Resolved *bool `json:"resolved,omitempty"`
}

// EdgeTTL defines edge cache TTL settings
type EdgeTTL struct {
	// Mode defines the edge TTL mode.
	// +kubebuilder:validation:Enum=override_origin;respect_origin;bypass_by_default
	// +optional
	Mode *string `json:"mode,omitempty"`

	// Default is the default edge TTL in seconds.
	// +optional
	Default *int64 `json:"default,omitempty"`

	// StatusCodeTTL is a list of status code specific TTL settings.
	// +optional
	StatusCodeTTL []StatusCodeTTL `json:"statusCodeTtl,omitempty"`
}

// StatusCodeTTL defines TTL settings for specific status codes
type StatusCodeTTL struct {
	// StatusCode is the specific status code.
	// +optional
	StatusCode *int64 `json:"statusCode,omitempty"`

	// StatusCodeRange defines a range of status codes.
	// +optional
	StatusCodeRange *StatusCodeRange `json:"statusCodeRange,omitempty"`

	// Value is the TTL value in seconds.
	// +required
	Value int64 `json:"value"`
}

// StatusCodeRange defines a range of status codes
type StatusCodeRange struct {
	// From is the start of the status code range.
	// +required
	From int64 `json:"from"`

	// To is the end of the status code range.
	// +required
	To int64 `json:"to"`
}

// BrowserTTL defines browser cache TTL settings
type BrowserTTL struct {
	// Mode defines the browser TTL mode.
	// +kubebuilder:validation:Enum=override_origin;respect_origin;bypass_by_default
	// +optional
	Mode *string `json:"mode,omitempty"`

	// Default is the default browser TTL in seconds.
	// +optional
	Default *int64 `json:"default,omitempty"`
}

// ServeStale defines serve stale settings
type ServeStale struct {
	// DisableStaleWhileUpdating indicates whether to disable serve stale while updating.
	// +optional
	DisableStaleWhileUpdating *bool `json:"disableStaleWhileUpdating,omitempty"`
}

// CacheRuleObservation is the observable fields of a Cache Rule.
type CacheRuleObservation struct {
	// ID of the created cache rule.
	ID string `json:"id,omitempty"`

	// RulesetID of the ruleset containing the cache rule.
	RulesetID string `json:"rulesetId,omitempty"`

	// Phase indicates the phase where the cache rule is executed.
	Phase string `json:"phase,omitempty"`

	// Version indicates the version of the cache rule.
	Version string `json:"version,omitempty"`

	// LastModified indicates when the cache rule was last modified.
	LastModified string `json:"lastModified,omitempty"`
}

// A CacheRuleSpec defines the desired state of a Cache Rule.
type CacheRuleSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       CacheRuleParameters `json:"forProvider"`
}

// A CacheRuleStatus represents the observed state of a Cache Rule.
type CacheRuleStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          CacheRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CacheRule represents a Cloudflare Cache Rule for advanced caching control.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type CacheRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheRuleSpec   `json:"spec"`
	Status CacheRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CacheRuleList contains a list of Cache Rule objects.
type CacheRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRule `json:"items"`
}
