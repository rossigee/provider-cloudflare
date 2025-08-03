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
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/cache/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/cache"
	"github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotCacheRule = "managed resource is not a CacheRule custom resource"
	errGetCreds     = "failed to get provider credentials"
	errNewClient    = "failed to create cache rule client"
)

// SetupCacheRule adds a controller that reconciles CacheRule managed resources.
func SetupCacheRule(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.CacheRuleGroupKind)

	o := controller.Options{
		RateLimiter:             rl,
		MaxConcurrentReconciles: 5,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CacheRuleGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			newClientFn: func(cfg clients.Config) (cache.CacheRuleClient, error) {
				return cache.NewCacheRuleClient(cfg, hc)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(5*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CacheRule{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube        client.Client
	newClientFn func(cfg clients.Config) (cache.CacheRuleClient, error)
}

// Connect typically produces an ExternalClient by:
// 1. Getting the managed resource's ProviderConfig.
// 2. Getting the credentials specified by the ProviderConfig.
// 3. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.CacheRule)
	if !ok {
		return nil, errors.New(errNotCacheRule)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newClientFn(*cfg)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service cache.CacheRuleClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CacheRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCacheRule)
	}

	rulesetID := cr.Status.AtProvider.RulesetID
	ruleID := cr.Status.AtProvider.ID

	if ruleID == "" || rulesetID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	rule, ruleset, err := c.service.GetCacheRule(ctx, rulesetID, ruleID, cr.Spec.ForProvider)
	if err != nil {
		if cache.IsCacheRuleNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get cache rule from Cloudflare API")
	}

	cr.Status.AtProvider = cache.GenerateCacheRuleObservation(rule, ruleset)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        cache.IsCacheRuleUpToDate(&cr.Spec.ForProvider, rule),
		ResourceLateInitialized: true,
		ConnectionDetails:       managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CacheRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCacheRule)
	}

	rule, ruleset, err := c.service.CreateCacheRule(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create cache rule in Cloudflare API")
	}

	cr.Status.AtProvider = cache.GenerateCacheRuleObservation(rule, ruleset)
	meta.SetExternalName(cr, rule.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CacheRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCacheRule)
	}

	rulesetID := cr.Status.AtProvider.RulesetID
	ruleID := cr.Status.AtProvider.ID

	rule, ruleset, err := c.service.UpdateCacheRule(ctx, rulesetID, ruleID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update cache rule in Cloudflare API")
	}

	cr.Status.AtProvider = cache.GenerateCacheRuleObservation(rule, ruleset)

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CacheRule)
	if !ok {
		return errors.New(errNotCacheRule)
	}

	rulesetID := cr.Status.AtProvider.RulesetID
	ruleID := cr.Status.AtProvider.ID

	if ruleID == "" || rulesetID == "" {
		return nil // Already deleted or never created
	}

	err := c.service.DeleteCacheRule(ctx, rulesetID, ruleID, cr.Spec.ForProvider)
	if err != nil && !cache.IsCacheRuleNotFound(err) {
		return errors.Wrap(err, "failed to delete cache rule from Cloudflare API")
	}

	return nil
}