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

package universalssl

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// MockUniversalSSLAPI implements the UniversalSSLAPI interface for testing
type MockUniversalSSLAPI struct {
	MockUniversalSSLSettingDetails func(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error)
	MockEditUniversalSSLSetting    func(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error)
}

func (m *MockUniversalSSLAPI) UniversalSSLSettingDetails(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
	if m.MockUniversalSSLSettingDetails != nil {
		return m.MockUniversalSSLSettingDetails(ctx, zoneID)
	}
	return cloudflare.UniversalSSLSetting{}, nil
}

func (m *MockUniversalSSLAPI) EditUniversalSSLSetting(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error) {
	if m.MockEditUniversalSSLSetting != nil {
		return m.MockEditUniversalSSLSetting(ctx, zoneID, setting)
	}
	return setting, nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockUniversalSSLAPI
	}

	type args struct {
		ctx    context.Context
		zoneID string
	}

	type want struct {
		obs *v1alpha1.UniversalSSLObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetUniversalSSLSuccess": {
			reason: "Get should return Universal SSL settings when API call succeeds",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockUniversalSSLSettingDetails: func(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
						return cloudflare.UniversalSSLSetting{
							Enabled: true,
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(true),
				},
				err: nil,
			},
		},
		"GetUniversalSSLDisabled": {
			reason: "Get should return disabled Universal SSL settings",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockUniversalSSLSettingDetails: func(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
						return cloudflare.UniversalSSLSetting{
							Enabled: false,
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(false),
				},
				err: nil,
			},
		},
		"GetUniversalSSLNotFound": {
			reason: "Get should return NotFoundError when Universal SSL settings are not found",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockUniversalSSLSettingDetails: func(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
						return cloudflare.UniversalSSLSetting{}, errors.New("not found")
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("universal ssl settings not found"),
			},
		},
		"GetUniversalSSLAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockUniversalSSLSettingDetails: func(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
						return cloudflare.UniversalSSLSetting{}, errBoom
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get universal ssl settings"),
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
		client *MockUniversalSSLAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.UniversalSSLParameters
	}

	type want struct {
		obs *v1alpha1.UniversalSSLObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateUniversalSSLEnable": {
			reason: "Update should enable Universal SSL when requested",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockEditUniversalSSLSetting: func(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.UniversalSSLSetting{}, errors.New("wrong zone ID")
						}
						if !setting.Enabled {
							return cloudflare.UniversalSSLSetting{}, errors.New("expected enabled to be true")
						}
						return cloudflare.UniversalSSLSetting{
							Enabled: true,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: true,
				},
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(true),
				},
				err: nil,
			},
		},
		"UpdateUniversalSSLDisable": {
			reason: "Update should disable Universal SSL when requested",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockEditUniversalSSLSetting: func(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.UniversalSSLSetting{}, errors.New("wrong zone ID")
						}
						if setting.Enabled {
							return cloudflare.UniversalSSLSetting{}, errors.New("expected enabled to be false")
						}
						return cloudflare.UniversalSSLSetting{
							Enabled: false,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: false,
				},
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(false),
				},
				err: nil,
			},
		},
		"UpdateUniversalSSLAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockUniversalSSLAPI{
					MockEditUniversalSSLSetting: func(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error) {
						return cloudflare.UniversalSSLSetting{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: true,
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot update universal ssl settings"),
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
		client *MockUniversalSSLAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.UniversalSSLParameters
		obs    v1alpha1.UniversalSSLObservation
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
			reason: "IsUpToDate should return true when settings match",
			fields: fields{
				client: &MockUniversalSSLAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: true,
				},
				obs: v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(true),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalse": {
			reason: "IsUpToDate should return false when settings don't match",
			fields: fields{
				client: &MockUniversalSSLAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: true,
				},
				obs: v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(false),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateNilObservation": {
			reason: "IsUpToDate should return true when observation enabled is nil",
			fields: fields{
				client: &MockUniversalSSLAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.UniversalSSLParameters{
					Zone:    zoneID,
					Enabled: true,
				},
				obs: v1alpha1.UniversalSSLObservation{
					Enabled: nil,
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

func TestConvertParametersToUniversalSSL(t *testing.T) {
	type args struct {
		params v1alpha1.UniversalSSLParameters
	}

	type want struct {
		setting cloudflare.UniversalSSLSetting
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertEnabled": {
			reason: "convertParametersToUniversalSSL should convert enabled parameters correctly",
			args: args{
				params: v1alpha1.UniversalSSLParameters{
					Zone:    "test-zone-id",
					Enabled: true,
				},
			},
			want: want{
				setting: cloudflare.UniversalSSLSetting{
					Enabled: true,
				},
			},
		},
		"ConvertDisabled": {
			reason: "convertParametersToUniversalSSL should convert disabled parameters correctly",
			args: args{
				params: v1alpha1.UniversalSSLParameters{
					Zone:    "test-zone-id",
					Enabled: false,
				},
			},
			want: want{
				setting: cloudflare.UniversalSSLSetting{
					Enabled: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToUniversalSSL(tc.args.params)
			if diff := cmp.Diff(tc.want.setting, got); diff != "" {
				t.Errorf("\n%s\nconvertParametersToUniversalSSL(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertUniversalSSLToObservation(t *testing.T) {
	type args struct {
		settings cloudflare.UniversalSSLSetting
	}

	type want struct {
		obs *v1alpha1.UniversalSSLObservation
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertEnabled": {
			reason: "convertUniversalSSLToObservation should convert enabled settings correctly",
			args: args{
				settings: cloudflare.UniversalSSLSetting{
					Enabled: true,
				},
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(true),
				},
			},
		},
		"ConvertDisabled": {
			reason: "convertUniversalSSLToObservation should convert disabled settings correctly",
			args: args{
				settings: cloudflare.UniversalSSLSetting{
					Enabled: false,
				},
			},
			want: want{
				obs: &v1alpha1.UniversalSSLObservation{
					Enabled: ptr.To(false),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertUniversalSSLToObservation(tc.args.settings)
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nconvertUniversalSSLToObservation(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"SSLNotFoundError": {
			reason: "isNotFound should return true for 'ssl not found' error",
			args: args{
				err: errors.New("ssl not found"),
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
				err: errors.New("NOT FOUND"),
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