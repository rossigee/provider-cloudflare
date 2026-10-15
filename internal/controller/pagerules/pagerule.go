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

package pagerules

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	pagerulesv1beta1 "github.com/rossigee/provider-cloudflare/apis/pagerules/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/pagerules"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	errNotPageRule = "managed resource is not a PageRule custom resource"
	errGetCreds    = "cannot get credentials"
)

func SetupPageRule(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(pagerulesv1beta1.PageRuleKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(pagerulesv1beta1.PageRuleGroupVersionKind),
		managed.WithExternalConnector(&pageRuleConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) pagerules.CloudflarePageRuleClient {
				return pagerules.NewClientFromAPI(api)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&pagerulesv1beta1.PageRule{}).
		Complete(r)
}

type pageRuleConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) pagerules.CloudflarePageRuleClient
}

func (c *pageRuleConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*pagerulesv1beta1.PageRule)
	if !ok {
		return nil, errors.New(errNotPageRule)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	api, err := clients.NewClient(*cfg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Cloudflare client")
	}

	return &pageRuleExternal{
		service: c.newServiceFn(api),
		kube:    c.kube,
	}, nil
}

type pageRuleExternal struct {
	service pagerules.CloudflarePageRuleClient
	kube    client.Client
}

func (c *pageRuleExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*pagerulesv1beta1.PageRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPageRule)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	rule, err := c.service.GetPageRule(ctx, cr.Spec.ForProvider.ZoneID, meta.GetExternalName(cr))
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get Page Rule")
	}

	cr.Status.AtProvider.ID = rule.ID
	cr.Status.AtProvider.ZoneID = cr.Spec.ForProvider.ZoneID
	cr.Status.AtProvider.Priority = rule.Priority
	cr.Status.AtProvider.Status = rule.Status

	// Convert targets
	cr.Status.AtProvider.Targets = make([]pagerulesv1beta1.Target, len(rule.Targets))
	for i, t := range rule.Targets {
		cr.Status.AtProvider.Targets[i] = pagerulesv1beta1.Target{
			Target: t.Target,
			Constraint: pagerulesv1beta1.TargetConstraint{
				Operator: t.Constraint.Operator,
				Value:    t.Constraint.Value,
			},
		}
	}

	// Convert actions
	cr.Status.AtProvider.Actions = make([]pagerulesv1beta1.Action, len(rule.Actions))
	for i, a := range rule.Actions {
		val := ""
		if a.Value != nil {
			if s, ok := a.Value.(string); ok {
				val = s
			} else {
				// best effort for other types
				val = fmt.Sprintf("%v", a.Value)
			}
		}
		cr.Status.AtProvider.Actions[i] = pagerulesv1beta1.Action{
			ID:    a.ID,
			Value: val,
		}
	}

	// TODO: proper diff; for now assume up to date if exists
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pageRuleExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*pagerulesv1beta1.PageRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPageRule)
	}

	rule := cloudflare.PageRule{
		Targets:  make([]cloudflare.PageRuleTarget, len(cr.Spec.ForProvider.Targets)),
		Actions:  make([]cloudflare.PageRuleAction, len(cr.Spec.ForProvider.Actions)),
		Priority: cr.Spec.ForProvider.Priority,
		Status:   cr.Spec.ForProvider.Status,
	}

	for i, t := range cr.Spec.ForProvider.Targets {
		rule.Targets[i] = cloudflare.PageRuleTarget{
			Target: t.Target,
			Constraint: struct {
				Operator string `json:"operator"`
				Value    string `json:"value"`
			}{
				Operator: t.Constraint.Operator,
				Value:    t.Constraint.Value,
			},
		}
	}

	for i, a := range cr.Spec.ForProvider.Actions {
		rule.Actions[i] = cloudflare.PageRuleAction{
			ID:    a.ID,
			Value: a.Value,
		}
	}

	created, err := c.service.CreatePageRule(ctx, cr.Spec.ForProvider.ZoneID, rule)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create Page Rule")
	}

	meta.SetExternalName(cr, created.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pageRuleExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*pagerulesv1beta1.PageRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPageRule)
	}

	rule := cloudflare.PageRule{
		Targets:  make([]cloudflare.PageRuleTarget, len(cr.Spec.ForProvider.Targets)),
		Actions:  make([]cloudflare.PageRuleAction, len(cr.Spec.ForProvider.Actions)),
		Priority: cr.Spec.ForProvider.Priority,
		Status:   cr.Spec.ForProvider.Status,
	}

	for i, t := range cr.Spec.ForProvider.Targets {
		rule.Targets[i] = cloudflare.PageRuleTarget{
			Target: t.Target,
			Constraint: struct {
				Operator string `json:"operator"`
				Value    string `json:"value"`
			}{
				Operator: t.Constraint.Operator,
				Value:    t.Constraint.Value,
			},
		}
	}

	for i, a := range cr.Spec.ForProvider.Actions {
		rule.Actions[i] = cloudflare.PageRuleAction{
			ID:    a.ID,
			Value: a.Value,
		}
	}

	err := c.service.UpdatePageRule(ctx, cr.Spec.ForProvider.ZoneID, meta.GetExternalName(cr), rule)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update Page Rule")
	}

	return managed.ExternalUpdate{}, nil
}

func (c *pageRuleExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*pagerulesv1beta1.PageRule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPageRule)
	}

	err := c.service.DeletePageRule(ctx, cr.Spec.ForProvider.ZoneID, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete Page Rule")
	}

	return managed.ExternalDelete{}, nil
}

func (c *pageRuleExternal) Disconnect(ctx context.Context) error {
	return nil
}
