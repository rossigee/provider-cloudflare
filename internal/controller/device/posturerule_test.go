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

package device

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/device/posturerule"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

// Unlike many Kubernetes projects Crossplane does not use third party testing
// libraries, per the common Go test review comments. Crossplane encourages the
// use of table driven unit tests. The tests of the crossplane-runtime project
// are representative of the testing style Crossplane encourages.
//
// https://github.com/golang/go/wiki/TestComments
// https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md#contributing-code

// mockDevicePostureRuleAPI mocks the posturerule.DevicePostureRuleAPI interface
type mockDevicePostureRuleAPI struct {
	MockDevicePostureRule       func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error)
	MockCreateDevicePostureRule func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	MockUpdateDevicePostureRule func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	MockDeleteDevicePostureRule func(ctx context.Context, accountID, ruleID string) error
}

func (m *mockDevicePostureRuleAPI) DevicePostureRules(ctx context.Context, accountID string) ([]cloudflare.DevicePostureRule, cloudflare.ResultInfo, error) {
	return nil, cloudflare.ResultInfo{}, errors.New("not implemented")
}

func (m *mockDevicePostureRuleAPI) DevicePostureRule(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
	if m.MockDevicePostureRule != nil {
		return m.MockDevicePostureRule(ctx, accountID, ruleID)
	}
	return cloudflare.DevicePostureRule{}, errors.New("cannot get device posture rule")
}

func (m *mockDevicePostureRuleAPI) CreateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
	if m.MockCreateDevicePostureRule != nil {
		return m.MockCreateDevicePostureRule(ctx, accountID, rule)
	}
	return cloudflare.DevicePostureRule{}, errors.New("cannot create device posture rule")
}

func (m *mockDevicePostureRuleAPI) UpdateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
	if m.MockUpdateDevicePostureRule != nil {
		return m.MockUpdateDevicePostureRule(ctx, accountID, rule)
	}
	return cloudflare.DevicePostureRule{}, errors.New("cannot update device posture rule")
}

func (m *mockDevicePostureRuleAPI) DeleteDevicePostureRule(ctx context.Context, accountID, ruleID string) error {
	if m.MockDeleteDevicePostureRule != nil {
		return m.MockDeleteDevicePostureRule(ctx, accountID, ruleID)
	}
	return errors.New("cannot delete device posture rule")
}

// Helper to create a CloudflareDevicePostureRuleClient with a mocked API
func newMockDevicePostureRuleClient(api posturerule.DevicePostureRuleAPI) *posturerule.CloudflareDevicePostureRuleClient {
	return posturerule.NewClient(api)
}

type devicePostureRuleModifier func(*v1beta1.DevicePostureRule)

func withDevicePostureRuleID(id string) devicePostureRuleModifier {
	return func(dpr *v1beta1.DevicePostureRule) {
		dpr.Status.AtProvider.ID = id
		rtmeta.SetExternalName(dpr, id)
	}
}

