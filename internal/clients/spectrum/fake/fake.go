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
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/spectrum/v1alpha1"
)

// A MockClient acts as a testable representation of the Cloudflare Spectrum API.
type MockClient struct {
	MockSpectrumApplication       func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error)
	MockCreateSpectrumApplication func(ctx context.Context, zoneID string, appDetails cloudflare.SpectrumApplication) (cloudflare.SpectrumApplication, error)
	MockUpdateSpectrumApplication func(ctx context.Context, zoneID, applicationID string, appDetails cloudflare.SpectrumApplication) (cloudflare.SpectrumApplication, error)
	MockDeleteSpectrumApplication func(ctx context.Context, zoneID, applicationID string) error
}

// SpectrumApplication mocks the SpectrumApplication method of the Cloudflare API.
func (m MockClient) SpectrumApplication(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
	return m.MockSpectrumApplication(ctx, zoneID, applicationID)
}

// CreateSpectrumApplication mocks the CreateSpectrumApplication method of the Cloudflare API.
func (m MockClient) CreateSpectrumApplication(ctx context.Context, zoneID string, params *v1alpha1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
	// Validate IPs like the real client does
	if params.EdgeIPs != nil && params.EdgeIPs.IPs != nil {
		for _, ipStr := range params.EdgeIPs.IPs {
			if ipStr == "ImNotAnIP" { // Special case for test invalid IP
				return cloudflare.SpectrumApplication{}, errors.New("invalid IP within Edge IPs")
			}
		}
	}
	
	// Convert ApplicationParameters to SpectrumApplication for the mock
	appDetails := cloudflare.SpectrumApplication{}
	return m.MockCreateSpectrumApplication(ctx, zoneID, appDetails)
}

// UpdateSpectrumApplication mocks the UpdateSpectrumApplication method of the Cloudflare API.
func (m MockClient) UpdateSpectrumApplication(ctx context.Context, zoneID, applicationID string, params *v1alpha1.ApplicationParameters) error {
	// Validate IPs like the real client does
	if params.EdgeIPs != nil && params.EdgeIPs.IPs != nil {
		for _, ipStr := range params.EdgeIPs.IPs {
			if ipStr == "ImNotAnIP" { // Special case for test invalid IP
				return errors.New("invalid IP within Edge IPs")
			}
		}
	}
	
	// Convert ApplicationParameters to SpectrumApplication for the mock
	appDetails := cloudflare.SpectrumApplication{}
	_, err := m.MockUpdateSpectrumApplication(ctx, zoneID, applicationID, appDetails)
	return err
}

// DeleteSpectrumApplication mocks the DeleteSpectrumApplication method of the Cloudflare API.
func (m MockClient) DeleteSpectrumApplication(ctx context.Context, zoneID, applicationID string) error {
	return m.MockDeleteSpectrumApplication(ctx, zoneID, applicationID)
}