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

package rule

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/emailrouting/v1alpha1"
)

// EmailRoutingRuleAPI defines the interface for Email Routing Rule operations
type EmailRoutingRuleAPI interface {
	CreateEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error)
	GetEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error)
	UpdateEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEmailRoutingRuleParameters) (cloudflare.EmailRoutingRule, error)
	DeleteEmailRoutingRule(ctx context.Context, rc *cloudflare.ResourceContainer, ruleTag string) (cloudflare.EmailRoutingRule, error)
	ListEmailRoutingRules(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListEmailRoutingRulesParameters) ([]cloudflare.EmailRoutingRule, *cloudflare.ResultInfo, error)
}

const (
	errCreateRule = "cannot create email routing rule"
	errUpdateRule = "cannot update email routing rule"
	errGetRule    = "cannot get email routing rule"
	errDeleteRule = "cannot delete email routing rule"
	errListRules  = "cannot list email routing rules"
)

// RuleClient provides operations for Email Routing Rules.
type RuleClient struct {
	client EmailRoutingRuleAPI
}

// NewClient creates a new Email Routing Rule client.
func NewClient(client EmailRoutingRuleAPI) *RuleClient {
	return &RuleClient{
		client: client,
	}
}

// NewClientFromAPI creates a new Email Routing Rule client from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *RuleClient {
	return NewClient(api)
}

// convertToObservation converts cloudflare-go email routing rule to Crossplane observation.
func convertToObservation(rule cloudflare.EmailRoutingRule) v1alpha1.RuleObservation {
	obs := v1alpha1.RuleObservation{
		Tag:      rule.Tag,
		Name:     rule.Name,
		Priority: &rule.Priority,
		Enabled:  rule.Enabled,
	}

	// Convert matchers
	if len(rule.Matchers) > 0 {
		obs.Matchers = make([]v1alpha1.RuleMatcher, len(rule.Matchers))
		for i, matcher := range rule.Matchers {
			obs.Matchers[i] = v1alpha1.RuleMatcher{
				Type:  matcher.Type,
				Field: matcher.Field,
				Value: matcher.Value,
			}
		}
	}

	// Convert actions
	if len(rule.Actions) > 0 {
		obs.Actions = make([]v1alpha1.RuleAction, len(rule.Actions))
		for i, action := range rule.Actions {
			obs.Actions[i] = v1alpha1.RuleAction{
				Type:  action.Type,
				Value: action.Value,
			}
		}
	}

	return obs
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.RuleParameters) cloudflare.CreateEmailRoutingRuleParameters {
	cfParams := cloudflare.CreateEmailRoutingRuleParameters{
		Name:     params.Name,
		Priority: params.Priority,
		Enabled:  params.Enabled,
	}

	// Convert matchers
	if len(params.Matchers) > 0 {
		cfParams.Matchers = make([]cloudflare.EmailRoutingRuleMatcher, len(params.Matchers))
		for i, matcher := range params.Matchers {
			cfParams.Matchers[i] = cloudflare.EmailRoutingRuleMatcher{
				Type:  matcher.Type,
				Field: matcher.Field,
				Value: matcher.Value,
			}
		}
	}

	// Convert actions
	if len(params.Actions) > 0 {
		cfParams.Actions = make([]cloudflare.EmailRoutingRuleAction, len(params.Actions))
		for i, action := range params.Actions {
			cfParams.Actions[i] = cloudflare.EmailRoutingRuleAction{
				Type:  action.Type,
				Value: action.Value,
			}
		}
	}

	return cfParams
}

// convertToUpdateParams converts Crossplane parameters to cloudflare-go update parameters.
func convertToUpdateParams(ruleTag string, params v1alpha1.RuleParameters) cloudflare.UpdateEmailRoutingRuleParameters {
	cfParams := cloudflare.UpdateEmailRoutingRuleParameters{
		Name:     params.Name,
		Priority: params.Priority,
		Enabled:  params.Enabled,
	}

	// Convert matchers
	if len(params.Matchers) > 0 {
		cfParams.Matchers = make([]cloudflare.EmailRoutingRuleMatcher, len(params.Matchers))
		for i, matcher := range params.Matchers {
			cfParams.Matchers[i] = cloudflare.EmailRoutingRuleMatcher{
				Type:  matcher.Type,
				Field: matcher.Field,
				Value: matcher.Value,
			}
		}
	}

	// Convert actions
	if len(params.Actions) > 0 {
		cfParams.Actions = make([]cloudflare.EmailRoutingRuleAction, len(params.Actions))
		for i, action := range params.Actions {
			cfParams.Actions[i] = cloudflare.EmailRoutingRuleAction{
				Type:  action.Type,
				Value: action.Value,
			}
		}
	}

	return cfParams
}

