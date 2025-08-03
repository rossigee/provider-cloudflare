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

package fallbackorigin

import (
	"context"
	"net/http"

	"github.com/cloudflare/cloudflare-go"

	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errFallbackOriginNotFound = "Fallback Origin not found"
)

// Client is a Cloudflare SSL for SaaS Fallback Origin API client
type Client interface {
	FallbackOrigin(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error)
	UpdateFallbackOrigin(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error)
	DeleteFallbackOrigin(ctx context.Context, zoneID string) error
}

type client struct {
	cf *cloudflare.API
}

// NewClient returns a new FallbackOrigin client
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &client{cf: cf}, nil
}

// FallbackOrigin retrieves Fallback Origin for a zone
func (c *client) FallbackOrigin(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
	response, err := c.cf.CustomHostnameFallbackOrigin(ctx, zoneID)
	if err != nil {
		return cloudflare.CustomHostnameFallbackOrigin{}, err
	}

	return response, nil
}

// UpdateFallbackOrigin updates Fallback Origin for a zone
func (c *client) UpdateFallbackOrigin(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
	response, err := c.cf.UpdateCustomHostnameFallbackOrigin(ctx, zoneID, origin)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// DeleteFallbackOrigin deletes Fallback Origin for a zone
func (c *client) DeleteFallbackOrigin(ctx context.Context, zoneID string) error {
	err := c.cf.DeleteCustomHostnameFallbackOrigin(ctx, zoneID)
	return err
}

// IsFallbackOriginNotFound returns true if the error indicates the fallback origin was not found
func IsFallbackOriginNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errFallbackOriginNotFound ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}

// GenerateObservation creates observation data from Fallback Origin
func GenerateObservation(origin cloudflare.CustomHostnameFallbackOrigin) v1alpha1.FallbackOriginObservation {
	obs := v1alpha1.FallbackOriginObservation{
		Status: origin.Status,
		Errors: origin.Errors,
	}

	return obs
}

// ParametersToFallbackOrigin converts FallbackOriginParameters to cloudflare.CustomHostnameFallbackOrigin
func ParametersToFallbackOrigin(params v1alpha1.FallbackOriginParameters) cloudflare.CustomHostnameFallbackOrigin {
	origin := cloudflare.CustomHostnameFallbackOrigin{}

	if params.Origin != nil {
		origin.Origin = *params.Origin
	}

	return origin
}

// UpToDate checks if the spec is up to date with the observed origin
func UpToDate(spec *v1alpha1.FallbackOriginParameters, origin cloudflare.CustomHostnameFallbackOrigin) bool {
	// If no origin specified in spec, check if origin is empty
	if spec.Origin == nil {
		return origin.Origin == ""
	}

	// Check if the specified origin matches the observed origin
	return *spec.Origin == origin.Origin
}