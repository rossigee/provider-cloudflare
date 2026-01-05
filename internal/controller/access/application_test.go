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

package access

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/access/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/access/application"

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

// mockAccessApplicationAPI mocks the application.AccessApplicationAPI interface
type mockAccessApplicationAPI struct {
	MockListAccessApplications func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListAccessApplicationsParams) ([]cloudflare.AccessApplication, *cloudflare.ResultInfo, error)
	MockGetAccessApplication   func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error)
	MockCreateAccessApplication func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error)
	MockUpdateAccessApplication func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error)
	MockDeleteAccessApplication func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error
}

func (m *mockAccessApplicationAPI) ListAccessApplications(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListAccessApplicationsParams) ([]cloudflare.AccessApplication, *cloudflare.ResultInfo, error) {
	if m.MockListAccessApplications != nil {
		return m.MockListAccessApplications(ctx, rc, params)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockAccessApplicationAPI) GetAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error) {
	if m.MockGetAccessApplication != nil {
		return m.MockGetAccessApplication(ctx, rc, applicationID)
	}
	return cloudflare.AccessApplication{}, errors.New("cannot get access application")
}

func (m *mockAccessApplicationAPI) CreateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error) {
	if m.MockCreateAccessApplication != nil {
		return m.MockCreateAccessApplication(ctx, rc, params)
	}
	return cloudflare.AccessApplication{}, errors.New("cannot create access application")
}

func (m *mockAccessApplicationAPI) UpdateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error) {
	if m.MockUpdateAccessApplication != nil {
		return m.MockUpdateAccessApplication(ctx, rc, params)
	}
	return cloudflare.AccessApplication{}, errors.New("cannot update access application")
}

func (m *mockAccessApplicationAPI) DeleteAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error {
	if m.MockDeleteAccessApplication != nil {
		return m.MockDeleteAccessApplication(ctx, rc, applicationID)
	}
	return errors.New("cannot delete access application")
}

// Helper to create a CloudflareAccessApplicationClient with a mocked API
func newMockAccessApplicationClient(api application.AccessApplicationAPI) *application.CloudflareAccessApplicationClient {
	return application.NewClient(api)
}

type accessApplicationModifier func(*v1beta1.AccessApplication)

func withAccessAppID(id string) accessApplicationModifier {
	return func(aa *v1beta1.AccessApplication) {
		aa.Status.AtProvider.ID = id
		rtmeta.SetExternalName(aa, id)
	}
}

func accessApplication(m ...accessApplicationModifier) *v1beta1.AccessApplication {
	aa := &v1beta1.AccessApplication{
		Spec: v1beta1.AccessApplicationSpec{
			ForProvider: v1beta1.AccessApplicationParameters{
				AccountID: "test-account-id",
				Name:      "test-app",
				Domain:    "test.example.com",
				Type:      "self_hosted",
			},
		},
	}
	for _, f := range m {
		f(aa)
	}
	return aa
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube        client.Client
		newServiceFn func(api *cloudflare.API) *application.CloudflareAccessApplicationClient
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
		"ErrNotAccessApplication": {
			reason: "Should return an error if the managed resource is not an AccessApplication",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotAccessApplication),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: accessApplication(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errGetCreds),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &accessApplicationConnector{
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
		service *application.CloudflareAccessApplicationClient
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
		"ErrNotAccessApplication": {
			reason: "Should return an error if the managed resource is not an AccessApplication",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotAccessApplication),
			},
		},
		"ErrGetAccessApplication": {
			reason: "Should return any error encountered getting the access application",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockGetAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot get access application"), "cannot get external resource"),
			},
		},
		"AccessApplicationNotFound": {
			reason: "Should report that the access application does not exist",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockGetAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{}, errors.New("access application not found")
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"AccessApplicationExistsAndUpToDate": {
			reason: "Should report that the access application exists and is up to date",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockGetAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{
							ID:   "test-app-id",
							Name: "test-app",
							Domain: "test.example.com",
							Type: "self_hosted",
						}, nil
					},
				}),
			},
			args: args{
				mg: accessApplication(
					withAccessAppID("test-app-id"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"AccessApplicationExistsButOutOfDate": {
			reason: "Should report that the access application exists but is not up to date",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockGetAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{
							ID:   "test-app-id",
							Name: "test-app",
							Domain: "test.example.com",
							Type: "self_hosted",
						}, nil
					},
				}),
			},
			args: args{
				mg: accessApplication(
					withAccessAppID("test-app-id"),
					func(aa *v1beta1.AccessApplication) {
						aa.Spec.ForProvider.Name = "updated-name"
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
			e := &accessApplicationExternal{service: tc.fields.service}
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
		service *application.CloudflareAccessApplicationClient
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
		"ErrNotAccessApplication": {
			reason: "Should return an error if the managed resource is not an AccessApplication",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotAccessApplication),
			},
		},
		"ErrCreateAccessApplication": {
			reason: "Should return any error encountered creating the access application",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockCreateAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: accessApplication(),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot create access application"), "cannot create external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when access application is created successfully",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockCreateAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{
							ID:   "new-app-id",
							Name: params.Name,
							Domain: params.Domain,
							Type: params.Type,
						}, nil
					},
				}),
			},
			args: args{
				mg: accessApplication(),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &accessApplicationExternal{service: tc.fields.service}
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
		service *application.CloudflareAccessApplicationClient
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
		"ErrNotAccessApplication": {
			reason: "Should return an error if the managed resource is not an AccessApplication",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotAccessApplication),
			},
		},
		"ErrUpdateAccessApplication": {
			reason: "Should return any error encountered updating the access application",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockUpdateAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot update access application"), "cannot update external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when access application is updated successfully",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockUpdateAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error) {
						return cloudflare.AccessApplication{
							ID:   "test-app-id",
							Name: params.Name,
						}, nil
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &accessApplicationExternal{service: tc.fields.service}
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
		service *application.CloudflareAccessApplicationClient
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
		"ErrNotAccessApplication": {
			reason: "Should return an error if the managed resource is not an AccessApplication",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotAccessApplication),
			},
		},
		"ErrDeleteAccessApplication": {
			reason: "Should return any error encountered deleting the access application",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockDeleteAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error {
						return errors.New("api error")
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("api error"), "cannot delete access application"),
			},
		},
		"Success": {
			reason: "Should return no error when access application is deleted successfully",
			fields: fields{
				service: newMockAccessApplicationClient(&mockAccessApplicationAPI{
					MockDeleteAccessApplication: func(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error {
						return nil
					},
				}),
			},
			args: args{
				mg: accessApplication(withAccessAppID("test-app-id")),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &accessApplicationExternal{service: tc.fields.service}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}