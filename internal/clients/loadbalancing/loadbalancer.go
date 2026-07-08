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
	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"net/http"
	"strconv"
)

const (
	errCreateLoadBalancer = "failed to create load balancer"
	errGetLoadBalancer    = "failed to get load balancer"
	errUpdateLoadBalancer = "failed to update load balancer"
	errDeleteLoadBalancer = "failed to delete load balancer"

	errCreateLoadBalancerMonitor = "failed to create load balancer monitor"
	errGetLoadBalancerMonitor    = "failed to get load balancer monitor"
	errUpdateLoadBalancerMonitor = "failed to update load balancer monitor"
	errDeleteLoadBalancerMonitor = "failed to delete load balancer monitor"

	errCreateLoadBalancerPool = "failed to create load balancer pool"
	errGetLoadBalancerPool    = "failed to get load balancer pool"
	errUpdateLoadBalancerPool = "failed to update load balancer pool"
	errDeleteLoadBalancerPool = "failed to delete load balancer pool"
)

// LoadBalancerClient interface for Cloudflare Load Balancer operations
type LoadBalancerClient interface {
	CreateLoadBalancer(ctx context.Context, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	GetLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	UpdateLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) error
}

// MonitorClient interface for Cloudflare Load Balancer Monitor operations
type MonitorClient interface {
	CreateMonitor(ctx context.Context, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	GetMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	UpdateMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	DeleteMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) error
}

// PoolClient interface for Cloudflare Load Balancer Pool operations
type PoolClient interface {
	CreatePool(ctx context.Context, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	GetPool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	UpdatePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error)
	DeletePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) error
}

// NewLoadBalancerClient creates a new Cloudflare Load Balancer client
func NewLoadBalancerClient(cfg clients.Config, httpClient *http.Client) (LoadBalancerClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &loadBalancerClient{api: api}, nil
}

// NewMonitorClient creates a new Cloudflare Load Balancer Monitor client
func NewMonitorClient(cfg clients.Config, httpClient *http.Client) (MonitorClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &monitorClient{api: api}, nil
}

// NewPoolClient creates a new Cloudflare Load Balancer Pool client
func NewPoolClient(cfg clients.Config, httpClient *http.Client) (PoolClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &poolClient{api: api}, nil
}

type loadBalancerClient struct {
	api *cloudflare.API
}

type monitorClient struct {
	api *cloudflare.API
}

type poolClient struct {
	api *cloudflare.API
}

// CreateLoadBalancer creates a new Cloudflare load balancer
func (c *loadBalancerClient) CreateLoadBalancer(ctx context.Context, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	lb := cloudflare.LoadBalancer{
		DefaultPools: params.DefaultPools,
	}

	if params.Name != nil {
		lb.Name = *params.Name
	}

	if params.Description != nil {
		lb.Description = *params.Description
	}

	if params.TTL != nil {
		lb.TTL = *params.TTL
	}

	if params.FallbackPool != nil {
		lb.FallbackPool = *params.FallbackPool
	}

	if params.RegionPools != nil {
		lb.RegionPools = params.RegionPools
	}

	if params.PopPools != nil {
		lb.PopPools = params.PopPools
	}

	if params.CountryPools != nil {
		lb.CountryPools = params.CountryPools
	}

	if params.Proxied != nil {
		lb.Proxied = *params.Proxied
	}

	if params.Enabled != nil {
		lb.Enabled = params.Enabled
	}

	if params.SessionAffinity != nil {
		lb.Persistence = *params.SessionAffinity
	}

	if params.SessionAffinityTTL != nil {
		lb.PersistenceTTL = *params.SessionAffinityTTL
	}

	if params.SessionAffinityAttributes != nil {
		lb.SessionAffinityAttributes = convertSessionAffinityAttributesToCloudflare(*params.SessionAffinityAttributes)
	}

	if params.SteeringPolicy != nil {
		lb.SteeringPolicy = *params.SteeringPolicy
	}

	if len(params.Rules) > 0 {
		lb.Rules = convertRulesToCloudflare(params.Rules)
	}

	if params.RandomSteering != nil {
		lb.RandomSteering = convertRandomSteeringToCloudflare(*params.RandomSteering)
	}

	if params.AdaptiveRouting != nil {
		lb.AdaptiveRouting = convertAdaptiveRoutingToCloudflare(*params.AdaptiveRouting)
	}

	if params.LocationStrategy != nil {
		lb.LocationStrategy = convertLocationStrategyToCloudflare(*params.LocationStrategy)
	}

	createParams := cloudflare.CreateLoadBalancerParams{
		LoadBalancer: lb,
	}

	rc := cloudflare.ZoneIdentifier(params.Zone)

	result, err := c.api.CreateLoadBalancer(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateLoadBalancer)
	}

	return &result, nil
}

