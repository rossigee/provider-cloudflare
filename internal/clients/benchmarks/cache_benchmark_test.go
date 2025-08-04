package benchmarks

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
)

// MockRulesetClient provides a simple mock for cache ruleset benchmarks
type MockRulesetClient struct{}

func (c *MockRulesetClient) CreateRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateRulesetParams) (cloudflare.Ruleset, error) {
	return cloudflare.Ruleset{
		ID:          "test-ruleset-id",
		Name:        params.Name,
		Description: params.Description,
		Kind:        params.Kind,
		Phase:       params.Phase,
		Rules:       params.Rules,
	}, nil
}

func (c *MockRulesetClient) GetRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, rulesetID string) (cloudflare.Ruleset, error) {
	return cloudflare.Ruleset{
		ID:          rulesetID,
		Name:        "test-cache-ruleset",
		Description: "Test cache ruleset",
		Kind:        "zone",
		Phase:       "http_request_cache_settings",
		Rules: []cloudflare.RulesetRule{
			{
				ID:          "rule-1",
				Expression:  `http.request.uri.path matches "\\.(css|js|png|jpg)$"`,
				Action:      "set_cache_settings",
				Description: "Cache static assets",
				Enabled:     &[]bool{true}[0],
			},
		},
	}, nil
}

func (c *MockRulesetClient) UpdateRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateRulesetParams) (cloudflare.Ruleset, error) {
	return cloudflare.Ruleset{
		ID:          params.ID,
		Kind:        "zone",
		Phase:       "http_request_cache_settings",
		Rules:       params.Rules,
	}, nil
}

func (c *MockRulesetClient) DeleteRuleset(ctx context.Context, rc *cloudflare.ResourceContainer, rulesetID string) error {
	return nil
}

func (c *MockRulesetClient) ListRulesets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListRulesetsParams) ([]cloudflare.Ruleset, error) {
	rulesets := make([]cloudflare.Ruleset, 5)
	for i := 0; i < 5; i++ {
		rulesets[i] = cloudflare.Ruleset{
			ID:          "test-ruleset-" + string(rune('0'+i)),
			Name:        "cache-ruleset-" + string(rune('0'+i)),
			Description: "Test cache ruleset",
			Kind:        "zone",
			Phase:       "http_request_cache_settings",
			Rules: []cloudflare.RulesetRule{
				{
					ID:          "rule-1",
					Expression:  `http.request.uri.path matches "\\.(css|js)$"`,
					Action:      "set_cache_settings",
					Description: "Cache assets",
					Enabled:     &[]bool{true}[0],
				},
			},
		}
	}
	return rulesets, nil
}

// BenchmarkCacheRuleCreate measures cache rule creation performance
func BenchmarkCacheRuleCreate(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.CreateRulesetParams{
		Name:        "test-cache-ruleset",
		Description: "Test cache ruleset",
		Kind:        "zone",
		Phase:       "http_request_cache_settings",
		Rules: []cloudflare.RulesetRule{
			{
				Expression:  `http.request.uri.path matches "\\.(css|js|png|jpg)$"`,
				Action:      "set_cache_settings",
				Description: "Cache static assets",
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

// BenchmarkCacheRuleGet measures cache rule retrieval performance
func BenchmarkCacheRuleGet(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.GetRuleset(context.Background(), rc, "test-ruleset-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheRuleUpdate measures cache rule update performance
func BenchmarkCacheRuleUpdate(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.UpdateRulesetParams{
		ID: "test-ruleset-id",
		Rules: []cloudflare.RulesetRule{
			{
				ID:          "rule-1",
				Expression:  `http.request.uri.path matches "\\.(css|js|png|jpg|gif|svg)$"`,
				Action:      "set_cache_settings",
				Description: "Cache static assets with more file types",
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

// BenchmarkCacheRuleDelete measures cache rule deletion performance
func BenchmarkCacheRuleDelete(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		err := client.DeleteRuleset(context.Background(), rc, "test-ruleset-id")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheRuleList measures cache rule listing performance
func BenchmarkCacheRuleList(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	params := cloudflare.ListRulesetsParams{}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := client.ListRulesets(context.Background(), rc, params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheRuleBulkOperations measures bulk cache rule operations
func BenchmarkCacheRuleBulkOperations(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	rulesetsToCreate := []cloudflare.CreateRulesetParams{
		{
			Name:        "static-assets-cache",
			Description: "Cache static assets",
			Kind:        "zone",
			Phase:       "http_request_cache_settings",
			Rules: []cloudflare.RulesetRule{
				{
					Expression:  `http.request.uri.path matches "\\.(css|js)$"`,
					Action:      "set_cache_settings",
					Description: "Cache CSS/JS",
					Enabled:     &[]bool{true}[0],
				},
			},
		},
		{
			Name:        "image-cache",
			Description: "Cache images",
			Kind:        "zone",
			Phase:       "http_request_cache_settings",
			Rules: []cloudflare.RulesetRule{
				{
					Expression:  `http.request.uri.path matches "\\.(png|jpg|gif)$"`,
					Action:      "set_cache_settings",
					Description: "Cache images",
					Enabled:     &[]bool{true}[0],
				},
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

// BenchmarkCacheRuleConcurrentOperations measures concurrent cache rule operations
func BenchmarkCacheRuleConcurrentOperations(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := client.GetRuleset(context.Background(), rc, "test-ruleset-id")
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkCacheRuleComplexRules measures performance with complex cache rules
func BenchmarkCacheRuleComplexRules(b *testing.B) {
	client := &MockRulesetClient{}

	rc := &cloudflare.ResourceContainer{
		Level:      cloudflare.ZoneRouteLevel,
		Identifier: "test-zone-id",
	}

	// Complex cache rule with multiple conditions and actions
	params := cloudflare.CreateRulesetParams{
		Name:        "complex-cache-ruleset",
		Description: "Complex cache ruleset with multiple rules",
		Kind:        "zone",
		Phase:       "http_request_cache_settings",
		Rules: []cloudflare.RulesetRule{
			{
				Expression:  `(http.request.uri.path matches "\\.(css|js|png|jpg|gif|svg|woff|woff2)$") and (http.request.method eq "GET")`,
				Action:      "set_cache_settings",
				Description: "Cache static assets with complex conditions",
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