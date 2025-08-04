package benchmarks

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

// BenchmarkSecurityRulesetCreate measures security ruleset creation performance
func BenchmarkSecurityRulesetCreate(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.CreateRulesetParams{
		Name:        "security-ruleset",
		Description: "Security protection ruleset",
		Kind:        "zone",
		Phase:       "http_request_firewall_custom",
		Rules: []cloudflare.RulesetRule{
			{
				Expression:  `ip.src in {192.0.2.100 203.0.113.100}`,
				Action:      "block",
				Description: "Block malicious IPs",
				Enabled:     &[]bool{true}[0],
			},
			{
				Expression:  `http.request.uri.path matches "^/admin/"`,
				Action:      "challenge",
				Description: "Challenge admin access",
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

// BenchmarkSecurityRulesetGet measures security ruleset retrieval performance
func BenchmarkSecurityRulesetGet(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.GetRuleset(context.Background(), rc, "test-security-ruleset-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSecurityRulesetUpdate measures security ruleset update performance
func BenchmarkSecurityRulesetUpdate(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.UpdateRulesetParams{
		ID: "test-security-ruleset-id",
		Rules: []cloudflare.RulesetRule{
			{
				ID:          "rule-1",
				Expression:  `ip.src in {192.0.2.100 203.0.113.100 198.51.100.100}`,
				Action:      "block",
				Description: "Block expanded malicious IP list",
				Enabled:     &[]bool{true}[0],
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.UpdateRuleset(context.Background(), rc, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSecurityRulesetDelete measures security ruleset deletion performance
func BenchmarkSecurityRulesetDelete(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := client.DeleteRuleset(context.Background(), rc, "test-security-ruleset-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSecurityRuleComplexConditions measures performance with complex security conditions
func BenchmarkSecurityRuleComplexConditions(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.CreateRulesetParams{
		Name:        "complex-security-ruleset",
		Description: "Complex security ruleset with advanced conditions",
		Kind:        "zone",
		Phase:       "http_request_firewall_custom",
		Rules: []cloudflare.RulesetRule{
			{
				Expression: `(ip.geoip.country in {"CN" "RU" "KP"}) and (http.request.uri.path matches "^/(admin|wp-admin|login)/") and (http.user_agent contains "bot" and not http.user_agent contains "Googlebot")`,
				Action:     "block",
				Description: "Block suspicious access from high-risk countries to admin areas",
				Enabled:    &[]bool{true}[0],
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

// BenchmarkSecurityRulesetConcurrentOperations measures concurrent security ruleset operations
func BenchmarkSecurityRulesetConcurrentOperations(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.GetRuleset(context.Background(), rc, "test-security-ruleset-id")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkSecurityRuleBulkOperations measures bulk security rule operations
func BenchmarkSecurityRuleBulkOperations(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	rulesetsToCreate := []cloudflare.CreateRulesetParams{
		{
			Name:        "ip-blocking-ruleset",
			Description: "IP blocking security ruleset",
			Kind:        "zone",
			Phase:       "http_request_firewall_custom",
			Rules: []cloudflare.RulesetRule{
				{Expression: `ip.src in {192.0.2.100}`, Action: "block", Description: "Block bad IP 1", Enabled: &[]bool{true}[0]},
			},
		},
		{
			Name:        "bot-protection-ruleset",
			Description: "Bot protection security ruleset",
			Kind:        "zone",
			Phase:       "http_request_firewall_custom",
			Rules: []cloudflare.RulesetRule{
				{Expression: `http.user_agent contains "bot"`, Action: "challenge", Description: "Challenge bots", Enabled: &[]bool{true}[0]},
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, params := range rulesetsToCreate {
			_, err := client.CreateRuleset(context.Background(), rc, params)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}