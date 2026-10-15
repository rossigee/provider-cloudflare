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

type ProjectParameters struct {
	AccountID string `json:"accountId"`
	Name      string `json:"name"`

	ProductionBranch     string            `json:"productionBranch,omitempty"`
	BuildCommand         string            `json:"buildCommand,omitempty"`
	DestinationDir       string            `json:"destinationDir,omitempty"`
	RootDir              string            `json:"rootDir,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	GitRepository        *GitRepository    `json:"gitRepository,omitempty"`
}

type GitRepository struct {
	URL       string `json:"url"`
	Branch    string `json:"branch,omitempty"`
	ConfigDir string `json:"configDir,omitempty"`
}

type ProjectObservation struct {
	ID                   string            `json:"id,omitempty"`
	Name                 string            `json:"name,omitempty"`
	Subdomain            string            `json:"subdomain,omitempty"`
	CustomDomains        []string          `json:"customDomains,omitempty"`
	ProductionBranch     string            `json:"productionBranch,omitempty"`
	BuildCommand         string            `json:"buildCommand,omitempty"`
	DestinationDir       string            `json:"destinationDir,omitempty"`
	RootDir              string            `json:"rootDir,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	CreatedOn            string            `json:"createdOn,omitempty"`
	ModifiedOn           string            `json:"modifiedOn,omitempty"`
	Domains              []string          `json:"domains,omitempty"`
}

type ProjectSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              ProjectParameters `json:"forProvider"`
}

type ProjectStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 ProjectObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline,omitempty"`

	Spec   ProjectSpec   `json:"spec"`
	Status ProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type ProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Project `json:"items"`
}
