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
	"math"
	"strings"
	"sync"
	"time"

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
	
	// Cache TTL for API responses within the same reconcile cycle
	cacheTimeout = 30 * time.Second
	
	// Retry configuration for rate limiting
	maxRetries = 3
	baseDelay  = 2 * time.Second
)

// scriptCache holds cached API responses to avoid duplicate calls within the same reconcile cycle
type scriptCache struct {
	mu                    sync.RWMutex
	workerData           map[string]*cachedWorkerData
	scriptContent        map[string]*cachedScriptContent
	scriptSettings       map[string]*cachedScriptSettings
}

type cachedWorkerData struct {
	data      cloudflare.WorkerScriptResponse
	timestamp time.Time
}

type cachedScriptContent struct {
	content   string
	timestamp time.Time
}

type cachedScriptSettings struct {
	settings  cloudflare.WorkerScriptSettingsResponse
	timestamp time.Time
}

// ScriptClient provides operations for Worker Scripts.
type ScriptClient struct {
	client    clients.ClientInterface
	accountID string
	cache     *scriptCache
}

// NewClient creates a new Worker Script client.
func NewClient(client clients.ClientInterface) *ScriptClient {
	return &ScriptClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
		cache: &scriptCache{
			workerData:     make(map[string]*cachedWorkerData),
			scriptContent:  make(map[string]*cachedScriptContent),
			scriptSettings: make(map[string]*cachedScriptSettings),
		},
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

// Cache helper methods
func (c *ScriptClient) getWorkerDataFromCache(scriptName string) (*cloudflare.WorkerScriptResponse, bool) {
	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()
	
	cached, exists := c.cache.workerData[scriptName]
	if !exists || time.Since(cached.timestamp) > cacheTimeout {
		return nil, false
	}
	return &cached.data, true
}

func (c *ScriptClient) setWorkerDataInCache(scriptName string, data cloudflare.WorkerScriptResponse) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	
	c.cache.workerData[scriptName] = &cachedWorkerData{
		data:      data,
		timestamp: time.Now(),
	}
}

func (c *ScriptClient) getScriptContentFromCache(scriptName string) (string, bool) {
	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()
	
	cached, exists := c.cache.scriptContent[scriptName]
	if !exists || time.Since(cached.timestamp) > cacheTimeout {
		return "", false
	}
	return cached.content, true
}

func (c *ScriptClient) setScriptContentInCache(scriptName string, content string) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	
	c.cache.scriptContent[scriptName] = &cachedScriptContent{
		content:   content,
		timestamp: time.Now(),
	}
}

func (c *ScriptClient) getScriptSettingsFromCache(scriptName string) (*cloudflare.WorkerScriptSettingsResponse, bool) {
	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()
	
	cached, exists := c.cache.scriptSettings[scriptName]
	if !exists || time.Since(cached.timestamp) > cacheTimeout {
		return nil, false
	}
	return &cached.settings, true
}

func (c *ScriptClient) setScriptSettingsInCache(scriptName string, settings cloudflare.WorkerScriptSettingsResponse) {
	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()
	
	c.cache.scriptSettings[scriptName] = &cachedScriptSettings{
		settings:  settings,
		timestamp: time.Now(),
	}
}

// isRateLimitError checks if an error is due to rate limiting
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "rate limit") || 
		   strings.Contains(errStr, "429") ||
		   strings.Contains(errStr, "too many requests")
}

// retryWithBackoff executes a function with exponential backoff on rate limit errors
func (c *ScriptClient) retryWithBackoff(ctx context.Context, operation func() error) error {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: baseDelay * 2^(attempt-1) with jitter
			delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
			// Add 10% jitter to avoid thundering herd
			jitter := time.Duration(float64(delay) * 0.1 * float64(2*time.Now().UnixNano()%2 - 1) / 1e9)
			delay += jitter
			
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		
		lastErr = operation()
		if lastErr == nil {
			return nil
		}
		
		// Only retry on rate limit errors
		if !isRateLimitError(lastErr) {
			return lastErr
		}
		
		// Don't retry if this was the last attempt
		if attempt == maxRetries {
			break
		}
	}
	
	return errors.Wrap(lastErr, "max retries exceeded")
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
	
	// Debug logging
	// TODO: Remove debug logging after issue is resolved
	if accountID == "" {
		return nil, errors.New("DEBUG: accountID is empty")
	}
	if createParams.ScriptName == "" {
		return nil, errors.New("DEBUG: ScriptName is empty")
	}
	if createParams.Script == "" {
		return nil, errors.New("DEBUG: Script content is empty")
	}
	
	resp, err := c.client.UploadWorker(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateScript)
	}

	// Debug logging for response
	if resp.ID == "" {
		return nil, errors.New("DEBUG: Response WorkerMetaData.ID is empty - accountID=" + accountID + ", scriptName=" + createParams.ScriptName)
	}
	
	// Success debug logging - convert and return observation
	obs := convertToObservation(resp.WorkerMetaData, &resp.WorkerScript)
	return &obs, nil
}

