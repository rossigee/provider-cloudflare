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

package pages

import (
	"context"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/pkg/errors"
	pagesv1beta1 "github.com/rossigee/provider-cloudflare/apis/pages/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"github.com/rossigee/provider-cloudflare/internal/clients/pages"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	errNotPagesDomain = "managed resource is not a PagesDomain custom resource"
)

func SetupPagesDomain(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(pagesv1beta1.DomainKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(pagesv1beta1.DomainGroupVersionKind),
		managed.WithExternalConnector(&pagesDomainConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) pages.CloudflarePagesDomainClient {
				return pages.NewDomainClientFromAPI(api)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&pagesv1beta1.Domain{}).
		Complete(r)
}

type pagesDomainConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) pages.CloudflarePagesDomainClient
}

func (c *pagesDomainConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*pagesv1beta1.Domain)
	if !ok {
		return nil, errors.New(errNotPagesDomain)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	api, err := clients.NewClient(*cfg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Cloudflare client")
	}

	return &pagesDomainExternal{
		service: c.newServiceFn(api),
		kube:    c.kube,
	}, nil
}

type pagesDomainExternal struct {
	service pages.CloudflarePagesDomainClient
	kube    client.Client
}

func (c *pagesDomainExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*pagesv1beta1.Domain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPagesDomain)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	params := cloudflare.PagesDomainParameters{
		AccountID:   cr.Spec.ForProvider.AccountID,
		ProjectName: cr.Spec.ForProvider.ProjectName,
		DomainName:  cr.Spec.ForProvider.Domain,
	}

	domain, err := c.service.GetPagesDomain(ctx, params)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get Pages domain")
	}

	cr.Status.AtProvider.ID = domain.ID
	cr.Status.AtProvider.ProjectName = cr.Spec.ForProvider.ProjectName
	cr.Status.AtProvider.Domain = domain.Name
	cr.Status.AtProvider.Status = domain.Status

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesDomainExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*pagesv1beta1.Domain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPagesDomain)
	}

	params := cloudflare.PagesDomainParameters{
		AccountID:   cr.Spec.ForProvider.AccountID,
		ProjectName: cr.Spec.ForProvider.ProjectName,
		DomainName:  cr.Spec.ForProvider.Domain,
	}

	domain, err := c.service.PagesAddDomain(ctx, params)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot add Pages domain")
	}

	meta.SetExternalName(cr, domain.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesDomainExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*pagesv1beta1.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPagesDomain)
	}

	params := cloudflare.PagesDomainParameters{
		AccountID:   cr.Spec.ForProvider.AccountID,
		ProjectName: cr.Spec.ForProvider.ProjectName,
		DomainName:  cr.Spec.ForProvider.Domain,
	}

	_, err := c.service.PagesPatchDomain(ctx, params)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update Pages domain")
	}

	return managed.ExternalUpdate{}, nil
}

func (c *pagesDomainExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*pagesv1beta1.Domain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPagesDomain)
	}

	params := cloudflare.PagesDomainParameters{
		AccountID:   cr.Spec.ForProvider.AccountID,
		ProjectName: cr.Spec.ForProvider.ProjectName,
		DomainName:  cr.Spec.ForProvider.Domain,
	}

	err := c.service.PagesDeleteDomain(ctx, params)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete Pages domain")
	}

	return managed.ExternalDelete{}, nil
}

func (c *pagesDomainExternal) Disconnect(ctx context.Context) error {
	return nil
}
