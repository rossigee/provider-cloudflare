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

package cache

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/cache/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateCacheRule = "failed to create cache rule"
	errGetCacheRule    = "failed to get cache rule"
	errUpdateCacheRule = "failed to update cache rule"
	errDeleteCacheRule = "failed to delete cache rule"
	errListRulesets    = "failed to list rulesets"
	errCreateRuleset   = "failed to create cache rule ruleset"
	errUpdateRuleset   = "failed to update cache rule ruleset"
	errDeleteRuleset   = "failed to delete cache rule ruleset"

	cacheRulesetPhase = "http_request_cache_settings"
	cacheRulesetKind  = "zone"
	cacheAction       = "set_cache_settings"
)

// CacheRuleClient interface for Cloudflare Cache Rule operations
type CacheRuleClient interface {
	CreateCacheRule(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	GetCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	UpdateCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	DeleteCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error
}

// NewCacheRuleClient creates a new Cloudflare Cache Rule client
func NewCacheRuleClient(cfg clients.Config, hc *http.Client) (CacheRuleClient, error) {
	api, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}
	return &cacheRuleClient{api: api}, nil
}

type cacheRuleClient struct {
	api *cloudflare.API
}

// CreateCacheRule creates a new cache rule in Cloudflare
func (c *cacheRuleClient) CreateCacheRule(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	// First, find or create the cache rules ruleset
	ruleset, err := c.findOrCreateCacheRuleset(ctx, rc, params)
	if err != nil {
		return nil, nil, errors.Wrap(err, errCreateCacheRule)
	}

	// Create the cache rule
	rule := convertCacheRuleParametersToCloudflare(params)

	updateParams := cloudflare.UpdateRulesetParams{
		ID:    ruleset.ID,
		Rules: append(ruleset.Rules, rule),
	}

	updatedRuleset, err := c.api.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return nil, nil, errors.Wrap(err, errCreateCacheRule)
	}

	// Find the newly created rule (it will be the last one)
	if len(updatedRuleset.Rules) == 0 {
		return nil, nil, errors.New("no rules found in updated ruleset")
	}

	newRule := updatedRuleset.Rules[len(updatedRuleset.Rules)-1]
	return &newRule, &updatedRuleset, nil
}

// GetCacheRule retrieves a cache rule from Cloudflare
func (c *cacheRuleClient) GetCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	ruleset, err := c.api.GetRuleset(ctx, rc, rulesetID)
	if err != nil {
		return nil, nil, errors.Wrap(err, errGetCacheRule)
	}

	// Find the specific rule
	for _, rule := range ruleset.Rules {
		if rule.ID == ruleID {
			return &rule, &ruleset, nil
		}
	}

	return nil, nil, fmt.Errorf("cache rule %s not found in ruleset %s", ruleID, rulesetID)
}

// UpdateCacheRule updates an existing cache rule in Cloudflare
func (c *cacheRuleClient) UpdateCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	// Get the current ruleset
	ruleset, err := c.api.GetRuleset(ctx, rc, rulesetID)
	if err != nil {
		return nil, nil, errors.Wrap(err, errUpdateCacheRule)
	}

	// Find and update the specific rule
	var updatedRule *cloudflare.RulesetRule
	for i, rule := range ruleset.Rules {
		if rule.ID == ruleID {
			newRule := convertCacheRuleParametersToCloudflare(params)
			newRule.ID = ruleID
			ruleset.Rules[i] = newRule
			updatedRule = &newRule
			break
		}
	}

	if updatedRule == nil {
		return nil, nil, fmt.Errorf("cache rule %s not found in ruleset %s", ruleID, rulesetID)
	}

	// Update the ruleset
	updateParams := cloudflare.UpdateRulesetParams{
		ID:    rulesetID,
		Rules: ruleset.Rules,
	}

	updatedRuleset, err := c.api.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return nil, nil, errors.Wrap(err, errUpdateCacheRule)
	}

	// Find the updated rule in the response
	for _, rule := range updatedRuleset.Rules {
		if rule.ID == ruleID {
			return &rule, &updatedRuleset, nil
		}
	}

	return updatedRule, &updatedRuleset, nil
}

