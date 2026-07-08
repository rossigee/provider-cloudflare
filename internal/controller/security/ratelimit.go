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

package security

import (
	"context"
	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/security/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/security/botmanagement"
	"github.com/rossigee/provider-cloudflare/internal/clients/security/ratelimit"
	"github.com/rossigee/provider-cloudflare/internal/clients/security/turnstile"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"time"
)

const (
	errNotRateLimit       = "managed resource is not a RateLimit custom resource"
	errNotBotManagement   = "managed resource is not a BotManagement custom resource"
	errNotTurnstile       = "managed resource is not a Turnstile custom resource"
	errGetCreds           = "cannot get credentials"
	errNewRateLimitClient = "cannot create new RateLimit client"
	errNewBotMgmtClient   = "cannot create new BotManagement client"
	errNewTurnstileClient = "cannot create new Turnstile client"
)

// SetupRateLimit adds a controller that reconciles RateLimit managed resources.
func SetupRateLimit(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(securityv1beta1.RateLimitKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(securityv1beta1.RateLimitGroupVersionKind),
		managed.WithExternalConnector(&rateLimitConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) *ratelimit.CloudflareRateLimitClient {
				return ratelimit.NewClientFromAPI(api)
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
		For(&securityv1beta1.RateLimit{}).
		Complete(r)
}

// A rateLimitConnector is expected to produce an ExternalClient when its Connect method
// is called.
type rateLimitConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) *ratelimit.CloudflareRateLimitClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *rateLimitConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*securityv1beta1.RateLimit)
	if !ok {
		return nil, errors.New(errNotRateLimit)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewRateLimitClient)
	}

	// Create the rate limit client
	return &rateLimitExternal{service: c.newServiceFn(client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type rateLimitExternal struct {
	service *ratelimit.CloudflareRateLimitClient
}

func (c *rateLimitExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*securityv1beta1.RateLimit)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRateLimit)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.Zone, meta.GetExternalName(cr))
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

func (c *rateLimitExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*securityv1beta1.RateLimit)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRateLimit)
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

func (c *rateLimitExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*securityv1beta1.RateLimit)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRateLimit)
	}

	obs, err := c.service.Update(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *rateLimitExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*securityv1beta1.RateLimit)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRateLimit)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	return managed.ExternalDelete{}, c.service.Delete(ctx, cr.Spec.ForProvider.Zone, meta.GetExternalName(cr))
}

func (c *rateLimitExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// SetupBotManagement adds a controller that reconciles BotManagement managed resources.
func SetupBotManagement(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(securityv1beta1.BotManagementKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(securityv1beta1.BotManagementGroupVersionKind),
		managed.WithExternalConnector(&botManagementConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) *botmanagement.CloudflareBotManagementClient {
				return botmanagement.NewClientFromAPI(api)
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
		For(&securityv1beta1.BotManagement{}).
		Complete(r)
}

// A botManagementConnector is expected to produce an ExternalClient when its Connect method
// is called.
type botManagementConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) *botmanagement.CloudflareBotManagementClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *botManagementConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*securityv1beta1.BotManagement)
	if !ok {
		return nil, errors.New(errNotBotManagement)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewBotMgmtClient)
	}

	// Create the bot management client
	return &botManagementExternal{service: c.newServiceFn(client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type botManagementExternal struct {
	service *botmanagement.CloudflareBotManagementClient
}

func (c *botManagementExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*securityv1beta1.BotManagement)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBotManagement)
	}

	// Bot Management is a zone-level configuration, it always "exists"
	// We just need to get the current configuration
	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.Zone)
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

func (c *botManagementExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*securityv1beta1.BotManagement)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBotManagement)
	}

	cr.Status.SetConditions(rtv1.Creating())

	// Bot Management is a configuration, not a created resource, so we just update it
	obs, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create external resource")
	}

	cr.Status.AtProvider = *obs
	// For Bot Management, we use the zone ID as the external name
	meta.SetExternalName(cr, cr.Spec.ForProvider.Zone)

	return managed.ExternalCreation{}, nil
}

func (c *botManagementExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*securityv1beta1.BotManagement)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotBotManagement)
	}

	obs, err := c.service.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *botManagementExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	// Bot Management is a zone-level configuration, we don't delete it
	// We could reset it to default values, but that might not be desired
	// For now, we'll just mark it as deleting but not actually change anything
	cr, ok := mg.(*securityv1beta1.BotManagement)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBotManagement)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	// Successfully "delete" by doing nothing - the configuration remains
	return managed.ExternalDelete{}, nil
}

func (c *botManagementExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// SetupTurnstile adds a controller that reconciles Turnstile managed resources.
func SetupTurnstile(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(securityv1beta1.TurnstileKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(securityv1beta1.TurnstileGroupVersionKind),
		managed.WithExternalConnector(&turnstileConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) *turnstile.CloudflareTurnstileClient {
				return turnstile.NewClientFromAPI(api)
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
		For(&securityv1beta1.Turnstile{}).
		Complete(r)
}

// A turnstileConnector is expected to produce an ExternalClient when its Connect method
// is called.
type turnstileConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) *turnstile.CloudflareTurnstileClient
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *turnstileConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*securityv1beta1.Turnstile)
	if !ok {
		return nil, errors.New(errNotTurnstile)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	client, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewTurnstileClient)
	}

	// Create the turnstile client
	return &turnstileExternal{service: c.newServiceFn(client)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type turnstileExternal struct {
	service *turnstile.CloudflareTurnstileClient
}

func (c *turnstileExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*securityv1beta1.Turnstile)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTurnstile)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr))
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

func (c *turnstileExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*securityv1beta1.Turnstile)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotTurnstile)
	}

	cr.Status.SetConditions(rtv1.Creating())

	obs, err := c.service.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create external resource")
	}

	cr.Status.AtProvider = *obs
	if obs.SiteKey != nil {
		meta.SetExternalName(cr, *obs.SiteKey)
	}

	return managed.ExternalCreation{}, nil
}

func (c *turnstileExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*securityv1beta1.Turnstile)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotTurnstile)
	}

	obs, err := c.service.Update(ctx, meta.GetExternalName(cr), cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update external resource")
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{}, nil
}

func (c *turnstileExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*securityv1beta1.Turnstile)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotTurnstile)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	return managed.ExternalDelete{}, c.service.Delete(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr))
}

func (c *turnstileExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// Setup adds controllers for Security resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	if err := SetupRateLimit(mgr, l, rl); err != nil {
		return err
	}
	if err := SetupBotManagement(mgr, l, rl); err != nil {
		return err
	}
	return SetupTurnstile(mgr, l, rl)
}
