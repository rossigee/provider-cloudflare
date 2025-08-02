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

package rule

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
	errNotRule = "managed resource is not a Rule custom resource"

	errClientConfig = "error getting client config"

	errRuleLookup   = "cannot lookup firewall rule"
	errRuleCreation = "cannot create firewall rule"
	errRuleUpdate   = "cannot update firewall rule"
	errRuleDeletion = "cannot delete firewall rule"
	errNoZone       = "no zone found"
	errNoFilter     = "no filter found"
	errRuleNotFound = "Rule not found"

	maxConcurrency = 5
)

// Client is a Cloudflare API client that implements methods for working
// with Firewall Rules.
type Client interface {
	FirewallRule(ctx context.Context, zoneID, ruleID string) (cloudflare.FirewallRule, error)
	CreateFirewallRule(ctx context.Context, zoneID string, rule cloudflare.FirewallRule) (*cloudflare.FirewallRule, error)
	UpdateFirewallRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.FirewallRule) error
	DeleteFirewallRule(ctx context.Context, zoneID, ruleID string) error
}

type clientImpl struct {
	cf *cloudflare.API
}

// NewClient returns a new Cloudflare API client for working with Firewall Rules.
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	cf, err := clients.NewClient(cfg, hc)
	if err != nil {
		return nil, err
	}

	return &clientImpl{cf: cf}, nil
}

// FirewallRule retrieves a Firewall Rule
func (c *clientImpl) FirewallRule(ctx context.Context, zoneID, ruleID string) (cloudflare.FirewallRule, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	rule, err := c.cf.FirewallRule(ctx, rc, ruleID)
	if err != nil {
		return cloudflare.FirewallRule{}, err
	}

	if rule.ID == "" {
		return cloudflare.FirewallRule{}, errors.New(errRuleNotFound)
	}

	return rule, nil
}

// CreateFirewallRule creates a new Firewall Rule
func (c *clientImpl) CreateFirewallRule(ctx context.Context, zoneID string, rule cloudflare.FirewallRule) (*cloudflare.FirewallRule, error) {
	rc := cloudflare.ZoneIdentifier(zoneID)
	params := []cloudflare.FirewallRuleCreateParams{{
		Filter: cloudflare.Filter{ID: rule.Filter.ID},
		Action: rule.Action,
	}}
	
	rules, err := c.cf.CreateFirewallRules(ctx, rc, params)
	if err != nil {
		return nil, err
	}
	
	if len(rules) == 0 {
		return nil, errors.New("no rule created")
	}

	return &rules[0], nil
}

// UpdateFirewallRule updates an existing Firewall Rule
func (c *clientImpl) UpdateFirewallRule(ctx context.Context, zoneID, ruleID string, rule cloudflare.FirewallRule) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	params := cloudflare.FirewallRuleUpdateParams{
		ID:     ruleID,
		Filter: cloudflare.Filter{ID: rule.Filter.ID},
		Action: rule.Action,
	}
	
	_, err := c.cf.UpdateFirewallRule(ctx, rc, params)
	return err
}

// DeleteFirewallRule deletes a Firewall Rule
func (c *clientImpl) DeleteFirewallRule(ctx context.Context, zoneID, ruleID string) error {
	rc := cloudflare.ZoneIdentifier(zoneID)
	err := c.cf.DeleteFirewallRule(ctx, rc, ruleID)
	return err
}

// IsRuleNotFound returns true if the error indicates the rule was not found
func IsRuleNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == errRuleNotFound ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}

// GenerateObservation creates observation data from a FirewallRule
func GenerateObservation(rule cloudflare.FirewallRule) v1alpha1.RuleObservation {
	return v1alpha1.RuleObservation{}
}

// LateInitialize initializes RuleParameters based on the remote resource
func LateInitialize(spec *v1alpha1.RuleParameters, rule cloudflare.FirewallRule) bool {
	if spec == nil {
		return false
	}

	li := false
	if spec.Paused == nil && rule.Paused {
		spec.Paused = &rule.Paused
		li = true
	}

	return li
}

