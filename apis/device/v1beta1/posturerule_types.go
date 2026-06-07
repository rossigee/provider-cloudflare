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

// DevicePostureRuleParameters define the desired state of a Cloudflare Device Posture Rule.
type DevicePostureRuleParameters struct {
	// AccountID is the account ID where this device posture rule will be created.
	// +required
	AccountID string `json:"accountId"`

	// Type is the type of device posture rule.
	// +kubebuilder:validation:Enum=file;application;tanium;gateway;warp;disk_encryption;serial_number;decryption_key;os_version;domain_joined;os;dFirewall;client_certificate;client_certificate_v2;workspace_one;sentinelone;intune;crowdstrike;firewall;disk_encryption_browsers;unique_client_id;kolide
	// +required
	Type string `json:"type"`

	// Name is the name of the device posture rule.
	// +required
	Name string `json:"name"`

	// Description is the description of the device posture rule.
	// +optional
	Description *string `json:"description,omitempty"`

	// Schedule is the schedule for the device posture rule.
	// +optional
	Schedule *string `json:"schedule,omitempty"`

	// Match is the match criteria for the device posture rule.
	// +optional
	Match []DevicePostureRuleMatch `json:"match,omitempty"`

	// Input is the input configuration for the device posture rule.
	// +optional
	Input *DevicePostureRuleInput `json:"input,omitempty"`

	// Expiration is the expiration configuration for the device posture rule.
	// +optional
	Expiration *string `json:"expiration,omitempty"`
}

// DevicePostureRuleMatch defines match criteria for device posture rules.
type DevicePostureRuleMatch struct {
	// Platform is the platform to match.
	// +kubebuilder:validation:Enum=windows;mac;linux;android;ios
	// +optional
	Platform *string `json:"platform,omitempty"`
}

// DevicePostureRuleInput defines input configuration for device posture rules.
type DevicePostureRuleInput struct {
	// ID is the ID for the input.
	// +optional
	ID *string `json:"id,omitempty"`

	// Path is the path for file-based rules.
	// +optional
	Path *string `json:"path,omitempty"`

	// Exists indicates whether the file should exist.
	// +optional
	Exists *bool `json:"exists,omitempty"`

	// Thumbprint is the certificate thumbprint.
	// +optional
	Thumbprint *string `json:"thumbprint,omitempty"`

	// SHA256 is the SHA256 hash.
	// +optional
	SHA256 *string `json:"sha256,omitempty"`

	// Running indicates whether the application should be running.
	// +optional
	Running *bool `json:"running,omitempty"`

	// RequireAll indicates whether all conditions must be met.
	// +optional
	RequireAll *bool `json:"requireAll,omitempty"`

	// Enabled indicates whether the rule is enabled.
	// +optional
	Enabled *bool `json:"enabled,omitempty"`

	// Version is the version requirement.
	// +optional
	Version *string `json:"version,omitempty"`

	// Operator is the comparison operator.
	// +kubebuilder:validation:Enum=>;<;>=;<=;==
	// +optional
	Operator *string `json:"operator,omitempty"`

	// Domain is the domain for domain joined rules.
	// +optional
	Domain *string `json:"domain,omitempty"`

	// OS is the operating system version.
	// +optional
	OS *string `json:"os,omitempty"`

	// Overall indicates the overall posture result.
	// +optional
	Overall *string `json:"overall,omitempty"`

	// SensorConfig is the sensor configuration.
	// +optional
	SensorConfig *string `json:"sensorConfig,omitempty"`

	// State is the state requirement.
	// +optional
	State *string `json:"state,omitempty"`

	// CertificateID is the certificate ID.
	// +optional
	CertificateID *string `json:"certificateId,omitempty"`

	// CN is the common name.
	// +optional
	CN *string `json:"cn,omitempty"`

	// CheckDisks is a list of disks to check.
	// +optional
	CheckDisks []string `json:"checkDisks,omitempty"`

	// CheckPrivateKey indicates whether to check for private key.
	// +optional
	CheckPrivateKey *bool `json:"checkPrivateKey,omitempty"`

	// ComplianceStatus is the compliance status.
	// +optional
	ComplianceStatus *string `json:"complianceStatus,omitempty"`

	// ConnectionID is the connection ID.
	// +optional
	ConnectionID *string `json:"connectionId,omitempty"`
}

// DevicePostureRuleObservation are the observable fields of a Device Posture Rule.
type DevicePostureRuleObservation struct {
	// ID is the unique identifier of the device posture rule.
	ID string `json:"id,omitempty"`

	// Type is the type of device posture rule.
	Type string `json:"type,omitempty"`

	// Name is the name of the device posture rule.
	Name string `json:"name,omitempty"`

	// Description is the description of the device posture rule.
	Description string `json:"description,omitempty"`

	// Schedule is the schedule for the device posture rule.
	Schedule string `json:"schedule,omitempty"`

	// CreatedAt is the timestamp when the rule was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// UpdatedAt is the timestamp when the rule was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`

	// Match is the match criteria for the device posture rule.
	Match []DevicePostureRuleMatch `json:"match,omitempty"`

	// Input is the input configuration for the device posture rule.
	Input *DevicePostureRuleInput `json:"input,omitempty"`

	// Expiration is the expiration configuration for the device posture rule.
	Expiration string `json:"expiration,omitempty"`
}

// A DevicePostureRuleSpec defines the desired state of a Device Posture Rule.
type DevicePostureRuleSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       DevicePostureRuleParameters `json:"forProvider"`
}

// A DevicePostureRuleStatus represents the observed state of a Device Posture Rule.
type DevicePostureRuleStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          DevicePostureRuleObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A DevicePostureRule represents a Cloudflare Device Posture Rule for endpoint security.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".spec.forProvider.name"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type DevicePostureRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DevicePostureRuleSpec   `json:"spec"`
	Status DevicePostureRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DevicePostureRuleList contains a list of Device Posture Rule objects.
type DevicePostureRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevicePostureRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevicePostureRule{}, &DevicePostureRuleList{})
}