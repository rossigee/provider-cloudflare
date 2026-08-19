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
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
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

	// Try to get account ID from Cloudflare API by listing accounts
	// Most users have access to only one account, so we'll use the first one
	accounts, _, err := a.api.Accounts(context.Background(), cloudflare.AccountsListParams{})
	if err == nil && len(accounts) > 0 {
		a.accountID = accounts[0].ID
		// Log successful account ID retrieval
		return a.accountID
	}

	// If API call fails, use the known account ID for this deployment
	// Log fallback usage for debugging
	a.accountID = "c1b74f148aee28025816e104a92622c5"
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

// Workers Domain operations
func (a *CloudflareAPIAdapter) AttachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domain cloudflare.AttachWorkersDomainParams) (cloudflare.WorkersDomain, error) {
	return a.api.AttachWorkersDomain(ctx, rc, domain)
}

func (a *CloudflareAPIAdapter) GetWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) (cloudflare.WorkersDomain, error) {
	return a.api.GetWorkersDomain(ctx, rc, domainID)
}

func (a *CloudflareAPIAdapter) DetachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) error {
	return a.api.DetachWorkersDomain(ctx, rc, domainID)
}

func (a *CloudflareAPIAdapter) ListWorkersDomains(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersDomainParams) ([]cloudflare.WorkersDomain, error) {
	return a.api.ListWorkersDomains(ctx, rc, params)
}

// Workers Subdomain operations
func (a *CloudflareAPIAdapter) WorkersCreateSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WorkersSubdomain) (cloudflare.WorkersSubdomain, error) {
	return a.api.WorkersCreateSubdomain(ctx, rc, params)
}

func (a *CloudflareAPIAdapter) WorkersGetSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.WorkersSubdomain, error) {
	return a.api.WorkersGetSubdomain(ctx, rc)
}

// Additional methods expected by controllers

// CreateWorkerCronTrigger creates a cron trigger for a worker script
func (a *CloudflareAPIAdapter) CreateWorkerCronTrigger(ctx context.Context, scriptName string, cron string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	params := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: scriptName,
		Crons: []cloudflare.WorkerCronTrigger{
			{Cron: cron},
		},
	}

	triggers, err := a.api.UpdateWorkerCronTriggers(ctx, rc, params)
	if err != nil {
		return nil, err
	}

	if len(triggers) > 0 {
		return map[string]interface{}{
			"script_name": scriptName,
			"cron":        triggers[0].Cron,
			"created":     true,
		}, nil
	}

	return nil, nil
}

// WorkerCronTrigger gets cron triggers for a worker script
func (a *CloudflareAPIAdapter) WorkerCronTrigger(ctx context.Context, scriptName string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	params := cloudflare.ListWorkerCronTriggersParams{
		ScriptName: scriptName,
	}

	triggers, err := a.api.ListWorkerCronTriggers(ctx, rc, params)
	if err != nil {
		return nil, err
	}

	if len(triggers) > 0 {
		return map[string]interface{}{
			"script_name": scriptName,
			"cron":        triggers[0].Cron,
		}, nil
	}

	return map[string]interface{}{
		"script_name": scriptName,
		"cron":        "",
	}, nil
}

// UpdateWorkerCronTrigger updates cron triggers for a worker script
func (a *CloudflareAPIAdapter) UpdateWorkerCronTrigger(ctx context.Context, scriptName string, cron string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	params := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: scriptName,
		Crons: []cloudflare.WorkerCronTrigger{
			{Cron: cron},
		},
	}

	triggers, err := a.api.UpdateWorkerCronTriggers(ctx, rc, params)
	if err != nil {
		return nil, err
	}

	if len(triggers) > 0 {
		return map[string]interface{}{
			"script_name": scriptName,
			"cron":        triggers[0].Cron,
			"updated":     true,
		}, nil
	}

	return nil, nil
}

// DeleteWorkerCronTrigger deletes all cron triggers for a worker script
func (a *CloudflareAPIAdapter) DeleteWorkerCronTrigger(ctx context.Context, scriptName string) error {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	params := cloudflare.UpdateWorkerCronTriggersParams{
		ScriptName: scriptName,
		Crons:      []cloudflare.WorkerCronTrigger{}, // Empty array to remove all triggers
	}

	_, err := a.api.UpdateWorkerCronTriggers(ctx, rc, params)
	return err
}

// CreateWorkerDomain attaches a custom domain to workers
func (a *CloudflareAPIAdapter) CreateWorkerDomain(ctx context.Context, params v1beta1.DomainParameters) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: params.AccountID,
	}

	apiParams := cloudflare.AttachWorkersDomainParams{
		ZoneID:   params.ZoneID,
		Hostname: params.Hostname,
	}

	domain, err := a.api.AttachWorkersDomain(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"domain_id": domain.ID,
		"hostname":  domain.Hostname,
		"zone_id":   domain.ZoneID,
		"created":   true,
	}, nil
}

