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

package clients

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
)

// CloudflareClient represents the interface for Cloudflare API operations
type CloudflareClient interface {
	// DNS Record operations
	CreateDNSRecord(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error)
	UpdateDNSRecord(ctx context.Context, zoneID, recordID string, rr cloudflare.DNSRecord) error
	DNSRecord(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error
	DNSRecords(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error)
	ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error)
	GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error)
	ExportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ExportDNSRecordsParams) (string, error)
	ImportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ImportDNSRecordsParams) error

	// Zone operations
	CreateZone(ctx context.Context, name string, jumpstart bool, zoneType string, account cloudflare.Account) (cloudflare.Zone, error)
	ZoneDetails(ctx context.Context, zoneID string) (cloudflare.Zone, error)
	DeleteZone(ctx context.Context, zoneID string) (cloudflare.ZoneID, error)
	EditZone(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error)
	ZoneActivationCheck(ctx context.Context, zoneID string) (cloudflare.Response, error)
	ListZones(ctx context.Context, z ...string) ([]cloudflare.Zone, error)
	ListZonesContext(ctx context.Context, opts ...cloudflare.ReqOption) (cloudflare.ZonesResponse, error)
	ZoneSetPaused(ctx context.Context, zoneID string, paused bool) (cloudflare.Zone, error)
	ZoneSetType(ctx context.Context, zoneID string, zoneType string) (cloudflare.Zone, error)
	ZoneSetVanityNS(ctx context.Context, zoneID string, ns []string) (cloudflare.Zone, error)
	ZoneSetPlan(ctx context.Context, zoneID string, planType string) error
	ZoneUpdatePlan(ctx context.Context, zoneID string, planType string) error
	PurgeEverything(ctx context.Context, zoneID string) (cloudflare.PurgeCacheResponse, error)
	PurgeCache(ctx context.Context, zoneID string, pcr cloudflare.PurgeCacheRequest) (cloudflare.PurgeCacheResponse, error)
	PurgeCacheContext(ctx context.Context, zoneID string, pcr cloudflare.PurgeCacheRequest) (cloudflare.PurgeCacheResponse, error)
	AvailableZoneRatePlans(ctx context.Context, zoneID string) ([]cloudflare.ZoneRatePlan, error)
	AvailableZonePlans(ctx context.Context, zoneID string) ([]cloudflare.ZonePlan, error)
	ZoneAnalyticsDashboard(ctx context.Context, zoneID string, options cloudflare.ZoneAnalyticsOptions) (cloudflare.ZoneAnalyticsData, error)
	ZoneAnalyticsByColocation(ctx context.Context, zoneID string, options cloudflare.ZoneAnalyticsOptions) ([]cloudflare.ZoneAnalyticsColocation, error)
	ZoneSettings(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error)
	UpdateZoneSettings(ctx context.Context, zoneID string, settings []cloudflare.ZoneSetting) (*cloudflare.ZoneSettingResponse, error)
	ZoneSSLSettings(ctx context.Context, zoneID string) (cloudflare.ZoneSSLSetting, error)
	UpdateZoneSSLSettings(ctx context.Context, zoneID string, sslValue string) (cloudflare.ZoneSSLSetting, error)
	FallbackOrigin(ctx context.Context, zoneID string) (cloudflare.FallbackOrigin, error)
	UpdateFallbackOrigin(ctx context.Context, zoneID string, fbo cloudflare.FallbackOrigin) (*cloudflare.FallbackOriginResponse, error)
	GetZoneSetting(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetZoneSettingParams) (cloudflare.ZoneSetting, error)
	UpdateZoneSetting(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateZoneSettingParams) (cloudflare.ZoneSetting, error)
	ZoneExport(ctx context.Context, zoneID string) (string, error)
	ZoneDNSSECSetting(ctx context.Context, zoneID string) (cloudflare.ZoneDNSSEC, error)
	DeleteZoneDNSSEC(ctx context.Context, zoneID string) (string, error)
	UpdateZoneDNSSEC(ctx context.Context, zoneID string, options cloudflare.ZoneDNSSECUpdateOptions) (cloudflare.ZoneDNSSEC, error)

	// Origin CA Certificate operations
	CreateOriginCACertificate(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error)
	GetOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error)
	ListOriginCACertificates(ctx context.Context, params cloudflare.ListOriginCertificatesParams) ([]cloudflare.OriginCACertificate, error)
	RevokeOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error)

	// Rate Limiting operations
	CreateRateLimit(ctx context.Context, zoneID string, rateLimit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	RateLimit(ctx context.Context, zoneID, rateLimitID string) (cloudflare.RateLimit, error)
	UpdateRateLimit(ctx context.Context, zoneID, rateLimitID string, rateLimit cloudflare.RateLimit) (cloudflare.RateLimit, error)
	DeleteRateLimit(ctx context.Context, zoneID, rateLimitID string) error
	ListRateLimits(ctx context.Context, zoneID string, pageOpts cloudflare.PaginationOptions) ([]cloudflare.RateLimit, cloudflare.ResultInfo, error)
	ListAllRateLimits(ctx context.Context, zoneID string) ([]cloudflare.RateLimit, error)

	// Bot Management operations
	GetBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error)
	UpdateBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error)

	// Turnstile operations
	CreateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	ListTurnstileWidgets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListTurnstileWidgetParams) ([]cloudflare.TurnstileWidget, *cloudflare.ResultInfo, error)
	GetTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error)
	UpdateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	RotateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, param cloudflare.RotateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	DeleteTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error

	// Workers Custom Domain operations
	AttachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domain cloudflare.AttachWorkersDomainParams) (cloudflare.WorkersDomain, error)
	GetWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) (cloudflare.WorkersDomain, error)
	DetachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) error
	ListWorkersDomains(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersDomainParams) ([]cloudflare.WorkersDomain, error)

	// Workers Subdomain operations
	WorkersCreateSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WorkersSubdomain) (cloudflare.WorkersSubdomain, error)
	WorkersGetSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.WorkersSubdomain, error)

	// Universal SSL operations
	UniversalSSLSettingDetails(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error)
	EditUniversalSSLSetting(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error)
	UniversalSSLVerificationDetails(ctx context.Context, zoneID string) ([]cloudflare.UniversalSSLVerificationDetails, error)
	UpdateUniversalSSLCertificatePackValidationMethod(ctx context.Context, zoneID string, certPackUUID string, setting cloudflare.UniversalSSLCertificatePackValidationMethodSetting) (cloudflare.UniversalSSLCertificatePackValidationMethodSetting, error)

	// Total TLS operations
	GetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error)
	SetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error)

	// Workers Script operations
	UploadWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerParams) (cloudflare.WorkerScriptResponse, error)
	GetWorker(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptResponse, error)
	DeleteWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkerParams) error
	GetWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (string, error)
	GetWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptSettingsResponse, error)
	UpdateWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersScriptContentParams) (cloudflare.WorkerScriptResponse, error)
	UpdateWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersScriptSettingsParams) (cloudflare.WorkerScriptSettingsResponse, error)
	GetWorkerWithDispatchNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string, dispatchNamespace string) (cloudflare.WorkerScriptResponse, error)
	ListWorkers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersParams) (cloudflare.WorkerListResponse, *cloudflare.ResultInfo, error)

	// Workers KV operations
	CreateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersKVNamespaceParams) (cloudflare.WorkersKVNamespaceResponse, error)
	ListWorkersKVNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVNamespacesParams) ([]cloudflare.WorkersKVNamespace, *cloudflare.ResultInfo, error)
	DeleteWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, namespaceID string) (cloudflare.Response, error)
	UpdateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersKVNamespaceParams) (cloudflare.Response, error)
	WriteWorkersKVEntry(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WriteWorkersKVEntryParams) (cloudflare.Response, error)
	WriteWorkersKVEntries(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WriteWorkersKVEntriesParams) (cloudflare.Response, error)
	DeleteWorkersKVEntries(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkersKVEntriesParams) (cloudflare.Response, error)

	// Workers CronTrigger operations
	ListWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error)
	UpdateWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error)

	// Workers Route operations
	CreateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error)
	GetWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, routeID string) (cloudflare.WorkerRouteResponse, error)
	ListWorkerRoutes(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerRoutesParams) (cloudflare.WorkerRoutesResponse, error)
	UpdateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error)
	DeleteWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, routeID string) (cloudflare.WorkerRouteResponse, error)

	// Workers Bindings operations
	ListWorkerBindings(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerBindingsParams) (cloudflare.WorkerBindingListResponse, error)

	// Workers Secrets operations
	SetWorkersSecret(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.SetWorkersSecretParams) (cloudflare.WorkersPutSecretResponse, error)
	DeleteWorkersSecret(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkersSecretParams) (cloudflare.Response, error)
	ListWorkersSecrets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersSecretsParams) (cloudflare.WorkersListSecretsResponse, error)

	// Workers Tail operations
	StartWorkersTail(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkersTail, error)
	ListWorkersTail(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersTailParameters) (cloudflare.WorkersTail, error)
	DeleteWorkersTail(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName, tailID string) error

	// Workers For Platforms operations
	ListWorkersForPlatformsDispatchNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer) (*cloudflare.ListWorkersForPlatformsDispatchNamespaceResponse, error)
	GetWorkersForPlatformsDispatchNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, name string) (*cloudflare.GetWorkersForPlatformsDispatchNamespaceResponse, error)
	CreateWorkersForPlatformsDispatchNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersForPlatformsDispatchNamespaceParams) (*cloudflare.GetWorkersForPlatformsDispatchNamespaceResponse, error)
	DeleteWorkersForPlatformsDispatchNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, name string) error

	// Workers Account Settings operations
	CreateWorkersAccountSettings(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersAccountSettingsParameters) (cloudflare.WorkersAccountSettings, error)
	WorkersAccountSettings(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WorkersAccountSettingsParameters) (cloudflare.WorkersAccountSettings, error)

	// Access Application operations
	ListAccessApplications(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListAccessApplicationsParams) ([]cloudflare.AccessApplication, *cloudflare.ResultInfo, error)
	GetAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) (cloudflare.AccessApplication, error)
	CreateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateAccessApplicationParams) (cloudflare.AccessApplication, error)
	UpdateAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateAccessApplicationParams) (cloudflare.AccessApplication, error)
	DeleteAccessApplication(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error
	RevokeAccessApplicationTokens(ctx context.Context, rc *cloudflare.ResourceContainer, applicationID string) error

	// Device Posture Rule operations
	CreateDevicePostureIntegration(ctx context.Context, accountID string, integration cloudflare.DevicePostureIntegration) (cloudflare.DevicePostureIntegration, error)
	UpdateDevicePostureIntegration(ctx context.Context, accountID string, integration cloudflare.DevicePostureIntegration) (cloudflare.DevicePostureIntegration, error)
	DevicePostureIntegration(ctx context.Context, accountID, integrationID string) (cloudflare.DevicePostureIntegration, error)
	DevicePostureIntegrations(ctx context.Context, accountID string) ([]cloudflare.DevicePostureIntegration, cloudflare.ResultInfo, error)
	DeleteDevicePostureIntegration(ctx context.Context, accountID, ruleID string) error
	DevicePostureRules(ctx context.Context, accountID string) ([]cloudflare.DevicePostureRule, cloudflare.ResultInfo, error)
	DevicePostureRule(ctx context.Context, accountID, ruleID string) (cloudflare.DevicePostureRule, error)
	CreateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	UpdateDevicePostureRule(ctx context.Context, accountID string, rule cloudflare.DevicePostureRule) (cloudflare.DevicePostureRule, error)
	DeleteDevicePostureRule(ctx context.Context, accountID, ruleID string) error

	// Tunnel operations
	ListTunnels(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelListParams) ([]cloudflare.Tunnel, *cloudflare.ResultInfo, error)
	GetTunnel(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) (cloudflare.Tunnel, error)
	CreateTunnel(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelCreateParams) (cloudflare.Tunnel, error)
	UpdateTunnel(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelUpdateParams) (cloudflare.Tunnel, error)
	UpdateTunnelConfiguration(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelConfigurationParams) (cloudflare.TunnelConfigurationResult, error)
	GetTunnelConfiguration(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) (cloudflare.TunnelConfigurationResult, error)
	ListTunnelConnections(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) ([]cloudflare.Connection, error)
	DeleteTunnel(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) error
	CleanupTunnelConnections(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) error
	GetTunnelToken(ctx context.Context, rc *cloudflare.ResourceContainer, tunnelID string) (string, error)

	// Tunnel Routes operations
	ListTunnelRoutes(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelRoutesListParams) ([]cloudflare.TunnelRoute, error)
	GetTunnelRouteForIP(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelRoutesForIPParams) (cloudflare.TunnelRoute, error)
	CreateTunnelRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelRoutesCreateParams) (cloudflare.TunnelRoute, error)
	DeleteTunnelRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelRoutesDeleteParams) error
	UpdateTunnelRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelRoutesUpdateParams) (cloudflare.TunnelRoute, error)

	// Tunnel Virtual Networks operations
	ListTunnelVirtualNetworks(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelVirtualNetworksListParams) ([]cloudflare.TunnelVirtualNetwork, error)
	CreateTunnelVirtualNetwork(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelVirtualNetworkCreateParams) (cloudflare.TunnelVirtualNetwork, error)
	DeleteTunnelVirtualNetwork(ctx context.Context, rc *cloudflare.ResourceContainer, vnetID string) error
	UpdateTunnelVirtualNetwork(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TunnelVirtualNetworkUpdateParams) (cloudflare.TunnelVirtualNetwork, error)

	// Rulesets operations
	ListRulesets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListRulesetsParams) ([]cloudflare.Ruleset, error)
	GetRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, rulesetID string) (cloudflare.Ruleset, error)
	CreateRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateRulesetParams) (cloudflare.Ruleset, error)
	DeleteRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, rulesetID string) error
	UpdateRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateRulesetParams) (cloudflare.Ruleset, error)
	DeleteRulesetRule(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteRulesetRuleParams) error
	GetEntrypointRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, phase string) (cloudflare.Ruleset, error)
	UpdateEntrypointRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateEntrypointRulesetParams) (cloudflare.Ruleset, error)

	// Load Balancing operations
	CreateLoadBalancerPool(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLoadBalancerPoolParams) (cloudflare.LoadBalancerPool, error)
	ListLoadBalancerPools(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLoadBalancerPoolParams) ([]cloudflare.LoadBalancerPool, error)
	GetLoadBalancerPool(ctx context.Context, rc *cloudflare.ResourceContainer, poolID string) (cloudflare.LoadBalancerPool, error)
	DeleteLoadBalancerPool(ctx context.Context, rc *cloudflare.ResourceContainer, poolID string) error
	UpdateLoadBalancerPool(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLoadBalancerPoolParams) (cloudflare.LoadBalancerPool, error)
	CreateLoadBalancerMonitor(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLoadBalancerMonitorParams) (cloudflare.LoadBalancerMonitor, error)
	ListLoadBalancerMonitors(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLoadBalancerMonitorParams) ([]cloudflare.LoadBalancerMonitor, error)
	GetLoadBalancerMonitor(ctx context.Context, rc *cloudflare.ResourceContainer, monitorID string) (cloudflare.LoadBalancerMonitor, error)
	DeleteLoadBalancerMonitor(ctx context.Context, rc *cloudflare.ResourceContainer, monitorID string) error
	UpdateLoadBalancerMonitor(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLoadBalancerMonitorParams) (cloudflare.LoadBalancerMonitor, error)
	CreateLoadBalancer(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLoadBalancerParams) (cloudflare.LoadBalancer, error)
	ListLoadBalancers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLoadBalancerParams) ([]cloudflare.LoadBalancer, error)
	GetLoadBalancer(ctx context.Context, rc *cloudflare.ResourceContainer, loadbalancerID string) (cloudflare.LoadBalancer, error)
	DeleteLoadBalancer(ctx context.Context, rc *cloudflare.ResourceContainer, loadbalancerID string) error
	UpdateLoadBalancer(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLoadBalancerParams) (cloudflare.LoadBalancer, error)
	GetLoadBalancerPoolHealth(ctx context.Context, rc *cloudflare.ResourceContainer, poolID string) (cloudflare.LoadBalancerPoolHealth, error)

	// Custom Hostname operations
	CreateCustomHostname(ctx context.Context, zoneID string, ch cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	CustomHostname(ctx context.Context, zoneID string, customHostnameID string) (cloudflare.CustomHostname, error)
	CustomHostnames(ctx context.Context, zoneID string, page int, filter cloudflare.CustomHostname) ([]cloudflare.CustomHostname, cloudflare.ResultInfo, error)
	CustomHostnameIDByName(ctx context.Context, zoneID string, hostname string) (string, error)
	UpdateCustomHostname(ctx context.Context, zoneID string, customHostnameID string, ch cloudflare.CustomHostname) (*cloudflare.CustomHostnameResponse, error)
	UpdateCustomHostnameSSL(ctx context.Context, zoneID string, customHostnameID string, ssl *cloudflare.CustomHostnameSSL) (*cloudflare.CustomHostnameResponse, error)
	DeleteCustomHostname(ctx context.Context, zoneID string, customHostnameID string) error
	UpdateCustomHostnameFallbackOrigin(ctx context.Context, zoneID string, chfo cloudflare.CustomHostnameFallbackOrigin) (*cloudflare.CustomHostnameFallbackOriginResponse, error)
	DeleteCustomHostnameFallbackOrigin(ctx context.Context, zoneID string) error
	CustomHostnameFallbackOrigin(ctx context.Context, zoneID string) (cloudflare.CustomHostnameFallbackOrigin, error)

	// SSL operations
	CreateSSL(ctx context.Context, zoneID string, options cloudflare.ZoneCustomSSLOptions) (cloudflare.ZoneCustomSSL, error)
	ListSSL(ctx context.Context, zoneID string) ([]cloudflare.ZoneCustomSSL, error)
	SSLDetails(ctx context.Context, zoneID, certificateID string) (cloudflare.ZoneCustomSSL, error)
	UpdateSSL(ctx context.Context, zoneID, certificateID string, options cloudflare.ZoneCustomSSLOptions) (cloudflare.ZoneCustomSSL, error)
	ReprioritizeSSL(ctx context.Context, zoneID string, p []cloudflare.ZoneCustomSSLPriority) ([]cloudflare.ZoneCustomSSL, error)
	DeleteSSL(ctx context.Context, zoneID, certificateID string) error

	// Spectrum operations
	SpectrumApplications(ctx context.Context, zoneID string) ([]cloudflare.SpectrumApplication, error)
	SpectrumApplication(ctx context.Context, zoneID string, applicationID string) (cloudflare.SpectrumApplication, error)
	CreateSpectrumApplication(ctx context.Context, zoneID string, appDetails cloudflare.SpectrumApplication) (cloudflare.SpectrumApplication, error)
	UpdateSpectrumApplication(ctx context.Context, zoneID, appID string, appDetails cloudflare.SpectrumApplication) (cloudflare.SpectrumApplication, error)
	DeleteSpectrumApplication(ctx context.Context, zoneID string, applicationID string) error

	// Cache Reserve operations
	GetCacheReserve(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetCacheReserveParams) (cloudflare.CacheReserve, error)
	UpdateCacheReserve(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateCacheReserveParams) (cloudflare.CacheReserve, error)
}

