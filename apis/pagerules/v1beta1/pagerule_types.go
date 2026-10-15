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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PageRuleParameters struct {
	ZoneID   string   `json:"zoneId"`
	Targets  []Target `json:"targets"`
	Actions  []Action `json:"actions"`
	Priority int      `json:"priority,omitempty"`
	Status   string   `json:"status,omitempty"`
}

type TargetConstraint struct {
	Operator string `json:"operator,omitempty"`
	Value    string `json:"value,omitempty"`
}

type Target struct {
	Target     string           `json:"target"`
	Constraint TargetConstraint `json:"constraint,omitempty"`
}

type Action struct {
	ID    string `json:"id"`
	Value string `json:"value,omitempty"`
}

type PageRuleObservation struct {
	ID         string   `json:"id,omitempty"`
	ZoneID     string   `json:"zoneId,omitempty"`
	Targets    []Target `json:"targets,omitempty"`
	Actions    []Action `json:"actions,omitempty"`
	Priority   int      `json:"priority,omitempty"`
	Status     string   `json:"status,omitempty"`
	CreatedOn  string   `json:"createdOn,omitempty"`
	ModifiedOn string   `json:"modifiedOn,omitempty"`
}

type PageRuleSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              PageRuleParameters `json:"forProvider"`
}

type PageRuleStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 PageRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}

type PageRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline,omitempty"`

	Spec   PageRuleSpec   `json:"spec"`
	Status PageRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type PageRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PageRule `json:"items"`
}
