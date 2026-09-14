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

package ssl

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/ssl/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/ssl/certificatepack"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// mockCertificatePackAPI mocks the certificatepack.CertificatePackAPI interface
type mockCertificatePackAPI struct {
	MockCertificatePack              func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error)
	MockCreateCertificatePack        func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error)
	MockDeleteCertificatePack        func(ctx context.Context, zoneID, certificateID string) error
	MockRestartCertificateValidation func(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error)
}

func (m *mockCertificatePackAPI) CertificatePack(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
	if m.MockCertificatePack != nil {
		return m.MockCertificatePack(ctx, zoneID, certificatePackID)
	}
	return cloudflare.CertificatePack{}, errors.New("not implemented")
}

func (m *mockCertificatePackAPI) CreateCertificatePack(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
	if m.MockCreateCertificatePack != nil {
		return m.MockCreateCertificatePack(ctx, zoneID, cert)
	}
	return cloudflare.CertificatePack{}, errors.New("cannot create certificate pack")
}

func (m *mockCertificatePackAPI) DeleteCertificatePack(ctx context.Context, zoneID, certificateID string) error {
	if m.MockDeleteCertificatePack != nil {
		return m.MockDeleteCertificatePack(ctx, zoneID, certificateID)
	}
	return errors.New("cannot delete certificate pack")
}

func (m *mockCertificatePackAPI) RestartCertificateValidation(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error) {
	if m.MockRestartCertificateValidation != nil {
		return m.MockRestartCertificateValidation(ctx, zoneID, certificateID)
	}
	return cloudflare.CertificatePack{}, errors.New("cannot restart certificate validation")
}

// Helper to create a CloudflareCertificatePackClient with a mocked API
func newMockCertificatePackClient(api certificatepack.CertificatePackAPI) *certificatepack.CloudflareCertificatePackClient {
	return certificatepack.NewClient(api)
}

type certificatePackModifier func(*v1beta1.CertificatePack)

func withCertificatePackID(id string) certificatePackModifier {
	return func(cp *v1beta1.CertificatePack) {
		cp.Status.AtProvider.ID = &id
		meta.SetExternalName(cp, id)
	}
}

func certificatePack(m ...certificatePackModifier) *v1beta1.CertificatePack {
	cp := &v1beta1.CertificatePack{
		Spec: v1beta1.CertificatePackSpec{
			ForProvider: v1beta1.CertificatePackParameters{
				Zone:             "test-zone-id",
				Type:             "advanced",
				Hosts:            []string{"example.com"},
				ValidationMethod: "txt",
			},
		},
	}
	for _, f := range m {
		f(cp)
	}
	return cp
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube                  client.Client
		newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   error
	}{
		"ErrNotCertificatePack": {
			reason: "Should return an error if the managed resource is not a CertificatePack",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotCertificatePack),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: certificatePack(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errGetCredsCert),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &certificatePackConnector{
				kube:                  tc.fields.kube,
				newCloudflareClientFn: tc.fields.newCloudflareClientFn,
			}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Connect(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	type fields struct {
		service *certificatepack.CloudflareCertificatePackClient
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotCertificatePack": {
			reason: "Should return an error if the managed resource is not a CertificatePack",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCertificatePack),
			},
		},
		"ErrGetCertificatePack": {
			reason: "Should return any error encountered getting the certificate pack",
			fields: fields{
				service: newMockCertificatePackClient(&mockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: certificatePack(withCertificatePackID("test-cert-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot get certificate pack"), "failed to get Certificate Pack"),
			},
		},
		"CertificatePackNotFound": {
			reason: "Should report that the certificate pack does not exist",
			fields: fields{
				service: newMockCertificatePackClient(&mockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{}, errors.New("not found")
					},
				}),
			},
			args: args{
				mg: certificatePack(withCertificatePackID("test-cert-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"CertificatePackExistsAndUpToDate": {
			reason: "Should report that the certificate pack exists and is up to date",
			fields: fields{
				service: newMockCertificatePackClient(&mockCertificatePackAPI{
					MockCertificatePack: func(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{
							ID:   "test-cert-id",
							Type: "advanced",
						}, nil
					},
				}),
			},
			args: args{
				mg: certificatePack(withCertificatePackID("test-cert-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &certificatePackExternal{
				service: tc.fields.service,
			}
			o, err := c.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Observe(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("%s\nc.Observe(...): -want observation, +got observation:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type fields struct {
		service *certificatepack.CloudflareCertificatePackClient
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotCertificatePack": {
			reason: "Should return an error if the managed resource is not a CertificatePack",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCertificatePack),
			},
		},
		"SuccessfulCreate": {
			reason: "Should successfully create a certificate pack",
			fields: fields{
				service: newMockCertificatePackClient(&mockCertificatePackAPI{
					MockCreateCertificatePack: func(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
						return cloudflare.CertificatePack{
							ID:   "created-cert-id",
							Type: "advanced",
						}, nil
					},
				}),
			},
			args: args{
				mg: certificatePack(),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &certificatePackExternal{
				service: tc.fields.service,
			}
			o, err := c.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Create(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("%s\nc.Create(...): -want creation, +got creation:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type fields struct {
		service *certificatepack.CloudflareCertificatePackClient
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		d   managed.ExternalDelete
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotCertificatePack": {
			reason: "Should return an error if the managed resource is not a CertificatePack",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCertificatePack),
			},
		},
		"SuccessfulDelete": {
			reason: "Should successfully delete a certificate pack",
			fields: fields{
				service: newMockCertificatePackClient(&mockCertificatePackAPI{
					MockDeleteCertificatePack: func(ctx context.Context, zoneID, certificateID string) error {
						return nil
					},
				}),
			},
			args: args{
				mg: certificatePack(withCertificatePackID("test-cert-id")),
			},
			want: want{
				d:   managed.ExternalDelete{},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &certificatePackExternal{
				service: tc.fields.service,
			}
			d, err := c.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\nc.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.d, d); diff != "" {
				t.Errorf("%s\nc.Delete(...): -want deletion, +got deletion:\n%s", tc.reason, diff)
			}
		})
	}
}
