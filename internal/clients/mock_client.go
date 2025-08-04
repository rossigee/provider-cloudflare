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
	"fmt"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

// MockCloudflareClient implements CloudflareClient for testing
type MockCloudflareClient struct {
	// DNS Records
	CreateDNSRecordFn func(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error)
	UpdateDNSRecordFn func(ctx context.Context, zoneID, recordID string, rr cloudflare.DNSRecord) error
	DNSRecordFn       func(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error)
	DeleteDNSRecordFn func(ctx context.Context, zoneID, recordID string) error
	DNSRecordsFn      func(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error)

	// Zones
	CreateZoneFn  func(ctx context.Context, name string, jumpstart bool, zoneType string, account cloudflare.Account) (cloudflare.Zone, error)
	ZoneDetailsFn func(ctx context.Context, zoneID string) (cloudflare.Zone, error)
	DeleteZoneFn  func(ctx context.Context, zoneID string) (cloudflare.ZoneID, error)
	EditZoneFn    func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error)

	// Call tracking for test verification
	CreateDNSRecordCalls []CreateDNSRecordCall
	UpdateDNSRecordCalls []UpdateDNSRecordCall
	DNSRecordCalls       []DNSRecordCall
	DeleteDNSRecordCalls []DeleteDNSRecordCall
}

// Call tracking structures
type CreateDNSRecordCall struct {
	ZoneID string
	Record cloudflare.DNSRecord
}

type UpdateDNSRecordCall struct {
	ZoneID   string
	RecordID string
	Record   cloudflare.DNSRecord
}

type DNSRecordCall struct {
	ZoneID   string
	RecordID string
}

type DeleteDNSRecordCall struct {
	ZoneID   string
	RecordID string
}

// NewMockCloudflareClient creates a new mock client with default implementations
func NewMockCloudflareClient() *MockCloudflareClient {
	return &MockCloudflareClient{
		CreateDNSRecordCalls: make([]CreateDNSRecordCall, 0),
		UpdateDNSRecordCalls: make([]UpdateDNSRecordCall, 0),
		DNSRecordCalls:       make([]DNSRecordCall, 0),
		DeleteDNSRecordCalls: make([]DeleteDNSRecordCall, 0),
	}
}

// DNS Record operations
func (m *MockCloudflareClient) CreateDNSRecord(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
	m.CreateDNSRecordCalls = append(m.CreateDNSRecordCalls, CreateDNSRecordCall{
		ZoneID: zoneID,
		Record: rr,
	})

	if m.CreateDNSRecordFn != nil {
		return m.CreateDNSRecordFn(ctx, zoneID, rr)
	}

	// Default implementation
	return &cloudflare.DNSRecordResponse{
		Result: cloudflare.DNSRecord{
			ID:      fmt.Sprintf("record-%s-%s", zoneID, rr.Name),
			Name:    rr.Name,
			Type:    rr.Type,
			Content: rr.Content,
			TTL:     rr.TTL,
		},
	}, nil
}

func (m *MockCloudflareClient) UpdateDNSRecord(ctx context.Context, zoneID, recordID string, rr cloudflare.DNSRecord) error {
	m.UpdateDNSRecordCalls = append(m.UpdateDNSRecordCalls, UpdateDNSRecordCall{
		ZoneID:   zoneID,
		RecordID: recordID,
		Record:   rr,
	})

	if m.UpdateDNSRecordFn != nil {
		return m.UpdateDNSRecordFn(ctx, zoneID, recordID, rr)
	}

	return nil
}

func (m *MockCloudflareClient) DNSRecord(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error) {
	m.DNSRecordCalls = append(m.DNSRecordCalls, DNSRecordCall{
		ZoneID:   zoneID,
		RecordID: recordID,
	})

	if m.DNSRecordFn != nil {
		return m.DNSRecordFn(ctx, zoneID, recordID)
	}

	// Default implementation
	return cloudflare.DNSRecord{
		ID:      recordID,
		Name:    "example.com",
		Type:    "A",
		Content: "192.168.1.1",
		TTL:     300,
	}, nil
}

