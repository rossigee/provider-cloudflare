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

package rule

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/transform/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	// Cloudflare returns this error when a ruleset or rule is not found
	errRulesetNotFound = "10007"
	errRuleNotFound    = "10014"
)

// Client is a Cloudflare API client that implements methods for working
// with Transform Rules via the Ruleset Engine.
type Client interface {
	CreateTransformRule(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error)
	UpdateTransformRule(ctx context.Context, zoneID string, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error)
	GetTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) (cloudflare.RulesetRule, error)
	DeleteTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) error
	ListTransformRules(ctx context.Context, zoneID string, phase string) ([]cloudflare.RulesetRule, error)
}

// clientImpl implements the Client interface by wrapping cloudflare.API
type clientImpl struct {
	*cloudflare.API
}

// NewClient returns a new Cloudflare API client for working with Transform Rules.
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	api, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}
	return &clientImpl{API: api}, nil
}

// CreateTransformRule creates a new transform rule in the appropriate ruleset
func (c *clientImpl) CreateTransformRule(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
	// Get or create the phase ruleset
	ruleset, err := c.getOrCreatePhaseRuleset(ctx, zoneID, spec.Phase)
	if err != nil {
		return cloudflare.RulesetRule{}, errors.Wrap(err, "failed to get or create phase ruleset")
	}

	// Convert spec to Cloudflare RulesetRule
	rule := c.specToRulesetRule(spec)
	
	// Add the rule to the ruleset
	newRules := append(ruleset.Rules, rule)
	
	// Update the ruleset with the new rule
	updateParams := cloudflare.UpdateRulesetParams{
		ID:          ruleset.ID,
		Description: ruleset.Description,
		Rules:       newRules,
	}

	rc := cloudflare.ZoneIdentifier(zoneID)
	updatedRuleset, err := c.API.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return cloudflare.RulesetRule{}, errors.Wrap(err, "failed to update ruleset with new rule")
	}

	// Find and return the newly created rule
	for _, r := range updatedRuleset.Rules {
		if r.Expression == spec.Expression && r.Action == spec.Action {
			return r, nil
		}
	}

	return cloudflare.RulesetRule{}, errors.New("failed to find newly created rule in ruleset")
}

// UpdateTransformRule updates an existing transform rule
func (c *clientImpl) UpdateTransformRule(ctx context.Context, zoneID string, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
	// Get the phase ruleset
	ruleset, err := c.getPhaseRuleset(ctx, zoneID, spec.Phase)
	if err != nil {
		return cloudflare.RulesetRule{}, errors.Wrap(err, "failed to get phase ruleset")
	}

	// Find and update the rule
	var updatedRule cloudflare.RulesetRule
	ruleFound := false
	for i, rule := range ruleset.Rules {
		if rule.ID == ruleID {
			// Update the rule with new spec
			updatedRule = c.specToRulesetRule(spec)
			updatedRule.ID = ruleID
			updatedRule.Version = rule.Version
			ruleset.Rules[i] = updatedRule
			ruleFound = true
			break
		}
	}

	if !ruleFound {
		return cloudflare.RulesetRule{}, errors.New("rule not found in ruleset")
	}

	// Update the ruleset
	updateParams := cloudflare.UpdateRulesetParams{
		ID:          ruleset.ID,
		Description: ruleset.Description,
		Rules:       ruleset.Rules,
	}

	rc := cloudflare.ZoneIdentifier(zoneID)
	_, err = c.API.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return cloudflare.RulesetRule{}, errors.Wrap(err, "failed to update ruleset")
	}

	return updatedRule, nil
}

// GetTransformRule retrieves a transform rule by ID from the specified phase
func (c *clientImpl) GetTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) (cloudflare.RulesetRule, error) {
	// Get the phase ruleset
	ruleset, err := c.getPhaseRuleset(ctx, zoneID, phase)
	if err != nil {
		return cloudflare.RulesetRule{}, errors.Wrap(err, "failed to get phase ruleset")
	}

	// Find the rule
	for _, rule := range ruleset.Rules {
		if rule.ID == ruleID {
			return rule, nil
		}
	}

	return cloudflare.RulesetRule{}, errors.New("rule not found")
}

