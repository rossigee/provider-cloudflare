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

package pagerules

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type CloudflarePageRuleClient interface {
	CreatePageRule(ctx context.Context, zoneID string, rule cloudflare.PageRule) (*cloudflare.PageRule, error)
	GetPageRule(ctx context.Context, zoneID, ruleID string) (cloudflare.PageRule, error)
	UpdatePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error
	DeletePageRule(ctx context.Context, zoneID, ruleID string) error
	ListPageRules(ctx context.Context, zoneID string) ([]cloudflare.PageRule, error)
	ChangePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error
}

type CloudflarePageRuleClientImpl struct {
	api *cloudflare.API
}

func NewClientFromAPI(api *cloudflare.API) *CloudflarePageRuleClientImpl {
	return &CloudflarePageRuleClientImpl{api: api}
}

func (c *CloudflarePageRuleClientImpl) CreatePageRule(ctx context.Context, zoneID string, rule cloudflare.PageRule) (*cloudflare.PageRule, error) {
	return c.api.CreatePageRule(ctx, zoneID, rule)
}

func (c *CloudflarePageRuleClientImpl) GetPageRule(ctx context.Context, zoneID, ruleID string) (cloudflare.PageRule, error) {
	return c.api.PageRule(ctx, zoneID, ruleID)
}

func (c *CloudflarePageRuleClientImpl) UpdatePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error {
	return c.api.UpdatePageRule(ctx, zoneID, ruleID, rule)
}

func (c *CloudflarePageRuleClientImpl) DeletePageRule(ctx context.Context, zoneID, ruleID string) error {
	return c.api.DeletePageRule(ctx, zoneID, ruleID)
}

func (c *CloudflarePageRuleClientImpl) ListPageRules(ctx context.Context, zoneID string) ([]cloudflare.PageRule, error) {
	return c.api.ListPageRules(ctx, zoneID)
}

func (c *CloudflarePageRuleClientImpl) ChangePageRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.PageRule) error {
	return c.api.ChangePageRule(ctx, zoneID, ruleID, rule)
}
