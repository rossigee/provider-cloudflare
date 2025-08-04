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

package totaltls

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// TotalTLSAPI defines the interface for Total TLS operations
type TotalTLSAPI interface {
	GetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error)
	SetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error)
}

// CloudflareTotalTLSClient is a Cloudflare API client for Total TLS settings.
type CloudflareTotalTLSClient struct {
	client TotalTLSAPI
}

// NewClient creates a new CloudflareTotalTLSClient.
func NewClient(client TotalTLSAPI) *CloudflareTotalTLSClient {
	return &CloudflareTotalTLSClient{client: client}
}

// Get retrieves Total TLS settings for a zone.
func (c *CloudflareTotalTLSClient) Get(ctx context.Context, zoneID string) (*v1alpha1.TotalTLSObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: zoneID,
	}

	settings, err := c.client.GetTotalTLS(ctx, rc)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("total tls settings not found")
		}
		return nil, errors.Wrap(err, "cannot get total tls settings")
	}

	return convertTotalTLSToObservation(settings), nil
}

// Update updates Total TLS settings for a zone.
func (c *CloudflareTotalTLSClient) Update(ctx context.Context, params v1alpha1.TotalTLSParameters) (*v1alpha1.TotalTLSObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.Zone,
	}

	settings := convertParametersToTotalTLS(params)
	
	result, err := c.client.SetTotalTLS(ctx, rc, settings)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update total tls settings")
	}

	return convertTotalTLSToObservation(result), nil
}

// IsUpToDate checks if the Total TLS settings are up to date.
func (c *CloudflareTotalTLSClient) IsUpToDate(ctx context.Context, params v1alpha1.TotalTLSParameters, obs v1alpha1.TotalTLSObservation) (bool, error) {
	// Compare configurable parameters
	if params.Enabled != nil && obs.Enabled != nil && *params.Enabled != *obs.Enabled {
		return false, nil
	}

	if params.CertificateAuthority != nil && obs.CertificateAuthority != nil && *params.CertificateAuthority != *obs.CertificateAuthority {
		return false, nil
	}

	if params.ValidityDays != nil && obs.ValidityDays != nil && *params.ValidityDays != *obs.ValidityDays {
		return false, nil
	}

	return true, nil
}

// convertParametersToTotalTLS converts TotalTLSParameters to cloudflare.TotalTLS.
func convertParametersToTotalTLS(params v1alpha1.TotalTLSParameters) cloudflare.TotalTLS {
	settings := cloudflare.TotalTLS{}

	if params.Enabled != nil {
		settings.Enabled = params.Enabled
	}

	if params.CertificateAuthority != nil {
		settings.CertificateAuthority = *params.CertificateAuthority
	}

	if params.ValidityDays != nil {
		settings.ValidityDays = *params.ValidityDays
	}

	return settings
}

// convertTotalTLSToObservation converts cloudflare.TotalTLS to TotalTLSObservation.
func convertTotalTLSToObservation(settings cloudflare.TotalTLS) *v1alpha1.TotalTLSObservation {
	obs := &v1alpha1.TotalTLSObservation{
		Enabled:              settings.Enabled,
		CertificateAuthority: &settings.CertificateAuthority,
		ValidityDays:         &settings.ValidityDays,
	}

	return obs
}

// isNotFound checks if an error indicates that total tls settings were not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "tls not found") ||
		strings.Contains(errStr, "does not exist")
}