func (m *MockCloudflareClient) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	m.DeleteDNSRecordCalls = append(m.DeleteDNSRecordCalls, DeleteDNSRecordCall{
		ZoneID:   zoneID,
		RecordID: recordID,
	})

	if m.DeleteDNSRecordFn != nil {
		return m.DeleteDNSRecordFn(ctx, zoneID, recordID)
	}

	return nil
}

func (m *MockCloudflareClient) DNSRecords(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) ([]cloudflare.DNSRecord, error) {
	if m.DNSRecordsFn != nil {
		return m.DNSRecordsFn(ctx, zoneID, rr)
	}

	return []cloudflare.DNSRecord{}, nil
}

// Zone operations
func (m *MockCloudflareClient) CreateZone(ctx context.Context, name string, jumpstart bool, zoneType string, account cloudflare.Account) (cloudflare.Zone, error) {
	if m.CreateZoneFn != nil {
		return m.CreateZoneFn(ctx, name, jumpstart, zoneType, account)
	}

	return cloudflare.Zone{
		ID:   fmt.Sprintf("zone-%s", name),
		Name: name,
	}, nil
}

func (m *MockCloudflareClient) ZoneDetails(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
	if m.ZoneDetailsFn != nil {
		return m.ZoneDetailsFn(ctx, zoneID)
	}

	return cloudflare.Zone{
		ID:   zoneID,
		Name: "example.com",
	}, nil
}

func (m *MockCloudflareClient) DeleteZone(ctx context.Context, zoneID string) (cloudflare.ZoneID, error) {
	if m.DeleteZoneFn != nil {
		return m.DeleteZoneFn(ctx, zoneID)
	}

	return cloudflare.ZoneID{
		ID: zoneID,
	}, nil
}

func (m *MockCloudflareClient) EditZone(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
	if m.EditZoneFn != nil {
		return m.EditZoneFn(ctx, zoneID, zoneOpts)
	}

	return cloudflare.Zone{
		ID:   zoneID,
		Name: "example.com",
	}, nil
}

// Workers Custom Domain operations
func (m *MockCloudflareClient) AttachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domain cloudflare.AttachWorkersDomainParams) (cloudflare.WorkersDomain, error) {
	return cloudflare.WorkersDomain{
		ID:       "test-domain-id",
		Hostname: domain.Hostname,
		Service:  domain.Service,
		ZoneID:   domain.ZoneID,
	}, nil
}

func (m *MockCloudflareClient) GetWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) (cloudflare.WorkersDomain, error) {
	return cloudflare.WorkersDomain{
		ID:       domainID,
		Hostname: "example.com",
		Service:  "test-service",
		ZoneID:   "test-zone-id",
	}, nil
}

func (m *MockCloudflareClient) DetachWorkersDomain(ctx context.Context, rc *cloudflare.ResourceContainer, domainID string) error {
	return nil
}

func (m *MockCloudflareClient) ListWorkersDomains(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersDomainParams) ([]cloudflare.WorkersDomain, error) {
	return []cloudflare.WorkersDomain{}, nil
}

// Workers Subdomain operations
func (m *MockCloudflareClient) WorkersCreateSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.WorkersSubdomain) (cloudflare.WorkersSubdomain, error) {
	return cloudflare.WorkersSubdomain{
		Name: params.Name,
	}, nil
}

func (m *MockCloudflareClient) WorkersGetSubdomain(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.WorkersSubdomain, error) {
	return cloudflare.WorkersSubdomain{
		Name: "test-subdomain",
	}, nil
}

// Universal SSL operations
func (m *MockCloudflareClient) UniversalSSLSettingDetails(ctx context.Context, zoneID string) (cloudflare.UniversalSSLSetting, error) {
	return cloudflare.UniversalSSLSetting{
		Enabled: true,
	}, nil
}

