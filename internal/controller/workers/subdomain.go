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

package workers

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
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

	workersv1alpha1 "github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	providerv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	subdomain "github.com/rossigee/provider-cloudflare/internal/clients/workers/subdomain"
)

const (
	errNotSubdomain           = "managed resource is not a Subdomain custom resource"
	errTrackPCUsageSubdomain  = "cannot track ProviderConfig usage"
	errGetPCSubdomain         = "cannot get ProviderConfig"
	errGetCredsSubdomain      = "cannot get credentials"
	errNewSubdomainClient     = "cannot create new Subdomain client"
)

// SetupSubdomain adds a controller that reconciles Subdomain managed resources.
func SetupSubdomain(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(workersv1alpha1.SubdomainKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(workersv1alpha1.SubdomainGroupVersionKind),
		managed.WithExternalConnecter(&subdomainConnector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &providerv1alpha1.ProviderConfigUsage{}),
			newServiceFn: subdomain.NewClient,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: nil, // Use default rate limiter
		}).
		For(&workersv1alpha1.Subdomain{}).
		Complete(r)
}

// A subdomainConnector is expected to produce an ExternalClient when its Connect method
// is called.
type subdomainConnector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(*cloudflare.API) *subdomain.CloudflareSubdomainClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *subdomainConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*workersv1alpha1.Subdomain)
	if !ok {
		return nil, errors.New(errNotSubdomain)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsageSubdomain)
	}

	pc := &providerv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPCSubdomain)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCredsSubdomain)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewSubdomainClient)
	}

	// Create the subdomain client
	return &subdomainExternal{service: c.newServiceFn(client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type subdomainExternal struct {
	service *subdomain.CloudflareSubdomainClient
}

func (c *subdomainExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*workersv1alpha1.Subdomain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubdomain)
	}

	// Workers Subdomain is an account-level configuration, it always "exists"
	// We just need to get the current configuration
	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.AccountID)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(clients.IsNotFound, err), "cannot get external resource")
	}

	cr.Status.AtProvider = *obs

	cr.Status.SetConditions(rtv1.Available())

	upToDate, err := c.service.IsUpToDate(ctx, cr.Spec.ForProvider, *obs)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot determine if resource is up to date")
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *subdomainExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*workersv1alpha1.Subdomain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubdomain)
	}

	cr.Status.SetConditions(rtv1.Creating())

	// Workers Subdomain is a configuration, not a created resource, so we just update it
	obs, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create external resource")
	}

	cr.Status.AtProvider = *obs
	// For Workers Subdomain, we use the account ID as the external name
	meta.SetExternalName(cr, cr.Spec.ForProvider.AccountID)

	return managed.ExternalCreation{}, nil
}

func (c *subdomainExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*workersv1alpha1.Subdomain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubdomain)
	}

	obs, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *subdomainExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	// Workers Subdomain is an account-level configuration, we don't delete it
	// We could reset it to empty, but that might not be desired
	// For now, we'll just mark it as deleting but not actually change anything
	cr, ok := mg.(*workersv1alpha1.Subdomain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSubdomain)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	// Successfully "delete" by doing nothing - the configuration remains
	return managed.ExternalDelete{}, nil
}

func (c *subdomainExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}