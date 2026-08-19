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

package ruleset

import (
	"context"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/rulesets/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateRuleset = "failed to create ruleset"
	errGetRuleset    = "failed to get ruleset"
	errUpdateRuleset = "failed to update ruleset"
	errDeleteRuleset = "failed to delete ruleset"
)

// Client interface for Cloudflare Ruleset operations
type Client interface {
	CreateRuleset(ctx context.Context, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error)
	GetRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error)
	UpdateRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error)
	DeleteRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) error
}

// NewClient creates a new Cloudflare Ruleset client
func NewClient(cfg clients.Config, httpClient *http.Client) (Client, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &client{api: api}, nil
}

type client struct {
	api *cloudflare.API
}

// CreateRuleset creates a new Cloudflare ruleset
func (c *client) CreateRuleset(ctx context.Context, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error) {
	createParams := cloudflare.CreateRulesetParams{
		Name:  params.Name,
		Kind:  params.Kind,
		Phase: params.Phase,
		Rules: convertRulesToCloudflare(params.Rules),
	}

	if params.Description != nil {
		createParams.Description = *params.Description
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	ruleset, err := c.api.CreateRuleset(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateRuleset)
	}

	return &ruleset, nil
}

// GetRuleset retrieves a Cloudflare ruleset
func (c *client) GetRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error) {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	ruleset, err := c.api.GetRuleset(ctx, rc, rulesetID)
	if err != nil {
		return nil, errors.Wrap(err, errGetRuleset)
	}

	return &ruleset, nil
}

// UpdateRuleset updates a Cloudflare ruleset
func (c *client) UpdateRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) (*cloudflare.Ruleset, error) {
	updateParams := cloudflare.UpdateRulesetParams{
		ID:    rulesetID,
		Rules: convertRulesToCloudflare(params.Rules),
	}

	if params.Description != nil {
		updateParams.Description = *params.Description
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	ruleset, err := c.api.UpdateRuleset(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateRuleset)
	}

	return &ruleset, nil
}

// DeleteRuleset deletes a Cloudflare ruleset
func (c *client) DeleteRuleset(ctx context.Context, rulesetID string, params v1beta1.RulesetParameters) error {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return errors.New("either zone or account must be specified")
	}

	err := c.api.DeleteRuleset(ctx, rc, rulesetID)
	if err != nil {
		return errors.Wrap(err, errDeleteRuleset)
	}

	return nil
}

// IsRulesetNotFound checks if error indicates ruleset not found
func IsRulesetNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// convertRulesToCloudflare converts v1beta1 rules to Cloudflare API format
func convertRulesToCloudflare(rules []v1beta1.RulesetRule) []cloudflare.RulesetRule {
	var cfRules []cloudflare.RulesetRule

	for _, rule := range rules {
		cfRule := cloudflare.RulesetRule{
			Action:     rule.Action,
			Expression: rule.Expression,
		}

		if rule.Description != nil {
			cfRule.Description = *rule.Description
		}

		if rule.Enabled != nil {
			cfRule.Enabled = rule.Enabled
		}

		cfRules = append(cfRules, cfRule)
	}

	return cfRules
}

// GenerateObservation creates observation from Cloudflare ruleset
func GenerateObservation(ruleset *cloudflare.Ruleset) v1beta1.RulesetObservation {
	observation := v1beta1.RulesetObservation{
		ID: ruleset.ID,
	}

	if ruleset.Version != nil && *ruleset.Version != "" {
		observation.Version = *ruleset.Version
	}

	return observation
}

// UpToDate determines if the Cloudflare ruleset is up to date
func UpToDate(params *v1beta1.RulesetParameters, ruleset *cloudflare.Ruleset) bool {
	if params.Name != ruleset.Name {
		return false
	}

	if params.Description != nil && *params.Description != ruleset.Description {
		return false
	}

	if params.Description == nil && ruleset.Description != "" {
		return false
	}

	if params.Kind != ruleset.Kind {
		return false
	}

	if params.Phase != ruleset.Phase {
		return false
	}

	// For simplicity, we'll consider rules changed if the count differs
	// A more sophisticated comparison could be implemented if needed
	if len(params.Rules) != len(ruleset.Rules) {
		return false
	}

	return true
}