func (m *MockCloudflareClient) EditUniversalSSLSetting(ctx context.Context, zoneID string, setting cloudflare.UniversalSSLSetting) (cloudflare.UniversalSSLSetting, error) {
	return setting, nil
}

// Total TLS operations
func (m *MockCloudflareClient) GetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.TotalTLS, error) {
	return cloudflare.TotalTLS{
		Enabled:              &[]bool{true}[0],
		CertificateAuthority: "digicert",
		ValidityDays:         90,
	}, nil
}

func (m *MockCloudflareClient) SetTotalTLS(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.TotalTLS) (cloudflare.TotalTLS, error) {
	return params, nil
}

// Certificate Pack operations
func (m *MockCloudflareClient) CreateCertificatePack(ctx context.Context, zoneID string, cert cloudflare.CertificatePackRequest) (cloudflare.CertificatePack, error) {
	return cloudflare.CertificatePack{
		ID:    "test-cert-pack-id",
		Type:  cert.Type,
		Hosts: cert.Hosts,
	}, nil
}

func (m *MockCloudflareClient) CertificatePack(ctx context.Context, zoneID, certificatePackID string) (cloudflare.CertificatePack, error) {
	return cloudflare.CertificatePack{
		ID:    certificatePackID,
		Type:  "advanced",
		Hosts: []string{"example.com"},
	}, nil
}

func (m *MockCloudflareClient) DeleteCertificatePack(ctx context.Context, zoneID, certificateID string) error {
	return nil
}

func (m *MockCloudflareClient) RestartCertificateValidation(ctx context.Context, zoneID, certificateID string) (cloudflare.CertificatePack, error) {
	return cloudflare.CertificatePack{
		ID:     certificateID,
		Type:   "advanced",
		Hosts:  []string{"example.com"},
		Status: "pending_validation",
	}, nil
}

// Origin CA Certificate operations
func (m *MockCloudflareClient) CreateOriginCACertificate(ctx context.Context, params cloudflare.CreateOriginCertificateParams) (*cloudflare.OriginCACertificate, error) {
	return &cloudflare.OriginCACertificate{
		ID:          "test-cert-id",
		Certificate: "-----BEGIN CERTIFICATE-----\nMockCertificate\n-----END CERTIFICATE-----",
		Hostnames:   params.Hostnames,
	}, nil
}

func (m *MockCloudflareClient) GetOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificate, error) {
	return &cloudflare.OriginCACertificate{
		ID:          certificateID,
		Certificate: "-----BEGIN CERTIFICATE-----\nMockCertificate\n-----END CERTIFICATE-----",
		Hostnames:   []string{"example.com"},
	}, nil
}

func (m *MockCloudflareClient) ListOriginCACertificates(ctx context.Context, params cloudflare.ListOriginCertificatesParams) ([]cloudflare.OriginCACertificate, error) {
	return []cloudflare.OriginCACertificate{}, nil
}

func (m *MockCloudflareClient) RevokeOriginCACertificate(ctx context.Context, certificateID string) (*cloudflare.OriginCACertificateID, error) {
	return &cloudflare.OriginCACertificateID{
		ID: certificateID,
	}, nil
}

// Rate Limit operations
func (m *MockCloudflareClient) CreateRateLimit(ctx context.Context, zoneID string, rateLimit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
	rateLimit.ID = "test-rate-limit-id"
	return rateLimit, nil
}

func (m *MockCloudflareClient) RateLimit(ctx context.Context, zoneID, rateLimitID string) (cloudflare.RateLimit, error) {
	return cloudflare.RateLimit{
		ID: rateLimitID,
	}, nil
}

func (m *MockCloudflareClient) UpdateRateLimit(ctx context.Context, zoneID, rateLimitID string, rateLimit cloudflare.RateLimit) (cloudflare.RateLimit, error) {
	rateLimit.ID = rateLimitID
	return rateLimit, nil
}

func (m *MockCloudflareClient) DeleteRateLimit(ctx context.Context, zoneID, rateLimitID string) error {
	return nil
}

