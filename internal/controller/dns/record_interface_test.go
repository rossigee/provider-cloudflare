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

package record

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/dns/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

// Using error constants from main record controller

var (
	testRecordName    = "test-record"
	testZoneID        = "test-zone-id"
	testRecordID      = "test-record-id"
	testRecordType    = "SRV"
	testRecordContent = "10 20 8080 target.example.com"
	testRecordTTL     = int64(300)
)

// Test helper for interface-based external struct
type interfaceExternal struct {
	client    clients.CloudflareClient
	validator clients.DNSRecordValidator
}

func (e *interfaceExternal) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Record)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRecord)
	}

	// Check if record exists
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errRecordNoZone)
	}

	record, err := e.client.DNSRecord(ctx, *cr.Spec.ForProvider.Zone, externalName)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errRecordLookup)
	}

	// Update status with observed values
	cr.Status.AtProvider = v1alpha1.RecordObservation{
		Proxiable:  record.Proxiable,
		FQDN:       record.Name,
		Zone:       "", // Zone name not available in new API response
		Locked:     false, // Locked field not available in new API response
		CreatedOn:  &metav1.Time{Time: record.CreatedOn},
		ModifiedOn: &metav1.Time{Time: record.ModifiedOn},
	}

	cr.Status.SetConditions(xpv1.Available())

	// Check if resource is up to date
	upToDate := true
	if cr.Spec.ForProvider.Content != record.Content {
		upToDate = false
	}
	if cr.Spec.ForProvider.TTL != nil && *cr.Spec.ForProvider.TTL != int64(record.TTL) {
		upToDate = false
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (e *interfaceExternal) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Record)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRecord)
	}

	// Validate record before creation
	if cr.Spec.ForProvider.Type != nil && cr.Spec.ForProvider.Content != "" {
		var priority *int
		if cr.Spec.ForProvider.Priority != nil {
			p := int(*cr.Spec.ForProvider.Priority)
			priority = &p
		}
		
		if err := e.validator.ValidateRecord(*cr.Spec.ForProvider.Type, cr.Spec.ForProvider.Content, priority); err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, "ValidationFailed")
		}
	}

	// Create DNS record
	rr := cloudflare.DNSRecord{
		Name:    cr.Spec.ForProvider.Name,
		Type:    getStringValue(cr.Spec.ForProvider.Type),
		Content: cr.Spec.ForProvider.Content,
		TTL:     int(getInt64Value(cr.Spec.ForProvider.TTL)),
	}

	if cr.Spec.ForProvider.Priority != nil {
		priority := uint16(*cr.Spec.ForProvider.Priority)
		rr.Priority = &priority
	}

	resp, err := e.client.CreateDNSRecord(ctx, getStringValue(cr.Spec.ForProvider.Zone), rr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRecordCreation)
	}

	meta.SetExternalName(cr, resp.Result.ID)

	return managed.ExternalCreation{}, nil
}

