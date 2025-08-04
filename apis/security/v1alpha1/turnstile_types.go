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
	"k8s.io/apimachinery/pkg/runtime/schema"

	rtv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// TurnstileParameters define the desired state of a Cloudflare Turnstile widget.
type TurnstileParameters struct {
	// AccountID is the account identifier to target for the resource.
	// +required
	AccountID string `json:"accountId"`

	// Name is the human readable widget name.
	// +required
	Name string `json:"name"`

	// Domains are the domains for which the widget is active.
	// +required
	Domains []string `json:"domains"`

	// Mode describes how Cloudflare will handle the traffic coming from human or bot.
	// Valid values: "non-interactive", "invisible", "managed"
	// +optional
	// +kubebuilder:validation:Enum=non-interactive;invisible;managed
	Mode *string `json:"mode,omitempty"`

	// BotFightMode indicates whether Bot Fight Mode is enabled for this widget.
	// If true, the widget will enable Cloudflare's Bot Fight Mode.
	// +optional
	BotFightMode *bool `json:"botFightMode,omitempty"`

	// Region is the region for this widget. Valid values: "world" or specific region codes.
	// +optional
	// +kubebuilder:validation:Enum=world
	Region *string `json:"region,omitempty"`

	// OffLabel indicates whether to show/hide Cloudflare branding on the widget.
	// If true, Cloudflare branding is hidden (requires appropriate subscription).
	// +optional
	OffLabel *bool `json:"offLabel,omitempty"`
}

// TurnstileObservation are the observable fields of a Turnstile widget.
type TurnstileObservation struct {
	// SiteKey is the site key of the widget.
	SiteKey *string `json:"siteKey,omitempty"`

	// Secret is the secret key of the widget (sensitive).
	Secret *string `json:"secret,omitempty"`

	// CreatedOn is when the widget was created.
	CreatedOn *metav1.Time `json:"createdOn,omitempty"`

	// ModifiedOn is when the widget was last modified.
	ModifiedOn *metav1.Time `json:"modifiedOn,omitempty"`

	// Name is the human readable widget name.
	Name *string `json:"name,omitempty"`

	// Domains are the domains for which the widget is active.
	Domains []string `json:"domains,omitempty"`

	// Mode describes how Cloudflare handles the traffic.
	Mode *string `json:"mode,omitempty"`

	// BotFightMode indicates whether Bot Fight Mode is enabled.
	BotFightMode *bool `json:"botFightMode,omitempty"`

	// Region is the region for this widget.
	Region *string `json:"region,omitempty"`

	// OffLabel indicates whether Cloudflare branding is hidden.
	OffLabel *bool `json:"offLabel,omitempty"`
}

// TurnstileSpec defines the desired state of Turnstile.
type TurnstileSpec struct {
	rtv1.ResourceSpec `json:",inline"`
	ForProvider       TurnstileParameters `json:"forProvider"`
}

// TurnstileStatus defines the observed state of Turnstile.
type TurnstileStatus struct {
	rtv1.ResourceStatus `json:",inline"`
	AtProvider          TurnstileObservation `json:"atProvider,omitempty"`
}

// A Turnstile is a managed resource that represents a Cloudflare Turnstile widget.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="SITEKEY",type="string",JSONPath=".status.atProvider.siteKey"
// +kubebuilder:printcolumn:name="NAME",type="string",JSONPath=".status.atProvider.name"
// +kubebuilder:printcolumn:name="MODE",type="string",JSONPath=".status.atProvider.mode"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,cloudflare}
// +kubebuilder:object:root=true
type Turnstile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              TurnstileSpec   `json:"spec"`
	Status            TurnstileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// TurnstileList contains a list of Turnstile objects.
type TurnstileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Turnstile `json:"items"`
}

// GetCondition of this Turnstile.
func (mg *Turnstile) GetCondition(ct rtv1.ConditionType) rtv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Turnstile.
func (mg *Turnstile) GetDeletionPolicy() rtv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// GetManagementPolicies of this Turnstile.
func (mg *Turnstile) GetManagementPolicies() rtv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// GetProviderConfigReference of this Turnstile.
func (mg *Turnstile) GetProviderConfigReference() *rtv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// GetPublishConnectionDetailsTo of this Turnstile.
func (mg *Turnstile) GetPublishConnectionDetailsTo() *rtv1.PublishConnectionDetailsTo {
	return mg.Spec.PublishConnectionDetailsTo
}

// GetWriteConnectionSecretToReference of this Turnstile.
func (mg *Turnstile) GetWriteConnectionSecretToReference() *rtv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetConditions of this Turnstile.
func (mg *Turnstile) SetConditions(c ...rtv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Turnstile.
func (mg *Turnstile) SetDeletionPolicy(r rtv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}

// SetManagementPolicies of this Turnstile.
func (mg *Turnstile) SetManagementPolicies(r rtv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// SetProviderConfigReference of this Turnstile.
func (mg *Turnstile) SetProviderConfigReference(r *rtv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// SetPublishConnectionDetailsTo of this Turnstile.
func (mg *Turnstile) SetPublishConnectionDetailsTo(r *rtv1.PublishConnectionDetailsTo) {
	mg.Spec.PublishConnectionDetailsTo = r
}

// SetWriteConnectionSecretToReference of this Turnstile.
func (mg *Turnstile) SetWriteConnectionSecretToReference(r *rtv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetGroupVersionKind returns the GroupVersionKind for Turnstile.
func (mg *Turnstile) GetGroupVersionKind() schema.GroupVersionKind {
	return TurnstileGroupVersionKind
}