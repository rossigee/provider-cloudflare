/*
Copyright 2021 The Crossplane Authors.

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

package filter

import (
	"context"
	"net/http"
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

	"github.com/rossigee/provider-cloudflare/apis/firewall/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotFilter      = "managed resource is not a Filter custom resource"
	errClientConfig   = "error getting client config"
	errFilterLookup   = "cannot lookup filter"
	errFilterCreation = "cannot create filter"
	errFilterUpdate   = "cannot update filter"
	errFilterDeletion = "cannot delete filter"
	errNoZone         = "no zone found"
	errFilterNotFound = "Filter not found"
	maxConcurrency    = 5
)

// Client is a Cloudflare API client that implements methods for working
// with Firewall Filters.
type Client interface {
	Filter(ctx context.Context, zoneID, filterID string) (cloudflare.Filter, error)
	CreateFilter(ctx context.Context, zoneID string, filter cloudflare.Filter) (*cloudflare.Filter, error)
	UpdateFilter(ctx context.Context, zoneID, filterID string, filter cloudflare.Filter) error
	DeleteFilter(ctx context.Context, zoneID, filterID string) error
}

type clientImpl struct {
	cf *cloudflare.API
}

// NewClient returns a new Cloudflare API client for working with Firewall Filters.
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &clientImpl{cf: cf}, nil
}

// Filter retrieves a Firewall Filter
func (c *clientImpl) Filter(ctx context.Context, zoneID, filterID string) (cloudflare.Filter, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	filter, err := c.cf.Filter(ctx, rc, filterID)
	if err != nil {
		return cloudflare.Filter{}, err
	}

	if filter.ID == "" {
		return cloudflare.Filter{}, errors.New(errFilterNotFound)
	}

	return filter, nil
}

// CreateFilter creates a new Firewall Filter
func (c *clientImpl) CreateFilter(ctx context.Context, zoneID string, filter cloudflare.Filter) (*cloudflare.Filter, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	params := []cloudflare.FilterCreateParams{{
		Expression:  filter.Expression,
		Description: filter.Description,
		Paused:      filter.Paused,
	}}
	
	filters, err := c.cf.CreateFilters(ctx, rc, params)
	if err != nil {
		return nil, err
	}
	
	if len(filters) == 0 {
		return nil, errors.New("no filter created")
	}

	return &filters[0], nil
}

// UpdateFilter updates an existing Firewall Filter
func (c *clientImpl) UpdateFilter(ctx context.Context, zoneID, filterID string, filter cloudflare.Filter) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	params := cloudflare.FilterUpdateParams{
		ID:          filterID,
		Expression:  filter.Expression,
		Description: filter.Description,
		Paused:      filter.Paused,
	}
	
	_, err := c.cf.UpdateFilter(ctx, rc, params)
	return err
}

// DeleteFilter deletes a Firewall Filter
func (c *clientImpl) DeleteFilter(ctx context.Context, zoneID, filterID string) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	err := c.cf.DeleteFilter(ctx, rc, filterID)
	return err
}

// IsFilterNotFound returns true if the error indicates the filter was not found
func IsFilterNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errFilterNotFound ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}

// GenerateObservation creates observation data from a Filter
func GenerateObservation(filter cloudflare.Filter) v1alpha1.FilterObservation {
	return v1alpha1.FilterObservation{}
}

// LateInitialize initializes FilterParameters based on the remote resource
func LateInitialize(spec *v1alpha1.FilterParameters, filter cloudflare.Filter) bool {
	if spec == nil {
		return false
	}

	li := false
	if spec.Paused == nil {
		spec.Paused = &filter.Paused
		li = true
	}

	return li
}

// UpToDate checks if the remote Filter is up to date with the requested resource parameters
func UpToDate(spec *v1alpha1.FilterParameters, filter cloudflare.Filter) bool {
	if spec == nil {
		return true
	}

	if spec.Expression != filter.Expression {
		return false
	}

	if spec.Description != nil && *spec.Description != filter.Description {
		return false
	}

	if spec.Paused != nil && *spec.Paused != filter.Paused {
		return false
	}

	return true
}

// CreateFilter creates a Filter from FilterParameters
func CreateFilter(ctx context.Context, client Client, params *v1alpha1.FilterParameters) (*cloudflare.Filter, error) {
	if params.Zone == nil {
		return nil, errors.New("zone is required")
	}

	filter := cloudflare.Filter{
		Expression: params.Expression,
	}

	if params.Description != nil {
		filter.Description = *params.Description
	}

	if params.Paused != nil {
		filter.Paused = *params.Paused
	}

	result, err := client.CreateFilter(ctx, *params.Zone, filter)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateFilter updates an existing Filter
func UpdateFilter(ctx context.Context, client Client, filterID string, params *v1alpha1.FilterParameters) error {
	if params.Zone == nil {
		return errors.New("zone is required")
	}

	filter := cloudflare.Filter{
		Expression: params.Expression,
	}

	if params.Description != nil {
		filter.Description = *params.Description
	}

	if params.Paused != nil {
		filter.Paused = *params.Paused
	}

	err := client.UpdateFilter(ctx, *params.Zone, filterID, filter)
	return err
}

// Setup adds a controller that reconciles Filter managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.FilterGroupKind)

	o := controller.Options{
		RateLimiter:             rl,
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.FilterGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (Client, error) {
				return NewClient(cfg, hc)
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
		For(&v1alpha1.Filter{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (Client, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Filter)
	if !ok {
		return nil, errors.New(errNotFilter)
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

	return &external{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Filter)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotFilter)
	}

	// Filter does not exist if we dont have an ID stored in external-name
	fid := meta.GetExternalName(cr)
	if fid == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errNoZone)
	}

	f, err := e.client.Filter(ctx, *cr.Spec.ForProvider.Zone, fid)

	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(IsFilterNotFound, err), errFilterLookup)
	}

	cr.Status.AtProvider = GenerateObservation(f)

	cr.Status.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: LateInitialize(&cr.Spec.ForProvider, f),
		ResourceUpToDate:        UpToDate(&cr.Spec.ForProvider, f),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Filter)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotFilter)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{}, errors.New(errNoZone)
	}

	nr, err := CreateFilter(ctx, e.client, &cr.Spec.ForProvider)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errFilterCreation)
	}

	cr.Status.AtProvider = GenerateObservation(*nr)

	// Update the external name with the ID of the new Rule
	meta.SetExternalName(cr, nr.ID)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Filter)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotFilter)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errNoZone), errFilterUpdate)
	}

	rid := meta.GetExternalName(cr)

	// Update should never be called on a nonexistent resource
	if rid == "" {
		return managed.ExternalUpdate{}, errors.New(errFilterUpdate)
	}

	return managed.ExternalUpdate{},
		errors.Wrap(
			UpdateFilter(ctx, e.client, meta.GetExternalName(cr), &cr.Spec.ForProvider),
			errFilterUpdate,
		)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Filter)
	if !ok {
		return errors.New(errNotFilter)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return errors.Wrap(errors.New(errNoZone), errFilterDeletion)
	}

	rid := meta.GetExternalName(cr)

	// Delete should never be called on a nonexistent resource
	if rid == "" {
		return errors.New(errFilterDeletion)
	}

	return errors.Wrap(
		e.client.DeleteFilter(ctx, *cr.Spec.ForProvider.Zone, meta.GetExternalName(cr)),
		errFilterDeletion)
}
