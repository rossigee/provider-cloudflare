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

package certificate

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

	"github.com/rossigee/provider-cloudflare/apis/originssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// MockOriginCACertificateAPI implements the OriginCACertificateAPI interface for testing
type MockOriginCACertificateAPI struct {
	MockGetOriginCACertificate    func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error)
	MockCreateOriginCACertificate func(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error)
	MockRevokeOriginCACertificate func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error)
}

func (m *MockOriginCACertificateAPI) GetOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
	if m.MockGetOriginCACertificate != nil {
		return m.MockGetOriginCACertificate(ctx, certificateID)
	}
	return &cloudflare.OriginCACertificate{}, nil
}

func (m *MockOriginCACertificateAPI) CreateOriginCACertificate(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
	if m.MockCreateOriginCACertificate != nil {
		return m.MockCreateOriginCACertificate(ctx, params)
	}
	return &cloudflare.OriginCACertificate{}, nil
}

func (m *MockOriginCACertificateAPI) RevokeOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error) {
	if m.MockRevokeOriginCACertificate != nil {
		return m.MockRevokeOriginCACertificate(ctx, certificateID)
	}
	return &cloudflare.OriginCACertificateID{}, nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	certificateID := "test-cert-id"

	type fields struct {
		client *MockOriginCACertificateAPI
	}

	type args struct {
		ctx           context.Context
		certificateID string
	}

	type want struct {
		obs *v1alpha1.CertificateObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetOriginCACertificateSuccess": {
			reason: "Get should return Origin CA Certificate when API call succeeds",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockGetOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
						if certificateID != "test-cert-id" {
							return nil, errors.New("wrong certificate ID")
						}
						expiresOn := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
						return &cloudflare.OriginCACertificate{
							ID:              "test-cert-id",
							Certificate:     "-----BEGIN CERTIFICATE-----\nMIIDTest...\n-----END CERTIFICATE-----",
							Hostnames:       []string{"example.com", "*.example.com"},
							RequestType:     "origin-rsa",
							RequestValidity: 7300,
							CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nMIICTest...\n-----END CERTIFICATE REQUEST-----",
							ExpiresOn:       expiresOn,
						}, nil
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:              "test-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nMIIDTest...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"example.com", "*.example.com"},
					RequestType:     "origin-rsa",
					RequestValidity: 7300,
					CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nMIICTest...\n-----END CERTIFICATE REQUEST-----",
					ExpiresOn:       &metav1.Time{Time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"GetOriginCACertificateMinimal": {
			reason: "Get should return Origin CA Certificate with minimal fields",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockGetOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
						return &cloudflare.OriginCACertificate{
							ID:          "minimal-cert-id",
							Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
							Hostnames:   []string{"minimal.example.com"},
						}, nil
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: "minimal-cert-id",
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:          "minimal-cert-id",
					Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
					Hostnames:   []string{"minimal.example.com"},
				},
				err: nil,
			},
		},
		"GetOriginCACertificateNotFound": {
			reason: "Get should return NotFoundError when Origin CA Certificate is not found",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockGetOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
						return nil, errors.New("certificate not found")
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("origin ca certificate not found"),
			},
		},
		"GetOriginCACertificateAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockGetOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get origin ca certificate"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.certificateID)
			
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

	type fields struct {
		client *MockOriginCACertificateAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.CertificateParameters
	}

	type want struct {
		obs *v1alpha1.CertificateObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateOriginCACertificateSuccess": {
			reason: "Create should create Origin CA Certificate when API call succeeds",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockCreateOriginCACertificate: func(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
						if len(params.Hostnames) != 2 {
							return nil, errors.New("wrong number of hostnames")
						}
						if params.Hostnames[0] != "example.com" || params.Hostnames[1] != "*.example.com" {
							return nil, errors.New("wrong hostnames")
						}
						if params.RequestType != "origin-rsa" {
							return nil, errors.New("wrong request type")
						}
						if params.RequestValidity != 365 {
							return nil, errors.New("wrong request validity")
						}
						expiresOn := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
						return &cloudflare.OriginCACertificate{
							ID:              "new-cert-id",
							Certificate:     "-----BEGIN CERTIFICATE-----\nNewCert...\n-----END CERTIFICATE-----",
							Hostnames:       params.Hostnames,
							RequestType:     params.RequestType,
							RequestValidity: params.RequestValidity,
							CSR:             params.CSR,
							ExpiresOn:       expiresOn,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames:       []string{"example.com", "*.example.com"},
					RequestType:     ptr.To("origin-rsa"),
					RequestValidity: ptr.To(365),
					CSR:             ptr.To("-----BEGIN CERTIFICATE REQUEST-----\nTestCSR...\n-----END CERTIFICATE REQUEST-----"),
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:              "new-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nNewCert...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"example.com", "*.example.com"},
					RequestType:     "origin-rsa",
					RequestValidity: 365,
					CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nTestCSR...\n-----END CERTIFICATE REQUEST-----",
					ExpiresOn:       &metav1.Time{Time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"CreateOriginCACertificateMinimal": {
			reason: "Create should create Origin CA Certificate with minimal parameters",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockCreateOriginCACertificate: func(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
						return &cloudflare.OriginCACertificate{
							ID:          "minimal-cert-id",
							Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
							Hostnames:   params.Hostnames,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"minimal.example.com"},
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:          "minimal-cert-id",
					Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
					Hostnames:   []string{"minimal.example.com"},
				},
				err: nil,
			},
		},
		"CreateOriginCACertificateECDSA": {
			reason: "Create should create ECDSA Origin CA Certificate",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockCreateOriginCACertificate: func(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
						if params.RequestType != "origin-ecc" {
							return nil, errors.New("wrong request type")
						}
						return &cloudflare.OriginCACertificate{
							ID:          "ecdsa-cert-id",
							Certificate: "-----BEGIN CERTIFICATE-----\nECDSA...\n-----END CERTIFICATE-----",
							Hostnames:   params.Hostnames,
							RequestType: params.RequestType,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames:   []string{"ecdsa.example.com"},
					RequestType: ptr.To("origin-ecc"),
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:          "ecdsa-cert-id",
					Certificate: "-----BEGIN CERTIFICATE-----\nECDSA...\n-----END CERTIFICATE-----",
					Hostnames:   []string{"ecdsa.example.com"},
					RequestType: "origin-ecc",
				},
				err: nil,
			},
		},
		"CreateOriginCACertificateAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockCreateOriginCACertificate: func(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"test.example.com"},
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot create origin ca certificate"),
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
	certificateID := "test-cert-id"

	type fields struct {
		client *MockOriginCACertificateAPI
	}

	type args struct {
		ctx           context.Context
		certificateID string
		params        v1alpha1.CertificateParameters
	}

	type want struct {
		obs *v1alpha1.CertificateObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateOriginCACertificateNotSupported": {
			reason: "Update should return error since Origin CA certificates cannot be updated",
			fields: fields{
				client: &MockOriginCACertificateAPI{},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"updated.example.com"},
				},
			},
			want: want{
				obs: nil,
				err: errors.New("origin ca certificates cannot be updated"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.certificateID, tc.args.params)
			
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
	certificateID := "test-cert-id"

	type fields struct {
		client *MockOriginCACertificateAPI
	}

	type args struct {
		ctx           context.Context
		certificateID string
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
		"DeleteOriginCACertificateSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockRevokeOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error) {
						if certificateID != "test-cert-id" {
							return nil, errors.New("wrong certificate ID")
						}
						return &cloudflare.OriginCACertificateID{ID: certificateID}, nil
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteOriginCACertificateNotFound": {
			reason: "Delete should succeed when Origin CA Certificate is not found (already revoked)",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockRevokeOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error) {
						return nil, errors.New("certificate not found")
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteOriginCACertificateAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockOriginCACertificateAPI{
					MockRevokeOriginCACertificate: func(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx:           context.Background(),
				certificateID: certificateID,
			},
			want: want{
				err: errors.Wrap(errBoom, "cannot revoke origin ca certificate"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.certificateID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type fields struct {
		client *MockOriginCACertificateAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.CertificateParameters
		obs    v1alpha1.CertificateObservation
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
			reason: "IsUpToDate should return true when hostnames match",
			fields: fields{
				client: &MockOriginCACertificateAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"example.com", "*.example.com"},
				},
				obs: v1alpha1.CertificateObservation{
					ID:        "cert-id",
					Hostnames: []string{"example.com", "*.example.com"},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateTrueDifferentOrder": {
			reason: "IsUpToDate should return true when hostnames match in different order",
			fields: fields{
				client: &MockOriginCACertificateAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"*.example.com", "example.com"},
				},
				obs: v1alpha1.CertificateObservation{
					ID:        "cert-id",
					Hostnames: []string{"example.com", "*.example.com"},
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseDifferentLength": {
			reason: "IsUpToDate should return false when hostnames have different lengths",
			fields: fields{
				client: &MockOriginCACertificateAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"example.com"},
				},
				obs: v1alpha1.CertificateObservation{
					ID:        "cert-id",
					Hostnames: []string{"example.com", "*.example.com"},
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseDifferentHostnames": {
			reason: "IsUpToDate should return false when hostnames are different",
			fields: fields{
				client: &MockOriginCACertificateAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"different.example.com"},
				},
				obs: v1alpha1.CertificateObservation{
					ID:        "cert-id",
					Hostnames: []string{"example.com"},
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

func TestConvertParametersToCreate(t *testing.T) {
	type args struct {
		params v1alpha1.CertificateParameters
	}

	type want struct {
		createParams cloudflare.CreateOriginCertificateParams
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllParameters": {
			reason: "convertParametersToCreate should convert all parameters correctly",
			args: args{
				params: v1alpha1.CertificateParameters{
					Hostnames:       []string{"example.com", "*.example.com"},
					RequestType:     ptr.To("origin-rsa"),
					RequestValidity: ptr.To(365),
					CSR:             ptr.To("-----BEGIN CERTIFICATE REQUEST-----\nTest...\n-----END CERTIFICATE REQUEST-----"),
				},
			},
			want: want{
				createParams: cloudflare.CreateOriginCertificateParams{
					Hostnames:       []string{"example.com", "*.example.com"},
					RequestType:     "origin-rsa",
					RequestValidity: 365,
					CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nTest...\n-----END CERTIFICATE REQUEST-----",
				},
			},
		},
		"ConvertMinimalParameters": {
			reason: "convertParametersToCreate should handle minimal parameters",
			args: args{
				params: v1alpha1.CertificateParameters{
					Hostnames: []string{"minimal.example.com"},
				},
			},
			want: want{
				createParams: cloudflare.CreateOriginCertificateParams{
					Hostnames: []string{"minimal.example.com"},
				},
			},
		},
		"ConvertECDSAParameters": {
			reason: "convertParametersToCreate should handle ECDSA parameters",
			args: args{
				params: v1alpha1.CertificateParameters{
					Hostnames:       []string{"ecdsa.example.com"},
					RequestType:     ptr.To("origin-ecc"),
					RequestValidity: ptr.To(7300),
				},
			},
			want: want{
				createParams: cloudflare.CreateOriginCertificateParams{
					Hostnames:       []string{"ecdsa.example.com"},
					RequestType:     "origin-ecc",
					RequestValidity: 7300,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToCreate(tc.args.params)
			if diff := cmp.Diff(tc.want.createParams, got); diff != "" {
				t.Errorf("\n%s\nconvertParametersToCreate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertCertificateToObservation(t *testing.T) {
	type args struct {
		cert *cloudflare.OriginCACertificate
	}

	type want struct {
		obs *v1alpha1.CertificateObservation
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertFullCertificate": {
			reason: "convertCertificateToObservation should convert all fields correctly",
			args: args{
				cert: &cloudflare.OriginCACertificate{
					ID:              "full-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nFull...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"full.example.com", "*.full.example.com"},
					RequestType:     "origin-rsa",
					RequestValidity: 365,
					CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nFull...\n-----END CERTIFICATE REQUEST-----",
					ExpiresOn:       time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
					RevokedAt:       time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:              "full-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nFull...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"full.example.com", "*.full.example.com"},
					RequestType:     "origin-rsa",
					RequestValidity: 365,
					CSR:             "-----BEGIN CERTIFICATE REQUEST-----\nFull...\n-----END CERTIFICATE REQUEST-----",
					ExpiresOn:       &metav1.Time{Time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
					RevokedAt:       &metav1.Time{Time: time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)},
				},
			},
		},
		"ConvertMinimalCertificate": {
			reason: "convertCertificateToObservation should handle minimal certificate",
			args: args{
				cert: &cloudflare.OriginCACertificate{
					ID:          "minimal-cert-id",
					Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
					Hostnames:   []string{"minimal.example.com"},
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:          "minimal-cert-id",
					Certificate: "-----BEGIN CERTIFICATE-----\nMinimal...\n-----END CERTIFICATE-----",
					Hostnames:   []string{"minimal.example.com"},
				},
			},
		},
		"ConvertECDSACertificate": {
			reason: "convertCertificateToObservation should handle ECDSA certificate",
			args: args{
				cert: &cloudflare.OriginCACertificate{
					ID:              "ecdsa-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nECDSA...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"ecdsa.example.com"},
					RequestType:     "origin-ecc",
					RequestValidity: 7300,
					ExpiresOn:       time.Date(2045, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			want: want{
				obs: &v1alpha1.CertificateObservation{
					ID:              "ecdsa-cert-id",
					Certificate:     "-----BEGIN CERTIFICATE-----\nECDSA...\n-----END CERTIFICATE-----",
					Hostnames:       []string{"ecdsa.example.com"},
					RequestType:     "origin-ecc",
					RequestValidity: 7300,
					ExpiresOn:       &metav1.Time{Time: time.Date(2045, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertCertificateToObservation(tc.args.cert)
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nconvertCertificateToObservation(...): -want, +got:\n%s\n", tc.reason, diff)
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
		"CertificateNotFoundError": {
			reason: "isNotFound should return true for 'certificate not found' error",
			args: args{
				err: errors.New("certificate not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"CaseInsensitiveError": {
			reason: "isNotFound should be case insensitive",
			args: args{
				err: errors.New("CERTIFICATE NOT FOUND"),
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