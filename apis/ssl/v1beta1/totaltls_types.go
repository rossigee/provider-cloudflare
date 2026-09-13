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
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
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
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              TotalTLSParameters `json:"forProvider"`
}

// TotalTLSStatus defines the observed state of Total TLS.
type TotalTLSStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 TotalTLSObservation `json:"atProvider,omitempty"`
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
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
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

// GetCondition gets the condition from the resource status.
func (mg *TotalTLS) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions sets the conditions on the resource status.
func (mg *TotalTLS) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies gets the management policies for the resource.
func (mg *TotalTLS) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies sets the management policies for the resource.
func (mg *TotalTLS) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = mp
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *TotalTLS) DeepCopyObject() runtime.Object {
	out := &TotalTLS{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *TotalTLS) DeepCopyInto(out *TotalTLS) {
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// GetItems returns the list items.
func (l *TotalTLSList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *TotalTLSList) DeepCopyObject() runtime.Object {
	out := &TotalTLSList{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *TotalTLSList) DeepCopyInto(out *TotalTLSList) {
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	out.Items = append([]TotalTLS(nil), in.Items...)
}
