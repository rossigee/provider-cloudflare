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
	errNotPagesProject = "managed resource is not a Pages Project custom resource"
	errGetCreds        = "cannot get credentials"
)

func SetupPagesProject(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(pagesv1beta1.ProjectKind)

	o := controller.Options{
		MaxConcurrentReconciles: 5,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(pagesv1beta1.ProjectGroupVersionKind),
		managed.WithExternalConnector(&pagesProjectConnector{
			kube: mgr.GetClient(),
			newServiceFn: func(api *cloudflare.API) pages.CloudflarePagesProjectClient {
				return pages.NewClientFromAPI(api)
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
		For(&pagesv1beta1.Project{}).
		Complete(r)
}

type pagesProjectConnector struct {
	kube         client.Client
	newServiceFn func(*cloudflare.API) pages.CloudflarePagesProjectClient
}

func (c *pagesProjectConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*pagesv1beta1.Project)
	if !ok {
		return nil, errors.New(errNotPagesProject)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	api, err := clients.NewClient(*cfg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create new Cloudflare client")
	}

	return &pagesProjectExternal{
		service: c.newServiceFn(api),
		kube:    c.kube,
	}, nil
}

type pagesProjectExternal struct {
	service pages.CloudflarePagesProjectClient
	kube    client.Client
}

func (c *pagesProjectExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*pagesv1beta1.Project)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPagesProject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	project, err := c.service.GetPagesProject(ctx, rc, meta.GetExternalName(cr))
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get Pages project")
	}

	cr.Status.AtProvider.ID = project.ID
	cr.Status.AtProvider.Name = project.Name
	cr.Status.AtProvider.Subdomain = project.SubDomain
	cr.Status.AtProvider.Domains = project.Domains
	cr.Status.AtProvider.ProductionBranch = project.ProductionBranch
	cr.Status.AtProvider.BuildCommand = project.BuildConfig.BuildCommand
	cr.Status.AtProvider.DestinationDir = project.BuildConfig.DestinationDir
	cr.Status.AtProvider.RootDir = project.BuildConfig.RootDir
	cr.Status.AtProvider.CreatedOn = project.CreatedOn.String()
	cr.Status.AtProvider.ModifiedOn = project.CreatedOn.String()

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesProjectExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*pagesv1beta1.Project)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPagesProject)
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	params := cloudflare.CreatePagesProjectParams{
		Name:             cr.Spec.ForProvider.Name,
		ProductionBranch: cr.Spec.ForProvider.ProductionBranch,
	}

	params.BuildConfig = cloudflare.PagesProjectBuildConfig{
		BuildCommand:   cr.Spec.ForProvider.BuildCommand,
		DestinationDir: cr.Spec.ForProvider.DestinationDir,
		RootDir:        cr.Spec.ForProvider.RootDir,
	}

	if cr.Spec.ForProvider.GitRepository != nil {
		params.Source = &cloudflare.PagesProjectSource{
			Config: &cloudflare.PagesProjectSourceConfig{
				Owner:            cr.Spec.ForProvider.GitRepository.URL,
				RepoName:         cr.Spec.ForProvider.GitRepository.Branch,
				ProductionBranch: cr.Spec.ForProvider.GitRepository.ConfigDir,
			},
		}
	}

	project, err := c.service.CreatePagesProject(ctx, rc, params)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create Pages project")
	}

	meta.SetExternalName(cr, project.ID)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *pagesProjectExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*pagesv1beta1.Project)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPagesProject)
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	params := cloudflare.UpdatePagesProjectParams{
		Name:             cr.Spec.ForProvider.Name,
		ProductionBranch: cr.Spec.ForProvider.ProductionBranch,
	}

	params.BuildConfig = cloudflare.PagesProjectBuildConfig{
		BuildCommand:   cr.Spec.ForProvider.BuildCommand,
		DestinationDir: cr.Spec.ForProvider.DestinationDir,
		RootDir:        cr.Spec.ForProvider.RootDir,
	}

	_, err := c.service.UpdatePagesProject(ctx, rc, params)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot update Pages project")
	}

	return managed.ExternalUpdate{}, nil
}

func (c *pagesProjectExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*pagesv1beta1.Project)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotPagesProject)
	}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: cr.Spec.ForProvider.AccountID,
	}

	err := c.service.DeletePagesProject(ctx, rc, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete Pages project")
	}

	return managed.ExternalDelete{}, nil
}

func (c *pagesProjectExternal) Disconnect(ctx context.Context) error {
	return nil
}
