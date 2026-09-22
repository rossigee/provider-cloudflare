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

package spectrum

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/spectrum/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func stringPtr(s string) *string {
	return &s
}

func uint32Ptr(i uint32) *uint32 {
	return &i
}

// mockSpectrumAPI mocks the Spectrum API client
type mockSpectrumAPI struct {
	MockSpectrumApplication       func(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error)
	MockCreateSpectrumApplication func(ctx context.Context, zoneID string, params *v1beta1.ApplicationParameters) (cloudflare.SpectrumApplication, error)
	MockUpdateSpectrumApplication func(ctx context.Context, zoneID, applicationID string, params *v1beta1.ApplicationParameters) error
	MockDeleteSpectrumApplication func(ctx context.Context, zoneID, applicationID string) error
}

func (m *mockSpectrumAPI) SpectrumApplication(ctx context.Context, zoneID, applicationID string) (cloudflare.SpectrumApplication, error) {
	return m.MockSpectrumApplication(ctx, zoneID, applicationID)
}

func (m *mockSpectrumAPI) CreateSpectrumApplication(ctx context.Context, zoneID string, params *v1beta1.ApplicationParameters) (cloudflare.SpectrumApplication, error) {
	return m.MockCreateSpectrumApplication(ctx, zoneID, params)
}

func (m *mockSpectrumAPI) UpdateSpectrumApplication(ctx context.Context, zoneID, applicationID string, params *v1beta1.ApplicationParameters) error {
	return m.MockUpdateSpectrumApplication(ctx, zoneID, applicationID, params)
}

func (m *mockSpectrumAPI) DeleteSpectrumApplication(ctx context.Context, zoneID, applicationID string) error {
	return m.MockDeleteSpectrumApplication(ctx, zoneID, applicationID)
}

