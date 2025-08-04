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

// TotalTLSParameters define the desired state of Cloudflare Total TLS for a zone.
type TotalTLSParameters struct {
	// Zone is the zone ID where this Total TLS configuration will be applied.
	// +required
	Zone string `json:"zone"`

	// Enabled indicates whether Total TLS is enabled for this zone.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// CertificateAuthority is the Certificate Authority to use for Total TLS.
	// Valid values: "google", "lets_encrypt"
	// +optional
	// +kubebuilder:validation:Enum=google;lets_encrypt
	CertificateAuthority *string `json:"certificateAuthority,omitempty"`

	// ValidityDays is the number of days the certificate is valid.
	// Valid values: 14, 30, 90
	// +optional
	// +kubebuilder:validation:Enum=14;30;90
	ValidityDays *int `json:"validityDays,omitempty"`
}

// TotalTLSObservation are the observable fields of Total TLS.
type TotalTLSObservation struct {
	// Enabled indicates whether Total TLS is enabled for this zone.
	Enabled *bool `json:"enabled,omitempty"`

	// CertificateAuthority is the Certificate Authority used for Total TLS.
	CertificateAuthority *string `json:"certificateAuthority,omitempty"`

	// ValidityDays is the number of days the certificate is valid.
	ValidityDays *int `json:"validityDays,omitempty"`
}

// TotalTLSSpec defines the desired state of Total TLS.
type TotalTLSSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       TotalTLSParameters `json:"forProvider"`
}

// TotalTLSStatus defines the observed state of Total TLS.
type TotalTLSStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          TotalTLSObservation `json:"atProvider,omitempty"`
}

// A TotalTLS is a managed resource that represents Cloudflare Total TLS configuration for a zone.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="ENABLED",type="boolean",JSONPath=".status.atProvider.enabled"
// +kubebuilder:printcolumn:name="CA",type="string",JSONPath=".status.atProvider.certificateAuthority"
// +kubebuilder:printcolumn:name="VALIDITY",type="integer",JSONPath=".status.atProvider.validityDays"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type TotalTLS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TotalTLSSpec   `json:"spec"`
	Status            TotalTLSStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TotalTLSList contains a list of Total TLS objects.
type TotalTLSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TotalTLS `json:"items"`
}

// GetCondition of this TotalTLS.
func (mg *TotalTLS) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this TotalTLS.
func (mg *TotalTLS) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this TotalTLS.
func (mg *TotalTLS) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this TotalTLS.
func (mg *TotalTLS) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this TotalTLS.
func (mg *TotalTLS) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this TotalTLS.
func (mg *TotalTLS) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this TotalTLS.
func (mg *TotalTLS) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this TotalTLS.
func (mg *TotalTLS) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this TotalTLS.
func (mg *TotalTLS) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this TotalTLS.
func (mg *TotalTLS) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this TotalTLS.
func (mg *TotalTLS) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this TotalTLS.
func (mg *TotalTLS) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for TotalTLS.
func (mg *TotalTLS) GetGroupVersionKind() schema.GroupVersionKind {
	return TotalTLSGroupVersionKind
}