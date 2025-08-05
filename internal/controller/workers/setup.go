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
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

// Setup creates all Workers controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	// Setup Route controller (existing pattern)
	if err := SetupRoute(mgr, l, rl); err != nil {
		return err
	}

	// Setup new Workers controllers with proper account management
	// Enable CronTrigger first, then add others as they're fixed
	if err := SetupCronTrigger(mgr, l, rl); err != nil {
		return err
	}
	
	// Enable Script and KV Namespace controllers - compilation issues resolved
	if err := SetupScript(mgr, l, rl); err != nil {
		return err
	}
	if err := SetupKVNamespace(mgr, l, rl); err != nil {
		return err
	}

	// Enable Domain and Subdomain controllers
	if err := SetupDomain(mgr, l, rl); err != nil {
		return err
	}
	if err := SetupSubdomain(mgr, l, rl); err != nil {
		return err
	}

	return nil
}