// Bot Management operations
func (m *MockCloudflareClient) GetBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer) (cloudflare.BotManagement, error) {
	return cloudflare.BotManagement{
		EnableJS:         &[]bool{true}[0],
		FightMode:        &[]bool{false}[0],
		AutoUpdateModel:  &[]bool{true}[0],
	}, nil
}

func (m *MockCloudflareClient) UpdateBotManagement(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateBotManagementParams) (cloudflare.BotManagement, error) {
	return cloudflare.BotManagement{
		EnableJS:         params.EnableJS,
		FightMode:        params.FightMode,
		AutoUpdateModel:  params.AutoUpdateModel,
	}, nil
}

// Turnstile operations
func (m *MockCloudflareClient) CreateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
	return cloudflare.TurnstileWidget{
		SiteKey:  "test-site-key",
		Secret:   "test-secret",
		Name:     params.Name,
		Domains:  params.Domains,
		Mode:     params.Mode,
	}, nil
}

func (m *MockCloudflareClient) GetTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) (cloudflare.TurnstileWidget, error) {
	return cloudflare.TurnstileWidget{
		SiteKey: siteKey,
		Secret:  "test-secret",
		Name:    "test-widget",
		Domains: []string{"example.com"},
		Mode:    "managed",
	}, nil
}

func (m *MockCloudflareClient) UpdateTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateTurnstileWidgetParams) (cloudflare.TurnstileWidget, error) {
	var name string
	if params.Name != nil {
		name = *params.Name
	}
	var domains []string
	if params.Domains != nil {
		domains = *params.Domains
	}
	var mode string
	if params.Mode != nil {
		mode = *params.Mode
	}
	return cloudflare.TurnstileWidget{
		SiteKey: params.SiteKey,
		Secret:  "test-secret",
		Name:    name,
		Domains: domains,
		Mode:    mode,
	}, nil
}

func (m *MockCloudflareClient) DeleteTurnstileWidget(ctx context.Context, rc *cloudflare.ResourceContainer, siteKey string) error {
	return nil
}

// Validation methods
func (m *MockCloudflareClient) ValidateSRVRecord(content string) error {
	return nil
}

func (m *MockCloudflareClient) ValidateMXRecord(content string, priority int) error {
	return nil
}

func (m *MockCloudflareClient) ValidateRecord(recordType, content string, priority *int) error {
	return nil
}

// Config methods
func (m *MockCloudflareClient) GetAPIKey() string {
	return "test-api-key"
}

func (m *MockCloudflareClient) GetEmail() string {
	return "test@example.com"
}

func (m *MockCloudflareClient) GetAPIToken() string {
	return "test-api-token"
}

func (m *MockCloudflareClient) IsValid() bool {
	return true
}

// ClientInterface methods for Workers
func (m *MockCloudflareClient) GetAccountID() string {
	return "test-account-id"
}

func (m *MockCloudflareClient) UploadWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerParams) (cloudflare.WorkerScriptResponse, error) {
	return cloudflare.WorkerScriptResponse{
		WorkerScript: cloudflare.WorkerScript{
			WorkerMetaData: cloudflare.WorkerMetaData{
				ID: params.ScriptName,
			},
		},
	}, nil
}

func (m *MockCloudflareClient) GetWorker(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptResponse, error) {
	return cloudflare.WorkerScriptResponse{
		WorkerScript: cloudflare.WorkerScript{
			WorkerMetaData: cloudflare.WorkerMetaData{
				ID: scriptName,
			},
		},
	}, nil
}

func (m *MockCloudflareClient) DeleteWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkerParams) error {
	return nil
}

func (m *MockCloudflareClient) GetWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (string, error) {
	return "addEventListener('fetch', event => { event.respondWith(new Response('Hello World!')) })", nil
}

func (m *MockCloudflareClient) GetWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptSettingsResponse, error) {
	return cloudflare.WorkerScriptSettingsResponse{
		WorkerMetaData: cloudflare.WorkerMetaData{
			ID: scriptName,
		},
	}, nil
}

