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

package crontrigger

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateCronTrigger = "cannot create workers cron trigger"
	errUpdateCronTrigger = "cannot update workers cron trigger"
	errGetCronTrigger    = "cannot get workers cron trigger"
	errDeleteCronTrigger = "cannot delete workers cron trigger"
	errListCronTriggers  = "cannot list workers cron triggers"
)

// CronTriggerClient provides operations for Workers Cron Triggers.
type CronTriggerClient struct {
	client    clients.ClientInterface
	accountID string
}

// NewClient creates a new Workers Cron Trigger client.
func NewClient(client clients.ClientInterface) *CronTriggerClient {
	return &CronTriggerClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
	}
}

// getAccountID gets the account ID from the Cloudflare API
func (c *CronTriggerClient) getAccountID(ctx context.Context) (string, error) {
	if c.accountID != "" {
		return c.accountID, nil
	}
	
	// For mock clients, use the GetAccountID method directly
	accountID := c.client.GetAccountID()
	if accountID != "" {
		c.accountID = accountID
		return c.accountID, nil
	}
	
	return "", errors.New("no account ID available")
}

// convertToObservation converts cloudflare-go cron trigger to Crossplane observation.
func convertToObservation(scriptName string, trigger cloudflare.WorkerCronTrigger) v1alpha1.CronTriggerObservation {
	obs := v1alpha1.CronTriggerObservation{
		ScriptName: scriptName,
		Cron:       trigger.Cron,
	}

	if trigger.CreatedOn != nil {
		obs.CreatedOn = &metav1.Time{Time: *trigger.CreatedOn}
	}

	if trigger.ModifiedOn != nil {
		obs.ModifiedOn = &metav1.Time{Time: *trigger.ModifiedOn}
	}

	return obs
}

// Create creates a new Workers Cron Trigger.
// Note: Cloudflare API manages cron triggers as a collection for a script,
// so we update the entire collection to include our new trigger.
func (c *CronTriggerClient) Create(ctx context.Context, params v1alpha1.CronTriggerParameters) (*v1alpha1.CronTriggerObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	// First, get existing triggers to avoid overwriting them
	listParams := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: params.ScriptName,
	}
	
	existingTriggers, err := c.client.ListWorkerCronTriggers(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errListCronTriggers)
	}
	
	// Add our new trigger to the existing ones
	allTriggers := append(existingTriggers, cloudflare.WorkerCronTrigger{
		Cron: params.Cron,
	})
	
	// Update the entire collection
	updateParams := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: params.ScriptName,
		Crons:      allTriggers,
	}
	
	updatedTriggers, err := c.client.UpdateWorkerCronTriggers(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateCronTrigger)
	}
	
	// Find our newly created trigger in the response
	for _, trigger := range updatedTriggers {
		if trigger.Cron == params.Cron {
			obs := convertToObservation(params.ScriptName, trigger)
			return &obs, nil
		}
	}
	
	return nil, errors.New("created cron trigger not found in response")
}

// Get retrieves a Workers Cron Trigger by finding it in the script's trigger list.
func (c *CronTriggerClient) Get(ctx context.Context, scriptName, cronExpression string) (*v1alpha1.CronTriggerObservation, error) {
	rc := cloudflare.AccountIdentifier(c.accountID)
	
	listParams := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: scriptName,
	}
	
	triggers, err := c.client.ListWorkerCronTriggers(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errGetCronTrigger)
	}

	for _, trigger := range triggers {
		if trigger.Cron == cronExpression {
			obs := convertToObservation(scriptName, trigger)
			return &obs, nil
		}
	}

	return nil, errors.New("cron trigger not found")
}

// Update updates an existing Workers Cron Trigger.
// Note: Since Cloudflare manages triggers as a collection, we replace the specific trigger.
func (c *CronTriggerClient) Update(ctx context.Context, oldCron string, params v1alpha1.CronTriggerParameters) (*v1alpha1.CronTriggerObservation, error) {
	rc := cloudflare.AccountIdentifier(c.accountID)
	
	// Get existing triggers
	listParams := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: params.ScriptName,
	}
	
	existingTriggers, err := c.client.ListWorkerCronTriggers(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errListCronTriggers)
	}
	
	// Replace the old trigger with the new one
	var updatedTriggers []cloudflare.WorkerCronTrigger
	found := false
	for _, trigger := range existingTriggers {
		if trigger.Cron == oldCron {
			updatedTriggers = append(updatedTriggers, cloudflare.WorkerCronTrigger{
				Cron: params.Cron,
			})
			found = true
		} else {
			updatedTriggers = append(updatedTriggers, trigger)
		}
	}
	
	if !found {
		return nil, errors.New("cron trigger to update not found")
	}
	
	// Update the entire collection
	updateParams := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: params.ScriptName,
		Crons:      updatedTriggers,
	}
	
	resultTriggers, err := c.client.UpdateWorkerCronTriggers(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateCronTrigger)
	}
	
	// Find our updated trigger in the response
	for _, trigger := range resultTriggers {
		if trigger.Cron == params.Cron {
			obs := convertToObservation(params.ScriptName, trigger)
			return &obs, nil
		}
	}
	
	return nil, errors.New("updated cron trigger not found in response")
}

// Delete removes a Workers Cron Trigger.
func (c *CronTriggerClient) Delete(ctx context.Context, scriptName, cronExpression string) error {
	rc := cloudflare.AccountIdentifier(c.accountID)
	
	// Get existing triggers
	listParams := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: scriptName,
	}
	
	existingTriggers, err := c.client.ListWorkerCronTriggers(ctx, rc, listParams)
	if err != nil {
		return errors.Wrap(err, errListCronTriggers)
	}
	
	// Remove the trigger from the collection
	var remainingTriggers []cloudflare.WorkerCronTrigger
	found := false
	for _, trigger := range existingTriggers {
		if trigger.Cron != cronExpression {
			remainingTriggers = append(remainingTriggers, trigger)
		} else {
			found = true
		}
	}
	
	if !found {
		// Trigger doesn't exist, which is fine for delete
		return nil
	}
	
	// Update the collection without our trigger
	updateParams := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: scriptName,
		Crons:      remainingTriggers,
	}
	
	_, err = c.client.UpdateWorkerCronTriggers(ctx, rc, updateParams)
	if err != nil {
		return errors.Wrap(err, errDeleteCronTrigger)
	}

	return nil
}

// List retrieves all Workers Cron Triggers for a script.
func (c *CronTriggerClient) List(ctx context.Context, scriptName string) ([]v1alpha1.CronTriggerObservation, error) {
	rc := cloudflare.AccountIdentifier(c.accountID)
	
	listParams := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: scriptName,
	}
	
	triggers, err := c.client.ListWorkerCronTriggers(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errListCronTriggers)
	}

	observations := make([]v1alpha1.CronTriggerObservation, len(triggers))
	for i, trigger := range triggers {
		observations[i] = convertToObservation(scriptName, trigger)
	}

	return observations, nil
}

// IsUpToDate checks if the Workers Cron Trigger is up to date.
func (c *CronTriggerClient) IsUpToDate(ctx context.Context, params v1alpha1.CronTriggerParameters, obs v1alpha1.CronTriggerObservation) (bool, error) {
	// For cron triggers, we only need to compare the cron expression
	return obs.Cron == params.Cron && obs.ScriptName == params.ScriptName, nil
}

// IsCronTriggerNotFound returns true if the error indicates the cron trigger was not found
func IsCronTriggerNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "cron trigger not found" ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}