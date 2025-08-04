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

package universalssl

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// UniversalSSLAPI defines the interface for Universal SSL operations
type UniversalSSLAPI interface {
	UniversalSSLSettingDetails(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error)
	EditUniversalSSLSetting(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error)
}

// CloudflareUniversalSSLClient is a Cloudflare API client for Universal SSL settings.
type CloudflareUniversalSSLClient struct {
	client UniversalSSLAPI
}

// NewClient creates a new CloudflareUniversalSSLClient.
func NewClient(client UniversalSSLAPI) *CloudflareUniversalSSLClient {
	return &CloudflareUniversalSSLClient{client: client}
}

// Get retrieves Universal SSL settings for a zone.
func (c *CloudflareUniversalSSLClient) Get(ctx context.Context, zoneID string) (*v1alpha1.UniversalSSLObservation, error) {
	settings, err := c.client.UniversalSSLSettingDetails(ctx, zoneID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("universal ssl settings not found")
		}
		return nil, errors.Wrap(err, "cannot get universal ssl settings")
	}

	return convertUniversalSSLToObservation(settings), nil
}

// Update updates Universal SSL settings for a zone.
func (c *CloudflareUniversalSSLClient) Update(ctx context.Context, params v1alpha1.UniversalSSLParameters) (*v1alpha1.UniversalSSLObservation, error) {
	settings := convertParametersToUniversalSSL(params)
	
	result, err := c.client.EditUniversalSSLSetting(ctx, params.Zone, settings)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update universal ssl settings")
	}

	return convertUniversalSSLToObservation(result), nil
}

// IsUpToDate checks if the Universal SSL settings are up to date.
func (c *CloudflareUniversalSSLClient) IsUpToDate(ctx context.Context, params v1alpha1.UniversalSSLParameters, obs v1alpha1.UniversalSSLObservation) (bool, error) {
	// Compare configurable parameters
	if obs.Enabled != nil && params.Enabled != *obs.Enabled {
		return false, nil
	}

	return true, nil
}

// convertParametersToUniversalSSL converts UniversalSSLParameters to cloudflare.UniversalSSLSetting.
func convertParametersToUniversalSSL(params v1alpha1.UniversalSSLParameters) cloudflare.UniversalSSLSetting {
	return cloudflare.UniversalSSLSetting{
		Enabled: params.Enabled,
	}
}

// convertUniversalSSLToObservation converts cloudflare.UniversalSSLSetting to UniversalSSLObservation.
func convertUniversalSSLToObservation(settings cloudflare.UniversalSSLSetting) *v1alpha1.UniversalSSLObservation {
	obs := &v1alpha1.UniversalSSLObservation{
		Enabled: &settings.Enabled,
	}

	return obs
}

// isNotFound checks if an error indicates that universal ssl settings were not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "ssl not found") ||
		strings.Contains(errStr, "does not exist")
}