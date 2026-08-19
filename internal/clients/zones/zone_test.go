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

package zones

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/zones/fake"
	"k8s.io/utils/ptr"
)

func TestLateInitialize(t *testing.T) {
	type args struct {
		zp  *v1beta1.ZoneParameters
		z   cloudflare.Zone
		czs *v1beta1.ZoneSettings
	}

	type want struct {
		o  bool
		zp *v1beta1.ZoneParameters
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"LateInitSpecNil": {
			reason: "LateInit should return false when not passed a spec",
			args:   args{},
			want: want{
				o: false,
			},
		},
		"Success": {
			reason: "LateInit should update fields from a Zone",
			args: args{
				zp: &v1beta1.ZoneParameters{
					AccountID: ptr.To("beef"),
					PlanID:    ptr.To("dead")},
				z: cloudflare.Zone{
					Account: cloudflare.Account{
						ID: "beef",
					},
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "dead",
						},
					},
					Paused:   false,
					VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
				},
				czs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: true,
				zp: &v1beta1.ZoneParameters{
					Paused:    ptr.To(false),
					AccountID: ptr.To("beef"),
					PlanID:    ptr.To("dead")},
			},
		},
		"SuccessSettings": {
			reason: "LateInit should update settings from a Zone",
			args: args{
				zp: &v1beta1.ZoneParameters{
					AccountID: ptr.To("beef"),
					Paused:    ptr.To(false),
					PlanID:    ptr.To("dead")},
				z: cloudflare.Zone{
					Account: cloudflare.Account{
						ID: "beef",
					},
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "dead",
						},
					},
					// This field should not be late-inited, as the value
					// is already set false in zp
					Paused:   false,
					VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
				},
				// 'Current' Settings are those settings that were observed
				// from the API.
				// AdvancedDDOS, Minify and SecurityHeader should be late-inited here.
				czs: &v1beta1.ZoneSettings{
					AdvancedDDOS: ptr.To("yes"),
					Minify: &v1beta1.MinifySettings{
						CSS:  ptr.To("on"),
						HTML: ptr.To("on"),
						JS:   ptr.To("on"),
					},
					MobileRedirect: &v1beta1.MobileRedirectSettings{
						Status:    ptr.To("on"),
						Subdomain: ptr.To("m"),
						StripURI:  ptr.To(false),
					},
					SecurityHeader: &v1beta1.SecurityHeaderSettings{
						StrictTransportSecurity: &v1beta1.StrictTransportSecuritySettings{
							Enabled:           ptr.To(true),
							MaxAge:            ptr.To[int64](86400),
							IncludeSubdomains: ptr.To(true),
							NoSniff:           ptr.To(true),
						},
					},
					Ciphers: []string{
						"ECDHE-RSA-AES128-GCM-SHA256",
						"AES128-SHA",
					},
					BrowserCacheTTL: ptr.To[int64](3600),
				},
			},
			want: want{
				o: false,
				zp: &v1beta1.ZoneParameters{
					Paused:    ptr.To(false),
					AccountID: ptr.To("beef"),
					PlanID:    ptr.To("dead")},
			},
		},
		"SuccessSettingsPartial": {
			reason: "LateInit should update partially set settings from a Zone",
			args: args{
				zp: &v1beta1.ZoneParameters{
					AccountID: ptr.To("beef"),
					Paused:    ptr.To(false),
					PlanID:    ptr.To("dead")},
				z: cloudflare.Zone{
					Account: cloudflare.Account{
						ID: "beef",
					},
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "dead",
						},
					},
					// This field should not be late-inited, as the value
					// is already set false in zp
					Paused:   false,
					VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
				},
				// 'Current' Settings are those settings that were observed
				// from the API.
				// CSS and HTML under Minify should be late-inited here.
				czs: &v1beta1.ZoneSettings{
					Minify: &v1beta1.MinifySettings{
						CSS:  ptr.To("on"),
						HTML: ptr.To("off"),
						JS:   ptr.To("on"),
					},
					MobileRedirect: &v1beta1.MobileRedirectSettings{
						Status:    ptr.To("on"),
						Subdomain: ptr.To("m"),
						StripURI:  ptr.To(false),
					},
					SecurityHeader: &v1beta1.SecurityHeaderSettings{
						StrictTransportSecurity: &v1beta1.StrictTransportSecuritySettings{
							Enabled:           ptr.To(false),
							MaxAge:            ptr.To[int64](66700),
							NoSniff:           ptr.To(true),
							IncludeSubdomains: ptr.To(true),
						},
					},
				},
			},
			want: want{
				o: false,
				zp: &v1beta1.ZoneParameters{
					Paused:    ptr.To(false),
					AccountID: ptr.To("beef"),
					PlanID:    ptr.To("dead")},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := LateInitialize(tc.args.zp, tc.args.z, tc.args.czs)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nLateInit(...): -want, +got:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.zp, tc.args.zp); diff != "" {
				t.Errorf("\n%s\nLateInit(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}
func TestUpToDate(t *testing.T) {
	type args struct {
		zp  *v1beta1.ZoneParameters
		z   cloudflare.Zone
		ozs *v1beta1.ZoneSettings
	}

	type want struct {
		o bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SpecNil": {
			reason: "UpToDate should return true when not passed a spec",
			args:   args{},
			want: want{
				o: true,
			},
		},
		"EmptyParams": {
			reason: "UpToDate should return true and not panic with nil values",
			args: args{
				zp: &v1beta1.ZoneParameters{},
				z: cloudflare.Zone{
					Paused: true,
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "beef",
						},
					},
					PlanPending: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "cake",
						},
					},
					VanityNS: []string{},
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: true,
			},
		},
		"Paused": {
			reason: "UpToDate should return false if Paused is not up to date",
			args: args{
				zp: &v1beta1.ZoneParameters{
					Paused: ptr.To(false),
				},
				z: cloudflare.Zone{
					Paused: true,
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: false,
			},
		},
		"PlanFalse": {
			reason: "UpToDate should return false if PlanID is not one of Plan or PlanPending IDs",
			args: args{
				zp: &v1beta1.ZoneParameters{
					PlanID: ptr.To("moo"),
				},
				z: cloudflare.Zone{
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "beef",
						},
					},
					PlanPending: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "cake",
						},
					},
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: false,
			},
		},
		"PlanTrue": {
			reason: "UpToDate should return true if PlanID is current Plan ID",
			args: args{
				zp: &v1beta1.ZoneParameters{
					PlanID: ptr.To("beef"),
				},
				z: cloudflare.Zone{
					Plan: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "beef",
						},
					},
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: true,
			},
		},
		"PlanPendingTrue": {
			reason: "UpToDate should return true if PlanID is pending Plan ID",
			args: args{
				zp: &v1beta1.ZoneParameters{
					PlanID: ptr.To("cake")},
				z: cloudflare.Zone{
					PlanPending: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "cake",
						},
					},
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: true,
			},
		},
		"SettingsFalse": {
			reason: "UpToDate should return true as settings are not checked",
			args: args{
				zp: &v1beta1.ZoneParameters{
					PlanID: ptr.To("cake")},
				z: cloudflare.Zone{
					PlanPending: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "cake",
						},
					},
				},
				ozs: &v1beta1.ZoneSettings{
					ZeroRTT: ptr.To("yes"),
				},
			},
			want: want{
				o: true,
			},
		},
		"VanityNSTrue": {
			reason: "UpToDate should return true if VanityNS field matches in any order",
			args: args{
				zp: &v1beta1.ZoneParameters{
					PlanID: ptr.To("cake")},
				z: cloudflare.Zone{
					PlanPending: cloudflare.ZonePlan{
						ZonePlanCommon: cloudflare.ZonePlanCommon{
							ID: "cake",
						},
					},
					VanityNS: []string{"ns1.lele.com", "ns2.woowoo.org"},
				},
				ozs: &v1beta1.ZoneSettings{},
			},
			want: want{
				o: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := UpToDate(tc.args.zp, tc.args.z, tc.args.ozs)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdateZone(t *testing.T) {
	errBoom := errors.New("boom")

	inputZoneID := "1234"
	nsKey := cfsMinify

	nsInputValue := v1beta1.MinifySettings{
		CSS:  ptr.To("on"),
		HTML: ptr.To("off"),
		JS:   ptr.To("bar"),
	}

	type fields struct {
		client Client
	}

	type args struct {
		ctx context.Context
		id  string
		zp  v1beta1.ZoneParameters
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
		"UpdateZoneNotFound": {
			reason: "UpdateZone should return errUpdateZone if the zone is not found",
			fields: fields{
				client: fake.MockClient{
					// When the ZoneDetails method is called on this fake client,
					// return an error that we can compare.
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{}, errBoom
					},
				},
			},
			args: args{
				id: inputZoneID,
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateZone),
			},
		},
		"UpdateZoneOptions": {
			reason: "UpdateZone should return no error when updating zone options",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:       zoneID,
							Name:     "testzone.com",
							Paused:   true,
							VanityNS: []string{"ns1.lele.com"},
						}, nil
					},
					// When EditZone is called, check it receives the expected arguments.
					// If it doesn't we return an error which will cause the test to fail.
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						var err error
						if zoneID != inputZoneID {
							err = errors.New("zoneID value incorrect")
						}
						if *zoneOpts.Paused != false {
							err = errors.New("zoneOpts.Paused value incorrect")
						}

						// VanityNS should not be updated in this test
						if zoneOpts.VanityNS != nil {
							err = errors.New("zoneOpts.VanityNS should not be set")
						}
						// Returned zone is discarded by UpdateZone
						return cloudflare.Zone{}, err
					},

					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{}, nil
					},
				},
			},
			args: args{
				id: inputZoneID,
				zp: v1beta1.ZoneParameters{
					Paused: ptr.To(false)},
			},
			want: want{
				err: nil,
			},
		},
		"UpdateZoneSettings": {
			reason: "UpdateZone should return no error when updating zone settings",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:       zoneID,
							Name:     "testzone.com",
							Paused:   true,
							VanityNS: []string{"ns1.lele.com"},
						}, nil
					},
					// When EditZone is called, check it receives the expected arguments.
					// If it doesn't we return an error which will cause the test to fail.
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						var err error
						if zoneID != inputZoneID {
							err = errors.New("zoneID value incorrect")
						}
						if *zoneOpts.Paused != false {
							err = errors.New("zoneOpts.Paused value incorrect")
						}

						// VanityNS should not be updated in this test
						if zoneOpts.VanityNS != nil {
							err = errors.New("zoneOpts.VanityNS should not be set")
						}
						// Returned zone is discarded by UpdateZone
						return cloudflare.Zone{}, err
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{
									ID:       nsKey,
									Editable: true,
									// Client should decode nested values from map string interface
									Value: map[string]interface{}{
										cfsMinifyCSS:  nsInputValue.CSS,
										cfsMinifyHTML: nsInputValue.HTML,
										cfsMinifyJS:   "foo", // This value should be overwritten
									},
								},
							},
						}, nil
					},
					MockUpdateZoneSettings: func(ctx context.Context, zoneID string, cs []cloudflare.ZoneSetting) (*cloudflare.ZoneSettingResponse, error) {
						if zoneID != inputZoneID {
							return nil, errors.New("zoneID value incorrect")
						}
						nsInputExpectedValue := map[string]interface{}{
							cfsMinifyCSS:  *nsInputValue.CSS,
							cfsMinifyHTML: *nsInputValue.HTML,
							cfsMinifyJS:   *nsInputValue.JS,
						}
						// Must match our requested setting ID and value.
						// If not, we return an error.
						for _, setting := range cs {
							if setting.ID == nsKey {
								if cmp.Equal(nsInputExpectedValue, setting.Value) {
									return nil, nil
								}
							}
						}
						return nil, errors.New("Nested complex setting not updated or invalid")
					},
				},
			},
			args: args{
				id: inputZoneID,
				zp: v1beta1.ZoneParameters{
					Paused: ptr.To(false)},
			},
			want: want{
				err: nil,
			},
		},
		"UpdateZonePlan": {
			reason: "UpdateZone should call ZoneSetPlan when plan ID needs to be updated",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:     zoneID,
							Name:   "testzone.com",
							Paused: false,
							Plan: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "current-plan-id",
								},
							},
							PlanPending: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "other-plan-id",
								},
							},
							VanityNS: []string{},
						}, nil
					},
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						// No zone options should need updating in this test
						return cloudflare.Zone{}, nil
					},
					MockZoneSetPlan: func(ctx context.Context, zoneID string, planType string) error {
						if zoneID != inputZoneID {
							return errors.New("zoneID value incorrect")
						}
						if planType != "new-plan-id" {
							return errors.New("planType value incorrect")
						}
						return nil
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{}, nil
					},
				},
			},
			args: args{
				id: inputZoneID,
				zp: v1beta1.ZoneParameters{
					PlanID: ptr.To("new-plan-id"),
				},
			},
			want: want{
				err: nil,
			},
		},
		"UpdateZonePlanError": {
			reason: "UpdateZone should return wrapped error when ZoneSetPlan fails",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:     zoneID,
							Name:   "testzone.com",
							Paused: false,
							Plan: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "current-plan-id",
								},
							},
							PlanPending: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "other-plan-id",
								},
							},
							VanityNS: []string{},
						}, nil
					},
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						// No zone options should need updating in this test
						return cloudflare.Zone{}, nil
					},
					MockZoneSetPlan: func(ctx context.Context, zoneID string, planType string) error {
						return errBoom
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{}, nil
					},
				},
			},
			args: args{
				id: inputZoneID,
				zp: v1beta1.ZoneParameters{
					PlanID: ptr.To("new-plan-id"),
				},
			},
			want: want{
				err: errors.Wrap(errBoom, errSetPlan),
			},
		},
		"UpdateZonePlanNoChange": {
			reason: "UpdateZone should not call ZoneSetPlan when plan ID matches current or pending plan",
			fields: fields{
				client: fake.MockClient{
					MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
						return cloudflare.Zone{
							ID:     zoneID,
							Name:   "testzone.com",
							Paused: false,
							Plan: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "current-plan-id",
								},
							},
							PlanPending: cloudflare.ZonePlan{
								ZonePlanCommon: cloudflare.ZonePlanCommon{
									ID: "pending-plan-id",
								},
							},
							VanityNS: []string{},
						}, nil
					},
					MockEditZone: func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
						// No zone options should need updating in this test
						return cloudflare.Zone{}, nil
					},
					MockZoneSetPlan: func(ctx context.Context, zoneID string, planType string) error {
						// This should not be called
						return errors.New("ZoneSetPlan should not be called when plan matches")
					},
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{}, nil
					},
				},
			},
			args: args{
				id: inputZoneID,
				zp: v1beta1.ZoneParameters{
					PlanID: ptr.To("current-plan-id"), // Matches current plan
				},
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := UpdateZone(tc.args.ctx, tc.fields.client, tc.args.id, tc.args.zp)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nUpdateZone(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestLoadSettingsForZone(t *testing.T) {
	errBoom := errors.New("boom")
	type fields struct {
		client Client
	}

	type args struct {
		ctx context.Context
		id  string
		zs  v1beta1.ZoneSettings
	}

	type want struct {
		err error
		o   v1beta1.ZoneSettings
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ErrorLookupSettings": {
			reason: "LoadSettingsForZone should return an error when the API call returns an error",
			fields: fields{
				client: fake.MockClient{
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				id: "abcd",
				zs: v1beta1.ZoneSettings{ZeroRTT: ptr.To("yes")},
			},
			want: want{
				err: errors.Wrap(errBoom, errLoadSettings),
				o:   v1beta1.ZoneSettings{ZeroRTT: ptr.To("yes")},
			},
		},
		"LoadUnknownSetting": {
			// Note: This is an academic test - all of the keys we use are static strings
			// So we _cannot_ load a setting into a struct without knowing about it.
			// We add this test to avoid regressions if the method used to load settings
			// is changed, as it has caused problems in the past.
			reason: "LoadSettingsForZone should not error when reading unknown settings",
			fields: fields{
				client: fake.MockClient{
					MockZoneSettings: func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
						return &cloudflare.ZoneSettingResponse{
							Result: []cloudflare.ZoneSetting{
								{ID: "unknownKey", Value: "foo"},
							},
						}, nil
					},
				},
			},
			args: args{
				id: "abcd",
				zs: v1beta1.ZoneSettings{
					AdvancedDDOS: ptr.To("yes"),
				},
			},
			want: want{
				err: nil,
				o: v1beta1.ZoneSettings{
					AdvancedDDOS: ptr.To("yes"),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.args.zs.DeepCopy()

			err := LoadSettingsForZone(tc.args.ctx, tc.fields.client, tc.args.id, &tc.args.zs)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nLoadSettingsForZone(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.o, *got); diff != "" {
				t.Errorf("\n%s\nLoadSettingsForZone(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestSecurityHeaderSettingsToMap(t *testing.T) {
	type args struct {
		settings *v1beta1.SecurityHeaderSettings
	}

	type want struct {
		o map[string]interface{}
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"Success": {
			reason: "securityHeaderSettingsToMap should return a valid map type",
			args: args{
				settings: &v1beta1.SecurityHeaderSettings{
					StrictTransportSecurity: &v1beta1.StrictTransportSecuritySettings{
						Enabled:           ptr.To(true),
						MaxAge:            ptr.To[int64](86400),
						IncludeSubdomains: ptr.To(true),
						NoSniff:           ptr.To(true),
					},
				},
			},
			want: want{
				o: map[string]interface{}{
					cfsStrictTransportSecurity: map[string]interface{}{
						cfsStrictTransportSecurityEnabled:           true,
						cfsStrictTransportSecurityIncludeSubdomains: true,
						cfsStrictTransportSecurityMaxAge:            int64(86400),
						cfsStrictTransportSecurityNoSniff:           true,
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := securityHeaderSettingsToMap(tc.args.settings)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nsecurityHeaderSettingsToMap(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestMobileRedirectSettingsToMap(t *testing.T) {
	type args struct {
		settings *v1beta1.MobileRedirectSettings
	}

	type want struct {
		o map[string]interface{}
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"Success": {
			reason: "mobileRedirectSettingsToMap should return a valid map type",
			args: args{
				settings: &v1beta1.MobileRedirectSettings{
					Status:    ptr.To("on"),
					Subdomain: ptr.To("m"),
					StripURI:  ptr.To(false),
				},
			},
			want: want{
				o: map[string]interface{}{
					cfsMobileRedirectStatus:    "on",
					cfsMobileRedirectSubdomain: "m",
					cfsMobileRedirectStripURI:  false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := mobileRedirectSettingsToMap(tc.args.settings)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nmobileRedirectSettingsToMap(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestMinifySettingsToMap(t *testing.T) {
	type args struct {
		settings *v1beta1.MinifySettings
	}

	type want struct {
		o map[string]interface{}
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"Success": {
			reason: "minifySettingsToMap should return a valid map type",
			args: args{
				settings: &v1beta1.MinifySettings{
					CSS:  ptr.To("on"),
					HTML: ptr.To("on"),
					JS:   ptr.To("on"),
				},
			},
			want: want{
				o: map[string]interface{}{
					cfsMinifyCSS:  "on",
					cfsMinifyHTML: "on",
					cfsMinifyJS:   "on",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := minifySettingsToMap(tc.args.settings)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nminifySettingsToMap(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}
