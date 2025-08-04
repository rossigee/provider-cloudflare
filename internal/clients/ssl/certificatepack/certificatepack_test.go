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

package certificatepack

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

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// MockCertificatePackAPI implements the CertificatePackAPI interface for testing
type MockCertificatePackAPI struct {
	MockCertificatePack              func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error)
	MockCreateCertificatePack        func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error)
	MockDeleteCertificatePack        func(ctx context.Context, zoneID, certificateID string) error
	MockRestartCertificateValidation func(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error)
}

func (m *MockCertificatePackAPI) CertificatePack(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
	if m.MockCertificatePack != nil {
		return m.MockCertificatePack(ctx, zoneID, certificatePackID)
	}
	return cloudflare.CertificatePack{}, nil
}

func (m *MockCertificatePackAPI) CreateCertificatePack(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
	if m.MockCreateCertificatePack != nil {
		return m.MockCreateCertificatePack(ctx, zoneID, cert)
	}
	return cloudflare.CertificatePack{}, nil
}

func (m *MockCertificatePackAPI) DeleteCertificatePack(ctx context.Context, zoneID, certificateID string) error {
	if m.MockDeleteCertificatePack != nil {
		return m.MockDeleteCertificatePack(ctx, zoneID, certificateID)
	}
	return nil
}

