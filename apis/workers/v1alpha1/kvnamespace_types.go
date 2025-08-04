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

// KVNamespaceParameters are the configurable fields of a Workers KV Namespace.
type KVNamespaceParameters struct {
	// Title is the human-readable name of the KV namespace.
	Title string `json:"title"`
}

// KVNamespaceObservation are the observable fields of a Workers KV Namespace.
type KVNamespaceObservation struct {
	// ID is the unique identifier for the KV namespace.
	ID string `json:"id,omitempty"`

	// Title is the human-readable name of the KV namespace.
	Title string `json:"title,omitempty"`
}

// A KVNamespaceSpec defines the desired state of a Workers KV Namespace.
type KVNamespaceSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       KVNamespaceParameters `json:"forProvider"`
}

// A KVNamespaceStatus represents the observed state of a Workers KV Namespace.
type KVNamespaceStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          KVNamespaceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A KVNamespace represents a Workers KV storage namespace.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="TITLE",type="string",JSONPath=".spec.forProvider.title"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type KVNamespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KVNamespaceSpec   `json:"spec"`
	Status KVNamespaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// KVNamespaceList contains a list of Workers KV Namespace objects
type KVNamespaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KVNamespace `json:"items"`
}

