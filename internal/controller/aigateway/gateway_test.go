/*
Copyright 2026 The Crossplane Authors.

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

package aigateway

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rossigee/provider-cloudflare/apis/aigateway/v1beta1"
	pcv1beta1 "github.com/rossigee/provider-cloudflare/apis/v1beta1"
	aigatewayclient "github.com/rossigee/provider-cloudflare/internal/clients/aigateway"
)

const (
	testAccountID = "acc-123"
	testGatewayID = "my-gateway"
	externalName  = "my-gateway"
)

var errBoom = errors.New("boom")

// mockClient implements aigatewayclient.Client with overridable functions.
type mockClient struct {
	MockCreate func(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error)
	MockGet    func(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error)
	MockUpdate func(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error)
	MockDelete func(ctx context.Context, accountID, gatewayID string) error
}

func (m *mockClient) Create(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
	return m.MockCreate(ctx, accountID, p)
}

func (m *mockClient) Get(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error) {
	return m.MockGet(ctx, accountID, gatewayID)
}

func (m *mockClient) Update(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
	return m.MockUpdate(ctx, accountID, gatewayID, p)
}

func (m *mockClient) Delete(ctx context.Context, accountID, gatewayID string) error {
	return m.MockDelete(ctx, accountID, gatewayID)
}

type gatewayModifier func(*v1beta1.Gateway)

func withExternalName(name string) gatewayModifier {
	return func(g *v1beta1.Gateway) { meta.SetExternalName(g, name) }
}

func withProviderConfig() gatewayModifier {
	return func(g *v1beta1.Gateway) {
		g.Spec.ProviderConfigReference = &xpv1.ProviderConfigReference{Name: "default"}
	}
}

func withCacheTTL(ttl int64) gatewayModifier {
	return func(g *v1beta1.Gateway) { g.Spec.ForProvider.CacheTTL = ttl }
}

func gateway(m ...gatewayModifier) *v1beta1.Gateway {
	g := &v1beta1.Gateway{
		Spec: v1beta1.GatewaySpec{
			ForProvider: v1beta1.GatewayParameters{
				AccountID:               testAccountID,
				GatewayID:               testGatewayID,
				CacheInvalidateOnUpdate: true,
				CacheTTL:                3600,
				CollectLogs:             true,
				RateLimitingInterval:    60,
				RateLimitingLimit:       100,
				RateLimitingTechnique:   ptr.To("fixed"),
			},
		},
	}
	for _, fn := range m {
		fn(g)
	}
	return g
}

// remoteGateway returns the observed state that matches gateway() with default settings.
func remoteGateway() aigatewayclient.Gateway {
	return aigatewayclient.Gateway{
		ID:                      testGatewayID,
		CacheInvalidateOnUpdate: true,
		CacheTTL:                3600,
		CollectLogs:             true,
		RateLimitingInterval:    60,
		RateLimitingLimit:       100,
		RateLimitingTechnique:   "fixed",
	}
}

func TestObserve(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		client *mockClient
		args   args
		want   want
	}{
		"ErrNotGateway": {
			reason: "An error should be returned if the managed resource is not a Gateway",
			client: &mockClient{},
			args:   args{mg: nil},
			want:   want{err: errors.New(errNotGateway)},
		},
		"NoExternalName": {
			reason: "The resource does not exist when no external name is set",
			client: &mockClient{},
			args:   args{mg: gateway()},
			want:   want{o: managed.ExternalObservation{ResourceExists: false}},
		},
		"NotFound": {
			reason: "The resource does not exist when the API reports it missing",
			client: &mockClient{
				MockGet: func(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error) {
					return aigatewayclient.Gateway{}, errors.Wrap(&cloudflare.Error{StatusCode: http.StatusNotFound}, "get")
				},
			},
			args: args{mg: gateway(withExternalName(externalName))},
			want: want{o: managed.ExternalObservation{ResourceExists: false}},
		},
		"LookupError": {
			reason: "Errors other than not found are returned",
			client: &mockClient{
				MockGet: func(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error) {
					return aigatewayclient.Gateway{}, errBoom
				},
			},
			args: args{mg: gateway(withExternalName(externalName))},
			want: want{err: errors.Wrap(errBoom, errGatewayLookup)},
		},
		"UpToDate": {
			reason: "The resource is up to date when the remote settings match the spec",
			client: &mockClient{
				MockGet: func(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error) {
					return remoteGateway(), nil
				},
			},
			args: args{mg: gateway(withExternalName(externalName))},
			want: want{o: managed.ExternalObservation{
				ResourceExists:    true,
				ResourceUpToDate:  true,
				ConnectionDetails: managed.ConnectionDetails{},
			}},
		},
		"NeedsUpdate": {
			reason: "The resource needs an update when a required setting differs from the remote value",
			client: &mockClient{
				MockGet: func(ctx context.Context, accountID, gatewayID string) (aigatewayclient.Gateway, error) {
					return remoteGateway(), nil
				},
			},
			args: args{mg: gateway(withExternalName(externalName), withCacheTTL(60))},
			want: want{o: managed.ExternalObservation{
				ResourceExists:    true,
				ResourceUpToDate:  false,
				ConnectionDetails: managed.ConnectionDetails{},
			}},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &gatewayExternal{client: tc.client}
			got, err := e.Observe(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nObserve(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nObserve(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	cases := map[string]struct {
		reason      string
		client      *mockClient
		mg          resource.Managed
		wantErr     error
		wantExtName string
	}{
		"ErrNotGateway": {
			reason:  "An error should be returned if the managed resource is not a Gateway",
			client:  &mockClient{},
			mg:      nil,
			wantErr: errors.New(errNotGateway),
		},
		"CreateError": {
			reason: "Errors from the API are returned",
			client: &mockClient{
				MockCreate: func(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
					return aigatewayclient.Gateway{}, errBoom
				},
			},
			mg:      gateway(),
			wantErr: errors.Wrap(errBoom, errGatewayCreate),
		},
		"Success": {
			reason: "The gateway identifier from the API becomes the external name",
			client: &mockClient{
				MockCreate: func(ctx context.Context, accountID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
					return aigatewayclient.Gateway{ID: testGatewayID}, nil
				},
			},
			mg:          gateway(),
			wantExtName: externalName,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &gatewayExternal{client: tc.client}
			_, err := e.Create(context.Background(), tc.mg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if tc.wantExtName != "" {
				if got := meta.GetExternalName(tc.mg.(*v1beta1.Gateway)); got != tc.wantExtName {
					t.Errorf("\n%s\nCreate(...): external name = %q, want %q", tc.reason, got, tc.wantExtName)
				}
			}
		})
	}
}

func TestConnect(t *testing.T) {
	const credsKey = "creds"

	// withCredentials returns a mock kube client whose ProviderConfig points at
	// a secret containing creds as the JSON credentials blob.
	withCredentials := func(creds string) *test.MockClient {
		return &test.MockClient{
			MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
				switch o := obj.(type) {
				case *pcv1beta1.ProviderConfig:
					o.Spec.Credentials.Source = xpv1.CredentialsSourceSecret
					o.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
						SecretReference: xpv1.SecretReference{Name: "creds", Namespace: "crossplane-system"},
						Key:             credsKey,
					}
				case *corev1.Secret:
					o.Data = map[string][]byte{credsKey: []byte(creds)}
				}
				return nil
			}),
		}
	}

	cases := map[string]struct {
		reason    string
		kube      client.Client
		mg        resource.Managed
		wantErr   error
		wantToken string
	}{
		"ErrNotGateway": {
			reason:  "An error should be returned if the managed resource is not a Gateway",
			kube:    &test.MockClient{},
			mg:      nil,
			wantErr: errors.New(errNotGateway),
		},
		"ErrGetCredentials": {
			reason: "An error should be returned if the ProviderConfig cannot be read",
			kube: &test.MockClient{
				MockGet: test.NewMockGetFn(errBoom),
			},
			mg:      gateway(withProviderConfig()),
			wantErr: errors.Wrap(errors.Wrap(errBoom, "cannot get ProviderConfig"), errGetCreds),
		},
		"ErrAPIKeyOnly": {
			reason:  "Credentials that carry no API token should be rejected",
			kube:    withCredentials(`{"apiKey":"key-123","email":"user@example.com"}`),
			mg:      gateway(withProviderConfig()),
			wantErr: errors.New(errNoAPIToken),
		},
		"Success": {
			reason:    "A Gateway with API token credentials should connect with that token",
			kube:      withCredentials(`{"token":"token-123"}`),
			mg:        gateway(withProviderConfig()),
			wantToken: "token-123",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var gotToken string
			c := &gatewayConnector{
				kube: tc.kube,
				newClientFn: func(token string) aigatewayclient.Client {
					gotToken = token
					return &mockClient{}
				},
			}

			_, err := c.Connect(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nConnect(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if gotToken != tc.wantToken {
				t.Errorf("\n%s\nConnect(...): token = %q, want %q", tc.reason, gotToken, tc.wantToken)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := map[string]struct {
		reason  string
		client  *mockClient
		mg      resource.Managed
		wantErr error
	}{
		"ErrNotGateway": {
			reason:  "An error should be returned if the managed resource is not a Gateway",
			client:  &mockClient{},
			mg:      nil,
			wantErr: errors.New(errNotGateway),
		},
		"UpdateError": {
			reason: "Errors from the API are returned",
			client: &mockClient{
				MockUpdate: func(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
					return aigatewayclient.Gateway{}, errBoom
				},
			},
			mg:      gateway(withExternalName(externalName)),
			wantErr: errors.Wrap(errBoom, errGatewayUpdate),
		},
		"Success": {
			reason: "The update should target the external name with the desired parameters",
			client: &mockClient{
				MockUpdate: func(ctx context.Context, accountID, gatewayID string, p v1beta1.GatewayParameters) (aigatewayclient.Gateway, error) {
					if accountID != testAccountID || gatewayID != externalName || p.CacheTTL != 60 {
						return aigatewayclient.Gateway{}, errBoom
					}
					return aigatewayclient.Gateway{ID: externalName}, nil
				},
			},
			mg: gateway(withExternalName(externalName), withCacheTTL(60)),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &gatewayExternal{client: tc.client}
			_, err := e.Update(context.Background(), tc.mg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := map[string]struct {
		reason  string
		client  *mockClient
		wantErr error
	}{
		"DeleteError": {
			reason: "Errors other than not found are returned",
			client: &mockClient{
				MockDelete: func(ctx context.Context, accountID, gatewayID string) error {
					return errBoom
				},
			},
			wantErr: errors.Wrap(errBoom, errGatewayDelete),
		},
		"AlreadyGone": {
			reason: "A gateway that no longer exists is treated as deleted",
			client: &mockClient{
				MockDelete: func(ctx context.Context, accountID, gatewayID string) error {
					return errors.Wrap(&cloudflare.Error{StatusCode: http.StatusNotFound}, "delete")
				},
			},
		},
		"Success": {
			reason: "A successful delete returns no error",
			client: &mockClient{
				MockDelete: func(ctx context.Context, accountID, gatewayID string) error {
					if accountID != testAccountID || gatewayID != externalName {
						return errBoom
					}
					return nil
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &gatewayExternal{client: tc.client}
			_, err := e.Delete(context.Background(), gateway(withExternalName(externalName)))

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}
