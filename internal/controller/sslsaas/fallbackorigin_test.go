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

package sslsaas

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	fallbackorigin "github.com/rossigee/provider-cloudflare/internal/clients/sslsaas/fallbackorigin"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

type mockFallbackOriginClient struct {
	MockFallbackOrigin       func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error)
	MockUpdateFallbackOrigin func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error)
	MockDeleteFallbackOrigin func(ctx context.Context, zoneID string) error
}

func (m *mockFallbackOriginClient) FallbackOrigin(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
	return m.MockFallbackOrigin(ctx, zoneID)
}

func (m *mockFallbackOriginClient) UpdateFallbackOrigin(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
	return m.MockUpdateFallbackOrigin(ctx, zoneID, origin)
}

func (m *mockFallbackOriginClient) DeleteFallbackOrigin(ctx context.Context, zoneID string) error {
	return m.MockDeleteFallbackOrigin(ctx, zoneID)
}

type fallbackOriginModifier func(*v1alpha1.FallbackOrigin)

func withFallbackZone(zone string) fallbackOriginModifier {
	return func(fo *v1alpha1.FallbackOrigin) { fo.Spec.ForProvider.Zone = &zone }
}

func withFallbackOrigin(origin string) fallbackOriginModifier {
	return func(fo *v1alpha1.FallbackOrigin) { fo.Spec.ForProvider.Origin = &origin }
}

func fallbackOriginCR(m ...fallbackOriginModifier) *v1alpha1.FallbackOrigin {
	fo := &v1alpha1.FallbackOrigin{
		Spec: v1alpha1.FallbackOriginSpec{
			ForProvider: v1alpha1.FallbackOriginParameters{},
		},
	}
	for _, f := range m {
		f(fo)
	}
	return fo
}

func TestFallbackOriginConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube                  client.Client
		newCloudflareClientFn func(cfg clients.Config) (fallbackorigin.Client, error)
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
		"ErrNotFallbackOrigin": {
			reason: "Should return an error if the managed resource is not a FallbackOrigin",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotFallbackOrigin),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: fallbackOriginCR(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errClientConfig),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &fallbackOriginConnector{
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

func TestFallbackOriginObserve(t *testing.T) {
	type fields struct {
		client fallbackorigin.Client
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
		"ErrNotFallbackOrigin": {
			reason: "Should return an error if the managed resource is not a FallbackOrigin",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotFallbackOrigin),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: fallbackOriginCR(),
			},
			want: want{
				err: errors.New(errFallbackOriginNoZone),
			},
		},
		"ErrGetFallbackOrigin": {
			reason: "Should return any error encountered getting the fallback origin",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockFallbackOrigin: func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
						return cloudflare.CustomHostnameFallbackOrigin{}, errors.New("boom")
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(withFallbackZone("test-zone-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errFallbackOriginLookup),
			},
		},
		"FallbackOriginNotFound": {
			reason: "Should report that the fallback origin does not exist",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockFallbackOrigin: func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
						return cloudflare.CustomHostnameFallbackOrigin{}, errors.New("Fallback Origin not found")
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(withFallbackZone("test-zone-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"FallbackOriginExistsAndUpToDate": {
			reason: "Should report that the fallback origin exists and is up to date",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockFallbackOrigin: func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
						return cloudflare.CustomHostnameFallbackOrigin{
							Origin: "fallback.example.com",
							Status: "active",
						}, nil
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("fallback.example.com"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"FallbackOriginExistsButNotUpToDate": {
			reason: "Should report that the fallback origin exists but is not up to date",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockFallbackOrigin: func(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error) {
						return cloudflare.CustomHostnameFallbackOrigin{
							Origin: "old-fallback.example.com",
							Status: "active",
						}, nil
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("new-fallback.example.com"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &fallbackOriginExternal{client: tc.fields.client}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Observe(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("%s\ne.Observe(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestFallbackOriginCreate(t *testing.T) {
	type fields struct {
		client fallbackorigin.Client
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
		"ErrNotFallbackOrigin": {
			reason: "Should return an error if the managed resource is not a FallbackOrigin",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotFallbackOrigin),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: fallbackOriginCR(),
			},
			want: want{
				err: errors.New(errFallbackOriginNoZone),
			},
		},
		"ErrCreateFallbackOrigin": {
			reason: "Should return any error encountered creating the fallback origin",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockUpdateFallbackOrigin: func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
						return nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("fallback.example.com"),
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errFallbackOriginCreation),
			},
		},
		"Success": {
			reason: "Should return no error when fallback origin is created successfully",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockUpdateFallbackOrigin: func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
						return &cloudflare.CustomHostnameFallbackOriginResponse{
							Result: cloudflare.CustomHostnameFallbackOrigin{
								Origin: "fallback.example.com",
								Status: "active",
							},
						}, nil
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("fallback.example.com"),
				),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &fallbackOriginExternal{client: tc.fields.client}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Create(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("%s\ne.Create(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestFallbackOriginUpdate(t *testing.T) {
	type fields struct {
		client fallbackorigin.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotFallbackOrigin": {
			reason: "Should return an error if the managed resource is not a FallbackOrigin",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotFallbackOrigin),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: fallbackOriginCR(),
			},
			want: want{
				err: errors.New(errFallbackOriginNoZone),
			},
		},
		"ErrUpdateFallbackOrigin": {
			reason: "Should return any error encountered updating the fallback origin",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockUpdateFallbackOrigin: func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
						return nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("fallback.example.com"),
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errFallbackOriginUpdate),
			},
		},
		"Success": {
			reason: "Should return no error when fallback origin is updated successfully",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockUpdateFallbackOrigin: func(ctx context.Context, zoneID string, origin cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error) {
						return &cloudflare.CustomHostnameFallbackOriginResponse{
							Result: cloudflare.CustomHostnameFallbackOrigin{
								Origin: "fallback.example.com",
								Status: "active",
							},
						}, nil
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(
					withFallbackZone("test-zone-id"),
					withFallbackOrigin("fallback.example.com"),
				),
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &fallbackOriginExternal{client: tc.fields.client}
			got, err := e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Update(...): -want error, +got error:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("%s\ne.Update(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestFallbackOriginDelete(t *testing.T) {
	type fields struct {
		client fallbackorigin.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
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
		"ErrNotFallbackOrigin": {
			reason: "Should return an error if the managed resource is not a FallbackOrigin",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotFallbackOrigin),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: fallbackOriginCR(),
			},
			want: want{
				err: errors.New(errFallbackOriginNoZone),
			},
		},
		"ErrDeleteFallbackOrigin": {
			reason: "Should return any error encountered deleting the fallback origin",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockDeleteFallbackOrigin: func(ctx context.Context, zoneID string) error {
						return errors.New("boom")
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(withFallbackZone("test-zone-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errFallbackOriginDeletion),
			},
		},
		"Success": {
			reason: "Should return no error when fallback origin is deleted successfully",
			fields: fields{
				client: &mockFallbackOriginClient{
					MockDeleteFallbackOrigin: func(ctx context.Context, zoneID string) error {
						return nil
					},
				},
			},
			args: args{
				mg: fallbackOriginCR(withFallbackZone("test-zone-id")),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &fallbackOriginExternal{client: tc.fields.client}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}