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

package kvnamespace

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateKVNamespace = "cannot create workers kv namespace"
	errUpdateKVNamespace = "cannot update workers kv namespace"
	errGetKVNamespace    = "cannot get workers kv namespace"
	errDeleteKVNamespace = "cannot delete workers kv namespace"
	errListKVNamespaces  = "cannot list workers kv namespaces"
)

// KVNamespaceClient provides operations for Workers KV Namespaces.
type KVNamespaceClient struct {
	client    clients.ClientInterface
	accountID string
}

// NewClient creates a new Workers KV Namespace client.
func NewClient(client clients.ClientInterface) *KVNamespaceClient {
	return &KVNamespaceClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
	}
}

// getAccountID gets the account ID from the Cloudflare API
func (c *KVNamespaceClient) getAccountID(ctx context.Context) (string, error) {
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

// convertToObservation converts cloudflare-go KV namespace to Crossplane observation.
func convertToObservation(namespace cloudflare.WorkersKVNamespace) v1alpha1.KVNamespaceObservation {
	return v1alpha1.KVNamespaceObservation{
		ID:    namespace.ID,
		Title: namespace.Title,
	}
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.KVNamespaceParameters) cloudflare.CreateWorkersKVNamespaceParams {
	return cloudflare.CreateWorkersKVNamespaceParams{
		Title: params.Title,
	}
}

// convertToCloudflareUpdateParams converts Crossplane parameters to cloudflare-go update parameters.
func convertToCloudflareUpdateParams(namespaceID string, params v1alpha1.KVNamespaceParameters) cloudflare.UpdateWorkersKVNamespaceParams {
	return cloudflare.UpdateWorkersKVNamespaceParams{
		NamespaceID: namespaceID,
		Title:       params.Title,
	}
}

// Create creates a new Workers KV Namespace.
func (c *KVNamespaceClient) Create(ctx context.Context, params v1alpha1.KVNamespaceParameters) (*v1alpha1.KVNamespaceObservation, error) {
	createParams := convertToCloudflareParams(params)
	
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	resp, err := c.client.CreateWorkersKVNamespace(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateKVNamespace)
	}

	obs := convertToObservation(resp.Result)
	return &obs, nil
}

// Get retrieves a Workers KV Namespace by finding it in the list.
// Note: Cloudflare API doesn't provide a direct get by ID for KV namespaces,
// so we need to list all namespaces and find the one with matching ID.
func (c *KVNamespaceClient) Get(ctx context.Context, namespaceID string) (*v1alpha1.KVNamespaceObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	listParams := cloudflare.ListWorkersKVNamespacesParams{}
	
	namespaces, _, err := c.client.ListWorkersKVNamespaces(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errGetKVNamespace)
	}

	for _, namespace := range namespaces {
		if namespace.ID == namespaceID {
			obs := convertToObservation(namespace)
			return &obs, nil
		}
	}

	return nil, clients.NewNotFoundError("kv namespace not found")
}

// Update updates an existing Workers KV Namespace.
func (c *KVNamespaceClient) Update(ctx context.Context, namespaceID string, params v1alpha1.KVNamespaceParameters) (*v1alpha1.KVNamespaceObservation, error) {
	updateParams := convertToCloudflareUpdateParams(namespaceID, params)
	
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	_, err = c.client.UpdateWorkersKVNamespace(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateKVNamespace)
	}

	// After update, get the current state
	obs, err := c.Get(ctx, namespaceID)
	if err != nil {
		return nil, errors.Wrap(err, errGetKVNamespace)
	}

	return obs, nil
}

// Delete removes a Workers KV Namespace.
func (c *KVNamespaceClient) Delete(ctx context.Context, namespaceID string) error {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	_, err = c.client.DeleteWorkersKVNamespace(ctx, rc, namespaceID)
	if err != nil {
		return errors.Wrap(err, errDeleteKVNamespace)
	}

	return nil
}

// List retrieves all Workers KV Namespaces.
func (c *KVNamespaceClient) List(ctx context.Context) ([]v1alpha1.KVNamespaceObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	listParams := cloudflare.ListWorkersKVNamespacesParams{}
	
	namespaces, _, err := c.client.ListWorkersKVNamespaces(ctx, rc, listParams)
	if err != nil {
		return nil, errors.Wrap(err, errListKVNamespaces)
	}

	observations := make([]v1alpha1.KVNamespaceObservation, len(namespaces))
	for i, namespace := range namespaces {
		observations[i] = convertToObservation(namespace)
	}

	return observations, nil
}

// IsUpToDate checks if the Workers KV Namespace is up to date.
func (c *KVNamespaceClient) IsUpToDate(ctx context.Context, params v1alpha1.KVNamespaceParameters, obs v1alpha1.KVNamespaceObservation) (bool, error) {
	// For KV namespaces, we only need to compare the title
	return obs.Title == params.Title, nil
}