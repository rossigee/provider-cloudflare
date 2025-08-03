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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	// - "(http.request.uri.path contains \"/images/\")"
	// - "(http.request.uri.path.extension eq \"jpg\")"
	// - "(http.host eq \"example.com\" and http.request.uri.path.extension in {\"css\" \"js\"})"
	// +required
	Expression string `json:"expression"`

	// Enabled specifies whether the cache rule is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Priority controls the order of rule execution (lower values execute first).
	// Valid range: 1-1000000000
	// +optional
	Priority *int `json:"priority,omitempty"`

	// Cache controls whether to cache the response.
	// When false, disables caching for matching requests.
	// +optional
	Cache *bool `json:"cache,omitempty"`

	// EdgeTTL controls the cache TTL at Cloudflare edge locations.
	// +optional
	EdgeTTL *EdgeTTL `json:"edgeTTL,omitempty"`

	// BrowserTTL controls the cache TTL in user browsers.
	// +optional
	BrowserTTL *BrowserTTL `json:"browserTTL,omitempty"`

	// ServeStale controls serving stale content from cache.
	// +optional
	ServeStale *ServeStale `json:"serveStale,omitempty"`

	// CacheKey controls how cache keys are generated.
	// +optional
	CacheKey *CacheKey `json:"cacheKey,omitempty"`

	// CacheReserve controls Cache Reserve settings.
	// +optional
	CacheReserve *CacheReserve `json:"cacheReserve,omitempty"`

	// OriginCacheControl controls whether to respect origin cache control headers.
	// +optional
	OriginCacheControl *bool `json:"originCacheControl,omitempty"`

	// RespectStrongETags controls whether to respect strong ETags from origin.
	// +optional
	RespectStrongETags *bool `json:"respectStrongETags,omitempty"`

	// OriginErrorPagePassthru controls whether to pass through origin error pages.
	// +optional
	OriginErrorPagePassthru *bool `json:"originErrorPagePassthru,omitempty"`

	// AdditionalCacheablePorts specifies additional ports where content should be cached.
	// +optional
	AdditionalCacheablePorts []int `json:"additionalCacheablePorts,omitempty"`

	// ReadTimeout specifies the read timeout for origin requests in seconds.
	// +optional
	ReadTimeout *int `json:"readTimeout,omitempty"`
}

// EdgeTTL controls cache TTL at Cloudflare edge locations
type EdgeTTL struct {
	// Mode controls how edge TTL is determined.
	// Valid values: "respect_origin", "override_origin", "bypass"
	// +required
	Mode string `json:"mode"`

	// Default is the default TTL in seconds when mode is "override_origin".
	// +optional
	Default *int `json:"default,omitempty"`

	// StatusCodeTTL allows setting different TTLs based on origin response status codes.
	// +optional
	StatusCodeTTL []StatusCodeTTL `json:"statusCodeTTL,omitempty"`
}

// StatusCodeTTL defines TTL settings for specific HTTP status codes
type StatusCodeTTL struct {
	// StatusCodeValue specifies a single status code (e.g., 200, 404).
	// Either StatusCodeValue or StatusCodeRange must be specified.
	// +optional
	StatusCodeValue *int `json:"statusCodeValue,omitempty"`

	// StatusCodeRange specifies a range of status codes.
	// Either StatusCodeValue or StatusCodeRange must be specified.
	// +optional
	StatusCodeRange *StatusCodeRange `json:"statusCodeRange,omitempty"`

	// Value is the TTL in seconds for the specified status code(s).
	// Use -1 to indicate "no cache".
	// +required
	Value int `json:"value"`
}

// StatusCodeRange defines a range of HTTP status codes
type StatusCodeRange struct {
	// From is the start of the status code range (inclusive).
	// +required
	From int `json:"from"`

	// To is the end of the status code range (inclusive).
	// +required
	To int `json:"to"`
}

// BrowserTTL controls cache TTL in user browsers
type BrowserTTL struct {
	// Mode controls how browser TTL is determined.
	// Valid values: "respect_origin", "override_origin", "bypass"
	// +required
	Mode string `json:"mode"`

	// Default is the default TTL in seconds when mode is "override_origin".
	// +optional
	Default *int `json:"default,omitempty"`
}

// ServeStale controls serving stale content from cache
type ServeStale struct {
	// DisableStaleWhileUpdating disables serving stale content while updating cache.
	// +optional
	DisableStaleWhileUpdating *bool `json:"disableStaleWhileUpdating,omitempty"`
}

// CacheKey controls how cache keys are generated
type CacheKey struct {
	// CacheByDeviceType includes device type in cache key.
	// +optional
	CacheByDeviceType *bool `json:"cacheByDeviceType,omitempty"`

	// IgnoreQueryStringsOrder ignores the order of query string parameters.
	// +optional
	IgnoreQueryStringsOrder *bool `json:"ignoreQueryStringsOrder,omitempty"`

	// CacheDeceptionArmor enables cache deception armor.
	// +optional
	CacheDeceptionArmor *bool `json:"cacheDeceptionArmor,omitempty"`

	// CustomKey allows customization of cache key components.
	// +optional
	CustomKey *CustomKey `json:"customKey,omitempty"`
}

