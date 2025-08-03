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

package loadbalancing

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreatePool = "failed to create load balancer pool"
	errGetPool    = "failed to get load balancer pool"
	errUpdatePool = "failed to update load balancer pool"
	errDeletePool = "failed to delete load balancer pool"
)

// PoolClient interface for Cloudflare Load Balancer Pool operations
type PoolClient interface {
	CreatePool(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	GetPool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	UpdatePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	DeletePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error
}

// NewPoolClient creates a new Cloudflare Load Balancer Pool client
func NewPoolClient(cfg clients.Config, httpClient *http.Client) (PoolClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &poolClient{api: api}, nil
}

type poolClient struct {
	api *cloudflare.API
}

// CreatePool creates a new Cloudflare load balancer pool
func (c *poolClient) CreatePool(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	pool := cloudflare.LoadBalancerPool{
		Name:    params.Name,
		Origins: convertOriginsToCloudflare(params.Origins),
	}

	if params.Description != nil {
		pool.Description = *params.Description
	}

	if params.Enabled != nil {
		pool.Enabled = *params.Enabled
	}

	if params.MinimumOrigins != nil {
		pool.MinimumOrigins = params.MinimumOrigins
	}

	if params.Monitor != nil {
		pool.Monitor = *params.Monitor
	}

	if params.NotificationEmail != nil {
		pool.NotificationEmail = *params.NotificationEmail
	}

	if params.Latitude != nil {
		if lat, err := strconv.ParseFloat(*params.Latitude, 32); err == nil {
			latFloat := float32(lat)
			pool.Latitude = &latFloat
		}
	}

	if params.Longitude != nil {
		if lng, err := strconv.ParseFloat(*params.Longitude, 32); err == nil {
			lngFloat := float32(lng)
			pool.Longitude = &lngFloat
		}
	}

	if params.LoadShedding != nil {
		pool.LoadShedding = convertLoadSheddingToCloudflare(*params.LoadShedding)
	}

	if params.OriginSteering != nil {
		pool.OriginSteering = convertOriginSteeringToCloudflare(*params.OriginSteering)
	}

	if len(params.CheckRegions) > 0 {
		pool.CheckRegions = params.CheckRegions
	}

	createParams := cloudflare.CreateLoadBalancerPoolParams{
		LoadBalancerPool: pool,
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	result, err := c.api.CreateLoadBalancerPool(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreatePool)
	}

	return &result, nil
}

// GetPool retrieves a Cloudflare load balancer pool
func (c *poolClient) GetPool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	pool, err := c.api.GetLoadBalancerPool(ctx, rc, poolID)
	if err != nil {
		return nil, errors.Wrap(err, errGetPool)
	}

	return &pool, nil
}

// UpdatePool updates a Cloudflare load balancer pool
func (c *poolClient) UpdatePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	pool := cloudflare.LoadBalancerPool{
		ID:      poolID,
		Name:    params.Name,
		Origins: convertOriginsToCloudflare(params.Origins),
	}

	if params.Description != nil {
		pool.Description = *params.Description
	}

	if params.Enabled != nil {
		pool.Enabled = *params.Enabled
	}

	if params.MinimumOrigins != nil {
		pool.MinimumOrigins = params.MinimumOrigins
	}

	if params.Monitor != nil {
		pool.Monitor = *params.Monitor
	}

	if params.NotificationEmail != nil {
		pool.NotificationEmail = *params.NotificationEmail
	}

	if params.Latitude != nil {
		if lat, err := strconv.ParseFloat(*params.Latitude, 32); err == nil {
			latFloat := float32(lat)
			pool.Latitude = &latFloat
		}
	}

	if params.Longitude != nil {
		if lng, err := strconv.ParseFloat(*params.Longitude, 32); err == nil {
			lngFloat := float32(lng)
			pool.Longitude = &lngFloat
		}
	}

	if params.LoadShedding != nil {
		pool.LoadShedding = convertLoadSheddingToCloudflare(*params.LoadShedding)
	}

	if params.OriginSteering != nil {
		pool.OriginSteering = convertOriginSteeringToCloudflare(*params.OriginSteering)
	}

	if len(params.CheckRegions) > 0 {
		pool.CheckRegions = params.CheckRegions
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	updateParams := cloudflare.UpdateLoadBalancerPoolParams{
		LoadBalancer: pool,
	}

	result, err := c.api.UpdateLoadBalancerPool(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdatePool)
	}

	return &result, nil
}

