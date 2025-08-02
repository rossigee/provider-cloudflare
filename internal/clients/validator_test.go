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

package clients

import (
	"testing"

	"github.com/pkg/errors"
)

func TestValidateSRVRecord(t *testing.T) {
	validator := NewDNSRecordValidator()

	type args struct {
		content string
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ValidSRVRecord": {
			reason: "Valid SRV record should pass validation",
			args: args{
				content: "10 20 8080 target.example.com",
			},
			want: want{
				err: nil,
			},
		},
		"ValidSRVRecordWithDot": {
			reason: "Valid SRV record with trailing dot should pass",
			args: args{
				content: "5 10 443 web.example.com.",
			},
			want: want{
				err: nil,
			},
		},
		"ValidSRVRecordMinimumPort": {
			reason: "SRV record with port 1 should pass",
			args: args{
				content: "0 0 1 ssh.example.com",
			},
			want: want{
				err: nil,
			},
		},
		"ValidSRVRecordMaximumValues": {
			reason: "SRV record with maximum values should pass",
			args: args{
				content: "65535 65535 65535 max.example.com",
			},
			want: want{
				err: nil,
			},
		},
		"EmptyContent": {
			reason: "Empty content should fail validation",
			args: args{
				content: "",
			},
			want: want{
				err: errors.New("SRV record content cannot be empty"),
			},
		},
		"InvalidFormat": {
			reason: "Invalid format should fail validation",
			args: args{
				content: "10 20 target.example.com", // missing port
			},
			want: want{
				err: errors.New("SRV record must have format: priority weight port target"),
			},
		},
		"InvalidPriority": {
			reason: "Invalid priority should fail validation",
			args: args{
				content: "invalid 20 8080 target.example.com",
			},
			want: want{
				err: errors.New("invalid SRV priority"),
			},
		},
		"PriorityOutOfRange": {
			reason: "Priority out of range should fail validation",
			args: args{
				content: "70000 20 8080 target.example.com",
			},
			want: want{
				err: errors.New("SRV priority must be between 0 and 65535"),
			},
		},
		"InvalidWeight": {
			reason: "Invalid weight should fail validation",
			args: args{
				content: "10 invalid 8080 target.example.com",
			},
			want: want{
				err: errors.New("invalid SRV weight"),
			},
		},
		"WeightOutOfRange": {
			reason: "Weight out of range should fail validation",
			args: args{
				content: "10 70000 8080 target.example.com",
			},
			want: want{
				err: errors.New("SRV weight must be between 0 and 65535"),
			},
		},
		"InvalidPort": {
			reason: "Invalid port should fail validation",
			args: args{
				content: "10 20 invalid target.example.com",
			},
			want: want{
				err: errors.New("invalid SRV port"),
			},
		},
		"PortZero": {
			reason: "Port 0 should fail validation",
			args: args{
				content: "10 20 0 target.example.com",
			},
			want: want{
				err: errors.New("SRV port must be between 1 and 65535"),
			},
		},
		"PortOutOfRange": {
			reason: "Port out of range should fail validation",
			args: args{
				content: "10 20 70000 target.example.com",
			},
			want: want{
				err: errors.New("SRV port must be between 1 and 65535"),
			},
		},
		"EmptyTarget": {
			reason: "Empty target should fail validation",
			args: args{
				content: "10 20 8080",
			},
			want: want{
				err: errors.New("SRV record must have format: priority weight port target"),
			},
		},
		"InvalidHostname": {
			reason: "Invalid hostname should fail validation",
			args: args{
				content: "10 20 8080 invalid..hostname",
			},
			want: want{
				err: errors.New("invalid SRV target hostname"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := validator.ValidateSRVRecord(tc.args.content)

			if tc.want.err == nil {
				if err != nil {
					t.Errorf("\n%s\nValidateSRVRecord(%s): expected no error, got %v", tc.reason, tc.args.content, err)
				}
			} else {
				if err == nil {
					t.Errorf("\n%s\nValidateSRVRecord(%s): expected error %v, got nil", tc.reason, tc.args.content, tc.want.err)
				} else if err.Error() != tc.want.err.Error() {
					t.Errorf("\n%s\nValidateSRVRecord(%s): expected error %v, got %v", tc.reason, tc.args.content, tc.want.err, err)
				}
			}
		})
	}
}

