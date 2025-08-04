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

// JobParameters are the configurable fields of a Logpush Job.
type JobParameters struct {
	// Dataset to push logs from. 
	// +kubebuilder:validation:Required
	Dataset string `json:"dataset"`

	// Enabled indicates if the logpush job is enabled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`

	// Kind is the logpush job type.
	// +kubebuilder:validation:Optional
	Kind *string `json:"kind,omitempty"`

	// Name of the logpush job.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// LogpullOptions to configure the logpush behavior.
	// +kubebuilder:validation:Optional
	LogpullOptions *string `json:"logpullOptions,omitempty"`

	// OutputOptions contains output configuration.
	// +kubebuilder:validation:Optional
	OutputOptions *OutputOptions `json:"outputOptions,omitempty"`

	// DestinationConf is the configuration for the destination.
	// +kubebuilder:validation:Required
	DestinationConf string `json:"destinationConf"`

	// Frequency of log pushes.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=high;low
	Frequency *string `json:"frequency,omitempty"`

	// Filter contains filtering configuration.
	// +kubebuilder:validation:Optional
	Filter *JobFilters `json:"filter,omitempty"`

	// MaxUploadBytes is the maximum upload size in bytes.
	// +kubebuilder:validation:Optional
	MaxUploadBytes *int `json:"maxUploadBytes,omitempty"`

	// MaxUploadRecords is the maximum number of records to upload.
	// +kubebuilder:validation:Optional
	MaxUploadRecords *int `json:"maxUploadRecords,omitempty"`

	// MaxUploadIntervalSeconds is the maximum upload interval in seconds.
	// +kubebuilder:validation:Optional
	MaxUploadIntervalSeconds *int `json:"maxUploadIntervalSeconds,omitempty"`
}

// OutputOptions contains output configuration for logpush jobs.
type OutputOptions struct {
	// FieldNames specifies which fields to include in the output.
	// +kubebuilder:validation:Optional
	FieldNames []string `json:"fieldNames,omitempty"`

	// OutputType specifies the output format.
	// +kubebuilder:validation:Optional
	OutputType *string `json:"outputType,omitempty"`

	// BatchPrefix is the prefix for batched outputs.
	// +kubebuilder:validation:Optional
	BatchPrefix *string `json:"batchPrefix,omitempty"`

	// BatchSuffix is the suffix for batched outputs.
	// +kubebuilder:validation:Optional
	BatchSuffix *string `json:"batchSuffix,omitempty"`

	// RecordPrefix is the prefix for each record.
	// +kubebuilder:validation:Optional
	RecordPrefix *string `json:"recordPrefix,omitempty"`

	// RecordSuffix is the suffix for each record.
	// +kubebuilder:validation:Optional
	RecordSuffix *string `json:"recordSuffix,omitempty"`

	// RecordTemplate is the template for each record.
	// +kubebuilder:validation:Optional
	RecordTemplate *string `json:"recordTemplate,omitempty"`

	// RecordDelimiter is the delimiter between records.
	// +kubebuilder:validation:Optional
	RecordDelimiter *string `json:"recordDelimiter,omitempty"`

	// FieldDelimiter is the delimiter between fields.
	// +kubebuilder:validation:Optional
	FieldDelimiter *string `json:"fieldDelimiter,omitempty"`

	// TimestampFormat specifies the timestamp format.
	// +kubebuilder:validation:Optional
	TimestampFormat *string `json:"timestampFormat,omitempty"`

	// SampleRate is the sampling rate for logs.
	// +kubebuilder:validation:Optional
	SampleRate *string `json:"sampleRate,omitempty"`
}

// JobFilters contains filtering configuration for logpush jobs.
type JobFilters struct {
	// Where contains the filter conditions.
	// +kubebuilder:validation:Optional
	Where *JobFilter `json:"where,omitempty"`
}

// JobFilter defines a single filter condition.
type JobFilter struct {
	// Key is the field name to filter on.
	// +kubebuilder:validation:Optional
	Key *string `json:"key,omitempty"`

	// Operator is the comparison operator.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=eq;!eq;lt;leq;gt;geq;startsWith;endsWith;!startsWith;!endsWith;contains;!contains;in;!in
	Operator *string `json:"operator,omitempty"`

	// Value is the value to compare against.
	// +kubebuilder:validation:Optional
	Value *string `json:"value,omitempty"`
}

// JobObservation are the observable fields of a Logpush Job.
type JobObservation struct {
	// ID of the logpush job.
	ID *int `json:"id,omitempty"`

	// Dataset to push logs from.
	Dataset string `json:"dataset,omitempty"`

	// Enabled indicates if the logpush job is enabled.
	Enabled *bool `json:"enabled,omitempty"`

	// Kind is the logpush job type.
	Kind *string `json:"kind,omitempty"`

	// Name of the logpush job.
	Name string `json:"name,omitempty"`

	// LogpullOptions to configure the logpush behavior.
	LogpullOptions *string `json:"logpullOptions,omitempty"`

	// OutputOptions contains output configuration.
	OutputOptions *OutputOptions `json:"outputOptions,omitempty"`

	// DestinationConf is the configuration for the destination.
	DestinationConf string `json:"destinationConf,omitempty"`

	// OwnershipChallenge for destination verification.
	OwnershipChallenge *string `json:"ownershipChallenge,omitempty"`

	// LastComplete timestamp of last successful upload.
	LastComplete *metav1.Time `json:"lastComplete,omitempty"`

	// LastError timestamp of last error.
	LastError *metav1.Time `json:"lastError,omitempty"`

	// ErrorMessage contains the last error message.
	ErrorMessage *string `json:"errorMessage,omitempty"`

	// Frequency of log pushes.
	Frequency *string `json:"frequency,omitempty"`

	// Filter contains filtering configuration.
	Filter *JobFilters `json:"filter,omitempty"`

	// MaxUploadBytes is the maximum upload size in bytes.
	MaxUploadBytes *int `json:"maxUploadBytes,omitempty"`

	// MaxUploadRecords is the maximum number of records to upload.
	MaxUploadRecords *int `json:"maxUploadRecords,omitempty"`

	// MaxUploadIntervalSeconds is the maximum upload interval in seconds.
	MaxUploadIntervalSeconds *int `json:"maxUploadIntervalSeconds,omitempty"`
}

// A JobSpec defines the desired state of a Logpush Job.
type JobSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       JobParameters `json:"forProvider"`
}

// A JobStatus represents the observed state of a Logpush Job.
type JobStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          JobObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Job is a Cloudflare Logpush Job.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:",inline"`

	Spec   JobSpec   `json:"spec"`
	Status JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:",inline"`
	Items           []Job `json:"items"`
}

// Job type metadata.
var (
	JobKind             = "Job"
	JobGroupKind        = schema.GroupKind{Group: Group, Kind: JobKind}
	JobKindAPIVersion   = JobKind + "." + GroupVersion.String()
	JobGroupVersionKind = GroupVersion.WithKind(JobKind)
)