// Get retrieves a Worker script with caching to reduce API calls.
func (c *ScriptClient) Get(ctx context.Context, scriptName string) (*v1alpha1.ScriptObservation, error) {
	// Try to get from cache first
	if cachedWorkerData, ok := c.getWorkerDataFromCache(scriptName); ok {
		if cachedSettings, ok := c.getScriptSettingsFromCache(scriptName); ok {
			// Both are cached, use cached data
			workerScript := cloudflare.WorkerScript{
				WorkerMetaData: cachedWorkerData.WorkerMetaData,
				Script:         cachedWorkerData.Script,
				UsageModel:     cachedWorkerData.UsageModel,
			}
			obs := convertToObservation(cachedSettings.WorkerMetaData, &workerScript)
			return &obs, nil
		}
	}

	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	// Get script content and metadata (only if not cached)
	var scriptResp cloudflare.WorkerScriptResponse
	if cachedWorkerData, ok := c.getWorkerDataFromCache(scriptName); ok {
		scriptResp = *cachedWorkerData
	} else {
		err = c.retryWithBackoff(ctx, func() error {
			scriptResp, err = c.client.GetWorker(ctx, rc, scriptName)
			return err
		})
		if err != nil {
			return nil, errors.Wrap(err, errGetScript)
		}
		// Cache the worker data
		c.setWorkerDataInCache(scriptName, scriptResp)
	}

	// Get script settings for additional metadata (only if not cached)
	var settingsResp cloudflare.WorkerScriptSettingsResponse
	if cachedSettings, ok := c.getScriptSettingsFromCache(scriptName); ok {
		settingsResp = *cachedSettings
	} else {
		err = c.retryWithBackoff(ctx, func() error {
			settingsResp, err = c.client.GetWorkersScriptSettings(ctx, rc, scriptName)
			return err
		})
		if err != nil {
			return nil, errors.Wrap(err, errGetScriptSettings)
		}
		// Cache the settings
		c.setScriptSettingsInCache(scriptName, settingsResp)
	}

	// Create a WorkerScript from the fields in scriptResp
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

// IsUpToDate checks if the Worker script is up to date using cached data when possible.
func (c *ScriptClient) IsUpToDate(ctx context.Context, params v1alpha1.ScriptParameters, obs v1alpha1.ScriptObservation) (bool, error) {
	// Try to get script content from cache first
	var currentScript string
	if cachedContent, ok := c.getScriptContentFromCache(params.ScriptName); ok {
		currentScript = cachedContent
	} else {
		// Get current script content for comparison
		accountID, err := c.getAccountID(ctx)
		if err != nil {
			return false, errors.Wrap(err, "failed to get account ID")
		}
		rc := cloudflare.AccountIdentifier(accountID)
		
		err = c.retryWithBackoff(ctx, func() error {
			currentScript, err = c.client.GetWorkersScriptContent(ctx, rc, params.ScriptName)
			return err
		})
		if err != nil {
			return false, errors.Wrap(err, errGetScript)
		}
		// Cache the script content
		c.setScriptContentInCache(params.ScriptName, currentScript)
	}

	// Compare script content
	if currentScript != params.Script {
		return false, nil
	}

	// Try to get settings from cache first
	var settingsResp cloudflare.WorkerScriptSettingsResponse
	if cachedSettings, ok := c.getScriptSettingsFromCache(params.ScriptName); ok {
		settingsResp = *cachedSettings
	} else {
		// Get current settings for metadata comparison
		accountID, err := c.getAccountID(ctx)
		if err != nil {
			return false, errors.Wrap(err, "failed to get account ID")
		}
		rc := cloudflare.AccountIdentifier(accountID)
		
		err = c.retryWithBackoff(ctx, func() error {
			settingsResp, err = c.client.GetWorkersScriptSettings(ctx, rc, params.ScriptName)
			return err
		})
		if err != nil {
			return false, errors.Wrap(err, errGetScriptSettings)
		}
		// Cache the settings
		c.setScriptSettingsInCache(params.ScriptName, settingsResp)
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