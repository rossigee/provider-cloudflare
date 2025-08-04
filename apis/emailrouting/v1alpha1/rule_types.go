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

// RuleParameters are the configurable fields of an Email Routing Rule.
type RuleParameters struct {
	// ZoneID is the zone identifier to target for the resource.
	// +kubebuilder:validation:Required
	ZoneID string `json:"zoneId"`

	// Name of the email routing rule.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Priority of the rule. Lower numbers have higher priority.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Priority int `json:"priority"`

	// Enabled indicates if the rule is enabled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Matchers define the conditions for the rule.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Matchers []RuleMatcher `json:"matchers"`

	// Actions define what happens when the rule matches.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Actions []RuleAction `json:"actions"`
}

// RuleMatcher defines a condition for an email routing rule.
type RuleMatcher struct {
	// Type of matcher.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=literal;all
	Type string `json:"type"`

	// Field to match against.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=to;from;subject
	Field string `json:"field"`

	// Value to match.
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// RuleAction defines an action for an email routing rule.
type RuleAction struct {
	// Type of action.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=forward;worker;drop
	Type string `json:"type"`

	// Value contains the action parameters.
	// For "forward" actions, this should be email addresses.
	// For "worker" actions, this should be worker script names.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Value []string `json:"value"`
}

// RuleObservation are the observable fields of an Email Routing Rule.
type RuleObservation struct {
	// Tag is the unique identifier for the rule.
	Tag string `json:"tag,omitempty"`

	// ZoneID is the zone identifier to target for the resource.
	ZoneID string `json:"zoneId,omitempty"`

	// Name of the email routing rule.
	Name string `json:"name,omitempty"`

	// Priority of the rule.
	Priority *int `json:"priority,omitempty"`

	// Enabled indicates if the rule is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Matchers define the conditions for the rule.
	Matchers []RuleMatcher `json:"matchers,omitempty"`

	// Actions define what happens when the rule matches.
	Actions []RuleAction `json:"actions,omitempty"`
}

// A RuleSpec defines the desired state of an Email Routing Rule.
type RuleSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       RuleParameters `json:"forProvider"`
}

// A RuleStatus represents the observed state of an Email Routing Rule.
type RuleStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          RuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Rule is a Cloudflare Email Routing Rule.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Rule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline"`

	Spec   RuleSpec   `json:"spec"`
	Status RuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RuleList contains a list of Rule
type RuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:",inline"`
	Items           []Rule `json:"items"`
}

// Rule type metadata.
var (
	RuleKind             = "Rule"
	RuleGroupKind        = schema.GroupKind{Group: Group, Kind: RuleKind}
	RuleKindAPIVersion   = RuleKind + "." + GroupVersion.String()
	RuleGroupVersionKind = GroupVersion.WithKind(RuleKind)
)