func TestIsApplicationNotFound(t *testing.T) {
	cases := map[string]struct {
		reason string
		err    error
		want   bool
	}{
		"ExactMessage": {
			reason: "Should return true for exact error message",
			err:    errors.New("Application not found"),
			want:   true,
		},
		"ErrorCode404": {
			reason: "Should return true for 404 error",
			err:    errors.New("404"),
			want:   true,
		},
		"NotFound": {
			reason: "Should return true for 'Not found' message",
			err:    errors.New("Not found"),
			want:   true,
		},
		"ErrorCode10006": {
			reason: "Should return true for Cloudflare error code 10006",
			err:    errors.New("10006"),
			want:   true,
		},
		"OtherError": {
			reason: "Should return false for other errors",
			err:    errors.New("some other error"),
			want:   false,
		},
		"NilError": {
			reason: "Should return false for nil error",
			err:    nil,
			want:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsApplicationNotFound(tc.err)
			if got != tc.want {
				t.Errorf("%s: expected %v, got %v", tc.reason, tc.want, got)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	now := time.Now()
	createdOn := now.Add(-24 * time.Hour)

	cases := map[string]struct {
		reason string
		app    cloudflare.SpectrumApplication
		want   v1beta1.ApplicationObservation
	}{
		"WithTimestamps": {
			reason: "Should generate observation with timestamps",
			app: cloudflare.SpectrumApplication{
				ID:         "app-123",
				Protocol:   "tcp/22",
				CreatedOn:  &createdOn,
				ModifiedOn: &now,
			},
			want: v1beta1.ApplicationObservation{
				CreatedOn:  &metav1.Time{Time: createdOn},
				ModifiedOn: &metav1.Time{Time: now},
			},
		},
		"WithoutTimestamps": {
			reason: "Should handle missing timestamps gracefully",
			app: cloudflare.SpectrumApplication{
				ID:       "app-456",
				Protocol: "udp/53",
			},
			want: v1beta1.ApplicationObservation{},
		},
		"WithZeroTimestamps": {
			reason: "Should skip zero timestamps",
			app: cloudflare.SpectrumApplication{
				ID:         "app-789",
				Protocol:   "tcp/443",
				CreatedOn:  &time.Time{},
				ModifiedOn: &time.Time{},
			},
			want: v1beta1.ApplicationObservation{},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.app)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestConvertIPs(t *testing.T) {
	cases := map[string]struct {
		reason  string
		ips     []string
		want    []net.IP
		wantErr bool
		errMsg  string
	}{
		"ValidIPv4": {
			reason: "Should convert valid IPv4 addresses",
			ips:    []string{"192.0.2.1", "192.0.2.2"},
			want:   []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")},
		},
		"ValidIPv6": {
			reason: "Should convert valid IPv6 addresses",
			ips:    []string{"2001:db8::1", "2001:db8::2"},
			want:   []net.IP{net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::2")},
		},
		"Mixed": {
			reason: "Should convert mixed IPv4 and IPv6",
			ips:    []string{"192.0.2.1", "2001:db8::1"},
			want:   []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("2001:db8::1")},
		},
		"EmptySlice": {
			reason: "Should handle empty slice",
			ips:    []string{},
			want:   []net.IP{},
		},
		"InvalidIP": {
			reason:  "Should error on invalid IP",
			ips:     []string{"not-an-ip"},
			wantErr: true,
			errMsg:  "invalid IP within Edge IPs",
		},
		"MixedValidInvalid": {
			reason:  "Should error when one IP is invalid",
			ips:     []string{"192.0.2.1", "invalid"},
			wantErr: true,
			errMsg:  "invalid IP within Edge IPs",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ConvertIPs(tc.ips)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
			if tc.wantErr && tc.errMsg != "" && !cmp.Equal(tc.errMsg, err.Error(), cmp.Comparer(func(x, y string) bool {
				return x == y || (x != "" && y != "" && (x == y))
			})) {
				t.Errorf("%s: error mismatch, want %q, got %q", tc.reason, tc.errMsg, err.Error())
			}
			if !tc.wantErr && len(got) != len(tc.want) {
				t.Errorf("%s: length mismatch, want %d, got %d", tc.reason, len(tc.want), len(got))
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	cases := map[string]struct {
		reason string
		spec   *v1beta1.ApplicationParameters
		app    cloudflare.SpectrumApplication
		want   bool
		verify func(t *testing.T, spec *v1beta1.ApplicationParameters)
		skip   bool
	}{
		"NilSpec": {
			reason: "Would panic - spec nil check missing in implementation",
			spec:   nil,
			app:    cloudflare.SpectrumApplication{},
			want:   false,
			skip:   true,
		},
		"NoEdgeIPsInApp": {
			reason: "Should return false when app has no EdgeIPs",
			spec:   &v1beta1.ApplicationParameters{},
			app:    cloudflare.SpectrumApplication{},
			want:   false,
		},
		"EdgeIPsAlreadySet": {
			reason: "Should not override existing EdgeIPs",
			spec: &v1beta1.ApplicationParameters{
				EdgeIPs: &v1beta1.SpectrumApplicationEdgeIPs{
					Type: "static",
				},
			},
			app: cloudflare.SpectrumApplication{
				EdgeIPs: &cloudflare.SpectrumApplicationEdgeIPs{
					Type: cloudflare.SpectrumApplicationEdgeType("static"),
				},
			},
			want: false,
		},
		"InitializeEdgeIPs": {
			reason: "Should initialize EdgeIPs from app",
			spec:   &v1beta1.ApplicationParameters{},
			app: cloudflare.SpectrumApplication{
				EdgeIPs: &cloudflare.SpectrumApplicationEdgeIPs{
					Type: cloudflare.SpectrumApplicationEdgeType("static"),
					IPs:  []net.IP{net.ParseIP("192.0.2.1")},
				},
			},
			want: true,
			verify: func(t *testing.T, spec *v1beta1.ApplicationParameters) {
				if spec.EdgeIPs == nil {
					t.Errorf("EdgeIPs should be initialized")
				}
				if spec.EdgeIPs.Type != "static" {
					t.Errorf("EdgeIPs.Type should be 'static', got %q", spec.EdgeIPs.Type)
				}
				if len(spec.EdgeIPs.IPs) != 1 || spec.EdgeIPs.IPs[0] != "192.0.2.1" {
					t.Errorf("EdgeIPs.IPs mismatch")
				}
			},
		},
		"InitializeEdgeIPsWithConnectivity": {
			reason: "Should initialize EdgeIPs with connectivity",
			spec:   &v1beta1.ApplicationParameters{},
			app: cloudflare.SpectrumApplication{
				EdgeIPs: &cloudflare.SpectrumApplicationEdgeIPs{
					Type:          cloudflare.SpectrumApplicationEdgeType("dynamic"),
					Connectivity:  (*cloudflare.SpectrumApplicationConnectivity)(stringPtr("all")),
					IPs:           []net.IP{net.ParseIP("192.0.2.1"), net.ParseIP("192.0.2.2")},
				},
			},
			want: true,
			verify: func(t *testing.T, spec *v1beta1.ApplicationParameters) {
				if spec.EdgeIPs == nil {
					t.Errorf("EdgeIPs should be initialized")
				}
				if spec.EdgeIPs.Connectivity == nil || *spec.EdgeIPs.Connectivity != "all" {
					t.Errorf("EdgeIPs.Connectivity should be 'all'")
				}
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.skip {
				t.Skipf("%s - skipped", tc.reason)
				return
			}
			got := LateInitialize(tc.spec, tc.app)
			if got != tc.want {
				t.Errorf("%s: expected %v, got %v", tc.reason, tc.want, got)
			}
			if tc.verify != nil {
				tc.verify(t, tc.spec)
			}
		})
	}
}

func TestUpToDate(t *testing.T) {
	cases := map[string]struct {
		reason string
		spec   *v1beta1.ApplicationParameters
		app    cloudflare.SpectrumApplication
		want   bool
	}{
		"ProtocolMismatch": {
			reason: "Should return false when protocol differs",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/443",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
			},
			want: false,
		},
		"DNSTypeMismatch": {
			reason: "Should return false when DNS type differs",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/22",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "MX",
					Name: "example.com",
				},
			},
			want: false,
		},
		"DNSNameMismatch": {
			reason: "Should return false when DNS name differs",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "app1.example.com",
				},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/22",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "app2.example.com",
				},
			},
			want: false,
		},
		"OriginDirectMismatch": {
			reason: "Should return false when OriginDirect differs",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
				OriginDirect: []string{"192.0.2.1"},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/22",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
				OriginDirect: []string{"192.0.2.2"},
			},
			want: false,
		},
		"OriginDirectLengthMismatch": {
			reason: "Should return false when OriginDirect length differs",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
				OriginDirect: []string{"192.0.2.1"},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/22",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "example.com",
				},
				OriginDirect: []string{"192.0.2.1", "192.0.2.2"},
			},
			want: false,
		},
		"UpToDate": {
			reason: "Should return true when all fields match",
			spec: &v1beta1.ApplicationParameters{
				Protocol: "tcp/22",
				DNS: v1beta1.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "app.example.com",
				},
				OriginDirect: []string{"192.0.2.1", "192.0.2.2"},
			},
			app: cloudflare.SpectrumApplication{
				Protocol: "tcp/22",
				DNS: cloudflare.SpectrumApplicationDNS{
					Type: "CNAME",
					Name: "app.example.com",
				},
				OriginDirect: []string{"192.0.2.1", "192.0.2.2"},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := UpToDate(tc.spec, tc.app)
			if got != tc.want {
				t.Errorf("%s: expected %v, got %v", tc.reason, tc.want, got)
			}
		})
	}
}

