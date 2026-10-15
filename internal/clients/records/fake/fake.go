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

package fake

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

// A MockClient acts as a testable representation of the Cloudflare API.
type MockClient struct {
	MockCreateDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error)
	MockUpdateDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error)
	MockGetDNSRecord     func(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error)
	MockDeleteDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error
	MockListDNSRecords   func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error)
	MockExportDNSRecords func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ExportDNSRecordsParams) (string, error)
	MockImportDNSRecords func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ImportDNSRecordsParams) error
}

// CreateDNSRecord mocks the CreateDNSRecord method of the Cloudflare API.
func (m MockClient) CreateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error) {
	if m.MockCreateDNSRecord != nil {
		return m.MockCreateDNSRecord(ctx, rc, params)
	}
	return cloudflare.DNSRecord{}, nil
}

// UpdateDNSRecord mocks the UpdateDNSRecord method of the Cloudflare API.
func (m MockClient) UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	if m.MockUpdateDNSRecord != nil {
		return m.MockUpdateDNSRecord(ctx, rc, params)
	}
	return cloudflare.DNSRecord{}, nil
}

// GetDNSRecord mocks the GetDNSRecord method of the Cloudflare API.
func (m MockClient) GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error) {
	if m.MockGetDNSRecord != nil {
		return m.MockGetDNSRecord(ctx, rc, recordID)
	}
	return cloudflare.DNSRecord{}, nil
}

// DeleteDNSRecord mocks the DeleteDNSRecord method of the Cloudflare API.
func (m MockClient) DeleteDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error {
	if m.MockDeleteDNSRecord != nil {
		return m.MockDeleteDNSRecord(ctx, rc, recordID)
	}
	return nil
}

// ListDNSRecords mocks the ListDNSRecords method of the Cloudflare API.
func (m MockClient) ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	if m.MockListDNSRecords != nil {
		return m.MockListDNSRecords(ctx, rc, params)
	}
	return []cloudflare.DNSRecord{}, &cloudflare.ResultInfo{}, nil
}

// ExportDNSRecords mocks the ExportDNSRecords method of the Cloudflare API.
func (m MockClient) ExportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ExportDNSRecordsParams) (string, error) {
	if m.MockExportDNSRecords != nil {
		return m.MockExportDNSRecords(ctx, rc, params)
	}
	return "", nil
}

// ImportDNSRecords mocks the ImportDNSRecords method of the Cloudflare API.
func (m MockClient) ImportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ImportDNSRecordsParams) error {
	if m.MockImportDNSRecords != nil {
		return m.MockImportDNSRecords(ctx, rc, params)
	}
	return nil
}
