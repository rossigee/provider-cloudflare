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

package clients

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

// CloudflareAPIAdapter adapts *cloudflare.API to implement ClientInterface
type CloudflareAPIAdapter struct {
	api       *cloudflare.API
	accountID string
}

// NewCloudflareAPIAdapter creates a new adapter for cloudflare.API
func NewCloudflareAPIAdapter(api *cloudflare.API) *CloudflareAPIAdapter {
	return &CloudflareAPIAdapter{
		api: api,
	}
}

// GetAccountID returns the account ID
func (a *CloudflareAPIAdapter) GetAccountID() string {
	if a.accountID != "" {
		return a.accountID
	}
	
	// Get account ID from Cloudflare API by listing accounts
	// Most users have access to only one account, so we'll use the first one
	accounts, _, err := a.api.Accounts(context.Background(), cloudflare.AccountsListParams{})
	if err == nil && len(accounts) > 0 {
		a.accountID = accounts[0].ID
	}
	
	return a.accountID
}

// UploadWorker wraps the cloudflare API
func (a *CloudflareAPIAdapter) UploadWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerParams) (cloudflare.WorkerScriptResponse, error) {
	return a.api.UploadWorker(ctx, rc, params)
}

// GetWorker wraps the cloudflare API
func (a *CloudflareAPIAdapter) GetWorker(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptResponse, error) {
	return a.api.GetWorker(ctx, rc, scriptName)
}

// DeleteWorker wraps the cloudflare API
func (a *CloudflareAPIAdapter) DeleteWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkerParams) error {
	return a.api.DeleteWorker(ctx, rc, params)
}

// GetWorkersScriptContent wraps the cloudflare API
func (a *CloudflareAPIAdapter) GetWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (string, error) {
	return a.api.GetWorkersScriptContent(ctx, rc, scriptName)
}

// GetWorkersScriptSettings wraps the cloudflare API  
func (a *CloudflareAPIAdapter) GetWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptSettingsResponse, error) {
	return a.api.GetWorkersScriptSettings(ctx, rc, scriptName)
}

// ListWorkers wraps the cloudflare API
func (a *CloudflareAPIAdapter) ListWorkers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersParams) (cloudflare.WorkerListResponse, *cloudflare.ResultInfo, error) {
	return a.api.ListWorkers(ctx, rc, params)
}

// CreateWorkersKVNamespace wraps the cloudflare API
func (a *CloudflareAPIAdapter) CreateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersKVNamespaceParams) (cloudflare.WorkersKVNamespaceResponse, error) {
	return a.api.CreateWorkersKVNamespace(ctx, rc, params)
}

// ListWorkersKVNamespaces wraps the cloudflare API
func (a *CloudflareAPIAdapter) ListWorkersKVNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVNamespacesParams) ([]cloudflare.WorkersKVNamespace, *cloudflare.ResultInfo, error) {
	return a.api.ListWorkersKVNamespaces(ctx, rc, params)
}

// DeleteWorkersKVNamespace wraps the cloudflare API
func (a *CloudflareAPIAdapter) DeleteWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, namespaceID string) (cloudflare.Response, error) {
	return a.api.DeleteWorkersKVNamespace(ctx, rc, namespaceID)
}

// UpdateWorkersKVNamespace wraps the cloudflare API
func (a *CloudflareAPIAdapter) UpdateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersKVNamespaceParams) (cloudflare.Response, error) {
	return a.api.UpdateWorkersKVNamespace(ctx, rc, params)
}

// ListWorkerCronTriggers wraps the cloudflare API
func (a *CloudflareAPIAdapter) ListWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	return a.api.ListWorkerCronTriggers(ctx, rc, params)
}

// UpdateWorkerCronTriggers wraps the cloudflare API
func (a *CloudflareAPIAdapter) UpdateWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	return a.api.UpdateWorkerCronTriggers(ctx, rc, params)
}

// ListWorkerRoutes wraps the cloudflare API
func (a *CloudflareAPIAdapter) ListWorkerRoutes(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerRoutesParams) (cloudflare.WorkerRoutesResponse, error) {
	return a.api.ListWorkerRoutes(ctx, rc, params)
}

// CreateWorkerRoute wraps the cloudflare API
func (a *CloudflareAPIAdapter) CreateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error) {
	return a.api.CreateWorkerRoute(ctx, rc, params)
}

// UpdateWorkerRoute wraps the cloudflare API
func (a *CloudflareAPIAdapter) UpdateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error) {
	return a.api.UpdateWorkerRoute(ctx, rc, params)
}

// DeleteWorkerRoute wraps the cloudflare API
func (a *CloudflareAPIAdapter) DeleteWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, routeID string) (cloudflare.WorkerRouteResponse, error) {
	return a.api.DeleteWorkerRoute(ctx, rc, routeID)
}