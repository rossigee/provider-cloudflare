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

package application

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/access/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

// AccessApplicationAPI defines the interface for Access Application operations
type AccessApplicationAPI interface {
	ListAccessApplications(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListAccessApplicationsParams) ([]cloudflare.AccessApplication, *cloudflare.ResultInfo, error)
	GetAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error)
	CreateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error)
	UpdateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error)
	DeleteAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error
}

// CloudflareAccessApplicationClient is a Cloudflare API client for Access Applications.
type CloudflareAccessApplicationClient struct {
	client AccessApplicationAPI
}

// NewClient creates a new CloudflareAccessApplicationClient.
func NewClient(client AccessApplicationAPI) *CloudflareAccessApplicationClient {
	return &CloudflareAccessApplicationClient{client: client}
}

// NewClientFromAPI creates a new CloudflareAccessApplicationClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareAccessApplicationClient {
	return NewClient(api)
}

// Get retrieves an Access Application.
func (c *CloudflareAccessApplicationClient) Get(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (*v1beta1.AccessApplicationObservation, error) {
	app, err := c.client.GetAccessApplication(ctx, rc, applicationID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("access application not found")
		}
		return nil, errors.Wrap(err, "cannot get access application")
	}

	return convertAccessApplicationToObservation(app), nil
}

// Create creates a new Access Application.
func (c *CloudflareAccessApplicationClient) Create(ctx context.Context, params v1beta1.AccessApplicationParameters) (*v1beta1.AccessApplicationObservation, error) {
	createParams := convertParametersToCreateAccessApplication(params)

	app, err := c.client.CreateAccessApplication(ctx, getResourceContainer(params), createParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create access application")
	}

	return convertAccessApplicationToObservation(app), nil
}

// Update updates an Access Application.
func (c *CloudflareAccessApplicationClient) Update(ctx context.Context, applicationID string, params v1beta1.AccessApplicationParameters) (*v1beta1.AccessApplicationObservation, error) {
	updateParams := convertParametersToUpdateAccessApplication(applicationID, params)

	app, err := c.client.UpdateAccessApplication(ctx, getResourceContainer(params), updateParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update access application")
	}

	return convertAccessApplicationToObservation(app), nil
}

// Delete deletes an Access Application.
func (c *CloudflareAccessApplicationClient) Delete(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error {
	err := c.client.DeleteAccessApplication(ctx, rc, applicationID)
	if err != nil && !isNotFound(err) {
		return errors.Wrap(err, "cannot delete access application")
	}
	return nil
}

// IsUpToDate checks if the Access Application is up to date.
func (c *CloudflareAccessApplicationClient) IsUpToDate(ctx context.Context, params v1beta1.AccessApplicationParameters, obs v1beta1.AccessApplicationObservation) (bool, error) {
	// Compare key parameters
	if params.Name != obs.Name {
		return false, nil
	}

	if params.Domain != obs.Domain {
		return false, nil
	}

	if params.Type != obs.Type {
		return false, nil
	}

	if params.SessionDuration != nil && *params.SessionDuration != obs.SessionDuration {
		return false, nil
	}

	if params.AutoRedirectToIdentity != nil && *params.AutoRedirectToIdentity != obs.AutoRedirectToIdentity {
		return false, nil
	}

	if params.EnableBindingCookie != nil && *params.EnableBindingCookie != obs.EnableBindingCookie {
		return false, nil
	}

	if params.AppLauncherVisible != nil && *params.AppLauncherVisible != obs.AppLauncherVisible {
		return false, nil
	}

	if params.ServiceAuth401Redirect != nil && *params.ServiceAuth401Redirect != obs.ServiceAuth401Redirect {
		return false, nil
	}

	if params.SkipInterstitial != nil && *params.SkipInterstitial != obs.SkipInterstitial {
		return false, nil
	}

	return true, nil
}

// getResourceContainer creates a ResourceContainer based on the parameters.
func getResourceContainer(params v1beta1.AccessApplicationParameters) *cloudflare.ResourceContainer {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: params.AccountID,
	}

	if params.ZoneID != nil {
		rc = &cloudflare.ResourceContainer{
			Level:      cloudflare.ZoneRouteLevel,
			Identifier: *params.ZoneID,
		}
	}

	return rc
}