func (m *MockCloudflareClient) ListWorkers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersParams) (cloudflare.WorkerListResponse, *cloudflare.ResultInfo, error) {
	return cloudflare.WorkerListResponse{
		WorkerList: []cloudflare.WorkerMetaData{},
	}, nil, nil
}

func (m *MockCloudflareClient) CreateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersKVNamespaceParams) (cloudflare.WorkersKVNamespaceResponse, error) {
	return cloudflare.WorkersKVNamespaceResponse{
		Result: cloudflare.WorkersKVNamespace{
			ID:    "test-namespace-id",
			Title: params.Title,
		},
	}, nil
}

func (m *MockCloudflareClient) ListWorkersKVNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVNamespacesParams) ([]cloudflare.WorkersKVNamespace, *cloudflare.ResultInfo, error) {
	return []cloudflare.WorkersKVNamespace{}, nil, nil
}

func (m *MockCloudflareClient) DeleteWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, namespaceID string) (cloudflare.Response, error) {
	return cloudflare.Response{}, nil
}

func (m *MockCloudflareClient) UpdateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersKVNamespaceParams) (cloudflare.Response, error) {
	return cloudflare.Response{}, nil
}

func (m *MockCloudflareClient) ListWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	return []cloudflare.WorkerCronTrigger{}, nil
}

func (m *MockCloudflareClient) UpdateWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	return params.Crons, nil
}

// Helper methods for test verification
func (m *MockCloudflareClient) WasCreateDNSRecordCalled() bool {
	return len(m.CreateDNSRecordCalls) > 0
}

func (m *MockCloudflareClient) GetLastCreateDNSRecordCall() *CreateDNSRecordCall {
	if len(m.CreateDNSRecordCalls) == 0 {
		return nil
	}
	return &m.CreateDNSRecordCalls[len(m.CreateDNSRecordCalls)-1]
}

func (m *MockCloudflareClient) WasUpdateDNSRecordCalled() bool {
	return len(m.UpdateDNSRecordCalls) > 0
}

func (m *MockCloudflareClient) GetLastUpdateDNSRecordCall() *UpdateDNSRecordCall {
	if len(m.UpdateDNSRecordCalls) == 0 {
		return nil
	}
	return &m.UpdateDNSRecordCalls[len(m.UpdateDNSRecordCalls)-1]
}

func (m *MockCloudflareClient) Reset() {
	m.CreateDNSRecordCalls = make([]CreateDNSRecordCall, 0)
	m.UpdateDNSRecordCalls = make([]UpdateDNSRecordCall, 0)
	m.DNSRecordCalls = make([]DNSRecordCall, 0)
	m.DeleteDNSRecordCalls = make([]DeleteDNSRecordCall, 0)
}

// MockConfig implements ConfigProvider for testing
type MockConfig struct {
	APIKey   string
	Email    string
	APIToken string
	Valid    bool
}

func (c *MockConfig) GetAPIKey() string {
	return c.APIKey
}

func (c *MockConfig) GetEmail() string {
	return c.Email
}

func (c *MockConfig) GetAPIToken() string {
	return c.APIToken
}

func (c *MockConfig) IsValid() bool {
	return c.Valid
}

// NewValidMockConfig creates a mock config with valid credentials
func NewValidMockConfig() *MockConfig {
	return &MockConfig{
		APIKey:   "valid-api-key",
		Email:    "test@example.com",
		APIToken: "valid-api-token",
		Valid:    true,
	}
}

// NewInvalidMockConfig creates a mock config with invalid credentials
func NewInvalidMockConfig() *MockConfig {
	return &MockConfig{
		APIKey:   "",
		Email:    "",
		APIToken: "",
		Valid:    false,
	}
}

// MockClient implements ClientInterface for Workers testing using testify/mock pattern
type MockClient struct {
	accountID        string
	calls            map[string][]interface{}
	responses        map[string]interface{}
	errors           map[string]error
	lastMethodCalled string
	callCounter      map[string]int
}

