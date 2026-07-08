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
	"fmt"
	"github.com/cloudflare/cloudflare-go"
	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
	"strings"
)

const (
	// Cloudflare returns this code when a record isnt found.
	errRecordNotFound = "81044"
)

// Client is a Cloudflare API client that implements methods for working
// with DNS Records.
type Client interface {
	CreateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateDNSRecordParams) (cloudflare.DNSRecord, error)
	UpdateDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateDNSRecordParams) (cloudflare.DNSRecord, error)
	GetDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) (cloudflare.DNSRecord, error)
	DeleteDNSRecord(ctx context.Context, rc *cloudflare.ResourceContainer, recordID string) error
}

// NewClient returns a new Cloudflare API client for working with DNS Records.
func NewClient(cfg clients.Config, hc *http.Client) (Client, error) {
	return clients.NewClient(cfg, hc)
}

// IsRecordNotFound returns true if the passed error indicates
// a Record was not found.
func IsRecordNotFound(err error) bool {
	return strings.Contains(err.Error(), errRecordNotFound)
}

// GenerateObservation creates an observation of a cloudflare Record.
func GenerateObservation(in cloudflare.DNSRecord) v1beta1.RecordObservation {
	return v1beta1.RecordObservation{
		Proxiable:  in.Proxiable,
		FQDN:       in.Name,
		Zone:       "",    // Zone name not available in new API response
		Locked:     false, // Locked field not available in new API response
		CreatedOn:  &metav1.Time{Time: in.CreatedOn},
		ModifiedOn: &metav1.Time{Time: in.ModifiedOn},
	}
}

// LateInitialize initializes RecordParameters based on the remote resource.
func LateInitialize(spec *v1beta1.RecordParameters, o cloudflare.DNSRecord) bool {
	if spec == nil {
		return false
	}

	li := false
	if spec.Proxied == nil && o.Proxied != nil {
		spec.Proxied = o.Proxied
		li = true
	}

	if spec.Priority == nil && o.Priority != nil {
		pri := int32(*o.Priority)
		spec.Priority = &pri
		li = true
	}

	return li
}

// UpToDate checks if the remote Record is up to date with the
// requested resource parameters.
func UpToDate(spec *v1beta1.RecordParameters, o cloudflare.DNSRecord) bool { //nolint:gocyclo
	// NOTE(bagricola): The complexity here is simply repeated
	// if statements checking for updated fields. You should think
	// before adding further complexity to this method, but adding
	// more field checks should not be an issue.
	if spec == nil {
		return true
	}

	// Check if mutable fields are up to date with resource

	// Compare names directly - new API handles zone naming differently
	fn := spec.Name

	if fn != o.Name {
		return false
	}

	if spec.Content != o.Content {
		return false
	}

	if spec.TTL != nil && *spec.TTL != int64(o.TTL) {
		return false
	}

	if spec.Proxied != nil && o.Proxied != nil && *spec.Proxied != *o.Proxied {
		return false
	}

	if spec.Priority != nil && o.Priority != nil && *spec.Priority != int32(*o.Priority) {
		return false
	}

	return true
}

// UpdateRecord updates mutable values on a DNS Record.
func UpdateRecord(ctx context.Context, client Client, zoneID, recordID string, spec *v1beta1.RecordParameters) error {
	rc := cloudflare.ZoneIdentifier(zoneID)

	params := cloudflare.UpdateDNSRecordParams{
		ID:      recordID,
		Type:    *spec.Type,
		Name:    spec.Name,
		Content: spec.Content,
	}

	if spec.TTL != nil {
		params.TTL = int(*spec.TTL)
	}

	if spec.Proxied != nil {
		params.Proxied = spec.Proxied
	}

	if spec.Priority != nil {
		priority := uint16(*spec.Priority)
		params.Priority = &priority
	}

	// For SRV records, use the Data field
	if *spec.Type == "SRV" && spec.Priority != nil && spec.Weight != nil && spec.Port != nil {
		srvData := map[string]interface{}{
			"priority": int(*spec.Priority),
			"weight":   int(*spec.Weight),
			"port":     int(*spec.Port),
			"target":   spec.Content,
		}
		params.Data = srvData
		params.Priority = nil
		params.Content = ""
	}

	// For TLSA records, parse content and use Data field
	if *spec.Type == "TLSA" {
		tlsaData, err := parseTLSAContent(spec.Content)
		if err != nil {
			return err
		}
		params.Data = tlsaData
		params.Content = ""
	}

	_, err := client.UpdateDNSRecord(ctx, rc, params)
	return err
}

// parseTLSAContent parses a TLSA content string into CloudFlare API format.
// Input format: "usage selector matching_type certificate"
// Example: "3 1 1 0b9fa5a59eed715c26c1020c711b4f6ec42d58b0015e14337a39dad301c5afc3"
func parseTLSAContent(content string) (map[string]interface{}, error) {
	parts := strings.Fields(content)
	if len(parts) != 4 {
		return nil, fmt.Errorf("TLSA content must have 4 space-separated fields (usage selector matching_type certificate), got %d", len(parts))
	}

	usage, err := strconv.Atoi(parts[0])
	if err != nil || usage < 0 || usage > 3 {
		return nil, fmt.Errorf("TLSA usage must be 0-3, got: %s", parts[0])
	}

	selector, err := strconv.Atoi(parts[1])
	if err != nil || selector < 0 || selector > 1 {
		return nil, fmt.Errorf("TLSA selector must be 0-1, got: %s", parts[1])
	}

	matchingType, err := strconv.Atoi(parts[2])
	if err != nil || matchingType < 0 || matchingType > 2 {
		return nil, fmt.Errorf("TLSA matching_type must be 0-2, got: %s", parts[2])
	}

	certificate := parts[3]
	if len(certificate) == 0 {
		return nil, fmt.Errorf("TLSA certificate cannot be empty")
	}

	return map[string]interface{}{
		"usage":         usage,
		"selector":      selector,
		"matching_type": matchingType,
		"certificate":   certificate,
	}, nil
}
