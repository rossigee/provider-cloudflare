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

// BucketParameters are the configurable fields of a Bucket.
type BucketParameters struct {
	// Name of the bucket. Must be globally unique.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// LocationHint for bucket location preference.
	// Valid values: "apac", "eeur", "enam", "weur", "wnam"
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=apac;eeur;enam;weur;wnam
	LocationHint *string `json:"locationHint,omitempty"`
}

// BucketObservation are the observable fields of a Bucket.
type BucketObservation struct {
	// Name of the bucket.
	Name string `json:"name,omitempty"`

	// CreationDate when the bucket was created.
	CreationDate *metav1.Time `json:"creationDate,omitempty"`

	// Location where the bucket is stored.
	Location string `json:"location,omitempty"`
}

// A BucketSpec defines the desired state of a Bucket.
type BucketSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       BucketParameters `json:"forProvider"`
}

// A BucketStatus represents the observed state of a Bucket.
type BucketStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          BucketObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Bucket is an example API type.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline"`

	Spec   BucketSpec   `json:"spec"`
	Status BucketStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:",inline"`
	Items           []Bucket `json:"items"`
}

// Bucket type metadata.
var (
	BucketKind             = "Bucket"
	BucketGroupKind        = schema.GroupKind{Group: Group, Kind: BucketKind}
	BucketKindAPIVersion   = BucketKind + "." + GroupVersion.String()
	BucketGroupVersionKind = GroupVersion.WithKind(BucketKind)
)