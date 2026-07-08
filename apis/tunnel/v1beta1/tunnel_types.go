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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/crossplane/crossplane/apis/v2/core/v2"
)


// TunnelParameters define the desired state of a Cloudflare Tunnel.
type TunnelParameters struct {
	// AccountID is the account ID where this tunnel will be created.
	// +required
	AccountID string `json:"accountId"`

	// Name is the name of the tunnel.
	// +required
	Name string `json:"name"`

	// Secret is the secret key for the tunnel.
	// +required
	Secret string `json:"secret"`

	// ConfigSrc is the configuration source for the tunnel.
	// +optional
	ConfigSrc *TunnelConfigSrc `json:"configSrc,omitempty"`
}

// TunnelConfigSrc defines the configuration source for the tunnel.
type TunnelConfigSrc struct {
	// CloudflareAccess is the Cloudflare Access configuration.
	// +optional
	CloudflareAccess *TunnelCloudflareAccess `json:"cloudflareAccess,omitempty"`

	// Ingress is a list of ingress rules.
	// +optional
	Ingress []TunnelIngress `json:"ingress,omitempty"`

	// OriginRequest is the origin request configuration.
	// +optional
	OriginRequest *TunnelOriginRequest `json:"originRequest,omitempty"`

	// WarpRouting is the WARP routing configuration.
	// +optional
	WarpRouting *TunnelWarpRouting `json:"warpRouting,omitempty"`
}

// TunnelCloudflareAccess defines Cloudflare Access configuration.
type TunnelCloudflareAccess struct {
	// TeamName is the team name for Cloudflare Access.
	// +optional
	TeamName *string `json:"teamName,omitempty"`

	// AudTag is a list of audience tags.
	// +optional
	AudTag []string `json:"audTag,omitempty"`

	// DefaultPassesThrough indicates if requests pass through by default.
	// +optional
	DefaultPassesThrough *bool `json:"defaultPassesThrough,omitempty"`
}

// TunnelIngress defines an ingress rule.
type TunnelIngress struct {
	// Hostname is the hostname for the ingress rule.
	// +optional
	Hostname *string `json:"hostname,omitempty"`

	// Path is the path for the ingress rule.
	// +optional
	Path *string `json:"path,omitempty"`

	// Service is the service for the ingress rule.
	// +required
	Service string `json:"service"`

	// OriginRequest is the origin request configuration for this ingress rule.
	// +optional
	OriginRequest *TunnelOriginRequest `json:"originRequest,omitempty"`
}

// TunnelOriginRequest defines origin request configuration.
type TunnelOriginRequest struct {
	// ConnectTimeout is the connection timeout.
	// +optional
	ConnectTimeout *int `json:"connectTimeout,omitempty"`

	// TLSTimeout is the TLS timeout.
	// +optional
	TLSTimeout *int `json:"tlsTimeout,omitempty"`

	// TCPKeepAlive is the TCP keep alive duration.
	// +optional
	TCPKeepAlive *int `json:"tcpKeepAlive,omitempty"`

	// NoHappyEyeballs indicates whether to disable happy eyeballs.
	// +optional
	NoHappyEyeballs *bool `json:"noHappyEyeballs,omitempty"`

	// KeepAliveConnections is the number of keep alive connections.
	// +optional
	KeepAliveConnections *int `json:"keepAliveConnections,omitempty"`

	// KeepAliveTimeout is the keep alive timeout.
	// +optional
	KeepAliveTimeout *int `json:"keepAliveTimeout,omitempty"`

	// HTTPHostHeader is the HTTP host header.
	// +optional
	HTTPHostHeader *string `json:"httpHostHeader,omitempty"`

	// OriginServerName is the origin server name.
	// +optional
	OriginServerName *string `json:"originServerName,omitempty"`

	// CAFile is the CA certificate file.
	// +optional
	CAFile []string `json:"caFile,omitempty"`

	// NoTLSVerify indicates whether to skip TLS verification.
	// +optional
	NoTLSVerify *bool `json:"noTlsVerify,omitempty"`

	// DisableChunkedEncoding indicates whether to disable chunked encoding.
	// +optional
	DisableChunkedEncoding *bool `json:"disableChunkedEncoding,omitempty"`

	// BastionMode indicates whether to enable bastion mode.
	// +optional
	BastionMode *bool `json:"bastionMode,omitempty"`

	// ProxyAddress is the proxy address.
	// +optional
	ProxyAddress *string `json:"proxyAddress,omitempty"`

	// ProxyPort is the proxy port.
	// +optional
	ProxyPort *int `json:"proxyPort,omitempty"`

	// ProxyType is the proxy type.
	// +optional
	ProxyType *string `json:"proxyType,omitempty"`

	// HTTP2Origin indicates whether to use HTTP/2 for origin.
	// +optional
	HTTP2Origin *bool `json:"http2Origin,omitempty"`
}

// TunnelWarpRouting defines WARP routing configuration.
type TunnelWarpRouting struct {
	// Enabled indicates whether WARP routing is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`
}

// TunnelObservation are the observable fields of a Tunnel.
type TunnelObservation struct {
	// ID is the unique identifier of the tunnel.
	ID string `json:"id,omitempty"`

	// Name is the name of the tunnel.
	Name string `json:"name,omitempty"`

	// CreatedAt is the timestamp when the tunnel was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// DeletedAt is the timestamp when the tunnel was deleted.
	DeletedAt *metav1.Time `json:"deletedAt,omitempty"`

	// Connections is a list of tunnel connections.
	Connections []TunnelConnection `json:"connections,omitempty"`
}

// TunnelConnection represents a tunnel connection.
type TunnelConnection struct {
	// ID is the connection ID.
	ID string `json:"id,omitempty"`

	// ClientID is the client ID.
	ClientID string `json:"clientId,omitempty"`

	// ClientVersion is the client version.
	ClientVersion string `json:"clientVersion,omitempty"`

	// OpenedAt is when the connection was opened.
	OpenedAt *metav1.Time `json:"openedAt,omitempty"`

	// OriginIP is the origin IP address.
	OriginIP string `json:"originIp,omitempty"`

	// ColoName is the colocation name.
	ColoName string `json:"coloName,omitempty"`
}

// A TunnelSpec defines the desired state of a Tunnel.
type TunnelSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       TunnelParameters `json:"forProvider"`
}

// A TunnelStatus represents the observed state of a Tunnel.
type TunnelStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          TunnelObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Tunnel represents a Cloudflare Tunnel for secure connectivity.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type Tunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TunnelSpec   `json:"spec"`
	Status TunnelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TunnelList contains a list of Tunnel objects.
type TunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tunnel `json:"items"`
}

}