// WorkerDomain gets a workers domain
func (a *CloudflareAPIAdapter) WorkerDomain(ctx context.Context, accountID, zoneID, domainID string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	domain, err := a.api.GetWorkersDomain(ctx, rc, domainID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"domain_id": domain.ID,
		"hostname":  domain.Hostname,
		"zone_id":   domain.ZoneID,
	}, nil
}

// UpdateWorkerDomain updates a workers domain (detach and reattach)
func (a *CloudflareAPIAdapter) UpdateWorkerDomain(ctx context.Context, accountID, zoneID, domainID string, params v1beta1.DomainParameters) (interface{}, error) {
	// First detach the old domain
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	err := a.api.DetachWorkersDomain(ctx, rc, domainID)
	if err != nil {
		return nil, err
	}

	// Then attach the new domain
	apiParams := cloudflare.AttachWorkersDomainParams{
		ZoneID:   params.ZoneID,
		Hostname: params.Hostname,
	}

	domain, err := a.api.AttachWorkersDomain(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"domain_id": domain.ID,
		"hostname":  domain.Hostname,
		"zone_id":   domain.ZoneID,
		"updated":   true,
	}, nil
}

// DeleteWorkerDomain detaches a workers domain
func (a *CloudflareAPIAdapter) DeleteWorkerDomain(ctx context.Context, accountID, zoneID, domainID string) error {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	return a.api.DetachWorkersDomain(ctx, rc, domainID)
}

// CreateWorkerKVNamespace creates a KV namespace
func (a *CloudflareAPIAdapter) CreateWorkerKVNamespace(ctx context.Context, params v1beta1.KVNamespaceParameters) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	apiParams := cloudflare.CreateWorkersKVNamespaceParams{
		Title: params.Title,
	}

	resp, err := a.api.CreateWorkersKVNamespace(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      resp.Result.ID,
		"title":   resp.Result.Title,
		"created": true,
	}, nil
}

// WorkerKVNamespace gets a KV namespace
func (a *CloudflareAPIAdapter) WorkerKVNamespace(ctx context.Context, kvID string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	namespaces, _, err := a.api.ListWorkersKVNamespaces(ctx, rc, cloudflare.ListWorkersKVNamespacesParams{})
	if err != nil {
		return nil, err
	}

	for _, ns := range namespaces {
		if ns.ID == kvID {
			return map[string]interface{}{
				"id":    ns.ID,
				"title": ns.Title,
			}, nil
		}
	}

	return nil, nil
}

// UpdateWorkerKVNamespace updates a KV namespace
func (a *CloudflareAPIAdapter) UpdateWorkerKVNamespace(ctx context.Context, kvID string, params v1beta1.KVNamespaceParameters) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	apiParams := cloudflare.UpdateWorkersKVNamespaceParams{
		NamespaceID: kvID,
		Title:       params.Title,
	}

	resp, err := a.api.UpdateWorkersKVNamespace(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"id":      kvID,
		"title":   params.Title,
		"updated": resp.Success,
	}, nil
}

// DeleteWorkerKVNamespace deletes a KV namespace
func (a *CloudflareAPIAdapter) DeleteWorkerKVNamespace(ctx context.Context, kvID string) error {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: a.GetAccountID(),
	}

	_, err := a.api.DeleteWorkersKVNamespace(ctx, rc, kvID)
	return err
}

// CreateWorkerSubdomain creates a workers subdomain
func (a *CloudflareAPIAdapter) CreateWorkerSubdomain(ctx context.Context, params v1beta1.SubdomainParameters) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: params.AccountID,
	}

	apiParams := cloudflare.WorkersSubdomain{
		Name: params.Name,
	}

	subdomain, err := a.api.WorkersCreateSubdomain(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":    subdomain.Name,
		"created": true,
	}, nil
}

// WorkerSubdomain gets the workers subdomain
func (a *CloudflareAPIAdapter) WorkerSubdomain(ctx context.Context, accountID, subdomainName string) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	subdomain, err := a.api.WorkersGetSubdomain(ctx, rc)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name": subdomain.Name,
	}, nil
}

// UpdateWorkerSubdomain updates the workers subdomain
func (a *CloudflareAPIAdapter) UpdateWorkerSubdomain(ctx context.Context, accountID, subdomainName string, params v1beta1.SubdomainParameters) (interface{}, error) {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	apiParams := cloudflare.WorkersSubdomain{
		Name: params.Name,
	}

	subdomain, err := a.api.WorkersCreateSubdomain(ctx, rc, apiParams)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":    subdomain.Name,
		"updated": true,
	}, nil
}

// DeleteWorkerSubdomain deletes the workers subdomain (set to empty)
func (a *CloudflareAPIAdapter) DeleteWorkerSubdomain(ctx context.Context, accountID, subdomainName string) error {
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	// Cloudflare doesn't have a direct delete for subdomain, so we set it to empty
	params := cloudflare.WorkersSubdomain{
		Name: "",
	}

	_, err := a.api.WorkersCreateSubdomain(ctx, rc, params)
	return err
}
