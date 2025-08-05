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

package ssl

import (
	"context"

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

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/ssl/certificatepack"
)

const (
	errNotCertificatePack = "managed resource is not a Certificate Pack custom resource"
	errTrackPCUsageCert   = "cannot track ProviderConfig usage"
	errGetPCCert          = "cannot get ProviderConfig"
	errGetCredsCert       = "cannot get credentials"
	errNewClientCert      = "cannot create new Service"
)

// SetupCertificatePackController adds a controller that reconciles Certificate Pack managed resources.
func SetupCertificatePackController(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.CertificatePackKind)

	o := controller.Options{
		RateLimiter: nil, // Use default rate limiter
	}

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CertificatePackGroupVersionKind),
		managed.WithExternalConnecter(&certificatePackConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (*cloudflare.API, error) {
				return clients.NewClient(cfg, nil)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1alpha1.CertificatePack{}).
		Complete(r)
}

// A certificatePackConnector is expected to produce an ExternalClient when its Connect method
// is called.
type certificatePackConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
}

// Connect typically produces an ExternalClient by:
// 1. Getting the managed resource's ProviderConfig.
// 2. Getting the credentials specified by the ProviderConfig.
// 3. Using the credentials to form a client.
func (c *certificatePackConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.CertificatePack)
	if !ok {
		return nil, errors.New(errNotCertificatePack)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCredsCert)
	}

	cloudflareClient, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, errors.Wrap(err, errNewClientCert)
	}

	service := certificatepack.NewClient(cloudflareClient)

	return &certificatePackExternal{service: service}, nil
}

// An certificatePackExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type certificatePackExternal struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *certificatepack.CloudflareCertificatePackClient
}

func (c *certificatePackExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CertificatePack)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCertificatePack)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	observation, err := c.service.Get(ctx, cr.Spec.ForProvider.Zone, meta.GetExternalName(cr))
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get Certificate Pack")
	}

	cr.Status.AtProvider = *observation

	cr.Status.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true, // Certificate packs don't have updatable parameters after creation
	}, nil
}

func (c *certificatePackExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CertificatePack)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCertificatePack)
	}

	cr.Status.SetConditions(rtv1.Creating())

	observation, err := c.service.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create Certificate Pack")
	}

	cr.Status.AtProvider = *observation

	if observation.ID != nil {
		meta.SetExternalName(cr, *observation.ID)
	}

	return managed.ExternalCreation{}, nil
}

func (c *certificatePackExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CertificatePack)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCertificatePack)
	}

	// Certificate packs generally don't support updates to their configuration
	// The only supported operation is restarting validation
	if cr.Status.AtProvider.Status != nil && *cr.Status.AtProvider.Status == "pending_validation" {
		observation, err := c.service.RestartValidation(ctx, cr.Spec.ForProvider.Zone, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to restart certificate validation")
		}

		cr.Status.AtProvider = *observation
	}

	return managed.ExternalUpdate{}, nil
}

func (c *certificatePackExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.CertificatePack)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotCertificatePack)
	}

	cr.Status.SetConditions(rtv1.Deleting())

	err := c.service.Delete(ctx, cr.Spec.ForProvider.Zone, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete Certificate Pack")
	}

	return managed.ExternalDelete{}, nil
}

func (c *certificatePackExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}