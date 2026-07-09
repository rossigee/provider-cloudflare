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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)


// PlacementMode represents the placement mode for a Worker script.
type PlacementMode string

const (
	// PlacementModeOff disables smart placement.
	PlacementModeOff PlacementMode = ""
	// PlacementModeSmart enables smart placement for optimal performance.
	PlacementModeSmart PlacementMode = "smart"
)

// WorkerBinding represents different types of bindings available to Workers.
type WorkerBinding struct {
	// Type specifies the binding type (kv_namespace, wasm_module, text_blob, json_data, etc.)
	Type string `json:"type"`

	// Name is the variable name used in the Worker script to access this binding.
	Name string `json:"name"`

	// NamespaceID for KV namespace bindings.
	// +optional
	NamespaceID *string `json:"namespaceId,omitempty"`

	// Part for WASM module bindings.
	// +optional
	Part *string `json:"part,omitempty"`

	// Text for text blob bindings.
	// +optional
	Text *string `json:"text,omitempty"`

	// JSON for JSON data bindings (as string).
	// +optional
	JSON *string `json:"json,omitempty"`
}

// TailConsumer represents a Worker that consumes logs from another Worker.
type TailConsumer struct {
	// Service is the name of the Worker service that will consume logs.
	Service string `json:"service"`

	// Environment specifies which environment of the service to use.
	// +optional
	Environment *string `json:"environment,omitempty"`

	// Namespace specifies the Workers for Platforms namespace.
	// +optional
	Namespace *string `json:"namespace,omitempty"`
}

// ScriptParameters are the configurable fields of a Worker Script.
type ScriptParameters struct {
	// ScriptName is the name of the Worker script.
	// +immutable
	ScriptName string `json:"scriptName"`

	// Script is the JavaScript/WebAssembly content of the Worker.
	// This can be raw script content or base64 encoded for binary content.
	// +required
	Script string `json:"script"`

	// Module indicates whether this is an ES module Worker (true) or service worker (false).
	// +optional
	Module *bool `json:"module,omitempty"`

	// CompatibilityDate sets the compatibility date for the Worker.
	// Format: YYYY-MM-DD
	// +optional
	CompatibilityDate *string `json:"compatibilityDate,omitempty"`

	// CompatibilityFlags enables specific compatibility flags for the Worker.
	// +optional
	CompatibilityFlags []string `json:"compatibilityFlags,omitempty"`

	// UsageModel specifies the usage model for the Worker.
	// +kubebuilder:validation:Enum=bundled;unbound
	// +optional
	UsageModel *string `json:"usageModel,omitempty"`

	// Bindings defines the bindings available to the Worker script.
	// +optional
	Bindings []WorkerBinding `json:"bindings,omitempty"`

	// PlacementMode controls smart placement for optimal performance.
	// +kubebuilder:validation:Enum="";smart
	// +optional
	Placement *PlacementMode `json:"placement,omitempty"`

	// TailConsumers defines Workers that will consume logs from this Worker.
	// +optional
	TailConsumers []TailConsumer `json:"tailConsumers,omitempty"`

	// LogPush indicates whether to enable log push for this Worker.
	// +optional
	LogPush *bool `json:"logPush,omitempty"`

	// Tags are arbitrary user-defined tags for organizing Workers.
	// +optional
	Tags []string `json:"tags,omitempty"`
}

// ScriptObservation are the observable fields of a Worker Script.
type ScriptObservation struct {
	// ID is the unique identifier of the Worker script.
	ID string `json:"id,omitempty"`

	// ETAG represents the current version/state of the script.
	ETAG string `json:"etag,omitempty"`

	// Size is the size of the Worker script in bytes.
	Size int64 `json:"size,omitempty"`

	// ModifiedOn indicates when the script was last modified.
	ModifiedOn string `json:"modifiedOn,omitempty"`

	// CreatedOn indicates when the script was created.
	CreatedOn string `json:"createdOn,omitempty"`

	// UsageModel indicates the current usage model.
	UsageModel string `json:"usageModel,omitempty"`

	// DeploymentId represents the current deployment identifier.
	DeploymentId string `json:"deploymentId,omitempty"`

	// Pipeline indicates if the script is part of a deployment pipeline.
	Pipeline string `json:"pipeline,omitempty"`

	// LogPush indicates whether log push is enabled.
	LogPush bool `json:"logPush,omitempty"`

	// LastDeployedFrom indicates the deployment source.
	LastDeployedFrom string `json:"lastDeployedFrom,omitempty"`
}

// A ScriptSpec defines the desired state of a Worker Script.
type ScriptSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       ScriptParameters `json:"forProvider"`
}

// A ScriptStatus represents the observed state of a Worker Script.
type ScriptStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          ScriptObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Script represents a Cloudflare Worker script for serverless edge computing.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type Script struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScriptSpec   `json:"spec"`
	Status ScriptStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScriptList contains a list of Worker Script objects.
type ScriptList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Script `json:"items"`

}
