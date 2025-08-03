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

// LoadBalancerParameters define the desired state of a Cloudflare Load Balancer
type LoadBalancerParameters struct {
	// Zone is the zone ID where this load balancer will be created.
	// Load balancers are zone-scoped resources.
	// +required
	Zone string `json:"zone"`

	// Name is the DNS name for this load balancer.
	// +optional
	Name *string `json:"name,omitempty"`

	// Description is a human-readable description of the load balancer.
	// +optional
	Description *string `json:"description,omitempty"`

	// TTL is the DNS TTL for the load balancer.
	// +optional
	TTL *int `json:"ttl,omitempty"`

	// FallbackPool is the pool ID to use when all other pools are unhealthy.
	// +optional
	FallbackPool *string `json:"fallbackPool,omitempty"`

	// FallbackPoolRef is a reference to a LoadBalancerPool resource to use as fallback.
	// +optional
	FallbackPoolRef *xpv1.Reference `json:"fallbackPoolRef,omitempty"`

	// FallbackPoolSelector selects a reference to a LoadBalancerPool resource as fallback.
	// +optional
	FallbackPoolSelector *xpv1.Selector `json:"fallbackPoolSelector,omitempty"`

	// DefaultPools is the list of pool IDs ordered by their failover priority.
	// +optional
	DefaultPools []string `json:"defaultPools,omitempty"`

	// DefaultPoolRefs is a list of references to LoadBalancerPool resources.
	// +optional
	DefaultPoolRefs []xpv1.Reference `json:"defaultPoolRefs,omitempty"`

	// DefaultPoolSelector selects references to LoadBalancerPool resources.
	// +optional
	DefaultPoolSelector *xpv1.Selector `json:"defaultPoolSelector,omitempty"`

	// RegionPools maps regions to pool lists for geo-steering.
	// +optional
	RegionPools map[string][]string `json:"regionPools,omitempty"`

	// PopPools maps Cloudflare PoPs to pool lists for pop-steering.
	// +optional
	PopPools map[string][]string `json:"popPools,omitempty"`

	// CountryPools maps countries to pool lists for country-steering.
	// +optional
	CountryPools map[string][]string `json:"countryPools,omitempty"`

	// Proxied indicates whether traffic should be proxied through Cloudflare.
	// +optional
	Proxied *bool `json:"proxied,omitempty"`

	// Enabled indicates whether the load balancer is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// SessionAffinity controls session stickiness.
	// Valid values: "none", "cookie", "ip_cookie"
	// +optional
	SessionAffinity *string `json:"sessionAffinity,omitempty"`

	// SessionAffinityTTL is the TTL for session affinity in seconds.
	// +optional
	SessionAffinityTTL *int `json:"sessionAffinityTtl,omitempty"`

	// SessionAffinityAttributes contains session affinity configuration.
	// +optional
	SessionAffinityAttributes *SessionAffinityAttributes `json:"sessionAffinityAttributes,omitempty"`

	// Rules contains traffic steering rules for advanced routing.
	// +optional
	Rules []LoadBalancerRule `json:"rules,omitempty"`

	// RandomSteering contains random steering configuration.
	// +optional
	RandomSteering *RandomSteering `json:"randomSteering,omitempty"`

	// AdaptiveRouting contains adaptive routing configuration.
	// +optional
	AdaptiveRouting *AdaptiveRouting `json:"adaptiveRouting,omitempty"`

	// LocationStrategy contains location strategy configuration.
	// +optional
	LocationStrategy *LocationStrategy `json:"locationStrategy,omitempty"`

	// SteeringPolicy controls pool selection logic.
	// Valid values: "off", "geo", "dynamic_latency", "random", "proximity", 
	// "least_outstanding_requests", "least_connections"
	// +optional
	SteeringPolicy *string `json:"steeringPolicy,omitempty"`
}

// SessionAffinityAttributes contains session affinity configuration
type SessionAffinityAttributes struct {
	// SameSite controls the SameSite attribute for session affinity cookies.
	// Valid values: "Auto", "Lax", "None", "Strict"
	// +optional
	SameSite *string `json:"sameSite,omitempty"`

	// Secure indicates whether the session affinity cookie should be secure.
	// +optional
	Secure *string `json:"secure,omitempty"`

	// DrainDuration is how long to honor session affinity when a pool becomes unhealthy.
	// +optional
	DrainDuration *int `json:"drainDuration,omitempty"`

	// ZeroDowntimeFailover controls zero-downtime failover.
	// Valid values: "none", "sticky", "temporary"
	// +optional
	ZeroDowntimeFailover *string `json:"zeroDowntimeFailover,omitempty"`

	// Headers contains headers for session affinity.
	// +optional
	Headers []string `json:"headers,omitempty"`

	// RequireAllHeaders indicates whether all headers must match for session affinity.
	// +optional
	RequireAllHeaders *bool `json:"requireAllHeaders,omitempty"`
}