func (e *interfaceExternal) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Record)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRecord)
	}

	// Validate record before update
	if cr.Spec.ForProvider.Type != nil && cr.Spec.ForProvider.Content != "" {
		var priority *int
		if cr.Spec.ForProvider.Priority != nil {
			p := int(*cr.Spec.ForProvider.Priority)
			priority = &p
		}
		
		if err := e.validator.ValidateRecord(*cr.Spec.ForProvider.Type, cr.Spec.ForProvider.Content, priority); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "ValidationFailed")
		}
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalUpdate{}, errors.New(errRecordUpdate)
	}

	// Update DNS record
	rr := cloudflare.DNSRecord{
		Name:    cr.Spec.ForProvider.Name,
		Type:    getStringValue(cr.Spec.ForProvider.Type),
		Content: cr.Spec.ForProvider.Content,
		TTL:     int(getInt64Value(cr.Spec.ForProvider.TTL)),
	}

	if cr.Spec.ForProvider.Priority != nil {
		priority := uint16(*cr.Spec.ForProvider.Priority)
		rr.Priority = &priority
	}

	if err := e.client.UpdateDNSRecord(ctx, getStringValue(cr.Spec.ForProvider.Zone), externalName, rr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errRecordUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

func (e *interfaceExternal) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Record)
	if !ok {
		return errors.New(errNotRecord)
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return errors.New(errRecordDeletion)
	}

	if err := e.client.DeleteDNSRecord(ctx, getStringValue(cr.Spec.ForProvider.Zone), externalName); err != nil {
		return errors.Wrap(err, errRecordDeletion)
	}

	return nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func getIntValue(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func getInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

type interfaceRecordModifier func(*v1alpha1.Record)

func withInterfaceConditions(c ...xpv1.Condition) interfaceRecordModifier {
	return func(r *v1alpha1.Record) { r.Status.Conditions = c }
}

func withInterfaceExternalName(name string) interfaceRecordModifier {
	return func(r *v1alpha1.Record) { 
		meta.SetExternalName(r, name)
	}
}

func withInterfaceSpec(s v1alpha1.RecordSpec) interfaceRecordModifier {
	return func(r *v1alpha1.Record) { r.Spec = s }
}

func withInterfaceStatus(s v1alpha1.RecordStatus) interfaceRecordModifier {
	return func(r *v1alpha1.Record) { r.Status = s }
}

func interfaceRecord(m ...interfaceRecordModifier) *v1alpha1.Record {
	cr := &v1alpha1.Record{
		ObjectMeta: metav1.ObjectMeta{
			Name: testRecordName,
		},
		Spec: v1alpha1.RecordSpec{
			ForProvider: v1alpha1.RecordParameters{
				Zone:    &testZoneID,
				Name:    testRecordName,
				Type:    &testRecordType,
				Content: testRecordContent,
				TTL:     &testRecordTTL,
			},
		},
	}

	for _, f := range m {
		f(cr)
	}

	return cr
}

func TestInterfaceObserve(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalObservation
		err    error
	}

	type args struct {
		mg resource.Managed
		cf clients.CloudflareClient
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"RecordExistsAndUpToDate": {
			reason: "Should return ResourceExists: true and ResourceUpToDate: true when record matches spec",
			args: args{
				mg: interfaceRecord(withInterfaceExternalName(testRecordID)),
				cf: &clients.MockCloudflareClient{
					DNSRecordFn: func(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error) {
						return cloudflare.DNSRecord{
							ID:      testRecordID,
							Name:    testRecordName,
							Type:    testRecordType,
							Content: testRecordContent,
							TTL:     int(testRecordTTL),
						}, nil
					},
				},
			},
			want: want{
				cr: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceConditions(xpv1.Available()),
					withInterfaceStatus(v1alpha1.RecordStatus{
						AtProvider: v1alpha1.RecordObservation{
							Proxiable:  false,
							FQDN:       testRecordName,
							Zone:       "",
							Locked:     false,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"RecordNotFound": {
			reason: "Should return ResourceExists: false when no external name is set",
			args: args{
				mg: interfaceRecord(),
			},
			want: want{
				cr: interfaceRecord(),
				result: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"RecordOutdated": {
			reason: "Should return ResourceExists: true and ResourceUpToDate: false when record content differs",
			args: args{
				mg: interfaceRecord(withInterfaceExternalName(testRecordID)),
				cf: &clients.MockCloudflareClient{
					DNSRecordFn: func(ctx context.Context, zoneID, recordID string) (cloudflare.DNSRecord, error) {
						return cloudflare.DNSRecord{
							ID:      testRecordID,
							Name:    testRecordName,
							Type:    testRecordType,
							Content: "different-content",
							TTL:     int(testRecordTTL),
						}, nil
					},
				},
			},
			want: want{
				cr: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceConditions(xpv1.Available()),
					withInterfaceStatus(v1alpha1.RecordStatus{
						AtProvider: v1alpha1.RecordObservation{
							Proxiable:  false,
							FQDN:       testRecordName,
							Zone:       "",
							Locked:     false,
						},
					}),
				),
				result: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
				},
			},
		},
		"RecordNoZone": {
			reason: "Should return error when record has no zone specified",
			args: args{
				mg: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceSpec(v1alpha1.RecordSpec{
						ForProvider: v1alpha1.RecordParameters{
							Name:    testRecordName,
							Type:    &testRecordType,
							Content: testRecordContent,
							TTL:     &testRecordTTL,
							// Zone is nil
						},
					}),
				),
			},
			want: want{
				cr: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceSpec(v1alpha1.RecordSpec{
						ForProvider: v1alpha1.RecordParameters{
							Name:    testRecordName,
							Type:    &testRecordType,
							Content: testRecordContent,
							TTL:     &testRecordTTL,
						},
					}),
				),
				err: errors.New(errRecordNoZone),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &interfaceExternal{
				client:    tc.args.cf,
				validator: clients.NewDNSRecordValidator(),
			}

			o, err := e.Observe(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestInterfaceCreate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalCreation
		err    error
	}

	type args struct {
		mg resource.Managed
		cf clients.CloudflareClient
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulSRVRecord": {
			reason: "Should successfully create SRV record with validation",
			args: args{
				mg: interfaceRecord(),
				cf: &clients.MockCloudflareClient{
					CreateDNSRecordFn: func(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
						return &cloudflare.DNSRecordResponse{
							Result: cloudflare.DNSRecord{
								ID:      testRecordID,
									Name:    rr.Name,
								Type:    rr.Type,
								Content: rr.Content,
								TTL:     rr.TTL,
							},
						}, nil
					},
				},
			},
			want: want{
				cr: interfaceRecord(withInterfaceExternalName(testRecordID)),
				result: managed.ExternalCreation{
				},
			},
		},
		"FailedSRVValidation": {
			reason: "Should fail to create SRV record with invalid content",
			args: args{
				mg: interfaceRecord(withInterfaceSpec(v1alpha1.RecordSpec{
					ForProvider: v1alpha1.RecordParameters{
						Zone:    &testZoneID,
						Name:    testRecordName,
						Type:    stringPtr("SRV"),
						Content: "invalid srv content", // Invalid SRV format
						TTL:     &testRecordTTL,
					},
				})),
				cf: clients.NewMockCloudflareClient(),
			},
			want: want{
				cr: interfaceRecord(withInterfaceSpec(v1alpha1.RecordSpec{
					ForProvider: v1alpha1.RecordParameters{
						Zone:    &testZoneID,
						Name:    testRecordName,
						Type:    stringPtr("SRV"),
						Content: "invalid srv content",
						TTL:     &testRecordTTL,
					},
				})),
				err: errors.Wrap(errors.New("SRV record must have format: priority weight port target"), "ValidationFailed"),
			},
		},
		"SuccessfulARecord": {
			reason: "Should successfully create A record",
			args: args{
				mg: interfaceRecord(withInterfaceSpec(v1alpha1.RecordSpec{
					ForProvider: v1alpha1.RecordParameters{
						Zone:    &testZoneID,
						Name:    testRecordName,
						Type:    stringPtr("A"),
						Content: "192.168.1.1",
						TTL:     &testRecordTTL,
					},
				})),
				cf: &clients.MockCloudflareClient{
					CreateDNSRecordFn: func(ctx context.Context, zoneID string, rr cloudflare.DNSRecord) (*cloudflare.DNSRecordResponse, error) {
						return &cloudflare.DNSRecordResponse{
							Result: cloudflare.DNSRecord{
								ID:      testRecordID,
									Name:    rr.Name,
								Type:    rr.Type,
								Content: rr.Content,
								TTL:     rr.TTL,
							},
						}, nil
					},
				},
			},
			want: want{
				cr: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceSpec(v1alpha1.RecordSpec{
						ForProvider: v1alpha1.RecordParameters{
							Zone:    &testZoneID,
							Name:    testRecordName,
							Type:    stringPtr("A"),
							Content: "192.168.1.1",
							TTL:     &testRecordTTL,
						},
					}),
				),
				result: managed.ExternalCreation{
				},
			},
		},
		"FailedARecordValidation": {
			reason: "Should fail to create A record with invalid IP",
			args: args{
				mg: interfaceRecord(withInterfaceSpec(v1alpha1.RecordSpec{
					ForProvider: v1alpha1.RecordParameters{
						Zone:    &testZoneID,
						Name:    testRecordName,
						Type:    stringPtr("A"),
						Content: "256.1.1.1", // Invalid IP
						TTL:     &testRecordTTL,
					},
				})),
				cf: clients.NewMockCloudflareClient(),
			},
			want: want{
				cr: interfaceRecord(withInterfaceSpec(v1alpha1.RecordSpec{
					ForProvider: v1alpha1.RecordParameters{
						Zone:    &testZoneID,
						Name:    testRecordName,
						Type:    stringPtr("A"),
						Content: "256.1.1.1",
						TTL:     &testRecordTTL,
					},
				})),
				err: errors.Wrap(errors.New("IPv4 address octets must be between 0 and 255"), "ValidationFailed"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &interfaceExternal{
				client:    tc.args.cf,
				validator: clients.NewDNSRecordValidator(),
			}

			o, err := e.Create(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestInterfaceUpdate(t *testing.T) {
	type want struct {
		cr     resource.Managed
		result managed.ExternalUpdate
		err    error
	}

	type args struct {
		mg resource.Managed
		cf clients.CloudflareClient
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulUpdate": {
			reason: "Should successfully update DNS record",
			args: args{
				mg: interfaceRecord(withInterfaceExternalName(testRecordID)),
				cf: &clients.MockCloudflareClient{
					UpdateDNSRecordFn: func(ctx context.Context, zoneID, recordID string, rr cloudflare.DNSRecord) error {
						return nil
					},
				},
			},
			want: want{
				cr:     interfaceRecord(withInterfaceExternalName(testRecordID)),
				result: managed.ExternalUpdate{},
			},
		},
		"FailedValidationOnUpdate": {
			reason: "Should fail to update record with invalid content",
			args: args{
				mg: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceSpec(v1alpha1.RecordSpec{
						ForProvider: v1alpha1.RecordParameters{
							Zone:    &testZoneID,
							Name:    testRecordName,
							Type:    stringPtr("A"),
							Content: "256.1.1.1", // Invalid IP
							TTL:     &testRecordTTL,
						},
					}),
				),
				cf: clients.NewMockCloudflareClient(),
			},
			want: want{
				cr: interfaceRecord(
					withInterfaceExternalName(testRecordID),
					withInterfaceSpec(v1alpha1.RecordSpec{
						ForProvider: v1alpha1.RecordParameters{
							Zone:    &testZoneID,
							Name:    testRecordName,
							Type:    stringPtr("A"),
							Content: "256.1.1.1",
							TTL:     &testRecordTTL,
						},
					}),
				),
				err: errors.Wrap(errors.New("IPv4 address octets must be between 0 and 255"), "ValidationFailed"),
			},
		},
		"NoExternalName": {
			reason: "Should fail when no external name is set",
			args: args{
				mg: interfaceRecord(),
				cf: clients.NewMockCloudflareClient(),
			},
			want: want{
				cr:  interfaceRecord(),
				err: errors.New(errRecordUpdate),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &interfaceExternal{
				client:    tc.args.cf,
				validator: clients.NewDNSRecordValidator(),
			}

			o, err := e.Update(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.result, o); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestInterfaceDelete(t *testing.T) {
	type want struct {
		cr  resource.Managed
		err error
	}

	type args struct {
		mg resource.Managed
		cf clients.CloudflareClient
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulDelete": {
			reason: "Should successfully delete DNS record",
			args: args{
				mg: interfaceRecord(withInterfaceExternalName(testRecordID)),
				cf: &clients.MockCloudflareClient{
					DeleteDNSRecordFn: func(ctx context.Context, zoneID, recordID string) error {
						return nil
					},
				},
			},
			want: want{
				cr: interfaceRecord(withInterfaceExternalName(testRecordID)),
			},
		},
		"NoExternalName": {
			reason: "Should fail when no external name is set",
			args: args{
				mg: interfaceRecord(),
				cf: clients.NewMockCloudflareClient(),
			},
			want: want{
				cr:  interfaceRecord(),
				err: errors.New(errRecordDeletion),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &interfaceExternal{
				client:    tc.args.cf,
				validator: clients.NewDNSRecordValidator(),
			}

			err := e.Delete(context.Background(), tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.cr, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

// Test mock client call tracking
func TestMockClientCallTracking(t *testing.T) {
	mockClient := clients.NewMockCloudflareClient()
	validator := clients.NewDNSRecordValidator()

	e := &interfaceExternal{
		client:    mockClient,
		validator: validator,
	}

	// Test Create call tracking
	cr := interfaceRecord()
	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if !mockClient.WasCreateDNSRecordCalled() {
		t.Error("Expected CreateDNSRecord to be called")
	}

	lastCall := mockClient.GetLastCreateDNSRecordCall()
	if lastCall == nil {
		t.Error("Expected to get last CreateDNSRecord call")
	} else {
		if lastCall.ZoneID != testZoneID {
			t.Errorf("Expected zone ID %s, got %s", testZoneID, lastCall.ZoneID)
		}
		if lastCall.Record.Type != testRecordType {
			t.Errorf("Expected record type %s, got %s", testRecordType, lastCall.Record.Type)
		}
	}

	// Test Update call tracking
	meta.SetExternalName(cr, testRecordID)
	_, err = e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !mockClient.WasUpdateDNSRecordCalled() {
		t.Error("Expected UpdateDNSRecord to be called")
	}

	updateCall := mockClient.GetLastUpdateDNSRecordCall()
	if updateCall == nil {
		t.Error("Expected to get last UpdateDNSRecord call")
	} else {
		if updateCall.RecordID != testRecordID {
			t.Errorf("Expected record ID %s, got %s", testRecordID, updateCall.RecordID)
		}
	}
}

type MockRoundTripper struct{}

func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("{}")),
	}, nil
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}