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

package turnstile

import (
	"context"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/security/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// MockTurnstileAPI implements the TurnstileAPI interface for testing
type MockTurnstileAPI struct {
	MockCreateTurnstileWidget func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	MockGetTurnstileWidget    func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error)
	MockUpdateTurnstileWidget func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	MockDeleteTurnstileWidget func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error
}

func (m *MockTurnstileAPI) CreateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
	if m.MockCreateTurnstileWidget != nil {
		return m.MockCreateTurnstileWidget(ctx, rc, params)
	}
	return cloudflare.TurnstileWidget{}, nil
}

func (m *MockTurnstileAPI) GetTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error) {
	if m.MockGetTurnstileWidget != nil {
		return m.MockGetTurnstileWidget(ctx, rc, siteKey)
	}
	return cloudflare.TurnstileWidget{}, nil
}

func (m *MockTurnstileAPI) UpdateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
	if m.MockUpdateTurnstileWidget != nil {
		return m.MockUpdateTurnstileWidget(ctx, rc, params)
	}
	return cloudflare.TurnstileWidget{}, nil
}

func (m *MockTurnstileAPI) DeleteTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error {
	if m.MockDeleteTurnstileWidget != nil {
		return m.MockDeleteTurnstileWidget(ctx, rc, siteKey)
	}
	return nil
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")
	accountID := "test-account-id"

	type fields struct {
		client *MockTurnstileAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.TurnstileParameters
	}

	type want struct {
		obs *v1alpha1.TurnstileObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateTurnstileSuccess": {
			reason: "Create should create Turnstile widget when API call succeeds",
			fields: fields{
				client: &MockTurnstileAPI{
					MockCreateTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.TurnstileWidget{}, errors.New("wrong resource type")
						}
						if params.Name != "Test Widget" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong name")
						}
						if len(params.Domains) != 2 || params.Domains[0] != "example.com" || params.Domains[1] != "*.example.com" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong domains")
						}
						createdTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						return cloudflare.TurnstileWidget{
							SiteKey:      "0x4AAAAAAABnPIDROzyCUvwj",
							Secret:       "0x4AAAAAAABnPIDROzyCUvwj_secret",
							Name:         params.Name,
							Domains:      params.Domains,
							Mode:         params.Mode,
							BotFightMode: params.BotFightMode,
							Region:       params.Region,
							OffLabel:     params.OffLabel,
							CreatedOn:    &createdTime,
							ModifiedOn:   &createdTime,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID:    accountID,
					Name:         "Test Widget",
					Domains:      []string{"example.com", "*.example.com"},
					Mode:         ptr.To("managed"),
					BotFightMode: ptr.To(false),
					Region:       ptr.To("world"),
					OffLabel:     ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.TurnstileObservation{
					SiteKey:      ptr.To("0x4AAAAAAABnPIDROzyCUvwj"),
					Secret:       ptr.To("0x4AAAAAAABnPIDROzyCUvwj_secret"),
					Name:         ptr.To("Test Widget"),
					Domains:      []string{"example.com", "*.example.com"},
					Mode:         ptr.To("managed"),
					BotFightMode: ptr.To(false),
					Region:       ptr.To("world"),
					OffLabel:     ptr.To(false),
					CreatedOn:    &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
					ModifiedOn:   &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"CreateTurnstileMinimal": {
			reason: "Create should create Turnstile widget with minimal parameters",
			fields: fields{
				client: &MockTurnstileAPI{
					MockCreateTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
						return cloudflare.TurnstileWidget{
							SiteKey:      "0x4AAAAAAAMinimal",
							Secret:       "0x4AAAAAAAMinimal_secret",
							Name:         params.Name,
							Domains:      params.Domains,
							Mode:         "non-interactive",
							BotFightMode: false,     // Zero value
							Region:       "",        // Zero value
							OffLabel:     false,     // Zero value
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Minimal Widget",
					Domains:   []string{"example.com"},
				},
			},
			want: want{
				obs: &v1alpha1.TurnstileObservation{
					SiteKey:      ptr.To("0x4AAAAAAAMinimal"),
					Secret:       ptr.To("0x4AAAAAAAMinimal_secret"),
					Name:         ptr.To("Minimal Widget"),
					Domains:      []string{"example.com"},
					Mode:         ptr.To("non-interactive"),
					BotFightMode: ptr.To(false),  // convertTurnstileToObservation always creates pointers
					Region:       ptr.To(""),     // convertTurnstileToObservation always creates pointers
					OffLabel:     ptr.To(false),  // convertTurnstileToObservation always creates pointers
				},
				err: nil,
			},
		},
		"CreateTurnstileAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockTurnstileAPI{
					MockCreateTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
						return cloudflare.TurnstileWidget{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Test Widget",
					Domains:   []string{"example.com"},
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot create turnstile widget"),
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

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	accountID := "test-account-id"
	siteKey := "0x4AAAAAAABnPIDROzyCUvwj"

	type fields struct {
		client *MockTurnstileAPI
	}

	type args struct {
		ctx       context.Context
		accountID string
		siteKey   string
	}

	type want struct {
		obs *v1alpha1.TurnstileObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetTurnstileSuccess": {
			reason: "Get should return Turnstile widget when API call succeeds",
			fields: fields{
				client: &MockTurnstileAPI{
					MockGetTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.TurnstileWidget{}, errors.New("wrong resource type")
						}
						if siteKey != "0x4AAAAAAABnPIDROzyCUvwj" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong site key")
						}
						createdTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						modifiedTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
						return cloudflare.TurnstileWidget{
							SiteKey:      siteKey,
							Secret:       "0x4AAAAAAABnPIDROzyCUvwj_secret",
							Name:         "Test Widget",
							Domains:      []string{"example.com", "*.example.com"},
							Mode:         "managed",
							BotFightMode: false,
							Region:       "world",
							OffLabel:     false,
							CreatedOn:    &createdTime,
							ModifiedOn:   &modifiedTime,
						}, nil
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				obs: &v1alpha1.TurnstileObservation{
					SiteKey:      ptr.To("0x4AAAAAAABnPIDROzyCUvwj"),
					Secret:       ptr.To("0x4AAAAAAABnPIDROzyCUvwj_secret"),
					Name:         ptr.To("Test Widget"),
					Domains:      []string{"example.com", "*.example.com"},
					Mode:         ptr.To("managed"),
					BotFightMode: ptr.To(false),
					Region:       ptr.To("world"),
					OffLabel:     ptr.To(false),
					CreatedOn:    &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
					ModifiedOn:   &metav1.Time{Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"GetTurnstileNotFound": {
			reason: "Get should return NotFoundError when Turnstile widget is not found",
			fields: fields{
				client: &MockTurnstileAPI{
					MockGetTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error) {
						return cloudflare.TurnstileWidget{}, errors.New("widget not found")
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("turnstile widget not found"),
			},
		},
		"GetTurnstileAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockTurnstileAPI{
					MockGetTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error) {
						return cloudflare.TurnstileWidget{}, errBoom
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get turnstile widget"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.accountID, tc.args.siteKey)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGet(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nGet(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")
	accountID := "test-account-id"
	siteKey := "0x4AAAAAAABnPIDROzyCUvwj"

	type fields struct {
		client *MockTurnstileAPI
	}

	type args struct {
		ctx     context.Context
		siteKey string
		params  v1alpha1.TurnstileParameters
	}

	type want struct {
		obs *v1alpha1.TurnstileObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateTurnstileSuccess": {
			reason: "Update should update Turnstile widget when API call succeeds",
			fields: fields{
				client: &MockTurnstileAPI{
					MockUpdateTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.TurnstileWidget{}, errors.New("wrong resource type")
						}
						if params.SiteKey != "0x4AAAAAAABnPIDROzyCUvwj" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong site key")
						}
						if params.Name == nil || *params.Name != "Updated Widget" {
							return cloudflare.TurnstileWidget{}, errors.New("wrong name")
						}
						modifiedTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
						return cloudflare.TurnstileWidget{
							SiteKey:      params.SiteKey,
							Secret:       "0x4AAAAAAABnPIDROzyCUvwj_secret",
							Name:         *params.Name,
							Domains:      *params.Domains,
							Mode:         *params.Mode,
							BotFightMode: *params.BotFightMode,
							Region:       "",         // Zero value for Region
							OffLabel:     *params.OffLabel,
							ModifiedOn:   &modifiedTime,
						}, nil
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				siteKey: siteKey,
				params: v1alpha1.TurnstileParameters{
					AccountID:    accountID,
					Name:         "Updated Widget",
					Domains:      []string{"updated.example.com"},
					Mode:         ptr.To("invisible"),
					BotFightMode: ptr.To(true),
					OffLabel:     ptr.To(true),
				},
			},
			want: want{
				obs: &v1alpha1.TurnstileObservation{
					SiteKey:      ptr.To("0x4AAAAAAABnPIDROzyCUvwj"),
					Secret:       ptr.To("0x4AAAAAAABnPIDROzyCUvwj_secret"),
					Name:         ptr.To("Updated Widget"),
					Domains:      []string{"updated.example.com"},
					Mode:         ptr.To("invisible"),
					BotFightMode: ptr.To(true),
					Region:       ptr.To(""),      // convertTurnstileToObservation always creates pointers
					OffLabel:     ptr.To(true),
					ModifiedOn:   &metav1.Time{Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"UpdateTurnstileAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockTurnstileAPI{
					MockUpdateTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
						return cloudflare.TurnstileWidget{}, errBoom
					},
				},
			},
			args: args{
				ctx:     context.Background(),
				siteKey: siteKey,
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Updated Widget",
					Domains:   []string{"example.com"},
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot update turnstile widget"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.siteKey, tc.args.params)
			
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
	accountID := "test-account-id"
	siteKey := "0x4AAAAAAABnPIDROzyCUvwj"

	type fields struct {
		client *MockTurnstileAPI
	}

	type args struct {
		ctx       context.Context
		accountID string
		siteKey   string
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
		"DeleteTurnstileSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockTurnstileAPI{
					MockDeleteTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error {
						if rc.Identifier != "test-account-id" {
							return errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return errors.New("wrong resource type")
						}
						if siteKey != "0x4AAAAAAABnPIDROzyCUvwj" {
							return errors.New("wrong site key")
						}
						return nil
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteTurnstileNotFound": {
			reason: "Delete should succeed when Turnstile widget is not found (already deleted)",
			fields: fields{
				client: &MockTurnstileAPI{
					MockDeleteTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error {
						return errors.New("widget not found")
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteTurnstileAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockTurnstileAPI{
					MockDeleteTurnstileWidget: func(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:       context.Background(),
				accountID: accountID,
				siteKey:   siteKey,
			},
			want: want{
				err: errors.Wrap(errBoom, "cannot delete turnstile widget"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.accountID, tc.args.siteKey)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	accountID := "test-account-id"

	type fields struct {
		client *MockTurnstileAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.TurnstileParameters
		obs    v1alpha1.TurnstileObservation
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
				client: &MockTurnstileAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID:    accountID,
					Name:         "Test Widget",
					Domains:      []string{"example.com", "api.example.com"},
					Mode:         ptr.To("managed"),
					BotFightMode: ptr.To(true),
					Region:       ptr.To("world"),
					OffLabel:     ptr.To(false),
				},
				obs: v1alpha1.TurnstileObservation{
					Name:         ptr.To("Test Widget"),
					Domains:      []string{"api.example.com", "example.com"}, // Different order
					Mode:         ptr.To("managed"),
					BotFightMode: ptr.To(true),
					Region:       ptr.To("world"),
					OffLabel:     ptr.To(false),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseName": {
			reason: "IsUpToDate should return false when name doesn't match",
			fields: fields{
				client: &MockTurnstileAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Updated Widget",
					Domains:   []string{"example.com"},
				},
				obs: v1alpha1.TurnstileObservation{
					Name:    ptr.To("Original Widget"),
					Domains: []string{"example.com"},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseDomains": {
			reason: "IsUpToDate should return false when domains don't match",
			fields: fields{
				client: &MockTurnstileAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Test Widget",
					Domains:   []string{"example.com", "api.example.com"},
				},
				obs: v1alpha1.TurnstileObservation{
					Name:    ptr.To("Test Widget"),
					Domains: []string{"example.com"}, // Missing domain
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseMode": {
			reason: "IsUpToDate should return false when mode doesn't match",
			fields: fields{
				client: &MockTurnstileAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID: accountID,
					Name:      "Test Widget",
					Domains:   []string{"example.com"},
					Mode:      ptr.To("invisible"),
				},
				obs: v1alpha1.TurnstileObservation{
					Name:    ptr.To("Test Widget"),
					Domains: []string{"example.com"},
					Mode:    ptr.To("managed"),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseBotFightMode": {
			reason: "IsUpToDate should return false when BotFightMode doesn't match",
			fields: fields{
				client: &MockTurnstileAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TurnstileParameters{
					AccountID:    accountID,
					Name:         "Test Widget",
					Domains:      []string{"example.com"},
					BotFightMode: ptr.To(true),
				},
				obs: v1alpha1.TurnstileObservation{
					Name:         ptr.To("Test Widget"),
					Domains:      []string{"example.com"},
					BotFightMode: ptr.To(false),
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

func TestEqualStringSlices(t *testing.T) {
	type args struct {
		a []string
		b []string
	}

	type want struct {
		equal bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"EqualEmpty": {
			reason: "equalStringSlices should return true for empty slices",
			args: args{
				a: []string{},
				b: []string{},
			},
			want: want{
				equal: true,
			},
		},
		"EqualSameOrder": {
			reason: "equalStringSlices should return true for same elements in same order",
			args: args{
				a: []string{"a", "b", "c"},
				b: []string{"a", "b", "c"},
			},
			want: want{
				equal: true,
			},
		},
		"EqualDifferentOrder": {
			reason: "equalStringSlices should return true for same elements in different order",
			args: args{
				a: []string{"a", "b", "c"},
				b: []string{"c", "a", "b"},
			},
			want: want{
				equal: true,
			},
		},
		"NotEqualDifferentLength": {
			reason: "equalStringSlices should return false for slices with different lengths",
			args: args{
				a: []string{"a", "b"},
				b: []string{"a", "b", "c"},
			},
			want: want{
				equal: false,
			},
		},
		"NotEqualDifferentElements": {
			reason: "equalStringSlices should return false for slices with different elements",
			args: args{
				a: []string{"a", "b", "c"},
				b: []string{"a", "b", "d"},
			},
			want: want{
				equal: false,
			},
		},
		"EqualWithDuplicates": {
			reason: "equalStringSlices should handle duplicates correctly",
			args: args{
				a: []string{"a", "b", "a"},
				b: []string{"a", "a", "b"},
			},
			want: want{
				equal: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := equalStringSlices(tc.args.a, tc.args.b)
			if diff := cmp.Diff(tc.want.equal, got); diff != "" {
				t.Errorf("\n%s\nequalStringSlices(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"WidgetNotFoundError": {
			reason: "isNotFound should return true for 'widget not found' error",
			args: args{
				err: errors.New("widget not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"DoesNotExistError": {
			reason: "isNotFound should return true for 'does not exist' error",
			args: args{
				err: errors.New("does not exist"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"CaseInsensitiveError": {
			reason: "isNotFound should be case insensitive",
			args: args{
				err: errors.New("WIDGET NOT FOUND"),
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