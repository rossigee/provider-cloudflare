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

type CloudflarePagesDomainClient interface {
	GetPagesDomains(ctx context.Context, params cloudflare.PagesDomainsParameters) ([]cloudflare.PagesDomain, error)
	GetPagesDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error)
	PagesPatchDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error)
	PagesAddDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error)
	PagesDeleteDomain(ctx context.Context, params cloudflare.PagesDomainParameters) error
}

type CloudflarePagesDomainClientImpl struct {
	api *cloudflare.API
}

func NewDomainClientFromAPI(api *cloudflare.API) *CloudflarePagesDomainClientImpl {
	return &CloudflarePagesDomainClientImpl{api: api}
}

func (c *CloudflarePagesDomainClientImpl) GetPagesDomains(ctx context.Context, params cloudflare.PagesDomainsParameters) ([]cloudflare.PagesDomain, error) {
	return c.api.GetPagesDomains(ctx, params)
}

func (c *CloudflarePagesDomainClientImpl) GetPagesDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	return c.api.GetPagesDomain(ctx, params)
}

func (c *CloudflarePagesDomainClientImpl) PagesPatchDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	return c.api.PagesPatchDomain(ctx, params)
}

func (c *CloudflarePagesDomainClientImpl) PagesAddDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	return c.api.PagesAddDomain(ctx, params)
}

func (c *CloudflarePagesDomainClientImpl) PagesDeleteDomain(ctx context.Context, params cloudflare.PagesDomainParameters) error {
	return c.api.PagesDeleteDomain(ctx, params)
}
