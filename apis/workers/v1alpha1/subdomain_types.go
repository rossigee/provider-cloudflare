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
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       SubdomainParameters `json:"forProvider"`
}

// SubdomainStatus defines the observed state of Subdomain.
type SubdomainStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          SubdomainObservation `json:"atProvider,omitempty"`
}

// A Subdomain is a managed resource that represents a Cloudflare Workers Subdomain configuration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
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

// GetCondition of this Subdomain.
func (mg *Subdomain) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Subdomain.
func (mg *Subdomain) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this Subdomain.
func (mg *Subdomain) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this Subdomain.
func (mg *Subdomain) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this Subdomain.
func (mg *Subdomain) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this Subdomain.
func (mg *Subdomain) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this Subdomain.
func (mg *Subdomain) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Subdomain.
func (mg *Subdomain) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this Subdomain.
func (mg *Subdomain) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this Subdomain.
func (mg *Subdomain) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this Subdomain.
func (mg *Subdomain) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this Subdomain.
func (mg *Subdomain) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for Subdomain.
func (mg *Subdomain) GetGroupVersionKind() schema.GroupVersionKind {
	return SubdomainGroupVersionKind
}