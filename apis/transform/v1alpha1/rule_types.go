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

// Transform Rule phases for different types of transformations
const (
	// PhaseRequestTransform applies URL and header transformations to incoming requests
	PhaseRequestTransform = "http_request_transform"
	// PhaseRequestLateTransform applies transformations after other request processing
	PhaseRequestLateTransform = "http_request_late_transform"
	// PhaseResponseHeadersTransform modifies response headers
	PhaseResponseHeadersTransform = "http_response_headers_transform"
)

// Transform Rule actions
const (
	// ActionRewrite rewrites URLs, query strings, and headers
	ActionRewrite = "rewrite"
	// ActionRedirect performs HTTP redirects
	ActionRedirect = "redirect"
)

// RuleParameters define the desired state of a Transform Rule
type RuleParameters struct {
	// Zone is the zone identifier where the transform rule should be created.
	// +crossplane:generate:reference:type=github.com/rossigee/provider-cloudflare/apis/zone/v1alpha1.Zone
	Zone *string `json:"zone,omitempty"`

	// ZoneRef is a reference to a Zone object.
	ZoneRef *xpv1.Reference `json:"zoneRef,omitempty"`

	// ZoneSelector selects a Zone object.
	ZoneSelector *xpv1.Selector `json:"zoneSelector,omitempty"`

	// Phase specifies the ruleset phase where this rule should be applied.
	// Valid values: http_request_transform, http_request_late_transform, http_response_headers_transform
	// +kubebuilder:validation:Enum=http_request_transform;http_request_late_transform;http_response_headers_transform
	Phase string `json:"phase"`

	// Expression defines the conditions for when this rule should be applied.
	// Uses Cloudflare's Rules Language syntax.
	Expression string `json:"expression"`

	// Action specifies what action to take when the expression matches.
	// Valid values: rewrite, redirect
	// +kubebuilder:validation:Enum=rewrite;redirect
	Action string `json:"action"`

	// Description provides a human-readable description of the rule.
	// +optional
	Description *string `json:"description,omitempty"`

	// Enabled controls whether the rule is active.
	// +optional
	// +kubebuilder:default=true
	Enabled *bool `json:"enabled,omitempty"`

	// Priority determines the order in which rules are executed.
	// Lower numbers have higher priority.
	// +optional
	Priority *int `json:"priority,omitempty"`

	// ActionParameters contains the configuration for the rule action.
	// +optional
	ActionParameters *RuleActionParameters `json:"actionParameters,omitempty"`
}

// RuleActionParameters contains action-specific configuration
type RuleActionParameters struct {
	// URI settings for URL rewriting and redirects
	// +optional
	URI *URITransform `json:"uri,omitempty"`

	// Headers settings for header transformations
	// +optional
	Headers map[string]HTTPHeaderTransform `json:"headers,omitempty"`

	// StatusCode for redirect actions (301, 302, 307, 308)
	// +optional
	// +kubebuilder:validation:Enum=301;302;307;308
	StatusCode *int `json:"statusCode,omitempty"`
}

// URITransform defines URL transformation settings
type URITransform struct {
	// Path settings for URL path transformations
	// +optional
	Path *PathTransform `json:"path,omitempty"`

	// Query settings for query string transformations
	// +optional
	Query *QueryTransform `json:"query,omitempty"`
}

// PathTransform defines URL path transformation settings
type PathTransform struct {
	// Value is the new path value for static rewrites
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is a dynamic expression for path transformation
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// QueryTransform defines query string transformation settings
type QueryTransform struct {
	// Value is the new query string for static rewrites
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is a dynamic expression for query transformation
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// HTTPHeaderTransform defines header transformation settings
type HTTPHeaderTransform struct {
	// Operation specifies what to do with the header
	// Valid values: set, add, remove
	// +kubebuilder:validation:Enum=set;add;remove
	Operation string `json:"operation"`

	// Value is the header value (for set and add operations)
	// +optional
	Value *string `json:"value,omitempty"`

	// Expression is a dynamic expression for header value
	// +optional
	Expression *string `json:"expression,omitempty"`
}

// RuleObservation contains the observed state of a Transform Rule
type RuleObservation struct {
	// ID is the identifier of the transform rule assigned by Cloudflare
	ID string `json:"id,omitempty"`

	// RulesetID is the identifier of the ruleset containing this rule
	RulesetID string `json:"rulesetId,omitempty"`

	// Version is the current version of the rule
	Version string `json:"version,omitempty"`

	// LastUpdated indicates when the rule was last modified
	LastUpdated *metav1.Time `json:"lastUpdated,omitempty"`
}

// RuleSpec defines the desired state of Rule
type RuleSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RuleParameters `json:"forProvider"`
}

// RuleStatus defines the observed state of Rule
type RuleStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RuleObservation `json:"atProvider,omitempty"`
}

// A Rule is a managed resource that represents a Cloudflare Transform Rule
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RuleSpec   `json:"spec"`
	Status RuleStatus `json:"status,omitempty"`
}

// RuleList contains a list of Rule
// +kubebuilder:object:root=true
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rule `json:"items"`
}