// GetLoadBalancer retrieves a Cloudflare load balancer
func (c *loadBalancerClient) GetLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	lb, err := c.api.GetLoadBalancer(ctx, rc, lbID)
	if err != nil {
		return nil, errors.Wrap(err, errGetLoadBalancer)
	}

	return &lb, nil
}

// UpdateLoadBalancer updates a Cloudflare load balancer
func (c *loadBalancerClient) UpdateLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
	lb := cloudflare.LoadBalancer{
		ID:           lbID,
		DefaultPools: params.DefaultPools,
	}

	if params.Name != nil {
		lb.Name = *params.Name
	}

	if params.Description != nil {
		lb.Description = *params.Description
	}

	if params.TTL != nil {
		lb.TTL = *params.TTL
	}

	if params.FallbackPool != nil {
		lb.FallbackPool = *params.FallbackPool
	}

	if params.RegionPools != nil {
		lb.RegionPools = params.RegionPools
	}

	if params.PopPools != nil {
		lb.PopPools = params.PopPools
	}

	if params.CountryPools != nil {
		lb.CountryPools = params.CountryPools
	}

	if params.Proxied != nil {
		lb.Proxied = *params.Proxied
	}

	if params.Enabled != nil {
		lb.Enabled = params.Enabled
	}

	if params.SessionAffinity != nil {
		lb.Persistence = *params.SessionAffinity
	}

	if params.SessionAffinityTTL != nil {
		lb.PersistenceTTL = *params.SessionAffinityTTL
	}

	if params.SessionAffinityAttributes != nil {
		lb.SessionAffinityAttributes = convertSessionAffinityAttributesToCloudflare(*params.SessionAffinityAttributes)
	}

	if params.SteeringPolicy != nil {
		lb.SteeringPolicy = *params.SteeringPolicy
	}

	if len(params.Rules) > 0 {
		lb.Rules = convertRulesToCloudflare(params.Rules)
	}

	if params.RandomSteering != nil {
		lb.RandomSteering = convertRandomSteeringToCloudflare(*params.RandomSteering)
	}

	if params.AdaptiveRouting != nil {
		lb.AdaptiveRouting = convertAdaptiveRoutingToCloudflare(*params.AdaptiveRouting)
	}

	if params.LocationStrategy != nil {
		lb.LocationStrategy = convertLocationStrategyToCloudflare(*params.LocationStrategy)
	}

	updateParams := cloudflare.UpdateLoadBalancerParams{
		LoadBalancer: lb,
	}

	rc := cloudflare.ZoneIdentifier(params.Zone)

	result, err := c.api.UpdateLoadBalancer(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateLoadBalancer)
	}

	return &result, nil
}

