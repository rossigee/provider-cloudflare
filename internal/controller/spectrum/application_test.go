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

package spectrum

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/spectrum/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	applications "github.com/rossigee/provider-cloudflare/internal/clients/spectrum"

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

type mockApplicationClient struct {
	MockSpectrumApplication       func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error)
	MockCreateSpectrumApplication func(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error)
	MockUpdateSpectrumApplication func(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error
	MockDeleteSpectrumApplication func(ctx context.Context, zoneID, applicationID string) error
}

func (m *mockApplicationClient) SpectrumApplication(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
	return m.MockSpectrumApplication(ctx, zoneID, applicationID)
}

func (m *mockApplicationClient) CreateSpectrumApplication(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
	return m.MockCreateSpectrumApplication(ctx, zoneID, params)
}

func (m *mockApplicationClient) UpdateSpectrumApplication(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error {
	return m.MockUpdateSpectrumApplication(ctx, zoneID, applicationID, params)
}

func (m *mockApplicationClient) DeleteSpectrumApplication(ctx context.Context, zoneID, applicationID string) error {
	return m.MockDeleteSpectrumApplication(ctx, zoneID, applicationID)
}

type applicationModifier func(*v1alpha1.Application)

func withZone(zone string) applicationModifier {
	return func(app *v1alpha1.Application) { app.Spec.ForProvider.Zone = &zone }
}

func withProtocol(protocol string) applicationModifier {
	return func(app *v1alpha1.Application) { app.Spec.ForProvider.Protocol = protocol }
}

func withExternalName(id string) applicationModifier {
	return func(app *v1alpha1.Application) {
		app.SetAnnotations(map[string]string{
			"crossplane.io/external-name": id,
		})
	}
}

func applicationCR(m ...applicationModifier) *v1alpha1.Application {
	app := &v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			ForProvider: v1alpha1.ApplicationParameters{
				Protocol: "TCP/80",
			},
		},
	}
	for _, f := range m {
		f(app)
	}
	return app
}


func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube                  client.Client
		newCloudflareClientFn func(cfg clients.Config) (applications.Client, error)
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
		"ErrNotApplication": {
			reason: "Should return an error if the managed resource is not an Application",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotApplication),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: applicationCR(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errClientConfig),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{
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
		client applications.Client
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
		"ErrNotApplication": {
			reason: "Should return an error if the managed resource is not an Application",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotApplication),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: applicationCR(withExternalName("test-app-id")),
			},
			want: want{
				err: errors.New(errApplicationNoZone),
			},
		},
		"ErrGetApplication": {
			reason: "Should return any error encountered getting the application",
			fields: fields{
				client: &mockApplicationClient{
					MockSpectrumApplication: func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
						return cloudflare.SpectrumApplication{}, errors.New("boom")
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errApplicationLookup),
			},
		},
		"ApplicationNotFound": {
			reason: "Should report that the application does not exist",
			fields: fields{
				client: &mockApplicationClient{
					MockSpectrumApplication: func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
						return cloudflare.SpectrumApplication{}, errors.New("Application not found")
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					withExternalName("test-app-id"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ApplicationExistsAndUpToDate": {
			reason: "Should report that the application exists and is up to date",
			fields: fields{
				client: &mockApplicationClient{
					MockSpectrumApplication: func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
						return cloudflare.SpectrumApplication{
							ID:       "test-app-id",
							Protocol: "TCP/80",
						}, nil
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					withProtocol("TCP/80"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: false,
					ResourceUpToDate:        true,
				},
			},
		},
		"NoExternalName": {
			reason: "Should report that the application does not exist when no external name is set",
			args: args{
				mg: applicationCR(withZone("test-zone-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
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
		client applications.Client
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
		"ErrNotApplication": {
			reason: "Should return an error if the managed resource is not an Application",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotApplication),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: applicationCR(),
			},
			want: want{
				err: errors.Wrap(errors.New(errApplicationNoZone), errApplicationCreation),
			},
		},
		"ErrCreateApplication": {
			reason: "Should return any error encountered creating the application",
			fields: fields{
				client: &mockApplicationClient{
					MockCreateSpectrumApplication: func(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
						return cloudflare.SpectrumApplication{}, errors.New("boom")
					},
				},
			},
			args: args{
				mg: applicationCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errApplicationCreation),
			},
		},
		"Success": {
			reason: "Should return no error when application is created successfully",
			fields: fields{
				client: &mockApplicationClient{
					MockCreateSpectrumApplication: func(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
						return cloudflare.SpectrumApplication{
							ID:       "test-app-id",
							Protocol: "TCP/80",
						}, nil
					},
				},
			},
			args: args{
				mg: applicationCR(withZone("test-zone-id")),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
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
		client applications.Client
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
		"ErrNotApplication": {
			reason: "Should return an error if the managed resource is not an Application",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotApplication),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: applicationCR(),
			},
			want: want{
				err: errors.Wrap(errors.New(errApplicationNoZone), errApplicationUpdate),
			},
		},
		"ErrNoExternalName": {
			reason: "Should return an error if no external name is set",
			args: args{
				mg: applicationCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.New(errApplicationUpdate),
			},
		},
		"ErrUpdateApplication": {
			reason: "Should return any error encountered updating the application",
			fields: fields{
				client: &mockApplicationClient{
					MockUpdateSpectrumApplication: func(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error {
						return errors.New("boom")
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errApplicationUpdate),
			},
		},
		"Success": {
			reason: "Should return no error when application is updated successfully",
			fields: fields{
				client: &mockApplicationClient{
					MockUpdateSpectrumApplication: func(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
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
		client applications.Client
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
		"ErrNotApplication": {
			reason: "Should return an error if the managed resource is not an Application",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotApplication),
			},
		},
		"ErrNoZone": {
			reason: "Should return an error if no zone is specified",
			args: args{
				mg: applicationCR(withExternalName("test-app-id")),
			},
			want: want{
				err: errors.Wrap(errors.New(errApplicationNoZone), errApplicationDeletion),
			},
		},
		"ErrNoExternalName": {
			reason: "Should return an error if no external name is set",
			args: args{
				mg: applicationCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.New(errApplicationDeletion),
			},
		},
		"ErrDeleteApplication": {
			reason: "Should return any error encountered deleting the application",
			fields: fields{
				client: &mockApplicationClient{
					MockDeleteSpectrumApplication: func(ctx context.Context, zoneID, applicationID string) error {
						return errors.New("boom")
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errApplicationDeletion),
			},
		},
		"Success": {
			reason: "Should return no error when application is deleted successfully",
			fields: fields{
				client: &mockApplicationClient{
					MockDeleteSpectrumApplication: func(ctx context.Context, zoneID, applicationID string) error {
						return nil
					},
				},
			},
			args: args{
				mg: applicationCR(
					withZone("test-zone-id"),
					func(app *v1alpha1.Application) {
						app.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-app-id",
						})
					},
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{client: tc.fields.client}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}