// NewMockClient creates a new MockClient
func NewMockClient() *MockClient {
	return &MockClient{
		accountID:   "test-account-id", 
		calls:       make(map[string][]interface{}),
		responses:   make(map[string]interface{}),
		errors:      make(map[string]error),
		callCounter: make(map[string]int),
	}
}

// On sets up a mock expectation (simplified testify/mock pattern)
func (m *MockClient) On(methodName string, args ...interface{}) *MockClient {
	// Store the expected args for this method call
	if m.calls[methodName] == nil {
		m.calls[methodName] = args
	}
	m.lastMethodCalled = methodName // Track this for Return()
	return m
}

// Return sets the return value for the mock
func (m *MockClient) Return(values ...interface{}) *MockClient {
	if m.lastMethodCalled != "" {
		if len(values) == 1 {
			// Check if the single value is an error
			if err, ok := values[0].(error); ok {
				m.errors[m.lastMethodCalled] = err
			} else {
				m.responses[m.lastMethodCalled] = values[0]
			}
		} else if len(values) == 2 {
			// First value is response, second value is error
			if values[0] != nil {
				m.responses[m.lastMethodCalled] = values[0]
			}
			if values[1] != nil {
				if err, ok := values[1].(error); ok {
					m.errors[m.lastMethodCalled] = err
				}
			}
		} else if len(values) == 3 {
			// For methods that return (result, resultInfo, error) like ListWorkersKVNamespaces
			if values[0] != nil {
				m.responses[m.lastMethodCalled] = values[0]
			}
			// values[1] is ResultInfo which we'll ignore for now
			if values[2] != nil {
				if err, ok := values[2].(error); ok {
					m.errors[m.lastMethodCalled] = err
				}
			}
		}
	}
	return m
}

// GetAccountID returns the mock account ID
func (m *MockClient) GetAccountID() string {
	if m.accountID != "" {
		return m.accountID
	}
	if response, ok := m.responses["GetAccountID"]; ok {
		return response.(string)
	}
	return "test-account-id"
}

// UploadWorker mocks the UploadWorker method
func (m *MockClient) UploadWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerParams) (cloudflare.WorkerScriptResponse, error) {
	if err, ok := m.errors["UploadWorker"]; ok {
		return cloudflare.WorkerScriptResponse{}, err
	}
	if response, ok := m.responses["UploadWorker"]; ok {
		return response.(cloudflare.WorkerScriptResponse), nil
	}
	return cloudflare.WorkerScriptResponse{
		WorkerScript: cloudflare.WorkerScript{
			WorkerMetaData: cloudflare.WorkerMetaData{
				ID: "test-script-id",
			},
		},
	}, nil
}

// GetWorker mocks the GetWorker method
func (m *MockClient) GetWorker(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptResponse, error) {
	if err, ok := m.errors["GetWorker"]; ok {
		return cloudflare.WorkerScriptResponse{}, err
	}
	if response, ok := m.responses["GetWorker"]; ok {
		return response.(cloudflare.WorkerScriptResponse), nil
	}
	return cloudflare.WorkerScriptResponse{
		WorkerScript: cloudflare.WorkerScript{
			WorkerMetaData: cloudflare.WorkerMetaData{
				ID: scriptName,
			},
		},
	}, nil
}

// DeleteWorker mocks the DeleteWorker method
func (m *MockClient) DeleteWorker(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeleteWorkerParams) error {
	if err, ok := m.errors["DeleteWorker"]; ok {
		return err
	}
	return nil
}

