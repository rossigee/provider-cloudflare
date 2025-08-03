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
	errCreateLoadBalancer = "failed to create load balancer"
	errGetLoadBalancer    = "failed to get load balancer"
	errUpdateLoadBalancer = "failed to update load balancer"
	errDeleteLoadBalancer = "failed to delete load balancer"
)

// LoadBalancerClient interface for Cloudflare Load Balancer operations
type LoadBalancerClient interface {
	CreateLoadBalancer(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	GetLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	UpdateLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error
}

// NewLoadBalancerClient creates a new Cloudflare Load Balancer client
func NewLoadBalancerClient(cfg clients.Config, httpClient *http.Client) (LoadBalancerClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &loadBalancerClient{api: api}, nil
}

type loadBalancerClient struct {
	api *cloudflare.API
}

// CreateLoadBalancer creates a new Cloudflare load balancer
func (c *loadBalancerClient) CreateLoadBalancer(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	createParams := cloudflare.CreateLoadBalancerParams{
		DefaultPools: params.DefaultPools,
	}

	if params.Name != nil {
		createParams.Name = *params.Name
	}

	if params.Description != nil {
		createParams.Description = *params.Description
	}

	if params.TTL != nil {
		createParams.TTL = *params.TTL
	}

	if params.FallbackPool != nil {
		createParams.FallbackPool = *params.FallbackPool
	}

	if params.RegionPools != nil {
		createParams.RegionPools = params.RegionPools
	}

	if params.PopPools != nil {
		createParams.PopPools = params.PopPools
	}

	if params.CountryPools != nil {
		createParams.CountryPools = params.CountryPools
	}

	if params.Proxied != nil {
		createParams.Proxied = *params.Proxied
	}

	if params.Enabled != nil {
		createParams.Enabled = *params.Enabled
	}

	if params.SessionAffinity != nil {
		createParams.SessionAffinity = *params.SessionAffinity
	}

	if params.SessionAffinityTTL != nil {
		createParams.SessionAffinityTTL = *params.SessionAffinityTTL
	}

	if params.SessionAffinityAttributes != nil {
		createParams.SessionAffinityAttributes = convertSessionAffinityAttributesToCloudflare(*params.SessionAffinityAttributes)
	}

	if params.SteeringPolicy != nil {
		createParams.SteeringPolicy = *params.SteeringPolicy
	}

	if len(params.Rules) > 0 {
		createParams.Rules = convertRulesToCloudflare(params.Rules)
	}

	if params.RandomSteering != nil {
		createParams.RandomSteering = convertRandomSteeringToCloudflare(*params.RandomSteering)
	}

	if params.AdaptiveRouting != nil {
		createParams.AdaptiveRouting = convertAdaptiveRoutingToCloudflare(*params.AdaptiveRouting)
	}

	if params.LocationStrategy != nil {
		createParams.LocationStrategy = convertLocationStrategyToCloudflare(*params.LocationStrategy)
	}

	rc := cloudflare.ZoneIdentifier(params.Zone)

	lb, err := c.api.CreateLoadBalancer(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateLoadBalancer)
	}

	return &lb, nil
}

// GetLoadBalancer retrieves a Cloudflare load balancer
func (c *loadBalancerClient) GetLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	lb, err := c.api.GetLoadBalancer(ctx, rc, lbID)
	if err != nil {
		return nil, errors.Wrap(err, errGetLoadBalancer)
	}

	return &lb, nil
}

// UpdateLoadBalancer updates a Cloudflare load balancer
func (c *loadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	updateParams := cloudflare.UpdateLoadBalancerParams{
		ID:           lbID,
		DefaultPools: params.DefaultPools,
	}

	if params.Name != nil {
		updateParams.Name = *params.Name
	}

	if params.Description != nil {
		updateParams.Description = *params.Description
	}

	if params.TTL != nil {
		updateParams.TTL = *params.TTL
	}

	if params.FallbackPool != nil {
		updateParams.FallbackPool = *params.FallbackPool
	}

	if params.RegionPools != nil {
		updateParams.RegionPools = params.RegionPools
	}

	if params.PopPools != nil {
		updateParams.PopPools = params.PopPools
	}

	if params.CountryPools != nil {
		updateParams.CountryPools = params.CountryPools
	}

	if params.Proxied != nil {
		updateParams.Proxied = *params.Proxied
	}

	if params.Enabled != nil {
		updateParams.Enabled = *params.Enabled
	}

	if params.SessionAffinity != nil {
		updateParams.SessionAffinity = *params.SessionAffinity
	}

	if params.SessionAffinityTTL != nil {
		updateParams.SessionAffinityTTL = *params.SessionAffinityTTL
	}

	if params.SessionAffinityAttributes != nil {
		updateParams.SessionAffinityAttributes = convertSessionAffinityAttributesToCloudflare(*params.SessionAffinityAttributes)
	}

	if params.SteeringPolicy != nil {
		updateParams.SteeringPolicy = *params.SteeringPolicy
	}

	if len(params.Rules) > 0 {
		updateParams.Rules = convertRulesToCloudflare(params.Rules)
	}

	if params.RandomSteering != nil {
		updateParams.RandomSteering = convertRandomSteeringToCloudflare(*params.RandomSteering)
	}

	if params.AdaptiveRouting != nil {
		updateParams.AdaptiveRouting = convertAdaptiveRoutingToCloudflare(*params.AdaptiveRouting)
	}

	if params.LocationStrategy != nil {
		updateParams.LocationStrategy = convertLocationStrategyToCloudflare(*params.LocationStrategy)
	}

	rc := cloudflare.ZoneIdentifier(params.Zone)

	lb, err := c.api.UpdateLoadBalancer(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateLoadBalancer)
	}

	return &lb, nil
}

