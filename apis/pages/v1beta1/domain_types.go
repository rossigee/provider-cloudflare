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

type DomainParameters struct {
	AccountID   string `json:"accountId"`
	ProjectName string `json:"projectName"`
	Domain      string `json:"domain"`
}

type DomainObservation struct {
	ID                 string   `json:"id,omitempty"`
	ProjectName        string   `json:"projectName,omitempty"`
	Domain             string   `json:"domain,omitempty"`
	CustomDomain       string   `json:"customDomain,omitempty"`
	Status             string   `json:"status,omitempty"`
	VerificationErrors []string `json:"verificationErrors,omitempty"`
	CreatedOn          string   `json:"createdOn,omitempty"`
	ModifiedOn         string   `json:"modifiedOn,omitempty"`
}

type DomainSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              DomainParameters `json:"forProvider"`
}

type DomainStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 DomainObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}

type Domain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline,omitempty"`

	Spec   DomainSpec   `json:"spec"`
	Status DomainStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type DomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Domain `json:"items"`
}