// convertParametersToCreateAccessApplication converts AccessApplicationParameters to cloudflare.CreateAccessApplicationParams.
func convertParametersToCreateAccessApplication(params v1beta1.AccessApplicationParameters) cloudflare.CreateAccessApplicationParams {
	app := cloudflare.CreateAccessApplicationParams{
		Name:   params.Name,
		Domain: params.Domain,
		Type:   cloudflare.AccessApplicationType(params.Type),
	}

	if params.SessionDuration != nil {
		app.SessionDuration = *params.SessionDuration
	}

	if params.AllowedIdps != nil {
		app.AllowedIdps = params.AllowedIdps
	}

	if params.AutoRedirectToIdentity != nil {
		app.AutoRedirectToIdentity = params.AutoRedirectToIdentity
	}

	if params.EnableBindingCookie != nil {
		app.EnableBindingCookie = params.EnableBindingCookie
	}

	if params.AppLauncherVisible != nil {
		app.AppLauncherVisible = params.AppLauncherVisible
	}

	if params.ServiceAuth401Redirect != nil {
		app.ServiceAuth401Redirect = params.ServiceAuth401Redirect
	}

	if params.SkipInterstitial != nil {
		app.SkipInterstitial = params.SkipInterstitial
	}

	if params.LogoURL != nil {
		app.LogoURL = *params.LogoURL
	}

	if params.BgColor != nil {
		app.BackgroundColor = *params.BgColor
	}

	if params.HeaderBgColor != nil {
		app.HeaderBackgroundColor = *params.HeaderBgColor
	}

	if params.FooterLinks != nil {
		app.FooterLinks = convertFooterLinks(params.FooterLinks)
	}

	if params.LandingPageDesign != nil {
		app.LandingPageDesign = convertLandingPageDesign(params.LandingPageDesign)
	}

	return app
}

// convertParametersToUpdateAccessApplication converts AccessApplicationParameters to cloudflare.UpdateAccessApplicationParams.
func convertParametersToUpdateAccessApplication(applicationID string, params v1beta1.AccessApplicationParameters) cloudflare.UpdateAccessApplicationParams {
	app := cloudflare.UpdateAccessApplicationParams{
		ID: applicationID,
	}

	// Set required fields
	app.Name = params.Name
	app.Domain = params.Domain
	app.Type = cloudflare.AccessApplicationType(params.Type)

	if params.SessionDuration != nil {
		app.SessionDuration = *params.SessionDuration
	}

	if params.AllowedIdps != nil {
		app.AllowedIdps = params.AllowedIdps
	}

	if params.AutoRedirectToIdentity != nil {
		app.AutoRedirectToIdentity = params.AutoRedirectToIdentity
	}

	if params.EnableBindingCookie != nil {
		app.EnableBindingCookie = params.EnableBindingCookie
	}

	if params.AppLauncherVisible != nil {
		app.AppLauncherVisible = params.AppLauncherVisible
	}

	if params.ServiceAuth401Redirect != nil {
		app.ServiceAuth401Redirect = params.ServiceAuth401Redirect
	}

	if params.SkipInterstitial != nil {
		app.SkipInterstitial = params.SkipInterstitial
	}

	if params.LogoURL != nil {
		app.LogoURL = *params.LogoURL
	}

	if params.BgColor != nil {
		app.BackgroundColor = *params.BgColor
	}

	if params.HeaderBgColor != nil {
		app.HeaderBackgroundColor = *params.HeaderBgColor
	}

	if params.FooterLinks != nil {
		app.FooterLinks = convertFooterLinks(params.FooterLinks)
	}

	if params.LandingPageDesign != nil {
		app.LandingPageDesign = convertLandingPageDesign(params.LandingPageDesign)
	}

	return app
}

