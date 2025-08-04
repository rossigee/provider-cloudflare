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

package script

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateScript      = "cannot create worker script"
	errUpdateScript      = "cannot update worker script"
	errGetScript         = "cannot get worker script"
	errDeleteScript      = "cannot delete worker script"
	errListScripts       = "cannot list worker scripts"
	errGetScriptSettings = "cannot get worker script settings"
)

// ScriptClient provides operations for Worker Scripts.
type ScriptClient struct {
	client    clients.ClientInterface
	accountID string
}

// NewClient creates a new Worker Script client.
func NewClient(client clients.ClientInterface) *ScriptClient {
	return &ScriptClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
	}
}

// getAccountID gets the account ID from the Cloudflare API
func (c *ScriptClient) getAccountID(ctx context.Context) (string, error) {
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

// convertToCloudflareBindings converts Crossplane bindings to cloudflare-go bindings.
func convertToCloudflareBindings(bindings []v1alpha1.WorkerBinding) map[string]cloudflare.WorkerBinding {
	cfBindings := make(map[string]cloudflare.WorkerBinding)
	
	for _, binding := range bindings {
		switch binding.Type {
		case "kv_namespace":
			if binding.NamespaceID != nil {
				cfBindings[binding.Name] = cloudflare.WorkerKvNamespaceBinding{
					NamespaceID: *binding.NamespaceID,
				}
			}
		case "wasm_module":
			// TODO: WebAssembly bindings require io.Reader, need to implement file handling
			// Skip for now
		case "text_blob":
			if binding.Text != nil {
				cfBindings[binding.Name] = cloudflare.WorkerPlainTextBinding{
					Text: *binding.Text,
				}
			}
		case "json_data":
			if binding.JSON != nil {
				cfBindings[binding.Name] = cloudflare.WorkerInheritBinding{
					OldName: *binding.JSON,
				}
			}
		}
	}
	
	return cfBindings
}

// convertToCloudflareConsumers converts Crossplane tail consumers to cloudflare-go consumers.
func convertToCloudflareConsumers(consumers []v1alpha1.TailConsumer) *[]cloudflare.WorkersTailConsumer {
	if len(consumers) == 0 {
		return nil
	}
	
	cfConsumers := make([]cloudflare.WorkersTailConsumer, len(consumers))
	for i, consumer := range consumers {
		cfConsumers[i] = cloudflare.WorkersTailConsumer{
			Service:     consumer.Service,
			Environment: consumer.Environment,
			Namespace:   consumer.Namespace,
		}
	}
	
	return &cfConsumers
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.ScriptParameters) cloudflare.CreateWorkerParams {
	createParams := cloudflare.CreateWorkerParams{
		ScriptName: params.ScriptName,
		Script:     params.Script,
		Bindings:   convertToCloudflareBindings(params.Bindings),
		Tags:       params.Tags,
	}

	if params.Module != nil {
		createParams.Module = *params.Module
	}

	if params.CompatibilityDate != nil {
		createParams.CompatibilityDate = *params.CompatibilityDate
	}

	if params.CompatibilityFlags != nil {
		createParams.CompatibilityFlags = params.CompatibilityFlags
	}

	if params.Logpush != nil {
		createParams.Logpush = params.Logpush
	}

	if params.TailConsumers != nil {
		createParams.TailConsumers = convertToCloudflareConsumers(params.TailConsumers)
	}

	if params.PlacementMode != nil {
		placement := &cloudflare.Placement{
			Mode: cloudflare.PlacementMode(*params.PlacementMode),
		}
		createParams.Placement = placement
	}

	if params.DispatchNamespace != nil {
		createParams.DispatchNamespaceName = params.DispatchNamespace
	}

	return createParams
}

// convertToObservation converts cloudflare-go worker metadata to Crossplane observation.
func convertToObservation(metadata cloudflare.WorkerMetaData, script *cloudflare.WorkerScript) v1alpha1.ScriptObservation {
	obs := v1alpha1.ScriptObservation{
		ID:    metadata.ID,
		ETAG:  metadata.ETAG,
		Size:  metadata.Size,
	}

	if !metadata.CreatedOn.IsZero() {
		obs.CreatedOn = &metav1.Time{Time: metadata.CreatedOn}
	}

	if !metadata.ModifiedOn.IsZero() {
		obs.ModifiedOn = &metav1.Time{Time: metadata.ModifiedOn}
	}

	if metadata.LastDeployedFrom != nil {
		obs.LastDeployedFrom = metadata.LastDeployedFrom
	}

	if metadata.DeploymentId != nil {
		obs.DeploymentID = metadata.DeploymentId
	}

	if metadata.Placement != nil && metadata.Placement.Status != "" {
		status := string(metadata.Placement.Status)
		obs.PlacementStatus = &status
	}

	if metadata.PipelineHash != nil {
		obs.PipelineHash = metadata.PipelineHash
	}

	if script != nil && script.UsageModel != "" {
		obs.UsageModel = &script.UsageModel
	}

	return obs
}

// Create creates a new Worker script.
func (c *ScriptClient) Create(ctx context.Context, params v1alpha1.ScriptParameters) (*v1alpha1.ScriptObservation, error) {
	createParams := convertToCloudflareParams(params)
	
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	resp, err := c.client.UploadWorker(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateScript)
	}

	obs := convertToObservation(resp.WorkerMetaData, &resp.WorkerScript)
	return &obs, nil
}

