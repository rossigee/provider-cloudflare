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

package turnstile

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/security/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// TurnstileAPI defines the interface for Turnstile operations
type TurnstileAPI interface {
	CreateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	GetTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error)
	UpdateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	DeleteTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error
}

// CloudflareTurnstileClient is a Cloudflare API client for Turnstile widgets.
type CloudflareTurnstileClient struct {
	client TurnstileAPI
}

// NewClient creates a new CloudflareTurnstileClient.
func NewClient(client TurnstileAPI) *CloudflareTurnstileClient {
	return &CloudflareTurnstileClient{client: client}
}

// NewClientFromAPI creates a new CloudflareTurnstileClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareTurnstileClient {
	return NewClient(api)
}

// Create creates a new Turnstile widget.
func (c *CloudflareTurnstileClient) Create(ctx context.Context, params v1alpha1.TurnstileParameters) (*v1alpha1.TurnstileObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.AccountID,
		Type:       cloudflare.AccountType,
	}

	createParams := convertParametersToCreateTurnstile(params)
	
	widget, err := c.client.CreateTurnstileWidget(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create turnstile widget")
	}

	return convertTurnstileToObservation(widget), nil
}

// Get retrieves a Turnstile widget by site key.
func (c *CloudflareTurnstileClient) Get(ctx context.Context, accountID, siteKey string) (*v1alpha1.TurnstileObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}

	widget, err := c.client.GetTurnstileWidget(ctx, rc, siteKey)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("turnstile widget not found")
		}
		return nil, errors.Wrap(err, "cannot get turnstile widget")
	}

	return convertTurnstileToObservation(widget), nil
}

// Update updates a Turnstile widget.
func (c *CloudflareTurnstileClient) Update(ctx context.Context, siteKey string, params v1alpha1.TurnstileParameters) (*v1alpha1.TurnstileObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.AccountID,
		Type:       cloudflare.AccountType,
	}

	updateParams := convertParametersToUpdateTurnstile(siteKey, params)
	
	widget, err := c.client.UpdateTurnstileWidget(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update turnstile widget")
	}

	return convertTurnstileToObservation(widget), nil
}

// Delete deletes a Turnstile widget.
func (c *CloudflareTurnstileClient) Delete(ctx context.Context, accountID, siteKey string) error {
	rc := &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}

	err := c.client.DeleteTurnstileWidget(ctx, rc, siteKey)
	if err != nil {
		if isNotFound(err) {
			return nil // Already deleted
		}
		return errors.Wrap(err, "cannot delete turnstile widget")
	}

	return nil
}

// IsUpToDate checks if the Turnstile widget is up to date.
func (c *CloudflareTurnstileClient) IsUpToDate(ctx context.Context, params v1alpha1.TurnstileParameters, obs v1alpha1.TurnstileObservation) (bool, error) {
	// Compare configurable parameters
	if obs.Name != nil && params.Name != *obs.Name {
		return false, nil
	}

	// Compare domains (order doesn't matter)
	if !equalStringSlices(params.Domains, obs.Domains) {
		return false, nil
	}

	if params.Mode != nil && obs.Mode != nil && *params.Mode != *obs.Mode {
		return false, nil
	}

	if params.BotFightMode != nil && obs.BotFightMode != nil && *params.BotFightMode != *obs.BotFightMode {
		return false, nil
	}

	if params.Region != nil && obs.Region != nil && *params.Region != *obs.Region {
		return false, nil
	}

	if params.OffLabel != nil && obs.OffLabel != nil && *params.OffLabel != *obs.OffLabel {
		return false, nil
	}

	return true, nil
}

// convertParametersToCreateTurnstile converts TurnstileParameters to cloudflare.CreateTurnstileWidgetParams.
func convertParametersToCreateTurnstile(params v1alpha1.TurnstileParameters) cloudflare.CreateTurnstileWidgetParams {
	createParams := cloudflare.CreateTurnstileWidgetParams{
		Name:    params.Name,
		Domains: params.Domains,
	}

	if params.Mode != nil {
		createParams.Mode = *params.Mode
	}

	if params.BotFightMode != nil {
		createParams.BotFightMode = *params.BotFightMode
	}

	if params.Region != nil {
		createParams.Region = *params.Region
	}

	if params.OffLabel != nil {
		createParams.OffLabel = *params.OffLabel
	}

	return createParams
}

// convertParametersToUpdateTurnstile converts TurnstileParameters to cloudflare.UpdateTurnstileWidgetParams.
func convertParametersToUpdateTurnstile(siteKey string, params v1alpha1.TurnstileParameters) cloudflare.UpdateTurnstileWidgetParams {
	updateParams := cloudflare.UpdateTurnstileWidgetParams{
		SiteKey: siteKey,
	}

	updateParams.Name = &params.Name
	updateParams.Domains = &params.Domains

	if params.Mode != nil {
		updateParams.Mode = params.Mode
	}

	if params.BotFightMode != nil {
		updateParams.BotFightMode = params.BotFightMode
	}

	if params.OffLabel != nil {
		updateParams.OffLabel = params.OffLabel
	}

	return updateParams
}

// convertTurnstileToObservation converts cloudflare.TurnstileWidget to TurnstileObservation.
func convertTurnstileToObservation(widget cloudflare.TurnstileWidget) *v1alpha1.TurnstileObservation {
	obs := &v1alpha1.TurnstileObservation{
		SiteKey:      &widget.SiteKey,
		Secret:       &widget.Secret,
		Name:         &widget.Name,
		Domains:      widget.Domains,
		Mode:         &widget.Mode,
		BotFightMode: &widget.BotFightMode,
		Region:       &widget.Region,
		OffLabel:     &widget.OffLabel,
	}

	if widget.CreatedOn != nil {
		obs.CreatedOn = &metav1.Time{Time: *widget.CreatedOn}
	}

	if widget.ModifiedOn != nil {
		obs.ModifiedOn = &metav1.Time{Time: *widget.ModifiedOn}
	}

	return obs
}

// isNotFound checks if an error indicates that the turnstile widget was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "widget not found") ||
		strings.Contains(errStr, "does not exist")
}

// equalStringSlices compares two string slices for equality (order doesn't matter).
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}

	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}

	return true
}