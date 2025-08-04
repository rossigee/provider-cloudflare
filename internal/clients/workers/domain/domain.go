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

package domain

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// CloudflareDomainClient is a Cloudflare API client for Workers Custom Domains.
type CloudflareDomainClient struct {
	client *cloudflare.API
}

// NewClient creates a new CloudflareDomainClient.
func NewClient(client *cloudflare.API) *CloudflareDomainClient {
	return &CloudflareDomainClient{client: client}
}

// Create attaches a worker to a custom domain.
func (c *CloudflareDomainClient) Create(ctx context.Context, params v1alpha1.DomainParameters) (*v1alpha1.DomainObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.AccountID,
		Type:       cloudflare.AccountType,
	}

	attachParams := convertParametersToAttachDomain(params)
	
	domain, err := c.client.AttachWorkersDomain(ctx, rc, attachParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot attach workers domain")
	}

	return convertDomainToObservation(domain), nil
}

// Get retrieves a Workers Custom Domain by ID.
func (c *CloudflareDomainClient) Get(ctx context.Context, accountID, domainID string) (*v1alpha1.DomainObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}

	domain, err := c.client.GetWorkersDomain(ctx, rc, domainID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("workers domain not found")
		}
		return nil, errors.Wrap(err, "cannot get workers domain")
	}

	return convertDomainToObservation(domain), nil
}

// Update updates a Workers Custom Domain (re-attachment).
func (c *CloudflareDomainClient) Update(ctx context.Context, domainID string, params v1alpha1.DomainParameters) (*v1alpha1.DomainObservation, error) {
	// For Workers domains, we need to detach and re-attach to update
	rc := &cloudflare.ResourceContainer{
		Identifier: params.AccountID,
		Type:       cloudflare.AccountType,
	}

	// Detach the existing domain
	err := c.client.DetachWorkersDomain(ctx, rc, domainID)
	if err != nil && !isNotFound(err) {
		return nil, errors.Wrap(err, "cannot detach workers domain for update")
	}

	// Re-attach with new parameters
	attachParams := convertParametersToAttachDomain(params)
	
	domain, err := c.client.AttachWorkersDomain(ctx, rc, attachParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot re-attach workers domain")
	}

	return convertDomainToObservation(domain), nil
}

// Delete detaches a Workers Custom Domain.
func (c *CloudflareDomainClient) Delete(ctx context.Context, accountID, domainID string) error {
	rc := &cloudflare.ResourceContainer{
		Identifier: accountID,
		Type:       cloudflare.AccountType,
	}

	err := c.client.DetachWorkersDomain(ctx, rc, domainID)
	if err != nil {
		if isNotFound(err) {
			return nil // Already detached
		}
		return errors.Wrap(err, "cannot detach workers domain")
	}

	return nil
}

// IsUpToDate checks if the Workers Custom Domain is up to date.
func (c *CloudflareDomainClient) IsUpToDate(ctx context.Context, params v1alpha1.DomainParameters, obs v1alpha1.DomainObservation) (bool, error) {
	// Compare configurable parameters
	if obs.ZoneID != nil && params.ZoneID != *obs.ZoneID {
		return false, nil
	}

	if obs.Hostname != nil && params.Hostname != *obs.Hostname {
		return false, nil
	}

	if obs.Service != nil && params.Service != *obs.Service {
		return false, nil
	}

	if obs.Environment != nil && params.Environment != *obs.Environment {
		return false, nil
	}

	return true, nil
}

// convertParametersToAttachDomain converts DomainParameters to cloudflare.AttachWorkersDomainParams.
func convertParametersToAttachDomain(params v1alpha1.DomainParameters) cloudflare.AttachWorkersDomainParams {
	return cloudflare.AttachWorkersDomainParams{
		ZoneID:      params.ZoneID,
		Hostname:    params.Hostname,
		Service:     params.Service,
		Environment: params.Environment,
	}
}

// convertDomainToObservation converts cloudflare.WorkersDomain to DomainObservation.
func convertDomainToObservation(domain cloudflare.WorkersDomain) *v1alpha1.DomainObservation {
	obs := &v1alpha1.DomainObservation{
		ID:          &domain.ID,
		ZoneID:      &domain.ZoneID,
		ZoneName:    &domain.ZoneName,
		Hostname:    &domain.Hostname,
		Service:     &domain.Service,
		Environment: &domain.Environment,
	}

	return obs
}

// isNotFound checks if an error indicates that the workers domain was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "domain not found") ||
		strings.Contains(errStr, "does not exist")
}