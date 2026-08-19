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
	"context"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	zonev1beta1 "github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteParameters are the configurable fields of a DNS Route.
type RouteParameters struct {
	// Pattern is the URL pattern of the route.
	Pattern string `json:"pattern"`

	// Script is the name of the worker script.
	// +optional
	Script *string `json:"script,omitempty"`

	// ZoneID this Worker Route is managed on.
	// +immutable
	// +optional
	Zone *string `json:"zone,omitempty"`

	// ZoneRef references the Zone object this Worker Route is managed on.
	// +immutable
	// +optional
	ZoneRef *xpv1.Reference `json:"zoneRef,omitempty"`

	// ZoneSelector selects the Zone object this Worker Route is managed on.
	// +immutable
	// +optional
	ZoneSelector *xpv1.Selector `json:"zoneSelector,omitempty"`
}

// RouteObservation is the observable fields of a Worker Route.
type RouteObservation struct{}

// A RouteSpec defines the desired state of a Worker Route.
type RouteSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     RouteParameters `json:"forProvider"`
}

// A RouteStatus represents the observed state of a Worker Route.
type RouteStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 RouteObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Route represents a single Worker Route managed on a Zone.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="PATTERN",type="string",JSONPath=".spec.forProvider.pattern"
// +kubebuilder:printcolumn:name="SCRIPT",type="string",JSONPath=".spec.forProvider.script"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type Route struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteSpec   `json:"spec"`
	Status RouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RouteList contains a list of Worker Route objects
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Route `json:"items"`
}

// RoutePattern resolves the pattern field out of a Worker Route's
// spec. This may be used by other resources to reference the route.
func RoutePattern() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		r, ok := mg.(*Route)
		if !ok {
			return ""
		}
		return r.Spec.ForProvider.Pattern
	}
}

// ResolveReferences resolves references to the Zone that this Worker Route
// is managed on.
func (wr *Route) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, wr)

	// Resolve spec.forProvider.zone
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(wr.Spec.ForProvider.Zone),
		Reference:    wr.Spec.ForProvider.ZoneRef,
		Selector:     wr.Spec.ForProvider.ZoneSelector,
		To:           reference.To{Managed: &zonev1beta1.Zone{}, List: &zonev1beta1.ZoneList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.zone")
	}
	wr.Spec.ForProvider.Zone = reference.ToPtrValue(rsp.ResolvedValue)
	wr.Spec.ForProvider.ZoneRef = rsp.ResolvedReference

	return nil
}
