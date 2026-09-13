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
	"context"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	zonev1beta1 "github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FilterParameters are the configurable fields of a Filter.
type FilterParameters struct {
	// Expression is the filter expression used to match traffic.
	Expression string `json:"expression"`

	// Description is a human readable description of this rule.
	// +kubebuilder:validation:MaxLength=500
	// +optional
	Description *string `json:"description,omitempty"`

	// Paused indicates if this rule is paused or not.
	// +optional
	Paused *bool `json:"paused,omitempty"`

	// ZoneID this Firewall Rule is for.
	// +immutable
	// +optional
	Zone *string `json:"zone,omitempty"`

	// ZoneRef references the zone object this Firewall Rule is for.
	// +immutable
	// +optional
	ZoneRef *xpv1.Reference `json:"zoneRef,omitempty"`

	// ZoneSelector selects the zone object this Firewall Rule is for.
	// +immutable
	// +optional
	ZoneSelector *xpv1.Selector `json:"zoneSelector,omitempty"`
}

// FilterObservation is the observable fields of a Filter.
type FilterObservation struct{}

// A FilterSpec defines the desired state of a Filter.
type FilterSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider                     FilterParameters `json:"forProvider"`
}

// A FilterStatus represents the observed state of a Filter.
type FilterStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 FilterObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Filter is a matching expression that can be referenced by one or more
// firewall rules.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type Filter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FilterSpec   `json:"spec"`
	Status FilterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FilterList contains a list of Filter
type FilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Filter `json:"items"`
}

// ResolveReferences of this Filter
func (f *Filter) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, f)

	// Resolve spec.forProvider.zone
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(f.Spec.ForProvider.Zone),
		Reference:    f.Spec.ForProvider.ZoneRef,
		Selector:     f.Spec.ForProvider.ZoneSelector,
		To:           reference.To{Managed: &zonev1beta1.Zone{}, List: &zonev1beta1.ZoneList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.zone")
	}
	f.Spec.ForProvider.Zone = reference.ToPtrValue(rsp.ResolvedValue)
	f.Spec.ForProvider.ZoneRef = rsp.ResolvedReference
	return nil
}

// GetCondition gets the condition from the resource status.
func (mg *Filter) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions sets the conditions on the resource status.
func (mg *Filter) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies gets the management policies for the resource.
func (mg *Filter) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies sets the management policies for the resource.
func (mg *Filter) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = mp
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *Filter) DeepCopyObject() runtime.Object {
	out := &Filter{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *Filter) DeepCopyInto(out *Filter) {
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// GetItems returns the list items.
func (l *FilterList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}

// DeepCopyObject returns a deep copy of this object as runtime.Object.
func (in *FilterList) DeepCopyObject() runtime.Object {
	out := &FilterList{}
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto fills DeepCopy receiver with DeepCopy of the provided receiver.
func (in *FilterList) DeepCopyInto(out *FilterList) {
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	out.Items = append([]Filter(nil), in.Items...)
}

// GetProviderConfigReference returns the ProviderConfig reference.
func (mg *Filter) GetProviderConfigReference() *xpv1.Reference {
	if mg.Spec.ProviderConfigReference != nil && mg.Spec.ProviderConfigReference.Name != "" {
		return &xpv1.Reference{Name: mg.Spec.ProviderConfigReference.Name}
	}
	return nil
}