// DeleteTransformRule deletes a transform rule from the specified phase
func (c *clientImpl) DeleteTransformRule(ctx context.Context, zoneID string, ruleID string, phase string) error {
	// Get the phase ruleset
	ruleset, err := c.getPhaseRuleset(ctx, zoneID, phase)
	if err != nil {
		return errors.Wrap(err, "failed to get phase ruleset")
	}

	// Remove the rule from the ruleset
	var newRules []cloudflare.RulesetRule
	ruleFound := false
	for _, rule := range ruleset.Rules {
		if rule.ID != ruleID {
			newRules = append(newRules, rule)
		} else {
			ruleFound = true
		}
	}

	if !ruleFound {
		return errors.New("rule not found in ruleset")
	}

	// Update the ruleset without the deleted rule
	updateParams := cloudflare.UpdateRulesetParams{
		ID:          ruleset.ID,
		Description: ruleset.Description,
		Rules:       newRules,
	}

	rc := cloudflare.ZoneIdentifier(zoneID)
	_, err = c.API.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return errors.Wrap(err, "failed to update ruleset after deleting rule")
	}

	return nil
}

// ListTransformRules lists all transform rules in the specified phase
func (c *clientImpl) ListTransformRules(ctx context.Context, zoneID string, phase string) ([]cloudflare.RulesetRule, error) {
	// Get the phase ruleset
	ruleset, err := c.getPhaseRuleset(ctx, zoneID, phase)
	if err != nil {
		if IsRulesetNotFound(err) {
			// No ruleset exists for this phase yet, return empty list
			return []cloudflare.RulesetRule{}, nil
		}
		return nil, errors.Wrap(err, "failed to get phase ruleset")
	}

	return ruleset.Rules, nil
}

// Helper methods

// getPhaseRuleset gets the ruleset for a specific phase
func (c *clientImpl) getPhaseRuleset(ctx context.Context, zoneID string, phase string) (cloudflare.Ruleset, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	
	// Try to get the entrypoint ruleset for this phase
	ruleset, err := c.API.GetEntrypointRuleset(ctx, rc, phase)
	if err != nil {
		return cloudflare.Ruleset{}, err
	}

	return ruleset, nil
}

// getOrCreatePhaseRuleset gets an existing ruleset or creates a new one for the phase
func (c *clientImpl) getOrCreatePhaseRuleset(ctx context.Context, zoneID string, phase string) (cloudflare.Ruleset, error) {
	// Try to get existing ruleset
	ruleset, err := c.getPhaseRuleset(ctx, zoneID, phase)
	if err == nil {
		return ruleset, nil
	}

	// If ruleset doesn't exist, create it
	if IsRulesetNotFound(err) {
		createParams := cloudflare.CreateRulesetParams{
			Name:        fmt.Sprintf("Transform Rules - %s", phase),
			Description: fmt.Sprintf("Transform rules for %s phase", phase),
			Kind:        string(cloudflare.RulesetKindZone),
			Phase:       phase,
			Rules:       []cloudflare.RulesetRule{},
		}

		rc := cloudflare.ZoneIdentifier(zoneID)
		newRuleset, err := c.API.CreateRuleset(ctx, rc, createParams)
		if err != nil {
			return cloudflare.Ruleset{}, errors.Wrap(err, "failed to create new ruleset")
		}

		return newRuleset, nil
	}

	return cloudflare.Ruleset{}, err
}

