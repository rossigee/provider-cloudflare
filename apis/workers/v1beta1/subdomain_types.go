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

// SubdomainParameters define the desired state of a Cloudflare Workers Subdomain.
type SubdomainParameters struct {
	// AccountID is the account identifier to target for the resource.
	// +required
	AccountID string `json:"accountId"`

	// Name is the subdomain name to create (e.g., "myaccount" for myaccount.workers.dev).
	// +required
	Name string `json:"name"`
}

// SubdomainObservation are the observable fields of a Workers Subdomain.
type SubdomainObservation struct {
	// Name is the subdomain name (e.g., "myaccount" for myaccount.workers.dev).
	Name *string `json:"name,omitempty"`
}

// SubdomainSpec defines the desired state of Subdomain.
type SubdomainSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     SubdomainParameters `json:"forProvider"`
}

// SubdomainStatus defines the observed state of Subdomain.
type SubdomainStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 SubdomainObservation `json:"atProvider,omitempty"`
}

// A Subdomain is a managed resource that represents a Cloudflare Workers Subdomain configuration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type Subdomain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SubdomainSpec   `json:"spec"`
	Status            SubdomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// SubdomainList contains a list of Subdomain objects.
type SubdomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subdomain `json:"items"`
}
