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

// CronTriggerParameters are the configurable fields of a Workers Cron Trigger.
type CronTriggerParameters struct {
	// ScriptName is the name of the Worker script to attach the cron trigger to.
	// +immutable
	ScriptName string `json:"scriptName"`

	// Cron is the cron expression for the schedule.
	// Examples: "0 0 * * *" (daily at midnight), "*/5 * * * *" (every 5 minutes)
	// Documentation: https://developers.cloudflare.com/workers/platform/cron-triggers/
	Cron string `json:"cron"`
}

// CronTriggerObservation are the observable fields of a Workers Cron Trigger.
type CronTriggerObservation struct {
	// ScriptName is the name of the Worker script.
	ScriptName string `json:"scriptName,omitempty"`

	// Cron is the cron expression for the schedule.
	Cron string `json:"cron,omitempty"`

	// CreatedOn is when the cron trigger was created.
	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	// ModifiedOn is when the cron trigger was last modified.
	ModifiedOn *metav1.Time `json:"modifiedOn,omitempty"`
}

// A CronTriggerSpec defines the desired state of a Workers Cron Trigger.
type CronTriggerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CronTriggerParameters `json:"forProvider"`
}

// A CronTriggerStatus represents the observed state of a Workers Cron Trigger.
type CronTriggerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          CronTriggerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CronTrigger represents a scheduled execution trigger for a Cloudflare Worker.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SCRIPT",type="string",JSONPath=".spec.forProvider.scriptName"
// +kubebuilder:printcolumn:name="CRON",type="string",JSONPath=".spec.forProvider.cron"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type CronTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronTriggerSpec   `json:"spec"`
	Status CronTriggerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronTriggerList contains a list of Workers Cron Trigger objects
type CronTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronTrigger `json:"items"`
}