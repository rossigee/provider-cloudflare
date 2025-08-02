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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// recordValidator implements DNSRecordValidator interface
type recordValidator struct{}

// NewDNSRecordValidator creates a new DNS record validator
func NewDNSRecordValidator() DNSRecordValidator {
	return &recordValidator{}
}

// ValidateSRVRecord validates SRV record content format
// SRV format: "priority weight port target"
func (rv *recordValidator) ValidateSRVRecord(content string) error {
	if content == "" {
		return errors.New("SRV record content cannot be empty")
	}

	// Use Split with space to preserve empty fields better
	parts := strings.Fields(content)
	if len(parts) != 4 {
		return errors.New("SRV record must have format: priority weight port target")
	}

	// Validate priority (0-65535)
	priority, err := strconv.Atoi(parts[0])
	if err != nil {
		return errors.New("invalid SRV priority")
	}
	if priority < 0 || priority > 65535 {
		return errors.New("SRV priority must be between 0 and 65535")
	}

	// Validate weight (0-65535)
	weight, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.New("invalid SRV weight")
	}
	if weight < 0 || weight > 65535 {
		return errors.New("SRV weight must be between 0 and 65535")
	}

	// Validate port (1-65535)
	port, err := strconv.Atoi(parts[2])
	if err != nil {
		return errors.New("invalid SRV port")
	}
	if port < 1 || port > 65535 {
		return errors.New("SRV port must be between 1 and 65535")
	}

	// Validate target (basic hostname validation)
	target := parts[3]
	if target == "" {
		return errors.New("SRV target cannot be empty")
	}

	// Basic hostname validation
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.?$`)
	if !hostnameRegex.MatchString(target) {
		return errors.New("invalid SRV target hostname")
	}

	return nil
}

// ValidateMXRecord validates MX record priority
func (rv *recordValidator) ValidateMXRecord(content string, priority int) error {
	if content == "" {
		return errors.New("MX record content cannot be empty")
	}

	if priority < 0 || priority > 65535 {
		return errors.New("MX priority must be between 0 and 65535")
	}

	// Basic hostname validation for MX target
	hostnameRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*\.?$`)
	if !hostnameRegex.MatchString(content) {
		return errors.New("invalid MX target hostname")
	}

	return nil
}

// ValidateRecord validates any DNS record based on type
func (rv *recordValidator) ValidateRecord(recordType, content string, priority *int) error {
	switch strings.ToUpper(recordType) {
	case "SRV":
		return rv.ValidateSRVRecord(content)
	case "MX":
		if priority == nil {
			return errors.New("MX record requires priority field")
		}
		return rv.ValidateMXRecord(content, *priority)
	case "URI":
		if priority == nil {
			return errors.New("URI record requires priority field")
		}
		if *priority < 0 || *priority > 65535 {
			return errors.New("URI priority must be between 0 and 65535")
		}
	case "A":
		return rv.validateIPv4(content)
	case "AAAA":
		return rv.validateIPv6(content)
	case "CNAME", "TXT", "NS":
		if content == "" {
			return fmt.Errorf("%s record content cannot be empty", recordType)
		}
	}

	return nil
}

// validateIPv4 validates IPv4 address format
func (rv *recordValidator) validateIPv4(ip string) error {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return errors.New("invalid IPv4 address format")
	}

	for _, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return errors.New("invalid IPv4 address format")
		}
		if num < 0 || num > 255 {
			return errors.New("IPv4 address octets must be between 0 and 255")
		}
	}

	return nil
}

// validateIPv6 validates IPv6 address format (basic)
func (rv *recordValidator) validateIPv6(ip string) error {
	if ip == "" {
		return errors.New("IPv6 address cannot be empty")
	}

	// Basic validation - contains colons and valid hex chars
	ipv6Regex := regexp.MustCompile(`^([0-9a-fA-F]{0,4}:){1,7}[0-9a-fA-F]{0,4}$|^::$|^::1$`)
	if !ipv6Regex.MatchString(ip) {
		return errors.New("invalid IPv6 address format")
	}

	return nil
}
