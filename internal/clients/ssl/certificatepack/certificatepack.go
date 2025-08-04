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

package certificatepack

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// CertificatePackAPI defines the interface for Certificate Pack operations
type CertificatePackAPI interface {
	CertificatePack(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error)
	CreateCertificatePack(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error)
	DeleteCertificatePack(ctx context.Context, zoneID, certificateID string) error
	RestartCertificateValidation(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error)
}

// CloudflareCertificatePackClient is a Cloudflare API client for Certificate Pack management.
type CloudflareCertificatePackClient struct {
	client CertificatePackAPI
}

// NewClient creates a new CloudflareCertificatePackClient.
func NewClient(client CertificatePackAPI) *CloudflareCertificatePackClient {
	return &CloudflareCertificatePackClient{client: client}
}

// Get retrieves a Certificate Pack by ID.
func (c *CloudflareCertificatePackClient) Get(ctx context.Context, zoneID, certificatePackID string) (*v1alpha1.CertificatePackObservation, error) {
	pack, err := c.client.CertificatePack(ctx, zoneID, certificatePackID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("certificate pack not found")
		}
		return nil, errors.Wrap(err, "cannot get certificate pack")
	}

	return convertCertificatePackToObservation(pack), nil
}

// Create creates a new Certificate Pack.
func (c *CloudflareCertificatePackClient) Create(ctx context.Context, params v1alpha1.CertificatePackParameters) (*v1alpha1.CertificatePackObservation, error) {
	request := convertParametersToCertificatePackRequest(params)
	
	pack, err := c.client.CreateCertificatePack(ctx, params.Zone, request)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create certificate pack")
	}

	return convertCertificatePackToObservation(pack), nil
}

// Delete deletes a Certificate Pack.
func (c *CloudflareCertificatePackClient) Delete(ctx context.Context, zoneID, certificatePackID string) error {
	err := c.client.DeleteCertificatePack(ctx, zoneID, certificatePackID)
	if err != nil {
		if isNotFound(err) {
			return nil // Already deleted
		}
		return errors.Wrap(err, "cannot delete certificate pack")
	}

	return nil
}

// RestartValidation restarts certificate validation for a Certificate Pack.
func (c *CloudflareCertificatePackClient) RestartValidation(ctx context.Context, zoneID, certificatePackID string) (*v1alpha1.CertificatePackObservation, error) {
	pack, err := c.client.RestartCertificateValidation(ctx, zoneID, certificatePackID)
	if err != nil {
		return nil, errors.Wrap(err, "cannot restart certificate validation")
	}

	return convertCertificatePackToObservation(pack), nil
}

// convertParametersToCertificatePackRequest converts CertificatePackParameters to cloudflare.CertificatePackRequest.
func convertParametersToCertificatePackRequest(params v1alpha1.CertificatePackParameters) cloudflare.CertificatePackRequest {
	request := cloudflare.CertificatePackRequest{
		Type:             params.Type,
		Hosts:            params.Hosts,
		ValidationMethod: params.ValidationMethod,
	}

	if params.ValidityDays != nil {
		request.ValidityDays = *params.ValidityDays
	}

	if params.CertificateAuthority != nil {
		request.CertificateAuthority = *params.CertificateAuthority
	}

	if params.CloudflareBranding != nil {
		request.CloudflareBranding = *params.CloudflareBranding
	}

	return request
}

// convertCertificatePackToObservation converts cloudflare.CertificatePack to CertificatePackObservation.
func convertCertificatePackToObservation(pack cloudflare.CertificatePack) *v1alpha1.CertificatePackObservation {
	obs := &v1alpha1.CertificatePackObservation{
		ID:               &pack.ID,
		Type:             &pack.Type,
		Hosts:            pack.Hosts,
		ValidationMethod: &pack.ValidationMethod,
		ValidityDays:     &pack.ValidityDays,
		Status:           &pack.Status,
	}

	if pack.CertificateAuthority != "" {
		obs.CertificateAuthority = &pack.CertificateAuthority
	}

	if pack.PrimaryCertificate != "" {
		obs.PrimaryCertificate = &pack.PrimaryCertificate
	}

	obs.CloudflareBranding = &pack.CloudflareBranding

	// Convert certificates
	if len(pack.Certificates) > 0 {
		obs.Certificates = make([]v1alpha1.CertificateInfo, len(pack.Certificates))
		for i, cert := range pack.Certificates {
			obs.Certificates[i] = v1alpha1.CertificateInfo{
				ID:     &cert.ID,
				Hosts:  cert.Hosts,
				Issuer: &cert.Issuer,
				Status: &cert.Status,
			}

			if !cert.ExpiresOn.IsZero() {
				obs.Certificates[i].ExpiresOn = &metav1.Time{Time: cert.ExpiresOn}
			}

			if !cert.UploadedOn.IsZero() {
				obs.Certificates[i].UploadedOn = &metav1.Time{Time: cert.UploadedOn}
			}

			if !cert.ModifiedOn.IsZero() {
				obs.Certificates[i].ModifiedOn = &metav1.Time{Time: cert.ModifiedOn}
			}
		}
	}

	// Convert validation records
	if len(pack.ValidationRecords) > 0 {
		obs.ValidationRecords = make([]v1alpha1.SSLValidationRecord, len(pack.ValidationRecords))
		for i, record := range pack.ValidationRecords {
			obs.ValidationRecords[i] = v1alpha1.SSLValidationRecord{}

			if record.TxtName != "" {
				obs.ValidationRecords[i].TxtName = &record.TxtName
			}

			if record.TxtValue != "" {
				obs.ValidationRecords[i].TxtValue = &record.TxtValue
			}

			if record.HTTPUrl != "" {
				obs.ValidationRecords[i].HTTPPath = &record.HTTPUrl
			}

			if record.HTTPBody != "" {
				obs.ValidationRecords[i].HTTPBody = &record.HTTPBody
			}

			if len(record.Emails) > 0 {
				obs.ValidationRecords[i].EmailAddresses = record.Emails
			}
		}
	}

	// Convert validation errors
	if len(pack.ValidationErrors) > 0 {
		obs.ValidationErrors = make([]v1alpha1.SSLValidationError, len(pack.ValidationErrors))
		for i, validationError := range pack.ValidationErrors {
			obs.ValidationErrors[i] = v1alpha1.SSLValidationError{
				Message: &validationError.Message,
			}
		}
	}

	return obs
}

// isNotFound checks if an error indicates that certificate pack was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "certificate not found") ||
		strings.Contains(errStr, "does not exist")
}