// DeleteCacheRule deletes a cache rule from Cloudflare
func (c *cacheRuleClient) DeleteCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error {
	rc := cloudflare.ZoneIdentifier(params.Zone)

	// Get the current ruleset
	ruleset, err := c.api.GetRuleset(ctx, rc, rulesetID)
	if err != nil {
		return errors.Wrap(err, errDeleteCacheRule)
	}

	// Remove the specific rule
	var newRules []cloudflare.RulesetRule
	found := false
	for _, rule := range ruleset.Rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("cache rule %s not found in ruleset %s", ruleID, rulesetID)
	}

	// If this was the last rule, delete the entire ruleset
	if len(newRules) == 0 {
		err = c.api.DeleteRuleset(ctx, rc, rulesetID)
		if err != nil {
			return errors.Wrap(err, errDeleteRuleset)
		}
		return nil
	}

	// Update the ruleset without the deleted rule
	updateParams := cloudflare.UpdateRulesetParams{
		ID:    rulesetID,
		Rules: newRules,
	}

	_, err = c.api.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return errors.Wrap(err, errDeleteCacheRule)
	}

	return nil
}

// findOrCreateCacheRuleset finds an existing cache rules ruleset or creates a new one
func (c *cacheRuleClient) findOrCreateCacheRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params v1alpha1.CacheRuleParameters) (*cloudflare.Ruleset, error) {
	// List existing rulesets to find the cache rules ruleset
	rulesets, err := c.api.ListRulesets(ctx, rc, cloudflare.ListRulesetsParams{})
	if err != nil {
		return nil, errors.Wrap(err, errListRulesets)
	}

	// Look for an existing cache rules ruleset
	for _, ruleset := range rulesets {
		if ruleset.Phase == cacheRulesetPhase && ruleset.Kind == cacheRulesetKind {
			return &ruleset, nil
		}
	}

	// Create a new cache rules ruleset
	createParams := cloudflare.CreateRulesetParams{
		Name:        "Cache Rules",
		Description: "Cloudflare Cache Rules",
		Kind:        cacheRulesetKind,
		Phase:       cacheRulesetPhase,
		Rules:       []cloudflare.RulesetRule{},
	}

	ruleset, err := c.api.CreateRuleset(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateRuleset)
	}

	return &ruleset, nil
}

// convertCacheRuleParametersToCloudflare converts cache rule parameters to Cloudflare format
func convertCacheRuleParametersToCloudflare(params v1alpha1.CacheRuleParameters) cloudflare.RulesetRule {
	rule := cloudflare.RulesetRule{
		Action:     cacheAction,
		Expression: params.Expression,
	}

	if params.Description != nil {
		rule.Description = *params.Description
	}

	if params.Enabled != nil {
		rule.Enabled = params.Enabled
	}

	// Convert action parameters
	actionParams := &cloudflare.RulesetRuleActionParameters{}

	if params.Cache != nil {
		actionParams.Cache = params.Cache
	}

	if params.EdgeTTL != nil {
		actionParams.EdgeTTL = convertEdgeTTLToCloudflare(*params.EdgeTTL)
	}

	if params.BrowserTTL != nil {
		actionParams.BrowserTTL = convertBrowserTTLToCloudflare(*params.BrowserTTL)
	}

	if params.ServeStale != nil {
		actionParams.ServeStale = convertServeStaleToCloudflare(*params.ServeStale)
	}

	if params.CacheKey != nil {
		actionParams.CacheKey = convertCacheKeyToCloudflare(*params.CacheKey)
	}

	if params.CacheReserve != nil {
		actionParams.CacheReserve = convertCacheReserveToCloudflare(*params.CacheReserve)
	}

	if params.OriginCacheControl != nil {
		actionParams.OriginCacheControl = params.OriginCacheControl
	}

	if params.RespectStrongETags != nil {
		actionParams.RespectStrongETags = params.RespectStrongETags
	}

	if params.OriginErrorPagePassthru != nil {
		actionParams.OriginErrorPagePassthru = params.OriginErrorPagePassthru
	}

	if len(params.AdditionalCacheablePorts) > 0 {
		actionParams.AdditionalCacheablePorts = params.AdditionalCacheablePorts
	}

	if params.ReadTimeout != nil {
		readTimeout := uint(*params.ReadTimeout)
		actionParams.ReadTimeout = &readTimeout
	}

	rule.ActionParameters = actionParams

	return rule
}

