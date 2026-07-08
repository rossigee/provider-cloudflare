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
	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	"k8s.io/utils/ptr"
	"strings"
	"testing"
)

func uint16Ptr(v uint16) *uint16 {
	return &v
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