// Get retrieves a Worker script.
func (c *ScriptClient) Get(ctx context.Context, scriptName string) (*v1alpha1.ScriptObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	// Get script content and metadata
	scriptResp, err := c.client.GetWorker(ctx, rc, scriptName)
	if err != nil {
		return nil, errors.Wrap(err, errGetScript)
	}

	// Get script settings for additional metadata
	settingsResp, err := c.client.GetWorkersScriptSettings(ctx, rc, scriptName)
	if err != nil {
		return nil, errors.Wrap(err, errGetScriptSettings)
	}

	// Create a WorkerScript from the embedded fields in scriptResp
	workerScript := cloudflare.WorkerScript{
		WorkerMetaData: scriptResp.WorkerMetaData,
		Script:         scriptResp.Script,
		UsageModel:     scriptResp.UsageModel,
	}
	obs := convertToObservation(settingsResp.WorkerMetaData, &workerScript)
	return &obs, nil
}

// Update updates an existing Worker script.
func (c *ScriptClient) Update(ctx context.Context, params v1alpha1.ScriptParameters) (*v1alpha1.ScriptObservation, error) {
	createParams := convertToCloudflareParams(params)
	
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	// Use UploadWorker which handles both create and update
	resp, err := c.client.UploadWorker(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateScript)
	}

	obs := convertToObservation(resp.WorkerMetaData, &resp.WorkerScript)
	return &obs, nil
}

// Delete removes a Worker script.
func (c *ScriptClient) Delete(ctx context.Context, scriptName string, dispatchNamespace *string) error {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	deleteParams := cloudflare.DeleteWorkerParams{
		ScriptName: scriptName,
	}
	
	if dispatchNamespace != nil {
		deleteParams.DispatchNamespace = dispatchNamespace
	}
	
	err = c.client.DeleteWorker(ctx, rc, deleteParams)
	if err != nil {
		return errors.Wrap(err, errDeleteScript)
	}

	return nil
}

// List retrieves all Worker scripts.
func (c *ScriptClient) List(ctx context.Context) ([]v1alpha1.ScriptObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	listParams := cloudflare.ListWorkersParams{}
	
	resp, _, err := c.client.ListWorkers(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errListScripts)
	}

	observations := make([]v1alpha1.ScriptObservation, len(resp.WorkerList))
	for i, worker := range resp.WorkerList {
		observations[i] = convertToObservation(worker, nil)
	}

	return observations, nil
}

// IsUpToDate checks if the Worker script is up to date.
func (c *ScriptClient) IsUpToDate(ctx context.Context, params v1alpha1.ScriptParameters, obs v1alpha1.ScriptObservation) (bool, error) {
	// Get current script content for comparison
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return false, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	currentScript, err := c.client.GetWorkersScriptContent(ctx, rc, params.ScriptName)
	if err != nil {
		return false, errors.Wrap(err, errGetScript)
	}

	// Compare script content
	if currentScript != params.Script {
		return false, nil
	}

	// Get current settings for metadata comparison
	settingsResp, err := c.client.GetWorkersScriptSettings(ctx, rc, params.ScriptName)
	if err != nil {
		return false, errors.Wrap(err, errGetScriptSettings)
	}

	// Compare key metadata fields that affect the script
	
	// Compare logpush setting
	if params.Logpush != nil {
		if settingsResp.Logpush == nil || *settingsResp.Logpush != *params.Logpush {
			return false, nil
		}
	} else if settingsResp.Logpush != nil && *settingsResp.Logpush {
		return false, nil
	}

	// TODO: Compare compatibility date
	// CompatibilityDate is not available in WorkerScriptSettingsResponse
	// May need to get this from a different API call or compare during creation only

	// Compare placement mode
	if params.PlacementMode != nil {
		if settingsResp.Placement == nil || 
		   string(settingsResp.Placement.Mode) != string(*params.PlacementMode) {
			return false, nil
		}
	}

	// For comprehensive comparison, we could compare bindings, compatibility flags, etc.
	// For now, we'll consider it up to date if script content and key settings match
	
	return true, nil
}