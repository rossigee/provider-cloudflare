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

// DomainParameters define the desired state of a Cloudflare Workers Custom Domain.
type DomainParameters struct {
	// AccountID is the account identifier to target for the resource.
	// +required
	AccountID string `json:"accountId"`

	// ZoneID is the zone identifier where the custom domain will be created.
	// +required
	ZoneID string `json:"zoneId"`

	// Hostname is the custom hostname to attach the Worker to.
	// +required
	Hostname string `json:"hostname"`

	// Service is the name of the Worker Script to attach to this domain.
	// +required
	Service string `json:"service"`

	// Environment is the environment to use for this domain attachment.
	// Valid values: "production", "staging"
	// +required
	// +kubebuilder:validation:Enum=production;staging
	Environment string `json:"environment"`
}

// DomainObservation are the observable fields of a Workers Custom Domain.
type DomainObservation struct {
	// ID is the unique identifier for this domain attachment.
	ID *string `json:"id,omitempty"`

	// ZoneID is the zone identifier where the custom domain is created.
	ZoneID *string `json:"zoneId,omitempty"`

	// ZoneName is the zone name where the custom domain is created.
	ZoneName *string `json:"zoneName,omitempty"`

	// Hostname is the custom hostname attached to the Worker.
	Hostname *string `json:"hostname,omitempty"`

	// Service is the name of the Worker Script attached to this domain.
	Service *string `json:"service,omitempty"`

	// Environment is the environment used for this domain attachment.
	Environment *string `json:"environment,omitempty"`
}

// DomainSpec defines the desired state of Domain.
type DomainSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     DomainParameters `json:"forProvider"`
}

// DomainStatus defines the observed state of Domain.
type DomainStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 DomainObservation `json:"atProvider,omitempty"`
}

// A Domain is a managed resource that represents a Cloudflare Workers Custom Domain.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="HOSTNAME",type="string",JSONPath=".status.atProvider.hostname"
// +kubebuilder:printcolumn:name="SERVICE",type="string",JSONPath=".status.atProvider.service"
// +kubebuilder:printcolumn:name="ENV",type="string",JSONPath=".status.atProvider.environment"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type Domain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              DomainSpec   `json:"spec"`
	Status            DomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// DomainList contains a list of Domain objects.
type DomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Domain `json:"items"`
}
