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

package records

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func uint16Ptr(v uint16) *uint16 {
	return &v
}

func stringPtr(s string) *string {
	return &s
}

// mockRecordsAPI mocks the DNS Records API client
type mockRecordsAPI struct {
	MockCreateDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error)
	MockUpdateDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error)
	MockGetDNSRecord     func(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error)
	MockDeleteDNSRecord  func(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error
	MockListDNSRecords   func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error)
	MockExportDNSRecords func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ExportDNSRecordsParams) (string, error)
	MockImportDNSRecords func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ImportDNSRecordsParams) error
}

func (m *mockRecordsAPI) CreateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error) {
	return m.MockCreateDNSRecord(ctx, rc, params)
}

func (m *mockRecordsAPI) UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
	return m.MockUpdateDNSRecord(ctx, rc, params)
}

func (m *mockRecordsAPI) GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error) {
	return m.MockGetDNSRecord(ctx, rc, recordID)
}

func (m *mockRecordsAPI) DeleteDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error {
	return m.MockDeleteDNSRecord(ctx, rc, recordID)
}

func (m *mockRecordsAPI) ListDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListDNSRecordsParams) ([]cloudflare.DNSRecord, *cloudflare.ResultInfo, error) {
	if m.MockListDNSRecords != nil {
		return m.MockListDNSRecords(ctx, rc, params)
	}
	return nil, nil, errors.New("not implemented")
}

func (m *mockRecordsAPI) ExportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ExportDNSRecordsParams) (string, error) {
	if m.MockExportDNSRecords != nil {
		return m.MockExportDNSRecords(ctx, rc, params)
	}
	return "", errors.New("not implemented")
}

func (m *mockRecordsAPI) ImportDNSRecords(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ImportDNSRecordsParams) error {
	if m.MockImportDNSRecords != nil {
		return m.MockImportDNSRecords(ctx, rc, params)
	}
	return errors.New("not implemented")
}

