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

package certificate

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/originssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// OriginCACertificateAPI defines the interface for Origin CA Certificate operations
type OriginCACertificateAPI interface {
	GetOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error)
	CreateOriginCACertificate(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error)
	RevokeOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error)
}

// CloudflareOriginCertificateClient is a Cloudflare API client for Origin CA certificates.
type CloudflareOriginCertificateClient struct {
	client OriginCACertificateAPI
}

// NewClient creates a new CloudflareOriginCertificateClient.
func NewClient(client OriginCACertificateAPI) *CloudflareOriginCertificateClient {
	return &CloudflareOriginCertificateClient{client: client}
}

// NewClientFromAPI creates a new CloudflareOriginCertificateClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareOriginCertificateClient {
	return NewClient(api)
}

// Get retrieves an Origin CA certificate.
func (c *CloudflareOriginCertificateClient) Get(ctx context.Context, certificateID string) (*v1alpha1.CertificateObservation, error) {
	cert, err := c.client.GetOriginCACertificate(ctx, certificateID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("origin ca certificate not found")
		}
		return nil, errors.Wrap(err, "cannot get origin ca certificate")
	}

	return convertCertificateToObservation(cert), nil
}

// Create creates a new Origin CA certificate.
func (c *CloudflareOriginCertificateClient) Create(ctx context.Context, params v1alpha1.CertificateParameters) (*v1alpha1.CertificateObservation, error) {
	createParams := convertParametersToCreate(params)
	
	cert, err := c.client.CreateOriginCACertificate(ctx, createParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create origin ca certificate")
	}

	return convertCertificateToObservation(cert), nil
}

// Update updates an Origin CA certificate. Note: Origin CA certificates cannot be updated,
// so this will return an error.
func (c *CloudflareOriginCertificateClient) Update(ctx context.Context, certificateID string, params v1alpha1.CertificateParameters) (*v1alpha1.CertificateObservation, error) {
	return nil, errors.New("origin ca certificates cannot be updated")
}

// Delete revokes an Origin CA certificate.
func (c *CloudflareOriginCertificateClient) Delete(ctx context.Context, certificateID string) error {
	_, err := c.client.RevokeOriginCACertificate(ctx, certificateID)
	if err != nil && !isNotFound(err) {
		return errors.Wrap(err, "cannot revoke origin ca certificate")
	}
	return nil
}

// IsUpToDate checks if the Origin CA certificate is up to date.
// Since certificates cannot be updated, this always returns true if the certificate exists.
func (c *CloudflareOriginCertificateClient) IsUpToDate(ctx context.Context, params v1alpha1.CertificateParameters, obs v1alpha1.CertificateObservation) (bool, error) {
	// Origin CA certificates cannot be updated, so if it exists, it's "up to date"
	// We can only check if the hostnames match what was requested
	if len(params.Hostnames) != len(obs.Hostnames) {
		return false, nil
	}
	
	// Check if all requested hostnames are present
	requestedMap := make(map[string]bool)
	for _, hostname := range params.Hostnames {
		requestedMap[hostname] = true
	}
	
	for _, hostname := range obs.Hostnames {
		if !requestedMap[hostname] {
			return false, nil
		}
	}
	
	return true, nil
}

// convertParametersToCreate converts CertificateParameters to CreateOriginCertificateParams.
func convertParametersToCreate(params v1alpha1.CertificateParameters) cloudflare.CreateOriginCertificateParams {
	createParams := cloudflare.CreateOriginCertificateParams{
		Hostnames: params.Hostnames,
	}
	
	if params.RequestType != nil {
		createParams.RequestType = *params.RequestType
	}
	
	if params.RequestValidity != nil {
		createParams.RequestValidity = *params.RequestValidity
	}
	
	if params.CSR != nil {
		createParams.CSR = *params.CSR
	}
	
	return createParams
}

// convertCertificateToObservation converts an OriginCACertificate to CertificateObservation.
func convertCertificateToObservation(cert *cloudflare.OriginCACertificate) *v1alpha1.CertificateObservation {
	obs := &v1alpha1.CertificateObservation{
		ID:              cert.ID,
		Certificate:     cert.Certificate,
		Hostnames:       cert.Hostnames,
		RequestType:     cert.RequestType,
		RequestValidity: cert.RequestValidity,
		CSR:             cert.CSR,
	}
	
	if !cert.ExpiresOn.IsZero() {
		obs.ExpiresOn = &metav1.Time{Time: cert.ExpiresOn}
	}
	
	if !cert.RevokedAt.IsZero() {
		obs.RevokedAt = &metav1.Time{Time: cert.RevokedAt}
	}
	
	return obs
}

// isNotFound checks if an error indicates that the certificate was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "certificate not found")
}