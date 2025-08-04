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

// CertificatePackParameters define the desired state of a Cloudflare Certificate Pack.
type CertificatePackParameters struct {
	// Zone is the zone ID where this certificate pack will be created.
	// +required
	Zone string `json:"zone"`

	// Type is the certificate pack type.
	// Valid values: "advanced"
	// +required
	// +kubebuilder:validation:Enum=advanced
	Type string `json:"type"`

	// Hosts are the hostnames to include in the certificate.
	// +required
	Hosts []string `json:"hosts"`

	// ValidationMethod is the method to use for domain validation.
	// Valid values: "txt", "http", "email"
	// +required
	// +kubebuilder:validation:Enum=txt;http;email
	ValidationMethod string `json:"validationMethod"`

	// ValidityDays is the number of days the certificate is valid.
	// Valid values: 14, 30, 90, 365
	// +optional
	// +kubebuilder:validation:Enum=14;30;90;365
	ValidityDays *int `json:"validityDays,omitempty"`

	// CertificateAuthority is the Certificate Authority to use.
	// Valid values: "digicert", "google", "lets_encrypt"
	// +optional
	// +kubebuilder:validation:Enum=digicert;google;lets_encrypt
	CertificateAuthority *string `json:"certificateAuthority,omitempty"`

	// CloudflareBranding indicates whether to show Cloudflare branding on the certificate.
	// +optional
	CloudflareBranding *bool `json:"cloudflareBranding,omitempty"`
}

// SSLValidationRecord represents SSL validation information.
type SSLValidationRecord struct {
	// TxtName is the TXT record name for DNS validation.
	TxtName *string `json:"txtName,omitempty"`

	// TxtValue is the TXT record value for DNS validation.
	TxtValue *string `json:"txtValue,omitempty"`

	// HTTPPath is the HTTP path for HTTP validation.
	HTTPPath *string `json:"httpPath,omitempty"`

	// HTTPBody is the HTTP body content for HTTP validation.
	HTTPBody *string `json:"httpBody,omitempty"`

	// EmailAddresses are the email addresses for email validation.
	EmailAddresses []string `json:"emailAddresses,omitempty"`
}

// SSLValidationError represents SSL validation errors.
type SSLValidationError struct {
	// Message is the error message.
	Message *string `json:"message,omitempty"`
}

// CertificateInfo represents certificate information.
type CertificateInfo struct {
	// ID is the certificate ID.
	ID *string `json:"id,omitempty"`

	// Hosts are the hostnames covered by the certificate.
	Hosts []string `json:"hosts,omitempty"`

	// Issuer is the certificate issuer.
	Issuer *string `json:"issuer,omitempty"`

	// Status is the certificate status.
	Status *string `json:"status,omitempty"`

	// ExpiresOn is when the certificate expires.
	ExpiresOn *metav1.Time `json:"expiresOn,omitempty"`

	// UploadedOn is when the certificate was uploaded.
	UploadedOn *metav1.Time `json:"uploadedOn,omitempty"`

	// ModifiedOn is when the certificate was last modified.
	ModifiedOn *metav1.Time `json:"modifiedOn,omitempty"`
}

// CertificatePackObservation are the observable fields of a Certificate Pack.
type CertificatePackObservation struct {
	// ID is the certificate pack ID.
	ID *string `json:"id,omitempty"`

	// Type is the certificate pack type.
	Type *string `json:"type,omitempty"`

	// Hosts are the hostnames included in the certificate.
	Hosts []string `json:"hosts,omitempty"`

	// Status is the certificate pack status.
	Status *string `json:"status,omitempty"`

	// ValidationMethod is the method used for domain validation.
	ValidationMethod *string `json:"validationMethod,omitempty"`

	// ValidityDays is the number of days the certificate is valid.
	ValidityDays *int `json:"validityDays,omitempty"`

	// CertificateAuthority is the Certificate Authority used.
	CertificateAuthority *string `json:"certificateAuthority,omitempty"`

	// CloudflareBranding indicates whether Cloudflare branding is shown.
	CloudflareBranding *bool `json:"cloudflareBranding,omitempty"`

	// PrimaryCertificate is the primary certificate ID.
	PrimaryCertificate *string `json:"primaryCertificate,omitempty"`

	// Certificates are the certificates in this pack.
	Certificates []CertificateInfo `json:"certificates,omitempty"`

	// ValidationRecords contain domain validation information.
	ValidationRecords []SSLValidationRecord `json:"validationRecords,omitempty"`

	// ValidationErrors contain any validation errors.
	ValidationErrors []SSLValidationError `json:"validationErrors,omitempty"`
}

// CertificatePackSpec defines the desired state of Certificate Pack.
type CertificatePackSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       CertificatePackParameters `json:"forProvider"`
}

// CertificatePackStatus defines the observed state of Certificate Pack.
type CertificatePackStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          CertificatePackObservation `json:"atProvider,omitempty"`
}

// A CertificatePack is a managed resource that represents a Cloudflare Certificate Pack.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.atProvider.status"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".status.atProvider.type"
// +kubebuilder:printcolumn:name="CA",type="string",JSONPath=".status.atProvider.certificateAuthority"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type CertificatePack struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CertificatePackSpec   `json:"spec"`
	Status            CertificatePackStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// CertificatePackList contains a list of Certificate Pack objects.
type CertificatePackList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificatePack `json:"items"`
}

// GetCondition of this CertificatePack.
func (mg *CertificatePack) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this CertificatePack.
func (mg *CertificatePack) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this CertificatePack.
func (mg *CertificatePack) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this CertificatePack.
func (mg *CertificatePack) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this CertificatePack.
func (mg *CertificatePack) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this CertificatePack.
func (mg *CertificatePack) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this CertificatePack.
func (mg *CertificatePack) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this CertificatePack.
func (mg *CertificatePack) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this CertificatePack.
func (mg *CertificatePack) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this CertificatePack.
func (mg *CertificatePack) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this CertificatePack.
func (mg *CertificatePack) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this CertificatePack.
func (mg *CertificatePack) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for CertificatePack.
func (mg *CertificatePack) GetGroupVersionKind() schema.GroupVersionKind {
	return CertificatePackGroupVersionKind
}