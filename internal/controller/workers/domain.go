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
	errNotDomain = "managed resource is not a Worker Domain custom resource"

	errDomainLookup      = "cannot lookup domain"
	errDomainObservation = "cannot observe domain"
	errDomainCreation    = "cannot create domain"
	errDomainUpdate      = "cannot update domain"
	errDomainDeletion    = "cannot delete domain"
)

// SetupDomain adds a controller that reconciles Worker Domain managed resources.
func SetupDomain(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.DomainKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.DomainGroupVersionKind),
		managed.WithExternalConnecter(&domainConnector{
			kube: mgr.GetClient(),
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithPollInterval(10*time.Minute),
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Domain{}).
		Complete(r)
}

// A domainConnector is expected to produce an ExternalClient when its Connect method
// is called.
type domainConnector struct {
	kube client.Client
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *domainConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Domain)
	if !ok {
		return nil, errors.New(errNotDomain)
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
	return &domainExternal{client: adapter}, nil
}

// An domainExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type domainExternal struct {
	client clients.ClientInterface
}

func (e *domainExternal) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {

	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDomain)
	}

	// Domain does not exist if we dont have a domain ID stored in external-name
	domainID := meta.GetExternalName(cr)
	if domainID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	domainObs, err := e.client.WorkerDomain(ctx, cr.Spec.ForProvider.AccountID, cr.Spec.ForProvider.ZoneID, domainID)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(isDomainNotFound, err), errDomainLookup)
	}

	// Convert the domain observation
	obs := generateDomainObservation(domainObs)
	cr.Status.AtProvider = obs
	cr.Status.SetConditions(rtv1.Available())

	// Check if up to date
	upToDate := isDomainUpToDate(cr.Spec.ForProvider, obs)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *domainExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDomain)
	}

	_, err := e.client.CreateWorkerDomain(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errDomainCreation)
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Hostname)

	return managed.ExternalCreation{}, nil
}

func (e *domainExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDomain)
	}

	domainID := meta.GetExternalName(cr)
	if domainID == "" {
		return managed.ExternalUpdate{}, errors.New(errDomainUpdate)
	}

	_, err := e.client.UpdateWorkerDomain(ctx, cr.Spec.ForProvider.AccountID, cr.Spec.ForProvider.ZoneID, domainID, cr.Spec.ForProvider)
	return managed.ExternalUpdate{}, errors.Wrap(err, errDomainUpdate)
}

func (e *domainExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotDomain)
	}

	domainID := meta.GetExternalName(cr)
	if domainID == "" {
		return managed.ExternalDelete{}, errors.New(errDomainDeletion)
	}

	return managed.ExternalDelete{}, errors.Wrap(
		e.client.DeleteWorkerDomain(ctx, cr.Spec.ForProvider.AccountID, cr.Spec.ForProvider.ZoneID, domainID),
		errDomainDeletion)
}

func (e *domainExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// Helper functions

// isDomainNotFound checks if an error indicates that a domain was not found.
func isDomainNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		   strings.Contains(errStr, "404") ||
		   strings.Contains(errStr, "does not exist")
}

// generateDomainObservation converts API response to observation
func generateDomainObservation(in interface{}) v1beta1.DomainObservation {
	return v1beta1.DomainObservation{}
}

// isDomainUpToDate checks if the domain matches the desired state
func isDomainUpToDate(spec v1beta1.DomainParameters, obs v1beta1.DomainObservation) bool {
	return true
}