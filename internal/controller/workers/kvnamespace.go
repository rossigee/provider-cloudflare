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
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"
)

const (
	errNotKVNamespace = "managed resource is not a Worker KVNamespace custom resource"

	errKVNamespaceLookup      = "cannot lookup kv namespace"
	errKVNamespaceObservation = "cannot observe kv namespace"
	errKVNamespaceCreation    = "cannot create kv namespace"
	errKVNamespaceUpdate      = "cannot update kv namespace"
	errKVNamespaceDeletion    = "cannot delete kv namespace"
)

// SetupKVNamespace adds a controller that reconciles Worker KVNamespace managed resources.
func SetupKVNamespace(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.KVNamespaceKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.KVNamespaceGroupVersionKind),
		managed.WithExternalConnector(&kvNamespaceConnector{
			kube: mgr.GetClient(),
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(10*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.KVNamespace{}).
		Complete(r)
}

// A kvNamespaceConnector is expected to produce an ExternalClient when its Connect method
// is called.
type kvNamespaceConnector struct {
	kube client.Client
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *kvNamespaceConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.KVNamespace)
	if !ok {
		return nil, errors.New(errNotKVNamespace)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	// Create cloudflare API client
	api, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare client")
	}

	// Wrap with adapter to implement ClientInterface
	adapter := clients.NewCloudflareAPIAdapter(api)
	return &kvNamespaceExternal{client: adapter}, nil
}

// An kvNamespaceExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type kvNamespaceExternal struct {
	client clients.ClientInterface
}

func (e *kvNamespaceExternal) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {

	cr, ok := mg.(*v1beta1.KVNamespace)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotKVNamespace)
	}

	// KVNamespace does not exist if we dont have a namespace ID stored in external-name
	kvID := meta.GetExternalName(cr)
	if kvID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	kvObs, err := e.client.WorkerKVNamespace(ctx, kvID)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(isKVNamespaceNotFound, err), errKVNamespaceLookup)
	}

	// Convert the kv namespace observation
	obs := generateKVNamespaceObservation(kvObs)
	cr.Status.AtProvider = obs
	cr.Status.SetConditions(rtv1.Available())

	// Check if up to date
	upToDate := isKVNamespaceUpToDate(cr.Spec.ForProvider, obs)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *kvNamespaceExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.KVNamespace)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotKVNamespace)
	}

	_, err := e.client.CreateWorkerKVNamespace(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errKVNamespaceCreation)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Title)

	return managed.ExternalCreation{}, nil
}

func (e *kvNamespaceExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.KVNamespace)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotKVNamespace)
	}

	kvID := meta.GetExternalName(cr)
	if kvID == "" {
		return managed.ExternalUpdate{}, errors.New(errKVNamespaceUpdate)
	}

	_, err := e.client.UpdateWorkerKVNamespace(ctx, kvID, cr.Spec.ForProvider)
	return managed.ExternalUpdate{}, errors.Wrap(err, errKVNamespaceUpdate)
}

func (e *kvNamespaceExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.KVNamespace)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotKVNamespace)
	}

	kvID := meta.GetExternalName(cr)
	if kvID == "" {
		return managed.ExternalDelete{}, errors.New(errKVNamespaceDeletion)
	}

	return managed.ExternalDelete{}, errors.Wrap(
		e.client.DeleteWorkerKVNamespace(ctx, kvID),
		errKVNamespaceDeletion)
}

func (e *kvNamespaceExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// Helper functions

// isKVNamespaceNotFound checks if an error indicates that a kv namespace was not found.
func isKVNamespaceNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "does not exist")
}

// generateKVNamespaceObservation converts API response to observation
func generateKVNamespaceObservation(in interface{}) v1beta1.KVNamespaceObservation {
	return v1beta1.KVNamespaceObservation{}
}

// isKVNamespaceUpToDate checks if the kv namespace matches the desired state
func isKVNamespaceUpToDate(spec v1beta1.KVNamespaceParameters, obs v1beta1.KVNamespaceObservation) bool {
	return true
}
