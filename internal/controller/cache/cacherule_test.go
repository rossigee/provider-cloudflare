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

package cache

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/cache/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/cache"

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

type mockCacheRuleClient struct {
	MockCreateCacheRule func(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	MockGetCacheRule    func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	MockUpdateCacheRule func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error)
	MockDeleteCacheRule func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error
}

func (m *mockCacheRuleClient) CreateCacheRule(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	return m.MockCreateCacheRule(ctx, params)
}

func (m *mockCacheRuleClient) GetCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	return m.MockGetCacheRule(ctx, rulesetID, ruleID, params)
}

func (m *mockCacheRuleClient) UpdateCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
	return m.MockUpdateCacheRule(ctx, rulesetID, ruleID, params)
}

func (m *mockCacheRuleClient) DeleteCacheRule(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error {
	return m.MockDeleteCacheRule(ctx, rulesetID, ruleID, params)
}

type cacheRuleModifier func(*v1alpha1.CacheRule)


func withRuleID(id string) cacheRuleModifier {
	return func(cr *v1alpha1.CacheRule) { cr.Status.AtProvider.ID = id }
}

func withRulesetID(id string) cacheRuleModifier {
	return func(cr *v1alpha1.CacheRule) { cr.Status.AtProvider.RulesetID = id }
}

func cacheRule(m ...cacheRuleModifier) *v1alpha1.CacheRule {
	cr := &v1alpha1.CacheRule{
		Spec: v1alpha1.CacheRuleSpec{
			ForProvider: v1alpha1.CacheRuleParameters{
				Zone:       "test-zone-id",
				Name:       "test-cache-rule",
				Expression: "(http.request.uri.path contains \"/images/\")",
			},
		},
	}
	for _, f := range m {
		f(cr)
	}
	return cr
}


func boolPtr(b bool) *bool {
	return &b
}

func TestConnect(t *testing.T) {
	mc := &test.MockClient{
		MockGet: test.NewMockGetFn(nil),
	}

	type fields struct {
		kube        client.Client
		newClientFn func(cfg clients.Config) (cache.CacheRuleClient, error)
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
		"ErrNotCacheRule": {
			reason: "Should return an error if the managed resource is not a CacheRule",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: nil,
			},
			want: errors.New(errNotCacheRule),
		},
		"ErrGetCredentials": {
			reason: "Should return any error encountered getting credentials",
			fields: fields{
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errors.New("boom")),
				},
			},
			args: args{
				mg: cacheRule(),
			},
			want: errors.Wrap(errors.New("providerConfigRef not set"), errGetCreds),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := &connector{
				kube:        tc.fields.kube,
				newClientFn: tc.fields.newClientFn,
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
		service cache.CacheRuleClient
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
		"ErrNotCacheRule": {
			reason: "Should return an error if the managed resource is not a CacheRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCacheRule),
			},
		},
		"ErrGetCacheRule": {
			reason: "Should return any error encountered getting the cache rule",
			fields: fields{
				service: &mockCacheRuleClient{
					MockGetCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return nil, nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), "failed to get cache rule from Cloudflare API"),
			},
		},
		"CacheRuleNotFound": {
			reason: "Should report that the cache rule does not exist",
			fields: fields{
				service: &mockCacheRuleClient{
					MockGetCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return nil, nil, &cloudflare.Error{StatusCode: 404}
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"CacheRuleExistsAndUpToDate": {
			reason: "Should report that the cache rule exists and is up to date",
			fields: fields{
				service: &mockCacheRuleClient{
					MockGetCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return &cloudflare.RulesetRule{
							ID:         "test-rule-id",
							Expression: "(http.request.uri.path contains \"/images/\")",
							Enabled:    boolPtr(true),
						}, &cloudflare.Ruleset{
							ID: "test-ruleset-id",
						}, nil
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
					func(cr *v1alpha1.CacheRule) {
						cr.Spec.ForProvider.Enabled = boolPtr(true)
					},
				),
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
		"CacheRuleExistsButOutOfDate": {
			reason: "Should report that the cache rule exists but is not up to date",
			fields: fields{
				service: &mockCacheRuleClient{
					MockGetCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return &cloudflare.RulesetRule{
							ID:         "test-rule-id",
							Expression: "(http.request.uri.path contains \"/css/\")",
							Enabled:    boolPtr(true),
						}, &cloudflare.Ruleset{
							ID: "test-ruleset-id",
						}, nil
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceUpToDate:        false,
					ResourceLateInitialized: true,
					ConnectionDetails:       managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{service: tc.fields.service}
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
		service cache.CacheRuleClient
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
		"ErrNotCacheRule": {
			reason: "Should return an error if the managed resource is not a CacheRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCacheRule),
			},
		},
		"ErrCreateCacheRule": {
			reason: "Should return any error encountered creating the cache rule",
			fields: fields{
				service: &mockCacheRuleClient{
					MockCreateCacheRule: func(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return nil, nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: cacheRule(),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), "failed to create cache rule in Cloudflare API"),
			},
		},
		"Success": {
			reason: "Should return no error when cache rule is created successfully",
			fields: fields{
				service: &mockCacheRuleClient{
					MockCreateCacheRule: func(ctx context.Context, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return &cloudflare.RulesetRule{
							ID: "test-rule-id",
						}, &cloudflare.Ruleset{
							ID: "test-ruleset-id",
						}, nil
					},
				},
			},
			args: args{
				mg: cacheRule(),
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
			e := &external{service: tc.fields.service}
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
		service cache.CacheRuleClient
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
		"ErrNotCacheRule": {
			reason: "Should return an error if the managed resource is not a CacheRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCacheRule),
			},
		},
		"ErrUpdateCacheRule": {
			reason: "Should return any error encountered updating the cache rule",
			fields: fields{
				service: &mockCacheRuleClient{
					MockUpdateCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return nil, nil, errors.New("boom")
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), "failed to update cache rule in Cloudflare API"),
			},
		},
		"Success": {
			reason: "Should return no error when cache rule is updated successfully",
			fields: fields{
				service: &mockCacheRuleClient{
					MockUpdateCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) (*cloudflare.RulesetRule, *cloudflare.Ruleset, error) {
						return &cloudflare.RulesetRule{
							ID: "test-rule-id",
						}, &cloudflare.Ruleset{
							ID: "test-ruleset-id",
						}, nil
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
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
			e := &external{service: tc.fields.service}
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
		service cache.CacheRuleClient
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
		"ErrNotCacheRule": {
			reason: "Should return an error if the managed resource is not a CacheRule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotCacheRule),
			},
		},
		"ErrDeleteCacheRule": {
			reason: "Should return any error encountered deleting the cache rule",
			fields: fields{
				service: &mockCacheRuleClient{
					MockDeleteCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error {
						return errors.New("boom")
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				err: errors.Wrap(errors.New("boom"), "failed to delete cache rule from Cloudflare API"),
			},
		},
		"Success": {
			reason: "Should return no error when cache rule is deleted successfully",
			fields: fields{
				service: &mockCacheRuleClient{
					MockDeleteCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error {
						return nil
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				err: nil,
			},
		},
		"AlreadyDeleted": {
			reason: "Should return no error when cache rule is already deleted",
			fields: fields{
				service: &mockCacheRuleClient{
					MockDeleteCacheRule: func(ctx context.Context, rulesetID, ruleID string, params v1alpha1.CacheRuleParameters) error {
						return &cloudflare.Error{StatusCode: 404}
					},
				},
			},
			args: args{
				mg: cacheRule(
					withRuleID("test-rule-id"),
					withRulesetID("test-ruleset-id"),
				),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{service: tc.fields.service}
			_, err := e.Delete(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s\ne.Delete(...): -want error, +got error:\n%s", tc.reason, diff)
			}
		})
	}
}