func TestValidateRecord(t *testing.T) {
	validator := NewDNSRecordValidator()

	type args struct {
		recordType string
		content    string
		priority   *int
	}

	type want struct {
		err error
	}

	priority10 := 10
	priority70000 := 70000

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ValidSRVRecord": {
			reason: "Valid SRV record should pass",
			args: args{
				recordType: "SRV",
				content:    "10 20 8080 target.example.com",
				priority:   nil,
			},
			want: want{
				err: nil,
			},
		},
		"ValidMXRecord": {
			reason: "Valid MX record should pass",
			args: args{
				recordType: "MX",
				content:    "mail.example.com",
				priority:   &priority10,
			},
			want: want{
				err: nil,
			},
		},
		"MXRecordNoPriority": {
			reason: "MX record without priority should fail",
			args: args{
				recordType: "MX",
				content:    "mail.example.com",
				priority:   nil,
			},
			want: want{
				err: errors.New("MX record requires priority field"),
			},
		},
		"MXRecordInvalidPriority": {
			reason: "MX record with invalid priority should fail",
			args: args{
				recordType: "MX",
				content:    "mail.example.com",
				priority:   &priority70000,
			},
			want: want{
				err: errors.New("MX priority must be between 0 and 65535"),
			},
		},
		"ValidARecord": {
			reason: "Valid A record should pass",
			args: args{
				recordType: "A",
				content:    "192.168.1.1",
				priority:   nil,
			},
			want: want{
				err: nil,
			},
		},
		"InvalidARecord": {
			reason: "Invalid A record should fail",
			args: args{
				recordType: "A",
				content:    "256.1.1.1",
				priority:   nil,
			},
			want: want{
				err: errors.New("IPv4 address octets must be between 0 and 255"),
			},
		},
		"ValidAAAARecord": {
			reason: "Valid AAAA record should pass",
			args: args{
				recordType: "AAAA",
				content:    "2001:db8::1",
				priority:   nil,
			},
			want: want{
				err: nil,
			},
		},
		"ValidCNAMERecord": {
			reason: "Valid CNAME record should pass",
			args: args{
				recordType: "CNAME",
				content:    "target.example.com",
				priority:   nil,
			},
			want: want{
				err: nil,
			},
		},
		"EmptyCNAMERecord": {
			reason: "Empty CNAME record should fail",
			args: args{
				recordType: "CNAME",
				content:    "",
				priority:   nil,
			},
			want: want{
				err: errors.New("CNAME record content cannot be empty"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := validator.ValidateRecord(tc.args.recordType, tc.args.content, tc.args.priority)

			if tc.want.err == nil {
				if err != nil {
					t.Errorf("\n%s\nValidateRecord(%s, %s): expected no error, got %v", tc.reason, tc.args.recordType, tc.args.content, err)
				}
			} else {
				if err == nil {
					t.Errorf("\n%s\nValidateRecord(%s, %s): expected error %v, got nil", tc.reason, tc.args.recordType, tc.args.content, tc.want.err)
				} else if err.Error() != tc.want.err.Error() {
					t.Errorf("\n%s\nValidateRecord(%s, %s): expected error %v, got %v", tc.reason, tc.args.recordType, tc.args.content, tc.want.err, err)
				}
			}
		})
	}
}

func TestValidateIPv4(t *testing.T) {
	validator := &recordValidator{}

	cases := map[string]struct {
		ip  string
		err bool
	}{
		"ValidIP":       {"192.168.1.1", false},
		"ValidLoopback": {"127.0.0.1", false},
		"ValidZero":     {"0.0.0.0", false},
		"ValidMax":      {"255.255.255.255", false},
		"InvalidOctet":  {"256.1.1.1", true},
		"TooFewOctets":  {"192.168.1", true},
		"TooManyOctets": {"192.168.1.1.1", true},
		"NonNumeric":    {"192.168.a.1", true},
		"Empty":         {"", true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := validator.validateIPv4(tc.ip)
			if tc.err && err == nil {
				t.Errorf("Expected error for IP %s, got nil", tc.ip)
			}
			if !tc.err && err != nil {
				t.Errorf("Expected no error for IP %s, got %v", tc.ip, err)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkValidateSRVRecord(b *testing.B) {
	validator := NewDNSRecordValidator()
	testContent := "10 20 8080 target.example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateSRVRecord(testContent)
	}
}

func BenchmarkValidateRecord(b *testing.B) {
	validator := NewDNSRecordValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateRecord("SRV", "10 20 8080 target.example.com", nil)
		_ = validator.ValidateRecord("A", "192.168.1.1", nil)
		priority := 10
		_ = validator.ValidateRecord("MX", "mail.example.com", &priority)
	}
}
