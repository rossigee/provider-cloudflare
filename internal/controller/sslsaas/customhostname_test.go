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
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1alpha1"
	pcv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	customhostname "github.com/rossigee/provider-cloudflare/internal/clients/sslsaas/customhostname"
	"github.com/rossigee/provider-cloudflare/internal/clients/sslsaas/customhostname/fake"

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

type customHostnameModifier func(*v1alpha1.CustomHostname)

func withExternalName(name string) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { meta.SetExternalName(ch, name) }
}

func withZone(zoneID string) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { 
		if zoneID == "" {
			ch.Spec.ForProvider.Zone = nil
		} else {
			ch.Spec.ForProvider.Zone = &zoneID
		}
	}
}

func withHostname(hostname string) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Spec.ForProvider.Hostname = hostname }
}

func withSSLMethod(method string) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { 
		if method == "" {
			ch.Spec.ForProvider.SSL.Method = nil
		} else {
			ch.Spec.ForProvider.SSL.Method = &method
		}
	}
}

func withSSLType(sslType string) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Spec.ForProvider.SSL.Type = &sslType }
}

func withConditions(c ...xpv1.Condition) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Status.Conditions = c }
}

func withSpec(s v1alpha1.CustomHostnameSpec) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Spec = s }
}

func withStatus(s v1alpha1.CustomHostnameStatus) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Status = s }
}

func withAtProvider(obs v1alpha1.CustomHostnameObservation) customHostnameModifier {
	return func(ch *v1alpha1.CustomHostname) { ch.Status.AtProvider = obs }
}

func customHostname(m ...customHostnameModifier) *v1alpha1.CustomHostname {
	cr := &v1alpha1.CustomHostname{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-custom-hostname",
		},
		Spec: v1alpha1.CustomHostnameSpec{
			ForProvider: v1alpha1.CustomHostnameParameters{
				Hostname: "example.com",
				Zone:     stringPtr("test-zone-id"),
				SSL: v1alpha1.CustomHostnameSSL{
					Method: stringPtr("http"),
					Type:   stringPtr("dv"),
				},
			},
		},
	}

	for _, f := range m {
		f(cr)
	}

	return cr
}

func TestCustomHostnameConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	_, errGetProviderConfig := clients.GetConfig(context.Background(), mc, &rtfake.Managed{})

	type fields struct {
		kube      client.Client
		newClient func(cfg clients.Config, hc *http.Client) (customhostname.Client, error)
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
		"ErrNotCustomHostname": {
			reason: "An error should be returned if the managed resource is not a *CustomHostname",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotCustomHostname),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: &v1alpha1.CustomHostname{
					Spec: v1alpha1.CustomHostnameSpec{
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
				newClient: customhostname.NewClient,
			},
			args: args{
				mg: &v1alpha1.CustomHostname{
					Spec: v1alpha1.CustomHostnameSpec{
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
			nc := func(cfg clients.Config) (customhostname.Client, error) {
				return tc.fields.newClient(cfg, nil)
			}
			e := &customHostnameConnector{kube: tc.fields.kube, newCloudflareClientFn: nc}
			_, err := e.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Connect(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCustomHostnameObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client customhostname.Client
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
		"ErrNotCustomHostname": {
			reason: "An error should be returned if the managed resource is not a *CustomHostname",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCustomHostname),
			},
		},
		"ErrNoCustomHostname": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(),
			},
			want: want{
				cr: customHostname(),
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrCustomHostnameNoZone": {
			reason: "We should return an error if the CustomHostname does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone(""),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone(""),
				),
				o:   managed.ExternalObservation{},
				err: errors.New(errCustomHostnameNoZone),
			},
		},
		"ErrCustomHostnameLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				client: &fake.MockClient{
					MockCustomHostname: func(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
						return cloudflare.CustomHostname{}, errBoom
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, errCustomHostnameLookup),
			},
		},
		"CustomHostnameNotFound": {
			reason: "We should return ResourceExists: false when custom hostname is not found",
			fields: fields{
				client: &fake.MockClient{
					MockCustomHostname: func(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
						return cloudflare.CustomHostname{}, errors.New("Custom Hostname not found")
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"Success": {
			reason: "We should return ResourceExists: true and no error when a CustomHostname is found",
			fields: fields{
				client: &fake.MockClient{
					MockCustomHostname: func(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
						return cloudflare.CustomHostname{
							ID:       hostnameID,
							Hostname: "example.com",
							Status:   cloudflare.ACTIVE,
						}, nil
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
					withAtProvider(v1alpha1.CustomHostnameObservation{
						Status: cloudflare.ACTIVE,
					}),
				),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessOutdated": {
			reason: "We should return ResourceExists: true and ResourceUpToDate: false when hostname differs",
			fields: fields{
				client: &fake.MockClient{
					MockCustomHostname: func(ctx context.Context, zoneID, hostnameID string) (cloudflare.CustomHostname, error) {
						return cloudflare.CustomHostname{
							ID:       hostnameID,
							Hostname: "different.com",
							Status:   cloudflare.ACTIVE,
						}, nil
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
					withAtProvider(v1alpha1.CustomHostnameObservation{
						Status: cloudflare.ACTIVE,
					}),
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
			e := customHostnameExternal{client: tc.fields.client}
			got, err := e.Observe(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
			// Verify AtProvider is set for successful cases  
			if tc.want.cr != nil {
				wantCH := tc.want.cr.(*v1alpha1.CustomHostname)
				actualCH := tc.args.mg.(*v1alpha1.CustomHostname)
				if wantCH.Status.AtProvider.Status != "" && actualCH.Status.AtProvider.Status != wantCH.Status.AtProvider.Status {
					t.Errorf("\n%s\nAtProvider.Status: want %s, got %s\n", tc.reason, wantCH.Status.AtProvider.Status, actualCH.Status.AtProvider.Status)
				}
			}
		})
	}
}

func TestCustomHostnameCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client customhostname.Client
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
		"ErrNotCustomHostname": {
			reason: "An error should be returned if the managed resource is not a *CustomHostname",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCustomHostname),
			},
		},
		"ErrCustomHostnameNoZone": {
			reason: "We should return an error if the CustomHostname does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(withZone("")),
			},
			want: want{
				cr:  customHostname(withZone("")),
				o:   managed.ExternalCreation{},
				err: errors.New(errCustomHostnameCreation),
			},
		},
		"ErrCustomHostnameNoSSLMethod": {
			reason: "We should return an error if the CustomHostname does not have SSL method",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(withSSLMethod("")),
			},
			want: want{
				cr:  customHostname(withSSLMethod("")),
				o:   managed.ExternalCreation{},
				err: errors.New(errCustomHostnameCreation),
			},
		},
		"ErrCustomHostnameCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				client: &fake.MockClient{
					MockCreateCustomHostname: func(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
						return &cloudflare.CustomHostnameResponse{}, errBoom
					},
				},
			},
			args: args{
				mg: customHostname(),
			},
			want: want{
				cr:  customHostname(withConditions(xpv1.Creating())),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, errCustomHostnameCreation),
			},
		},
		"Success": {
			reason: "We should return no error when a CustomHostname is created",
			fields: fields{
				client: &fake.MockClient{
					MockCreateCustomHostname: func(ctx context.Context, zoneID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
						return &cloudflare.CustomHostnameResponse{
							Result: cloudflare.CustomHostname{
								ID:       "new-hostname-id",
								Hostname: hostname.Hostname,
								Status:   cloudflare.PENDING,
							},
						}, nil
					},
				},
			},
			args: args{
				mg: customHostname(),
			},
			want: want{
				cr: customHostname(
					withExternalName("new-hostname-id"),
					withAtProvider(v1alpha1.CustomHostnameObservation{
						Status: cloudflare.PENDING,
					}),
				),
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := customHostnameExternal{client: tc.fields.client}
			got, err := e.Create(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
			// Verify external name and AtProvider status for successful cases
			if tc.want.cr != nil {
				wantCH := tc.want.cr.(*v1alpha1.CustomHostname)
				actualCH := tc.args.mg.(*v1alpha1.CustomHostname)
				if meta.GetExternalName(wantCH) != "" && meta.GetExternalName(actualCH) != meta.GetExternalName(wantCH) {
					t.Errorf("\n%s\nExternal name: want %s, got %s\n", tc.reason, meta.GetExternalName(wantCH), meta.GetExternalName(actualCH))
				}
				if wantCH.Status.AtProvider.Status != "" && actualCH.Status.AtProvider.Status != wantCH.Status.AtProvider.Status {
					t.Errorf("\n%s\nAtProvider.Status: want %s, got %s\n", tc.reason, wantCH.Status.AtProvider.Status, actualCH.Status.AtProvider.Status)
				}
			}
		})
	}
}

func TestCustomHostnameUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client customhostname.Client
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
		"ErrNotCustomHostname": {
			reason: "An error should be returned if the managed resource is not a *CustomHostname",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCustomHostname),
			},
		},
		"ErrNoExternalName": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(),
			},
			want: want{
				cr:  customHostname(),
				o:   managed.ExternalUpdate{},
				err: errors.New(errCustomHostnameUpdate),
			},
		},
		"ErrCustomHostnameNoZone": {
			reason: "We should return an error if the CustomHostname does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone(""),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone(""),
				),
				o:   managed.ExternalUpdate{},
				err: errors.New(errCustomHostnameUpdate),
			},
		},
		"ErrCustomHostnameUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateCustomHostname: func(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
						return &cloudflare.CustomHostnameResponse{}, errBoom
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, errCustomHostnameUpdate),
			},
		},
		"Success": {
			reason: "We should return no error when a CustomHostname is updated",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateCustomHostname: func(ctx context.Context, zoneID, hostnameID string, hostname cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error) {
						return &cloudflare.CustomHostnameResponse{
							Result: cloudflare.CustomHostname{
								ID:       hostnameID,
								Hostname: hostname.Hostname,
								Status:   cloudflare.ACTIVE,
							},
						}, nil
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := customHostnameExternal{client: tc.fields.client}
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

func TestCustomHostnameDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client customhostname.Client
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
		"ErrNotCustomHostname": {
			reason: "An error should be returned if the managed resource is not a *CustomHostname",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCustomHostname),
			},
		},
		"ErrCustomHostnameNoZone": {
			reason: "We should return an error if the CustomHostname does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(withZone("")),
			},
			want: want{
				cr:  customHostname(withZone("")),
				err: errors.New(errCustomHostnameDeletion),
			},
		},
		"ErrNoExternalName": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: customHostname(),
			},
			want: want{
				cr:  customHostname(),
				err: errors.New(errCustomHostnameDeletion),
			},
		},
		"ErrCustomHostnameDelete": {
			reason: "We should return any errors during the delete process",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteCustomHostname: func(ctx context.Context, zoneID, hostnameID string) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
				err: errors.Wrap(errBoom, errCustomHostnameDeletion),
			},
		},
		"Success": {
			reason: "We should return no error when a CustomHostname is deleted",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteCustomHostname: func(ctx context.Context, zoneID, hostnameID string) error {
						return nil
					},
				},
			},
			args: args{
				mg: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: customHostname(
					withExternalName("test-hostname-id"),
					withZone("test-zone-id"),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := customHostnameExternal{client: tc.fields.client}
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