func (m *MockCertificatePackAPI) RestartCertificateValidation(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error) {
	if m.MockRestartCertificateValidation != nil {
		return m.MockRestartCertificateValidation(ctx, zoneID, certificateID)
	}
	return cloudflare.CertificatePack{}, nil
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	certPackID := "test-cert-pack-id"

	type fields struct {
		client *MockCertificatePackAPI
	}

	type args struct {
		ctx               context.Context
		zoneID            string
		certificatePackID string
	}

	type want struct {
		obs *v1alpha1.CertificatePackObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetCertificatePackSuccess": {
			reason: "Get should return Certificate Pack when API call succeeds",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.CertificatePack{}, errors.New("wrong zone ID")
						}
						if certificatePackID != "test-cert-pack-id" {
							return cloudflare.CertificatePack{}, errors.New("wrong certificate pack ID")
						}
						return cloudflare.CertificatePack{
							ID:                   "test-cert-pack-id",
							Type:                 "advanced",
							Hosts:                []string{"example.com", "*.example.com"},
							ValidationMethod:     "txt",
							ValidityDays:         90,
							CertificateAuthority: "digicert",
							CloudflareBranding:   false,
							Status:               "active",
						}, nil
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: &v1alpha1.CertificatePackObservation{
					ID:                   ptr.To("test-cert-pack-id"),
					Type:                 ptr.To("advanced"),
					Hosts:                []string{"example.com", "*.example.com"},
					ValidationMethod:     ptr.To("txt"),
					ValidityDays:         ptr.To(90),
					CertificateAuthority: ptr.To("digicert"),
					CloudflareBranding:   ptr.To(false),
					Status:               ptr.To("active"),
				},
				err: nil,
			},
		},
		"GetCertificatePackWithCertificates": {
			reason: "Get should return Certificate Pack with certificate details",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						expiryTime := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
						uploadTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						modifyTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
						
						return cloudflare.CertificatePack{
							ID:               "test-cert-pack-id",
							Type:             "advanced",
							Hosts:            []string{"example.com"},
							ValidationMethod: "http",
							ValidityDays:     365,
							Status:           "pending_validation",
							Certificates: []cloudflare.CertificatePackCertificate{
								{
									ID:         "cert-1",
									Hosts:      []string{"example.com"},
									Issuer:     "DigiCert Inc",
									Status:     "active",
									ExpiresOn:  expiryTime,
									UploadedOn: uploadTime,
									ModifiedOn: modifyTime,
								},
							},
						}, nil
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: &v1alpha1.CertificatePackObservation{
					ID:               ptr.To("test-cert-pack-id"),
					Type:             ptr.To("advanced"),
					Hosts:            []string{"example.com"},
					ValidationMethod: ptr.To("http"),
					ValidityDays:     ptr.To(365),
					Status:           ptr.To("pending_validation"),
					CloudflareBranding: ptr.To(false),
					Certificates: []v1alpha1.CertificateInfo{
						{
							ID:         ptr.To("cert-1"),
							Hosts:      []string{"example.com"},
							Issuer:     ptr.To("DigiCert Inc"),
							Status:     ptr.To("active"),
							ExpiresOn:  &metav1.Time{Time: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)},
							UploadedOn: &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
							ModifiedOn: &metav1.Time{Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)},
						},
					},
				},
				err: nil,
			},
		},
		"GetCertificatePackNotFound": {
			reason: "Get should return NotFoundError when Certificate Pack is not found",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errors.New("not found")
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: nil,
				err: clients.NewNotFoundError("certificate pack not found"),
			},
		},
		"GetCertificatePackAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errBoom
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot get certificate pack"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.zoneID, tc.args.certificatePackID)
			
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
		client *MockCertificatePackAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.CertificatePackParameters
	}

	type want struct {
		obs *v1alpha1.CertificatePackObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateCertificatePackSuccess": {
			reason: "Create should create Certificate Pack when API call succeeds",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCreateCertificatePack: func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.CertificatePack{}, errors.New("wrong zone ID")
						}
						if cert.Type != "advanced" {
							return cloudflare.CertificatePack{}, errors.New("wrong type")
						}
						if len(cert.Hosts) != 2 || cert.Hosts[0] != "example.com" || cert.Hosts[1] != "*.example.com" {
							return cloudflare.CertificatePack{}, errors.New("wrong hosts")
						}
						if cert.ValidationMethod != "txt" {
							return cloudflare.CertificatePack{}, errors.New("wrong validation method")
						}
						return cloudflare.CertificatePack{
							ID:                   "new-cert-pack-id",
							Type:                 cert.Type,
							Hosts:                cert.Hosts,
							ValidationMethod:     cert.ValidationMethod,
							ValidityDays:         cert.ValidityDays,
							CertificateAuthority: cert.CertificateAuthority,
							CloudflareBranding:   cert.CloudflareBranding,
							Status:               "pending_validation",
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificatePackParameters{
					Zone:                 zoneID,
					Type:                 "advanced",
					Hosts:                []string{"example.com", "*.example.com"},
					ValidationMethod:     "txt",
					ValidityDays:         ptr.To(90),
					CertificateAuthority: ptr.To("digicert"),
					CloudflareBranding:   ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.CertificatePackObservation{
					ID:                   ptr.To("new-cert-pack-id"),
					Type:                 ptr.To("advanced"),
					Hosts:                []string{"example.com", "*.example.com"},
					ValidationMethod:     ptr.To("txt"),
					ValidityDays:         ptr.To(90),
					CertificateAuthority: ptr.To("digicert"),
					CloudflareBranding:   ptr.To(false),
					Status:               ptr.To("pending_validation"),
				},
				err: nil,
			},
		},
		"CreateCertificatePackBasic": {
			reason: "Create should create basic Certificate Pack with minimal parameters",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCreateCertificatePack: func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{
							ID:               "basic-cert-pack-id",
							Type:             cert.Type,
							Hosts:            cert.Hosts,
							ValidationMethod: cert.ValidationMethod,
							Status:           "active",
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificatePackParameters{
					Zone:             zoneID,
					Type:             "universal",
					Hosts:            []string{"example.com"},
					ValidationMethod: "http",
				},
			},
			want: want{
				obs: &v1alpha1.CertificatePackObservation{
					ID:               ptr.To("basic-cert-pack-id"),
					Type:             ptr.To("universal"),
					Hosts:            []string{"example.com"},
					ValidationMethod: ptr.To("http"),
					ValidityDays:     ptr.To(0),
					Status:           ptr.To("active"),
					CloudflareBranding: ptr.To(false),
				},
				err: nil,
			},
		},
		"CreateCertificatePackAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockCreateCertificatePack: func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.CertificatePackParameters{
					Zone:             zoneID,
					Type:             "advanced",
					Hosts:            []string{"example.com"},
					ValidationMethod: "txt",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot create certificate pack"),
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

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	certPackID := "test-cert-pack-id"

	type fields struct {
		client *MockCertificatePackAPI
	}

	type args struct {
		ctx               context.Context
		zoneID            string
		certificatePackID string
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
		"DeleteCertificatePackSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockDeleteCertificatePack: func(ctx context.Context, zoneID, certificateID string) error {
						if zoneID != "test-zone-id" {
							return errors.New("wrong zone ID")
						}
						if certificateID != "test-cert-pack-id" {
							return errors.New("wrong certificate pack ID")
						}
						return nil
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteCertificatePackNotFound": {
			reason: "Delete should succeed when Certificate Pack is not found (already deleted)",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockDeleteCertificatePack: func(ctx context.Context, zoneID, certificateID string) error {
						return errors.New("not found")
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteCertificatePackAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockDeleteCertificatePack: func(ctx context.Context, zoneID, certificateID string) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				err: errors.Wrap(errBoom, "cannot delete certificate pack"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.zoneID, tc.args.certificatePackID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestRestartValidation(t *testing.T) {
	errBoom := errors.New("boom")
	zoneID := "test-zone-id"
	certPackID := "test-cert-pack-id"

	type fields struct {
		client *MockCertificatePackAPI
	}

	type args struct {
		ctx               context.Context
		zoneID            string
		certificatePackID string
	}

	type want struct {
		obs *v1alpha1.CertificatePackObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"RestartValidationSuccess": {
			reason: "RestartValidation should restart validation when API call succeeds",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockRestartCertificateValidation: func(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error) {
						if zoneID != "test-zone-id" {
							return cloudflare.CertificatePack{}, errors.New("wrong zone ID")
						}
						if certificateID != "test-cert-pack-id" {
							return cloudflare.CertificatePack{}, errors.New("wrong certificate pack ID")
						}
						return cloudflare.CertificatePack{
							ID:               "test-cert-pack-id",
							Type:             "advanced",
							Hosts:            []string{"example.com"},
							ValidationMethod: "txt",
							Status:           "pending_validation",
						}, nil
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: &v1alpha1.CertificatePackObservation{
					ID:                 ptr.To("test-cert-pack-id"),
					Type:               ptr.To("advanced"),
					Hosts:              []string{"example.com"},
					ValidationMethod:   ptr.To("txt"),
					ValidityDays:       ptr.To(0),
					Status:             ptr.To("pending_validation"),
					CloudflareBranding: ptr.To(false),
				},
				err: nil,
			},
		},
		"RestartValidationAPIError": {
			reason: "RestartValidation should return wrapped error when API call fails",
			fields: fields{
				client: &MockCertificatePackAPI{
					MockRestartCertificateValidation: func(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errBoom
					},
				},
			},
			args: args{
				ctx:               context.Background(),
				zoneID:            zoneID,
				certificatePackID: certPackID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, "cannot restart certificate validation"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.RestartValidation(tc.args.ctx, tc.args.zoneID, tc.args.certificatePackID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nRestartValidation(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nRestartValidation(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertParametersToCertificatePackRequest(t *testing.T) {
	type args struct {
		params v1alpha1.CertificatePackParameters
	}

	type want struct {
		request cloudflare.CertificatePackRequest
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllParameters": {
			reason: "convertParametersToCertificatePackRequest should convert all parameters correctly",
			args: args{
				params: v1alpha1.CertificatePackParameters{
					Zone:                 "test-zone-id",
					Type:                 "advanced",
					Hosts:                []string{"example.com", "*.example.com"},
					ValidationMethod:     "txt",
					ValidityDays:         ptr.To(90),
					CertificateAuthority: ptr.To("digicert"),
					CloudflareBranding:   ptr.To(false),
				},
			},
			want: want{
				request: cloudflare.CertificatePackRequest{
					Type:                 "advanced",
					Hosts:                []string{"example.com", "*.example.com"},
					ValidationMethod:     "txt",
					ValidityDays:         90,
					CertificateAuthority: "digicert",
					CloudflareBranding:   false,
				},
			},
		},
		"ConvertMinimalParameters": {
			reason: "convertParametersToCertificatePackRequest should handle minimal parameters",
			args: args{
				params: v1alpha1.CertificatePackParameters{
					Zone:             "test-zone-id",
					Type:             "universal",
					Hosts:            []string{"example.com"},
					ValidationMethod: "http",
				},
			},
			want: want{
				request: cloudflare.CertificatePackRequest{
					Type:             "universal",
					Hosts:            []string{"example.com"},
					ValidationMethod: "http",
					// ValidityDays, CertificateAuthority, CloudflareBranding should be zero values
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertParametersToCertificatePackRequest(tc.args.params)
			if diff := cmp.Diff(tc.want.request, got); diff != "" {
				t.Errorf("\n%s\nconvertParametersToCertificatePackRequest(...): -want, +got:\n%s\n", tc.reason, diff)
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