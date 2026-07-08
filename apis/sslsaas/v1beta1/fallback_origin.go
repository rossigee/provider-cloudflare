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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
)


// FallbackOriginParameters represents the settings of a FallbackOrigin
type FallbackOriginParameters struct {
	// Origin for the Fallback Origin.
	// +kubebuilder:validation:Format=hostname
	// +kubebuilder:validation:MaxLength=255
	// +optional
	Origin *string `json:"origin,omitempty"`

	// OriginRef references the Record object this Fallback Origin should point to.
	// +immutable
	// +optional
	OriginRef *xpv1.Reference `json:"originRef,omitempty"`

	// OriginSelector selects the Record object this Fallback Origin should point to.
	// +optional
	OriginSelector *xpv1.Selector `json:"originSelector,omitempty"`

	// ZoneID this Fallback Origin is for.
	// +immutable
	// +optional
	Zone *string `json:"zone,omitempty"`

	// ZoneRef references the zone object this Fallback Origin is for.
	// +immutable
	// +optional
	ZoneRef *xpv1.Reference `json:"zoneRef,omitempty"`

	// ZoneSelector selects the zone object this Fallback Origin is for.
	// +optional
	ZoneSelector *xpv1.Selector `json:"zoneSelector,omitempty"`
}

// FallbackOriginObservation are the observable fields of a Fallback Origin.
type FallbackOriginObservation struct {
	// Status of the fallback origin and if its completed deployment
	Status string `json:"status,omitempty"`

	// Errors if there any of the fallback origin
	Errors []string `json:"errors,omitempty"`
}

// A FallbackOriginSpec defines the desired state of a Fallback Origin.
type FallbackOriginSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       FallbackOriginParameters `json:"forProvider"`
}

// A FallbackOriginStatus represents the observed state of a Fallback Origin.
type FallbackOriginStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          FallbackOriginObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A FallbackOrigin is a fallback origin required to use SSL for SaaS.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type FallbackOrigin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FallbackOriginSpec   `json:"spec"`
	Status FallbackOriginStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FallbackOriginList contains a list of FallbackOrigin
type FallbackOriginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FallbackOrigin `json:"items"`
}
}

// ResolveReferences of this Fallback Origin
func (dr *FallbackOrigin) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, dr)

	// Resolve spec.forProvider.origin to FQDN from DNS Record
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(dr.Spec.ForProvider.Origin),
		Reference:    dr.Spec.ForProvider.OriginRef,
		Selector:     dr.Spec.ForProvider.OriginSelector,
		To:           reference.To{Managed: &dnsv1beta1.Record{}, List: &dnsv1beta1.RecordList{}},
		Extract:      dnsv1beta1.RecordFQDN(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.origin")
	}
	dr.Spec.ForProvider.Origin = reference.ToPtrValue(rsp.ResolvedValue)
	dr.Spec.ForProvider.OriginRef = rsp.ResolvedReference

	// Resolve spec.forProvider.zone
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(dr.Spec.ForProvider.Zone),
		Reference:    dr.Spec.ForProvider.ZoneRef,
		Selector:     dr.Spec.ForProvider.ZoneSelector,
		To:           reference.To{Managed: &zonev1beta1.Zone{}, List: &zonev1beta1.ZoneList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.zone")
	}
	dr.Spec.ForProvider.Zone = reference.ToPtrValue(rsp.ResolvedValue)
	dr.Spec.ForProvider.ZoneRef = rsp.ResolvedReference

	return nil
}