// DeleteLoadBalancer deletes a Cloudflare load balancer
func (c *loadBalancerClient) DeleteLoadBalancer(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	err := c.api.DeleteLoadBalancer(ctx, rc, lbID)
	if err != nil {
		return errors.Wrap(err, errDeleteLoadBalancer)
	}

	return nil
}

// IsLoadBalancerNotFound checks if error indicates load balancer not found
func IsLoadBalancerNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// convertSessionAffinityAttributesToCloudflare converts session affinity attributes to Cloudflare format
func convertSessionAffinityAttributesToCloudflare(attrs v1alpha1.SessionAffinityAttributes) *cloudflare.SessionAffinityAttributes {
	cfAttrs := &cloudflare.SessionAffinityAttributes{}

	if attrs.SameSite != nil {
		cfAttrs.SameSite = *attrs.SameSite
	}

	if attrs.Secure != nil {
		cfAttrs.Secure = *attrs.Secure
	}

	if attrs.DrainDuration != nil {
		cfAttrs.DrainDuration = *attrs.DrainDuration
	}

	if attrs.ZeroDowntimeFailover != nil {
		cfAttrs.ZeroDowntimeFailover = *attrs.ZeroDowntimeFailover
	}

	if len(attrs.Headers) > 0 {
		cfAttrs.Headers = attrs.Headers
	}

	if attrs.RequireAllHeaders != nil {
		cfAttrs.RequireAllHeaders = *attrs.RequireAllHeaders
	}

	return cfAttrs
}

// convertRulesToCloudflare converts load balancer rules to Cloudflare format
func convertRulesToCloudflare(rules []v1alpha1.LoadBalancerRule) []cloudflare.LoadBalancerRule {
	var cfRules []cloudflare.LoadBalancerRule

	for _, rule := range rules {
		cfRule := cloudflare.LoadBalancerRule{
			Name:      rule.Name,
			Condition: rule.Condition,
			Priority:  rule.Priority,
		}

		if rule.Disabled != nil {
			cfRule.Disabled = *rule.Disabled
		}

		if rule.Terminates != nil {
			cfRule.Terminates = *rule.Terminates
		}

		if rule.FixedResponse != nil {
			cfRule.FixedResponse = convertFixedResponseToCloudflare(*rule.FixedResponse)
		}

		if rule.Overrides != nil {
			cfRule.Overrides = convertRuleOverridesToCloudflare(*rule.Overrides)
		}

		cfRules = append(cfRules, cfRule)
	}

	return cfRules
}

// convertFixedResponseToCloudflare converts fixed response to Cloudflare format
func convertFixedResponseToCloudflare(fixedResponse v1alpha1.LoadBalancerFixedResponse) *cloudflare.LoadBalancerFixedResponse {
	cfFixedResponse := &cloudflare.LoadBalancerFixedResponse{}

	if fixedResponse.MessageBody != nil {
		cfFixedResponse.MessageBody = *fixedResponse.MessageBody
	}

	if fixedResponse.StatusCode != nil {
		cfFixedResponse.StatusCode = *fixedResponse.StatusCode
	}

	if fixedResponse.ContentType != nil {
		cfFixedResponse.ContentType = *fixedResponse.ContentType
	}

	if fixedResponse.Location != nil {
		cfFixedResponse.Location = *fixedResponse.Location
	}

	return cfFixedResponse
}

