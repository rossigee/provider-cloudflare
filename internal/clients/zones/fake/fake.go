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
	MockCreateZone          func(ctx context.Context, name string, jumpstart bool, account cloudflare.Account, zoneType string) (cloudflare.Zone, error)
	MockDeleteZone          func(ctx context.Context, zoneID string) (cloudflare.ZoneID, error)
	MockEditZone            func(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error)
	MockUpdateZoneSettings  func(ctx context.Context, zoneID string, cs []cloudflare.ZoneSetting) (*cloudflare.ZoneSettingResponse, error)
	MockZoneDetails         func(ctx context.Context, zoneID string) (cloudflare.Zone, error)
	MockZoneIDByName        func(zoneName string) (string, error)
	MockZoneSetPlan         func(ctx context.Context, zoneID string, planType string) error
	MockZoneSettings        func(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error)
	MockZoneDNSSECSetting   func(ctx context.Context, zoneID string) (cloudflare.ZoneDNSSEC, error)
	MockUpdateZoneDNSSEC    func(ctx context.Context, zoneID string, options cloudflare.ZoneDNSSECUpdateOptions) (cloudflare.ZoneDNSSEC, error)
	MockDeleteZoneDNSSEC    func(ctx context.Context, zoneID string) (string, error)
	MockZoneActivationCheck func(ctx context.Context, zoneID string) (cloudflare.Response, error)
	MockPurgeEverything     func(ctx context.Context, zoneID string) (cloudflare.PurgeCacheResponse, error)
	MockPurgeCache          func(ctx context.Context, zoneID string, pcr cloudflare.PurgeCacheRequest) (cloudflare.PurgeCacheResponse, error)
	MockListZones           func(ctx context.Context, z ...string) ([]cloudflare.Zone, error)
}

// CreateZone mocks the CreateZone method of the Cloudflare API.
func (m MockClient) CreateZone(ctx context.Context, name string, jumpstart bool, account cloudflare.Account, zoneType string) (cloudflare.Zone, error) {
	return m.MockCreateZone(ctx, name, jumpstart, account, zoneType)
}

// DeleteZone mocks the DeleteZone method of the Cloudflare API.
func (m MockClient) DeleteZone(ctx context.Context, zoneID string) (cloudflare.ZoneID, error) {
	return m.MockDeleteZone(ctx, zoneID)
}

// EditZone mocks the EditZone method of the Cloudflare API.
func (m MockClient) EditZone(ctx context.Context, zoneID string, zoneOpts cloudflare.ZoneOptions) (cloudflare.Zone, error) {
	return m.MockEditZone(ctx, zoneID, zoneOpts)
}

// UpdateZoneSettings mocks the UpdateZoneSettings method of the Cloudflare API.
func (m MockClient) UpdateZoneSettings(ctx context.Context, zoneID string, cs []cloudflare.ZoneSetting) (*cloudflare.ZoneSettingResponse, error) {
	return m.MockUpdateZoneSettings(ctx, zoneID, cs)
}

// ZoneDetails mocks the ZoneDetails method of the Cloudflare API.
func (m MockClient) ZoneDetails(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
	return m.MockZoneDetails(ctx, zoneID)
}

// ZoneIDByName mocks the ZoneIDByName method of the Cloudflare API.
func (m MockClient) ZoneIDByName(zoneName string) (string, error) {
	return m.MockZoneIDByName(zoneName)
}

// ZoneSetPlan mocks the ZoneSetPlan method of the Cloudflare API.
func (m MockClient) ZoneSetPlan(ctx context.Context, zoneID string, planType string) error {
	return m.MockZoneSetPlan(ctx, zoneID, planType)
}

// ZoneSettings mocks the ZoneSettings method of the Cloudflare API.
func (m MockClient) ZoneSettings(ctx context.Context, zoneID string) (*cloudflare.ZoneSettingResponse, error) {
	return m.MockZoneSettings(ctx, zoneID)
}

// ZoneDNSSECSetting mocks the ZoneDNSSECSetting method of the Cloudflare API.
func (m MockClient) ZoneDNSSECSetting(ctx context.Context, zoneID string) (cloudflare.ZoneDNSSEC, error) {
	if m.MockZoneDNSSECSetting != nil {
		return m.MockZoneDNSSECSetting(ctx, zoneID)
	}
	return cloudflare.ZoneDNSSEC{Status: "disabled"}, nil
}

// UpdateZoneDNSSEC mocks the UpdateZoneDNSSEC method of the Cloudflare API.
func (m MockClient) UpdateZoneDNSSEC(ctx context.Context, zoneID string, options cloudflare.ZoneDNSSECUpdateOptions) (cloudflare.ZoneDNSSEC, error) {
	if m.MockUpdateZoneDNSSEC != nil {
		return m.MockUpdateZoneDNSSEC(ctx, zoneID, options)
	}
	return cloudflare.ZoneDNSSEC{Status: options.Status}, nil
}

// DeleteZoneDNSSEC mocks the DeleteZoneDNSSEC method of the Cloudflare API.
func (m MockClient) DeleteZoneDNSSEC(ctx context.Context, zoneID string) (string, error) {
	if m.MockDeleteZoneDNSSEC != nil {
		return m.MockDeleteZoneDNSSEC(ctx, zoneID)
	}
	return "disabled", nil
}

// ZoneActivationCheck mocks the ZoneActivationCheck method of the Cloudflare API.
func (m MockClient) ZoneActivationCheck(ctx context.Context, zoneID string) (cloudflare.Response, error) {
	if m.MockZoneActivationCheck != nil {
		return m.MockZoneActivationCheck(ctx, zoneID)
	}
	return cloudflare.Response{Success: true}, nil
}

// PurgeEverything mocks the PurgeEverything method of the Cloudflare API.
func (m MockClient) PurgeEverything(ctx context.Context, zoneID string) (cloudflare.PurgeCacheResponse, error) {
	if m.MockPurgeEverything != nil {
		return m.MockPurgeEverything(ctx, zoneID)
	}
	return cloudflare.PurgeCacheResponse{Response: cloudflare.Response{Success: true}}, nil
}

// PurgeCache mocks the PurgeCache method of the Cloudflare API.
func (m MockClient) PurgeCache(ctx context.Context, zoneID string, pcr cloudflare.PurgeCacheRequest) (cloudflare.PurgeCacheResponse, error) {
	if m.MockPurgeCache != nil {
		return m.MockPurgeCache(ctx, zoneID, pcr)
	}
	return cloudflare.PurgeCacheResponse{Response: cloudflare.Response{Success: true}}, nil
}

// ListZones mocks the ListZones method of the Cloudflare API.
func (m MockClient) ListZones(ctx context.Context, z ...string) ([]cloudflare.Zone, error) {
	if m.MockListZones != nil {
		return m.MockListZones(ctx, z...)
	}
	return []cloudflare.Zone{}, nil
}
