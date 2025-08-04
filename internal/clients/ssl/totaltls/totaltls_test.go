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

package totaltls

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

// MockTotalTLSAPI implements the TotalTLSAPI interface for testing
type MockTotalTLSAPI struct {
	MockGetTotalTLS func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error)
	MockSetTotalTLS func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error)
}

func (m *MockTotalTLSAPI) GetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
	if m.MockGetTotalTLS != nil {
		return m.MockGetTotalTLS(ctx, rc)
	}
	return cloudflare.TotalTLS{}, nil
}

func (m *MockTotalTLSAPI) SetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
	if m.MockSetTotalTLS != nil {
		return m.MockSetTotalTLS(ctx, rc, params)
	}
	return params, nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"

	type fields struct {
		client *MockTotalTLSAPI
	}

	type args struct {
		ctx    context.Context
		zoneID string
	}

	type want struct {
		obs *v1alpha1.TotalTLSObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetTotalTLSSuccess": {
			reason: "Get should return Total TLS settings when API call succeeds",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockGetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
						if rc.Identifier != zoneID {
							return cloudflare.TotalTLS{}, errors.New("wrong zone ID")
						}
						return cloudflare.TotalTLS{
							Enabled:              ptr.To(true),
							CertificateAuthority: "digicert",
							ValidityDays:         90,
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
				err: nil,
			},
		},
		"GetTotalTLSDisabled": {
			reason: "Get should return disabled Total TLS settings",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockGetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
						return cloudflare.TotalTLS{
							Enabled:              ptr.To(false),
							CertificateAuthority: "letsencrypt",
							ValidityDays:         30,
						}, nil
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(false),
					CertificateAuthority: ptr.To("letsencrypt"),
					ValidityDays:         ptr.To(30),
				},
				err: nil,
			},
		},
		"GetTotalTLSNotFound": {
			reason: "Get should return NotFoundError when Total TLS settings are not found",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockGetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
						return cloudflare.TotalTLS{}, errors.New("not found")
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("total tls settings not found"),
			},
		},
		"GetTotalTLSAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockGetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
						return cloudflare.TotalTLS{}, errBoom
					},
				},
			},
			args: args{
				ctx:    context.Background(),
				zoneID: zoneID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get total tls settings"),
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
		client *MockTotalTLSAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.TotalTLSParameters
	}

	type want struct {
		obs *v1alpha1.TotalTLSObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateTotalTLSEnable": {
			reason: "Update should enable Total TLS with specified settings",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockSetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
						if rc.Identifier != zoneID {
							return cloudflare.TotalTLS{}, errors.New("wrong zone ID")
						}
						if params.Enabled == nil || !*params.Enabled {
							return cloudflare.TotalTLS{}, errors.New("expected enabled to be true")
						}
						if params.CertificateAuthority != "digicert" {
							return cloudflare.TotalTLS{}, errors.New("expected CA to be digicert")
						}
						if params.ValidityDays != 90 {
							return cloudflare.TotalTLS{}, errors.New("expected validity days to be 90")
						}
						return cloudflare.TotalTLS{
							Enabled:              ptr.To(true),
							CertificateAuthority: "digicert",
							ValidityDays:         90,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:                 zoneID,
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
				err: nil,
			},
		},
		"UpdateTotalTLSDisable": {
			reason: "Update should disable Total TLS when requested",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockSetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
						if rc.Identifier != zoneID {
							return cloudflare.TotalTLS{}, errors.New("wrong zone ID")
						}
						if params.Enabled == nil || *params.Enabled {
							return cloudflare.TotalTLS{}, errors.New("expected enabled to be false")
						}
						return cloudflare.TotalTLS{
							Enabled:              ptr.To(false),
							CertificateAuthority: "letsencrypt",
							ValidityDays:         30,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:    zoneID,
					Enabled: ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(false),
					CertificateAuthority: ptr.To("letsencrypt"),
					ValidityDays:         ptr.To(30),
				},
				err: nil,
			},
		},
		"UpdateTotalTLSPartialSettings": {
			reason: "Update should handle partial settings correctly",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockSetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
						// Should only set enabled, keep other defaults
						if params.Enabled == nil || !*params.Enabled {
							return cloudflare.TotalTLS{}, errors.New("expected enabled to be true")
						}
						return cloudflare.TotalTLS{
							Enabled:              ptr.To(true),
							CertificateAuthority: "google",
							ValidityDays:         60,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:    zoneID,
					Enabled: ptr.To(true),
					// CertificateAuthority and ValidityDays are nil
				},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("google"),
					ValidityDays:         ptr.To(60),
				},
				err: nil,
			},
		},
		"UpdateTotalTLSAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockTotalTLSAPI{
					MockSetTotalTLS: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
						return cloudflare.TotalTLS{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:    zoneID,
					Enabled: ptr.To(true),
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot update total tls settings"),
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
		client *MockTotalTLSAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.TotalTLSParameters
		obs    v1alpha1.TotalTLSObservation
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
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:                 zoneID,
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
				obs: v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseEnabled": {
			reason: "IsUpToDate should return false when enabled setting doesn't match",
			fields: fields{
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:    zoneID,
					Enabled: ptr.To(true),
				},
				obs: v1alpha1.TotalTLSObservation{
					Enabled: ptr.To(false),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseCA": {
			reason: "IsUpToDate should return false when certificate authority doesn't match",
			fields: fields{
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:                 zoneID,
					CertificateAuthority: ptr.To("digicert"),
				},
				obs: v1alpha1.TotalTLSObservation{
					CertificateAuthority: ptr.To("letsencrypt"),
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseValidityDays": {
			reason: "IsUpToDate should return false when validity days doesn't match",
			fields: fields{
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:         zoneID,
					ValidityDays: ptr.To(90),
				},
				obs: v1alpha1.TotalTLSObservation{
					ValidityDays: ptr.To(30),
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
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone: zoneID,
					// All other params are nil
				},
				obs: v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateNilObservation": {
			reason: "IsUpToDate should return true when observation values are nil",
			fields: fields{
				client: &MockTotalTLSAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.TotalTLSParameters{
					Zone:    zoneID,
					Enabled: ptr.To(true),
				},
				obs: v1alpha1.TotalTLSObservation{
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

func TestConvertParametersToTotalTLS(t *testing.T) {
	type args struct {
		params v1alpha1.TotalTLSParameters
	}

	type want struct {
		settings cloudflare.TotalTLS
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllParameters": {
			reason: "convertParametersToTotalTLS should convert all parameters correctly",
			args: args{
				params: v1alpha1.TotalTLSParameters{
					Zone:                 "test-zone-id",
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
			},
			want: want{
				settings: cloudflare.TotalTLS{
					Enabled:              ptr.To(true),
					CertificateAuthority: "digicert",
					ValidityDays:         90,
				},
			},
		},
		"ConvertPartialParameters": {
			reason: "convertParametersToTotalTLS should handle nil parameters correctly",
			args: args{
				params: v1alpha1.TotalTLSParameters{
					Zone:    "test-zone-id",
					Enabled: ptr.To(false),
					// CertificateAuthority and ValidityDays are nil
				},
			},
			want: want{
				settings: cloudflare.TotalTLS{
					Enabled: ptr.To(false),
					// CertificateAuthority and ValidityDays should be zero values
				},
			},
		},
		"ConvertEmptyParameters": {
			reason: "convertParametersToTotalTLS should handle all nil parameters",
			args: args{
				params: v1alpha1.TotalTLSParameters{
					Zone: "test-zone-id",
					// All other parameters are nil
				},
			},
			want: want{
				settings: cloudflare.TotalTLS{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToTotalTLS(tc.args.params)
			if diff := cmp.Diff(tc.want.settings, got); diff != "" {
				t.Errorf("\n%s\nconvertParametersToTotalTLS(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertTotalTLSToObservation(t *testing.T) {
	type args struct {
		settings cloudflare.TotalTLS
	}

	type want struct {
		obs *v1alpha1.TotalTLSObservation
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllSettings": {
			reason: "convertTotalTLSToObservation should convert all settings correctly",
			args: args{
				settings: cloudflare.TotalTLS{
					Enabled:              ptr.To(true),
					CertificateAuthority: "digicert",
					ValidityDays:         90,
				},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(true),
					CertificateAuthority: ptr.To("digicert"),
					ValidityDays:         ptr.To(90),
				},
			},
		},
		"ConvertDisabledSettings": {
			reason: "convertTotalTLSToObservation should convert disabled settings correctly",
			args: args{
				settings: cloudflare.TotalTLS{
					Enabled:              ptr.To(false),
					CertificateAuthority: "letsencrypt",
					ValidityDays:         30,
				},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              ptr.To(false),
					CertificateAuthority: ptr.To("letsencrypt"),
					ValidityDays:         ptr.To(30),
				},
			},
		},
		"ConvertEmptySettings": {
			reason: "convertTotalTLSToObservation should handle empty settings",
			args: args{
				settings: cloudflare.TotalTLS{},
			},
			want: want{
				obs: &v1alpha1.TotalTLSObservation{
					Enabled:              nil,
					CertificateAuthority: ptr.To(""),
					ValidityDays:         ptr.To(0),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertTotalTLSToObservation(tc.args.settings)
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nconvertTotalTLSToObservation(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"TLSNotFoundError": {
			reason: "isNotFound should return true for 'tls not found' error",
			args: args{
				err: errors.New("tls not found"),
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
				err: errors.New("TLS NOT FOUND"),
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