// specToRulesetRule converts a v1alpha1.RuleParameters to cloudflare.RulesetRule
func (c *clientImpl) specToRulesetRule(spec *v1alpha1.RuleParameters) cloudflare.RulesetRule {
	rule := cloudflare.RulesetRule{
		Expression: spec.Expression,
		Action:     spec.Action,
	}

	if spec.Description != nil {
		rule.Description = *spec.Description
	}

	if spec.Enabled != nil {
		rule.Enabled = spec.Enabled
	} else {
		// Default to enabled
		enabled := true
		rule.Enabled = &enabled
	}

	// Convert action parameters
	if spec.ActionParameters != nil {
		actionParams := &cloudflare.RulesetRuleActionParameters{}

		// URI transformations
		if spec.ActionParameters.URI != nil {
			uriParams := &cloudflare.RulesetRuleActionParametersURI{}

			if spec.ActionParameters.URI.Path != nil {
				pathParams := &cloudflare.RulesetRuleActionParametersURIPath{}
				if spec.ActionParameters.URI.Path.Value != nil {
					pathParams.Value = *spec.ActionParameters.URI.Path.Value
				}
				if spec.ActionParameters.URI.Path.Expression != nil {
					pathParams.Expression = *spec.ActionParameters.URI.Path.Expression
				}
				uriParams.Path = pathParams
			}

			if spec.ActionParameters.URI.Query != nil {
				queryParams := &cloudflare.RulesetRuleActionParametersURIQuery{}
				if spec.ActionParameters.URI.Query.Value != nil {
					queryParams.Value = spec.ActionParameters.URI.Query.Value
				}
				if spec.ActionParameters.URI.Query.Expression != nil {
					queryParams.Expression = *spec.ActionParameters.URI.Query.Expression
				}
				uriParams.Query = queryParams
			}

			actionParams.URI = uriParams
		}

		// Header transformations
		if spec.ActionParameters.Headers != nil && len(spec.ActionParameters.Headers) > 0 {
			headers := make(map[string]cloudflare.RulesetRuleActionParametersHTTPHeader)
			for name, header := range spec.ActionParameters.Headers {
				cfHeader := cloudflare.RulesetRuleActionParametersHTTPHeader{
					Operation: header.Operation,
				}
				if header.Value != nil {
					cfHeader.Value = *header.Value
				}
				if header.Expression != nil {
					cfHeader.Expression = *header.Expression
				}
				headers[name] = cfHeader
			}
			actionParams.Headers = headers
		}

		// Status code for redirects
		if spec.ActionParameters.StatusCode != nil {
			actionParams.StatusCode = uint16(*spec.ActionParameters.StatusCode)
		}

		rule.ActionParameters = actionParams
	}

	return rule
}

// IsRulesetNotFound returns true if the passed error indicates a ruleset was not found
func IsRulesetNotFound(err error) bool {
	return strings.Contains(err.Error(), errRulesetNotFound)
}

// IsRuleNotFound returns true if the passed error indicates a rule was not found
func IsRuleNotFound(err error) bool {
	return strings.Contains(err.Error(), errRuleNotFound)
}

// GenerateObservation creates an observation from a Cloudflare RulesetRule
func GenerateObservation(rule cloudflare.RulesetRule, rulesetID string) v1alpha1.RuleObservation {
	obs := v1alpha1.RuleObservation{
		ID:        rule.ID,
		RulesetID: rulesetID,
	}

	if rule.Version != nil {
		obs.Version = *rule.Version
	}

	if rule.LastUpdated != nil {
		obs.LastUpdated = &metav1.Time{Time: *rule.LastUpdated}
	}

	return obs
}

// UpToDate checks if the remote rule is up to date with the requested resource parameters
func UpToDate(spec *v1alpha1.RuleParameters, rule cloudflare.RulesetRule) bool {
	if spec == nil {
		return true
	}

	// Check expression
	if spec.Expression != rule.Expression {
		return false
	}

	// Check action
	if spec.Action != rule.Action {
		return false
	}

	// Check description
	if spec.Description != nil && rule.Description != *spec.Description {
		return false
	}
	if spec.Description == nil && rule.Description != "" {
		return false
	}

	// Check enabled status
	if spec.Enabled != nil && rule.Enabled != nil {
		if *spec.Enabled != *rule.Enabled {
			return false
		}
	}

	// Check action parameters (simplified - can be expanded based on needs)
	if spec.ActionParameters != nil {
		if rule.ActionParameters == nil {
			return false
		}
		// Additional parameter checking can be added here as needed
	} else if rule.ActionParameters != nil {
		return false
	}

	return true
}