// GetWorkersScriptContent mocks the GetWorkersScriptContent method
func (m *MockClient) GetWorkersScriptContent(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (string, error) {
	if err, ok := m.errors["GetWorkersScriptContent"]; ok {
		return "", err
	}
	if response, ok := m.responses["GetWorkersScriptContent"]; ok {
		return response.(string), nil
	}
	return "addEventListener('fetch', event => { event.respondWith(new Response('Hello World!')) })", nil
}

// GetWorkersScriptSettings mocks the GetWorkersScriptSettings method
func (m *MockClient) GetWorkersScriptSettings(ctx context.Context, rc *cloudflare.ResourceContainer, scriptName string) (cloudflare.WorkerScriptSettingsResponse, error) {
	if err, ok := m.errors["GetWorkersScriptSettings"]; ok {
		return cloudflare.WorkerScriptSettingsResponse{}, err
	}
	if response, ok := m.responses["GetWorkersScriptSettings"]; ok {
		return response.(cloudflare.WorkerScriptSettingsResponse), nil
	}
	return cloudflare.WorkerScriptSettingsResponse{
		WorkerMetaData: cloudflare.WorkerMetaData{
			ID: scriptName,
		},
	}, nil
}

// ListWorkers mocks the ListWorkers method
func (m *MockClient) ListWorkers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersParams) (cloudflare.WorkerListResponse, *cloudflare.ResultInfo, error) {
	if err, ok := m.errors["ListWorkers"]; ok {
		return cloudflare.WorkerListResponse{}, nil, err
	}
	if response, ok := m.responses["ListWorkers"]; ok {
		return response.(cloudflare.WorkerListResponse), nil, nil
	}
	return cloudflare.WorkerListResponse{
		WorkerList: []cloudflare.WorkerMetaData{},
	}, nil, nil
}

// CreateWorkersKVNamespace mocks the CreateWorkersKVNamespace method
func (m *MockClient) CreateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkersKVNamespaceParams) (cloudflare.WorkersKVNamespaceResponse, error) {
	if err, ok := m.errors["CreateWorkersKVNamespace"]; ok {
		return cloudflare.WorkersKVNamespaceResponse{}, err
	}
	if response, ok := m.responses["CreateWorkersKVNamespace"]; ok {
		return response.(cloudflare.WorkersKVNamespaceResponse), nil
	}
	return cloudflare.WorkersKVNamespaceResponse{
		Result: cloudflare.WorkersKVNamespace{
			ID:    "test-namespace-id",
			Title: params.Title,
		},
	}, nil
}

// ListWorkersKVNamespaces mocks the ListWorkersKVNamespaces method
func (m *MockClient) ListWorkersKVNamespaces(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkersKVNamespacesParams) ([]cloudflare.WorkersKVNamespace, *cloudflare.ResultInfo, error) {
	if err, ok := m.errors["ListWorkersKVNamespaces"]; ok {
		return nil, nil, err
	}
	if response, ok := m.responses["ListWorkersKVNamespaces"]; ok {
		// Handle both slice and response struct types
		switch resp := response.(type) {
		case []cloudflare.WorkersKVNamespace:
			return resp, &cloudflare.ResultInfo{}, nil
		case struct{
			Result []cloudflare.WorkersKVNamespace
			ResultInfo *cloudflare.ResultInfo
		}:
			return resp.Result, resp.ResultInfo, nil
		}
	}
	return []cloudflare.WorkersKVNamespace{
		{
			ID:    "test-namespace-id",
			Title: "test-namespace",
		},
	}, &cloudflare.ResultInfo{}, nil
}

// DeleteWorkersKVNamespace mocks the DeleteWorkersKVNamespace method
func (m *MockClient) DeleteWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, namespaceID string) (cloudflare.Response, error) {
	if err, ok := m.errors["DeleteWorkersKVNamespace"]; ok {
		return cloudflare.Response{}, err
	}
	if response, ok := m.responses["DeleteWorkersKVNamespace"]; ok {
		return response.(cloudflare.Response), nil
	}
	return cloudflare.Response{}, nil
}

// UpdateWorkersKVNamespace mocks the UpdateWorkersKVNamespace method
func (m *MockClient) UpdateWorkersKVNamespace(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkersKVNamespaceParams) (cloudflare.Response, error) {
	if err, ok := m.errors["UpdateWorkersKVNamespace"]; ok {
		return cloudflare.Response{}, err
	}
	if response, ok := m.responses["UpdateWorkersKVNamespace"]; ok {
		return response.(cloudflare.Response), nil
	}
	return cloudflare.Response{}, nil
}

