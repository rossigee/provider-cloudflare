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