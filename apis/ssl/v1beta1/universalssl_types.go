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
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
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
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider                     UniversalSSLParameters `json:"forProvider"`
}

// UniversalSSLStatus defines the observed state of Universal SSL.
type UniversalSSLStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 UniversalSSLObservation `json:"atProvider,omitempty"`
}

// A UniversalSSL is a managed resource that represents Cloudflare Universal SSL configuration for a zone.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="ENABLED",type="boolean",JSONPath=".status.atProvider.enabled"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
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


// GetCondition gets the condition from the resource status.
func (mg *UniversalSSL) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions sets the conditions on the resource status.
func (mg *UniversalSSL) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies gets the management policies for the resource.
func (mg *UniversalSSL) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies sets the management policies for the resource.
func (mg *UniversalSSL) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = mp
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *UniversalSSL) DeepCopyObject() runtime.Object {
	out := &UniversalSSL{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *UniversalSSL) DeepCopyInto(out *UniversalSSL) {
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// GetItems returns the list items.
func (l *UniversalSSLList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *UniversalSSLList) DeepCopyObject() runtime.Object {
	out := &UniversalSSLList{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *UniversalSSLList) DeepCopyInto(out *UniversalSSLList) {
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	out.Items = append([]UniversalSSL(nil), in.Items...)
}
