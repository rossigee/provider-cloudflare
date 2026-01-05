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

package tunnel

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/tunnel/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/tunnel"

	"sigs.k8s.io/controller-runtime/pkg/client"

	rtmeta "github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
)



// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

// mockTunnelAPI mocks the tunnel.TunnelAPI interface
type mockTunnelAPI struct {
	MockArgoTunnels               func(ctx context.Context, accountID string) ([]cloudflare.ArgoTunnel, error)
	MockArgoTunnel                func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error)
	MockCreateArgoTunnel          func(ctx context.Context, accountID, name, secret string) (cloudflare.ArgoTunnel, error)
	MockDeleteArgoTunnel          func(ctx context.Context, accountID, tunnelUUID string) error
	MockCleanupArgoTunnelConnections func(ctx context.Context, accountID, tunnelUUID string) error
}

func (m *mockTunnelAPI) ArgoTunnels(ctx context.Context, accountID string) ([]cloudflare.ArgoTunnel, error) {
	if m.MockArgoTunnels != nil {
		return m.MockArgoTunnels(ctx, accountID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockTunnelAPI) ArgoTunnel(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
	if m.MockArgoTunnel != nil {
		return m.MockArgoTunnel(ctx, accountID, tunnelUUID)
	}
	return cloudflare.ArgoTunnel{}, errors.New("cannot get tunnel")
}

func (m *mockTunnelAPI) CreateArgoTunnel(ctx context.Context, accountID, name, secret string) (cloudflare.ArgoTunnel, error) {
	if m.MockCreateArgoTunnel != nil {
		return m.MockCreateArgoTunnel(ctx, accountID, name, secret)
	}
	return cloudflare.ArgoTunnel{}, errors.New("cannot create tunnel")
}

func (m *mockTunnelAPI) DeleteArgoTunnel(ctx context.Context, accountID, tunnelUUID string) error {
	if m.MockDeleteArgoTunnel != nil {
		return m.MockDeleteArgoTunnel(ctx, accountID, tunnelUUID)
	}
	return errors.New("cannot delete tunnel")
}

func (m *mockTunnelAPI) CleanupArgoTunnelConnections(ctx context.Context, accountID, tunnelUUID string) error {
	if m.MockCleanupArgoTunnelConnections != nil {
		return m.MockCleanupArgoTunnelConnections(ctx, accountID, tunnelUUID)
	}
	return nil
}

// Helper to create a CloudflareTunnelClient with a mocked API
func newMockTunnelClient(api tunnel.TunnelAPI) *tunnel.CloudflareTunnelClient {
	return tunnel.NewClient(api)
}

type tunnelModifier func(*v1beta1.Tunnel)

func withTunnelID(id string) tunnelModifier {
	return func(t *v1beta1.Tunnel) {
		t.Status.AtProvider.ID = id
		rtmeta.SetExternalName(t, id)
	}
}

func tunnelResource(m ...tunnelModifier) *v1beta1.Tunnel {
	t := &v1beta1.Tunnel{
		Spec: v1beta1.TunnelSpec{
			ForProvider: v1beta1.TunnelParameters{
				AccountID: "test-account-id",
				Name:      "test-tunnel",
				Secret:    "dGVzdC1zZWNyZXQ=",
			},
		},
	}
	for _, f := range m {
		f(t)
	}
	return t
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube        client.Client
		newServiceFn func(api *cloudflare.API) *tunnel.CloudflareTunnelClient
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
		"ErrNotTunnel": {
			reason: "Should return an error if the managed resource is not a Tunnel",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotTunnel),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: tunnelResource(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errGetCreds),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &tunnelConnector{
				kube:         tc.fields.kube,
				newServiceFn: tc.fields.newServiceFn,
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
		service *tunnel.CloudflareTunnelClient
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
		"ErrNotTunnel": {
			reason: "Should return an error if the managed resource is not a Tunnel",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotTunnel),
			},
		},
		"ErrGetTunnel": {
			reason: "Should return any error encountered getting the tunnel",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot get tunnel"), "cannot get external resource"),
			},
		},
		"TunnelNotFound": {
			reason: "Should report that the tunnel does not exist",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{}, errors.New("tunnel not found")
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"TunnelExistsAndUpToDate": {
			reason: "Should report that the tunnel exists and is up to date",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{
							ID:   "test-tunnel-id",
							Name: "test-tunnel",
						}, nil
					},
				}),
			},
			args: args{
				mg: tunnelResource(
					withTunnelID("test-tunnel-id"),
					func(t *v1beta1.Tunnel) {
						t.Spec.ForProvider.Name = "test-tunnel"
					},
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"TunnelExistsButOutOfDate": {
			reason: "Should report that the tunnel exists but is not up to date",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{
							ID:   "test-tunnel-id",
							Name: "test-tunnel",
						}, nil
					},
				}),
			},
			args: args{
				mg: tunnelResource(
					withTunnelID("test-tunnel-id"),
					func(t *v1beta1.Tunnel) {
						t.Spec.ForProvider.Name = "updated-tunnel"
					},
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
			e := &tunnelExternal{service: tc.fields.service}
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

func TestCreate(t *testing.T) {
	type fields struct {
		service *tunnel.CloudflareTunnelClient
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
		"ErrNotTunnel": {
			reason: "Should return an error if the managed resource is not a Tunnel",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotTunnel),
			},
		},
		"ErrCreateTunnel": {
			reason: "Should return any error encountered creating the tunnel",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockCreateArgoTunnel: func(ctx context.Context, accountID, name, secret string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: tunnelResource(),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot create tunnel"), "cannot create external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when tunnel is created successfully",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockCreateArgoTunnel: func(ctx context.Context, accountID, name, secret string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{
							ID:   "new-tunnel-id",
							Name: name,
						}, nil
					},
				}),
			},
			args: args{
				mg: tunnelResource(),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tunnelExternal{service: tc.fields.service}
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

func TestUpdate(t *testing.T) {
	type fields struct {
		service *tunnel.CloudflareTunnelClient
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
		"ErrNotTunnel": {
			reason: "Should return an error if the managed resource is not a Tunnel",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotTunnel),
			},
		},
		"ErrUpdateTunnel": {
			reason: "Should return any error encountered updating the tunnel",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot get tunnel for update"), "cannot update external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when tunnel is updated successfully",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) (cloudflare.ArgoTunnel, error) {
						return cloudflare.ArgoTunnel{
							ID:   "test-tunnel-id",
							Name: "test-tunnel",
						}, nil
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tunnelExternal{service: tc.fields.service}
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

func TestDelete(t *testing.T) {
	type fields struct {
		service *tunnel.CloudflareTunnelClient
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
		"ErrNotTunnel": {
			reason: "Should return an error if the managed resource is not a Tunnel",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotTunnel),
			},
		},
		"ErrDeleteTunnel": {
			reason: "Should return any error encountered deleting the tunnel",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockCleanupArgoTunnelConnections: func(ctx context.Context, accountID, tunnelUUID string) error {
						return nil
					},
					MockDeleteArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) error {
						return errors.New("cannot delete tunnel")
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("cannot delete tunnel"), "cannot delete tunnel"),
			},
		},
		"Success": {
			reason: "Should return no error when tunnel is deleted successfully",
			fields: fields{
				service: newMockTunnelClient(&mockTunnelAPI{
					MockCleanupArgoTunnelConnections: func(ctx context.Context, accountID, tunnelUUID string) error {
						return nil
					},
					MockDeleteArgoTunnel: func(ctx context.Context, accountID, tunnelUUID string) error {
						return nil
					},
				}),
			},
			args: args{
				mg: tunnelResource(withTunnelID("test-tunnel-id")),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &tunnelExternal{service: tc.fields.service}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}