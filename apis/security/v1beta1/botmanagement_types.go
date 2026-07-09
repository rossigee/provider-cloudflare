/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)


// BotManagementParameters define the desired state of Cloudflare Bot Management for a zone.
type BotManagementParameters struct {
	// Zone is the zone ID where this bot management configuration will be applied.
	// +required
	Zone string `json:"zone"`

	// EnableJS indicates whether to enable JavaScript detections and challenges.
	// +optional
	EnableJS *bool `json:"enableJS,omitempty"`

	// FightMode indicates whether Bot Fight Mode is enabled.
	// This helps mitigate automated traffic with a free plan.
	// +optional
	FightMode *bool `json:"fightMode,omitempty"`

	// SBFMDefinitelyAutomated configures the action for traffic that is definitely automated.
	// Valid values: "allow", "block", "challenge", "js_challenge", "managed_challenge"
	// +optional
	// +kubebuilder:validation:Enum=allow;block;challenge;js_challenge;managed_challenge
	SBFMDefinitelyAutomated *string `json:"sbfmDefinitelyAutomated,omitempty"`

	// SBFMLikelyAutomated configures the action for traffic that is likely automated.
	// Valid values: "allow", "block", "challenge", "js_challenge", "managed_challenge"
	// +optional
	// +kubebuilder:validation:Enum=allow;block;challenge;js_challenge;managed_challenge
	SBFMLikelyAutomated *string `json:"sbfmLikelyAutomated,omitempty"`

	// SBFMVerifiedBots configures the action for verified bots.
	// Valid values: "allow", "block", "challenge", "js_challenge", "managed_challenge"
	// +optional
	// +kubebuilder:validation:Enum=allow;block;challenge;js_challenge;managed_challenge
	SBFMVerifiedBots *string `json:"sbfmVerifiedBots,omitempty"`

	// SBFMStaticResourceProtection indicates whether to enable static resource protection.
	// This protects static resources like images, CSS, and JS files.
	// +optional
	SBFMStaticResourceProtection *bool `json:"sbfmStaticResourceProtection,omitempty"`

	// OptimizeWordpress indicates whether to enable WordPress-specific optimizations.
	// This provides enhanced protection for WordPress sites.
	// +optional
	OptimizeWordpress *bool `json:"optimizeWordpress,omitempty"`

	// SuppressSessionScore indicates whether to suppress session score calculation.
	// +optional
	SuppressSessionScore *bool `json:"suppressSessionScore,omitempty"`

	// AutoUpdateModel indicates whether to automatically update the bot detection model.
	// +optional
	AutoUpdateModel *bool `json:"autoUpdateModel,omitempty"`

	// AIBotsProtection configures protection level for AI/ML bots.
	// Valid values: "allow", "block", "challenge", "js_challenge", "managed_challenge"
	// +optional
	// +kubebuilder:validation:Enum=allow;block;challenge;js_challenge;managed_challenge
	AIBotsProtection *string `json:"aiBotsProtection,omitempty"`
}

// BotManagementObservation are the observable fields of Bot Management.
type BotManagementObservation struct {
	// EnableJS indicates whether JavaScript detections and challenges are enabled.
	EnableJS *bool `json:"enableJS,omitempty"`

	// FightMode indicates whether Bot Fight Mode is enabled.
	FightMode *bool `json:"fightMode,omitempty"`

	// SBFMDefinitelyAutomated shows the action for traffic that is definitely automated.
	SBFMDefinitelyAutomated *string `json:"sbfmDefinitelyAutomated,omitempty"`

	// SBFMLikelyAutomated shows the action for traffic that is likely automated.
	SBFMLikelyAutomated *string `json:"sbfmLikelyAutomated,omitempty"`

	// SBFMVerifiedBots shows the action for verified bots.
	SBFMVerifiedBots *string `json:"sbfmVerifiedBots,omitempty"`

	// SBFMStaticResourceProtection indicates whether static resource protection is enabled.
	SBFMStaticResourceProtection *bool `json:"sbfmStaticResourceProtection,omitempty"`

	// OptimizeWordpress indicates whether WordPress-specific optimizations are enabled.
	OptimizeWordpress *bool `json:"optimizeWordpress,omitempty"`

	// SuppressSessionScore indicates whether session score calculation is suppressed.
	SuppressSessionScore *bool `json:"suppressSessionScore,omitempty"`

	// AutoUpdateModel indicates whether the bot detection model is automatically updated.
	AutoUpdateModel *bool `json:"autoUpdateModel,omitempty"`

	// UsingLatestModel indicates whether the zone is using the latest bot detection model.
	UsingLatestModel *bool `json:"usingLatestModel,omitempty"`

	// AIBotsProtection shows the protection level for AI/ML bots.
	AIBotsProtection *string `json:"aiBotsProtection,omitempty"`
}

// BotManagementSpec defines the desired state of Bot Management.
type BotManagementSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider       BotManagementParameters `json:"forProvider"`
}

// BotManagementStatus defines the observed state of Bot Management.
type BotManagementStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider          BotManagementObservation `json:"atProvider,omitempty"`
}

// A BotManagement is a managed resource that represents Cloudflare Bot Management configuration for a zone.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ZONE",type="string",JSONPath=".spec.forProvider.zone"
// +kubebuilder:printcolumn:name="FIGHT_MODE",type="boolean",JSONPath=".status.atProvider.fightMode"
// +kubebuilder:printcolumn:name="JS_ENABLED",type="boolean",JSONPath=".status.atProvider.enableJS"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type BotManagement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BotManagementSpec   `json:"spec"`
	Status            BotManagementStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// BotManagementList contains a list of Bot Management objects.
type BotManagementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BotManagement `json:"items"`
}