// convertRuleOverridesToCloudflare converts rule overrides to Cloudflare format
func convertRuleOverridesToCloudflare(overrides v1alpha1.LoadBalancerRuleOverrides) *cloudflare.LoadBalancerRuleOverrides {
	cfOverrides := &cloudflare.LoadBalancerRuleOverrides{}

	if overrides.SessionAffinity != nil {
		cfOverrides.SessionAffinity = *overrides.SessionAffinity
	}

	if overrides.SessionAffinityTTL != nil {
		cfOverrides.SessionAffinityTTL = *overrides.SessionAffinityTTL
	}

	if overrides.SessionAffinityAttributes != nil {
		cfOverrides.SessionAffinityAttributes = convertSessionAffinityAttributesToCloudflare(*overrides.SessionAffinityAttributes)
	}

	if overrides.TTL != nil {
		cfOverrides.TTL = *overrides.TTL
	}

	if overrides.SteeringPolicy != nil {
		cfOverrides.SteeringPolicy = *overrides.SteeringPolicy
	}

	if overrides.FallbackPool != nil {
		cfOverrides.FallbackPool = *overrides.FallbackPool
	}

	if len(overrides.DefaultPools) > 0 {
		cfOverrides.DefaultPools = overrides.DefaultPools
	}

	if overrides.PopPools != nil {
		cfOverrides.PopPools = overrides.PopPools
	}

	if overrides.RegionPools != nil {
		cfOverrides.RegionPools = overrides.RegionPools
	}

	if overrides.CountryPools != nil {
		cfOverrides.CountryPools = overrides.CountryPools
	}

	if overrides.RandomSteering != nil {
		cfOverrides.RandomSteering = convertRandomSteeringToCloudflare(*overrides.RandomSteering)
	}

	if overrides.AdaptiveRouting != nil {
		cfOverrides.AdaptiveRouting = convertAdaptiveRoutingToCloudflare(*overrides.AdaptiveRouting)
	}

	if overrides.LocationStrategy != nil {
		cfOverrides.LocationStrategy = convertLocationStrategyToCloudflare(*overrides.LocationStrategy)
	}

	return cfOverrides
}

// convertRandomSteeringToCloudflare converts random steering to Cloudflare format
func convertRandomSteeringToCloudflare(steering v1alpha1.RandomSteering) *cloudflare.RandomSteering {
	cfSteering := &cloudflare.RandomSteering{}

	if steering.DefaultWeight != nil {
		if weight, err := strconv.ParseFloat(*steering.DefaultWeight, 64); err == nil {
			cfSteering.DefaultWeight = weight
		}
	}

	if len(steering.PoolWeights) > 0 {
		cfSteering.PoolWeights = make(map[string]float64)
		for poolID, weightStr := range steering.PoolWeights {
			if weight, err := strconv.ParseFloat(weightStr, 64); err == nil {
				cfSteering.PoolWeights[poolID] = weight
			}
		}
	}

	return cfSteering
}

// convertAdaptiveRoutingToCloudflare converts adaptive routing to Cloudflare format
func convertAdaptiveRoutingToCloudflare(routing v1alpha1.AdaptiveRouting) *cloudflare.AdaptiveRouting {
	cfRouting := &cloudflare.AdaptiveRouting{}

	if routing.FailoverAcrossPools != nil {
		cfRouting.FailoverAcrossPools = *routing.FailoverAcrossPools
	}

	return cfRouting
}

// convertLocationStrategyToCloudflare converts location strategy to Cloudflare format
func convertLocationStrategyToCloudflare(strategy v1alpha1.LocationStrategy) *cloudflare.LocationStrategy {
	cfStrategy := &cloudflare.LocationStrategy{}

	if strategy.Mode != nil {
		cfStrategy.Mode = *strategy.Mode
	}

	if strategy.PreferECSRegion != nil {
		cfStrategy.PreferECSRegion = *strategy.PreferECSRegion
	}

	return cfStrategy
}

// GenerateLoadBalancerObservation creates observation from Cloudflare load balancer
func GenerateLoadBalancerObservation(lb *cloudflare.LoadBalancer) v1alpha1.LoadBalancerObservation {
	observation := v1alpha1.LoadBalancerObservation{
		ID: lb.ID,
	}

	if lb.CreatedOn != nil {
		createdOn := lb.CreatedOn.String()
		observation.CreatedOn = &createdOn
	}

	if lb.ModifiedOn != nil {
		modifiedOn := lb.ModifiedOn.String()
		observation.ModifiedOn = &modifiedOn
	}

	return observation
}

// IsLoadBalancerUpToDate determines if the Cloudflare load balancer is up to date
func IsLoadBalancerUpToDate(params *v1alpha1.LoadBalancerParameters, lb *cloudflare.LoadBalancer) bool {
	if params.Name != nil && *params.Name != lb.Name {
		return false
	}

	if params.Description != nil && *params.Description != lb.Description {
		return false
	}

	if params.Description == nil && lb.Description != "" {
		return false
	}

	if params.TTL != nil && *params.TTL != lb.TTL {
		return false
	}

	if params.FallbackPool != nil && *params.FallbackPool != lb.FallbackPool {
		return false
	}

	if params.Proxied != nil && *params.Proxied != lb.Proxied {
		return false
	}

	if params.Enabled != nil && *params.Enabled != lb.Enabled {
		return false
	}

	if params.SessionAffinity != nil && *params.SessionAffinity != lb.SessionAffinity {
		return false
	}

	if params.SessionAffinityTTL != nil && *params.SessionAffinityTTL != lb.SessionAffinityTTL {
		return false
	}

	if params.SteeringPolicy != nil && *params.SteeringPolicy != lb.SteeringPolicy {
		return false
	}

	// For complex comparisons like pools, rules, etc., we'll keep it simple
	// A more sophisticated comparison could be implemented if needed
	if len(params.DefaultPools) != len(lb.DefaultPools) {
		return false
	}

	return true
}