// CustomKey allows customization of cache key components
type CustomKey struct {
	// Query controls query string inclusion in cache key.
	// +optional
	Query *CustomKeyQuery `json:"query,omitempty"`

	// Header controls header inclusion in cache key.
	// +optional
	Header *CustomKeyHeader `json:"header,omitempty"`

	// Cookie controls cookie inclusion in cache key.
	// +optional
	Cookie *CustomKeyFields `json:"cookie,omitempty"`

	// User controls user-specific attributes in cache key.
	// +optional
	User *CustomKeyUser `json:"user,omitempty"`

	// Host controls host-related attributes in cache key.
	// +optional
	Host *CustomKeyHost `json:"host,omitempty"`
}

// CustomKeyQuery controls query string inclusion in cache key
type CustomKeyQuery struct {
	// Include specifies query parameters to include in cache key.
	// If All is true, this field is ignored.
	// +optional
	Include []string `json:"include,omitempty"`

	// Exclude specifies query parameters to exclude from cache key.
	// If All is true, this field is ignored.
	// +optional
	Exclude []string `json:"exclude,omitempty"`

	// All includes all query parameters in cache key.
	// When true, Include and Exclude are ignored.
	// +optional
	All *bool `json:"all,omitempty"`

	// Ignore ignores all query parameters in cache key.
	// +optional
	Ignore *bool `json:"ignore,omitempty"`
}

// CustomKeyHeader controls header inclusion in cache key
type CustomKeyHeader struct {
	// Include specifies headers to include in cache key.
	// +optional
	Include []string `json:"include,omitempty"`

	// CheckPresence specifies headers whose presence (not value) to include in cache key.
	// +optional
	CheckPresence []string `json:"checkPresence,omitempty"`

	// ExcludeOrigin excludes origin headers from cache key.
	// +optional
	ExcludeOrigin *bool `json:"excludeOrigin,omitempty"`

	// Contains specifies header values that must be present for cache key inclusion.
	// Key is header name, value is list of required values.
	// +optional
	Contains map[string][]string `json:"contains,omitempty"`
}

// CustomKeyFields controls general field inclusion in cache key
type CustomKeyFields struct {
	// Include specifies fields to include in cache key.
	// +optional
	Include []string `json:"include,omitempty"`

	// CheckPresence specifies fields whose presence (not value) to include in cache key.
	// +optional
	CheckPresence []string `json:"checkPresence,omitempty"`
}

// CustomKeyUser controls user-specific attributes in cache key
type CustomKeyUser struct {
	// DeviceType includes device type in cache key.
	// +optional
	DeviceType *bool `json:"deviceType,omitempty"`

	// Geo includes geographic information in cache key.
	// +optional
	Geo *bool `json:"geo,omitempty"`

	// Lang includes language preference in cache key.
	// +optional
	Lang *bool `json:"lang,omitempty"`
}

// CustomKeyHost controls host-related attributes in cache key
type CustomKeyHost struct {
	// Resolved includes resolved hostname in cache key.
	// +optional
	Resolved *bool `json:"resolved,omitempty"`
}

// CacheReserve controls Cache Reserve settings
type CacheReserve struct {
	// Eligible controls whether content is eligible for Cache Reserve.
	// +optional
	Eligible *bool `json:"eligible,omitempty"`

	// MinimumFileSize is the minimum file size in bytes for Cache Reserve eligibility.
	// +optional
	MinimumFileSize *int `json:"minimumFileSize,omitempty"`
}

// CacheRuleObservation represents the observed state of a Cloudflare Cache Rule
type CacheRuleObservation struct {
	// ID is the cache rule ID.
	ID string `json:"id,omitempty"`

	// RulesetID is the ID of the underlying ruleset containing this cache rule.
	RulesetID string `json:"rulesetId,omitempty"`

	// Version is the version of the cache rule.
	Version string `json:"version,omitempty"`

	// LastUpdated is when the cache rule was last updated.
	LastUpdated *string `json:"lastUpdated,omitempty"`

	// CreatedOn is when the cache rule was created.
	CreatedOn *string `json:"createdOn,omitempty"`

	// ModifiedOn is when the cache rule was last modified.
	ModifiedOn *string `json:"modifiedOn,omitempty"`
}

// A CacheRuleSpec defines the desired state of a CacheRule.
type CacheRuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CacheRuleParameters `json:"forProvider"`
}

// A CacheRuleStatus represents the observed state of a CacheRule.
type CacheRuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CacheRuleObservation `json:"atProvider,omitempty"`
}

// A CacheRule is a managed resource that represents a Cloudflare Cache Rule
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="EXPRESSION",type="string",JSONPath=".spec.forProvider.expression"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type CacheRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CacheRuleSpec   `json:"spec"`
	Status CacheRuleStatus `json:"status,omitempty"`
}

// CacheRuleList contains a list of CacheRules
// +kubebuilder:object:root=true
type CacheRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CacheRule `json:"items"`
}