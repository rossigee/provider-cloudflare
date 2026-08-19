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

package posturerule

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// DevicePostureRuleAPI defines the interface for Device Posture Rule operations
type DevicePostureRuleAPI interface {
	DevicePostureRules(ctx context.Context, accountID string) ([]cloudflare.DevicePostureRule, cloudflare.ResultInfo, error)
	DevicePostureRule(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error)
	CreateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	UpdateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	DeleteDevicePostureRule(ctx context.Context, accountID, ruleID string) error
}

// CloudflareDevicePostureRuleClient is a Cloudflare API client for Device Posture Rules.
type CloudflareDevicePostureRuleClient struct {
	client DevicePostureRuleAPI
}

// NewClient creates a new CloudflareDevicePostureRuleClient.
func NewClient(client DevicePostureRuleAPI) *CloudflareDevicePostureRuleClient {
	return &CloudflareDevicePostureRuleClient{client: client}
}

// NewClientFromAPI creates a new CloudflareDevicePostureRuleClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareDevicePostureRuleClient {
	return NewClient(api)
}

// Get retrieves a Device Posture Rule.
func (c *CloudflareDevicePostureRuleClient) Get(ctx context.Context, accountID, ruleID string) (*v1beta1.DevicePostureRuleObservation, error) {
	rule, err := c.client.DevicePostureRule(ctx, accountID, ruleID)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("device posture rule not found")
		}
		return nil, errors.Wrap(err, "cannot get device posture rule")
	}

	return convertDevicePostureRuleToObservation(rule), nil
}

// Create creates a new Device Posture Rule.
func (c *CloudflareDevicePostureRuleClient) Create(ctx context.Context, params v1beta1.DevicePostureRuleParameters) (*v1beta1.DevicePostureRuleObservation, error) {
	createRule := convertParametersToDevicePostureRule(params)

	rule, err := c.client.CreateDevicePostureRule(ctx, params.AccountID, createRule)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create device posture rule")
	}

	return convertDevicePostureRuleToObservation(rule), nil
}

// Update updates a Device Posture Rule.
func (c *CloudflareDevicePostureRuleClient) Update(ctx context.Context, ruleID string, params v1beta1.DevicePostureRuleParameters) (*v1beta1.DevicePostureRuleObservation, error) {
	updateRule := convertParametersToDevicePostureRule(params)
	updateRule.ID = ruleID

	rule, err := c.client.UpdateDevicePostureRule(ctx, params.AccountID, updateRule)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update device posture rule")
	}

	return convertDevicePostureRuleToObservation(rule), nil
}

// Delete deletes a Device Posture Rule.
func (c *CloudflareDevicePostureRuleClient) Delete(ctx context.Context, accountID, ruleID string) error {
	err := c.client.DeleteDevicePostureRule(ctx, accountID, ruleID)
	if err != nil && !isNotFound(err) {
		return errors.Wrap(err, "cannot delete device posture rule")
	}
	return nil
}

// IsUpToDate checks if the Device Posture Rule is up to date.
func (c *CloudflareDevicePostureRuleClient) IsUpToDate(ctx context.Context, params v1beta1.DevicePostureRuleParameters, obs v1beta1.DevicePostureRuleObservation) (bool, error) {
	// Compare key parameters
	if params.Name != obs.Name {
		return false, nil
	}

	if params.Type != obs.Type {
		return false, nil
	}

	if params.Description != nil && *params.Description != obs.Description {
		return false, nil
	}

	if params.Schedule != nil && *params.Schedule != obs.Schedule {
		return false, nil
	}

	if params.Expiration != nil && *params.Expiration != obs.Expiration {
		return false, nil
	}

	return true, nil
}

// convertParametersToDevicePostureRule converts DevicePostureRuleParameters to cloudflare.DevicePostureRule.
func convertParametersToDevicePostureRule(params v1beta1.DevicePostureRuleParameters) cloudflare.DevicePostureRule {
	rule := cloudflare.DevicePostureRule{
		Type: params.Type,
		Name: params.Name,
	}

	if params.Description != nil {
		rule.Description = *params.Description
	}

	if params.Schedule != nil {
		rule.Schedule = *params.Schedule
	}

	if params.Match != nil {
		rule.Match = convertMatch(params.Match)
	}

	if params.Input != nil {
		rule.Input = *convertInput(params.Input)
	}

	if params.Expiration != nil {
		rule.Expiration = *params.Expiration
	}

	return rule
}

// convertDevicePostureRuleToObservation converts cloudflare.DevicePostureRule to DevicePostureRuleObservation.
func convertDevicePostureRuleToObservation(rule cloudflare.DevicePostureRule) *v1beta1.DevicePostureRuleObservation {
	obs := &v1beta1.DevicePostureRuleObservation{
		ID:          rule.ID,
		Type:        rule.Type,
		Name:        rule.Name,
		Description: rule.Description,
		Schedule:    rule.Schedule,
		Expiration:  rule.Expiration,
	}

	if rule.Match != nil {
		obs.Match = convertMatchFromCloudflare(rule.Match)
	}

	// Input is not a pointer in DevicePostureRule
	obs.Input = convertInputFromCloudflare(&rule.Input)

	return obs
}

