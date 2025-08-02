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

package test

import (
	"context"
	"os"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

// IntegrationTest demonstrates how to run integration tests with real Cloudflare API
// Set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID environment variables to run
func TestSRVRecordIntegration(t *testing.T) {
	apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
	zoneID := os.Getenv("CLOUDFLARE_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		t.Skip("Skipping integration test - set CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID environment variables to run")
	}

	// Initialize Cloudflare client
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		t.Fatalf("Failed to create Cloudflare client: %v", err)
	}

	ctx := context.Background()

	// Test creating an SRV record
	srvRecord := cloudflare.DNSRecord{
		Type:    "SRV",
		Name:    "_test._tcp",
		Content: "10 20 8080 target.example.com",
		TTL:     300,
	}

	t.Run("CreateSRVRecord", func(t *testing.T) {
		resp, err := api.CreateDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateDNSRecordParams{
			Type:    srvRecord.Type,
			Name:    srvRecord.Name,
			Content: srvRecord.Content,
			TTL:     srvRecord.TTL,
		})
		if err != nil {
			t.Fatalf("Failed to create SRV record: %v", err)
		}

		// Verify the record was created correctly
		if resp.Type != "SRV" {
			t.Errorf("Expected SRV record type, got %s", resp.Type)
		}

		if resp.Content != srvRecord.Content {
			t.Errorf("Expected content %s, got %s", srvRecord.Content, resp.Content)
		}

		// Clean up - delete the test record
		t.Cleanup(func() {
			err := api.DeleteDNSRecord(ctx, cloudflare.ZoneIdentifier(zoneID), resp.ID)
			if err != nil {
				t.Logf("Failed to cleanup test record: %v", err)
			}
		})
	})
}

// BenchmarkSRVRecordValidation benchmarks SRV record content validation
func BenchmarkSRVRecordValidation(b *testing.B) {
	testCases := []string{
		"10 20 8080 target.example.com",
		"5 10 443 web.example.com",
		"0 0 22 ssh.example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, content := range testCases {
			// Simple validation - in real implementation this would be more comprehensive
			if len(content) < 10 {
				b.Error("Invalid SRV content")
			}
		}
	}
}

// ExampleSRVRecord demonstrates how to create SRV records with the provider
func Example_srvRecord() {
	// This would be a YAML manifest for the Crossplane resource
	srvExample := `
apiVersion: dns.cloudflare.crossplane.io/v1alpha1
kind: Record
metadata:
  name: srv-record-example
spec:
  forProvider:
    type: SRV
    name: "_service._tcp.example"
    content: "10 20 8080 target.example.com"
    zone: "example.com"
    ttl: 300
  providerConfigRef:
    name: cloudflare-config
`
	// In a real test, this would be applied to a Kubernetes cluster
	// with Crossplane and the Cloudflare provider installed
	_ = srvExample
}