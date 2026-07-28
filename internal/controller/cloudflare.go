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
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Setup creates all CloudFlare controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, wl workqueue.TypedRateLimiter[any]) error {
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
