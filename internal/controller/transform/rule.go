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
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/transform/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/transform/rule"
	"github.com/rossigee/provider-cloudflare/internal/metrics"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"
)

const (
	errNotRule = "managed resource is not a Transform Rule custom resource"

	errClientConfig = "error getting client config"

	errRuleLookup   = "cannot lookup Transform Rule"
	errRuleCreation = "cannot create Transform Rule"
	errRuleUpdate   = "cannot update Transform Rule"
	errRuleDeletion = "cannot delete Transform Rule"
	errRuleNoZone   = "no zone found"

	maxConcurrency = 5
)

// Setup adds a controller that reconciles Transform Rule managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.RuleGroupKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RuleGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube: mgr.GetClient(),
			newTransformRuleClientFn: func(cfg clients.Config) (transformrule.Client, error) {
				return transformrule.NewClient(cfg, hc)
			},
		}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		// Do not initialize external-name field.
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Rule{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube                     client.Client
	newTransformRuleClientFn func(cfg clients.Config) (transformrule.Client, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Rule)
	if !ok {
		return nil, errors.New(errNotRule)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	client, err := c.newTransformRuleClientFn(*config)
	if err != nil {
		return nil, err
	}

	return &external{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client transformrule.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Rule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRule)
	}
	_, span := tracing.StartSpan(ctx, "rule.observe",
		tracing.SpanAttrs("rule", cr.GetName(), "observe")...)
	defer span.End()

	// Rule does not exist if we don't have an ID stored in external-name
	rid := meta.GetExternalName(cr)
	if rid == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errRuleNoZone)
	}

	rule, err := e.client.GetTransformRule(ctx, *cr.Spec.ForProvider.Zone, rid, cr.Spec.ForProvider.Phase)
	if err != nil {
		if transformrule.IsRuleNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errRuleLookup)
	}

	cr.Status.AtProvider = transformrule.GenerateObservation(rule, "")
	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: transformrule.UpToDate(&cr.Spec.ForProvider, rule),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Rule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRule)
	}
	_, span := tracing.StartSpan(ctx, "rule.create",
		tracing.SpanAttrs("rule", cr.GetName(), "create")...)
	defer span.End()

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{}, errors.Wrap(errors.New(errRuleNoZone), errRuleCreation)
	}

	cr.SetConditions(rtv1.Creating())

	rule, err := e.client.CreateTransformRule(ctx, *cr.Spec.ForProvider.Zone, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRuleCreation)
	}

	// Update the external name with the ID of the new Rule
	meta.SetExternalName(cr, rule.ID)
	cr.Status.AtProvider = transformrule.GenerateObservation(rule, "")

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Rule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRule)
	}
	_, span := tracing.StartSpan(ctx, "rule.update",
		tracing.SpanAttrs("rule", cr.GetName(), "update")...)
	defer span.End()

	rid := meta.GetExternalName(cr)
	if rid == "" {
		return managed.ExternalUpdate{}, errors.New(errRuleUpdate)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errRuleNoZone), errRuleUpdate)
	}

	rule, err := e.client.UpdateTransformRule(ctx, *cr.Spec.ForProvider.Zone, rid, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errRuleUpdate)
	}

	cr.Status.AtProvider = transformrule.GenerateObservation(rule, "")

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Rule)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRule)
	}
	_, span := tracing.StartSpan(ctx, "rule.delete",
		tracing.SpanAttrs("rule", cr.GetName(), "delete")...)
	defer span.End()

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalDelete{}, errors.Wrap(errors.New(errRuleNoZone), errRuleDeletion)
	}

	rid := meta.GetExternalName(cr)
	if rid == "" {
		return managed.ExternalDelete{}, errors.New(errRuleDeletion)
	}

	err := e.client.DeleteTransformRule(ctx, *cr.Spec.ForProvider.Zone, rid, cr.Spec.ForProvider.Phase)
	return managed.ExternalDelete{}, errors.Wrap(err, errRuleDeletion)
}
func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}
