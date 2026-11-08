/*
Copyright 2026 The Crossplane Authors.

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
)

// GatewayParameters are the configurable fields of an AI Gateway.
type GatewayParameters struct {
	// AccountID is the Cloudflare account that owns the gateway.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="accountId is immutable"
	AccountID string `json:"accountId"`

	// GatewayID is the identifier of the gateway, which is also its name in the API.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="gatewayId is immutable"
	GatewayID string `json:"gatewayId"`

	// CacheInvalidateOnUpdate clears the gateway cache when its configuration changes.
	CacheInvalidateOnUpdate bool `json:"cacheInvalidateOnUpdate"`

	// CacheTTL is the cache time-to-live, in seconds.
	CacheTTL int64 `json:"cacheTTL"`

	// CollectLogs enables request log collection for the gateway.
	CollectLogs bool `json:"collectLogs"`

	// RateLimitingInterval is the rate limiting window, in seconds.
	RateLimitingInterval int64 `json:"rateLimitingInterval"`

	// RateLimitingLimit is the maximum number of requests allowed per rate limiting interval.
	RateLimitingLimit int64 `json:"rateLimitingLimit"`

	// RateLimitingTechnique is the rate limiting algorithm.
	// +kubebuilder:validation:Enum=fixed;sliding
	// +optional
	RateLimitingTechnique *string `json:"rateLimitingTechnique,omitempty"`

	// Authentication requires a cf-aig-authorization token on requests to the gateway.
	// +optional
	Authentication *bool `json:"authentication,omitempty"`
}

// GatewayObservation are the observable fields of an AI Gateway.
type GatewayObservation struct {
	// ID is the gateway identifier.
	ID string `json:"id,omitempty"`

	// CreatedAt is the time the gateway was created, in RFC 3339 format.
	CreatedAt string `json:"createdAt,omitempty"`

	// ModifiedAt is the time the gateway was last modified, in RFC 3339 format.
	ModifiedAt string `json:"modifiedAt,omitempty"`
}

// A GatewaySpec defines the desired state of an AI Gateway.
type GatewaySpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              GatewayParameters `json:"forProvider"`
}

// A GatewayStatus represents the observed state of an AI Gateway.
type GatewayStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 GatewayObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Gateway represents a Cloudflare AI Gateway.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="GATEWAY",type="string",JSONPath=".spec.forProvider.gatewayId"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec"`
	Status GatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of AI Gateway objects.
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}

// GetCondition gets the condition from the resource status.
func (mg *Gateway) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions sets the conditions on the resource status.
func (mg *Gateway) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies gets the management policies for the resource.
func (mg *Gateway) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies sets the management policies for the resource.
func (mg *Gateway) SetManagementPolicies(mp xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = mp
}

// GetItems returns the list items.
func (l *GatewayList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
