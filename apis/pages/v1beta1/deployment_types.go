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

type DeploymentParameters struct {
	AccountID   string `json:"accountId"`
	ProjectName string `json:"projectName"`
	Branch      string `json:"branch,omitempty"`

	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	Aliases              []string          `json:"aliases,omitempty"`
}

type DeploymentObservation struct {
	ID                   string            `json:"id,omitempty"`
	ProjectName          string            `json:"projectName,omitempty"`
	Branch               string            `json:"branch,omitempty"`
	Environment          string            `json:"environment,omitempty"`
	URL                  string            `json:"url,omitempty"`
	Aliases              []string          `json:"aliases,omitempty"`
	ShortID              string            `json:"shortId,omitempty"`
	LatestStage          *DeploymentStage  `json:"latestStage,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`
	CreatedOn            string            `json:"createdOn,omitempty"`
	ModifiedOn           string            `json:"modifiedOn,omitempty"`
}

type DeploymentStage struct {
	Name      string `json:"name,omitempty"`
	Status    string `json:"status,omitempty"`
	StartedOn string `json:"startedOn,omitempty"`
	EndedOn   string `json:"endedOn,omitempty"`
	Duration  int    `json:"duration,omitempty"`
}

type DeploymentSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              DeploymentParameters `json:"forProvider"`
}

type DeploymentStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 DeploymentObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}

type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline,omitempty"`

	Spec   DeploymentSpec   `json:"spec"`
	Status DeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployment `json:"items"`
}
