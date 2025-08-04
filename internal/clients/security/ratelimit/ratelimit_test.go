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

package ratelimit

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/security/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// MockRateLimitAPI implements the RateLimitAPI interface for testing
type MockRateLimitAPI struct {
	MockRateLimit       func(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error)
	MockCreateRateLimit func(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	MockUpdateRateLimit func(ctx context.Context, zoneID, limitID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	MockDeleteRateLimit func(ctx context.Context, zoneID, limitID string) error
}

func (m *MockRateLimitAPI) RateLimit(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error) {
	if m.MockRateLimit != nil {
		return m.MockRateLimit(ctx, zoneID, limitID)
	}
	return cloudflare.RateLimit{}, nil
}

func (m *MockRateLimitAPI) CreateRateLimit(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
	if m.MockCreateRateLimit != nil {
		return m.MockCreateRateLimit(ctx, zoneID, limit)
	}
	return limit, nil
}

func (m *MockRateLimitAPI) UpdateRateLimit(ctx context.Context, zoneID, limitID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
	if m.MockUpdateRateLimit != nil {
		return m.MockUpdateRateLimit(ctx, zoneID, limitID, limit)
	}
	return limit, nil
}

func (m *MockRateLimitAPI) DeleteRateLimit(ctx context.Context, zoneID, limitID string) error {
	if m.MockDeleteRateLimit != nil {
		return m.MockDeleteRateLimit(ctx, zoneID, limitID)
	}
	return nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	rateLimitID := "test-rate-limit-id"

	type fields struct {
		client *MockRateLimitAPI
	}

	type args struct {
		ctx         context.Context
		zoneID      string
		rateLimitID string
	}

	type want struct {
		obs *v1alpha1.RateLimitObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetRateLimitSuccess": {
			reason: "Get should return Rate Limit when API call succeeds",
			fields: fields{
				client: &MockRateLimitAPI{
					MockRateLimit: func(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.RateLimit{}, errors.New("wrong zone ID")
						}
						if limitID != "test-rate-limit-id" {
							return cloudflare.RateLimit{}, errors.New("wrong rate limit ID")
						}
						return cloudflare.RateLimit{
							ID:          "test-rate-limit-id",
							Disabled:    false,
							Description: "Test rate limit",
							Threshold:   100,
							Period:      60,
							Match: cloudflare.RateLimitTrafficMatcher{
								Request: cloudflare.RateLimitRequestMatcher{
									Methods:    []string{"GET", "POST"},
									Schemes:    []string{"HTTP", "HTTPS"},
									URLPattern: "*.example.com/*",
								},
							},
							Action: cloudflare.RateLimitAction{
								Mode:    "simulate",
								Timeout: 86400,
							},
						}, nil
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				obs: &v1alpha1.RateLimitObservation{
					ID:          "test-rate-limit-id",
					Disabled:    false,
					Description: "Test rate limit",
					Threshold:   100,
					Period:      60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods:    []string{"GET", "POST"},
							Schemes:    []string{"HTTP", "HTTPS"},
							URLPattern: ptr.To("*.example.com/*"),
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "simulate",
						Timeout: ptr.To(86400),
					},
				},
				err: nil,
			},
		},
		"GetRateLimitWithComplexMatch": {
			reason: "Get should return Rate Limit with complex matching rules",
			fields: fields{
				client: &MockRateLimitAPI{
					MockRateLimit: func(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error) {
						return cloudflare.RateLimit{
							ID:          "complex-rate-limit-id",
							Disabled:    true,
							Description: "Complex rate limit",
							Threshold:   50,
							Period:      300,
							Match: cloudflare.RateLimitTrafficMatcher{
								Request: cloudflare.RateLimitRequestMatcher{
									Methods:    []string{"POST"},
									Schemes:    []string{"HTTPS"},
									URLPattern: "/api/*",
								},
								Response: cloudflare.RateLimitResponseMatcher{
									Statuses:      []int{429, 503},
									OriginTraffic: ptr.To(true),
									Headers: []cloudflare.RateLimitResponseMatcherHeader{
										{
											Name:  "X-Rate-Limit",
											Op:    "eq",
											Value: "exceeded",
										},
									},
								},
							},
							Action: cloudflare.RateLimitAction{
								Mode:    "ban",
								Timeout: 3600,
								Response: &cloudflare.RateLimitActionResponse{
									ContentType: "application/json",
									Body:        `{"error": "rate limited"}`,
								},
							},
							Bypass: []cloudflare.RateLimitKeyValue{
								{Name: "url", Value: "api.example.com/*"},
							},
							Correlate: &cloudflare.RateLimitCorrelate{
								By: "nat",
							},
						}, nil
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: "complex-rate-limit-id",
			},
			want: want{
				obs: &v1alpha1.RateLimitObservation{
					ID:          "complex-rate-limit-id",
					Disabled:    true,
					Description: "Complex rate limit",
					Threshold:   50,
					Period:      300,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods:    []string{"POST"},
							Schemes:    []string{"HTTPS"},
							URLPattern: ptr.To("/api/*"),
						},
						Response: &v1alpha1.RateLimitResponseMatcher{
							Statuses:      []int{429, 503},
							OriginTraffic: ptr.To(true),
							Headers: []v1alpha1.RateLimitResponseMatcherHeader{
								{
									Name:  "X-Rate-Limit",
									Op:    "eq",
									Value: "exceeded",
								},
							},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "ban",
						Timeout: ptr.To(3600),
						Response: &v1alpha1.RateLimitActionResponse{
							ContentType: "application/json",
							Body:        `{"error": "rate limited"}`,
						},
					},
					Bypass: []v1alpha1.RateLimitKeyValue{
						{Name: "url", Value: "api.example.com/*"},
					},
					Correlate: &v1alpha1.RateLimitCorrelate{
						By: "nat",
					},
				},
				err: nil,
			},
		},
		"GetRateLimitNotFound": {
			reason: "Get should return NotFoundError when Rate Limit is not found",
			fields: fields{
				client: &MockRateLimitAPI{
					MockRateLimit: func(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error) {
						return cloudflare.RateLimit{}, errors.New("rate limit not found")
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("rate limit not found"),
			},
		},
		"GetRateLimitAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockRateLimitAPI{
					MockRateLimit: func(ctx context.Context, zoneID, limitID string) (cloudflare.RateLimit, error) {
						return cloudflare.RateLimit{}, errBoom
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get rate limit"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.zoneID, tc.args.rateLimitID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGet(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nGet(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockRateLimitAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.RateLimitParameters
	}

	type want struct {
		obs *v1alpha1.RateLimitObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateRateLimitSuccess": {
			reason: "Create should create Rate Limit when API call succeeds",
			fields: fields{
				client: &MockRateLimitAPI{
					MockCreateRateLimit: func(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.RateLimit{}, errors.New("wrong zone ID")
						}
						if limit.Threshold != 100 {
							return cloudflare.RateLimit{}, errors.New("wrong threshold")
						}
						if limit.Period != 60 {
							return cloudflare.RateLimit{}, errors.New("wrong period")
						}
						// Return the created rate limit with an ID
						limit.ID = "new-rate-limit-id"
						return limit, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:        zoneID,
					Disabled:    ptr.To(false),
					Description: ptr.To("Test rate limit"),
					Threshold:   100,
					Period:      60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET", "POST"},
							Schemes: []string{"HTTP", "HTTPS"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "simulate",
						Timeout: ptr.To(86400),
					},
				},
			},
			want: want{
				obs: &v1alpha1.RateLimitObservation{
					ID:          "new-rate-limit-id",
					Disabled:    false,
					Description: "Test rate limit",
					Threshold:   100,
					Period:      60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET", "POST"},
							Schemes: []string{"HTTP", "HTTPS"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "simulate",
						Timeout: ptr.To(86400),
					},
				},
				err: nil,
			},
		},
		"CreateRateLimitMinimal": {
			reason: "Create should create Rate Limit with minimal parameters",
			fields: fields{
				client: &MockRateLimitAPI{
					MockCreateRateLimit: func(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
						limit.ID = "minimal-rate-limit-id"
						return limit, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:      zoneID,
					Threshold: 10,
					Period:    300,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "block",
					},
				},
			},
			want: want{
				obs: &v1alpha1.RateLimitObservation{
					ID:        "minimal-rate-limit-id",
					Disabled:  false,
					Threshold: 10,
					Period:    300,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "block",
					},
				},
				err: nil,
			},
		},
		"CreateRateLimitAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockRateLimitAPI{
					MockCreateRateLimit: func(ctx context.Context, zoneID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
						return cloudflare.RateLimit{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:      zoneID,
					Threshold: 100,
					Period:    60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot create rate limit"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Create(tc.args.ctx, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	rateLimitID := "test-rate-limit-id"

	type fields struct {
		client *MockRateLimitAPI
	}

	type args struct {
		ctx         context.Context
		rateLimitID string
		params      v1alpha1.RateLimitParameters
	}

	type want struct {
		obs *v1alpha1.RateLimitObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateRateLimitSuccess": {
			reason: "Update should update Rate Limit when API call succeeds",
			fields: fields{
				client: &MockRateLimitAPI{
					MockUpdateRateLimit: func(ctx context.Context, zoneID, limitID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.RateLimit{}, errors.New("wrong zone ID")
						}
						if limitID != "test-rate-limit-id" {
							return cloudflare.RateLimit{}, errors.New("wrong rate limit ID")
						}
						if limit.Threshold != 200 {
							return cloudflare.RateLimit{}, errors.New("wrong threshold")
						}
						return limit, nil
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				rateLimitID: rateLimitID,
				params: v1alpha1.RateLimitParameters{
					Zone:        zoneID,
					Disabled:    ptr.To(true),
					Description: ptr.To("Updated rate limit"),
					Threshold:   200,
					Period:      120,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "block",
					},
				},
			},
			want: want{
				obs: &v1alpha1.RateLimitObservation{
					ID:          "test-rate-limit-id",
					Disabled:    true,
					Description: "Updated rate limit",
					Threshold:   200,
					Period:      120,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "block",
					},
				},
				err: nil,
			},
		},
		"UpdateRateLimitAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockRateLimitAPI{
					MockUpdateRateLimit: func(ctx context.Context, zoneID, limitID string, limit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
						return cloudflare.RateLimit{}, errBoom
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				rateLimitID: rateLimitID,
				params: v1alpha1.RateLimitParameters{
					Zone:      zoneID,
					Threshold: 100,
					Period:    60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot update rate limit"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.rateLimitID, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	rateLimitID := "test-rate-limit-id"

	type fields struct {
		client *MockRateLimitAPI
	}

	type args struct {
		ctx         context.Context
		zoneID      string
		rateLimitID string
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"DeleteRateLimitSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockRateLimitAPI{
					MockDeleteRateLimit: func(ctx context.Context, zoneID, limitID string) error {
						if zoneID != "test-zone-id" {
							return errors.New("wrong zone ID")
						}
						if limitID != "test-rate-limit-id" {
							return errors.New("wrong rate limit ID")
						}
						return nil
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteRateLimitNotFound": {
			reason: "Delete should succeed when Rate Limit is not found (already deleted)",
			fields: fields{
				client: &MockRateLimitAPI{
					MockDeleteRateLimit: func(ctx context.Context, zoneID, limitID string) error {
						return errors.New("rate limit not found")
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteRateLimitAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockRateLimitAPI{
					MockDeleteRateLimit: func(ctx context.Context, zoneID, limitID string) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:         context.Background(),
				zoneID:      zoneID,
				rateLimitID: rateLimitID,
			},
			want: want{
				err: errors.Wrap(errBoom, "cannot delete rate limit"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.zoneID, tc.args.rateLimitID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	zoneID := "test-zone-id"

	type fields struct {
		client *MockRateLimitAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.RateLimitParameters
		obs    v1alpha1.RateLimitObservation
	}

	type want struct {
		upToDate bool
		err      error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"IsUpToDateTrue": {
			reason: "IsUpToDate should return true when all settings match",
			fields: fields{
				client: &MockRateLimitAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:        zoneID,
					Disabled:    ptr.To(false),
					Description: ptr.To("Test rate limit"),
					Threshold:   100,
					Period:      60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET", "POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "simulate",
						Timeout: ptr.To(86400),
					},
				},
				obs: v1alpha1.RateLimitObservation{
					ID:          "test-id",
					Disabled:    false,
					Description: "Test rate limit",
					Threshold:   100,
					Period:      60,
					Match: v1alpha1.RateLimitTrafficMatcher{
						Request: v1alpha1.RateLimitRequestMatcher{
							Methods: []string{"GET", "POST"},
						},
					},
					Action: v1alpha1.RateLimitAction{
						Mode:    "simulate",
						Timeout: ptr.To(86400),
					},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseDisabled": {
			reason: "IsUpToDate should return false when disabled setting doesn't match",
			fields: fields{
				client: &MockRateLimitAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:     zoneID,
					Disabled: ptr.To(true),
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
				obs: v1alpha1.RateLimitObservation{
					Disabled: false,
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseThreshold": {
			reason: "IsUpToDate should return false when threshold doesn't match",
			fields: fields{
				client: &MockRateLimitAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone:      zoneID,
					Threshold: 100,
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
				obs: v1alpha1.RateLimitObservation{
					Threshold: 50,
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseActionMode": {
			reason: "IsUpToDate should return false when action mode doesn't match",
			fields: fields{
				client: &MockRateLimitAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.RateLimitParameters{
					Zone: zoneID,
					Action: v1alpha1.RateLimitAction{
						Mode: "block",
					},
				},
				obs: v1alpha1.RateLimitObservation{
					Action: v1alpha1.RateLimitAction{
						Mode: "simulate",
					},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.IsUpToDate(tc.args.ctx, tc.args.params, tc.args.obs)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	type args struct {
		err error
	}

	type want struct {
		isNotFound bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NilError": {
			reason: "isNotFound should return false for nil error",
			args: args{
				err: nil,
			},
			want: want{
				isNotFound: false,
			},
		},
		"NotFoundError": {
			reason: "isNotFound should return true for 'not found' error",
			args: args{
				err: errors.New("not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"ResourceNotFoundError": {
			reason: "isNotFound should return true for 'resource not found' error",
			args: args{
				err: errors.New("resource not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"RateLimitNotFoundError": {
			reason: "isNotFound should return true for 'rate limit not found' error",
			args: args{
				err: errors.New("rate limit not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"CaseInsensitiveError": {
			reason: "isNotFound should be case insensitive",
			args: args{
				err: errors.New("RATE LIMIT NOT FOUND"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "isNotFound should return false for other errors",
			args: args{
				err: errors.New("some other error"),
			},
			want: want{
				isNotFound: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := isNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.isNotFound, got); diff != "" {
				t.Errorf("\n%s\nisNotFound(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}