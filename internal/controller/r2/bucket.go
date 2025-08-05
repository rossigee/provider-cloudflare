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

package r2

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go"
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

	"github.com/rossigee/provider-cloudflare/apis/r2/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	bucketclient "github.com/rossigee/provider-cloudflare/internal/clients/r2/bucket"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotBucket = "managed resource is not a Bucket custom resource"

	errBucketClientConfig = "error getting bucket client config"

	errBucketLookup   = "cannot lookup Bucket"
	errBucketCreation = "cannot create Bucket"
	errBucketUpdate   = "cannot update Bucket"
	errBucketDeletion = "cannot delete Bucket"

	bucketMaxConcurrency = 5
)

// SetupBucket adds a controller that reconciles Bucket managed resources.
func SetupBucket(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1alpha1.BucketKind)

	o := controller.Options{
		RateLimiter: nil, // Use default rate limiter
		MaxConcurrentReconciles: bucketMaxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.BucketGroupVersionKind),
		managed.WithExternalConnecter(&bucketConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (*cloudflare.API, error) {
				return clients.NewClient(cfg, hc)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(5*time.Minute),
		// Do not initialize external-name field.
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.Bucket{}).
		Complete(r)
}

// A bucketConnector is expected to produce an ExternalClient when its Connect method
// is called.
type bucketConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *bucketConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return nil, errors.New(errNotBucket)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errBucketClientConfig)
	}

	client, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, err
	}

	// Create the bucket client wrapper
	bucketClient := bucketclient.NewClient(client)

	return &bucketExternal{client: bucketClient}, nil
}

// An bucketExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type bucketExternal struct {
	client *bucketclient.BucketClient
}

func (c *bucketExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBucket)
	}

	// Bucket does not exist if we don't have an ID stored in external-name
	bucketName := meta.GetExternalName(cr)
	if bucketName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	observation, err := c.client.Get(ctx, bucketName)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(bucketclient.IsBucketNotFound, err), errBucketLookup)
	}

	cr.Status.AtProvider = *observation
	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: observation.Name == cr.Spec.ForProvider.Name,
	}, nil
}

func (c *bucketExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBucket)
	}

	cr.SetConditions(rtv1.Creating())

	observation, err := c.client.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errBucketCreation)
	}

	// Update the external name with the bucket name
	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
	cr.Status.AtProvider = *observation

	return managed.ExternalCreation{}, nil
}

func (c *bucketExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBucket)
	}

	// R2 buckets don't support updates beyond creation parameters
	// If updates are needed, the bucket would need to be recreated
	return managed.ExternalUpdate{}, nil
}

func (c *bucketExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Bucket)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBucket)
	}

	bucketName := meta.GetExternalName(cr)
	if bucketName == "" {
		// Nothing to delete if no external name is set
		return managed.ExternalDelete{}, nil
	}

	err := c.client.Delete(ctx, bucketName)
	if err != nil && !bucketclient.IsBucketNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errBucketDeletion)
	}

	return managed.ExternalDelete{}, nil
}

func (c *bucketExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}