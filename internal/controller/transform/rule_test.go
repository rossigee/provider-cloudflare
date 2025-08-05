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

package transform

import (
	"context"
	"net/http"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	pcv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	"github.com/rossigee/provider-cloudflare/apis/transform/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	transformrule "github.com/rossigee/provider-cloudflare/internal/clients/transform/rule"
	"github.com/rossigee/provider-cloudflare/internal/clients/transform/rule/fake"

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

type ruleModifier func(*v1alpha1.Rule)

func withExternalName(name string) ruleModifier {
	return func(r *v1alpha1.Rule) { meta.SetExternalName(r, name) }
}

func withZone(zoneID string) ruleModifier {
	return func(r *v1alpha1.Rule) {
		if zoneID == "" {
			r.Spec.ForProvider.Zone = nil
		} else {
			r.Spec.ForProvider.Zone = &zoneID
		}
	}
}

func withExpression(expression string) ruleModifier {
	return func(r *v1alpha1.Rule) { r.Spec.ForProvider.Expression = expression }
}

func withConditions(c ...xpv1.Condition) ruleModifier {
	return func(r *v1alpha1.Rule) { r.Status.Conditions = c }
}


func withStatus(s v1alpha1.RuleStatus) ruleModifier {
	return func(r *v1alpha1.Rule) { r.Status = s }
}

func rule(m ...ruleModifier) *v1alpha1.Rule {
	cr := &v1alpha1.Rule{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-rule",
		},
		Spec: v1alpha1.RuleSpec{
			ForProvider: v1alpha1.RuleParameters{
				Phase:      "http_request_transform",
				Expression: `http.request.uri.path eq "/test"`,
				Action:     "rewrite",
				Zone:       ptr.To("test-zone-id"),
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
		newClient func(cfg clients.Config, hc *http.Client) (transformrule.Client, error)
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
		"ErrNotRule": {
			reason: "An error should be returned if the managed resource is not a *Rule",
			args: args{
				mg: nil,
			},
			want: errors.New(errNotRule),
		},
		"ErrGetConfig": {
			reason: "Any errors from GetConfig should be wrapped",
			fields: fields{
				kube: mc,
			},
			args: args{
				mg: &v1alpha1.Rule{
					Spec: v1alpha1.RuleSpec{
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
				newClient: transformrule.NewClient,
			},
			args: args{
				mg: &v1alpha1.Rule{
					Spec: v1alpha1.RuleSpec{
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
			nc := func(cfg clients.Config) (transformrule.Client, error) {
				return tc.fields.newClient(cfg, nil)
			}
			e := &connector{kube: tc.fields.kube, newTransformRuleClientFn: nc}
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
		client transformrule.Client
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
		"ErrNotRule": {
			reason: "An error should be returned if the managed resource is not a *Rule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRule),
			},
		},
		"ErrNoRule": {
			reason: "We should return ResourceExists: false when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(),
			},
			want: want{
				cr: rule(),
				o:  managed.ExternalObservation{ResourceExists: false},
			},
		},
		"ErrRuleNoZone": {
			reason: "We should return an error if the Rule does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone(""),
				),
			},
			want: want{
				cr: rule(
					withExternalName("test-rule-id"),
					withZone(""),
				),
				o:   managed.ExternalObservation{},
				err: errors.New(errRuleNoZone),
			},
		},
		"ErrRuleLookup": {
			reason: "We should return an empty observation and an error if the API returned an error",
			fields: fields{
				client: &fake.MockClient{
					MockGetTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{}, errBoom
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
				o:   managed.ExternalObservation{},
				err: errors.Wrap(errBoom, errRuleLookup),
			},
		},
		"RuleNotFound": {
			reason: "We should return ResourceExists: false when rule is not found",
			fields: fields{
				client: &fake.MockClient{
					MockGetTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{}, errors.New("API Error 10014: Rule not found")
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
				o: managed.ExternalObservation{ResourceExists: false},
			},
		},
		"Success": {
			reason: "We should return ResourceExists: true and no error when a Rule is found",
			fields: fields{
				client: &fake.MockClient{
					MockGetTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{
							ID:         ruleID,
							Expression: `http.request.uri.path eq "/test"`,
							Action:     "rewrite",
						}, nil
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.RuleStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
						AtProvider: v1alpha1.RuleObservation{
							ID: "test-rule-id",
						},
					}),
				),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"SuccessOutdated": {
			reason: "We should return ResourceExists: true and ResourceUpToDate: false when rule differs",
			fields: fields{
				client: &fake.MockClient{
					MockGetTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{
							ID:         ruleID,
							Expression: `http.request.uri.path eq "/different"`,
							Action:     "rewrite",
						}, nil
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				cr: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
					withConditions(xpv1.Available()),
					withStatus(v1alpha1.RuleStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Available()},
							},
						},
						AtProvider: v1alpha1.RuleObservation{
							ID: "test-rule-id",
						},
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
		client transformrule.Client
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
		"ErrNotRule": {
			reason: "An error should be returned if the managed resource is not a *Rule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRule),
			},
		},
		"ErrRuleNoZone": {
			reason: "We should return an error if the Rule does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(withZone("")),
			},
			want: want{
				cr:  rule(withZone("")),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errors.New(errRuleNoZone), errRuleCreation),
			},
		},
		"ErrRuleCreate": {
			reason: "We should return any errors during the create process",
			fields: fields{
				client: &fake.MockClient{
					MockCreateTransformRule: func(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{}, errBoom
					},
				},
			},
			args: args{
				mg: rule(),
			},
			want: want{
				cr:  rule(withConditions(xpv1.Creating())),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errBoom, errRuleCreation),
			},
		},
		"Success": {
			reason: "We should return no error when a Rule is created",
			fields: fields{
				client: &fake.MockClient{
					MockCreateTransformRule: func(ctx context.Context, zoneID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{
							ID:         "new-rule-id",
							Expression: spec.Expression,
							Action:     spec.Action,
						}, nil
					},
				},
			},
			args: args{
				mg: rule(),
			},
			want: want{
				cr: rule(
					withExternalName("new-rule-id"),
					withConditions(xpv1.Creating()),
					withStatus(v1alpha1.RuleStatus{
						ResourceStatus: xpv1.ResourceStatus{
							ConditionedStatus: xpv1.ConditionedStatus{
								Conditions: []xpv1.Condition{xpv1.Creating()},
							},
						},
						AtProvider: v1alpha1.RuleObservation{
							ID: "new-rule-id",
						},
					}),
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
		client transformrule.Client
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
		"ErrNotRule": {
			reason: "An error should be returned if the managed resource is not a *Rule*",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRule),
			},
		},
		"ErrNoRule": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(
					withZone("foo.com"),
					withExpression(`http.request.uri.path eq "/test"`),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errRuleUpdate),
			},
		},
		"ErrRuleNoZone": {
			reason: "We should return an error if the Rule does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(
					withExternalName("1234beef"),
					withExpression(`http.request.uri.path eq "/test"`),
					withZone(""),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errors.New(errRuleNoZone), errRuleUpdate),
			},
		},
		"ErrRuleUpdate": {
			reason: "We should return any errors during the update process",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateTransformRule: func(ctx context.Context, zoneID, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{}, errBoom
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("1234beef"),
					withZone("foo.com"),
					withExpression(`http.request.uri.path eq "/test"`),
				),
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errBoom, errRuleUpdate),
			},
		},
		"Success": {
			reason: "We should return no error when a rule is updated",
			fields: fields{
				client: &fake.MockClient{
					MockUpdateTransformRule: func(ctx context.Context, zoneID, ruleID string, spec *v1alpha1.RuleParameters) (cloudflare.RulesetRule, error) {
						return cloudflare.RulesetRule{
							ID:         ruleID,
							Expression: spec.Expression,
							Action:     spec.Action,
						}, nil
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("1234beef"),
					withZone("foo.com"),
					withExpression(`http.request.uri.path eq "/test"`),
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
		client transformrule.Client
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
		"ErrNotRule": {
			reason: "An error should be returned if the managed resource is not a *Rule",
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotRule),
			},
		},
		"ErrNoExternalName": {
			reason: "We should return an error when no external name is set",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(),
			},
			want: want{
				err: errors.New(errRuleDeletion),
			},
		},
		"ErrRuleNoZone": {
			reason: "We should return an error if the Rule does not have a zone",
			fields: fields{
				client: &fake.MockClient{},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone(""),
				),
			},
			want: want{
				err: errors.Wrap(errors.New(errRuleNoZone), errRuleDeletion),
			},
		},
		"ErrRuleDeletion": {
			reason: "We should return any errors during the deletion process",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) error {
						return errBoom
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
				),
			},
			want: want{
				err: errors.Wrap(errBoom, errRuleDeletion),
			},
		},
		"Success": {
			reason: "We should return no error when a Rule is deleted",
			fields: fields{
				client: &fake.MockClient{
					MockDeleteTransformRule: func(ctx context.Context, zoneID, ruleID, phase string) error {
						return nil
					},
				},
			},
			args: args{
				mg: rule(
					withExternalName("test-rule-id"),
					withZone("test-zone-id"),
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