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
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	"github.com/cloudflare/cloudflare-go"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	workerscript "github.com/rossigee/provider-cloudflare/internal/clients/workers/script"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	errNotScript = "managed resource is not a Worker Script custom resource"

	errClientConfig = "error getting client config"

	errScriptLookup      = "cannot lookup script"
	errScriptObservation = "cannot observe script"
	errScriptCreation    = "cannot create script"
	errScriptUpdate      = "cannot update script"
	errScriptDeletion    = "cannot delete script"

	maxConcurrency = 5
)

// SetupScript adds a controller that reconciles Worker Script managed resources.
func SetupScript(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.ScriptKind)

	o := controller.Options{
		RateLimiter:             nil, // Use default rate limiter
		MaxConcurrentReconciles: maxConcurrency,
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.ScriptGroupVersionKind),
		managed.WithExternalConnector(&scriptConnector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(client clients.ClientInterface) *workerscript.ScriptClient {
				return workerscript.NewClient(client)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))),
		managed.WithPollInterval(5*time.Minute),
		// Do not initialize external-name field.
		managed.WithInitializers(),
		managed.WithManagementPolicies(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Script{}).
		Complete(r)
}

// A scriptConnector is expected to produce an ExternalClient when its Connect method
// is called.
type scriptConnector struct {
	kube                  client.Client
	newCloudflareClientFn func(client clients.ClientInterface) *workerscript.ScriptClient
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *scriptConnector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Script)
	if !ok {
		return nil, errors.New(errNotScript)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	// Create cloudflare API client
	api, err := clients.NewClient(*config, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare client")
	}

	// Wrap with adapter to implement ClientInterface
	adapter := clients.NewCloudflareAPIAdapter(api)
	client := c.newCloudflareClientFn(adapter)
	return &scriptExternal{client: client, kube: c.kube}, nil
}

// An scriptExternal observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type scriptExternal struct {
	client *workerscript.ScriptClient
	kube   client.Client
}

func (e *scriptExternal) Observe(ctx context.Context,
	mg resource.Managed) (managed.ExternalObservation, error) {

	cr, ok := mg.(*v1beta1.Script)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotScript)
	}

	// Script does not exist if we dont have a name stored in external-name
	scriptName := meta.GetExternalName(cr)
	if scriptName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	scriptObs, err := e.client.Get(ctx, scriptName)
	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(isScriptNotFound, err), errScriptLookup)
	}

	cr.Status.AtProvider = *scriptObs
	cr.Status.SetConditions(xpv1.Available())

	upToDate, err := e.client.IsUpToDate(ctx, cr.Spec.ForProvider, *scriptObs)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: true}, err
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *scriptExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Script)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotScript)
	}

	scriptObs, err := e.client.Create(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errScriptCreation)
	}

	cr.Status.AtProvider = *scriptObs
	meta.SetExternalName(cr, cr.Spec.ForProvider.ScriptName)

	// Sync secret_text bindings. Cloudflare never returns a secret's value,
	// so the values must be pushed on every Create.
	if err := e.syncSecrets(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot sync script secrets")
	}

	return managed.ExternalCreation{}, nil
}

func (e *scriptExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Script)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotScript)
	}

	scriptName := meta.GetExternalName(cr)
	// Update should never be called on a nonexistent resource
	if scriptName == "" {
		return managed.ExternalUpdate{}, errors.New(errScriptUpdate)
	}

	_, err := e.client.Update(ctx, cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errScriptUpdate)
	}

	// Sync secret_text bindings. Cloudflare never returns a secret's value,
	// so the values must be pushed on every Update.
	if err := e.syncSecrets(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "cannot sync script secrets")
	}

	return managed.ExternalUpdate{}, nil
}

func (e *scriptExternal) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Script)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotScript)
	}

	scriptName := meta.GetExternalName(cr)

	// Delete should never be called on a nonexistent resource
	if scriptName == "" {
		return managed.ExternalDelete{}, errors.New(errScriptDeletion)
	}

	return managed.ExternalDelete{}, errors.Wrap(
		e.client.Delete(ctx, scriptName),
		errScriptDeletion)
}


// syncSecrets pushes the value of every secret_text binding to Cloudflare.
// Cloudflare's script-secrets endpoint is write-only, so a value that was
// never uploaded (or that changed in the referenced Kubernetes secret) must
// be pushed again on every reconcile.
func (e *scriptExternal) syncSecrets(ctx context.Context, cr *v1beta1.Script) error {
	accountID := e.client.GetAccountID()
	if accountID == "" {
		// The account ID is resolved when the script is created. Pushing
		// secrets without it would target an empty container identifier.
		return errors.New("cloudflare account ID not resolved; cannot sync script secrets")
	}
	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.AccountRouteLevel,
		Identifier: accountID,
	}

	for i := range cr.Spec.ForProvider.Bindings {
		b := &cr.Spec.ForProvider.Bindings[i]
		if b.Type != "secret_text" || b.SecretRef == nil {
			continue
		}
		sel := b.SecretRef
		ns := cr.GetNamespace()
		if sel.Namespace != nil && *sel.Namespace != "" {
			ns = *sel.Namespace
		}
		sec := &corev1.Secret{}
		if err := e.kube.Get(ctx, types.NamespacedName{Name: sel.Name, Namespace: ns}, sec); err != nil {
			return errors.Wrap(err, "cannot read secret for secret_text binding")
		}
		value, ok := sec.Data[sel.Key]
		if !ok {
			// fall back to the only key when none is specified
			if sel.Key == "" && len(sec.Data) == 1 {
				for _, v := range sec.Data {
					value = v
					ok = true
				}
			}
		}
		if !ok {
			return errors.New("secret key not found for secret_text binding")
		}
		_ = e.client.GetAccountID()
		params := cloudflare.SetWorkersSecretParams{
			ScriptName: cr.Spec.ForProvider.ScriptName,
			Secret: &cloudflare.WorkersPutSecretRequest{
				Name: b.Name,
				Text: string(value),
				Type: cloudflare.WorkerBindingType("secret_text"),
			},
		}
		if _, err := e.client.SetWorkersSecret(ctx, rc, params); err != nil {
			return errors.Wrap(err, "cannot set worker secret")
		}
	}
	return nil
}

func (e *scriptExternal) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// isScriptNotFound checks if an error indicates that a script was not found.
func isScriptNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "does not exist")
}
