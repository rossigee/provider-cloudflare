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
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/statemetrics"
	"github.com/rossigee/provider-cloudflare/apis"
	accessv1beta1 "github.com/rossigee/provider-cloudflare/apis/access/v1beta1"
	cachev1beta1 "github.com/rossigee/provider-cloudflare/apis/cache/v1beta1"
	devicev1beta1 "github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	dnsv1beta1 "github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	emailroutingv1beta1 "github.com/rossigee/provider-cloudflare/apis/emailrouting/v1beta1"
	firewallv1beta1 "github.com/rossigee/provider-cloudflare/apis/firewall/v1beta1"
	loadbalancingv1beta1 "github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
	logpushv1beta1 "github.com/rossigee/provider-cloudflare/apis/logpush/v1beta1"
	originsslv1beta1 "github.com/rossigee/provider-cloudflare/apis/originssl/v1beta1"
	r2v1beta1 "github.com/rossigee/provider-cloudflare/apis/r2/v1beta1"
	rulesetsv1beta1 "github.com/rossigee/provider-cloudflare/apis/rulesets/v1beta1"
	securityv1beta1 "github.com/rossigee/provider-cloudflare/apis/security/v1beta1"
	spectrumv1beta1 "github.com/rossigee/provider-cloudflare/apis/spectrum/v1beta1"
	sslv1beta1 "github.com/rossigee/provider-cloudflare/apis/ssl/v1beta1"
	sslsaasv1beta1 "github.com/rossigee/provider-cloudflare/apis/sslsaas/v1beta1"
	transformv1beta1 "github.com/rossigee/provider-cloudflare/apis/transform/v1beta1"
	tunnelv1beta1 "github.com/rossigee/provider-cloudflare/apis/tunnel/v1beta1"
	workersv1beta1 "github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	zonev1beta1 "github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/controller"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
	"github.com/rossigee/provider-cloudflare/internal/version"
	"gopkg.in/alecthomas/kingpin.v2"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	metricserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
	var (
		app                     = kingpin.New(filepath.Base(os.Args[0]), "CloudFlare DNS and Zone support for Crossplane.").DefaultEnvars()
		debug                   = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncPeriod              = app.Flag("sync", "Controller manager sync period such as 300ms, 1.5h, or 2h45m").Short('s').Default("1h").Duration()
		leaderElection          = app.Flag("leader-election", "Use leader election for the controller manager.").Short('l').Default("false").OverrideDefaultFromEnvar("LEADER_ELECTION").Bool()
		pollStateMetricInterval = app.Flag("poll-state-metric", "State metric recording interval").Default("5s").Duration()
		metricsBindAddress      = app.Flag("metrics-bind-address", "The address the metrics endpoint binds to.").Default(":8080").String()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	log := logging.NewLogrLogger(zl.WithName("provider-cloudflare"))

	signalCtx := ctrl.SetupSignalHandler()
	shutdownTracing := tracing.Init("provider-cloudflare")
	defer shutdownTracing(signalCtx)

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

	sync := *syncPeriod
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:           s,
		Logger:           zl,
		LeaderElection:   *leaderElection,
		LeaderElectionID: "crossplane-leader-election-provider-cloudflare",
		Cache: cache.Options{
			SyncPeriod: &sync,
		},
		Metrics: metricserver.Options{
			BindAddress: *metricsBindAddress,
		},
		Controller: config.Controller{
			// 30 controllers/watchers x initial List+Watch startup can exceed the
			// controller-runtime default of 2m when the apiserver is slow or many
			// CRDs are installed. Raise the ceiling so the manager does not fatal-exit
			// before caches finish syncing.
			CacheSyncTimeout: 10 * time.Minute,
		},
	})
	kingpin.FatalIfError(err, "Cannot create controller manager")

	mrStateMetrics := statemetrics.NewMRStateMetrics()
	metrics.Registry.MustRegister(mrStateMetrics)

	rl := workqueue.DefaultTypedControllerRateLimiter[any]()
	kingpin.FatalIfError(apis.VerifySchemeRegistration(), "Scheme verification failed")
	log.Info("CloudFlare APIs added to scheme successfully")
	kingpin.FatalIfError(controller.SetupMinimal(mgr, log, rl), "Cannot setup minimal CloudFlare controllers")

	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &zonev1beta1.ZoneList{}, *pollStateMetricInterval)), "Cannot register state metrics for Zone")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &dnsv1beta1.RecordList{}, *pollStateMetricInterval)), "Cannot register state metrics for Record")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &spectrumv1beta1.ApplicationList{}, *pollStateMetricInterval)), "Cannot register state metrics for Application")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.ScriptList{}, *pollStateMetricInterval)), "Cannot register state metrics for Script")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.CronTriggerList{}, *pollStateMetricInterval)), "Cannot register state metrics for CronTrigger")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.DomainList{}, *pollStateMetricInterval)), "Cannot register state metrics for WorkerDomain")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.KVNamespaceList{}, *pollStateMetricInterval)), "Cannot register state metrics for KVNamespace")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.RouteList{}, *pollStateMetricInterval)), "Cannot register state metrics for WorkerRoute")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &workersv1beta1.SubdomainList{}, *pollStateMetricInterval)), "Cannot register state metrics for Subdomain")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &sslv1beta1.CertificatePackList{}, *pollStateMetricInterval)), "Cannot register state metrics for CertificatePack")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &sslv1beta1.TotalTLSList{}, *pollStateMetricInterval)), "Cannot register state metrics for TotalTLS")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &sslv1beta1.UniversalSSLList{}, *pollStateMetricInterval)), "Cannot register state metrics for UniversalSSL")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &originsslv1beta1.CertificateList{}, *pollStateMetricInterval)), "Cannot register state metrics for OriginCertificate")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &sslsaasv1beta1.CustomHostnameList{}, *pollStateMetricInterval)), "Cannot register state metrics for CustomHostname")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &sslsaasv1beta1.FallbackOriginList{}, *pollStateMetricInterval)), "Cannot register state metrics for FallbackOrigin")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &transformv1beta1.RuleList{}, *pollStateMetricInterval)), "Cannot register state metrics for TransformRule")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &rulesetsv1beta1.RulesetList{}, *pollStateMetricInterval)), "Cannot register state metrics for Ruleset")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &securityv1beta1.RateLimitList{}, *pollStateMetricInterval)), "Cannot register state metrics for RateLimit")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &securityv1beta1.BotManagementList{}, *pollStateMetricInterval)), "Cannot register state metrics for BotManagement")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &securityv1beta1.TurnstileList{}, *pollStateMetricInterval)), "Cannot register state metrics for Turnstile")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &loadbalancingv1beta1.LoadBalancerList{}, *pollStateMetricInterval)), "Cannot register state metrics for LoadBalancer")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &loadbalancingv1beta1.LoadBalancerPoolList{}, *pollStateMetricInterval)), "Cannot register state metrics for LoadBalancerPool")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &loadbalancingv1beta1.LoadBalancerMonitorList{}, *pollStateMetricInterval)), "Cannot register state metrics for LoadBalancerMonitor")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &cachev1beta1.CacheRuleList{}, *pollStateMetricInterval)), "Cannot register state metrics for CacheRule")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &r2v1beta1.BucketList{}, *pollStateMetricInterval)), "Cannot register state metrics for R2Bucket")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &emailroutingv1beta1.RuleList{}, *pollStateMetricInterval)), "Cannot register state metrics for EmailRoutingRule")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &accessv1beta1.AccessApplicationList{}, *pollStateMetricInterval)), "Cannot register state metrics for AccessApplication")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &tunnelv1beta1.TunnelList{}, *pollStateMetricInterval)), "Cannot register state metrics for Tunnel")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &devicev1beta1.DevicePostureRuleList{}, *pollStateMetricInterval)), "Cannot register state metrics for DevicePostureRule")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &firewallv1beta1.RuleList{}, *pollStateMetricInterval)), "Cannot register state metrics for FirewallRule")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &firewallv1beta1.FilterList{}, *pollStateMetricInterval)), "Cannot register state metrics for Filter")
	kingpin.FatalIfError(mgr.Add(statemetrics.NewMRStateRecorder(mgr.GetClient(), log, mrStateMetrics, &logpushv1beta1.JobList{}, *pollStateMetricInterval)), "Cannot register state metrics for LogpushJob")

	kingpin.FatalIfError(mgr.AddHealthzCheck("healthz", healthz.Ping), "Cannot add health check")
	kingpin.FatalIfError(mgr.AddReadyzCheck("readyz", healthz.Ping), "Cannot add ready check")

	kingpin.FatalIfError(mgr.Start(signalCtx), "Cannot start controller manager")
}
