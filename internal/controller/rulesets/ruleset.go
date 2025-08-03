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
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	rtv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/rulesets/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	ruleset "github.com/rossigee/provider-cloudflare/internal/clients/rulesets"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotRuleset = "managed resource is not a Ruleset custom resource"

	errClientConfig = "error getting client config"

	errRulesetLookup   = "cannot lookup ruleset"
	errRulesetCreation = "cannot create ruleset"
	errRulesetUpdate   = "cannot update ruleset"
	errRulesetDeletion = "cannot delete ruleset"
	errRulesetNoScope  = "cannot create ruleset: no zone or account specified"
)

const (
	maxConcurrency = 5
)

// Setup adds a controller that reconciles Ruleset managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	return SetupRuleset(mgr, l, rl)
}

// SetupRuleset adds a controller that reconciles Ruleset managed resources.
func SetupRuleset(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.RulesetGroupKind)

	o := controller.Options{
		RateLimiter:             rl,
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RulesetGroupVersionKind),
		managed.WithExternalConnecter(&rulesetConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (ruleset.Client, error) {
				return ruleset.NewClient(cfg, hc)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(5*time.Minute),
		// Initialize external-name field.
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Ruleset{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type rulesetConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (ruleset.Client, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *rulesetConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Ruleset)
	if !ok {
		return nil, errors.New(errNotRuleset)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	client, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, err
	}

	return &rulesetExternal{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type rulesetExternal struct {
	client ruleset.Client
}

func (e *rulesetExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Ruleset)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRuleset)
	}

	// Validate that either zone or account is specified
	if cr.Spec.ForProvider.Zone == nil && cr.Spec.ForProvider.Account == nil {
		return managed.ExternalObservation{}, errors.New(errRulesetNoScope)
	}

	// Ruleset does not exist if we dont have an ID stored in external-name
	rulesetID := meta.GetExternalName(cr)
	if rulesetID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	rs, err := e.client.GetRuleset(ctx, rulesetID, cr.Spec.ForProvider)

	if err != nil {
		if ruleset.IsRulesetNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errRulesetLookup)
	}

	cr.Status.AtProvider = ruleset.GenerateObservation(rs)

	// Mark as ready
	cr.Status.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: ruleset.UpToDate(&cr.Spec.ForProvider, rs),
	}, nil
}

func (e *rulesetExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Ruleset)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRuleset)
	}

	// Validate that either zone or account is specified
	if cr.Spec.ForProvider.Zone == nil && cr.Spec.ForProvider.Account == nil {
		return managed.ExternalCreation{}, errors.New(errRulesetNoScope)
	}

	cr.SetConditions(rtv1.Creating())

	rs, err := e.client.CreateRuleset(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRulesetCreation)
	}

	cr.Status.AtProvider = ruleset.GenerateObservation(rs)
	meta.SetExternalName(cr, rs.ID)

	return managed.ExternalCreation{}, nil
}

func (e *rulesetExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Ruleset)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRuleset)
	}

	// Validate that either zone or account is specified
	if cr.Spec.ForProvider.Zone == nil && cr.Spec.ForProvider.Account == nil {
		return managed.ExternalUpdate{}, errors.New(errRulesetNoScope)
	}

	rulesetID := meta.GetExternalName(cr)

	// Update should never be called on a nonexistent resource
	if rulesetID == "" {
		return managed.ExternalUpdate{}, errors.New(errRulesetUpdate)
	}

	rs, err := e.client.UpdateRuleset(ctx, rulesetID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errRulesetUpdate)
	}

	cr.Status.AtProvider = ruleset.GenerateObservation(rs)

	return managed.ExternalUpdate{}, nil
}

func (e *rulesetExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Ruleset)
	if !ok {
		return errors.New(errNotRuleset)
	}

	// Validate that either zone or account is specified
	if cr.Spec.ForProvider.Zone == nil && cr.Spec.ForProvider.Account == nil {
		return errors.New(errRulesetNoScope)
	}

	rulesetID := meta.GetExternalName(cr)

	// Delete should never be called on a nonexistent resource
	if rulesetID == "" {
		return errors.New(errRulesetDeletion)
	}

	return errors.Wrap(
		e.client.DeleteRuleset(ctx, rulesetID, cr.Spec.ForProvider),
		errRulesetDeletion)
}