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

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

type mockTracker struct{}

func (m *mockTracker) Track(ctx context.Context, mg resource.Managed) error {
	return nil
}

type loadbalancerModifier func(*v1alpha1.LoadBalancer)


func withZone(zone string) loadbalancerModifier {
	return func(lb *v1alpha1.LoadBalancer) { lb.Spec.ForProvider.Zone = zone }
}

func withName(name string) loadbalancerModifier {
	return func(lb *v1alpha1.LoadBalancer) { lb.Spec.ForProvider.Name = &name }
}

func withID(id string) loadbalancerModifier {
	return func(lb *v1alpha1.LoadBalancer) { lb.Status.AtProvider.ID = id }
}

func loadbalancer(m ...loadbalancerModifier) *v1alpha1.LoadBalancer {
	cr := &v1alpha1.LoadBalancer{}
	for _, f := range m {
		f(cr)
	}
	return cr
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube      client.Client
		usage     resource.Tracker
		newClient func(cfg clients.Config, hc *http.Client) (loadbalancing.LoadBalancerClient, error)
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
		"ErrNotLoadBalancer": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancer",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotLoadBalancer),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube:  mc,
				usage: &mockTracker{},
			},
			args: args{
				mg: &v1alpha1.LoadBalancer{
					Spec: v1alpha1.LoadBalancerSpec{
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
				newClient: func(cfg clients.Config, hc *http.Client) (loadbalancing.LoadBalancerClient, error) {
					return &fake.MockLoadBalancerClient{}, nil
				},
			},
			args: args{
				mg: &v1alpha1.LoadBalancer{
					Spec: v1alpha1.LoadBalancerSpec{
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
			c := &connector{kube: tc.fields.kube, usage: tc.fields.usage, newServiceFn: tc.fields.newClient}
			_, err := c.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nc.Connect(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.LoadBalancerClient
		kube    client.Client
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
		"ErrNotLoadBalancer": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancer",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotLoadBalancer),
			},
		},
		"ErrNoLoadBalancer": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				service: &fake.MockLoadBalancerClient{},
			},
			args: args{
				mg: &v1alpha1.LoadBalancer{},
			},
			want: want{
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrLoadBalancerLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockGetLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withID("1234beef"),
					withZone("example.com"),
				),
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, "failed to get load balancer from Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return ResourceExists: true and no error when a load balancer is found",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockGetLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return &cloudflare.LoadBalancer{
							ID: lbID,
						}, nil
					},
				},
			},
			args: args{
				mg: loadbalancer(withID("1234beef"), withZone("example.com")),
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
			e := external{service: tc.fields.service, kube: tc.fields.kube}
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

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.LoadBalancerClient
		kube    client.Client
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
		"ErrNotLoadBalancer": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancer",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotLoadBalancer),
			},
		},
		"ErrLoadBalancerCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockCreateLoadBalancer: func(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withZone("example.com"),
					withName("test-lb"),
				),
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, "failed to create load balancer in Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return ExternalNameAssigned: true and no error when a load balancer is created",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockCreateLoadBalancer: func(ctx context.Context, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return &cloudflare.LoadBalancer{
							ID:   "1234beef",
							Name: *params.Name,
						}, nil
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withZone("example.com"),
					withName("test-lb"),
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
			e := external{service: tc.fields.service, kube: tc.fields.kube}
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

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.LoadBalancerClient
		kube    client.Client
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
		"ErrNotLoadBalancer": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancer",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotLoadBalancer),
			},
		},
		"ErrLoadBalancerUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockUpdateLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withID("1234beef"),
					withZone("example.com"),
					withName("test-lb"),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, "failed to update load balancer in Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return no error when a load balancer is updated",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockUpdateLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) (*cloudflare.LoadBalancer, error) {
						return &cloudflare.LoadBalancer{}, nil
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withID("1234beef"),
					withZone("example.com"),
					withName("test-lb"),
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
			e := external{service: tc.fields.service, kube: tc.fields.kube}
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

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		service loadbalancing.LoadBalancerClient
		kube    client.Client
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
		"ErrNotLoadBalancer": {
			reason: "An error should be returned if the managed resource is not a *LoadBalancer",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotLoadBalancer),
			},
		},
		"ErrLoadBalancerDelete": {
			reason: "We should return any errors during the delete process",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockDeleteLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withID("1234beef"),
					withZone("example.com"),
				),
			},
			want: want{
				err: errors.Wrap(errBoom, "failed to delete load balancer from Cloudflare API"),
			},
		},
		"Success": {
			reason: "We should return no error when a load balancer is deleted",
			fields: fields{
				service: &fake.MockLoadBalancerClient{
					MockDeleteLoadBalancer: func(ctx context.Context, lbID string, params v1alpha1.LoadBalancerParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: loadbalancer(
					withID("1234beef"),
					withZone("example.com"),
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{service: tc.fields.service, kube: tc.fields.kube}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}