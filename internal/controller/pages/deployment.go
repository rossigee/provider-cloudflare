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
	errNotPagesDeployment = "managed resource is not a PagesDeployment custom resource"
)

func SetupPagesDeployment(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(pagesv1beta1.DeploymentKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(pagesv1beta1.DeploymentGroupVersionKind),
		managed.WithExternalConnector(&pagesDeploymentConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) pages.CloudflarePagesDeploymentClient {
				return pages.NewDeploymentClientFromAPI(api)
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
		For(&pagesv1beta1.Deployment{}).
		Complete(r)
}

type pagesDeploymentConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) pages.CloudflarePagesDeploymentClient
}

func (c *pagesDeploymentConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*pagesv1beta1.Deployment)
	if !ok {
		return nil, errors.New(errNotPagesDeployment)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	api, err := clients.NewClient(*cfg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Cloudflare client")
	}

	return &pagesDeploymentExternal{
		service: c.newServiceFn(api),
		kube:    c.kube,
	}, nil
}

type pagesDeploymentExternal struct {
	service pages.CloudflarePagesDeploymentClient
	kube    client.Client
}

func (c *pagesDeploymentExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*pagesv1beta1.Deployment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPagesDeployment)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	deployment, err := c.service.GetPagesDeploymentInfo(ctx, rc, cr.Spec.ForProvider.ProjectName, meta.GetExternalName(cr))
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get Pages deployment")
	}

	cr.Status.AtProvider.ID = deployment.ID
	cr.Status.AtProvider.ProjectName = deployment.ProjectName
	cr.Status.AtProvider.Environment = deployment.Environment
	cr.Status.AtProvider.URL = deployment.URL
	if deployment.Aliases != nil {
		cr.Status.AtProvider.Aliases = deployment.Aliases
	}
	cr.Status.AtProvider.ShortID = deployment.ShortID
	cr.Status.AtProvider.LatestStage = &pagesv1beta1.DeploymentStage{
		Name:   deployment.LatestStage.Name,
		Status: deployment.LatestStage.Status,
	}

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesDeploymentExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*pagesv1beta1.Deployment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPagesDeployment)
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	params := cloudflare.CreatePagesDeploymentParams{
		Branch: cr.Spec.ForProvider.Branch,
	}

	deployment, err := c.service.CreatePagesDeployment(ctx, rc, params)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create Pages deployment")
	}

	meta.SetExternalName(cr, deployment.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesDeploymentExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Pages deployments are typically immutable after creation
	return managed.ExternalUpdate{}, nil
}

func (c *pagesDeploymentExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*pagesv1beta1.Deployment)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPagesDeployment)
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	params := cloudflare.DeletePagesDeploymentParams{
		ProjectName: cr.Spec.ForProvider.ProjectName,
	}

	err := c.service.DeletePagesDeployment(ctx, rc, params)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete Pages deployment")
	}

	return managed.ExternalDelete{}, nil
}

func (c *pagesDeploymentExternal) Disconnect(ctx context.Context) error {
	return nil
}