// ListWorkerCronTriggers mocks the ListWorkerCronTriggers method
func (m *MockClient) ListWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	if err, ok := m.errors["ListWorkerCronTriggers"]; ok {
		return nil, err
	}
	if response, ok := m.responses["ListWorkerCronTriggers"]; ok {
		return response.([]cloudflare.WorkerCronTrigger), nil
	}
	return []cloudflare.WorkerCronTrigger{}, nil
}

// UpdateWorkerCronTriggers mocks the UpdateWorkerCronTriggers method
func (m *MockClient) UpdateWorkerCronTriggers(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerCronTriggersParams) ([]cloudflare.WorkerCronTrigger, error) {
	if err, ok := m.errors["UpdateWorkerCronTriggers"]; ok {
		return nil, err
	}
	if response, ok := m.responses["UpdateWorkerCronTriggers"]; ok {
		return response.([]cloudflare.WorkerCronTrigger), nil
	}
	return params.Crons, nil
}

// ListWorkerRoutes mocks the ListWorkerRoutes method
func (m *MockClient) ListWorkerRoutes(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListWorkerRoutesParams) (cloudflare.WorkerRoutesResponse, error) {
	if err, ok := m.errors["ListWorkerRoutes"]; ok {
		return cloudflare.WorkerRoutesResponse{}, err
	}
	if response, ok := m.responses["ListWorkerRoutes"]; ok {
		return response.(cloudflare.WorkerRoutesResponse), nil
	}
	return cloudflare.WorkerRoutesResponse{Routes: []cloudflare.WorkerRoute{}}, nil
}

// CreateWorkerRoute mocks the CreateWorkerRoute method
func (m *MockClient) CreateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error) {
	if err, ok := m.errors["CreateWorkerRoute"]; ok {
		return cloudflare.WorkerRouteResponse{}, err
	}
	if response, ok := m.responses["CreateWorkerRoute"]; ok {
		return response.(cloudflare.WorkerRouteResponse), nil
	}
	return cloudflare.WorkerRouteResponse{
		WorkerRoute: cloudflare.WorkerRoute{
			ID:      "test-route-id",
			Pattern: params.Pattern,
			ScriptName: params.Script,
		},
	}, nil
}

// UpdateWorkerRoute mocks the UpdateWorkerRoute method
func (m *MockClient) UpdateWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateWorkerRouteParams) (cloudflare.WorkerRouteResponse, error) {
	if err, ok := m.errors["UpdateWorkerRoute"]; ok {
		return cloudflare.WorkerRouteResponse{}, err
	}
	if response, ok := m.responses["UpdateWorkerRoute"]; ok {
		return response.(cloudflare.WorkerRouteResponse), nil
	}
	return cloudflare.WorkerRouteResponse{
		WorkerRoute: cloudflare.WorkerRoute{
			ID:      params.ID,
			Pattern: params.Pattern,
			ScriptName: params.Script,
		},
	}, nil
}

// DeleteWorkerRoute mocks the DeleteWorkerRoute method
func (m *MockClient) DeleteWorkerRoute(ctx context.Context, rc *cloudflare.ResourceContainer, routeID string) (cloudflare.WorkerRouteResponse, error) {
	if err, ok := m.errors["DeleteWorkerRoute"]; ok {
		return cloudflare.WorkerRouteResponse{}, err
	}
	if response, ok := m.responses["DeleteWorkerRoute"]; ok {
		return response.(cloudflare.WorkerRouteResponse), nil
	}
	return cloudflare.WorkerRouteResponse{}, nil
}

// NewNotFoundError creates a not found error for testing
func NewNotFoundError(message string) error {
	return fmt.Errorf("not found: %s", message)
}

// IsNotFound checks if an error indicates a resource was not found
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "not found")
}
