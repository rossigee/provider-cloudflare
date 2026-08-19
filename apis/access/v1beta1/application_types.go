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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccessApplicationParameters define the desired state of a Cloudflare Access Application.
type AccessApplicationParameters struct {
	// AccountID is the account ID where this access application will be created.
	// +required
	AccountID string `json:"accountId"`

	// ZoneID is the zone ID where this access application will be created.
	// +optional
	ZoneID *string `json:"zoneId,omitempty"`

	// Name is the name of the access application.
	// +required
	Name string `json:"name"`

	// Domain is the domain for the access application.
	// +required
	Domain string `json:"domain"`

	// Type is the type of the access application.
	// +kubebuilder:validation:Enum=self_hosted;ssh;browser_ssh;warp;bookmark;saas;app_launcher;device_enrollment_permissions;browser_isolation_permissions
	// +required
	Type string `json:"type"`

	// SessionDuration is the lifespan of the access application.
	// +optional
	SessionDuration *string `json:"sessionDuration,omitempty"`

	// AllowedIdps is a list of allowed Identity Provider IDs.
	// +optional
	AllowedIdps []string `json:"allowedIdps,omitempty"`

	// AutoRedirectToIdentity indicates whether to automatically redirect to the identity provider.
	// +optional
	AutoRedirectToIdentity *bool `json:"autoRedirectToIdentity,omitempty"`

	// EnableBindingCookie indicates whether to enable binding cookie.
	// +optional
	EnableBindingCookie *bool `json:"enableBindingCookie,omitempty"`

	// AppLauncherVisible indicates whether the app is visible in the app launcher.
	// +optional
	AppLauncherVisible *bool `json:"appLauncherVisible,omitempty"`

	// ServiceAuth401Redirect indicates whether to enable 401 redirect for service auth.
	// +optional
	ServiceAuth401Redirect *bool `json:"serviceAuth401Redirect,omitempty"`

	// SkipInterstitial indicates whether to skip the interstitial page.
	// +optional
	SkipInterstitial *bool `json:"skipInterstitial,omitempty"`

	// LogoURL is the URL of the logo for the application.
	// +optional
	LogoURL *string `json:"logoUrl,omitempty"`

	// BgColor is the background color for the application.
	// +optional
	BgColor *string `json:"bgColor,omitempty"`

	// HeaderBgColor is the header background color for the application.
	// +optional
	HeaderBgColor *string `json:"headerBgColor,omitempty"`

	// FooterLinks is a list of footer links for the application.
	// +optional
	FooterLinks []AccessApplicationFooterLink `json:"footerLinks,omitempty"`

	// LandingPageDesign defines the landing page design.
	// +optional
	LandingPageDesign *AccessApplicationLandingPageDesign `json:"landingPageDesign,omitempty"`

	// Policies references the policies to apply to this application.
	// +optional
	Policies []string `json:"policies,omitempty"`
}

// AccessApplicationFooterLink defines a footer link for the access application.
type AccessApplicationFooterLink struct {
	// Name is the name of the footer link.
	// +required
	Name string `json:"name"`

	// URL is the URL of the footer link.
	// +required
	URL string `json:"url"`
}

// AccessApplicationLandingPageDesign defines the landing page design.
type AccessApplicationLandingPageDesign struct {
	// Title is the title of the landing page.
	// +optional
	Title *string `json:"title,omitempty"`

	// Message is the message of the landing page.
	// +optional
	Message *string `json:"message,omitempty"`

	// ImageURL is the URL of the image for the landing page.
	// +optional
	ImageURL *string `json:"imageUrl,omitempty"`

	// ButtonColor is the color of the button on the landing page.
	// +optional
	ButtonColor *string `json:"buttonColor,omitempty"`

	// ButtonTextColor is the text color of the button on the landing page.
	// +optional
	ButtonTextColor *string `json:"buttonTextColor,omitempty"`
}

// AccessApplicationObservation are the observable fields of an Access Application.
type AccessApplicationObservation struct {
	// ID is the unique identifier of the access application.
	ID string `json:"id,omitempty"`

	// CreatedAt is the timestamp when the application was created.
	CreatedAt *metav1.Time `json:"createdAt,omitempty"`

	// UpdatedAt is the timestamp when the application was last updated.
	UpdatedAt *metav1.Time `json:"updatedAt,omitempty"`

	// Name is the name of the access application.
	Name string `json:"name,omitempty"`

	// Domain is the domain for the access application.
	Domain string `json:"domain,omitempty"`

	// Type is the type of the access application.
	Type string `json:"type,omitempty"`

	// SessionDuration is the lifespan of the access application.
	SessionDuration string `json:"sessionDuration,omitempty"`

	// AllowedIdps is a list of allowed Identity Provider IDs.
	AllowedIdps []string `json:"allowedIdps,omitempty"`

	// AutoRedirectToIdentity indicates whether to automatically redirect to the identity provider.
	AutoRedirectToIdentity bool `json:"autoRedirectToIdentity,omitempty"`

	// EnableBindingCookie indicates whether to enable binding cookie.
	EnableBindingCookie bool `json:"enableBindingCookie,omitempty"`

	// AppLauncherVisible indicates whether the app is visible in the app launcher.
	AppLauncherVisible bool `json:"appLauncherVisible,omitempty"`

	// ServiceAuth401Redirect indicates whether to enable 401 redirect for service auth.
	ServiceAuth401Redirect bool `json:"serviceAuth401Redirect,omitempty"`

	// SkipInterstitial indicates whether to skip the interstitial page.
	SkipInterstitial bool `json:"skipInterstitial,omitempty"`

	// LogoURL is the URL of the logo for the application.
	LogoURL string `json:"logoUrl,omitempty"`

	// BgColor is the background color for the application.
	BgColor string `json:"bgColor,omitempty"`

	// HeaderBgColor is the header background color for the application.
	HeaderBgColor string `json:"headerBgColor,omitempty"`

	// FooterLinks is a list of footer links for the application.
	FooterLinks []AccessApplicationFooterLink `json:"footerLinks,omitempty"`

	// LandingPageDesign defines the landing page design.
	LandingPageDesign *AccessApplicationLandingPageDesign `json:"landingPageDesign,omitempty"`

	// Aud is the audience tag for the application.
	Aud string `json:"aud,omitempty"`
}

// A AccessApplicationSpec defines the desired state of an Access Application.
type AccessApplicationSpec struct {
	xpv1.ClusterManagedResourceSpec `json:",inline"`
	ForProvider                     AccessApplicationParameters `json:"forProvider"`
}

// A AccessApplicationStatus represents the observed state of an Access Application.
type AccessApplicationStatus struct {
	xpv1.ManagedResourceStatus `json:",inline"`
	AtProvider                 AccessApplicationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A AccessApplication represents a Cloudflare Access Application for Zero Trust security.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="DOMAIN",type="string",JSONPath=".spec.forProvider.domain"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,cloudflare}
type AccessApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccessApplicationSpec   `json:"spec"`
	Status AccessApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccessApplicationList contains a list of Access Application objects.
type AccessApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccessApplication `json:"items"`
}