// Helper conversion functions
func convertEdgeTTLToCloudflare(edgeTTL v1alpha1.EdgeTTL) *cloudflare.RulesetRuleActionParametersEdgeTTL {
	cfEdgeTTL := &cloudflare.RulesetRuleActionParametersEdgeTTL{
		Mode: edgeTTL.Mode,
	}

	if edgeTTL.Default != nil {
		defaultTTL := uint(*edgeTTL.Default)
		cfEdgeTTL.Default = &defaultTTL
	}

	if len(edgeTTL.StatusCodeTTL) > 0 {
		for _, statusTTL := range edgeTTL.StatusCodeTTL {
			cfStatusTTL := cloudflare.RulesetRuleActionParametersStatusCodeTTL{
				Value: &statusTTL.Value,
			}

			if statusTTL.StatusCodeValue != nil {
				statusCodeValue := uint(*statusTTL.StatusCodeValue)
				cfStatusTTL.StatusCodeValue = &statusCodeValue
			}

			if statusTTL.StatusCodeRange != nil {
				fromCode := uint(statusTTL.StatusCodeRange.From)
				toCode := uint(statusTTL.StatusCodeRange.To)
				cfStatusTTL.StatusCodeRange = &cloudflare.RulesetRuleActionParametersStatusCodeRange{
					From: &fromCode,
					To:   &toCode,
				}
			}

			cfEdgeTTL.StatusCodeTTL = append(cfEdgeTTL.StatusCodeTTL, cfStatusTTL)
		}
	}

	return cfEdgeTTL
}

func convertBrowserTTLToCloudflare(browserTTL v1alpha1.BrowserTTL) *cloudflare.RulesetRuleActionParametersBrowserTTL {
	cfBrowserTTL := &cloudflare.RulesetRuleActionParametersBrowserTTL{
		Mode: browserTTL.Mode,
	}

	if browserTTL.Default != nil {
		defaultTTL := uint(*browserTTL.Default)
		cfBrowserTTL.Default = &defaultTTL
	}

	return cfBrowserTTL
}

func convertServeStaleToCloudflare(serveStale v1alpha1.ServeStale) *cloudflare.RulesetRuleActionParametersServeStale {
	return &cloudflare.RulesetRuleActionParametersServeStale{
		DisableStaleWhileUpdating: serveStale.DisableStaleWhileUpdating,
	}
}

func convertCacheKeyToCloudflare(cacheKey v1alpha1.CacheKey) *cloudflare.RulesetRuleActionParametersCacheKey {
	cfCacheKey := &cloudflare.RulesetRuleActionParametersCacheKey{
		CacheByDeviceType:       cacheKey.CacheByDeviceType,
		IgnoreQueryStringsOrder: cacheKey.IgnoreQueryStringsOrder,
		CacheDeceptionArmor:     cacheKey.CacheDeceptionArmor,
	}

	if cacheKey.CustomKey != nil {
		cfCacheKey.CustomKey = convertCustomKeyToCloudflare(*cacheKey.CustomKey)
	}

	return cfCacheKey
}

func convertCustomKeyToCloudflare(customKey v1alpha1.CustomKey) *cloudflare.RulesetRuleActionParametersCustomKey {
	cfCustomKey := &cloudflare.RulesetRuleActionParametersCustomKey{}

	if customKey.Query != nil {
		cfCustomKey.Query = convertCustomKeyQueryToCloudflare(*customKey.Query)
	}

	if customKey.Header != nil {
		cfCustomKey.Header = convertCustomKeyHeaderToCloudflare(*customKey.Header)
	}

	if customKey.Cookie != nil {
		cfCustomKey.Cookie = convertCustomKeyFieldsToCloudflare(*customKey.Cookie)
	}

	if customKey.User != nil {
		cfCustomKey.User = convertCustomKeyUserToCloudflare(*customKey.User)
	}

	if customKey.Host != nil {
		cfCustomKey.Host = convertCustomKeyHostToCloudflare(*customKey.Host)
	}

	return cfCustomKey
}

