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

// RulesetRule defines a single rule within a ruleset
type RulesetRule struct {
	// Action specifies what to do when the rule matches
	// Valid values: "allow", "block", "challenge", "js_challenge", "log", "skip", "rewrite", "redirect"
	// +required
	Action string `json:"action"`

	// Expression defines the conditions for when this rule matches
	// Uses Cloudflare's filter expression syntax
	// +required
	Expression string `json:"expression"`

	// Description is a human-readable description of the rule
	// +optional
	Description *string `json:"description,omitempty"`

	// Enabled indicates whether this rule is active
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

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
	// Common phases: "http_request_firewall_custom", "http_request_transform", "http_response_headers_transform"
	// +required
	Phase string `json:"phase"`

	// Rules is the list of rules in this ruleset
	// +optional
	Rules []RulesetRule `json:"rules,omitempty"`
}

// RulesetObservation contains the observable fields of a Ruleset
type RulesetObservation struct {
	// ID is the unique identifier for the ruleset
	ID string `json:"id,omitempty"`

	// Name is the name of the ruleset
	Name string `json:"name,omitempty"`

	// Description is the description of the ruleset
	Description string `json:"description,omitempty"`

	// Kind is the kind of ruleset
	Kind string `json:"kind,omitempty"`

	// Phase is the phase when the ruleset is executed
	Phase string `json:"phase,omitempty"`

	// Version is the version of the ruleset
	Version string `json:"version,omitempty"`

	// LastModified indicates when this ruleset was last modified
	LastModified *metav1.Time `json:"lastModified,omitempty"`
}

// A RulesetSpec defines the desired state of a Ruleset.
type RulesetSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       RulesetParameters `json:"forProvider"`
}

// A RulesetStatus represents the observed state of a Ruleset.
type RulesetStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          RulesetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Ruleset provides advanced security and filtering capabilities using Cloudflare's Ruleset Engine
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="KIND",type="string",JSONPath=".spec.forProvider.kind"
// +kubebuilder:printcolumn:name="PHASE",type="string",JSONPath=".spec.forProvider.phase"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
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
func init() {
	SchemeBuilder.Register(&Ruleset{}, &RulesetList{})
}
