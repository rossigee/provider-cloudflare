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

package subdomain

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// CloudflareSubdomainClient is a Cloudflare API client for Workers Subdomain configuration.
type CloudflareSubdomainClient struct {
	client *cloudflare.API
}

// NewClient creates a new CloudflareSubdomainClient.
func NewClient(client *cloudflare.API) *CloudflareSubdomainClient {
	return &CloudflareSubdomainClient{client: client}
}

// Get retrieves the Workers Subdomain configuration for an account.
func (c *CloudflareSubdomainClient) Get(ctx context.Context, accountID string) (*v1alpha1.SubdomainObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}

	subdomain, err := c.client.WorkersGetSubdomain(ctx, rc)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("workers subdomain not found")
		}
		return nil, errors.Wrap(err, "cannot get workers subdomain")
	}

	return convertSubdomainToObservation(subdomain), nil
}

// Update updates the Workers Subdomain configuration for an account.
func (c *CloudflareSubdomainClient) Update(ctx context.Context, params v1alpha1.SubdomainParameters) (*v1alpha1.SubdomainObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.AccountID,
		Type:       cloudflare.AccountType,
	}

	createParams := convertParametersToSubdomain(params)
	
	subdomain, err := c.client.WorkersCreateSubdomain(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update workers subdomain")
	}

	return convertSubdomainToObservation(subdomain), nil
}

// IsUpToDate checks if the Workers Subdomain configuration is up to date.
func (c *CloudflareSubdomainClient) IsUpToDate(ctx context.Context, params v1alpha1.SubdomainParameters, obs v1alpha1.SubdomainObservation) (bool, error) {
	// Compare configurable parameters
	if obs.Name != nil && params.Name != *obs.Name {
		return false, nil
	}

	return true, nil
}

// convertParametersToSubdomain converts SubdomainParameters to cloudflare.WorkersSubdomain.
func convertParametersToSubdomain(params v1alpha1.SubdomainParameters) cloudflare.WorkersSubdomain {
	return cloudflare.WorkersSubdomain{
		Name: params.Name,
	}
}

// convertSubdomainToObservation converts cloudflare.WorkersSubdomain to SubdomainObservation.
func convertSubdomainToObservation(subdomain cloudflare.WorkersSubdomain) *v1alpha1.SubdomainObservation {
	obs := &v1alpha1.SubdomainObservation{
		Name: &subdomain.Name,
	}

	return obs
}

// isNotFound checks if an error indicates that the workers subdomain was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "subdomain not found") ||
		strings.Contains(errStr, "does not exist")
}