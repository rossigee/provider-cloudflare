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

package ratelimit

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/security/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// RateLimitAPI defines the interface for Rate Limit operations
type RateLimitAPI interface {
	RateLimit(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error)
	CreateRateLimit(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	UpdateRateLimit(ctx context.Context, zoneID, limitID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	DeleteRateLimit(ctx context.Context, zoneID, limitID string) error
}

// CloudflareRateLimitClient is a Cloudflare API client for Rate Limits.
type CloudflareRateLimitClient struct {
	client RateLimitAPI
}

// NewClient creates a new CloudflareRateLimitClient.
func NewClient(client RateLimitAPI) *CloudflareRateLimitClient {
	return &CloudflareRateLimitClient{client: client}
}

// NewClientFromAPI creates a new CloudflareRateLimitClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareRateLimitClient {
	return NewClient(api)
}

// Get retrieves a Rate Limit.
func (c *CloudflareRateLimitClient) Get(ctx context.Context, zoneID, rateLimitID string) (*v1alpha1.RateLimitObservation, error) {
	rateLimit, err := c.client.RateLimit(ctx, zoneID, rateLimitID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("rate limit not found")
		}
		return nil, errors.Wrap(err, "cannot get rate limit")
	}

	return convertRateLimitToObservation(rateLimit), nil
}

// Create creates a new Rate Limit.
func (c *CloudflareRateLimitClient) Create(ctx context.Context, params v1alpha1.RateLimitParameters) (*v1alpha1.RateLimitObservation, error) {
	createRateLimit := convertParametersToRateLimit(params)
	
	rateLimit, err := c.client.CreateRateLimit(ctx, params.Zone, createRateLimit)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create rate limit")
	}

	return convertRateLimitToObservation(rateLimit), nil
}

// Update updates a Rate Limit.
func (c *CloudflareRateLimitClient) Update(ctx context.Context, rateLimitID string, params v1alpha1.RateLimitParameters) (*v1alpha1.RateLimitObservation, error) {
	updateRateLimit := convertParametersToRateLimit(params)
	updateRateLimit.ID = rateLimitID
	
	rateLimit, err := c.client.UpdateRateLimit(ctx, params.Zone, rateLimitID, updateRateLimit)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update rate limit")
	}

	return convertRateLimitToObservation(rateLimit), nil
}

// Delete deletes a Rate Limit.
func (c *CloudflareRateLimitClient) Delete(ctx context.Context, zoneID, rateLimitID string) error {
	err := c.client.DeleteRateLimit(ctx, zoneID, rateLimitID)
	if err != nil && !isNotFound(err) {
		return errors.Wrap(err, "cannot delete rate limit")
	}
	return nil
}

// IsUpToDate checks if the Rate Limit is up to date.
func (c *CloudflareRateLimitClient) IsUpToDate(ctx context.Context, params v1alpha1.RateLimitParameters, obs v1alpha1.RateLimitObservation) (bool, error) {
	// Compare key parameters
	if params.Disabled != nil && *params.Disabled != obs.Disabled {
		return false, nil
	}
	
	if params.Description != nil && *params.Description != obs.Description {
		return false, nil
	}
	
	if params.Threshold != obs.Threshold {
		return false, nil
	}
	
	if params.Period != obs.Period {
		return false, nil
	}
	
	if params.Action.Mode != obs.Action.Mode {
		return false, nil
	}
	
	if params.Action.Timeout != nil && obs.Action.Timeout != nil && *params.Action.Timeout != *obs.Action.Timeout {
		return false, nil
	}
	
	// Compare match rules (simplified comparison)
	if len(params.Match.Request.Methods) != len(obs.Match.Request.Methods) {
		return false, nil
	}
	
	return true, nil
}