func devicePostureRule(m ...devicePostureRuleModifier) *v1beta1.DevicePostureRule {
	dpr := &v1beta1.DevicePostureRule{
		Spec: v1beta1.DevicePostureRuleSpec{
			ForProvider: v1beta1.DevicePostureRuleParameters{
				AccountID: "test-account-id",
				Name:      "test-rule",
				Type:      "os_version",
			},
		},
	}
	for _, f := range m {
		f(dpr)
	}
	return dpr
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube         client.Client
		newServiceFn func(api *cloudflare.API) *posturerule.CloudflareDevicePostureRuleClient
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
		"ErrNotDevicePostureRule": {
			reason: "Should return an error if the managed resource is not a DevicePostureRule",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotDevicePostureRule),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: devicePostureRule(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errGetCreds),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &devicePostureRuleConnector{
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
		service *posturerule.CloudflareDevicePostureRuleClient
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
		"ErrNotDevicePostureRule": {
			reason: "Should return an error if the managed resource is not a DevicePostureRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotDevicePostureRule),
			},
		},
		"ErrGetDevicePostureRule": {
			reason: "Should return any error encountered getting the device posture rule",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot get device posture rule"), "cannot get external resource"),
			},
		},
		"DevicePostureRuleNotFound": {
			reason: "Should report that the device posture rule does not exist",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("device posture rule not found")
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"DevicePostureRuleExistsAndUpToDate": {
			reason: "Should report that the device posture rule exists and is up to date",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   "test-rule-id",
							Name: "test-rule",
							Type: "os_version",
						}, nil
					},
				}),
			},
			args: args{
				mg: devicePostureRule(
					withDevicePostureRuleID("test-rule-id"),
					func(dpr *v1beta1.DevicePostureRule) {
						dpr.Spec.ForProvider.Type = "os_version"
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
		"DevicePostureRuleExistsButOutOfDate": {
			reason: "Should report that the device posture rule exists but is not up to date",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDevicePostureRule: func(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   "test-rule-id",
							Name: "test-rule",
							Type: "os_version",
						}, nil
					},
				}),
			},
			args: args{
				mg: devicePostureRule(
					withDevicePostureRuleID("test-rule-id"),
					func(dpr *v1beta1.DevicePostureRule) {
						dpr.Spec.ForProvider.Name = "updated-rule"
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
			e := &devicePostureRuleExternal{service: tc.fields.service}
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
		service *posturerule.CloudflareDevicePostureRuleClient
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
		"ErrNotDevicePostureRule": {
			reason: "Should return an error if the managed resource is not a DevicePostureRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotDevicePostureRule),
			},
		},
		"ErrCreateDevicePostureRule": {
			reason: "Should return any error encountered creating the device posture rule",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockCreateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: devicePostureRule(),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot create device posture rule"), "cannot create external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when device posture rule is created successfully",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockCreateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   "test-rule-id",
							Name: "test-rule",
							Type: "os_version",
						}, nil
					},
				}),
			},
			args: args{
				mg: devicePostureRule(),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &devicePostureRuleExternal{service: tc.fields.service}
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
		service *posturerule.CloudflareDevicePostureRuleClient
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
		"ErrNotDevicePostureRule": {
			reason: "Should return an error if the managed resource is not a DevicePostureRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotDevicePostureRule),
			},
		},
		"ErrUpdateDevicePostureRule": {
			reason: "Should return any error encountered updating the device posture rule",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockUpdateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{}, errors.New("api error")
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errors.New("api error"), "cannot update device posture rule"), "cannot update external resource"),
			},
		},
		"Success": {
			reason: "Should return no error when device posture rule is updated successfully",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockUpdateDevicePostureRule: func(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error) {
						return cloudflare.DevicePostureRule{
							ID:   "test-rule-id",
							Name: "test-rule",
							Type: "os_version",
						}, nil
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &devicePostureRuleExternal{service: tc.fields.service}
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
		service *posturerule.CloudflareDevicePostureRuleClient
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
		"ErrNotDevicePostureRule": {
			reason: "Should return an error if the managed resource is not a DevicePostureRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotDevicePostureRule),
			},
		},
		"ErrDeleteDevicePostureRule": {
			reason: "Should return any error encountered deleting the device posture rule",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDeleteDevicePostureRule: func(ctx context.Context, accountID, ruleID string) error {
						return errors.New("api error")
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("api error"), "cannot delete device posture rule"),
			},
		},
		"Success": {
			reason: "Should return no error when device posture rule is deleted successfully",
			fields: fields{
				service: newMockDevicePostureRuleClient(&mockDevicePostureRuleAPI{
					MockDeleteDevicePostureRule: func(ctx context.Context, accountID, ruleID string) error {
						return nil
					},
				}),
			},
			args: args{
				mg: devicePostureRule(withDevicePostureRuleID("test-rule-id")),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &devicePostureRuleExternal{service: tc.fields.service}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}