// convertAccessApplicationToObservation converts cloudflare.AccessApplication to AccessApplicationObservation.
func convertAccessApplicationToObservation(app cloudflare.AccessApplication) *v1beta1.AccessApplicationObservation {
	obs := &v1beta1.AccessApplicationObservation{
		ID:              app.ID,
		Name:            app.Name,
		Domain:          app.Domain,
		Type:            string(app.Type),
		SessionDuration: app.SessionDuration,
		AllowedIdps:     app.AllowedIdps,
		LogoURL:         app.LogoURL,
		BgColor:         app.BackgroundColor,
		HeaderBgColor:   app.HeaderBackgroundColor,
		Aud:             app.AUD,
	}

	// Handle pointer boolean fields
	if app.AutoRedirectToIdentity != nil {
		obs.AutoRedirectToIdentity = *app.AutoRedirectToIdentity
	}
	if app.EnableBindingCookie != nil {
		obs.EnableBindingCookie = *app.EnableBindingCookie
	}
	if app.AppLauncherVisible != nil {
		obs.AppLauncherVisible = *app.AppLauncherVisible
	}
	if app.ServiceAuth401Redirect != nil {
		obs.ServiceAuth401Redirect = *app.ServiceAuth401Redirect
	}
	if app.SkipInterstitial != nil {
		obs.SkipInterstitial = *app.SkipInterstitial
	}

	if app.CreatedAt != nil {
		obs.CreatedAt = &metav1.Time{Time: *app.CreatedAt}
	}

	if app.UpdatedAt != nil {
		obs.UpdatedAt = &metav1.Time{Time: *app.UpdatedAt}
	}

	if app.FooterLinks != nil {
		obs.FooterLinks = convertFooterLinksFromCloudflare(app.FooterLinks)
	}

	if app.LandingPageDesign.Title != "" || app.LandingPageDesign.Message != "" || app.LandingPageDesign.ImageURL != "" {
		obs.LandingPageDesign = convertLandingPageDesignFromCloudflare(app.LandingPageDesign)
	}

	return obs
}

// convertFooterLinks converts []v1beta1.AccessApplicationFooterLink to []cloudflare.AccessFooterLink.
func convertFooterLinks(links []v1beta1.AccessApplicationFooterLink) []cloudflare.AccessFooterLink {
	cfLinks := make([]cloudflare.AccessFooterLink, len(links))
	for i, link := range links {
		cfLinks[i] = cloudflare.AccessFooterLink{
			Name: link.Name,
			URL:  link.URL,
		}
	}
	return cfLinks
}

// convertFooterLinksFromCloudflare converts []cloudflare.AccessFooterLink to []v1beta1.AccessApplicationFooterLink.
func convertFooterLinksFromCloudflare(links []cloudflare.AccessFooterLink) []v1beta1.AccessApplicationFooterLink {
	v1Links := make([]v1beta1.AccessApplicationFooterLink, len(links))
	for i, link := range links {
		v1Links[i] = v1beta1.AccessApplicationFooterLink{
			Name: link.Name,
			URL:  link.URL,
		}
	}
	return v1Links
}

// convertLandingPageDesign converts *v1beta1.AccessApplicationLandingPageDesign to cloudflare.AccessLandingPageDesign.
func convertLandingPageDesign(design *v1beta1.AccessApplicationLandingPageDesign) cloudflare.AccessLandingPageDesign {
	if design == nil {
		return cloudflare.AccessLandingPageDesign{}
	}

	cfDesign := cloudflare.AccessLandingPageDesign{}

	if design.Title != nil {
		cfDesign.Title = *design.Title
	}

	if design.Message != nil {
		cfDesign.Message = *design.Message
	}

	if design.ImageURL != nil {
		cfDesign.ImageURL = *design.ImageURL
	}

	if design.ButtonColor != nil {
		cfDesign.ButtonColor = *design.ButtonColor
	}

	if design.ButtonTextColor != nil {
		cfDesign.ButtonTextColor = *design.ButtonTextColor
	}

	return cfDesign
}

// convertLandingPageDesignFromCloudflare converts cloudflare.AccessLandingPageDesign to *v1beta1.AccessApplicationLandingPageDesign.
func convertLandingPageDesignFromCloudflare(design cloudflare.AccessLandingPageDesign) *v1beta1.AccessApplicationLandingPageDesign {
	v1Design := &v1beta1.AccessApplicationLandingPageDesign{
		Title:           &design.Title,
		Message:         &design.Message,
		ImageURL:        &design.ImageURL,
		ButtonColor:     &design.ButtonColor,
		ButtonTextColor: &design.ButtonTextColor,
	}

	return v1Design
}

// isNotFound checks if an error indicates that the access application was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "application not found") ||
		strings.Contains(errStr, "does not exist")
}
