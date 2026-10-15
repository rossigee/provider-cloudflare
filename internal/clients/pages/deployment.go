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

package pages

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflarePagesDeploymentClient interface {
	CreatePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesDeploymentParams) (cloudflare.PagesProjectDeployment, error)
	GetPagesDeploymentInfo(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error)
	DeletePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeletePagesDeploymentParams) error
	ListPagesDeployments(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesDeploymentsParams) ([]cloudflare.PagesProjectDeployment, *cloudflare.ResultInfo, error)
	RetryPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error)
	RollbackPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error)
	GetPagesDeploymentLogs(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetPagesDeploymentLogsParams) (cloudflare.PagesDeploymentLogs, error)
}

type CloudflarePagesDeploymentClientImpl struct {
	api *cloudflare.API
}

func NewDeploymentClientFromAPI(api *cloudflare.API) *CloudflarePagesDeploymentClientImpl {
	return &CloudflarePagesDeploymentClientImpl{api: api}
}

func (c *CloudflarePagesDeploymentClientImpl) CreatePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesDeploymentParams) (cloudflare.PagesProjectDeployment, error) {
	return c.api.CreatePagesDeployment(ctx, rc, params)
}

func (c *CloudflarePagesDeploymentClientImpl) GetPagesDeploymentInfo(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	return c.api.GetPagesDeploymentInfo(ctx, rc, projectName, deploymentID)
}

func (c *CloudflarePagesDeploymentClientImpl) DeletePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeletePagesDeploymentParams) error {
	return c.api.DeletePagesDeployment(ctx, rc, params)
}

func (c *CloudflarePagesDeploymentClientImpl) ListPagesDeployments(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesDeploymentsParams) ([]cloudflare.PagesProjectDeployment, *cloudflare.ResultInfo, error) {
	return c.api.ListPagesDeployments(ctx, rc, params)
}

func (c *CloudflarePagesDeploymentClientImpl) RetryPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	return c.api.RetryPagesDeployment(ctx, rc, projectName, deploymentID)
}

func (c *CloudflarePagesDeploymentClientImpl) RollbackPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	return c.api.RollbackPagesDeployment(ctx, rc, projectName, deploymentID)
}

func (c *CloudflarePagesDeploymentClientImpl) GetPagesDeploymentLogs(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetPagesDeploymentLogsParams) (cloudflare.PagesDeploymentLogs, error) {
	return c.api.GetPagesDeploymentLogs(ctx, rc, params)
}
