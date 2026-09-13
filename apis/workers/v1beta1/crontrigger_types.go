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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              CronTriggerParameters `json:"forProvider"`
}

// A CronTriggerStatus represents the observed state of a Workers Cron Trigger.
type CronTriggerStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 CronTriggerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A CronTrigger represents a scheduled execution trigger for a Cloudflare Worker.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SCRIPT",type="string",JSONPath=".spec.forProvider.scriptName"
// +kubebuilder:printcolumn:name="CRON",type="string",JSONPath=".spec.forProvider.cron"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
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

// GetCondition gets the condition from the resource status.
func (mg *CronTrigger) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions sets the conditions on the resource status.
func (mg *CronTrigger) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies gets the management policies for the resource.
func (mg *CronTrigger) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies sets the management policies for the resource.
func (mg *CronTrigger) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = mp
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *CronTrigger) DeepCopyObject() runtime.Object {
	out := &CronTrigger{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *CronTrigger) DeepCopyInto(out *CronTrigger) {
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// GetItems returns the list items.
func (l *CronTriggerList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *CronTriggerList) DeepCopyObject() runtime.Object {
	out := &CronTriggerList{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *CronTriggerList) DeepCopyInto(out *CronTriggerList) {
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	out.Items = append([]CronTrigger(nil), in.Items...)
}
