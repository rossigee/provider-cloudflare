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

package loadbalancing

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-cloudflare/apis/v1beta1"

	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/loadbalancing"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"
)

const (
	errNotLoadBalancer = "managed resource is not a LoadBalancer custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errNewClient       = "cannot create new Service"
)

// SetupLoadBalancer adds a controller that reconciles LoadBalancer managed resources.
func SetupLoadBalancer(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.LoadBalancerGroupKind.String())

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.LoadBalancerGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube: mgr.GetClient(),
			newServiceFn: func(cfg clients.Config, httpClient *http.Client) (loadbalancing.LoadBalancerClient, error) {
				return loadbalancing.NewLoadBalancerClient(cfg, httpClient)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.LoadBalancer{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	newServiceFn func(cfg clients.Config, httpClient *http.Client) (loadbalancing.LoadBalancerClient, error)
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.LoadBalancer)
	if !ok {
		return nil, errors.New(errNotLoadBalancer)
	}

	pc := &apisv1beta1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc, err := c.newServiceFn(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: svc, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service loadbalancing.LoadBalancerClient
	kube    client.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.LoadBalancer)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotLoadBalancer)
	}
	_, span := tracing.StartSpan(ctx, "loadbalancer.observe",
		tracing.SpanAttrs("loadbalancer", cr.GetName(), "observe")...)
	defer span.End()

	if cr.Status.AtProvider.ID == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalObservation{}, err
	}

	lb, err := c.service.GetLoadBalancer(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		if loadbalancing.IsLoadBalancerNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get load balancer from Cloudflare API")
	}

	cr.Status.AtProvider = loadbalancing.GenerateLoadBalancerObservation(lb)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        loadbalancing.IsLoadBalancerUpToDate(&cr.Spec.ForProvider, lb),
		ResourceLateInitialized: c.lateInitialize(&cr.Spec.ForProvider, lb),
		ConnectionDetails:       managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.LoadBalancer)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotLoadBalancer)
	}
	_, span := tracing.StartSpan(ctx, "loadbalancer.create",
		tracing.SpanAttrs("loadbalancer", cr.GetName(), "create")...)
	defer span.End()

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalCreation{}, err
	}

	lb, err := c.service.CreateLoadBalancer(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create load balancer in Cloudflare API")
	}

	cr.Status.AtProvider = loadbalancing.GenerateLoadBalancerObservation(lb)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.LoadBalancer)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotLoadBalancer)
	}
	_, span := tracing.StartSpan(ctx, "loadbalancer.update",
		tracing.SpanAttrs("loadbalancer", cr.GetName(), "update")...)
	defer span.End()

	// Resolve references before making API call
	if err := c.resolveReferences(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	_, err := c.service.UpdateLoadBalancer(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update load balancer in Cloudflare API")
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.LoadBalancer)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotLoadBalancer)
	}
	_, span := tracing.StartSpan(ctx, "loadbalancer.delete",
		tracing.SpanAttrs("loadbalancer", cr.GetName(), "delete")...)
	defer span.End()

	err := c.service.DeleteLoadBalancer(ctx, cr.Status.AtProvider.ID, cr.Spec.ForProvider)
	if err != nil && !loadbalancing.IsLoadBalancerNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete load balancer from Cloudflare API")
	}

	return managed.ExternalDelete{}, nil
}

func (c *external) lateInitialize(spec *v1beta1.LoadBalancerParameters, lb *cloudflare.LoadBalancer) bool {
	li := false

	if spec.Name == nil && lb.Name != "" {
		spec.Name = &lb.Name
		li = true
	}

	if spec.Description == nil && lb.Description != "" {
		spec.Description = &lb.Description
		li = true
	}

	if spec.TTL == nil && lb.TTL != 0 {
		spec.TTL = &lb.TTL
		li = true
	}

	if spec.Proxied == nil {
		spec.Proxied = &lb.Proxied
		li = true
	}

	if spec.Enabled == nil {
		spec.Enabled = lb.Enabled
		li = true
	}

	if spec.SessionAffinity == nil && lb.Persistence != "" {
		spec.SessionAffinity = &lb.Persistence
		li = true
	}

	if spec.SessionAffinityTTL == nil && lb.PersistenceTTL != 0 {
		spec.SessionAffinityTTL = &lb.PersistenceTTL
		li = true
	}

	if spec.SteeringPolicy == nil && lb.SteeringPolicy != "" {
		spec.SteeringPolicy = &lb.SteeringPolicy
		li = true
	}

	return li
}

func (c *external) resolveReferences(ctx context.Context, cr *v1beta1.LoadBalancer) error {
	// Resolve FallbackPoolRef
	if cr.Spec.ForProvider.FallbackPoolRef != nil {
		r := cr.Spec.ForProvider.FallbackPoolRef
		pool := &v1beta1.LoadBalancerPool{}
		if err := c.kube.Get(ctx, types.NamespacedName{Name: r.Name}, pool); err != nil {
			return errors.Wrap(err, "cannot get referenced fallback pool")
		}
		if pool.Status.AtProvider.ID == "" {
			return errors.New("referenced fallback pool does not have an ID yet")
		}
		cr.Spec.ForProvider.FallbackPool = &pool.Status.AtProvider.ID
	}

	// Resolve DefaultPoolRefs
	if len(cr.Spec.ForProvider.DefaultPoolRefs) > 0 {
		poolIDs := make([]string, 0, len(cr.Spec.ForProvider.DefaultPoolRefs))
		for _, ref := range cr.Spec.ForProvider.DefaultPoolRefs {
			pool := &v1beta1.LoadBalancerPool{}
			if err := c.kube.Get(ctx, types.NamespacedName{Name: ref.Name}, pool); err != nil {
				return errors.Wrap(err, fmt.Sprintf("cannot get referenced default pool %s", ref.Name))
			}
			if pool.Status.AtProvider.ID == "" {
				return errors.New(fmt.Sprintf("referenced default pool %s does not have an ID yet", ref.Name))
			}
			poolIDs = append(poolIDs, pool.Status.AtProvider.ID)
		}
		cr.Spec.ForProvider.DefaultPools = poolIDs
	}

	// Resolve DefaultPoolSelector
	if cr.Spec.ForProvider.DefaultPoolSelector != nil {
		pools := &v1beta1.LoadBalancerPoolList{}
		if err := c.kube.List(ctx, pools, client.MatchingLabels(cr.Spec.ForProvider.DefaultPoolSelector.MatchLabels)); err != nil {
			return errors.Wrap(err, "cannot list pools for default pool selector")
		}
		poolIDs := make([]string, 0, len(pools.Items))
		for _, pool := range pools.Items {
			if pool.Status.AtProvider.ID == "" {
				continue // Skip pools that don't have IDs yet
			}
			poolIDs = append(poolIDs, pool.Status.AtProvider.ID)
		}
		cr.Spec.ForProvider.DefaultPools = poolIDs
	}

	// Resolve FallbackPoolSelector
	if cr.Spec.ForProvider.FallbackPoolSelector != nil {
		pools := &v1beta1.LoadBalancerPoolList{}
		if err := c.kube.List(ctx, pools, client.MatchingLabels(cr.Spec.ForProvider.FallbackPoolSelector.MatchLabels)); err != nil {
			return errors.Wrap(err, "cannot list pools for fallback pool selector")
		}
		if len(pools.Items) > 0 && pools.Items[0].Status.AtProvider.ID != "" {
			cr.Spec.ForProvider.FallbackPool = &pools.Items[0].Status.AtProvider.ID
		}
	}

	return nil
}
func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}
