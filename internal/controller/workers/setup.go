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
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime"
)

// Setup creates all Workers controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	// Setup Script controller - the primary Workers resource
	if err := SetupScript(mgr, l, rl); err != nil {
		return err
	}

	// Setup Route controller
	if err := SetupRoute(mgr, l, rl); err != nil {
		return err
	}

	// Setup CronTrigger controller
	if err := SetupCronTrigger(mgr, l, rl); err != nil {
		return err
	}

	// Setup Domain controller
	if err := SetupDomain(mgr, l, rl); err != nil {
		return err
	}

	// Setup KVNamespace controller
	if err := SetupKVNamespace(mgr, l, rl); err != nil {
		return err
	}

	// Setup Subdomain controller
	if err := SetupSubdomain(mgr, l, rl); err != nil {
		return err
	}

	return nil
}
