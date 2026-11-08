/*
Copyright 2026 The Crossplane Authors.

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

package aigateway

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/rossigee/provider-cloudflare/apis/aigateway/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	aigatewayclient "github.com/rossigee/provider-cloudflare/internal/clients/aigateway"
)

const (
	errNotGateway     = "managed resource is not an AI Gateway custom resource"
	errGetCreds       = "cannot get credentials"
	errNoAPIToken     = "AI Gateway requires API token authentication"
	errGatewayLookup  = "cannot lookup AI Gateway"
	errGatewayCreate  = "cannot create AI Gateway"
	errGatewayUpdate  = "cannot update AI Gateway"
	errGatewayDelete  = "cannot delete AI Gateway"
	errGatewayCreated = "AI Gateway was created without an identifier"
)

// SetupGateway adds a controller that reconciles AI Gateway managed resources.
func SetupGateway(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.GatewayKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.GatewayGroupVersionKind),
		managed.WithExternalConnector(&gatewayConnector{
			kube:        mgr.GetClient(),
			newClientFn: aigatewayclient.NewClient,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(10*time.Minute),
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Gateway{}).
		Complete(r)
}

// gatewayConnector produces an external client for a Gateway from its ProviderConfig.
type gatewayConnector struct {
	kube        client.Client
	newClientFn func(token string) aigatewayclient.Client
}

// Connect returns an external client that authenticates with the API token from the ProviderConfig.
func (c *gatewayConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	if _, ok := mg.(*v1beta1.Gateway); !ok {
		return nil, errors.New(errNotGateway)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}
	if cfg.AuthByAPIToken == nil || cfg.Token == nil {
		return nil, errors.New(errNoAPIToken)
	}

	return &gatewayExternal{
		client: c.newClientFn(*cfg.Token),
	}, nil
}

type gatewayExternal struct {
	client aigatewayclient.Client
}

func (e *gatewayExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Gateway)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotGateway)
	}

	gatewayID := meta.GetExternalName(cr)
	if gatewayID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	gw, err := e.client.Get(ctx, cr.Spec.ForProvider.AccountID, gatewayID)
	if err != nil {
		if aigatewayclient.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGatewayLookup)
	}

	cr.Status.AtProvider = v1beta1.GatewayObservation{
		ID:         gw.ID,
		CreatedAt:  gw.CreatedAt.Format(time.RFC3339),
		ModifiedAt: gw.ModifiedAt.Format(time.RFC3339),
	}
	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  aigatewayclient.IsUpToDate(cr.Spec.ForProvider, gw),
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *gatewayExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Gateway)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotGateway)
	}

	gw, err := e.client.Create(ctx, cr.Spec.ForProvider.AccountID, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGatewayCreate)
	}
	if gw.ID == "" {
		return managed.ExternalCreation{}, errors.New(errGatewayCreated)
	}

	meta.SetExternalName(cr, gw.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *gatewayExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Gateway)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotGateway)
	}

	_, err := e.client.Update(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr), cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGatewayUpdate)
	}

	return managed.ExternalUpdate{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (e *gatewayExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Gateway)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotGateway)
	}

	err := e.client.Delete(ctx, cr.Spec.ForProvider.AccountID, meta.GetExternalName(cr))
	if err != nil && !aigatewayclient.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, errGatewayDelete)
	}

	return managed.ExternalDelete{}, nil
}

// Disconnect is a no-op because the API client holds no connection state.
func (e *gatewayExternal) Disconnect(ctx context.Context) error {
	return nil
}
