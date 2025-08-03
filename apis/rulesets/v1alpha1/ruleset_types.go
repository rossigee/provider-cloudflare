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

// RulesetParameters define the desired state of a Cloudflare Ruleset
type RulesetParameters struct {
	// Zone is the zone ID where this ruleset will be applied.
	// Either Zone or Account must be specified, but not both.
	// +optional
	Zone *string `json:"zone,omitempty"`

	// Account is the account ID where this ruleset will be applied.
	// Either Zone or Account must be specified, but not both.
	// +optional
	Account *string `json:"account,omitempty"`

	// Name is the name of the ruleset.
	// +required
	Name string `json:"name"`

	// Description is a description of the ruleset.
	// +optional
	Description *string `json:"description,omitempty"`

	// Kind specifies the kind of ruleset.
	// Valid values: "managed", "custom", "root", "zone"
	// +required
	Kind string `json:"kind"`

	// Phase specifies when the ruleset is executed.
	// Valid values: "ddos_l4", "ddos_l7", "http_request_firewall_custom", 
	// "http_request_firewall_managed", "http_request_transform", 
	// "http_request_late_transform", "http_response_headers_transform",
	// "http_response_firewall", "http_log_custom", "magic_transit"
	// +required
	Phase string `json:"phase"`

	// Rules is the list of rules in this ruleset.
	// +optional
	Rules []RulesetRule `json:"rules,omitempty"`
}

// RulesetRule represents a single rule in a ruleset
type RulesetRule struct {
	// ID is the rule ID (read-only).
	// +optional
	ID *string `json:"id,omitempty"`

	// Action specifies what action to take when the rule matches.
	// Valid values: "allow", "block", "challenge", "js_challenge", 
	// "managed_challenge", "log", "bypass", "rewrite", "redirect", 
	// "route", "score", "serve_error", "set_config", "skip", "execute"
	// +required
	Action string `json:"action"`

	// ActionParameters provides additional configuration for the action.
	// +optional
	ActionParameters *RulesetRuleActionParameters `json:"actionParameters,omitempty"`

	// Expression is the Cloudflare expression that determines when this rule matches.
	// +required
	Expression string `json:"expression"`

	// Description is a human-readable description of the rule.
	// +optional
	Description *string `json:"description,omitempty"`

	// Enabled specifies whether the rule is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Ref is a reference to another rule or ruleset.
	// +optional
	Ref *string `json:"ref,omitempty"`

	// ScoreThreshold is the score threshold for managed rules.
	// +optional
	ScoreThreshold *int `json:"scoreThreshold,omitempty"`

	// RateLimit contains rate limiting configuration.
	// +optional
	RateLimit *RulesetRuleRateLimit `json:"rateLimit,omitempty"`

	// Logging contains logging configuration for the rule.
	// +optional
	Logging *RulesetRuleLogging `json:"logging,omitempty"`
}

// RulesetRuleActionParameters contains action-specific parameters
type RulesetRuleActionParameters struct {
	// ID for execute action referencing another ruleset.
	// +optional
	ID *string `json:"id,omitempty"`

	// Ruleset for execute action referencing another ruleset.
	// +optional
	Ruleset *string `json:"ruleset,omitempty"`

	// Rulesets for execute action referencing multiple rulesets.
	// +optional
	Rulesets []string `json:"rulesets,omitempty"`

	// Rules for overriding specific managed rules.
	// +optional
	Rules map[string][]string `json:"rules,omitempty"`

	// URI contains URI transformation parameters.
	// +optional
	URI *RulesetRuleActionParametersURI `json:"uri,omitempty"`

	// Headers contains HTTP header transformation parameters.
	// +optional
	Headers map[string]RulesetRuleActionParametersHTTPHeader `json:"headers,omitempty"`

	// Response contains block response configuration.
	// +optional
	Response *RulesetRuleActionParametersBlockResponse `json:"response,omitempty"`

	// HostHeader for origin override.
	// +optional
	HostHeader *string `json:"hostHeader,omitempty"`

	// Origin contains origin override parameters.
	// +optional
	Origin *RulesetRuleActionParametersOrigin `json:"origin,omitempty"`

	// Overrides contains rule override parameters for managed rulesets.
	// +optional
	Overrides *RulesetRuleActionParametersOverrides `json:"overrides,omitempty"`

	// Products for managed rulesets.
	// +optional
	Products []string `json:"products,omitempty"`

	// Phases for managed rulesets.
	// +optional
	Phases []string `json:"phases,omitempty"`
}

// RulesetRuleActionParametersURI contains URI transformation parameters
type RulesetRuleActionParametersURI struct {
	// Path contains path transformation parameters.
	// +optional
	Path *RulesetRuleActionParametersURIPath `json:"path,omitempty"`

	// Query contains query transformation parameters.
	// +optional
	Query *RulesetRuleActionParametersURIQuery `json:"query,omitempty"`

	// Origin specifies whether to apply to origin requests.
	// +optional
	Origin *bool `json:"origin,omitempty"`
}

