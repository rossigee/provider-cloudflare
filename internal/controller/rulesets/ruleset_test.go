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

package rulesets

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/rulesets/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	ruleset "github.com/rossigee/provider-cloudflare/internal/clients/rulesets"

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

type mockRulesetClient struct {
	MockCreateRuleset func(ctx context.Context, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error)
	MockGetRuleset    func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error)
	MockUpdateRuleset func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error)
	MockDeleteRuleset func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) error
}

func (m *mockRulesetClient) CreateRuleset(ctx context.Context, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
	return m.MockCreateRuleset(ctx, params)
}

func (m *mockRulesetClient) GetRuleset(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
	return m.MockGetRuleset(ctx, rulesetID, params)
}

func (m *mockRulesetClient) UpdateRuleset(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
	return m.MockUpdateRuleset(ctx, rulesetID, params)
}

func (m *mockRulesetClient) DeleteRuleset(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) error {
	return m.MockDeleteRuleset(ctx, rulesetID, params)
}

type rulesetModifier func(*v1alpha1.Ruleset)

func withZone(zone string) rulesetModifier {
	return func(rs *v1alpha1.Ruleset) { rs.Spec.ForProvider.Zone = &zone }
}


func withRulesetID(id string) rulesetModifier {
	return func(rs *v1alpha1.Ruleset) { rs.Status.AtProvider.ID = id }
}

func rulesetCR(m ...rulesetModifier) *v1alpha1.Ruleset {
	rs := &v1alpha1.Ruleset{
		Spec: v1alpha1.RulesetSpec{
			ForProvider: v1alpha1.RulesetParameters{
				Name:        "test-ruleset",
				Description: stringPtr("Test ruleset"),
				Kind:        "zone",
				Phase:       "http_request_firewall_custom",
			},
		},
	}
	for _, f := range m {
		f(rs)
	}
	return rs
}