// UpToDate checks if the remote FirewallRule is up to date with the requested resource parameters
func UpToDate(spec *v1alpha1.RuleParameters, rule cloudflare.FirewallRule) bool {
	if spec == nil {
		return true
	}

	if spec.Action != rule.Action {
		return false
	}

	if spec.Filter != nil && *spec.Filter != rule.Filter.ID {
		return false
	}

	if spec.Paused != nil && *spec.Paused != rule.Paused {
		return false
	}

	return true
}

// CreateRule creates a FirewallRule from RuleParameters
func CreateRule(ctx context.Context, client Client, params *v1alpha1.RuleParameters) (*cloudflare.FirewallRule, error) {
	if params.Zone == nil {
		return nil, errors.New("zone is required")
	}

	if params.Filter == nil {
		return nil, errors.New("filter is required")
	}

	rule := cloudflare.FirewallRule{
		Filter: cloudflare.Filter{ID: *params.Filter},
		Action: params.Action,
	}

	if params.Paused != nil {
		rule.Paused = *params.Paused
	}

	result, err := client.CreateFirewallRule(ctx, *params.Zone, rule)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateRule updates an existing FirewallRule
func UpdateRule(ctx context.Context, client Client, ruleID string, params *v1alpha1.RuleParameters) error {
	if params.Zone == nil {
		return errors.New("zone is required")
	}

	if params.Filter == nil {
		return errors.New("filter is required")
	}

	rule := cloudflare.FirewallRule{
		Filter: cloudflare.Filter{ID: *params.Filter},
		Action: params.Action,
	}

	if params.Paused != nil {
		rule.Paused = *params.Paused
	}

	err := client.UpdateFirewallRule(ctx, *params.Zone, ruleID, rule)
	return err
}

// Setup adds a controller that reconciles Rule managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha1.RuleGroupKind)

	o := controller.Options{
		RateLimiter:             rl,
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RuleGroupVersionKind),
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
		For(&v1alpha1.Rule{}).
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
	_, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return nil, errors.New(errNotRule)
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
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRule)
	}

	// Rule does not exist if we dont have an ID stored in external-name
	rid := meta.GetExternalName(cr)
	if rid == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errNoZone)
	}

	r, err := e.client.FirewallRule(ctx, *cr.Spec.ForProvider.Zone, rid)

	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(IsRuleNotFound, err), errRuleLookup)
	}

	cr.Status.AtProvider = GenerateObservation(r)

	cr.Status.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: LateInitialize(&cr.Spec.ForProvider, r),
		ResourceUpToDate:        UpToDate(&cr.Spec.ForProvider, r),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRule)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{}, errors.New(errNoZone)
	}

	if cr.Spec.ForProvider.Filter == nil {
		return managed.ExternalCreation{}, errors.New(errNoFilter)
	}

	nr, err := CreateRule(ctx, e.client, &cr.Spec.ForProvider)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRuleCreation)
	}

	cr.Status.AtProvider = GenerateObservation(*nr)

	// Update the external name with the ID of the new Rule
	meta.SetExternalName(cr, nr.ID)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRule)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errNoZone), errRuleUpdate)
	}

	rid := meta.GetExternalName(cr)

	// Update should never be called on a nonexistent resource
	if rid == "" {
		return managed.ExternalUpdate{}, errors.New(errRuleUpdate)
	}

	return managed.ExternalUpdate{},
		errors.Wrap(
			UpdateRule(ctx, e.client, meta.GetExternalName(cr), &cr.Spec.ForProvider),
			errRuleUpdate,
		)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Rule)
	if !ok {
		return errors.New(errNotRule)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return errors.Wrap(errors.New(errNoZone), errRuleDeletion)
	}

	rid := meta.GetExternalName(cr)

	// Delete should never be called on a nonexistent resource
	if rid == "" {
		return errors.New(errRuleDeletion)
	}

	return errors.Wrap(
		e.client.DeleteFirewallRule(ctx, *cr.Spec.ForProvider.Zone, meta.GetExternalName(cr)),
		errRuleDeletion)
}
