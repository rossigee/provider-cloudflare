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

package controller

import (
	"context"
	"os"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/rossigee/provider-cloudflare/internal/controller/access"
	"github.com/rossigee/provider-cloudflare/internal/controller/cache"
	"github.com/rossigee/provider-cloudflare/internal/controller/device"
	record "github.com/rossigee/provider-cloudflare/internal/controller/dns"
	"github.com/rossigee/provider-cloudflare/internal/controller/emailrouting"
	"github.com/rossigee/provider-cloudflare/internal/controller/loadbalancing"
	"github.com/rossigee/provider-cloudflare/internal/controller/originssl"
	"github.com/rossigee/provider-cloudflare/internal/controller/providerconfig"
	"github.com/rossigee/provider-cloudflare/internal/controller/r2"
	"github.com/rossigee/provider-cloudflare/internal/controller/rulesets"
	"github.com/rossigee/provider-cloudflare/internal/controller/security"
	application "github.com/rossigee/provider-cloudflare/internal/controller/spectrum"
	"github.com/rossigee/provider-cloudflare/internal/controller/ssl"
	"github.com/rossigee/provider-cloudflare/internal/controller/sslsaas"
	"github.com/rossigee/provider-cloudflare/internal/controller/transform"
	"github.com/rossigee/provider-cloudflare/internal/controller/tunnel"
	"github.com/rossigee/provider-cloudflare/internal/controller/workers"
	"github.com/rossigee/provider-cloudflare/internal/controller/zone"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Setup creates all CloudFlare controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, wl workqueue.TypedRateLimiter[any]) error {
	if err := setupRBAC(mgr.GetClient(), l); err != nil {
		l.Info("RBAC setup warning (may be transient)", "error", err)
	}
	for _, setup := range []func(ctrl.Manager, logging.Logger, workqueue.TypedRateLimiter[any]) error{
		// config.Setup, // Temporarily disabled for v2 compatibility debugging
		zone.Setup,
		record.Setup,
		application.Setup,
		workers.Setup, // Workers client implementation now complete
		ssl.Setup,
		sslsaas.Setup,
		transform.Setup,
		rulesets.Setup,
		security.Setup,
		loadbalancing.Setup,
		originssl.Setup,
		cache.Setup,
		r2.Setup,
		emailrouting.Setup,
		access.Setup,
		tunnel.Setup,
		device.Setup,
	} {
		if err := setup(mgr, l, wl); err != nil {
			return err
		}
	}
	// providerconfig.Setup has a different signature - call it separately
	if err := providerconfig.Setup(mgr); err != nil {
		return err
	}
	return nil
}

// SetupMinimal creates minimal controllers with only config, zone, and dns record support.
func SetupMinimal(mgr ctrl.Manager, l logging.Logger, wl workqueue.TypedRateLimiter[any]) error {
	if err := setupRBAC(mgr.GetClient(), l); err != nil {
		l.Info("RBAC setup warning (may be transient)", "error", err)
	}
	for _, setup := range []func(ctrl.Manager, logging.Logger, workqueue.TypedRateLimiter[any]) error{
		// config.Setup, // Temporarily disabled for v2 compatibility debugging
		zone.Setup,
		record.Setup,
		application.Setup,
		workers.Setup, // Workers client implementation now complete
		ssl.Setup,
		sslsaas.Setup,
		transform.Setup,
		rulesets.Setup,
		security.Setup,
		loadbalancing.Setup,
		originssl.Setup,
		cache.Setup,
		r2.Setup,
		emailrouting.Setup,
		access.Setup,
		tunnel.Setup,
		device.Setup,
	} {
		if err := setup(mgr, l, wl); err != nil {
			return err
		}
	}
	// providerconfig.Setup has a different signature - call it separately
	if err := providerconfig.Setup(mgr); err != nil {
		return err
	}
	return nil
}