// convertParametersToRateLimit converts RateLimitParameters to cloudflare.RateLimit.
func convertParametersToRateLimit(params v1alpha1.RateLimitParameters) cloudflare.RateLimit {
	rateLimit := cloudflare.RateLimit{
		Threshold: params.Threshold,
		Period:    params.Period,
		Match:     convertTrafficMatcher(params.Match),
		Action:    convertAction(params.Action),
	}
	
	if params.Disabled != nil {
		rateLimit.Disabled = *params.Disabled
	}
	
	if params.Description != nil {
		rateLimit.Description = *params.Description
	}
	
	if params.Bypass != nil {
		rateLimit.Bypass = convertBypass(params.Bypass)
	}
	
	if params.Correlate != nil {
		rateLimit.Correlate = convertCorrelate(params.Correlate)
	}
	
	return rateLimit
}

// convertRateLimitToObservation converts cloudflare.RateLimit to RateLimitObservation.
func convertRateLimitToObservation(rateLimit cloudflare.RateLimit) *v1alpha1.RateLimitObservation {
	obs := &v1alpha1.RateLimitObservation{
		ID:          rateLimit.ID,
		Disabled:    rateLimit.Disabled,
		Description: rateLimit.Description,
		Threshold:   rateLimit.Threshold,
		Period:      rateLimit.Period,
		Match:       convertTrafficMatcherFromCloudflare(rateLimit.Match),
		Action:      convertActionFromCloudflare(rateLimit.Action),
	}
	
	if rateLimit.Bypass != nil {
		obs.Bypass = convertBypassFromCloudflare(rateLimit.Bypass)
	}
	
	if rateLimit.Correlate != nil {
		obs.Correlate = convertCorrelateFromCloudflare(rateLimit.Correlate)
	}
	
	return obs
}

// convertTrafficMatcher converts v1alpha1.RateLimitTrafficMatcher to cloudflare.RateLimitTrafficMatcher.
func convertTrafficMatcher(matcher v1alpha1.RateLimitTrafficMatcher) cloudflare.RateLimitTrafficMatcher {
	cfMatcher := cloudflare.RateLimitTrafficMatcher{
		Request: cloudflare.RateLimitRequestMatcher{
			Methods: matcher.Request.Methods,
			Schemes: matcher.Request.Schemes,
		},
	}
	
	if matcher.Request.URLPattern != nil {
		cfMatcher.Request.URLPattern = *matcher.Request.URLPattern
	}
	
	if matcher.Response != nil {
		cfMatcher.Response = cloudflare.RateLimitResponseMatcher{
			Statuses:      matcher.Response.Statuses,
			OriginTraffic: matcher.Response.OriginTraffic,
		}
		
		if matcher.Response.Headers != nil {
			cfMatcher.Response.Headers = make([]cloudflare.RateLimitResponseMatcherHeader, len(matcher.Response.Headers))
			for i, header := range matcher.Response.Headers {
				cfMatcher.Response.Headers[i] = cloudflare.RateLimitResponseMatcherHeader{
					Name:  header.Name,
					Op:    header.Op,
					Value: header.Value,
				}
			}
		}
	}
	
	return cfMatcher
}

// convertTrafficMatcherFromCloudflare converts cloudflare.RateLimitTrafficMatcher to v1alpha1.RateLimitTrafficMatcher.
func convertTrafficMatcherFromCloudflare(matcher cloudflare.RateLimitTrafficMatcher) v1alpha1.RateLimitTrafficMatcher {
	v1Matcher := v1alpha1.RateLimitTrafficMatcher{
		Request: v1alpha1.RateLimitRequestMatcher{
			Methods: matcher.Request.Methods,
			Schemes: matcher.Request.Schemes,
		},
	}
	
	if matcher.Request.URLPattern != "" {
		v1Matcher.Request.URLPattern = &matcher.Request.URLPattern
	}
	
	// Check if response matcher has any data
	if len(matcher.Response.Statuses) > 0 || matcher.Response.OriginTraffic != nil || len(matcher.Response.Headers) > 0 {
		v1Matcher.Response = &v1alpha1.RateLimitResponseMatcher{
			Statuses:      matcher.Response.Statuses,
			OriginTraffic: matcher.Response.OriginTraffic,
		}
		
		if len(matcher.Response.Headers) > 0 {
			v1Matcher.Response.Headers = make([]v1alpha1.RateLimitResponseMatcherHeader, len(matcher.Response.Headers))
			for i, header := range matcher.Response.Headers {
				v1Matcher.Response.Headers[i] = v1alpha1.RateLimitResponseMatcherHeader{
					Name:  header.Name,
					Op:    header.Op,
					Value: header.Value,
				}
			}
		}
	}
	
	return v1Matcher
}

