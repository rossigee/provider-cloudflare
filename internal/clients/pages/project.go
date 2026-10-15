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

type CloudflarePagesProjectClient interface {
	CreatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesProjectParams) (cloudflare.PagesProject, error)
	GetPagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) (cloudflare.PagesProject, error)
	UpdatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdatePagesProjectParams) (cloudflare.PagesProject, error)
	DeletePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) error
	ListPagesProjects(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesProjectsParams) ([]cloudflare.PagesProject, cloudflare.ResultInfo, error)
}

type CloudflarePagesProjectClientImpl struct {
	api *cloudflare.API
}

func NewClientFromAPI(api *cloudflare.API) *CloudflarePagesProjectClientImpl {
	return &CloudflarePagesProjectClientImpl{api: api}
}

func (c *CloudflarePagesProjectClientImpl) CreatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesProjectParams) (cloudflare.PagesProject, error) {
	return c.api.CreatePagesProject(ctx, rc, params)
}

func (c *CloudflarePagesProjectClientImpl) GetPagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) (cloudflare.PagesProject, error) {
	return c.api.GetPagesProject(ctx, rc, projectName)
}

func (c *CloudflarePagesProjectClientImpl) UpdatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdatePagesProjectParams) (cloudflare.PagesProject, error) {
	return c.api.UpdatePagesProject(ctx, rc, params)
}

func (c *CloudflarePagesProjectClientImpl) DeletePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) error {
	return c.api.DeletePagesProject(ctx, rc, projectName)
}

func (c *CloudflarePagesProjectClientImpl) ListPagesProjects(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesProjectsParams) ([]cloudflare.PagesProject, cloudflare.ResultInfo, error) {
	return c.api.ListPagesProjects(ctx, rc, params)
}