func convertCustomKeyQueryToCloudflare(query v1alpha1.CustomKeyQuery) *cloudflare.RulesetRuleActionParametersCustomKeyQuery {
	cfQuery := &cloudflare.RulesetRuleActionParametersCustomKeyQuery{
		Ignore: query.Ignore,
	}

	if len(query.Include) > 0 || (query.All != nil && *query.All) {
		cfQuery.Include = &cloudflare.RulesetRuleActionParametersCustomKeyList{}
		if query.All != nil && *query.All {
			cfQuery.Include.All = true
		} else {
			cfQuery.Include.List = query.Include
		}
	}

	if len(query.Exclude) > 0 {
		cfQuery.Exclude = &cloudflare.RulesetRuleActionParametersCustomKeyList{
			List: query.Exclude,
		}
	}

	return cfQuery
}

func convertCustomKeyHeaderToCloudflare(header v1alpha1.CustomKeyHeader) *cloudflare.RulesetRuleActionParametersCustomKeyHeader {
	return &cloudflare.RulesetRuleActionParametersCustomKeyHeader{
		RulesetRuleActionParametersCustomKeyFields: cloudflare.RulesetRuleActionParametersCustomKeyFields{
			Include:       header.Include,
			CheckPresence: header.CheckPresence,
		},
		ExcludeOrigin: header.ExcludeOrigin,
		Contains:      header.Contains,
	}
}

func convertCustomKeyFieldsToCloudflare(fields v1alpha1.CustomKeyFields) *cloudflare.RulesetRuleActionParametersCustomKeyCookie {
	return &cloudflare.RulesetRuleActionParametersCustomKeyCookie{
		Include:       fields.Include,
		CheckPresence: fields.CheckPresence,
	}
}

func convertCustomKeyUserToCloudflare(user v1alpha1.CustomKeyUser) *cloudflare.RulesetRuleActionParametersCustomKeyUser {
	return &cloudflare.RulesetRuleActionParametersCustomKeyUser{
		DeviceType: user.DeviceType,
		Geo:        user.Geo,
		Lang:       user.Lang,
	}
}

func convertCustomKeyHostToCloudflare(host v1alpha1.CustomKeyHost) *cloudflare.RulesetRuleActionParametersCustomKeyHost {
	return &cloudflare.RulesetRuleActionParametersCustomKeyHost{
		Resolved: host.Resolved,
	}
}

func convertCacheReserveToCloudflare(cacheReserve v1alpha1.CacheReserve) *cloudflare.RulesetRuleActionParametersCacheReserve {
	cfCacheReserve := &cloudflare.RulesetRuleActionParametersCacheReserve{
		Eligible: cacheReserve.Eligible,
	}

	if cacheReserve.MinimumFileSize != nil {
		minFileSize := uint(*cacheReserve.MinimumFileSize)
		cfCacheReserve.MinimumFileSize = &minFileSize
	}

	return cfCacheReserve
}

// IsCacheRuleNotFound checks if error indicates cache rule not found
func IsCacheRuleNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// GenerateCacheRuleObservation creates observation from Cloudflare cache rule
func GenerateCacheRuleObservation(rule *cloudflare.RulesetRule, ruleset *cloudflare.Ruleset) v1alpha1.CacheRuleObservation {
	observation := v1alpha1.CacheRuleObservation{
		ID:        rule.ID,
		RulesetID: ruleset.ID,
	}

	if rule.Version != nil {
		observation.Version = *rule.Version
	}

	if rule.LastUpdated != nil {
		lastUpdated := rule.LastUpdated.String()
		observation.LastUpdated = &lastUpdated
	}

	if ruleset.LastUpdated != nil {
		modifiedOn := ruleset.LastUpdated.String()
		observation.ModifiedOn = &modifiedOn
	}

	return observation
}

// IsCacheRuleUpToDate determines if the cache rule is up to date
func IsCacheRuleUpToDate(params *v1alpha1.CacheRuleParameters, rule *cloudflare.RulesetRule) bool {
	// Check basic fields
	if params.Expression != rule.Expression {
		return false
	}

	if params.Description != nil && *params.Description != rule.Description {
		return false
	}

	if params.Description == nil && rule.Description != "" {
		return false
	}

	if params.Enabled != nil && rule.Enabled != nil && *params.Enabled != *rule.Enabled {
		return false
	}

	// For a more sophisticated comparison, we would need to compare action parameters
	// This is a simplified check focusing on the most common fields
	return true
}