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

package main

import (
	"context"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/rossigee/provider-cloudflare/apis"
	"github.com/rossigee/provider-cloudflare/internal/controller"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
	"github.com/rossigee/provider-cloudflare/internal/version"
	"gopkg.in/alecthomas/kingpin.v2"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	"os"
	"path/filepath"
	"runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	var (
		app            = kingpin.New(filepath.Base(os.Args[0]), "CloudFlare DNS and Zone support for Crossplane.").DefaultEnvars()
		debug          = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncPeriod     = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration()
		leaderElection = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-cloudflare"))

	shutdownTracing := tracing.Init("provider-cloudflare")
	defer shutdownTracing(context.Background())

	// Always set the controller-runtime logger to prevent logging errors
	ctrl.SetLogger(zl)

	log.Info("Provider starting up",
		"provider", "provider-cloudflare",
		"version", version.Version,
		"go-version", runtime.Version(),
		"platform", runtime.GOOS+"/"+runtime.GOARCH,
		"sync-period", syncPeriod.String(),
		"leader-election", *leaderElection,
		"leader-election-id", "crossplane-leader-election-provider-cloudflare",
		"debug-mode", *debug)

	s := apimachineryruntime.NewScheme()
	kingpin.FatalIfError(scheme.AddToScheme(s), "Cannot add k8s types to scheme")
	kingpin.FatalIfError(apis.AddToScheme(s), "Cannot add CloudFlare APIs to scheme")

	cfg, err := ctrl.GetConfig()
	kingpin.FatalIfError(err, "Cannot get API server rest config")

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             s,
		LeaderElection:   *leaderElection,
		LeaderElectionID: "crossplane-leader-election-provider-cloudflare",
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	rl := workqueue.DefaultTypedControllerRateLimiter[any]()
	kingpin.FatalIfError(apis.VerifySchemeRegistration(), "Scheme verification failed")
	log.Info("CloudFlare APIs added to scheme successfully")
	kingpin.FatalIfError(controller.SetupMinimal(mgr, log, rl), "Cannot setup minimal CloudFlare controllers")

	kingpin.FatalIfError(mgr.AddHealthzCheck("healthz", healthz.Ping), "Cannot add health check")
	kingpin.FatalIfError(mgr.AddReadyzCheck("readyz", healthz.Ping), "Cannot add ready check")

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
