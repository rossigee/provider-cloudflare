/*
Copyright 2026 The Crossplane Authors.

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

package aigateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cloudflare "github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"

	"github.com/rossigee/provider-cloudflare/apis/aigateway/v1beta1"
)

const (
	testAccountID = "acc-123"
	testGatewayID = "my-gateway"

	gatewayJSON = `{"id":"my-gateway","cache_invalidate_on_update":true,"cache_ttl":3600,` +
		`"collect_logs":true,"created_at":"2026-01-02T03:04:05Z","modified_at":"2026-01-03T03:04:05Z",` +
		`"rate_limiting_interval":60,"rate_limiting_limit":100,"rate_limiting_technique":"fixed",` +
		`"authentication":true}`
)

// envelope wraps a result the way the Cloudflare API does.
func envelope(result string) string {
	return fmt.Sprintf(`{"success":true,"errors":[],"messages":[],"result":%s}`, result)
}

// newTestClient returns a Client whose requests are served by the handler.
func newTestClient(t *testing.T, handler http.HandlerFunc) Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		handler(w, r)
	}))
	t.Cleanup(srv.Close)
	return newSDKClient(option.WithAPIToken("test-token"), option.WithBaseURL(srv.URL))
}

func TestGet(t *testing.T) {
	type want struct {
		gw  Gateway
		err bool
	}
	cases := map[string]struct {
		status int
		body   string
		want   want
	}{
		"Success": {
			status: http.StatusOK,
			body:   envelope(gatewayJSON),
			want: want{
				gw: Gateway{
					ID:                      "my-gateway",
					CacheInvalidateOnUpdate: true,
					CacheTTL:                3600,
					CollectLogs:             true,
					RateLimitingInterval:    60,
					RateLimitingLimit:       100,
					RateLimitingTechnique:   "fixed",
					Authentication:          true,
					CreatedAt:               mustTime("2026-01-02T03:04:05Z"),
					ModifiedAt:              mustTime("2026-01-03T03:04:05Z"),
				},
			},
		},
		"NotFound": {
			status: http.StatusNotFound,
			body:   `{"success":false,"errors":[{"code":7000,"message":"No route for that URI"}],"messages":[],"result":null}`,
			want:   want{err: true},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if got, wantPath := r.URL.Path, "/accounts/"+testAccountID+"/ai-gateway/gateways/"+testGatewayID; got != wantPath {
					t.Errorf("path = %s, want %s", got, wantPath)
				}
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			})

			got, err := c.Get(context.Background(), testAccountID, testGatewayID)
			if tc.want.err {
				if err == nil {
					t.Fatalf("Get(...): expected error, got none")
				}
				if !IsNotFound(err) {
					t.Errorf("IsNotFound(%v) = false, want true", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get(...): unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want.gw, got); diff != "" {
				t.Errorf("Get(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	var sent map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		_, _ = io.WriteString(w, envelope(gatewayJSON))
	})

	gw, err := c.Create(context.Background(), testAccountID, v1beta1.GatewayParameters{
		AccountID:               testAccountID,
		GatewayID:               testGatewayID,
		CacheInvalidateOnUpdate: true,
		CacheTTL:                3600,
		CollectLogs:             true,
		RateLimitingInterval:    60,
		RateLimitingLimit:       100,
		RateLimitingTechnique:   ptr.To("fixed"),
		Authentication:          ptr.To(true),
	})
	if err != nil {
		t.Fatalf("Create(...): unexpected error: %v", err)
	}
	if gw.ID != testGatewayID {
		t.Errorf("Create(...) ID = %q, want %q", gw.ID, testGatewayID)
	}

	wantBody := map[string]any{
		"id":                         testGatewayID,
		"cache_invalidate_on_update": true,
		"cache_ttl":                  float64(3600),
		"collect_logs":               true,
		"rate_limiting_interval":     float64(60),
		"rate_limiting_limit":        float64(100),
		"rate_limiting_technique":    "fixed",
		"authentication":             true,
	}
	if diff := cmp.Diff(wantBody, sent); diff != "" {
		t.Errorf("Create(...) request body: -want, +got:\n%s", diff)
	}
}

func TestUpdateOmitsUnsetOptionalFields(t *testing.T) {
	var sent map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&sent); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		_, _ = io.WriteString(w, envelope(gatewayJSON))
	})

	if _, err := c.Update(context.Background(), testAccountID, testGatewayID, v1beta1.GatewayParameters{
		AccountID:            testAccountID,
		GatewayID:            testGatewayID,
		CacheTTL:             3600,
		CollectLogs:          true,
		RateLimitingInterval: 60,
		RateLimitingLimit:    100,
	}); err != nil {
		t.Fatalf("Update(...): unexpected error: %v", err)
	}
	for _, key := range []string{"rate_limiting_technique", "authentication"} {
		if _, ok := sent[key]; ok {
			t.Errorf("Update(...) sent %q although it was not set", key)
		}
	}
}

func TestDelete(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		_, _ = io.WriteString(w, envelope(`{"id":"my-gateway"}`))
	})

	if err := c.Delete(context.Background(), testAccountID, testGatewayID); err != nil {
		t.Errorf("Delete(...): unexpected error: %v", err)
	}
}

func TestNewClient(t *testing.T) {
	if c := NewClient("test-token"); c == nil {
		t.Fatal("NewClient(...) returned nil")
	}
}

func TestIsNotFound(t *testing.T) {
	cases := map[string]struct {
		err  error
		want bool
	}{
		"Nil": {
			err:  nil,
			want: false,
		},
		"NotFoundDirect": {
			err:  &cloudflare.Error{StatusCode: http.StatusNotFound},
			want: true,
		},
		"NotFoundWrapped": {
			err:  errors.Wrap(&cloudflare.Error{StatusCode: http.StatusNotFound}, errGet),
			want: true,
		},
		"ServerError": {
			err:  &cloudflare.Error{StatusCode: http.StatusInternalServerError},
			want: false,
		},
		"OtherError": {
			err:  errors.New("not found"),
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsNotFound(tc.err); got != tc.want {
				t.Errorf("IsNotFound(%v) = %t, want %t", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	base := v1beta1.GatewayParameters{
		AccountID:               testAccountID,
		GatewayID:               testGatewayID,
		CacheInvalidateOnUpdate: true,
		CacheTTL:                3600,
		CollectLogs:             true,
		RateLimitingInterval:    60,
		RateLimitingLimit:       100,
	}
	observed := Gateway{
		ID:                      testGatewayID,
		CacheInvalidateOnUpdate: true,
		CacheTTL:                3600,
		CollectLogs:             true,
		RateLimitingInterval:    60,
		RateLimitingLimit:       100,
		RateLimitingTechnique:   "sliding",
		Authentication:          false,
	}

	cases := map[string]struct {
		reason string
		params v1beta1.GatewayParameters
		obs    Gateway
		want   bool
	}{
		"MatchingRequiredFieldsOnly": {
			reason: "Unset optional fields are not compared",
			params: base,
			obs:    observed,
			want:   true,
		},
		"CacheTTLDiffers": {
			reason: "A changed required field is out of date",
			params: func() v1beta1.GatewayParameters { p := base; p.CacheTTL = 60; return p }(),
			obs:    observed,
			want:   false,
		},
		"TechniqueDiffers": {
			reason: "A set rate limiting technique that differs is out of date",
			params: func() v1beta1.GatewayParameters { p := base; p.RateLimitingTechnique = ptr.To("fixed"); return p }(),
			obs:    observed,
			want:   false,
		},
		"TechniqueMatches": {
			reason: "A set rate limiting technique that matches is up to date",
			params: func() v1beta1.GatewayParameters { p := base; p.RateLimitingTechnique = ptr.To("sliding"); return p }(),
			obs:    observed,
			want:   true,
		},
		"AuthenticationDiffers": {
			reason: "A set authentication flag that differs is out of date",
			params: func() v1beta1.GatewayParameters { p := base; p.Authentication = ptr.To(true); return p }(),
			obs:    observed,
			want:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := IsUpToDate(tc.params, tc.obs); got != tc.want {
				t.Errorf("%s\nIsUpToDate(...) = %t, want %t", tc.reason, got, tc.want)
			}
		})
	}
}

func mustTime(s string) (t time.Time) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}
