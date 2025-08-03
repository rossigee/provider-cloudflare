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

package loadbalancing

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
	pcv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/loadbalancing"
	"github.com/rossigee/provider-cloudflare/internal/clients/loadbalancing/fake"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

type poolModifier func(*v1alpha1.LoadBalancerPool)


func withPoolAccount(account string) poolModifier {
	return func(pool *v1alpha1.LoadBalancerPool) { pool.Spec.ForProvider.Account = &account }
}

func withPoolName(name string) poolModifier {
	return func(pool *v1alpha1.LoadBalancerPool) { pool.Spec.ForProvider.Name = name }
}

func withPoolID(id string) poolModifier {
	return func(pool *v1alpha1.LoadBalancerPool) { pool.Status.AtProvider.ID = id }
}

func pool(m ...poolModifier) *v1alpha1.LoadBalancerPool {
	cr := &v1alpha1.LoadBalancerPool{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestPoolConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube      client.Client
		usage     resource.Tracker
		newClient func(cfg clients.Config, hc *http.Client) (loadbalancing.PoolClient, error)
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
		"ErrNotLoadBalancerPool": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancerPool",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotPool),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube:  mc,
				usage: &mockTracker{},
			},
			args: args{
				mg: &v1alpha1.LoadBalancerPool{
					Spec: v1alpha1.LoadBalancerPoolSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "test-config",
							},
						},
					},
				},
			},
			want: errors.Wrap(errors.New("cannot get ProviderConfig: no extraction handler registered for source: "), errGetCreds),
		},
		"ConnectReturnOK": {
			reason: "Connect should return no error when passed the correct values",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						switch o := obj.(type) {
						case *pcv1alpha1.ProviderConfig:
							o.Spec.Credentials.Source = "Secret"
							o.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
								Key: "creds",
							}
						case *corev1.Secret:
							o.Data = map[string][]byte{
								"creds": []byte("{\"token\":\"foo\"}"),
							}
						}
						return nil
					}),
				},
				usage: &mockTracker{},
				newClient: func(cfg clients.Config, hc *http.Client) (loadbalancing.PoolClient, error) {
					return &fake.MockPoolClient{}, nil
				},
			},
			args: args{
				mg: &v1alpha1.LoadBalancerPool{
					Spec: v1alpha1.LoadBalancerPoolSpec{
						ResourceSpec: xpv1.ResourceSpec{
							ProviderConfigReference: &xpv1.Reference{
								Name: "blah",
							},
						},
					},
				},
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &poolConnector{kube: tc.fields.kube, usage: tc.fields.usage, newServiceFn: tc.fields.newClient}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nc.Connect(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestPoolObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.PoolClient
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
		"ErrNotLoadBalancerPool": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancerPool",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotPool),
			},
		},
		"ErrNoLoadBalancerPool": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				service: &fake.MockPoolClient{},
			},
			args: args{
				mg: &v1alpha1.LoadBalancerPool{},
			},
			want: want{
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrLoadBalancerPoolLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				service: &fake.MockPoolClient{
					MockGetPool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: pool(
					withPoolID("1234beef"),
					withPoolAccount("example-account"),
				),
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, "failed to get load balancer pool from Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return ResourceExists: true and no error when a load balancer pool is found",
			fields: fields{
				service: &fake.MockPoolClient{
					MockGetPool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return &cloudflare.LoadBalancerPool{
							ID: poolID,
						}, nil
					},
				},
			},
			args: args{
				mg: pool(withPoolID("1234beef"), withPoolAccount("example-account")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: true,
					ConnectionDetails:       managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := poolExternal{service: tc.fields.service}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestPoolCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.PoolClient
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
		"ErrNotLoadBalancerPool": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancerPool",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotPool),
			},
		},
		"ErrLoadBalancerPoolCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				service: &fake.MockPoolClient{
					MockCreatePool: func(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: pool(
					withPoolAccount("example-account"),
					withPoolName("test-pool"),
				),
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, "failed to create load balancer pool in Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return ExternalNameAssigned: true and no error when a load balancer pool is created",
			fields: fields{
				service: &fake.MockPoolClient{
					MockCreatePool: func(ctx context.Context, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return &cloudflare.LoadBalancerPool{
							ID:   "1234beef",
							Name: params.Name,
						}, nil
					},
				},
			},
			args: args{
				mg: pool(
					withPoolAccount("example-account"),
					withPoolName("test-pool"),
				),
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := poolExternal{service: tc.fields.service}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestPoolUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.PoolClient
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
		"ErrNotLoadBalancerPool": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancerPool",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotPool),
			},
		},
		"ErrLoadBalancerPoolUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				service: &fake.MockPoolClient{
					MockUpdatePool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: pool(
					withPoolID("1234beef"),
					withPoolAccount("example-account"),
					withPoolName("test-pool"),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, "failed to update load balancer pool in Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return no error when a load balancer pool is updated",
			fields: fields{
				service: &fake.MockPoolClient{
					MockUpdatePool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) (*cloudflare.LoadBalancerPool, error) {
						return &cloudflare.LoadBalancerPool{}, nil
					},
				},
			},
			args: args{
				mg: pool(
					withPoolID("1234beef"),
					withPoolAccount("example-account"),
					withPoolName("test-pool"),
				),
			},
			want: want{
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := poolExternal{service: tc.fields.service}
			got, err := e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestPoolDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.PoolClient
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
		"ErrNotLoadBalancerPool": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancerPool",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotPool),
			},
		},
		"ErrLoadBalancerPoolDelete": {
			reason: "We should return any errors during the delete process",
			fields: fields{
				service: &fake.MockPoolClient{
					MockDeletePool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: pool(
					withPoolID("1234beef"),
					withPoolAccount("example-account"),
				),
			},
			want: want{
				err: errors.Wrap(errBoom, "failed to delete load balancer pool from Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return no error when a load balancer pool is deleted",
			fields: fields{
				service: &fake.MockPoolClient{
					MockDeletePool: func(ctx context.Context, poolID string, params v1alpha1.LoadBalancerPoolParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: pool(
					withPoolID("1234beef"),
					withPoolAccount("example-account"),
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := poolExternal{service: tc.fields.service}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}