func setupRBAC(c client.Client, l logging.Logger) error {
	ctx := context.Background()

	rules := []rbacv1.PolicyRule{
		{APIGroups: []string{"access.cloudflare.m.crossplane.io"}, Resources: []string{"accessapplications", "accessapplications/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"spectrum.cloudflare.m.crossplane.io"}, Resources: []string{"applications", "applications/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"security.cloudflare.m.crossplane.io"}, Resources: []string{"botmanagements", "botmanagements/status", "ratelimits", "ratelimits/status", "turnstiles", "turnstiles/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"r2.cloudflare.m.crossplane.io"}, Resources: []string{"buckets", "buckets/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"cache.cloudflare.m.crossplane.io"}, Resources: []string{"cacherules", "cacherules/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"ssl.cloudflare.m.crossplane.io"}, Resources: []string{"certificatepacks", "certificatepacks/status", "totaltlses", "totaltlses/status", "universalssls", "universalssls/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"originssl.cloudflare.m.crossplane.io"}, Resources: []string{"certificates", "certificates/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{APIGroups: []string{"cloudflare.crossplane.io"}, Resources: []string{"providerconfigs", "providerconfigs/status", "providerconfigusages", "providerconfigusages/status"}, Verbs: []string{"get", "list", "watch", "update", "patch", "create"}},
		{
			APIGroups: []string{"access.cloudflare.m.crossplane.io", "spectrum.cloudflare.m.crossplane.io", "security.cloudflare.m.crossplane.io", "r2.cloudflare.m.crossplane.io", "cache.cloudflare.m.crossplane.io", "ssl.cloudflare.m.crossplane.io", "originssl.cloudflare.m.crossplane.io", "cloudflare.crossplane.io"},
			Resources: []string{"*/finalizers"},
			Verbs:     []string{"update"},
		},
		{APIGroups: []string{"", "coordination.k8s.io"}, Resources: []string{"secrets", "configmaps", "events", "leases"}, Verbs: []string{"*"}},
	}

	system := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{Name: "crossplane:provider:provider-cloudflare:system", Labels: map[string]string{"rbac.crossplane.io/system": "provider-cloudflare"}},
		Rules:      rules,
	}
	if err := c.Create(ctx, system); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	if err := c.Update(ctx, system); err != nil {
		l.Info("system role update", "err", err)
	}

	binding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "crossplane:provider:provider-cloudflare:system"},
		RoleRef:    rbacv1.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: "crossplane:provider:provider-cloudflare:system"},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: os.Getenv("REVISION_NAME"), Namespace: "crossplane-system"}},
	}
	if err := c.Create(ctx, binding); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	if err := c.Update(ctx, binding); err != nil {
		l.Info("system binding update", "err", err)
	}

	edit := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "crossplane:provider:provider-cloudflare:aggregate-to-edit",
			Labels: map[string]string{"rbac.crossplane.io/aggregate-to-edit": "true", "rbac.crossplane.io/aggregate-to-admin": "true", "rbac.crossplane.io/aggregate-to-crossplane": "true", "rbac.crossplane.io/system": "provider-cloudflare"},
		},
		Rules: withVerbs(rules, []string{"*"}),
	}
	if err := c.Create(ctx, edit); err != nil && !errors.IsAlreadyExists(err) {
		l.Info("aggregate-to-edit create warning (non-fatal)", "err", err)
	}
	_ = c.Update(ctx, edit)

	view := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "crossplane:provider:provider-cloudflare:aggregate-to-view",
			Labels: map[string]string{"rbac.crossplane.io/aggregate-to-view": "true", "rbac.crossplane.io/system": "provider-cloudflare"},
		},
		Rules: withVerbs(rules, []string{"get", "list", "watch"}),
	}
	if err := c.Create(ctx, view); err != nil && !errors.IsAlreadyExists(err) {
		l.Info("aggregate-to-view create warning (non-fatal)", "err", err)
	}
	_ = c.Update(ctx, view)

	l.Info("provider self-managed RBAC roles ensured")
	return nil
}

func withVerbs(r []rbacv1.PolicyRule, verbs []string) []rbacv1.PolicyRule {
	out := make([]rbacv1.PolicyRule, len(r))
	for i := range r {
		out[i] = r[i]
		out[i].Verbs = verbs
	}
	return out
}