// LoadBalancerRule represents a traffic steering rule
type LoadBalancerRule struct {
	// Name is a human-readable name for the rule.
	// +required
	Name string `json:"name"`

	// Condition is the expression that determines when this rule applies.
	// +required
	Condition string `json:"condition"`

	// Priority controls the order of rule execution (lower values execute first).
	// +required
	Priority int `json:"priority"`

	// Disabled indicates whether the rule is disabled.
	// +optional
	Disabled *bool `json:"disabled,omitempty"`

	// Terminates indicates whether to stop processing rules after this one.
	// +optional
	Terminates *bool `json:"terminates,omitempty"`

	// FixedResponse contains a fixed response configuration.
	// +optional
	FixedResponse *LoadBalancerFixedResponse `json:"fixedResponse,omitempty"`

	// Overrides contains pool and routing overrides for this rule.
	// +optional
	Overrides *LoadBalancerRuleOverrides `json:"overrides,omitempty"`
}

// LoadBalancerFixedResponse contains fixed response configuration
type LoadBalancerFixedResponse struct {
	// MessageBody is the response body content.
	// +optional
	MessageBody *string `json:"messageBody,omitempty"`

	// StatusCode is the HTTP status code to return.
	// +optional
	StatusCode *int `json:"statusCode,omitempty"`

	// ContentType is the Content-Type header value.
	// +optional
	ContentType *string `json:"contentType,omitempty"`

	// Location is the Location header value for redirects.
	// +optional
	Location *string `json:"location,omitempty"`
}

// LoadBalancerRuleOverrides contains rule-specific overrides
type LoadBalancerRuleOverrides struct {
	// SessionAffinity overrides the session affinity setting.
	// +optional
	SessionAffinity *string `json:"sessionAffinity,omitempty"`

	// SessionAffinityTTL overrides the session affinity TTL.
	// +optional
	SessionAffinityTTL *int `json:"sessionAffinityTtl,omitempty"`

	// SessionAffinityAttributes overrides session affinity attributes.
	// +optional
	SessionAffinityAttributes *SessionAffinityAttributes `json:"sessionAffinityAttributes,omitempty"`

	// TTL overrides the DNS TTL.
	// +optional
	TTL *int `json:"ttl,omitempty"`

	// SteeringPolicy overrides the steering policy.
	// +optional
	SteeringPolicy *string `json:"steeringPolicy,omitempty"`

	// FallbackPool overrides the fallback pool.
	// +optional
	FallbackPool *string `json:"fallbackPool,omitempty"`

	// DefaultPools overrides the default pools.
	// +optional
	DefaultPools []string `json:"defaultPools,omitempty"`

	// PopPools overrides the PoP pools.
	// +optional
	PopPools map[string][]string `json:"popPools,omitempty"`

	// RegionPools overrides the region pools.
	// +optional
	RegionPools map[string][]string `json:"regionPools,omitempty"`

	// CountryPools overrides the country pools.
	// +optional
	CountryPools map[string][]string `json:"countryPools,omitempty"`

	// RandomSteering overrides random steering configuration.
	// +optional
	RandomSteering *RandomSteering `json:"randomSteering,omitempty"`

	// AdaptiveRouting overrides adaptive routing configuration.
	// +optional
	AdaptiveRouting *AdaptiveRouting `json:"adaptiveRouting,omitempty"`

	// LocationStrategy overrides location strategy configuration.
	// +optional
	LocationStrategy *LocationStrategy `json:"locationStrategy,omitempty"`
}

// RandomSteering configures pool weights for random steering
type RandomSteering struct {
	// DefaultWeight is the default weight for pools not specified in PoolWeights.
	// +optional
	DefaultWeight *string `json:"defaultWeight,omitempty"`

	// PoolWeights maps pool IDs to their weights.
	// +optional
	PoolWeights map[string]string `json:"poolWeights,omitempty"`
}

// AdaptiveRouting controls adaptive routing features
type AdaptiveRouting struct {
	// FailoverAcrossPools enables failover across pools when no healthy origins exist.
	// +optional
	FailoverAcrossPools *bool `json:"failoverAcrossPools,omitempty"`
}

// LocationStrategy controls how client location is determined
type LocationStrategy struct {
	// Mode determines how to get the client IP for location-based steering.
	// Valid values: "pop", "resolver_ip"
	// +optional
	Mode *string `json:"mode,omitempty"`

	// PreferECSRegion controls whether to prefer the ECS region.
	// +optional
	PreferECSRegion *string `json:"preferECSRegion,omitempty"`
}

// LoadBalancerObservation represents the observed state of a Cloudflare Load Balancer
type LoadBalancerObservation struct {
	// ID is the load balancer ID.
	ID string `json:"id,omitempty"`

	// CreatedOn is when the load balancer was created.
	CreatedOn *string `json:"createdOn,omitempty"`

	// ModifiedOn is when the load balancer was last modified.
	ModifiedOn *string `json:"modifiedOn,omitempty"`
}

// LoadBalancerSpec defines the desired state of LoadBalancer
type LoadBalancerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       LoadBalancerParameters `json:"forProvider"`
}

// LoadBalancerStatus defines the observed state of LoadBalancer
type LoadBalancerStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LoadBalancerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancer is a managed resource that represents a Cloudflare Load Balancer
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="STEERING",type="string",JSONPath=".spec.forProvider.steeringPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type LoadBalancer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerSpec   `json:"spec"`
	Status LoadBalancerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerList contains a list of LoadBalancers
type LoadBalancerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancer `json:"items"`
}