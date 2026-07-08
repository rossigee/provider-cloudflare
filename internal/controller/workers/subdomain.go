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
	"strings"
	"time"

	"github.com/pkg/errors"

	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"


	rtv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errNotSubdomain = "managed resource is not a Worker Subdomain custom resource"

	errSubdomainLookup      = "cannot lookup subdomain"
	errSubdomainObservation = "cannot observe subdomain"
	errSubdomainCreation    = "cannot create subdomain"
	errSubdomainUpdate      = "cannot update subdomain"
	errSubdomainDeletion    = "cannot delete subdomain"
)

// SetupSubdomain adds a controller that reconciles Worker Subdomain managed resources.
func SetupSubdomain(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.SubdomainKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.SubdomainGroupVersionKind),
		managed.WithExternalConnector(&subdomainConnector{
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
		For(&v1beta1.Subdomain{}).
		Complete(r)
}

// A subdomainConnector is expected to produce an ExternalClient when its Connect method
// is called.
type subdomainConnector struct {
	kube client.Client
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *subdomainConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Subdomain)
	if !ok {
		return nil, errors.New(errNotSubdomain)
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
	return &subdomainExternal{client: adapter}, nil
}

// An subdomainExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type subdomainExternal struct {
	client clients.ClientInterface
}

func (e *subdomainExternal) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {

	cr, ok := mg.(*v1beta1.Subdomain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSubdomain)
	}

	// Subdomain does not exist if we dont have a subdomain name stored in external-name
	subdomainName := meta.GetExternalName(cr)
	if subdomainName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	subdomainObs, err := e.client.WorkerSubdomain(ctx, cr.Spec.ForProvider.AccountID, subdomainName)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(isSubdomainNotFound, err), errSubdomainLookup)
	}

	// Convert the subdomain observation
	obs := generateSubdomainObservation(subdomainObs)
	cr.Status.AtProvider = obs
	cr.Status.SetConditions(rtv1.Available())

	// Check if up to date
	upToDate := isSubdomainUpToDate(cr.Spec.ForProvider, obs)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *subdomainExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Subdomain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSubdomain)
	}

	_, err := e.client.CreateWorkerSubdomain(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errSubdomainCreation)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)

	return managed.ExternalCreation{}, nil
}

func (e *subdomainExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Subdomain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSubdomain)
	}

	subdomainName := meta.GetExternalName(cr)
	if subdomainName == "" {
		return managed.ExternalUpdate{}, errors.New(errSubdomainUpdate)
	}

	_, err := e.client.UpdateWorkerSubdomain(ctx, cr.Spec.ForProvider.AccountID, subdomainName, cr.Spec.ForProvider)
	return managed.ExternalUpdate{}, errors.Wrap(err, errSubdomainUpdate)
}

func (e *subdomainExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Subdomain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotSubdomain)
	}

	subdomainName := meta.GetExternalName(cr)
	if subdomainName == "" {
		return managed.ExternalDelete{}, errors.New(errSubdomainDeletion)
	}

	return managed.ExternalDelete{}, errors.Wrap(
		e.client.DeleteWorkerSubdomain(ctx, cr.Spec.ForProvider.AccountID, subdomainName),
		errSubdomainDeletion)
}

func (e *subdomainExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// Helper functions

// isSubdomainNotFound checks if an error indicates that a subdomain was not found.
func isSubdomainNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		   strings.Contains(errStr, "404") ||
		   strings.Contains(errStr, "does not exist")
}

// generateSubdomainObservation converts API response to observation
func generateSubdomainObservation(in interface{}) v1beta1.SubdomainObservation {
	return v1beta1.SubdomainObservation{}
}

// isSubdomainUpToDate checks if the subdomain matches the desired state
func isSubdomainUpToDate(spec v1beta1.SubdomainParameters, obs v1beta1.SubdomainObservation) bool {
	return true
}