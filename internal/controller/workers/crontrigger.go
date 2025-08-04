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

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	crontriggerclient "github.com/rossigee/provider-cloudflare/internal/clients/workers/crontrigger"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
)

const (
	errNotCronTrigger = "managed resource is not a CronTrigger custom resource"

	errCronTriggerClientConfig = "error getting cron trigger client config"

	errCronTriggerLookup   = "cannot lookup CronTrigger"
	errCronTriggerCreation = "cannot create CronTrigger"
	errCronTriggerUpdate   = "cannot update CronTrigger"
	errCronTriggerDeletion = "cannot delete CronTrigger"

	cronTriggerMaxConcurrency = 5
)

// SetupCronTrigger adds a controller that reconciles CronTrigger managed resources.
func SetupCronTrigger(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.CronTriggerGroupKind)

	o := controller.Options{
		RateLimiter:             rl,
		MaxConcurrentReconciles: cronTriggerMaxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.CronTriggerGroupVersionKind),
		managed.WithExternalConnecter(&cronTriggerConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (*cloudflare.API, error) {
				return clients.NewClient(cfg, hc)
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
		For(&v1alpha1.CronTrigger{}).
		Complete(r)
}

// A cronTriggerConnector is expected to produce an ExternalClient when its Connect method
// is called.
type cronTriggerConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (*cloudflare.API, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *cronTriggerConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.CronTrigger)
	if !ok {
		return nil, errors.New(errNotCronTrigger)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errCronTriggerClientConfig)
	}

	client, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, err
	}

	// Create the cron trigger client wrapper
	adapter := clients.NewCloudflareAPIAdapter(client)
	cronTriggerClient := crontriggerclient.NewClient(adapter)

	return &cronTriggerExternal{client: cronTriggerClient}, nil
}

// An cronTriggerExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type cronTriggerExternal struct {
	client *crontriggerclient.CronTriggerClient
}

func (c *cronTriggerExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.CronTrigger)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotCronTrigger)
	}

	// CronTrigger does not exist if we don't have an ID stored in external-name
	triggerID := meta.GetExternalName(cr)
	if triggerID == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// For cron triggers, we identify them by script name + cron expression
	scriptName := cr.Spec.ForProvider.ScriptName
	cronExpression := cr.Spec.ForProvider.Cron
	
	observation, err := c.client.Get(ctx, scriptName, cronExpression)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(crontriggerclient.IsCronTriggerNotFound, err), errCronTriggerLookup)
	}

	cr.Status.AtProvider = *observation
	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: observation.Cron == cronExpression && observation.ScriptName == scriptName,
	}, nil
}

func (c *cronTriggerExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.CronTrigger)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotCronTrigger)
	}

	cr.SetConditions(rtv1.Creating())

	observation, err := c.client.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCronTriggerCreation)
	}

	// Update the external name with a unique ID for this cron trigger
	meta.SetExternalName(cr, cr.Spec.ForProvider.ScriptName+":"+cr.Spec.ForProvider.Cron)
	cr.Status.AtProvider = *observation

	return managed.ExternalCreation{}, nil
}

func (c *cronTriggerExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.CronTrigger)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotCronTrigger)
	}

	// For updates, we need the old cron expression to find and replace
	// We'll use the current observation's cron expression as the old one
	oldCron := cr.Status.AtProvider.Cron
	if oldCron == "" {
		return managed.ExternalUpdate{}, errors.New("cannot update cron trigger without existing cron expression")
	}

	observation, err := c.client.Update(ctx, oldCron, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCronTriggerUpdate)
	}

	cr.Status.AtProvider = *observation

	return managed.ExternalUpdate{}, nil
}

func (c *cronTriggerExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.CronTrigger)
	if !ok {
		return errors.New(errNotCronTrigger)
	}

	scriptName := cr.Spec.ForProvider.ScriptName
	cronExpression := cr.Spec.ForProvider.Cron

	err := c.client.Delete(ctx, scriptName, cronExpression)
	if err != nil && !crontriggerclient.IsCronTriggerNotFound(err) {
		return errors.Wrap(err, errCronTriggerDeletion)
	}

	return nil
}