// DeleteLoadBalancer deletes a Cloudflare load balancer
func (c *loadBalancerClient) DeleteLoadBalancer(ctx context.Context, lbID string, params v1beta1.LoadBalancerParameters) error {
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
func convertSessionAffinityAttributesToCloudflare(attrs v1beta1.SessionAffinityAttributes) *cloudflare.SessionAffinityAttributes {
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

// convertSessionAffinityAttributesForRuleOverrides converts session affinity attributes for rule overrides to Cloudflare format
func convertSessionAffinityAttributesForRuleOverrides(attrs v1beta1.SessionAffinityAttributes) *cloudflare.LoadBalancerRuleOverridesSessionAffinityAttrs {
	cfAttrs := &cloudflare.LoadBalancerRuleOverridesSessionAffinityAttrs{}

	if attrs.SameSite != nil {
		cfAttrs.SameSite = *attrs.SameSite
	}

	if attrs.Secure != nil {
		cfAttrs.Secure = *attrs.Secure
	}

	if attrs.ZeroDowntimeFailover != nil {
		cfAttrs.ZeroDowntimeFailover = *attrs.ZeroDowntimeFailover
	}

	if len(attrs.Headers) > 0 {
		cfAttrs.Headers = attrs.Headers
	}

	if attrs.RequireAllHeaders != nil {
		cfAttrs.RequireAllHeaders = attrs.RequireAllHeaders
	}

	// Note: DrainDuration is not included as it's not supported in rule overrides

	return cfAttrs
}

// convertRulesToCloudflare converts load balancer rules to Cloudflare format
func convertRulesToCloudflare(rules []v1beta1.LoadBalancerRule) []*cloudflare.LoadBalancerRule {
	var cfRules []*cloudflare.LoadBalancerRule

	for _, rule := range rules {
		cfRule := &cloudflare.LoadBalancerRule{
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
			cfRule.Overrides = *convertRuleOverridesToCloudflare(*rule.Overrides)
		}

		cfRules = append(cfRules, cfRule)
	}

	return cfRules
}

// convertFixedResponseToCloudflare converts fixed response to Cloudflare format
func convertFixedResponseToCloudflare(fixedResponse v1beta1.LoadBalancerFixedResponse) *cloudflare.LoadBalancerFixedResponseData {
	cfFixedResponse := &cloudflare.LoadBalancerFixedResponseData{}

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
func convertRuleOverridesToCloudflare(overrides v1beta1.LoadBalancerRuleOverrides) *cloudflare.LoadBalancerRuleOverrides {
	cfOverrides := &cloudflare.LoadBalancerRuleOverrides{}

	if overrides.SessionAffinity != nil {
		cfOverrides.Persistence = *overrides.SessionAffinity
	}

	if overrides.SessionAffinityTTL != nil {
		ttl := uint(*overrides.SessionAffinityTTL)
		cfOverrides.PersistenceTTL = &ttl
	}

	if overrides.SessionAffinityAttributes != nil {
		cfOverrides.SessionAffinityAttrs = convertSessionAffinityAttributesForRuleOverrides(*overrides.SessionAffinityAttributes)
	}

	if overrides.TTL != nil {
		cfOverrides.TTL = uint(*overrides.TTL)
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
		cfOverrides.PoPPools = overrides.PopPools
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
func convertRandomSteeringToCloudflare(steering v1beta1.RandomSteering) *cloudflare.RandomSteering {
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
func convertAdaptiveRoutingToCloudflare(routing v1beta1.AdaptiveRouting) *cloudflare.AdaptiveRouting {
	cfRouting := &cloudflare.AdaptiveRouting{}

	if routing.FailoverAcrossPools != nil {
		cfRouting.FailoverAcrossPools = routing.FailoverAcrossPools
	}

	return cfRouting
}

// convertLocationStrategyToCloudflare converts location strategy to Cloudflare format
func convertLocationStrategyToCloudflare(strategy v1beta1.LocationStrategy) *cloudflare.LocationStrategy {
	cfStrategy := &cloudflare.LocationStrategy{}

	if strategy.Mode != nil {
		cfStrategy.Mode = *strategy.Mode
	}

	if strategy.PreferECSRegion != nil {
		cfStrategy.PreferECS = *strategy.PreferECSRegion
	}

	return cfStrategy
}

// GenerateLoadBalancerObservation creates observation from Cloudflare load balancer
func GenerateLoadBalancerObservation(lb *cloudflare.LoadBalancer) v1beta1.LoadBalancerObservation {
	observation := v1beta1.LoadBalancerObservation{
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
func IsLoadBalancerUpToDate(params *v1beta1.LoadBalancerParameters, lb *cloudflare.LoadBalancer) bool {
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

	if params.Enabled != nil && lb.Enabled != nil && *params.Enabled != *lb.Enabled {
		return false
	}

	if params.SessionAffinity != nil && *params.SessionAffinity != lb.Persistence {
		return false
	}

	if params.SessionAffinityTTL != nil && *params.SessionAffinityTTL != lb.PersistenceTTL {
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

// GenerateMonitorObservation creates observation from Cloudflare load balancer monitor
func GenerateMonitorObservation(monitor *cloudflare.LoadBalancerMonitor) v1beta1.LoadBalancerMonitorObservation {
	observation := v1beta1.LoadBalancerMonitorObservation{
		ID: monitor.ID,
	}

	if monitor.CreatedOn != nil {
		createdOn := monitor.CreatedOn.String()
		observation.CreatedOn = &createdOn
	}

	if monitor.ModifiedOn != nil {
		modifiedOn := monitor.ModifiedOn.String()
		observation.ModifiedOn = &modifiedOn
	}

	return observation
}

// GeneratePoolObservation creates observation from Cloudflare load balancer pool
func GeneratePoolObservation(pool *cloudflare.LoadBalancerPool) v1beta1.LoadBalancerPoolObservation {
	observation := v1beta1.LoadBalancerPoolObservation{
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

	// Note: Healthy status might not be directly available in the pool struct
	// This would need to be determined from origin health checks

	return observation
}

// IsMonitorUpToDate determines if the Cloudflare load balancer monitor is up to date
func IsMonitorUpToDate(params *v1beta1.LoadBalancerMonitorParameters, monitor *cloudflare.LoadBalancerMonitor) bool {
	if params.Description != nil && *params.Description != monitor.Description {
		return false
	}

	if params.Description == nil && monitor.Description != "" {
		return false
	}

	if params.Type != monitor.Type {
		return false
	}

	if params.Method != nil && *params.Method != monitor.Method {
		return false
	}

	if params.Path != nil && *params.Path != monitor.Path {
		return false
	}

	if params.Timeout != nil && *params.Timeout != monitor.Timeout {
		return false
	}

	if params.Retries != nil && *params.Retries != monitor.Retries {
		return false
	}

	if params.Interval != nil && *params.Interval != monitor.Interval {
		return false
	}

	if params.ConsecutiveUp != nil && *params.ConsecutiveUp != monitor.ConsecutiveUp {
		return false
	}

	if params.ConsecutiveDown != nil && *params.ConsecutiveDown != monitor.ConsecutiveDown {
		return false
	}

	if params.Port != nil && uint16(*params.Port) != monitor.Port {
		return false
	}

	if params.ExpectedBody != nil && *params.ExpectedBody != monitor.ExpectedBody {
		return false
	}

	if params.ExpectedCodes != nil && *params.ExpectedCodes != monitor.ExpectedCodes {
		return false
	}

	if params.FollowRedirects != nil && *params.FollowRedirects != monitor.FollowRedirects {
		return false
	}

	if params.AllowInsecure != nil && *params.AllowInsecure != monitor.AllowInsecure {
		return false
	}

	if params.ProbeZone != nil && *params.ProbeZone != monitor.ProbeZone {
		return false
	}

	return true
}

// CreateMonitor creates a new Cloudflare load balancer monitor
func (c *monitorClient) CreateMonitor(ctx context.Context, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	monitor := cloudflare.LoadBalancerMonitor{
		Type: params.Type,
	}

	if params.Description != nil {
		monitor.Description = *params.Description
	}

	if params.Method != nil {
		monitor.Method = *params.Method
	}

	if params.Path != nil {
		monitor.Path = *params.Path
	}

	if params.Header != nil {
		monitor.Header = params.Header
	}

	if params.Timeout != nil {
		monitor.Timeout = *params.Timeout
	}

	if params.Retries != nil {
		monitor.Retries = *params.Retries
	}

	if params.Interval != nil {
		monitor.Interval = *params.Interval
	}

	if params.ConsecutiveUp != nil {
		monitor.ConsecutiveUp = *params.ConsecutiveUp
	}

	if params.ConsecutiveDown != nil {
		monitor.ConsecutiveDown = *params.ConsecutiveDown
	}

	if params.Port != nil {
		monitor.Port = uint16(*params.Port)
	}

	if params.ExpectedBody != nil {
		monitor.ExpectedBody = *params.ExpectedBody
	}

	if params.ExpectedCodes != nil {
		monitor.ExpectedCodes = *params.ExpectedCodes
	}

	if params.FollowRedirects != nil {
		monitor.FollowRedirects = *params.FollowRedirects
	}

	if params.AllowInsecure != nil {
		monitor.AllowInsecure = *params.AllowInsecure
	}

	if params.ProbeZone != nil {
		monitor.ProbeZone = *params.ProbeZone
	}

	var rc *cloudflare.ResourceContainer
	if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else {
		return nil, errors.New("either Account or Zone must be specified for load balancer monitor")
	}

	createParams := cloudflare.CreateLoadBalancerMonitorParams{
		LoadBalancerMonitor: monitor,
	}

	result, err := c.api.CreateLoadBalancerMonitor(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateLoadBalancerMonitor)
	}

	return &result, nil
}

// GetMonitor retrieves a Cloudflare load balancer monitor
func (c *monitorClient) GetMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	var rc *cloudflare.ResourceContainer
	if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else {
		return nil, errors.New("either Account or Zone must be specified for load balancer monitor")
	}

	monitor, err := c.api.GetLoadBalancerMonitor(ctx, rc, monitorID)
	if err != nil {
		return nil, errors.Wrap(err, errGetLoadBalancerMonitor)
	}

	return &monitor, nil
}

// UpdateMonitor updates a Cloudflare load balancer monitor
func (c *monitorClient) UpdateMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	monitor := cloudflare.LoadBalancerMonitor{
		ID:   monitorID,
		Type: params.Type,
	}

	if params.Description != nil {
		monitor.Description = *params.Description
	}

	if params.Method != nil {
		monitor.Method = *params.Method
	}

	if params.Path != nil {
		monitor.Path = *params.Path
	}

	if params.Header != nil {
		monitor.Header = params.Header
	}

	if params.Timeout != nil {
		monitor.Timeout = *params.Timeout
	}

	if params.Retries != nil {
		monitor.Retries = *params.Retries
	}

	if params.Interval != nil {
		monitor.Interval = *params.Interval
	}

	if params.ConsecutiveUp != nil {
		monitor.ConsecutiveUp = *params.ConsecutiveUp
	}

	if params.ConsecutiveDown != nil {
		monitor.ConsecutiveDown = *params.ConsecutiveDown
	}

	if params.Port != nil {
		monitor.Port = uint16(*params.Port)
	}

	if params.ExpectedBody != nil {
		monitor.ExpectedBody = *params.ExpectedBody
	}

	if params.ExpectedCodes != nil {
		monitor.ExpectedCodes = *params.ExpectedCodes
	}

	if params.FollowRedirects != nil {
		monitor.FollowRedirects = *params.FollowRedirects
	}

	if params.AllowInsecure != nil {
		monitor.AllowInsecure = *params.AllowInsecure
	}

	if params.ProbeZone != nil {
		monitor.ProbeZone = *params.ProbeZone
	}

	var rc *cloudflare.ResourceContainer
	if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else {
		return nil, errors.New("either Account or Zone must be specified for load balancer monitor")
	}

	updateParams := cloudflare.UpdateLoadBalancerMonitorParams{
		LoadBalancerMonitor: monitor,
	}

	result, err := c.api.UpdateLoadBalancerMonitor(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateLoadBalancerMonitor)
	}

	return &result, nil
}

// DeleteMonitor deletes a Cloudflare load balancer monitor
func (c *monitorClient) DeleteMonitor(ctx context.Context, monitorID string, params v1beta1.LoadBalancerMonitorParameters) error {
	var rc *cloudflare.ResourceContainer
	if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else {
		return errors.New("either Account or Zone must be specified for load balancer monitor")
	}

	err := c.api.DeleteLoadBalancerMonitor(ctx, rc, monitorID)
	if err != nil {
		return errors.Wrap(err, errDeleteLoadBalancerMonitor)
	}

	return nil
}

// IsMonitorNotFound checks if error indicates monitor not found
func IsMonitorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// CreatePool creates a new Cloudflare load balancer pool
func (c *poolClient) CreatePool(ctx context.Context, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	pool := cloudflare.LoadBalancerPool{}

	if params.Name != nil {
		pool.Name = *params.Name
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

	if len(params.Origins) > 0 {
		pool.Origins = convertOriginsToCloudflare(params.Origins)
	}

	// Note: OriginSteering may not be supported in the current cloudflare-go version
	// if params.OriginSteering != nil {
	//     pool.OriginSteering = convertOriginSteeringToCloudflare(*params.OriginSteering)
	// }

	if len(params.CheckRegions) > 0 {
		pool.CheckRegions = params.CheckRegions
	}

	if params.Latitude != nil {
		lat := float32(*params.Latitude)
		pool.Latitude = &lat
	}

	if params.Longitude != nil {
		lng := float32(*params.Longitude)
		pool.Longitude = &lng
	}

	// Pools are account-level resources
	rc := cloudflare.AccountIdentifier("") // Empty string means use the account from credentials

	createParams := cloudflare.CreateLoadBalancerPoolParams{
		LoadBalancerPool: pool,
	}

	result, err := c.api.CreateLoadBalancerPool(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateLoadBalancerPool)
	}

	return &result, nil
}

// GetPool retrieves a Cloudflare load balancer pool
func (c *poolClient) GetPool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	// Pools are account-level resources
	rc := cloudflare.AccountIdentifier("") // Empty string means use the account from credentials

	pool, err := c.api.GetLoadBalancerPool(ctx, rc, poolID)
	if err != nil {
		return nil, errors.Wrap(err, errGetLoadBalancerPool)
	}

	return &pool, nil
}

// UpdatePool updates a Cloudflare load balancer pool
func (c *poolClient) UpdatePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
	pool := cloudflare.LoadBalancerPool{
		ID: poolID,
	}

	if params.Name != nil {
		pool.Name = *params.Name
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

	if len(params.Origins) > 0 {
		pool.Origins = convertOriginsToCloudflare(params.Origins)
	}

	// Note: OriginSteering may not be supported in the current cloudflare-go version
	// if params.OriginSteering != nil {
	//     pool.OriginSteering = convertOriginSteeringToCloudflare(*params.OriginSteering)
	// }

	if len(params.CheckRegions) > 0 {
		pool.CheckRegions = params.CheckRegions
	}

	if params.Latitude != nil {
		lat := float32(*params.Latitude)
		pool.Latitude = &lat
	}

	if params.Longitude != nil {
		lng := float32(*params.Longitude)
		pool.Longitude = &lng
	}

	// Pools are account-level resources
	rc := cloudflare.AccountIdentifier("") // Empty string means use the account from credentials

	updateParams := cloudflare.UpdateLoadBalancerPoolParams{
		LoadBalancer: pool,
	}

	result, err := c.api.UpdateLoadBalancerPool(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateLoadBalancerPool)
	}

	return &result, nil
}

// DeletePool deletes a Cloudflare load balancer pool
func (c *poolClient) DeletePool(ctx context.Context, poolID string, params v1beta1.LoadBalancerPoolParameters) error {
	// Pools are account-level resources
	rc := cloudflare.AccountIdentifier("") // Empty string means use the account from credentials

	err := c.api.DeleteLoadBalancerPool(ctx, rc, poolID)
	if err != nil {
		return errors.Wrap(err, errDeleteLoadBalancerPool)
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

// IsPoolUpToDate determines if the Cloudflare load balancer pool is up to date
func IsPoolUpToDate(params *v1beta1.LoadBalancerPoolParameters, pool *cloudflare.LoadBalancerPool) bool {
	if params.Name != nil && *params.Name != pool.Name {
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

	if params.MinimumOrigins != nil && pool.MinimumOrigins != nil && *params.MinimumOrigins != *pool.MinimumOrigins {
		return false
	}

	if params.Monitor != nil && *params.Monitor != pool.Monitor {
		return false
	}

	if params.NotificationEmail != nil && *params.NotificationEmail != pool.NotificationEmail {
		return false
	}

	if params.NotificationEmail == nil && pool.NotificationEmail != "" {
		return false
	}

	// For origins and other complex fields, we'll do basic length checks
	if len(params.Origins) != len(pool.Origins) {
		return false
	}

	return true
}

// convertOriginsToCloudflare converts origins to Cloudflare format
func convertOriginsToCloudflare(origins []v1beta1.LoadBalancerOrigin) []cloudflare.LoadBalancerOrigin {
	var cfOrigins []cloudflare.LoadBalancerOrigin

	for _, origin := range origins {
		cfOrigin := cloudflare.LoadBalancerOrigin{
			Name:    origin.Name,
			Address: origin.Address,
		}

		if origin.Enabled != nil {
			cfOrigin.Enabled = *origin.Enabled
		}

		if origin.Weight != nil {
			cfOrigin.Weight = *origin.Weight
		}

		if origin.Header != nil {
			cfOrigin.Header = origin.Header
		}

		cfOrigins = append(cfOrigins, cfOrigin)
	}

	return cfOrigins
}

// convertOriginSteeringToCloudflare converts origin steering to Cloudflare format
// Note: Commented out as OriginSteering may not be supported in current cloudflare-go version
// func convertOriginSteeringToCloudflare(steering v1beta1.OriginSteering) *cloudflare.OriginSteering {
//     cfSteering := &cloudflare.OriginSteering{}
//
//     if steering.Policy != nil {
//         cfSteering.Policy = *steering.Policy
//     }
//
//     return cfSteering
// }
