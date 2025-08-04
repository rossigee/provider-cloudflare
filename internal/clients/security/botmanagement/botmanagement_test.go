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

package botmanagement

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

// MockBotManagementAPI implements the BotManagementAPI interface for testing
type MockBotManagementAPI struct {
	MockGetBotManagement    func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error)
	MockUpdateBotManagement func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error)
}

func (m *MockBotManagementAPI) GetBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
	if m.MockGetBotManagement != nil {
		return m.MockGetBotManagement(ctx, rc)
	}
	return cloudflare.BotManagement{}, nil
}

func (m *MockBotManagementAPI) UpdateBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error) {
	if m.MockUpdateBotManagement != nil {
		return m.MockUpdateBotManagement(ctx, rc, params)
	}
	return cloudflare.BotManagement{}, nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockBotManagementAPI
	}

	type args struct {
		ctx    context.Context
		zoneID string
	}

	type want struct {
		obs *v1alpha1.BotManagementObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetBotManagementSuccess": {
			reason: "Get should return Bot Management settings when API call succeeds",
			fields: fields{
				client: &MockBotManagementAPI{
					MockGetBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.BotManagement{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.BotManagement{}, errors.New("wrong resource type")
						}
						return cloudflare.BotManagement{
							EnableJS:                     ptr.To(true),
							FightMode:                    ptr.To(true),
							SBFMDefinitelyAutomated:      ptr.To("block"),
							SBFMLikelyAutomated:          ptr.To("managed_challenge"),
							SBFMVerifiedBots:             ptr.To("allow"),
							SBFMStaticResourceProtection: ptr.To(false),
							OptimizeWordpress:            ptr.To(true),
							SuppressSessionScore:         ptr.To(false),
							AutoUpdateModel:              ptr.To(true),
							UsingLatestModel:             ptr.To(true),
							AIBotsProtection:             ptr.To("allow"),
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					UsingLatestModel:             ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
				err: nil,
			},
		},
		"GetBotManagementMinimal": {
			reason: "Get should return Bot Management with minimal settings",
			fields: fields{
				client: &MockBotManagementAPI{
					MockGetBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
						return cloudflare.BotManagement{
							EnableJS:         ptr.To(false),
							FightMode:        ptr.To(false),
							UsingLatestModel: ptr.To(false),
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:         ptr.To(false),
					FightMode:        ptr.To(false),
					UsingLatestModel: ptr.To(false),
				},
				err: nil,
			},
		},
		"GetBotManagementNotFound": {
			reason: "Get should return NotFoundError when Bot Management is not found",
			fields: fields{
				client: &MockBotManagementAPI{
					MockGetBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
						return cloudflare.BotManagement{}, errors.New("bot management not found")
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("bot management configuration not found"),
			},
		},
		"GetBotManagementAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockBotManagementAPI{
					MockGetBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
						return cloudflare.BotManagement{}, errBoom
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get bot management configuration"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.zoneID)
			
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
	zoneID := "test-zone-id"

	type fields struct {
		client *MockBotManagementAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.BotManagementParameters
	}

	type want struct {
		obs *v1alpha1.BotManagementObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateBotManagementSuccess": {
			reason: "Update should update Bot Management when API call succeeds",
			fields: fields{
				client: &MockBotManagementAPI{
					MockUpdateBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error) {
						if rc.Identifier != "test-zone-id" {
							return cloudflare.BotManagement{}, errors.New("wrong zone ID")
						}
						if rc.Type != cloudflare.ZoneType {
							return cloudflare.BotManagement{}, errors.New("wrong resource type")
						}
						if params.EnableJS == nil || !*params.EnableJS {
							return cloudflare.BotManagement{}, errors.New("expected EnableJS to be true")
						}
						if params.FightMode == nil || !*params.FightMode {
							return cloudflare.BotManagement{}, errors.New("expected FightMode to be true")
						}
						return cloudflare.BotManagement{
							EnableJS:                     params.EnableJS,
							FightMode:                    params.FightMode,
							SBFMDefinitelyAutomated:      params.SBFMDefinitelyAutomated,
							SBFMLikelyAutomated:          params.SBFMLikelyAutomated,
							SBFMVerifiedBots:             params.SBFMVerifiedBots,
							SBFMStaticResourceProtection: params.SBFMStaticResourceProtection,
							OptimizeWordpress:            params.OptimizeWordpress,
							SuppressSessionScore:         params.SuppressSessionScore,
							AutoUpdateModel:              params.AutoUpdateModel,
							UsingLatestModel:             ptr.To(true),
							AIBotsProtection:             params.AIBotsProtection,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:                         zoneID,
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					UsingLatestModel:             ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
				err: nil,
			},
		},
		"UpdateBotManagementPartial": {
			reason: "Update should handle partial Bot Management updates",
			fields: fields{
				client: &MockBotManagementAPI{
					MockUpdateBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error) {
						// Should only update specified fields
						if params.EnableJS == nil || !*params.EnableJS {
							return cloudflare.BotManagement{}, errors.New("expected EnableJS to be true")
						}
						return cloudflare.BotManagement{
							EnableJS:         params.EnableJS,
							FightMode:        ptr.To(false), // Default value
							UsingLatestModel: ptr.To(true),
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:     zoneID,
					EnableJS: ptr.To(true),
					// Other fields are nil
				},
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:         ptr.To(true),
					FightMode:        ptr.To(false),
					UsingLatestModel: ptr.To(true),
				},
				err: nil,
			},
		},
		"UpdateBotManagementAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockBotManagementAPI{
					MockUpdateBotManagement: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error) {
						return cloudflare.BotManagement{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:     zoneID,
					EnableJS: ptr.To(true),
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot update bot management configuration"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	zoneID := "test-zone-id"

	type fields struct {
		client *MockBotManagementAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.BotManagementParameters
		obs    v1alpha1.BotManagementObservation
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
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:                    zoneID,
					EnableJS:                ptr.To(true),
					FightMode:               ptr.To(true),
					SBFMDefinitelyAutomated: ptr.To("block"),
					OptimizeWordpress:       ptr.To(false),
				},
				obs: v1alpha1.BotManagementObservation{
					EnableJS:                ptr.To(true),
					FightMode:               ptr.To(true),
					SBFMDefinitelyAutomated: ptr.To("block"),
					OptimizeWordpress:       ptr.To(false),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseEnableJS": {
			reason: "IsUpToDate should return false when EnableJS doesn't match",
			fields: fields{
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:     zoneID,
					EnableJS: ptr.To(true),
				},
				obs: v1alpha1.BotManagementObservation{
					EnableJS: ptr.To(false),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseFightMode": {
			reason: "IsUpToDate should return false when FightMode doesn't match",
			fields: fields{
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:      zoneID,
					FightMode: ptr.To(true),
				},
				obs: v1alpha1.BotManagementObservation{
					FightMode: ptr.To(false),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseSBFMDefinitelyAutomated": {
			reason: "IsUpToDate should return false when SBFMDefinitelyAutomated doesn't match",
			fields: fields{
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:                    zoneID,
					SBFMDefinitelyAutomated: ptr.To("block"),
				},
				obs: v1alpha1.BotManagementObservation{
					SBFMDefinitelyAutomated: ptr.To("allow"),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseAIBotsProtection": {
			reason: "IsUpToDate should return false when AIBotsProtection doesn't match",
			fields: fields{
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone:             zoneID,
					AIBotsProtection: ptr.To("block"),
				},
				obs: v1alpha1.BotManagementObservation{
					AIBotsProtection: ptr.To("allow"),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateNilParams": {
			reason: "IsUpToDate should return true when parameters are nil",
			fields: fields{
				client: &MockBotManagementAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BotManagementParameters{
					Zone: zoneID,
					// All other params are nil
				},
				obs: v1alpha1.BotManagementObservation{
					EnableJS:  ptr.To(true),
					FightMode: ptr.To(false),
				},
			},
			want: want{
				upToDate: true,
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

func TestConvertParametersToBotManagement(t *testing.T) {
	type args struct {
		params v1alpha1.BotManagementParameters
	}

	type want struct {
		updateParams cloudflare.UpdateBotManagementParams
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllParameters": {
			reason: "convertParametersToBotManagement should convert all parameters correctly",
			args: args{
				params: v1alpha1.BotManagementParameters{
					Zone:                         "test-zone-id",
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
			},
			want: want{
				updateParams: cloudflare.UpdateBotManagementParams{
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
			},
		},
		"ConvertPartialParameters": {
			reason: "convertParametersToBotManagement should handle nil parameters correctly",
			args: args{
				params: v1alpha1.BotManagementParameters{
					Zone:     "test-zone-id",
					EnableJS: ptr.To(false),
					// Other parameters are nil
				},
			},
			want: want{
				updateParams: cloudflare.UpdateBotManagementParams{
					EnableJS: ptr.To(false),
					// Other fields should be nil
				},
			},
		},
		"ConvertEmptyParameters": {
			reason: "convertParametersToBotManagement should handle all nil parameters",
			args: args{
				params: v1alpha1.BotManagementParameters{
					Zone: "test-zone-id",
					// All other parameters are nil
				},
			},
			want: want{
				updateParams: cloudflare.UpdateBotManagementParams{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToBotManagement(tc.args.params)
			if diff := cmp.Diff(tc.want.updateParams, got); diff != "" {
				t.Errorf("\n%s\nconvertParametersToBotManagement(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertBotManagementToObservation(t *testing.T) {
	type args struct {
		botManagement cloudflare.BotManagement
	}

	type want struct {
		obs *v1alpha1.BotManagementObservation
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllSettings": {
			reason: "convertBotManagementToObservation should convert all settings correctly",
			args: args{
				botManagement: cloudflare.BotManagement{
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					UsingLatestModel:             ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:                     ptr.To(true),
					FightMode:                    ptr.To(true),
					SBFMDefinitelyAutomated:      ptr.To("block"),
					SBFMLikelyAutomated:          ptr.To("managed_challenge"),
					SBFMVerifiedBots:             ptr.To("allow"),
					SBFMStaticResourceProtection: ptr.To(false),
					OptimizeWordpress:            ptr.To(true),
					SuppressSessionScore:         ptr.To(false),
					AutoUpdateModel:              ptr.To(true),
					UsingLatestModel:             ptr.To(true),
					AIBotsProtection:             ptr.To("allow"),
				},
			},
		},
		"ConvertMinimalSettings": {
			reason: "convertBotManagementToObservation should handle minimal settings",
			args: args{
				botManagement: cloudflare.BotManagement{
					EnableJS:         ptr.To(false),
					UsingLatestModel: ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{
					EnableJS:         ptr.To(false),
					UsingLatestModel: ptr.To(false),
				},
			},
		},
		"ConvertEmptySettings": {
			reason: "convertBotManagementToObservation should handle empty settings",
			args: args{
				botManagement: cloudflare.BotManagement{},
			},
			want: want{
				obs: &v1alpha1.BotManagementObservation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertBotManagementToObservation(tc.args.botManagement)
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nconvertBotManagementToObservation(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"BotManagementNotFoundError": {
			reason: "isNotFound should return true for 'bot management not found' error",
			args: args{
				err: errors.New("bot management not found"),
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
				err: errors.New("BOT MANAGEMENT NOT FOUND"),
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