// DeletePool deletes a Cloudflare load balancer pool
func (c *poolClient) DeletePool(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return errors.New("either zone or account must be specified")
	}

	err := c.api.DeleteLoadBalancerPool(ctx, rc, poolID)
	if err != nil {
		return errors.Wrap(err, errDeletePool)
	}

	return nil
}

// IsPoolNotFound checks if error indicates pool not found
func IsPoolNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// convertOriginsToCloudflare converts v1alpha1 origins to Cloudflare API format
func convertOriginsToCloudflare(origins []v1alpha1.LoadBalancerOrigin) []cloudflare.LoadBalancerOrigin {
	var cfOrigins []cloudflare.LoadBalancerOrigin

	for _, origin := range origins {
		cfOrigin := cloudflare.LoadBalancerOrigin{
			Name:    origin.Name,
			Address: origin.Address,
		}

		if origin.Enabled != nil {
			cfOrigin.Enabled = *origin.Enabled
		} else {
			cfOrigin.Enabled = true // Default to enabled
		}

		if origin.Weight != nil {
			if weight, err := strconv.ParseFloat(*origin.Weight, 64); err == nil {
				cfOrigin.Weight = weight
			}
		}

		if origin.Header != nil {
			cfOrigin.Header = origin.Header
		}

		if origin.VirtualNetworkID != nil {
			cfOrigin.VirtualNetworkID = *origin.VirtualNetworkID
		}

		cfOrigins = append(cfOrigins, cfOrigin)
	}

	return cfOrigins
}

// convertLoadSheddingToCloudflare converts load shedding to Cloudflare format
func convertLoadSheddingToCloudflare(loadShedding v1alpha1.LoadBalancerLoadShedding) *cloudflare.LoadBalancerLoadShedding {
	cfLoadShedding := &cloudflare.LoadBalancerLoadShedding{}

	if loadShedding.DefaultPercent != nil {
		if percent, err := strconv.ParseFloat(*loadShedding.DefaultPercent, 32); err == nil {
			cfLoadShedding.DefaultPercent = float32(percent)
		}
	}

	if loadShedding.DefaultPolicy != nil {
		cfLoadShedding.DefaultPolicy = *loadShedding.DefaultPolicy
	}

	if loadShedding.SessionPercent != nil {
		if percent, err := strconv.ParseFloat(*loadShedding.SessionPercent, 32); err == nil {
			cfLoadShedding.SessionPercent = float32(percent)
		}
	}

	if loadShedding.SessionPolicy != nil {
		cfLoadShedding.SessionPolicy = *loadShedding.SessionPolicy
	}

	return cfLoadShedding
}

// convertOriginSteeringToCloudflare converts origin steering to Cloudflare format
func convertOriginSteeringToCloudflare(steering v1alpha1.LoadBalancerOriginSteering) *cloudflare.LoadBalancerOriginSteering {
	cfSteering := &cloudflare.LoadBalancerOriginSteering{}

	if steering.Policy != nil {
		cfSteering.Policy = *steering.Policy
	}

	return cfSteering
}

// GeneratePoolObservation creates observation from Cloudflare load balancer pool
func GeneratePoolObservation(pool *cloudflare.LoadBalancerPool) v1alpha1.LoadBalancerPoolObservation {
	observation := v1alpha1.LoadBalancerPoolObservation{
		ID: pool.ID,
	}

	if pool.CreatedOn != nil {
		createdOn := pool.CreatedOn.String()
		observation.CreatedOn = &createdOn
	}

	if pool.ModifiedOn != nil {
		modifiedOn := pool.ModifiedOn.String()
		observation.ModifiedOn = &modifiedOn
	}

	observation.Healthy = pool.Healthy

	return observation
}

// IsPoolUpToDate determines if the Cloudflare load balancer pool is up to date
func IsPoolUpToDate(params *v1alpha1.LoadBalancerPoolParameters, pool *cloudflare.LoadBalancerPool) bool {
	if params.Name != pool.Name {
		return false
	}

	if params.Description != nil && *params.Description != pool.Description {
		return false
	}

	if params.Description == nil && pool.Description != "" {
		return false
	}

	if params.Enabled != nil && *params.Enabled != pool.Enabled {
		return false
	}

	if len(params.Origins) != len(pool.Origins) {
		return false
	}

	// For simplicity, we'll consider origins changed if the count differs
	// A more sophisticated comparison could be implemented if needed

	return true
}