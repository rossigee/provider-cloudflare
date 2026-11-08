/*
Copyright 2026 The Crossplane Authors.

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

// Package aigateway wraps the Cloudflare AI Gateway API used by the Gateway resource.
// It uses cloudflare-go v7, which is separate from the v0 SDK used by the other clients.
package aigateway

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/ai_gateway"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/aigateway/v1beta1"
)

const (
	errCreate = "cannot create AI Gateway"
	errGet    = "cannot get AI Gateway"
	errUpdate = "cannot update AI Gateway"
	errDelete = "cannot delete AI Gateway"
)

// Gateway is an AI Gateway as returned by the Cloudflare API.
type Gateway struct {
	ID                      string
	CacheInvalidateOnUpdate bool
	CacheTTL                int64
	CollectLogs             bool
	RateLimitingInterval    int64
	RateLimitingLimit       int64
	RateLimitingTechnique   string
	Authentication          bool
	CreatedAt               time.Time
	ModifiedAt              time.Time
}

// Client manages AI Gateways in a Cloudflare account.
type Client interface {
	Create(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (Gateway, error)
	Get(ctx context.Context, accountID, gatewayID string) (Gateway, error)
	Update(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (Gateway, error)
	Delete(ctx context.Context, accountID, gatewayID string) error
}

type sdkClient struct {
	sdk *cloudflare.Client
}

// NewClient returns a Client that authenticates with a Cloudflare API token.
func NewClient(token string) Client {
	return newSDKClient(option.WithAPIToken(token))
}

func newSDKClient(opts ...option.RequestOption) Client {
	return &sdkClient{sdk: cloudflare.NewClient(opts...)}
}

// Create creates an AI Gateway.
func (c *sdkClient) Create(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (Gateway, error) {
	params := ai_gateway.AIGatewayNewParams{
		AccountID:               cloudflare.F(accountID),
		ID:                      cloudflare.F(p.GatewayID),
		CacheInvalidateOnUpdate: cloudflare.F(p.CacheInvalidateOnUpdate),
		CacheTTL:                cloudflare.F(p.CacheTTL),
		CollectLogs:             cloudflare.F(p.CollectLogs),
		RateLimitingInterval:    cloudflare.F(p.RateLimitingInterval),
		RateLimitingLimit:       cloudflare.F(p.RateLimitingLimit),
	}
	if p.Authentication != nil {
		params.Authentication = cloudflare.F(*p.Authentication)
	}
	if p.RateLimitingTechnique != nil {
		params.RateLimitingTechnique = cloudflare.F(ai_gateway.AIGatewayNewParamsRateLimitingTechnique(*p.RateLimitingTechnique))
	}

	res, err := c.sdk.AIGateway.New(ctx, params)
	if err != nil {
		return Gateway{}, errors.Wrap(err, errCreate)
	}
	return Gateway{
		ID:                      res.ID,
		CacheInvalidateOnUpdate: res.CacheInvalidateOnUpdate,
		CacheTTL:                res.CacheTTL,
		CollectLogs:             res.CollectLogs,
		RateLimitingInterval:    res.RateLimitingInterval,
		RateLimitingLimit:       res.RateLimitingLimit,
		RateLimitingTechnique:   string(res.RateLimitingTechnique),
		Authentication:          res.Authentication,
		CreatedAt:               res.CreatedAt,
		ModifiedAt:              res.ModifiedAt,
	}, nil
}

// Get returns an AI Gateway. A missing gateway is reported by IsNotFound.
func (c *sdkClient) Get(ctx context.Context, accountID, gatewayID string) (Gateway, error) {
	res, err := c.sdk.AIGateway.Get(ctx, gatewayID, ai_gateway.AIGatewayGetParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return Gateway{}, errors.Wrap(err, errGet)
	}
	return Gateway{
		ID:                      res.ID,
		CacheInvalidateOnUpdate: res.CacheInvalidateOnUpdate,
		CacheTTL:                res.CacheTTL,
		CollectLogs:             res.CollectLogs,
		RateLimitingInterval:    res.RateLimitingInterval,
		RateLimitingLimit:       res.RateLimitingLimit,
		RateLimitingTechnique:   string(res.RateLimitingTechnique),
		Authentication:          res.Authentication,
		CreatedAt:               res.CreatedAt,
		ModifiedAt:              res.ModifiedAt,
	}, nil
}

// Update updates the settings of an AI Gateway. Optional settings that are
// not set in p are left unchanged.
func (c *sdkClient) Update(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (Gateway, error) {
	params := ai_gateway.AIGatewayUpdateParams{
		AccountID:               cloudflare.F(accountID),
		CacheInvalidateOnUpdate: cloudflare.F(p.CacheInvalidateOnUpdate),
		CacheTTL:                cloudflare.F(p.CacheTTL),
		CollectLogs:             cloudflare.F(p.CollectLogs),
		RateLimitingInterval:    cloudflare.F(p.RateLimitingInterval),
		RateLimitingLimit:       cloudflare.F(p.RateLimitingLimit),
	}
	if p.Authentication != nil {
		params.Authentication = cloudflare.F(*p.Authentication)
	}
	if p.RateLimitingTechnique != nil {
		params.RateLimitingTechnique = cloudflare.F(ai_gateway.AIGatewayUpdateParamsRateLimitingTechnique(*p.RateLimitingTechnique))
	}

	res, err := c.sdk.AIGateway.Update(ctx, gatewayID, params)
	if err != nil {
		return Gateway{}, errors.Wrap(err, errUpdate)
	}
	return Gateway{
		ID:                      res.ID,
		CacheInvalidateOnUpdate: res.CacheInvalidateOnUpdate,
		CacheTTL:                res.CacheTTL,
		CollectLogs:             res.CollectLogs,
		RateLimitingInterval:    res.RateLimitingInterval,
		RateLimitingLimit:       res.RateLimitingLimit,
		RateLimitingTechnique:   string(res.RateLimitingTechnique),
		Authentication:          res.Authentication,
		CreatedAt:               res.CreatedAt,
		ModifiedAt:              res.ModifiedAt,
	}, nil
}

// Delete deletes an AI Gateway.
func (c *sdkClient) Delete(ctx context.Context, accountID, gatewayID string) error {
	_, err := c.sdk.AIGateway.Delete(ctx, gatewayID, ai_gateway.AIGatewayDeleteParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return errors.Wrap(err, errDelete)
	}
	return nil
}

// IsNotFound reports whether err is a Cloudflare API "not found" response.
func IsNotFound(err error) bool {
	var apiErr *cloudflare.Error
	return stderrors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// IsUpToDate reports whether the observed gateway matches the desired parameters.
// Optional parameters that are not set are not compared.
func IsUpToDate(p v1beta1.GatewayParameters, g Gateway) bool {
	if p.CacheInvalidateOnUpdate != g.CacheInvalidateOnUpdate ||
		p.CacheTTL != g.CacheTTL ||
		p.CollectLogs != g.CollectLogs ||
		p.RateLimitingInterval != g.RateLimitingInterval ||
		p.RateLimitingLimit != g.RateLimitingLimit {
		return false
	}
	if p.RateLimitingTechnique != nil && *p.RateLimitingTechnique != g.RateLimitingTechnique {
		return false
	}
	if p.Authentication != nil && *p.Authentication != g.Authentication {
		return false
	}
	return true
}
