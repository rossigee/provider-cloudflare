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

// CertificateParameters define the desired state of a Cloudflare Origin CA Certificate.
type CertificateParameters struct {
	// Hostnames is the list of hostnames or wildcard names (beginning with "*.")
	// for which this certificate is valid.
	// +kubebuilder:validation:MinItems=1
	Hostnames []string `json:"hostnames"`

	// RequestType is the signature type to create the certificate with. Options: "origin-rsa", "origin-ecc", "keyless-certificate".
	// +kubebuilder:validation:Enum=origin-rsa;origin-ecc;keyless-certificate
	// +optional
	RequestType *string `json:"requestType,omitempty"`

	// RequestValidity is the number of days for which the certificate should be valid.
	// Valid values: 7, 30, 90, 365, 730, 1095, 5475
	// +kubebuilder:validation:Enum=7;30;90;365;730;1095;5475
	// +optional
	RequestValidity *int `json:"requestValidity,omitempty"`

	// CSR is the Certificate Signing Request. Must be newline-encoded.
	// If not provided, Cloudflare will generate a private key and CSR.
	// +optional
	CSR *string `json:"csr,omitempty"`
}

// CertificateObservation represents the observed state of a Cloudflare Origin CA Certificate.
type CertificateObservation struct {
	// ID is the certificate ID.
	ID string `json:"id,omitempty"`

	// Certificate is the PEM-encoded certificate.
	Certificate string `json:"certificate,omitempty"`

	// Hostnames is the list of hostnames for which this certificate is valid.
	Hostnames []string `json:"hostnames,omitempty"`

	// ExpiresOn is the date and time when the certificate expires.
	ExpiresOn *metav1.Time `json:"expiresOn,omitempty"`

	// RequestType is the signature type of the certificate.
	RequestType string `json:"requestType,omitempty"`

	// RequestValidity is the number of days for which the certificate is valid.
	RequestValidity int `json:"requestValidity,omitempty"`

	// RevokedAt is the date and time when the certificate was revoked (if applicable).
	RevokedAt *metav1.Time `json:"revokedAt,omitempty"`

	// CSR is the Certificate Signing Request used to generate this certificate.
	CSR string `json:"csr,omitempty"`
}

// CertificateSpec defines the desired state of a Certificate.
type CertificateSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       CertificateParameters `json:"forProvider"`
}

// CertificateStatus defines the observed state of a Certificate.
type CertificateStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          CertificateObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Certificate is a managed resource that represents a Cloudflare Origin CA Certificate.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="HOSTNAMES",type="string",JSONPath=".status.atProvider.hostnames"
// +kubebuilder:printcolumn:name="EXPIRES",type="string",JSONPath=".status.atProvider.expiresOn"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec"`
	Status CertificateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CertificateList contains a list of Certificate
type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

// GetCondition of this Certificate.
func (mg *Certificate) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Certificate.
func (mg *Certificate) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this Certificate.
func (mg *Certificate) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this Certificate.
func (mg *Certificate) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this Certificate.
func (mg *Certificate) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this Certificate.
func (mg *Certificate) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this Certificate.
func (mg *Certificate) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Certificate.
func (mg *Certificate) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this Certificate.
func (mg *Certificate) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this Certificate.
func (mg *Certificate) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this Certificate.
func (mg *Certificate) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this Certificate.
func (mg *Certificate) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for Certificate.
func (mg *Certificate) GetGroupVersionKind() schema.GroupVersionKind {
	return CertificateGroupVersionKind
}