// Create creates a new Email Routing Rule.
func (c *RuleClient) Create(ctx context.Context, params v1alpha1.RuleParameters) (*v1alpha1.RuleObservation, error) {
	rc := cloudflare.ZoneIdentifier(params.ZoneID)
	
	createParams := convertToCloudflareParams(params)
	
	rule, err := c.client.CreateEmailRoutingRule(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateRule)
	}

	obs := convertToObservation(rule)
	return &obs, nil
}

// Get retrieves an Email Routing Rule.
func (c *RuleClient) Get(ctx context.Context, zoneID, ruleTag string) (*v1alpha1.RuleObservation, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	rule, err := c.client.GetEmailRoutingRule(ctx, rc, ruleTag)
	if err != nil {
		return nil, errors.Wrap(err, errGetRule)
	}

	obs := convertToObservation(rule)
	return &obs, nil
}

// Update updates an existing Email Routing Rule.
func (c *RuleClient) Update(ctx context.Context, ruleTag string, params v1alpha1.RuleParameters) (*v1alpha1.RuleObservation, error) {
	rc := cloudflare.ZoneIdentifier(params.ZoneID)
	
	updateParams := convertToUpdateParams(ruleTag, params)
	
	rule, err := c.client.UpdateEmailRoutingRule(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateRule)
	}

	obs := convertToObservation(rule)
	return &obs, nil
}

// Delete removes an Email Routing Rule.
func (c *RuleClient) Delete(ctx context.Context, zoneID, ruleTag string) error {
	rc := cloudflare.ZoneIdentifier(zoneID)

	_, err := c.client.DeleteEmailRoutingRule(ctx, rc, ruleTag)
	if err != nil && !IsRuleNotFound(err) {
		return errors.Wrap(err, errDeleteRule)
	}

	return nil
}

// List retrieves all Email Routing Rules for a zone.
func (c *RuleClient) List(ctx context.Context, zoneID string) ([]v1alpha1.RuleObservation, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)

	rules, _, err := c.client.ListEmailRoutingRules(ctx, rc, cloudflare.ListEmailRoutingRulesParameters{})
	if err != nil {
		return nil, errors.Wrap(err, errListRules)
	}

	observations := make([]v1alpha1.RuleObservation, len(rules))
	for i, rule := range rules {
		observations[i] = convertToObservation(rule)
	}

	return observations, nil
}

// IsUpToDate checks if the Email Routing Rule is up to date.
func (c *RuleClient) IsUpToDate(ctx context.Context, params v1alpha1.RuleParameters, obs v1alpha1.RuleObservation) (bool, error) {
	// Compare key fields to determine if update is needed
	if obs.Name != params.Name ||
		(obs.Priority != nil && *obs.Priority != params.Priority) {
		return false, nil
	}

	if params.Enabled != nil && (obs.Enabled == nil || *obs.Enabled != *params.Enabled) {
		return false, nil
	}

	// Compare matchers
	if len(params.Matchers) != len(obs.Matchers) {
		return false, nil
	}
	for i, matcher := range params.Matchers {
		if i >= len(obs.Matchers) ||
			matcher.Type != obs.Matchers[i].Type ||
			matcher.Field != obs.Matchers[i].Field ||
			matcher.Value != obs.Matchers[i].Value {
			return false, nil
		}
	}

	// Compare actions
	if len(params.Actions) != len(obs.Actions) {
		return false, nil
	}
	for i, action := range params.Actions {
		if i >= len(obs.Actions) ||
			action.Type != obs.Actions[i].Type ||
			len(action.Value) != len(obs.Actions[i].Value) {
			return false, nil
		}
		for j, value := range action.Value {
			if j >= len(obs.Actions[i].Value) || value != obs.Actions[i].Value[j] {
				return false, nil
			}
		}
	}

	return true, nil
}

// IsRuleNotFound returns true if the error indicates the rule was not found
func IsRuleNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "rule not found" ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}