package benchmarks

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/rossigee/provider-cloudflare/internal/clients/zones/fake"
)

// BenchmarkZoneCreateSimple measures simple zone creation performance
func BenchmarkZoneCreateSimple(b *testing.B) {
	client := &fake.MockClient{
		MockCreateZone: func(ctx context.Context, name string, jumpstart bool, account cloudflare.Account, zoneType string) (cloudflare.Zone, error) {
			return cloudflare.Zone{
				ID:       "test-zone-id",
				Name:     name,
				Type:     zoneType,
				Account:  account,
				Paused:   false,
				VanityNS: []string{"ns1.example.com", "ns2.example.com"},
			}, nil
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.CreateZone(context.Background(), "example.com", false, cloudflare.Account{ID: "test-account"}, "full")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkZoneDetailsSimple measures simple zone details retrieval performance
func BenchmarkZoneDetailsSimple(b *testing.B) {
	client := &fake.MockClient{
		MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
			return cloudflare.Zone{
				ID:   zoneID,
				Name: "example.com",
				Type: "full",
				Account: cloudflare.Account{
					ID:   "test-account",
					Name: "Test Account",
				},
				Paused:   false,
				VanityNS: []string{"ns1.example.com", "ns2.example.com"},
			}, nil
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.ZoneDetails(context.Background(), "test-zone-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSimpleRulesetCreate measures simple ruleset creation performance
func BenchmarkSimpleRulesetCreate(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.CreateRulesetParams{
		Name:        "test-simple-ruleset",
		Description: "Test simple ruleset",
		Kind:        "zone",
		Phase:       "http_request_firewall_custom",
		Rules: []cloudflare.RulesetRule{
			{
				Expression:  `ip.src eq 192.0.2.1`,
				Action:      "block",
				Description: "Block test IP",
				Enabled:     &[]bool{true}[0],
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.CreateRuleset(context.Background(), rc, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentOperations measures concurrent operations performance
func BenchmarkConcurrentOperations(b *testing.B) {
	client := &fake.MockClient{
		MockZoneDetails: func(ctx context.Context, zoneID string) (cloudflare.Zone, error) {
			return cloudflare.Zone{
				ID:   zoneID,
				Name: "example.com",
				Type: "full",
			}, nil
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.ZoneDetails(context.Background(), "test-zone-id")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}