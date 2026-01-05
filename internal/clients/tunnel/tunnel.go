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

package tunnel

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/tunnel/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// TunnelAPI defines the interface for Tunnel operations
type TunnelAPI interface {
	ArgoTunnels(ctx context.Context, accountID string) ([]cloudflare.ArgoTunnel, error)
	ArgoTunnel(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error)
	CreateArgoTunnel(ctx context.Context, accountID, name, secret string) (cloudflare.ArgoTunnel, error)
	DeleteArgoTunnel(ctx context.Context, accountID, tunnelUUID string) error
	CleanupArgoTunnelConnections(ctx context.Context, accountID, tunnelUUID string) error
}

// CloudflareTunnelClient is a Cloudflare API client for Tunnels.
type CloudflareTunnelClient struct {
	client TunnelAPI
}

// NewClient creates a new CloudflareTunnelClient.
func NewClient(client TunnelAPI) *CloudflareTunnelClient {
	return &CloudflareTunnelClient{client: client}
}

// NewClientFromAPI creates a new CloudflareTunnelClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareTunnelClient {
	return NewClient(api)
}

// Get retrieves a Tunnel.
func (c *CloudflareTunnelClient) Get(ctx context.Context, accountID, tunnelID string) (*v1beta1.TunnelObservation, error) {
	tunnel, err := c.client.ArgoTunnel(ctx, accountID, tunnelID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("tunnel not found")
		}
		return nil, errors.Wrap(err, "cannot get tunnel")
	}

	return convertArgoTunnelToObservation(tunnel), nil
}

// Create creates a new Tunnel.
func (c *CloudflareTunnelClient) Create(ctx context.Context, params v1beta1.TunnelParameters) (*v1beta1.TunnelObservation, error) {
	tunnel, err := c.client.CreateArgoTunnel(ctx, params.AccountID, params.Name, params.Secret)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create tunnel")
	}

	return convertArgoTunnelToObservation(tunnel), nil
}

// Update updates a Tunnel. Note: Cloudflare tunnels are immutable except for configuration.
func (c *CloudflareTunnelClient) Update(ctx context.Context, tunnelID string, params v1beta1.TunnelParameters) (*v1beta1.TunnelObservation, error) {
	// Tunnels themselves are not directly updatable via API
	// Configuration updates would be handled separately
	// For now, just return the current state
	tunnel, err := c.client.ArgoTunnel(ctx, params.AccountID, tunnelID)
	if err != nil {
		return nil, errors.Wrap(err, "cannot get tunnel for update")
	}

	return convertArgoTunnelToObservation(tunnel), nil
}

// Delete deletes a Tunnel.
func (c *CloudflareTunnelClient) Delete(ctx context.Context, accountID, tunnelID string) error {
	// First try to cleanup connections
	_ = c.client.CleanupArgoTunnelConnections(ctx, accountID, tunnelID)

	// Then delete the tunnel
	err := c.client.DeleteArgoTunnel(ctx, accountID, tunnelID)
	if err != nil && !isNotFound(err) {
		return errors.Wrap(err, "cannot delete tunnel")
	}
	return nil
}

// IsUpToDate checks if the Tunnel is up to date.
func (c *CloudflareTunnelClient) IsUpToDate(ctx context.Context, params v1beta1.TunnelParameters, obs v1beta1.TunnelObservation) (bool, error) {
	// Compare key parameters
	if params.Name != obs.Name {
		return false, nil
	}

	// Tunnels are primarily identified by name and ID
	// Configuration updates would be handled separately
	return true, nil
}

// convertArgoTunnelToObservation converts cloudflare.ArgoTunnel to TunnelObservation.
func convertArgoTunnelToObservation(tunnel cloudflare.ArgoTunnel) *v1beta1.TunnelObservation {
	obs := &v1beta1.TunnelObservation{
		ID:   tunnel.ID,
		Name: tunnel.Name,
	}

	if tunnel.CreatedAt != nil {
		obs.CreatedAt = &metav1.Time{Time: *tunnel.CreatedAt}
	}

	if tunnel.DeletedAt != nil {
		obs.DeletedAt = &metav1.Time{Time: *tunnel.DeletedAt}
	}

	// Convert connections if available
	if tunnel.Connections != nil {
		obs.Connections = make([]v1beta1.TunnelConnection, len(tunnel.Connections))
		for i, conn := range tunnel.Connections {
			obs.Connections[i] = v1beta1.TunnelConnection{
				ColoName: conn.ColoName,
			}
			// Note: ArgoTunnelConnection only has ColoName, UUID, and IsPendingReconnect
			// The other fields from our API spec are not available in the current Cloudflare API
	}
	}

	return obs
}

// isNotFound checks if an error indicates that the tunnel was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "tunnel not found") ||
		strings.Contains(errStr, "does not exist")
}