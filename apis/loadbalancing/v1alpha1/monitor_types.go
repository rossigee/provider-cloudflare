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

// LoadBalancerMonitorParameters define the desired state of a Cloudflare Load Balancer Monitor
type LoadBalancerMonitorParameters struct {
	// Account is the account ID where this monitor will be created.
	// Either Account or Zone must be specified, but not both.
	// +optional
	Account *string `json:"account,omitempty"`

	// Zone is the zone ID where this monitor will be created.
	// Either Account or Zone must be specified, but not both.
	// +optional
	Zone *string `json:"zone,omitempty"`

	// Type is the protocol to use for the health check.
	// Valid values: "http", "https", "tcp", "udp_icmp", "icmp_ping", "smtp"
	// +required
	Type string `json:"type"`

	// Description is a human-readable description of the monitor.
	// +optional
	Description *string `json:"description,omitempty"`

	// Method is the HTTP method to use for the health check (only for http/https).
	// Valid values: "GET", "HEAD"
	// +optional
	Method *string `json:"method,omitempty"`

	// Path is the path to use for the health check (only for http/https).
	// +optional
	Path *string `json:"path,omitempty"`

	// Header contains HTTP headers to send with the health check request.
	// +optional
	Header map[string][]string `json:"header,omitempty"`

	// Timeout is the timeout in seconds before marking the health check as failed.
	// Valid range: 1-10 seconds.
	// +optional
	Timeout *int `json:"timeout,omitempty"`

	// Retries is the number of retries to attempt in case of a timeout before marking as failed.
	// Valid range: 0-5 retries.
	// +optional
	Retries *int `json:"retries,omitempty"`

	// Interval is the interval in seconds between health checks.
	// Valid range: 5-3600 seconds.
	// +optional
	Interval *int `json:"interval,omitempty"`

	// ConsecutiveUp is the number of consecutive successful health checks required
	// before marking an origin as healthy.
	// Valid range: 1-20.
	// +optional
	ConsecutiveUp *int `json:"consecutiveUp,omitempty"`

	// ConsecutiveDown is the number of consecutive failed health checks required
	// before marking an origin as unhealthy.
	// Valid range: 1-20.
	// +optional
	ConsecutiveDown *int `json:"consecutiveDown,omitempty"`

	// Port is the port to use for the health check.
	// If not specified, defaults to 80 for HTTP and 443 for HTTPS.
	// +optional
	Port *int `json:"port,omitempty"`

	// ExpectedBody is the expected response body for the health check.
	// Only applies to HTTP/HTTPS monitors.
	// +optional
	ExpectedBody *string `json:"expectedBody,omitempty"`

	// ExpectedCodes is the expected HTTP response code or code range.
	// Examples: "200", "2xx", "200,202,301"
	// Only applies to HTTP/HTTPS monitors.
	// +optional
	ExpectedCodes *string `json:"expectedCodes,omitempty"`

	// FollowRedirects indicates whether to follow redirects.
	// Only applies to HTTP/HTTPS monitors.
	// +optional
	FollowRedirects *bool `json:"followRedirects,omitempty"`

	// AllowInsecure indicates whether to allow insecure connections.
	// Only applies to HTTPS monitors.
	// +optional
	AllowInsecure *bool `json:"allowInsecure,omitempty"`

	// ProbeZone is the zone ID to use for the probe.
	// If not specified, probes will be sent from all Cloudflare data centers.
	// +optional
	ProbeZone *string `json:"probeZone,omitempty"`
}

// LoadBalancerMonitorObservation represents the observed state of a Cloudflare Load Balancer Monitor
type LoadBalancerMonitorObservation struct {
	// ID is the monitor ID.
	ID string `json:"id,omitempty"`

	// CreatedOn is when the monitor was created.
	CreatedOn *string `json:"createdOn,omitempty"`

	// ModifiedOn is when the monitor was last modified.
	ModifiedOn *string `json:"modifiedOn,omitempty"`
}

// LoadBalancerMonitorSpec defines the desired state of LoadBalancerMonitor
type LoadBalancerMonitorSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       LoadBalancerMonitorParameters `json:"forProvider"`
}

// LoadBalancerMonitorStatus defines the observed state of LoadBalancerMonitor
type LoadBalancerMonitorStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          LoadBalancerMonitorObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerMonitor is a managed resource that represents a Cloudflare Load Balancer Monitor
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.atProvider.id"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="INTERVAL",type="string",JSONPath=".spec.forProvider.interval"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
type LoadBalancerMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LoadBalancerMonitorSpec   `json:"spec"`
	Status LoadBalancerMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LoadBalancerMonitorList contains a list of LoadBalancerMonitors
type LoadBalancerMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoadBalancerMonitor `json:"items"`
}