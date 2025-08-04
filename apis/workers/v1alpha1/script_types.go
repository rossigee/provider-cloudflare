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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
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
	Script string `json:"script"`

	// Module indicates if this is an ES Module script (true) or Service Worker (false).
	// +optional
	Module *bool `json:"module,omitempty"`

	// CompatibilityDate sets the Worker runtime version (format: YYYY-MM-DD).
	// Documentation: https://developers.cloudflare.com/workers/platform/compatibility-dates/
	// +optional
	CompatibilityDate *string `json:"compatibilityDate,omitempty"`

	// CompatibilityFlags enables or disables specific Workers runtime features.
	// Documentation: https://developers.cloudflare.com/workers/platform/compatibility-dates/#compatibility-flags
	// +optional
	CompatibilityFlags []string `json:"compatibilityFlags,omitempty"`

	// Bindings provide access to KV namespaces, WASM modules, and other resources.
	// +optional
	Bindings []WorkerBinding `json:"bindings,omitempty"`

	// PlacementMode controls where the Worker runs for optimal performance.
	// +optional
	PlacementMode *PlacementMode `json:"placementMode,omitempty"`

	// Logpush enables Worker log collection and forwarding.
	// Documentation: https://developers.cloudflare.com/workers/platform/logpush/
	// +optional
	Logpush *bool `json:"logpush,omitempty"`

	// TailConsumers specifies Workers that will consume logs from this Worker.
	// Documentation: https://developers.cloudflare.com/workers/platform/tail-workers/
	// +optional
	TailConsumers []TailConsumer `json:"tailConsumers,omitempty"`

	// Tags help manage Workers at scale.
	// Documentation: https://developers.cloudflare.com/cloudflare-for-platforms/workers-for-platforms/platform/tags/
	// +optional
	Tags []string `json:"tags,omitempty"`

	// DispatchNamespace uploads the Worker to a Workers for Platforms dispatch namespace.
	// +optional
	DispatchNamespace *string `json:"dispatchNamespace,omitempty"`
}

// ScriptObservation are the observable fields of a Worker Script.
type ScriptObservation struct {
	// ID is the unique identifier for the Worker script.
	ID string `json:"id,omitempty"`

	// ETAG is the entity tag for the Worker script.
	ETAG string `json:"etag,omitempty"`

	// Size is the size of the Worker script in bytes.
	Size int `json:"size,omitempty"`

	// CreatedOn is when the Worker script was created.
	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	// ModifiedOn is when the Worker script was last modified.
	ModifiedOn *metav1.Time `json:"modifiedOn,omitempty"`

	// LastDeployedFrom indicates the source of the last deployment.
	LastDeployedFrom *string `json:"lastDeployedFrom,omitempty"`

	// DeploymentID is the unique identifier for the current deployment.
	DeploymentID *string `json:"deploymentId,omitempty"`

	// PlacementStatus shows the current placement status.
	PlacementStatus *string `json:"placementStatus,omitempty"`

	// PipelineHash is a hash of the Worker's processing pipeline.
	PipelineHash *string `json:"pipelineHash,omitempty"`

	// UsageModel indicates the billing model for the Worker.
	UsageModel *string `json:"usageModel,omitempty"`
}

// A ScriptSpec defines the desired state of a Worker Script.
type ScriptSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ScriptParameters `json:"forProvider"`
}

// A ScriptStatus represents the observed state of a Worker Script.
type ScriptStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ScriptObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Script represents a Cloudflare Worker script.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SCRIPT",type="string",JSONPath=".spec.forProvider.scriptName"
// +kubebuilder:printcolumn:name="SIZE",type="integer",JSONPath=".status.atProvider.size"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Script struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScriptSpec   `json:"spec"`
	Status ScriptStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScriptList contains a list of Worker Script objects
type ScriptList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Script `json:"items"`
}

