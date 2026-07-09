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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)


// LoadBalancerOrigin represents an origin server in a load balancer pool
type LoadBalancerOrigin struct {
	// Name is the name of the origin.
	// +required
	Name string `json:"name"`

	// Address is the IP address or hostname of the origin.
	// +required
	Address string `json:"address"`

	// Enabled indicates whether this origin is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Weight is the weight of this origin in load balancing decisions.
	// +optional
	Weight *float64 `json:"weight,omitempty"`

	// Header contains HTTP headers to send to this origin.
	// +optional
	Header map[string][]string `json:"header,omitempty"`
}

// LoadBalancerPoolParameters define the desired state of a Cloudflare Load Balancer Pool
type LoadBalancerPoolParameters struct {
	// Name is the name of the pool.
	// +optional
	Name *string `json:"name,omitempty"`

	// Description is a human-readable description of the pool.
	// +optional
	Description *string `json:"description,omitempty"`

	// Enabled indicates whether the pool is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// MinimumOrigins is the minimum number of healthy origins required for the pool to be considered healthy.
	// +optional
	MinimumOrigins *int `json:"minimumOrigins,omitempty"`

	// Monitor is the ID of the monitor to use for health checking.
	// +optional
	Monitor *string `json:"monitor,omitempty"`

	// MonitorRef is a reference to a LoadBalancerMonitor resource.
	// +optional
	MonitorRef *xpv1.Reference `json:"monitorRef,omitempty"`

	// MonitorSelector selects a reference to a LoadBalancerMonitor resource.
	// +optional
	MonitorSelector *xpv1.Selector `json:"monitorSelector,omitempty"`

	// Origins is the list of origins within this pool.
	// +required
	Origins []LoadBalancerOrigin `json:"origins"`

	// NotificationEmail is the email address to send health status notifications to.
	// +optional
	NotificationEmail *string `json:"notificationEmail,omitempty"`

	// OriginSteering controls how origins are selected within the pool.
	// +optional
	OriginSteering *OriginSteering `json:"originSteering,omitempty"`

	// CheckRegions defines the regions from which to run health checks.
	// +optional
	CheckRegions []string `json:"checkRegions,omitempty"`

	// Latitude is the latitude for the pool.
	// +optional
	Latitude *float64 `json:"latitude,omitempty"`

	// Longitude is the longitude for the pool.
	// +optional
	Longitude *float64 `json:"longitude,omitempty"`
}

// OriginSteering controls how origins are selected within a pool
type OriginSteering struct {
	// Policy defines the origin steering policy.
	// Valid values: "random", "hash", "least_outstanding_requests"
	// +optional
	Policy *string `json:"policy,omitempty"`
}

// LoadBalancerPoolObservation represents the observed state of a Cloudflare Load Balancer Pool
type LoadBalancerPoolObservation struct {
	// ID is the pool ID.
	ID string `json:"id,omitempty"`

	// CreatedOn is when the pool was created.
	CreatedOn *string `json:"createdOn,omitempty"`

	// ModifiedOn is when the pool was last modified.
	ModifiedOn *string `json:"modifiedOn,omitempty"`

	// Healthy indicates whether the pool is healthy.
	Healthy *bool `json:"healthy,omitempty"`
}

// LoadBalancerPoolSpec defines the desired state of LoadBalancerPool
type LoadBalancerPoolSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       LoadBalancerPoolParameters `json:"forProvider"`
}

// LoadBalancerPoolStatus defines the observed state of LoadBalancerPool
type LoadBalancerPoolStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          LoadBalancerPoolObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerPool is a managed resource that represents a Cloudflare Load Balancer Pool
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="HEALTHY",type="boolean",JSONPath=".status.atProvider.healthy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
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
