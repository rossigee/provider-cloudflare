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

package workers

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	pcv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	workers "github.com/rossigee/provider-cloudflare/internal/clients/workers"
	"github.com/rossigee/provider-cloudflare/internal/clients/workers/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	rtfake "github.com/crossplane/crossplane-runtime/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

type routeModifier func(*v1alpha1.Route)

func withExternalName(name string) routeModifier {
	return func(r *v1alpha1.Route) { meta.SetExternalName(r, name) }
}

func withZone(zoneID string) routeModifier {
	return func(r *v1alpha1.Route) { 
		if zoneID == "" {
			r.Spec.ForProvider.Zone = nil
		} else {
			r.Spec.ForProvider.Zone = &zoneID
		}
	}
}

func withConditions(c ...xpv1.Condition) routeModifier {
	return func(r *v1alpha1.Route) { r.Status.Conditions = c }
}


func route(m ...routeModifier) *v1alpha1.Route {
	cr := &v1alpha1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-route",
		},
		Spec: v1alpha1.RouteSpec{
			ForProvider: v1alpha1.RouteParameters{
				Pattern: "example.com/*",
				Zone:    stringPtr("test-zone-id"),
				Script:  stringPtr("test-worker"),
			},
		},
	}

	for _, f := range m {
		f(cr)
	}

	return cr
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	_, errGetProviderConfig := clients.GetConfig(context.Background(), mc, &rtfake.Managed{})

	type fields struct {
		kube      client.Client
		newClient func(cfg clients.Config, hc *http.Client) (workers.Client, error)
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
		"ErrNotRoute": {
			reason: "An error should be returned if the managed resource is not a *Route",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotRoute),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						ResourceSpec: xpv1.ResourceSpec{},
					},
				},
			},
			want: errors.Wrap(errGetProviderConfig, errClientConfig),
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
								"creds": []byte(`{"APIKey":"foo","Email":"foo@bar.com"}`),
							}
						}
						return nil
					}),
				},
				newClient: workers.NewClient,
			},
			args: args{
				mg: &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
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
			nc := func(cfg clients.Config) (workers.Client, error) {
				return tc.fields.newClient(cfg, nil)
			}
			e := &connector{kube: tc.fields.kube, newCloudflareClientFn: nc}
			_, err := e.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Connect(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client workers.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		cr  resource.Managed
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotRoute": {
			reason: "An error should be returned if the managed resource is not a *Route",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRoute),
			},
		},
		"ErrNoRoute": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(),
			},
			want: want{
				cr: route(),
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrRouteNoZone": {
			reason: "We should return an error if the Route does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone(""),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone(""),
				),
				o:   managed.ExternalObservation{},
				err: errors.New(errRouteNoZone),
			},
		},
		"ErrRouteLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				client: &fake.MockClient{
					MockWorkerRoute: func(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{}, errBoom
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, errRouteLookup),
			},
		},
		"RouteNotFound": {
			reason: "We should return ResourceExists: false when route is not found",
			fields: fields{
				client: &fake.MockClient{
					MockWorkerRoute: func(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{}, errors.New("Worker Route not found")
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"Success": {
			reason: "We should return ResourceExists: true and no error when a Route is found",
			fields: fields{
				client: &fake.MockClient{
					MockWorkerRoute: func(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{
							ID:         routeID,
							Pattern:    "example.com/*",
							ScriptName: "test-worker",
						}, nil
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
					withConditions(xpv1.Available()),
				),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessOutdated": {
			reason: "We should return ResourceExists: true and ResourceUpToDate: false when route differs",
			fields: fields{
				client: &fake.MockClient{
					MockWorkerRoute: func(ctx context.Context, zoneID, routeID string) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{
							ID:         routeID,
							Pattern:    "different.com/*",
							ScriptName: "different-worker",
						}, nil
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
					withConditions(xpv1.Available()),
				),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
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
		client workers.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		cr  resource.Managed
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotRoute": {
			reason: "An error should be returned if the managed resource is not a *Route",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRoute),
			},
		},
		"ErrRouteNoZone": {
			reason: "We should return an error if the Route does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(withZone("")),
			},
			want: want{
				cr:  route(withZone("")),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errors.New(errRouteNoZone), errRouteCreation),
			},
		},
		"ErrRouteCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				client: &fake.MockClient{
					MockCreateWorkerRoute: func(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{}, errBoom
					},
				},
			},
			args: args{
				mg: route(),
			},
			want: want{
				cr:  route(withConditions(xpv1.Creating())),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, errRouteCreation),
			},
		},
		"Success": {
			reason: "We should return no error when a Route is created",
			fields: fields{
				client: &fake.MockClient{
					MockCreateWorkerRoute: func(ctx context.Context, zoneID string, params *v1alpha1.RouteParameters) (cloudflare.WorkerRoute, error) {
						return cloudflare.WorkerRoute{
							ID:         "new-route-id",
							Pattern:    params.Pattern,
							ScriptName: *params.Script,
						}, nil
					},
				},
			},
			args: args{
				mg: route(),
			},
			want: want{
				cr: route(
					withExternalName("new-route-id"),
					withConditions(xpv1.Creating()),
				),
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
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
		client workers.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		cr  resource.Managed
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotRoute": {
			reason: "An error should be returned if the managed resource is not a *Route",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRoute),
			},
		},
		"ErrNoExternalName": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(),
			},
			want: want{
				cr:  route(),
				o:   managed.ExternalUpdate{},
				err: errors.New(errRouteUpdate),
			},
		},
		"ErrRouteNoZone": {
			reason: "We should return an error if the Route does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone(""),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone(""),
				),
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errors.New(errRouteNoZone), errRouteUpdate),
			},
		},
		"ErrRouteUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateWorkerRoute: func(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, errRouteUpdate),
			},
		},
		"Success": {
			reason: "We should return no error when a Route is updated",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateWorkerRoute: func(ctx context.Context, zoneID, routeID string, params *v1alpha1.RouteParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			got, err := e.Update(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
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
		client workers.Client
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrNotRoute": {
			reason: "An error should be returned if the managed resource is not a *Route",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRoute),
			},
		},
		"ErrRouteNoZone": {
			reason: "We should return an error if the Route does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(withZone("")),
			},
			want: want{
				cr:  route(withZone("")),
				err: errors.Wrap(errors.New(errRouteNoZone), errRouteDeletion),
			},
		},
		"ErrNoExternalName": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: route(),
			},
			want: want{
				cr:  route(),
				err: errors.New(errRouteDeletion),
			},
		},
		"ErrRouteDelete": {
			reason: "We should return any errors during the delete process",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteWorkerRoute: func(ctx context.Context, zoneID, routeID string) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
				err: errors.Wrap(errBoom, errRouteDeletion),
			},
		},
		"Success": {
			reason: "We should return no error when a Route is deleted",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteWorkerRoute: func(ctx context.Context, zoneID, routeID string) error {
						return nil
					},
				},
			},
			args: args{
				mg: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: route(
					withExternalName("test-route-id"),
					withZone("test-zone-id"),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}