// convertAction converts v1alpha1.RateLimitAction to cloudflare.RateLimitAction.
func convertAction(action v1alpha1.RateLimitAction) cloudflare.RateLimitAction {
	cfAction := cloudflare.RateLimitAction{
		Mode: action.Mode,
	}
	
	if action.Timeout != nil {
		cfAction.Timeout = *action.Timeout
	}
	
	if action.Response != nil {
		cfAction.Response = &cloudflare.RateLimitActionResponse{
			ContentType: action.Response.ContentType,
			Body:        action.Response.Body,
		}
	}
	
	return cfAction
}

// convertActionFromCloudflare converts cloudflare.RateLimitAction to v1alpha1.RateLimitAction.
func convertActionFromCloudflare(action cloudflare.RateLimitAction) v1alpha1.RateLimitAction {
	v1Action := v1alpha1.RateLimitAction{
		Mode: action.Mode,
	}
	
	if action.Timeout != 0 {
		v1Action.Timeout = &action.Timeout
	}
	
	if action.Response != nil {
		v1Action.Response = &v1alpha1.RateLimitActionResponse{
			ContentType: action.Response.ContentType,
			Body:        action.Response.Body,
		}
	}
	
	return v1Action
}

// convertBypass converts []v1alpha1.RateLimitKeyValue to []cloudflare.RateLimitKeyValue.
func convertBypass(bypass []v1alpha1.RateLimitKeyValue) []cloudflare.RateLimitKeyValue {
	cfBypass := make([]cloudflare.RateLimitKeyValue, len(bypass))
	for i, kv := range bypass {
		cfBypass[i] = cloudflare.RateLimitKeyValue{
			Name:  kv.Name,
			Value: kv.Value,
		}
	}
	return cfBypass
}

// convertBypassFromCloudflare converts []cloudflare.RateLimitKeyValue to []v1alpha1.RateLimitKeyValue.
func convertBypassFromCloudflare(bypass []cloudflare.RateLimitKeyValue) []v1alpha1.RateLimitKeyValue {
	v1Bypass := make([]v1alpha1.RateLimitKeyValue, len(bypass))
	for i, kv := range bypass {
		v1Bypass[i] = v1alpha1.RateLimitKeyValue{
			Name:  kv.Name,
			Value: kv.Value,
		}
	}
	return v1Bypass
}

// convertCorrelate converts *v1alpha1.RateLimitCorrelate to *cloudflare.RateLimitCorrelate.
func convertCorrelate(correlate *v1alpha1.RateLimitCorrelate) *cloudflare.RateLimitCorrelate {
	if correlate == nil {
		return nil
	}
	return &cloudflare.RateLimitCorrelate{
		By: correlate.By,
	}
}

// convertCorrelateFromCloudflare converts *cloudflare.RateLimitCorrelate to *v1alpha1.RateLimitCorrelate.
func convertCorrelateFromCloudflare(correlate *cloudflare.RateLimitCorrelate) *v1alpha1.RateLimitCorrelate {
	if correlate == nil {
		return nil
	}
	return &v1alpha1.RateLimitCorrelate{
		By: correlate.By,
	}
}

// isNotFound checks if an error indicates that the rate limit was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "rate limit not found")
}