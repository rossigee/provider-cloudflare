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

	providerv1alpha1 "github.com/rossigee/provider-cloudflare/apis/v1alpha1"
	workersv1alpha1 "github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	kvnamespace "github.com/rossigee/provider-cloudflare/internal/clients/workers/kvnamespace"
)

const (
	errNotKVNamespace        = "managed resource is not a KV Namespace custom resource"
	errTrackPCUsageKV        = "cannot track ProviderConfig usage"
	errGetPCKV               = "cannot get ProviderConfig"
	errGetCredsKV            = "cannot get credentials"
	errNewKVNamespaceClient  = "cannot create new KV Namespace client"
)

// SetupKVNamespace adds a controller that reconciles KVNamespace managed resources.
func SetupKVNamespace(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(workersv1alpha1.KVNamespaceGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(workersv1alpha1.KVNamespaceGroupVersionKind),
		managed.WithExternalConnecter(&kvConnector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &providerv1alpha1.ProviderConfigUsage{}),
			newServiceFn: kvnamespace.NewClient,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: nil, // Use default rate limiter
		}).
		For(&workersv1alpha1.KVNamespace{}).
		Complete(r)
}

// A kvConnector is expected to produce an ExternalClient when its Connect method
// is called.
type kvConnector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(clients.ClientInterface) *kvnamespace.KVNamespaceClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *kvConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*workersv1alpha1.KVNamespace)
	if !ok {
		return nil, errors.New(errNotKVNamespace)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsageKV)
	}

	pc := &providerv1alpha1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPCKV)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCredsKV)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewKVNamespaceClient)
	}

	// Create the KV namespace client wrapper
	adapter := clients.NewCloudflareAPIAdapter(client)
	return &kvExternal{service: c.newServiceFn(adapter)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type kvExternal struct {
	service *kvnamespace.KVNamespaceClient
}

func (c *kvExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*workersv1alpha1.KVNamespace)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotKVNamespace)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	obs, err := c.service.Get(ctx, meta.GetExternalName(cr))
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

func (c *kvExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*workersv1alpha1.KVNamespace)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotKVNamespace)
	}

	cr.Status.SetConditions(rtv1.Creating())

	obs, err := c.service.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create external resource")
	}

	cr.Status.AtProvider = *obs
	meta.SetExternalName(cr, obs.ID)

	return managed.ExternalCreation{}, nil
}

func (c *kvExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*workersv1alpha1.KVNamespace)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotKVNamespace)
	}

	obs, err := c.service.Update(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *kvExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*workersv1alpha1.KVNamespace)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotKVNamespace)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	err := c.service.Delete(ctx, meta.GetExternalName(cr))
	return managed.ExternalDelete{}, err
}

func (c *kvExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}