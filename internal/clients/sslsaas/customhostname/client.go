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

package customhostname

import (
	"context"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCustomHostnameNotFound = "Custom Hostname not found"
)

// Client is a Cloudflare SSL for SaaS API client
type Client interface {
	CustomHostname(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error)
	CreateCustomHostname(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	UpdateCustomHostname(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	DeleteCustomHostname(ctx context.Context, zoneID, hostnameID string) error
}

type client struct {
	cf *cloudflare.API
}

// NewClient returns a new CustomHostname client
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &client{cf: cf}, nil
}

// CustomHostname retrieves a Custom Hostname
func (c *client) CustomHostname(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
	hostname, err := c.cf.CustomHostname(ctx, zoneID, hostnameID)
	if err != nil {
		return cloudflare.CustomHostname{}, err
	}

	if hostname.ID == "" {
		return cloudflare.CustomHostname{}, errors.New(errCustomHostnameNotFound)
	}

	return hostname, nil
}

// CreateCustomHostname creates a new Custom Hostname
func (c *client) CreateCustomHostname(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
	response, err := c.cf.CreateCustomHostname(ctx, zoneID, hostname)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateCustomHostname updates an existing Custom Hostname
func (c *client) UpdateCustomHostname(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
	hostname.ID = hostnameID
	response, err := c.cf.UpdateCustomHostname(ctx, zoneID, hostnameID, hostname)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteCustomHostname deletes a Custom Hostname
func (c *client) DeleteCustomHostname(ctx context.Context, zoneID, hostnameID string) error {
	err := c.cf.DeleteCustomHostname(ctx, zoneID, hostnameID)
	return err
}

// IsCustomHostnameNotFound returns true if the error indicates the hostname was not found
func IsCustomHostnameNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errCustomHostnameNotFound ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}

// GenerateObservation creates observation data from a Custom Hostname
func GenerateObservation(hostname cloudflare.CustomHostname) v1alpha1.CustomHostnameObservation {
	obs := v1alpha1.CustomHostnameObservation{
		Status: hostname.Status,
	}

	// Map SSL status if available
	if hostname.SSL != nil {
		obs.SSL = v1alpha1.CustomHostnameSSLObserved{
			Status:               hostname.SSL.Status,
			HTTPUrl:              hostname.SSL.HTTPUrl,
			HTTPBody:             hostname.SSL.HTTPBody,
			CertificateAuthority: hostname.SSL.CertificateAuthority,
			CnameName:            hostname.SSL.CnameName,
			CnameTarget:          hostname.SSL.CnameTarget,
		}

		// Map validation errors if present
		for _, verr := range hostname.SSL.ValidationErrors {
			obs.SSL.ValidationErrors = append(obs.SSL.ValidationErrors, v1alpha1.CustomHostnameSSLValidationErrors{
				Message: verr.Message,
			})
		}
	}

	// Map ownership verification if available
	if hostname.OwnershipVerification.Name != "" {
		obs.OwnershipVerification.DNSRecord = &v1alpha1.CustomHostnameOwnershipVerificationDNS{
			Name:  &hostname.OwnershipVerification.Name,
			Type:  &hostname.OwnershipVerification.Type,
			Value: &hostname.OwnershipVerification.Value,
		}
	}
	if hostname.OwnershipVerificationHTTP.HTTPUrl != "" {
		obs.OwnershipVerification.HTTPFile = &v1alpha1.CustomHostnameOwnershipVerificationHTTP{
			URL:  &hostname.OwnershipVerificationHTTP.HTTPUrl,
			Body: &hostname.OwnershipVerificationHTTP.HTTPBody,
		}
	}

	return obs
}

// ParametersToCustomHostname converts CustomHostnameParameters to cloudflare.CustomHostname
func ParametersToCustomHostname(params v1alpha1.CustomHostnameParameters) cloudflare.CustomHostname {
	hostname := cloudflare.CustomHostname{
		Hostname: params.Hostname,
	}

	if params.CustomOriginServer != nil {
		hostname.CustomOriginServer = *params.CustomOriginServer
	}

	// Map SSL settings
	ssl := &cloudflare.CustomHostnameSSL{}
	if params.SSL.Method != nil {
		ssl.Method = *params.SSL.Method
	}
	if params.SSL.Type != nil {
		ssl.Type = *params.SSL.Type
	}
	if params.SSL.Wildcard != nil {
		ssl.Wildcard = params.SSL.Wildcard
	}
	if params.SSL.CustomCertificate != nil {
		ssl.CustomCertificate = *params.SSL.CustomCertificate
	}
	if params.SSL.CustomKey != nil {
		ssl.CustomKey = *params.SSL.CustomKey
	}

	// Map SSL settings
	ssl.Settings = cloudflare.CustomHostnameSSLSettings{}
	if params.SSL.Settings.HTTP2 != nil {
		ssl.Settings.HTTP2 = *params.SSL.Settings.HTTP2
	}
	if params.SSL.Settings.TLS13 != nil {
		ssl.Settings.TLS13 = *params.SSL.Settings.TLS13
	}
	if params.SSL.Settings.MinTLSVersion != nil {
		ssl.Settings.MinTLSVersion = *params.SSL.Settings.MinTLSVersion
	}
	if len(params.SSL.Settings.Ciphers) > 0 {
		ssl.Settings.Ciphers = params.SSL.Settings.Ciphers
	}

	hostname.SSL = ssl

	return hostname
}

// UpToDate checks if the spec is up to date with the observed hostname
func UpToDate(spec *v1alpha1.CustomHostnameParameters, hostname cloudflare.CustomHostname) bool {
	// Check hostname
	if spec.Hostname != hostname.Hostname {
		return false
	}

	// Check custom origin server
	if spec.CustomOriginServer != nil && *spec.CustomOriginServer != hostname.CustomOriginServer {
		return false
	}

	// Check SSL settings
	if hostname.SSL != nil {
		if spec.SSL.Method != nil && *spec.SSL.Method != hostname.SSL.Method {
			return false
		}
		if spec.SSL.Type != nil && *spec.SSL.Type != hostname.SSL.Type {
			return false
		}
		if spec.SSL.Wildcard != nil && *spec.SSL.Wildcard != *hostname.SSL.Wildcard {
			return false
		}
		if spec.SSL.CustomCertificate != nil && *spec.SSL.CustomCertificate != hostname.SSL.CustomCertificate {
			return false
		}
		if spec.SSL.CustomKey != nil && *spec.SSL.CustomKey != hostname.SSL.CustomKey {
			return false
		}

		// Check SSL settings - cloudflare.CustomHostnameSSLSettings is a struct, not a pointer
		if spec.SSL.Settings.HTTP2 != nil && *spec.SSL.Settings.HTTP2 != hostname.SSL.Settings.HTTP2 {
			return false
		}
		if spec.SSL.Settings.TLS13 != nil && *spec.SSL.Settings.TLS13 != hostname.SSL.Settings.TLS13 {
			return false
		}
		if spec.SSL.Settings.MinTLSVersion != nil && *spec.SSL.Settings.MinTLSVersion != hostname.SSL.Settings.MinTLSVersion {
			return false
		}
	}

	return true
}