// RulesetRuleActionParametersURIPath contains path transformation parameters
type RulesetRuleActionParametersURIPath struct {
	// Value is the static value to set.
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is the dynamic expression to evaluate.
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// RulesetRuleActionParametersURIQuery contains query transformation parameters
type RulesetRuleActionParametersURIQuery struct {
	// Value is the static value to set.
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is the dynamic expression to evaluate.
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// RulesetRuleActionParametersHTTPHeader contains HTTP header transformation parameters
type RulesetRuleActionParametersHTTPHeader struct {
	// Operation specifies the header operation.
	// Valid values: "set", "add", "remove"
	// +required
	Operation string `json:"operation"`

	// Value is the header value for set/add operations.
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is the dynamic expression for the header value.
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// RulesetRuleActionParametersBlockResponse contains block response configuration
type RulesetRuleActionParametersBlockResponse struct {
	// StatusCode is the HTTP status code to return.
	// +required
	StatusCode int `json:"statusCode"`

	// ContentType is the response content type.
	// +required
	ContentType string `json:"contentType"`

	// Content is the response body content.
	// +required
	Content string `json:"content"`
}

// RulesetRuleActionParametersOrigin contains origin override parameters
type RulesetRuleActionParametersOrigin struct {
	// Host is the origin hostname.
	// +optional
	Host *string `json:"host,omitempty"`

	// Port is the origin port.
	// +optional
	Port *int `json:"port,omitempty"`
}

// RulesetRuleActionParametersOverrides contains rule override parameters
type RulesetRuleActionParametersOverrides struct {
	// Enabled specifies whether overrides are enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Action to override for all rules.
	// +optional
	Action *string `json:"action,omitempty"`

	// SensitivityLevel to override for all rules.
	// +optional
	SensitivityLevel *string `json:"sensitivityLevel,omitempty"`

	// Categories contains category-specific overrides.
	// +optional
	Categories []RulesetRuleActionParametersCategories `json:"categories,omitempty"`

	// Rules contains rule-specific overrides.
	// +optional
	Rules []RulesetRuleActionParametersRules `json:"rules,omitempty"`
}

// RulesetRuleActionParametersCategories contains category override parameters
type RulesetRuleActionParametersCategories struct {
	// Category is the rule category.
	// +required
	Category string `json:"category"`

	// Action to override for this category.
	// +optional
	Action *string `json:"action,omitempty"`

	// Enabled specifies whether this category is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// RulesetRuleActionParametersRules contains rule-specific override parameters
type RulesetRuleActionParametersRules struct {
	// ID is the rule ID to override.
	// +required
	ID string `json:"id"`

	// Action to override for this rule.
	// +optional
	Action *string `json:"action,omitempty"`

	// Enabled specifies whether this rule is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// ScoreThreshold to override for this rule.
	// +optional
	ScoreThreshold *int `json:"scoreThreshold,omitempty"`

	// SensitivityLevel to override for this rule.
	// +optional
	SensitivityLevel *string `json:"sensitivityLevel,omitempty"`
}

// RulesetRuleRateLimit contains rate limiting configuration
type RulesetRuleRateLimit struct {
	// Characteristics define what to rate limit on.
	// +optional
	Characteristics []string `json:"characteristics,omitempty"`

	// RequestsPerPeriod is the number of requests allowed per period.
	// +optional
	RequestsPerPeriod *int `json:"requestsPerPeriod,omitempty"`

	// ScorePerPeriod is the score per period for rate limiting.
	// +optional
	ScorePerPeriod *int `json:"scorePerPeriod,omitempty"`

	// Period is the time period in seconds.
	// +optional
	Period *int `json:"period,omitempty"`

	// MitigationTimeout is how long to block after rate limit hit.
	// +optional
	MitigationTimeout *int `json:"mitigationTimeout,omitempty"`

	// CountingExpression is the expression for counting requests.
	// +optional
	CountingExpression *string `json:"countingExpression,omitempty"`

	// RequestsToOrigin specifies whether to count requests to origin.
	// +optional
	RequestsToOrigin *bool `json:"requestsToOrigin,omitempty"`
}

// RulesetRuleLogging contains logging configuration
type RulesetRuleLogging struct {
	// Enabled specifies whether logging is enabled for this rule.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// RulesetObservation represents the observed state of a Cloudflare Ruleset
type RulesetObservation struct {
	// ID is the ruleset ID.
	ID string `json:"id,omitempty"`

	// Version is the ruleset version.
	Version *string `json:"version,omitempty"`

	// LastUpdated is when the ruleset was last updated.
	LastUpdated *string `json:"lastUpdated,omitempty"`

	// ShareableEntitlementName is the shareable entitlement name.
	ShareableEntitlementName *string `json:"shareableEntitlementName,omitempty"`
}

// RulesetSpec defines the desired state of Ruleset
type RulesetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RulesetParameters `json:"forProvider"`
}

// RulesetStatus defines the observed state of Ruleset
type RulesetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RulesetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// Ruleset is a managed resource that represents a Cloudflare Ruleset
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="KIND",type="string",JSONPath=".spec.forProvider.kind"
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".spec.forProvider.phase"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Ruleset struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RulesetSpec   `json:"spec"`
	Status RulesetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RulesetList contains a list of Rulesets
type RulesetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ruleset `json:"items"`
}