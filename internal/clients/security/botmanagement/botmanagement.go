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

package botmanagement

import (
	"context"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/security/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// BotManagementAPI defines the interface for Bot Management operations
type BotManagementAPI interface {
	GetBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error)
	UpdateBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error)
}

// CloudflareBotManagementClient is a Cloudflare API client for Bot Management.
type CloudflareBotManagementClient struct {
	client BotManagementAPI
}

// NewClient creates a new CloudflareBotManagementClient.
func NewClient(client BotManagementAPI) *CloudflareBotManagementClient {
	return &CloudflareBotManagementClient{client: client}
}

// NewClientFromAPI creates a new CloudflareBotManagementClient from a Cloudflare API instance.
// This is a wrapper for compatibility with the controller pattern.
func NewClientFromAPI(api *cloudflare.API) *CloudflareBotManagementClient {
	return NewClient(api)
}

// Get retrieves Bot Management configuration for a zone.
func (c *CloudflareBotManagementClient) Get(ctx context.Context, zoneID string) (*v1alpha1.BotManagementObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: zoneID,
		Type:       cloudflare.ZoneType,
	}

	botManagement, err := c.client.GetBotManagement(ctx, rc)
	if err != nil {
		if isNotFound(err) {
			return nil, clients.NewNotFoundError("bot management configuration not found")
		}
		return nil, errors.Wrap(err, "cannot get bot management configuration")
	}

	return convertBotManagementToObservation(botManagement), nil
}

// Update updates Bot Management configuration for a zone.
func (c *CloudflareBotManagementClient) Update(ctx context.Context, params v1alpha1.BotManagementParameters) (*v1alpha1.BotManagementObservation, error) {
	rc := &cloudflare.ResourceContainer{
		Identifier: params.Zone,
		Type:       cloudflare.ZoneType,
	}

	updateParams := convertParametersToBotManagement(params)
	
	botManagement, err := c.client.UpdateBotManagement(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, "cannot update bot management configuration")
	}

	return convertBotManagementToObservation(botManagement), nil
}

// IsUpToDate checks if the Bot Management configuration is up to date.
func (c *CloudflareBotManagementClient) IsUpToDate(ctx context.Context, params v1alpha1.BotManagementParameters, obs v1alpha1.BotManagementObservation) (bool, error) {
	// Compare all configurable parameters
	if params.EnableJS != nil && obs.EnableJS != nil && *params.EnableJS != *obs.EnableJS {
		return false, nil
	}
	
	if params.FightMode != nil && obs.FightMode != nil && *params.FightMode != *obs.FightMode {
		return false, nil
	}
	
	if params.SBFMDefinitelyAutomated != nil && obs.SBFMDefinitelyAutomated != nil && 
		*params.SBFMDefinitelyAutomated != *obs.SBFMDefinitelyAutomated {
		return false, nil
	}
	
	if params.SBFMLikelyAutomated != nil && obs.SBFMLikelyAutomated != nil && 
		*params.SBFMLikelyAutomated != *obs.SBFMLikelyAutomated {
		return false, nil
	}
	
	if params.SBFMVerifiedBots != nil && obs.SBFMVerifiedBots != nil && 
		*params.SBFMVerifiedBots != *obs.SBFMVerifiedBots {
		return false, nil
	}
	
	if params.SBFMStaticResourceProtection != nil && obs.SBFMStaticResourceProtection != nil && 
		*params.SBFMStaticResourceProtection != *obs.SBFMStaticResourceProtection {
		return false, nil
	}
	
	if params.OptimizeWordpress != nil && obs.OptimizeWordpress != nil && 
		*params.OptimizeWordpress != *obs.OptimizeWordpress {
		return false, nil
	}
	
	if params.SuppressSessionScore != nil && obs.SuppressSessionScore != nil && 
		*params.SuppressSessionScore != *obs.SuppressSessionScore {
		return false, nil
	}
	
	if params.AutoUpdateModel != nil && obs.AutoUpdateModel != nil && 
		*params.AutoUpdateModel != *obs.AutoUpdateModel {
		return false, nil
	}
	
	if params.AIBotsProtection != nil && obs.AIBotsProtection != nil && 
		*params.AIBotsProtection != *obs.AIBotsProtection {
		return false, nil
	}
	
	return true, nil
}

// convertParametersToBotManagement converts BotManagementParameters to cloudflare.UpdateBotManagementParams.
func convertParametersToBotManagement(params v1alpha1.BotManagementParameters) cloudflare.UpdateBotManagementParams {
	updateParams := cloudflare.UpdateBotManagementParams{}
	
	if params.EnableJS != nil {
		updateParams.EnableJS = params.EnableJS
	}
	
	if params.FightMode != nil {
		updateParams.FightMode = params.FightMode
	}
	
	if params.SBFMDefinitelyAutomated != nil {
		updateParams.SBFMDefinitelyAutomated = params.SBFMDefinitelyAutomated
	}
	
	if params.SBFMLikelyAutomated != nil {
		updateParams.SBFMLikelyAutomated = params.SBFMLikelyAutomated
	}
	
	if params.SBFMVerifiedBots != nil {
		updateParams.SBFMVerifiedBots = params.SBFMVerifiedBots
	}
	
	if params.SBFMStaticResourceProtection != nil {
		updateParams.SBFMStaticResourceProtection = params.SBFMStaticResourceProtection
	}
	
	if params.OptimizeWordpress != nil {
		updateParams.OptimizeWordpress = params.OptimizeWordpress
	}
	
	if params.SuppressSessionScore != nil {
		updateParams.SuppressSessionScore = params.SuppressSessionScore
	}
	
	if params.AutoUpdateModel != nil {
		updateParams.AutoUpdateModel = params.AutoUpdateModel
	}
	
	if params.AIBotsProtection != nil {
		updateParams.AIBotsProtection = params.AIBotsProtection
	}
	
	return updateParams
}

// convertBotManagementToObservation converts cloudflare.BotManagement to BotManagementObservation.
func convertBotManagementToObservation(botManagement cloudflare.BotManagement) *v1alpha1.BotManagementObservation {
	obs := &v1alpha1.BotManagementObservation{
		EnableJS:                     botManagement.EnableJS,
		FightMode:                    botManagement.FightMode,
		SBFMDefinitelyAutomated:      botManagement.SBFMDefinitelyAutomated,
		SBFMLikelyAutomated:          botManagement.SBFMLikelyAutomated,
		SBFMVerifiedBots:             botManagement.SBFMVerifiedBots,
		SBFMStaticResourceProtection: botManagement.SBFMStaticResourceProtection,
		OptimizeWordpress:            botManagement.OptimizeWordpress,
		SuppressSessionScore:         botManagement.SuppressSessionScore,
		AutoUpdateModel:              botManagement.AutoUpdateModel,
		UsingLatestModel:             botManagement.UsingLatestModel,
		AIBotsProtection:             botManagement.AIBotsProtection,
	}
	
	return obs
}

// isNotFound checks if an error indicates that the bot management configuration was not found.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "resource not found") ||
		strings.Contains(errStr, "bot management not found") ||
		strings.Contains(errStr, "does not exist")
}