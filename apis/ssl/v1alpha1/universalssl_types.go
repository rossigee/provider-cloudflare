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

// UniversalSSLParameters define the desired state of Cloudflare Universal SSL for a zone.
type UniversalSSLParameters struct {
	// Zone is the zone ID where this Universal SSL configuration will be applied.
	// +required
	Zone string `json:"zone"`

	// Enabled indicates whether Universal SSL is enabled for this zone.
	// +required
	Enabled bool `json:"enabled"`
}

// UniversalSSLObservation are the observable fields of Universal SSL.
type UniversalSSLObservation struct {
	// Enabled indicates whether Universal SSL is enabled for this zone.
	Enabled *bool `json:"enabled,omitempty"`
}

// UniversalSSLSpec defines the desired state of Universal SSL.
type UniversalSSLSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       UniversalSSLParameters `json:"forProvider"`
}

// UniversalSSLStatus defines the observed state of Universal SSL.
type UniversalSSLStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          UniversalSSLObservation `json:"atProvider,omitempty"`
}

// A UniversalSSL is a managed resource that represents Cloudflare Universal SSL configuration for a zone.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="ENABLED",type="boolean",JSONPath=".status.atProvider.enabled"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type UniversalSSL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              UniversalSSLSpec   `json:"spec"`
	Status            UniversalSSLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// UniversalSSLList contains a list of Universal SSL objects.
type UniversalSSLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UniversalSSL `json:"items"`
}

// GetCondition of this UniversalSSL.
func (mg *UniversalSSL) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this UniversalSSL.
func (mg *UniversalSSL) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this UniversalSSL.
func (mg *UniversalSSL) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this UniversalSSL.
func (mg *UniversalSSL) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this UniversalSSL.
func (mg *UniversalSSL) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this UniversalSSL.
func (mg *UniversalSSL) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this UniversalSSL.
func (mg *UniversalSSL) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this UniversalSSL.
func (mg *UniversalSSL) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this UniversalSSL.
func (mg *UniversalSSL) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this UniversalSSL.
func (mg *UniversalSSL) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this UniversalSSL.
func (mg *UniversalSSL) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this UniversalSSL.
func (mg *UniversalSSL) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for UniversalSSL.
func (mg *UniversalSSL) GetGroupVersionKind() schema.GroupVersionKind {
	return UniversalSSLGroupVersionKind
}