func stringPtr(s string) *string {
	return &s
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube                  client.Client
		newCloudflareClientFn func(cfg clients.Config) (ruleset.Client, error)
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
		"ErrNotRuleset": {
			reason: "Should return an error if the managed resource is not a Ruleset",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotRuleset),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: rulesetCR(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errClientConfig),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &rulesetConnector{
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
		client ruleset.Client
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
		"ErrNotRuleset": {
			reason: "Should return an error if the managed resource is not a Ruleset",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRuleset),
			},
		},
		"ErrNoScope": {
			reason: "Should return an error if neither zone nor account is specified",
			args: args{
				mg: &v1alpha1.Ruleset{
					Spec: v1alpha1.RulesetSpec{
						ForProvider: v1alpha1.RulesetParameters{
							Name: "test-ruleset",
						},
					},
				},
			},
			want: want{
				err: errors.New(errRulesetNoScope),
			},
		},
		"ErrGetRuleset": {
			reason: "Should return any error encountered getting the ruleset",
			fields: fields{
				client: &mockRulesetClient{
					MockGetRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					withRulesetID("test-ruleset-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errRulesetLookup),
			},
		},
		"RulesetNotFound": {
			reason: "Should report that the ruleset does not exist",
			fields: fields{
				client: &mockRulesetClient{
					MockGetRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return nil, &cloudflare.Error{StatusCode: 404}
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					withRulesetID("test-ruleset-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
						})
					},
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"RulesetExistsAndUpToDate": {
			reason: "Should report that the ruleset exists and is up to date",
			fields: fields{
				client: &mockRulesetClient{
					MockGetRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return &cloudflare.Ruleset{
							ID:          "test-ruleset-id",
							Name:        "test-ruleset",
							Description: "Test ruleset",
							Kind:        "zone",
							Phase:       "http_request_firewall_custom",
						}, nil
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
						})
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
		"NoExternalName": {
			reason: "Should report that the ruleset does not exist when no external name is set",
			args: args{
				mg: rulesetCR(withZone("test-zone-id")),
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
			e := &rulesetExternal{client: tc.fields.client}
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
		client ruleset.Client
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
		"ErrNotRuleset": {
			reason: "Should return an error if the managed resource is not a Ruleset",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRuleset),
			},
		},
		"ErrNoScope": {
			reason: "Should return an error if neither zone nor account is specified",
			args: args{
				mg: &v1alpha1.Ruleset{
					Spec: v1alpha1.RulesetSpec{
						ForProvider: v1alpha1.RulesetParameters{
							Name: "test-ruleset",
						},
					},
				},
			},
			want: want{
				err: errors.New(errRulesetNoScope),
			},
		},
		"ErrCreateRuleset": {
			reason: "Should return any error encountered creating the ruleset",
			fields: fields{
				client: &mockRulesetClient{
					MockCreateRuleset: func(ctx context.Context, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: rulesetCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errRulesetCreation),
			},
		},
		"Success": {
			reason: "Should return no error when ruleset is created successfully",
			fields: fields{
				client: &mockRulesetClient{
					MockCreateRuleset: func(ctx context.Context, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return &cloudflare.Ruleset{
							ID:          "test-ruleset-id",
							Name:        "test-ruleset",
							Description: "Test ruleset",
							Kind:        "zone",
							Phase:       "http_request_firewall_custom",
						}, nil
					},
				},
			},
			args: args{
				mg: rulesetCR(withZone("test-zone-id")),
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &rulesetExternal{client: tc.fields.client}
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
		client ruleset.Client
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
		"ErrNotRuleset": {
			reason: "Should return an error if the managed resource is not a Ruleset",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRuleset),
			},
		},
		"ErrNoScope": {
			reason: "Should return an error if neither zone nor account is specified",
			args: args{
				mg: &v1alpha1.Ruleset{
					Spec: v1alpha1.RulesetSpec{
						ForProvider: v1alpha1.RulesetParameters{
							Name: "test-ruleset",
						},
					},
				},
			},
			want: want{
				err: errors.New(errRulesetNoScope),
			},
		},
		"ErrNoExternalName": {
			reason: "Should return an error if no external name is set",
			args: args{
				mg: rulesetCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.New(errRulesetUpdate),
			},
		},
		"ErrUpdateRuleset": {
			reason: "Should return any error encountered updating the ruleset",
			fields: fields{
				client: &mockRulesetClient{
					MockUpdateRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errRulesetUpdate),
			},
		},
		"Success": {
			reason: "Should return no error when ruleset is updated successfully",
			fields: fields{
				client: &mockRulesetClient{
					MockUpdateRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) (*cloudflare.Ruleset, error) {
						return &cloudflare.Ruleset{
							ID:          "test-ruleset-id",
							Name:        "test-ruleset",
							Description: "Test ruleset",
							Kind:        "zone",
							Phase:       "http_request_firewall_custom",
						}, nil
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
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
			e := &rulesetExternal{client: tc.fields.client}
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
		client ruleset.Client
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
		"ErrNotRuleset": {
			reason: "Should return an error if the managed resource is not a Ruleset",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRuleset),
			},
		},
		"ErrNoScope": {
			reason: "Should return an error if neither zone nor account is specified",
			args: args{
				mg: &v1alpha1.Ruleset{
					Spec: v1alpha1.RulesetSpec{
						ForProvider: v1alpha1.RulesetParameters{
							Name: "test-ruleset",
						},
					},
				},
			},
			want: want{
				err: errors.New(errRulesetNoScope),
			},
		},
		"ErrNoExternalName": {
			reason: "Should return an error if no external name is set",
			args: args{
				mg: rulesetCR(withZone("test-zone-id")),
			},
			want: want{
				err: errors.New(errRulesetDeletion),
			},
		},
		"ErrDeleteRuleset": {
			reason: "Should return any error encountered deleting the ruleset",
			fields: fields{
				client: &mockRulesetClient{
					MockDeleteRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) error {
						return errors.New("boom")
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
						})
					},
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), errRulesetDeletion),
			},
		},
		"Success": {
			reason: "Should return no error when ruleset is deleted successfully",
			fields: fields{
				client: &mockRulesetClient{
					MockDeleteRuleset: func(ctx context.Context, rulesetID string, params v1alpha1.RulesetParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: rulesetCR(
					withZone("test-zone-id"),
					func(rs *v1alpha1.Ruleset) {
						rs.SetAnnotations(map[string]string{
							"crossplane.io/external-name": "test-ruleset-id",
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
			e := &rulesetExternal{client: tc.fields.client}
			err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}