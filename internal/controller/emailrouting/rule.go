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

package emailrouting

import (
	"context"
	"fmt"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-cloudflare/apis/emailrouting/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	emailroutingruleclient "github.com/rossigee/provider-cloudflare/internal/clients/emailrouting/rule"
)

const (
	errNotRule       = "managed resource is not a Rule custom resource"
	errClientConfig  = "error getting client config"
	errNewClient     = "cannot create new Service"
	errCreateRule    = "cannot create email routing rule"
	errUpdateRule    = "cannot update email routing rule"
	errDeleteRule    = "cannot delete email routing rule"
	errGetRule       = "cannot get email routing rule"
)

// SetupRule adds a controller that reconciles Rule managed resources.
func SetupRule(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.RuleKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RuleGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			newServiceFn: emailroutingruleclient.NewClientFromAPI,
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha1.Rule{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	newServiceFn func(api *cloudflare.API) *emailroutingruleclient.RuleClient
}

// Connect typically produces an ExternalClient by:
// 1. Getting the managed resource's credentials.
// 2. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return nil, errors.New(errNotRule)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	// Create Cloudflare API client using the configuration
	api, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, errNewClient)
	}

	return &external{service: c.newServiceFn(api)}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service *emailroutingruleclient.RuleClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRule)
	}

	// These fmt statements should be removed in the real implementation.
	fmt.Printf("Observing: %+v", cr)

	ruleTag := meta.GetExternalName(cr)
	if ruleTag == "" {
		// Rule doesn't exist yet
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	obs, err := c.service.Get(ctx, cr.Spec.ForProvider.ZoneID, ruleTag)
	if err != nil {
		if emailroutingruleclient.IsRuleNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetRule)
	}

	cr.Status.AtProvider = *obs

	upToDate, err := c.service.IsUpToDate(ctx, cr.Spec.ForProvider, *obs)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRule)
	}

	fmt.Printf("Creating: %+v", cr)

	obs, err := c.service.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateRule)
	}

	cr.Status.AtProvider = *obs
	meta.SetExternalName(cr, obs.Tag)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRule)
	}

	fmt.Printf("Updating: %+v", cr)

	ruleTag := meta.GetExternalName(cr)
	obs, err := c.service.Update(ctx, ruleTag, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateRule)
	}

	cr.Status.AtProvider = *obs

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return errors.New(errNotRule)
	}

	fmt.Printf("Deleting: %+v", cr)

	ruleTag := meta.GetExternalName(cr)
	if ruleTag == "" {
		// Rule doesn't exist, nothing to delete
		return nil
	}

	return errors.Wrap(c.service.Delete(ctx, cr.Spec.ForProvider.ZoneID, ruleTag), errDeleteRule)
}