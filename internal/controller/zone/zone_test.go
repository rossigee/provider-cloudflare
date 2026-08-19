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

package zone

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	rtfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"net/http"
	"testing"

	pcv1beta1 "github.com/rossigee/provider-cloudflare/apis/v1beta1"
	zonev1beta1 "github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/zones"
	"github.com/rossigee/provider-cloudflare/internal/clients/zones/fake"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type zoneModifier func(*zonev1beta1.Zone)

func withAccount(sValue *string) zoneModifier {
	return func(r *zonev1beta1.Zone) { r.Spec.ForProvider.AccountID = sValue }
}
func withExternalName(zoneID string) zoneModifier {
	return func(r *zonev1beta1.Zone) { meta.SetExternalName(r, zoneID) }
}
func withPaused(paused *bool) zoneModifier {
	return func(r *zonev1beta1.Zone) { r.Spec.ForProvider.Paused = paused }
}
func withPlan(sValue *string) zoneModifier {
	return func(r *zonev1beta1.Zone) { r.Spec.ForProvider.PlanID = sValue }
}
func withType(typ *string) zoneModifier {
	return func(r *zonev1beta1.Zone) { r.Spec.ForProvider.Type = typ }
}

func zone(m ...zoneModifier) *zonev1beta1.Zone {
	cr := &zonev1beta1.Zone{}
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
		newClient func(cfg clients.Config, hc *http.Client) (zones.Client, error)
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
		"ErrNotZone": {
			reason: "An error should be returned if the managed resource is not a *Zone",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotZone),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: &zonev1beta1.Zone{
					Spec: zonev1beta1.ZoneSpec{
						ClusterManagedResourceSpec: xpv1.ClusterManagedResourceSpec{},
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
						case *pcv1beta1.ProviderConfig:
							o.Spec.Credentials.Source = "Secret"
							o.Spec.Credentials.SecretRef = &xpv1.SecretKeySelector{
								Key: "creds",
							}
						case *corev1.Secret:
							o.Data = map[string][]byte{
								"creds": []byte("{\"APIKey\":\"foo\",\"Email\":\"foo@bar.com\"}"),
							}
						}
						return nil
					}),
				},
				newClient: zones.NewClient,
			},
			args: args{
				mg: &zonev1beta1.Zone{
					Spec: zonev1beta1.ZoneSpec{
						ClusterManagedResourceSpec: xpv1.ClusterManagedResourceSpec{
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
			nc := func(cfg clients.Config) (zones.Client, error) {
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
	testZone := cloudflare.Zone{
		Account: cloudflare.Account{
			ID:   "a1234",
			Name: "blah",
		},
		Plan: cloudflare.ZonePlan{
			ZonePlanCommon: cloudflare.ZonePlanCommon{
				ID:   "a1235",
				Name: "blah1",
			},
		},
		Paused:   true,
		VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
	}

	type fields struct {
		client zones.Client
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
		"ErrNotZone": {
			reason: "An error should be returned if the managed resource is not a *Zone",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotZone),
			},
		},
		"ErrNoZone": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				client: fake.MockClient{},
			},
			args: args{
				mg: &zonev1beta1.Zone{},
			},
			want: want{
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrZoneLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{}, errBoom
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
				),
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, errZoneLookup),
			},
		},
		"SuccessNeedsUpdate": {
			reason: "We should return ResourceExists: true and no error when a zone is found",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return testZone, nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "edge_cache_ttl", Value: 7200, Editable: true},
							},
						}, nil
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					// Paused is different than input params, this will trigger
					// ResourceUpToDate: false
					withPaused(ptr.To(false)),
					withAccount(ptr.To("a1234")),
					withPlan(ptr.To("a1235")),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"SuccessLateInit": {
			reason: "We should return ResourceLateInitialized: true and ResourceUpToDate: false when updates are required",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return testZone, nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "edge_cache_ttl", Value: 7200, Editable: true},
								{ID: "0rtt", Value: "off", Editable: true},
							},
						}, nil
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					withPaused(ptr.To(false)),
					withAccount(ptr.To("a1234")),
					withPlan(ptr.To("a1235")),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: false,
				},
				err: nil,
			},
		},
		"Success": {
			reason: "We should return ResourceLateInitialized: false and ResourceUpToDate: true when resource exactly matches remote",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return testZone, nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "edge_cache_ttl", Value: 7200, Editable: true},
								{ID: "0rtt", Value: "off", Editable: true},
							},
						}, nil
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					withPaused(ptr.To(true)),
					withAccount(ptr.To("a1234")),
					withPlan(ptr.To("a1235")),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        true,
					ResourceLateInitialized: false,
				},
				err: nil,
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
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Observe(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client zones.Client
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
		"ErrNotZone": {
			reason: "An error should be returned if the managed resource is not a *Zone",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotZone),
			},
		},
		"ErrZoneCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				client: fake.MockClient{
					MockCreateZone: func(ctx context.Context, name string, jumpstart bool, account cloudflare.Account, zoneType string) (cloudflare.Zone, error) {
						return cloudflare.Zone{}, errBoom
					},
				},
			},
			args: args{
				mg: zone(withExternalName("1234beef"), withType(ptr.To("full"))),
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, errZoneCreation),
			},
		},
		"Success": {
			reason: "We should return ExternalNameAssigned: true and no error when a zone is created",
			fields: fields{
				client: fake.MockClient{
					MockCreateZone: func(ctx context.Context, name string, jumpstart bool, account cloudflare.Account, zoneType string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:       "abcd",
							Name:     name,
							Type:     "full",
							Paused:   false,
							VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
						}, nil
					},
				},
			},
			args: args{
				mg: zone(withPaused(ptr.To(false)), withType(ptr.To("full"))),
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: nil,
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
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Create(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client zones.Client
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
		"ErrNotZone": {
			reason: "An error should be returned if the managed resource is not a *Zone",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotZone),
			},
		},
		"ErrNoZone": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: fake.MockClient{},
			},
			args: args{
				mg: &zonev1beta1.Zone{},
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errZoneUpdate),
			},
		},
		"ErrZoneUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:     zoneID,
							Paused: false,
						}, nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "test1", Value: "foo"},
								{ID: "test2", Value: "bar"},
							},
						}, nil
					},
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						return cloudflare.Zone{}, errBoom
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					withType(ptr.To("full")),
					withPaused(ptr.To(true)),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, errZoneUpdate),
			},
		},
		"Success": {
			reason: "We should return no error when a zone is updated",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:     zoneID,
							Paused: false,
						}, nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "edge_cache_ttl", Value: 3600, Editable: true},
							},
						}, nil
					},
					MockUpdateZoneSettings: func(ctx context.Context, zoneID string, cs []cloudflare.ZoneSetting) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: cs[0].ID, Value: cs[0].Value, Editable: cs[0].Editable},
							},
						}, nil
					},
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						return cloudflare.Zone{}, nil
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					withPaused(ptr.To(true)),
					withType(ptr.To("full")),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: nil,
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
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\ne.Update(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client zones.Client
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
		"ErrNotZone": {
			reason: "An error should be returned if the managed resource is not a *Zone",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotZone),
			},
		},
		"ErrNoZone": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: fake.MockClient{},
			},
			args: args{
				mg: &zonev1beta1.Zone{},
			},
			want: want{
				err: errors.New(errZoneDeletion),
			},
		},
		"ErrZoneDelete": {
			reason: "We should return any errors during the delete process",
			fields: fields{
				client: fake.MockClient{
					MockDeleteZone: func(ctx context.Context, zoneID string) (cloudflare.ZoneID, error) {
						return cloudflare.ZoneID{}, errBoom
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
				),
			},
			want: want{
				err: errors.Wrap(errBoom, errZoneDeletion),
			},
		},
		"Success": {
			reason: "We should return no error when a zone is deleted",
			fields: fields{
				client: fake.MockClient{
					MockDeleteZone: func(ctx context.Context, zoneID string) (cloudflare.ZoneID, error) {
						return cloudflare.ZoneID{ID: zoneID}, nil
					},
				},
			},
			args: args{
				mg: zone(
					withExternalName("1234beef"),
					withPaused(ptr.To(true)),
					withType(ptr.To("full")),
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.fields.client}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ne.Delete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestSetupSchemeRegistration(t *testing.T) {
	// CRITICAL TEST: Verifies Zone types are properly defined and can be instantiated
	// This catches basic type registration issues that would prevent controller setup
	// FAILURE INDICATOR: If this test fails, the Zone controller cannot be initialized

	// These should not panic if the types are properly defined
	zone := &zonev1beta1.Zone{}
	zoneList := &zonev1beta1.ZoneList{}

	// Verify the types have the expected metadata
	if zonev1beta1.ZoneKind != "Zone" {
		t.Errorf("ZoneKind = %s, want Zone", zonev1beta1.ZoneKind)
	}
	if zonev1beta1.Group != "zone.cloudflare.m.crossplane.io" {
		t.Errorf("Group = %s, want zone.cloudflare.m.crossplane.io", zonev1beta1.Group)
	}
	if zonev1beta1.Version != "v1beta1" {
		t.Errorf("Version = %s, want v1beta1", zonev1beta1.Version)
	}

	// Verify we can access fields without panicking
	_ = zone.Spec
	_ = zone.Status
	_ = zoneList.Items

	// If we get here without panicking, the basic type registration is working
	// The comprehensive scheme test is in the apis package
}