// DNSRecordValidator provides validation for DNS records
type DNSRecordValidator interface {
	ValidateSRVRecord(content string) error
	ValidateMXRecord(content string, priority int) error
	ValidateRecord(recordType, content string, priority *int) error
}

// ConfigProvider provides configuration for Cloudflare clients
type ConfigProvider interface {
	GetAPIKey() string
	GetEmail() string
	GetAPIToken() string
	IsValid() bool
}

// ClientInterface defines the interface for Workers API operations
type ClientInterface interface {
	GetAccountID() string
	UploadWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerParams) (cloudflare.WorkerScriptResponse, error)
	GetWorker(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptResponse, error)
	DeleteWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkerParams) error
	GetWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (string, error)
	GetWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptSettingsResponse, error)
	ListWorkers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersParams) (cloudflare.WorkerListResponse, *cloudflare.ResultInfo, error)
	CreateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersKVNamespaceParams) (cloudflare.WorkersKVNamespaceResponse, error)
	ListWorkersKVNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVNamespacesParams) ([]cloudflare.WorkersKVNamespace, *cloudflare.ResultInfo, error)
	DeleteWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, namespaceID string) (cloudflare.Response, error)
	UpdateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersKVNamespaceParams) (cloudflare.Response, error)
	ListWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error)
	UpdateWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error)
	ListWorkerRoutes(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerRoutesParams) (cloudflare.WorkerRoutesResponse, error)
	CreateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error)
	UpdateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error)
	DeleteWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, routeID string) (cloudflare.WorkerRouteResponse, error)

	// Additional methods expected by controllers
	CreateWorkerCronTrigger(ctx context.Context, scriptName string, cron string) (interface{}, error)
	WorkerCronTrigger(ctx context.Context, scriptName string) (interface{}, error)
	UpdateWorkerCronTrigger(ctx context.Context, scriptName string, cron string) (interface{}, error)
	DeleteWorkerCronTrigger(ctx context.Context, scriptName string) error

	CreateWorkerDomain(ctx context.Context, params v1beta1.DomainParameters) (interface{}, error)
	WorkerDomain(ctx context.Context, accountID, zoneID, domainID string) (interface{}, error)
	UpdateWorkerDomain(ctx context.Context, accountID, zoneID, domainID string, params v1beta1.DomainParameters) (interface{}, error)
	DeleteWorkerDomain(ctx context.Context, accountID, zoneID, domainID string) error

	CreateWorkerKVNamespace(ctx context.Context, params v1beta1.KVNamespaceParameters) (interface{}, error)
	WorkerKVNamespace(ctx context.Context, kvID string) (interface{}, error)
	UpdateWorkerKVNamespace(ctx context.Context, kvID string, params v1beta1.KVNamespaceParameters) (interface{}, error)
	DeleteWorkerKVNamespace(ctx context.Context, kvID string) error

	CreateWorkerSubdomain(ctx context.Context, params v1beta1.SubdomainParameters) (interface{}, error)
	WorkerSubdomain(ctx context.Context, accountID, subdomainName string) (interface{}, error)
	UpdateWorkerSubdomain(ctx context.Context, accountID, subdomainName string, params v1beta1.SubdomainParameters) (interface{}, error)
	DeleteWorkerSubdomain(ctx context.Context, accountID, subdomainName string) error
}
