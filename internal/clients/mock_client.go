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
