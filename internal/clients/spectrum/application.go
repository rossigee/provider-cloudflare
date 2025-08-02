/*
Copyright 2021 The Crossplane Authors.

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

package spectrum

import (
	"context"
	"net"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/spectrum/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errApplicationNotFound = "Application not found"
)

// Client is a Cloudflare Spectrum API client
type Client interface {
	SpectrumApplication(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error)
	CreateSpectrumApplication(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error)
	UpdateSpectrumApplication(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error
	DeleteSpectrumApplication(ctx context.Context, zoneID, applicationID string) error
}

type client struct {
	cf *cloudflare.API
}

// NewClient returns a new Spectrum Applications client
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &client{cf: cf}, nil
}

// SpectrumApplication retrieves a Spectrum Application
func (c *client) SpectrumApplication(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
	return c.cf.SpectrumApplication(ctx, zoneID, applicationID)
}

// CreateSpectrumApplication creates a new Spectrum Application
func (c *client) CreateSpectrumApplication(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
	app := cloudflare.SpectrumApplication{
		Protocol: params.Protocol,
		DNS: cloudflare.SpectrumApplicationDNS{
			Type: params.DNS.Type,
			Name: params.DNS.Name,
		},
		OriginDirect: params.OriginDirect,
	}

	// Set origin port if specified
	if params.OriginPort != nil {
		app.OriginPort = &cloudflare.SpectrumApplicationOriginPort{}
		if params.OriginPort.Port != nil {
			app.OriginPort.Port = uint16(*params.OriginPort.Port)
		}
		if params.OriginPort.Start != nil {
			app.OriginPort.Start = uint16(*params.OriginPort.Start)
		}
		if params.OriginPort.End != nil {
			app.OriginPort.End = uint16(*params.OriginPort.End)
		}
	}

	// Set origin DNS if specified
	if params.OriginDNS != nil {
		app.OriginDNS = &cloudflare.SpectrumApplicationOriginDNS{
			Name: params.OriginDNS.Name,
		}
	}

	// Set edge IPs if specified
	if params.EdgeIPs != nil {
		app.EdgeIPs = &cloudflare.SpectrumApplicationEdgeIPs{
			Type: cloudflare.SpectrumApplicationEdgeType(params.EdgeIPs.Type),
		}

		if params.EdgeIPs.Connectivity != nil {
			app.EdgeIPs.Connectivity = (*cloudflare.SpectrumApplicationConnectivity)(params.EdgeIPs.Connectivity)
		}

		if params.EdgeIPs.IPs != nil {
			ips, err := ConvertIPs(params.EdgeIPs.IPs)
			if err != nil {
				return cloudflare.SpectrumApplication{}, err
			}
			app.EdgeIPs.IPs = ips
		}
	}

	// Set optional fields
	if params.ProxyProtocol != nil {
		app.ProxyProtocol = cloudflare.ProxyProtocol(*params.ProxyProtocol)
	}

	if params.IPFirewall != nil {
		app.IPFirewall = *params.IPFirewall
	}

	if params.TLS != nil {
		app.TLS = *params.TLS
	}

	if params.TrafficType != nil {
		app.TrafficType = *params.TrafficType
	}

	if params.ArgoSmartRouting != nil {
		app.ArgoSmartRouting = *params.ArgoSmartRouting
	}

	return c.cf.CreateSpectrumApplication(ctx, zoneID, app)
}

// UpdateSpectrumApplication updates an existing Spectrum Application
func (c *client) UpdateSpectrumApplication(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error {
	app := cloudflare.SpectrumApplication{
		ID:           applicationID,
		Protocol:     params.Protocol,
		DNS: cloudflare.SpectrumApplicationDNS{
			Type: params.DNS.Type,
			Name: params.DNS.Name,
		},
		OriginDirect: params.OriginDirect,
	}

	// Set origin port if specified
	if params.OriginPort != nil {
		app.OriginPort = &cloudflare.SpectrumApplicationOriginPort{}
		if params.OriginPort.Port != nil {
			app.OriginPort.Port = uint16(*params.OriginPort.Port)
		}
		if params.OriginPort.Start != nil {
			app.OriginPort.Start = uint16(*params.OriginPort.Start)
		}
		if params.OriginPort.End != nil {
			app.OriginPort.End = uint16(*params.OriginPort.End)
		}
	}

	// Set origin DNS if specified
	if params.OriginDNS != nil {
		app.OriginDNS = &cloudflare.SpectrumApplicationOriginDNS{
			Name: params.OriginDNS.Name,
		}
	}

	// Set edge IPs if specified
	if params.EdgeIPs != nil {
		app.EdgeIPs = &cloudflare.SpectrumApplicationEdgeIPs{
			Type: cloudflare.SpectrumApplicationEdgeType(params.EdgeIPs.Type),
		}

		if params.EdgeIPs.Connectivity != nil {
			app.EdgeIPs.Connectivity = (*cloudflare.SpectrumApplicationConnectivity)(params.EdgeIPs.Connectivity)
		}

		if params.EdgeIPs.IPs != nil {
			ips, err := ConvertIPs(params.EdgeIPs.IPs)
			if err != nil {
				return err
			}
			app.EdgeIPs.IPs = ips
		}
	}

	// Set optional fields
	if params.ProxyProtocol != nil {
		app.ProxyProtocol = cloudflare.ProxyProtocol(*params.ProxyProtocol)
	}

	if params.IPFirewall != nil {
		app.IPFirewall = *params.IPFirewall
	}

	if params.TLS != nil {
		app.TLS = *params.TLS
	}

	if params.TrafficType != nil {
		app.TrafficType = *params.TrafficType
	}

	if params.ArgoSmartRouting != nil {
		app.ArgoSmartRouting = *params.ArgoSmartRouting
	}

	_, err := c.cf.UpdateSpectrumApplication(ctx, zoneID, applicationID, app)
	return err
}

// DeleteSpectrumApplication deletes a Spectrum Application
func (c *client) DeleteSpectrumApplication(ctx context.Context, zoneID, applicationID string) error {
	return c.cf.DeleteSpectrumApplication(ctx, zoneID, applicationID)
}

// IsApplicationNotFound returns true if the error indicates the application was not found
func IsApplicationNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare not found error or our specific error message
	return err.Error() == errApplicationNotFound || 
		   err.Error() == "404" ||
		   err.Error() == "Not found"
}

// GenerateObservation creates observation data from a Spectrum Application
func GenerateObservation(app cloudflare.SpectrumApplication) v1alpha1.ApplicationObservation {
	obs := v1alpha1.ApplicationObservation{}

	if app.CreatedOn != nil && !app.CreatedOn.IsZero() {
		obs.CreatedOn = &metav1.Time{Time: *app.CreatedOn}
	}

	if app.ModifiedOn != nil && !app.ModifiedOn.IsZero() {
		obs.ModifiedOn = &metav1.Time{Time: *app.ModifiedOn}
	}

	return obs
}

// LateInitialize fills in any missing fields in the spec from the observed application
func LateInitialize(spec *v1alpha1.ApplicationParameters, app cloudflare.SpectrumApplication) bool {
	// No late initialization needed for Spectrum applications currently
	return false
}

// UpToDate checks if the spec is up to date with the observed application
func UpToDate(spec *v1alpha1.ApplicationParameters, app cloudflare.SpectrumApplication) bool {
	// Check protocol
	if spec.Protocol != app.Protocol {
		return false
	}

	// Check DNS configuration
	if spec.DNS.Type != app.DNS.Type || spec.DNS.Name != app.DNS.Name {
		return false
	}

	// Check origin direct
	if len(spec.OriginDirect) != len(app.OriginDirect) {
		return false
	}
	for i, origin := range spec.OriginDirect {
		if origin != app.OriginDirect[i] {
			return false
		}
	}

	// Additional checks for other fields would go here...
	
	return true
}

// ConvertIPs converts string IPs to net.IP slice
func ConvertIPs(ipStrings []string) ([]net.IP, error) {
	ips := make([]net.IP, len(ipStrings))
	for i, ipStr := range ipStrings {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, errors.Errorf("invalid IP address: %s", ipStr)
		}
		ips[i] = ip
	}
	return ips, nil
}