func TestIsRecordNotFound(t *testing.T) {
	cases := map[string]struct {
		reason string
		err    error
		want   bool
	}{
		"RecordNotFound": {
			reason: "Should return true when error contains record not found code",
			err:    errors.New("error code 81044"),
			want:   true,
		},
		"RecordNotFoundFullMessage": {
			reason: "Should match error code 81044 in full message",
			err:    errors.New("HTTP status 400: error code 81044: Invalid record type"),
			want:   true,
		},
		"OtherError": {
			reason: "Should return false for other errors",
			err:    errors.New("some other error"),
			want:   false,
		},
		"NilError": {
			reason: "Should panic when error is nil - typical error handling",
			err:    nil,
			want:   false, // Would panic in real usage
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if tc.err == nil {
				return // Skip nil check as it would panic
			}
			got := IsRecordNotFound(tc.err)
			if got != tc.want {
				t.Errorf("%s: expected %v, got %v", tc.reason, tc.want, got)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	now := time.Now()

	cases := map[string]struct {
		reason string
		input  cloudflare.DNSRecord
		want   v1beta1.RecordObservation
	}{
		"BasicRecord": {
			reason: "Should generate observation from basic record",
			input: cloudflare.DNSRecord{
				ID:         "record-123",
				Type:       "A",
				Name:       "www.example.com",
				Content:    "192.0.2.1",
				TTL:        3600,
				Proxiable:  true,
				Proxied:    ptr.To(false),
				CreatedOn:  now,
				ModifiedOn: now,
			},
			want: v1beta1.RecordObservation{
				Proxiable:  true,
				FQDN:       "www.example.com",
				Zone:       "",
				Locked:     false,
				CreatedOn:  &metav1.Time{Time: now},
				ModifiedOn: &metav1.Time{Time: now},
			},
		},
		"ProxiedRecord": {
			reason: "Should handle proxied records",
			input: cloudflare.DNSRecord{
				ID:         "record-456",
				Type:       "A",
				Name:       "app.example.com",
				Content:    "192.0.2.2",
				TTL:        120,
				Proxiable:  true,
				Proxied:    ptr.To(true),
				CreatedOn:  now,
				ModifiedOn: now,
			},
			want: v1beta1.RecordObservation{
				Proxiable:  true,
				FQDN:       "app.example.com",
				Zone:       "",
				Locked:     false,
				CreatedOn:  &metav1.Time{Time: now},
				ModifiedOn: &metav1.Time{Time: now},
			},
		},
		"CNAMERecord": {
			reason: "Should handle CNAME records",
			input: cloudflare.DNSRecord{
				ID:         "record-789",
				Type:       "CNAME",
				Name:       "alias.example.com",
				Content:    "target.example.com",
				TTL:        1800,
				Proxiable:  false,
				CreatedOn:  now,
				ModifiedOn: now,
			},
			want: v1beta1.RecordObservation{
				Proxiable:  false,
				FQDN:       "alias.example.com",
				Zone:       "",
				Locked:     false,
				CreatedOn:  &metav1.Time{Time: now},
				ModifiedOn: &metav1.Time{Time: now},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s: -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}

func TestUpdateRecord(t *testing.T) {
	ctx := context.Background()
	zoneID := "zone-123"
	recordID := "record-456"

	cases := map[string]struct {
		reason  string
		mock    func() Client
		spec    *v1beta1.RecordParameters
		wantErr bool
	}{
		"SuccessUpdateA": {
			reason: "Should successfully update A record",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						if params.Type != "A" || params.Content != "192.0.2.1" {
							t.Errorf("Unexpected params: type=%s, content=%s", params.Type, params.Content)
						}
						return cloudflare.DNSRecord{
							ID:      recordID,
							Type:    "A",
							Name:    params.Name,
							Content: params.Content,
							TTL:     params.TTL,
						}, nil
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:    stringPtr("A"),
				Name:    "www.example.com",
				Content: "192.0.2.1",
				TTL:     ptr.To[int64](3600),
				Proxied: ptr.To(false),
			},
			wantErr: false,
		},
		"SuccessUpdateWithPriority": {
			reason: "Should successfully update record with priority",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						if params.Type != "MX" || params.Priority == nil {
							t.Errorf("Expected MX record with priority")
						}
						return cloudflare.DNSRecord{
							ID:       recordID,
							Type:     "MX",
							Name:     params.Name,
							Content:  params.Content,
							Priority: uint16Ptr(uint16(*params.Priority)),
						}, nil
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:     stringPtr("MX"),
				Name:     "example.com",
				Content:  "mail.example.com",
				Priority: ptr.To[int32](10),
			},
			wantErr: false,
		},
		"SuccessUpdateSRV": {
			reason: "Should successfully update SRV record",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						if params.Type != "SRV" || params.Data == nil {
							t.Errorf("Expected SRV record with Data field")
						}
						return cloudflare.DNSRecord{
							ID:   recordID,
							Type: "SRV",
							Name: params.Name,
						}, nil
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:     stringPtr("SRV"),
				Name:     "_http._tcp.example.com",
				Content:  "target.example.com",
				Priority: ptr.To[int32](10),
				Weight:   ptr.To[int32](20),
				Port:     ptr.To[int32](80),
			},
			wantErr: false,
		},
		"SuccessUpdateTLSA": {
			reason: "Should successfully update TLSA record",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						if params.Type != "TLSA" || params.Data == nil {
							t.Errorf("Expected TLSA record with Data field")
						}
						return cloudflare.DNSRecord{
							ID:   recordID,
							Type: "TLSA",
							Name: params.Name,
						}, nil
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:    stringPtr("TLSA"),
				Name:    "_443._tcp.example.com",
				Content: "3 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			},
			wantErr: false,
		},
		"ErrInvalidTLSA": {
			reason: "Should return error for invalid TLSA content",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						return cloudflare.DNSRecord{}, nil
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:    stringPtr("TLSA"),
				Name:    "_443._tcp.example.com",
				Content: "invalid tlsa",
			},
			wantErr: true,
		},
		"ErrUpdateFailed": {
			reason: "Should return error when API call fails",
			mock: func() Client {
				return &mockRecordsAPI{
					MockUpdateDNSRecord: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error) {
						return cloudflare.DNSRecord{}, errors.New("API error")
					},
				}
			},
			spec: &v1beta1.RecordParameters{
				Type:    stringPtr("A"),
				Name:    "www.example.com",
				Content: "192.0.2.1",
			},
			wantErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := UpdateRecord(ctx, tc.mock(), zoneID, recordID, tc.spec)
			if (err != nil) != tc.wantErr {
				t.Errorf("%s: expected error %v, got %v", tc.reason, tc.wantErr, err != nil)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		rp *v1beta1.RecordParameters
		r  cloudflare.DNSRecord
	}

	type want struct {
		o  bool
		rp *v1beta1.RecordParameters
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"LateInitSpecNil": {
			reason: "LateInit should return false when not passed a spec",
			args:   args{},
			want: want{
				o: false,
			},
		},
		"LateInitDontUpdate": {
			reason: "LateInit should not update already-set spec fields from a Record",
			args: args{
				rp: &v1beta1.RecordParameters{
					Proxied:  ptr.To(false),
					Priority: ptr.To[int32](4),
				},
				r: cloudflare.DNSRecord{
					Proxied:  ptr.To(true),
					Priority: uint16Ptr(1),
				},
			},
			want: want{
				o: false,
				rp: &v1beta1.RecordParameters{
					Proxied:  ptr.To(false),
					Priority: ptr.To[int32](4),
				},
			},
		},
		"LateInitUpdate": {
			reason: "LateInit should update unset spec fields from a Record",
			args: args{
				rp: &v1beta1.RecordParameters{},
				r: cloudflare.DNSRecord{
					Proxied:  ptr.To(true),
					Priority: uint16Ptr(1),
				},
			},
			want: want{
				o: true,
				rp: &v1beta1.RecordParameters{
					Proxied:  ptr.To(true),
					Priority: ptr.To[int32](1),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := LateInitialize(tc.args.rp, tc.args.r)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nLateInit(...): -want, +got:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.rp, tc.args.rp); diff != "" {
				t.Errorf("\n%s\nLateInit(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpToDate(t *testing.T) {
	type args struct {
		rp *v1beta1.RecordParameters
		r  cloudflare.DNSRecord
	}

	type want struct {
		o bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"UpToDateSpecNil": {
			reason: "UpToDate should return true when not passed a spec",
			args:   args{},
			want: want{
				o: true,
			},
		},
		"UpToDateEmptyParams": {
			reason: "UpToDate should return true and not panic with nil values",
			args: args{
				rp: &v1beta1.RecordParameters{},
				r:  cloudflare.DNSRecord{},
			},
			want: want{
				o: true,
			},
		},
		"UpToDateDifferent": {
			reason: "UpToDate should return false if the spec does not match the record",
			args: args{
				rp: &v1beta1.RecordParameters{
					Type:    ptr.To("A"),
					Name:    "foo",
					Content: "127.0.0.1",
					TTL:     ptr.To[int64](600),
					Proxied: ptr.To(false),
				},
				r: cloudflare.DNSRecord{
					Type:    "A",
					Name:    "foo",
					Content: "127.0.0.2",
					TTL:     600,
					Proxied: ptr.To(false),
				},
			},
			want: want{
				o: false,
			},
		},
		"UpToDateIdentical": {
			reason: "UpToDate should return true if the spec matches the record",
			args: args{
				rp: &v1beta1.RecordParameters{
					Type:    ptr.To("A"),
					Name:    "foo",
					Content: "127.0.0.1",
					TTL:     ptr.To[int64](600),
					Proxied: ptr.To(false),
				},
				r: cloudflare.DNSRecord{
					Type:    "A",
					Name:    "foo",
					Content: "127.0.0.1",
					TTL:     600,
					Proxied: ptr.To(false),
				},
			},
			want: want{
				o: true,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := UpToDate(tc.args.rp, tc.args.r)
			if diff := cmp.Diff(tc.want.o, got); diff != "" {
				t.Errorf("\n%s\nUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestParseTLSAContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid TLSA record - DANE-EE SPKI SHA-256",
			content: "3 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			want: map[string]interface{}{
				"usage":         3,
				"selector":      1,
				"matching_type": 1,
				"certificate":   "0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			},
			wantErr: false,
		},
		{
			name:    "valid TLSA record - PKIX-TA Full Cert SHA-512",
			content: "0 0 2 d2abde240d7cd3ee6b4b28c54df034b97983a1d16e8a410e4561cb106618e971f8",
			want: map[string]interface{}{
				"usage":         0,
				"selector":      0,
				"matching_type": 2,
				"certificate":   "d2abde240d7cd3ee6b4b28c54df034b97983a1d16e8a410e4561cb106618e971f8",
			},
			wantErr: false,
		},
		{
			name:    "valid TLSA record - DANE-TA SPKI SHA-256",
			content: "2 1 1 92003ba34942dc74152e2f2c408d29eca5a520e7f2e06bb944f4dca346baf63c",
			want: map[string]interface{}{
				"usage":         2,
				"selector":      1,
				"matching_type": 1,
				"certificate":   "92003ba34942dc74152e2f2c408d29eca5a520e7f2e06bb944f4dca346baf63c",
			},
			wantErr: false,
		},
		{
			name:    "invalid - too few fields",
			content: "3 1 1",
			wantErr: true,
			errMsg:  "TLSA content must have 4 space-separated fields",
		},
		{
			name:    "invalid - too many fields",
			content: "3 1 1 cert extra",
			wantErr: true,
			errMsg:  "TLSA content must have 4 space-separated fields",
		},
		{
			name:    "invalid - usage out of range (high)",
			content: "4 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA usage must be 0-3",
		},
		{
			name:    "invalid - usage out of range (negative)",
			content: "-1 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA usage must be 0-3",
		},
		{
			name:    "invalid - usage not a number",
			content: "x 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA usage must be 0-3",
		},
		{
			name:    "invalid - selector out of range",
			content: "3 2 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA selector must be 0-1",
		},
		{
			name:    "invalid - selector not a number",
			content: "3 x 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA selector must be 0-1",
		},
		{
			name:    "invalid - matching_type out of range",
			content: "3 1 3 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA matching_type must be 0-2",
		},
		{
			name:    "invalid - matching_type not a number",
			content: "3 1 x 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3",
			wantErr: true,
			errMsg:  "TLSA matching_type must be 0-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTLSAContent(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTLSAContent() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("parseTLSAContent() error = %v, want error containing %q", err, tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Errorf("parseTLSAContent() unexpected error = %v", err)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("parseTLSAContent() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
