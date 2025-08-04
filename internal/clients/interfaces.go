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
)

// CloudflareClient represents the interface for Cloudflare API operations
type CloudflareClient interface {
	// DNS Record operations
	CreateDNSRecord(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error)
	UpdateDNSRecord(ctx context.Context, zoneID, recordID string, rr cloudflare.DNSRecord) error
	DNSRecord(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error
	DNSRecords(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error)

	// Zone operations
	CreateZone(ctx context.Context, name string, jumpstart bool, zoneType string, account cloudflare.Account) (cloudflare.Zone, error)
	ZoneDetails(ctx context.Context, zoneID string) (cloudflare.Zone, error)
	DeleteZone(ctx context.Context, zoneID string) (cloudflare.ZoneID, error)
	EditZone(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error)

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

	// Bot Management operations
	GetBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error)
	UpdateBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error)

	// Turnstile operations
	CreateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
	GetTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error)
	UpdateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error)
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

	// Total TLS operations
	GetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error)
	SetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error)

	// Certificate Pack operations
	CreateCertificatePack(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error)
	CertificatePack(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error)
	DeleteCertificatePack(ctx context.Context, zoneID, certificateID string) error
	RestartCertificateValidation(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error)
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
}