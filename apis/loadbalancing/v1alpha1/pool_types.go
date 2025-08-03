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

// LoadBalancerPoolParameters define the desired state of a Cloudflare Load Balancer Pool
type LoadBalancerPoolParameters struct {
	// Account is the account ID where this pool will be created.
	// Either Account or Zone must be specified, but not both.
	// +optional
	Account *string `json:"account,omitempty"`

	// Zone is the zone ID where this pool will be created.
	// Either Account or Zone must be specified, but not both.
	// +optional
	Zone *string `json:"zone,omitempty"`

	// Name is the name of the pool.
	// +required
	Name string `json:"name"`

	// Description is a human-readable description of the pool.
	// +optional
	Description *string `json:"description,omitempty"`

	// Enabled indicates whether the pool is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// MinimumOrigins is the minimum number of origins that must be healthy
	// for the pool to serve traffic.
	// +optional
	MinimumOrigins *int `json:"minimumOrigins,omitempty"`

	// Monitor is the ID of the monitor to use for health checks.
	// +optional
	Monitor *string `json:"monitor,omitempty"`

	// MonitorRef is a reference to a LoadBalancerMonitor resource.
	// +optional
	MonitorRef *xpv1.Reference `json:"monitorRef,omitempty"`

	// MonitorSelector selects a reference to a LoadBalancerMonitor resource.
	// +optional
	MonitorSelector *xpv1.Selector `json:"monitorSelector,omitempty"`

	// Origins is the list of origin servers in this pool.
	// +required
	Origins []LoadBalancerOrigin `json:"origins"`

	// NotificationEmail is the email address to notify when the pool status changes.
	// +optional
	NotificationEmail *string `json:"notificationEmail,omitempty"`

	// Latitude is the latitude for proximity-based load balancing.
	// +optional
	Latitude *string `json:"latitude,omitempty"`

	// Longitude is the longitude for proximity-based load balancing.
	// +optional
	Longitude *string `json:"longitude,omitempty"`

	// LoadShedding contains load shedding configuration.
	// +optional
	LoadShedding *LoadBalancerLoadShedding `json:"loadShedding,omitempty"`

	// OriginSteering controls origin selection for new sessions.
	// +optional
	OriginSteering *LoadBalancerOriginSteering `json:"originSteering,omitempty"`

	// CheckRegions defines the geographic regions from where to run health checks.
	// Valid values: "WNAM", "ENAM", "WEU", "EEU", "NSAM", "SSAM", "OC", "ME", "NAF", "SAF", "IN", "SEAS", "NEAS"
	// If empty, health checks will be performed from all regions.
	// +optional
	CheckRegions []string `json:"checkRegions,omitempty"`
}

// LoadBalancerOrigin represents a single origin server in a pool
type LoadBalancerOrigin struct {
	// Name is a human-readable name for this origin.
	// +required
	Name string `json:"name"`

	// Address is the IP address or hostname of the origin server.
	// +required
	Address string `json:"address"`

	// Enabled indicates whether this origin is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Weight is the relative weight of this origin for load balancing.
	// Higher weights receive more traffic.
	// +optional
	Weight *string `json:"weight,omitempty"`

	// Header contains HTTP headers to send to this origin.
	// +optional
	Header map[string][]string `json:"header,omitempty"`

	// VirtualNetworkID is the ID of the virtual network this origin belongs to.
	// +optional
	VirtualNetworkID *string `json:"virtualNetworkId,omitempty"`
}

// LoadBalancerOriginSteering controls origin selection for new sessions
type LoadBalancerOriginSteering struct {
	// Policy determines the origin selection algorithm.
	// Valid values: "random", "hash", "least_outstanding_requests", "least_connections"
	// +optional
	Policy *string `json:"policy,omitempty"`
}

// LoadBalancerLoadShedding contains load shedding configuration
type LoadBalancerLoadShedding struct {
	// DefaultPercent is the default load shedding percentage.
	// +optional
	DefaultPercent *string `json:"defaultPercent,omitempty"`

	// DefaultPolicy is the default load shedding policy.
	// Valid values: "random", "hash"
	// +optional
	DefaultPolicy *string `json:"defaultPolicy,omitempty"`

	// SessionPercent is the load shedding percentage for session affinity.
	// +optional
	SessionPercent *string `json:"sessionPercent,omitempty"`

	// SessionPolicy is the load shedding policy for session affinity.
	// Valid values: "random", "hash"
	// +optional
	SessionPolicy *string `json:"sessionPolicy,omitempty"`
}

// LoadBalancerPoolObservation represents the observed state of a Cloudflare Load Balancer Pool
type LoadBalancerPoolObservation struct {
	// ID is the pool ID.
	ID string `json:"id,omitempty"`

	// CreatedOn is when the pool was created.
	CreatedOn *string `json:"createdOn,omitempty"`

	// ModifiedOn is when the pool was last modified.
	ModifiedOn *string `json:"modifiedOn,omitempty"`

	// Healthy indicates whether the pool is currently healthy.
	Healthy *bool `json:"healthy,omitempty"`
}

// LoadBalancerPoolSpec defines the desired state of LoadBalancerPool
type LoadBalancerPoolSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       LoadBalancerPoolParameters `json:"forProvider"`
}

// LoadBalancerPoolStatus defines the observed state of LoadBalancerPool
type LoadBalancerPoolStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LoadBalancerPoolObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerPool is a managed resource that represents a Cloudflare Load Balancer Pool
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="HEALTHY",type="string",JSONPath=".status.atProvider.healthy"
// +kubebuilder:printcolumn:name="ORIGINS",type="string",JSONPath=".spec.forProvider.origins[*].name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type LoadBalancerPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerPoolSpec   `json:"spec"`
	Status LoadBalancerPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerPoolList contains a list of LoadBalancerPools
type LoadBalancerPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancerPool `json:"items"`
}