// convertMatch converts []v1beta1.DevicePostureRuleMatch to []cloudflare.DevicePostureRuleMatch.
func convertMatch(matches []v1beta1.DevicePostureRuleMatch) []cloudflare.DevicePostureRuleMatch {
	cfMatches := make([]cloudflare.DevicePostureRuleMatch, len(matches))
	for i, match := range matches {
		cfMatches[i] = cloudflare.DevicePostureRuleMatch{}
		if match.Platform != nil {
			cfMatches[i].Platform = *match.Platform
		}
	}
	return cfMatches
}

// convertMatchFromCloudflare converts []cloudflare.DevicePostureRuleMatch to []v1beta1.DevicePostureRuleMatch.
func convertMatchFromCloudflare(matches []cloudflare.DevicePostureRuleMatch) []v1beta1.DevicePostureRuleMatch {
	v1Matches := make([]v1beta1.DevicePostureRuleMatch, len(matches))
	for i, match := range matches {
		v1Matches[i] = v1beta1.DevicePostureRuleMatch{
			Platform: &match.Platform,
		}
	}
	return v1Matches
}

// convertInput converts *v1beta1.DevicePostureRuleInput to cloudflare.DevicePostureRuleInput.
func convertInput(input *v1beta1.DevicePostureRuleInput) *cloudflare.DevicePostureRuleInput {
	if input == nil {
		return &cloudflare.DevicePostureRuleInput{}
	}

	cfInput := &cloudflare.DevicePostureRuleInput{}

	if input.ID != nil {
		cfInput.ID = *input.ID
	}

	if input.Path != nil {
		cfInput.Path = *input.Path
	}

	if input.Exists != nil {
		cfInput.Exists = input.Exists
	}

	if input.Thumbprint != nil {
		cfInput.Thumbprint = *input.Thumbprint
	}

	if input.SHA256 != nil {
		cfInput.Sha256 = *input.SHA256
	}

	if input.Running != nil {
		cfInput.Running = input.Running
	}

	if input.RequireAll != nil {
		cfInput.RequireAll = input.RequireAll
	}

	if input.Enabled != nil {
		cfInput.Enabled = input.Enabled
	}

	if input.Version != nil {
		cfInput.Version = *input.Version
	}

	if input.Operator != nil {
		cfInput.Operator = *input.Operator
	}

	if input.Domain != nil {
		cfInput.Domain = *input.Domain
	}

	if input.OS != nil {
		cfInput.Os = *input.OS
	}

	if input.Overall != nil {
		cfInput.Overall = *input.Overall
	}

	if input.SensorConfig != nil {
		cfInput.SensorConfig = *input.SensorConfig
	}

	if input.State != nil {
		cfInput.State = *input.State
	}

	if input.CertificateID != nil {
		cfInput.CertificateID = *input.CertificateID
	}

	if input.CN != nil {
		cfInput.CommonName = *input.CN
	}

	if input.CheckDisks != nil {
		cfInput.CheckDisks = input.CheckDisks
	}

	if input.CheckPrivateKey != nil {
		cfInput.CheckPrivateKey = input.CheckPrivateKey
	}

	if input.ComplianceStatus != nil {
		cfInput.ComplianceStatus = *input.ComplianceStatus
	}

	if input.ConnectionID != nil {
		cfInput.ConnectionID = *input.ConnectionID
	}

	return cfInput
}

// convertInputFromCloudflare converts *cloudflare.DevicePostureRuleInput to *v1beta1.DevicePostureRuleInput.
func convertInputFromCloudflare(input *cloudflare.DevicePostureRuleInput) *v1beta1.DevicePostureRuleInput {
	if input == nil {
		return nil
	}

	v1Input := &v1beta1.DevicePostureRuleInput{
		ID:               &input.ID,
		Path:             &input.Path,
		Exists:           input.Exists,
		Thumbprint:       &input.Thumbprint,
		SHA256:           &input.Sha256,
		Running:          input.Running,
		RequireAll:       input.RequireAll,
		Enabled:          input.Enabled,
		Version:          &input.Version,
		Operator:         &input.Operator,
		Domain:           &input.Domain,
		OS:               &input.Os,
		Overall:          &input.Overall,
		SensorConfig:     &input.SensorConfig,
		State:            &input.State,
		CertificateID:    &input.CertificateID,
		CN:               &input.CommonName,
		CheckDisks:       input.CheckDisks,
		CheckPrivateKey:  input.CheckPrivateKey,
		ComplianceStatus: &input.ComplianceStatus,
		ConnectionID:     &input.ConnectionID,
	}

	return v1Input
}

// isNotFound checks if an error indicates that the device posture rule was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "device posture rule not found") ||
		strings.Contains(errStr, "does not exist")
}
