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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	rtv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/logging"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/cloudflare/cloudflare-go"

	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
	records "github.com/rossigee/provider-cloudflare/internal/clients/records"
	metrics "github.com/rossigee/provider-cloudflare/internal/metrics"
	"github.com/rossigee/provider-cloudflare/internal/tracing"
)

const (
	errNotRecord = "managed resource is not a Record custom resource"

	errClientConfig = "error getting client config"

	errRecordLookup   = "cannot lookup record"
	errRecordCreation = "cannot create record"
	errRecordUpdate   = "cannot update record"
	errRecordDeletion = "cannot delete record"
	errRecordNoZone   = "no zone found"

	maxConcurrency = 5

	// recordStatusActive = "active"
)

// Setup adds a controller that reconciles Record managed resources.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.TypedRateLimiter[any]) error {
	name := managed.ControllerName(v1beta1.RecordGroupKind)

	o := controller.Options{
		RateLimiter: nil, // Use default rate limiter
		MaxConcurrentReconciles: maxConcurrency,
	}

	hc := metrics.NewInstrumentedHTTPClient(name)
	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.RecordGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube: mgr.GetClient(),
			newCloudflareClientFn: func(cfg clients.Config) (records.Client, error) {
				return records.NewClient(cfg, hc)
			},
		}),
		managed.WithLogger(l.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorder(name))), 
		managed.WithPollInterval(5*time.Minute),
		// Do not initialize external-name field.
		managed.WithInitializers(),
	)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o).
		For(&v1beta1.Record{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube                  client.Client
	newCloudflareClientFn func(cfg clients.Config) (records.Client, error)
}

// Connect produces a valid configuration for a Cloudflare API
// instance, and returns it as an external client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, ok := mg.(*v1beta1.Record)
	if !ok {
		return nil, errors.New(errNotRecord)
	}

	// Get client configuration
	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errClientConfig)
	}

	client, err := c.newCloudflareClientFn(*config)
	if err != nil {
		return nil, err
	}

	return &external{client: client}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client records.Client
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	_, span := tracing.StartSpan(ctx, "record.observe",
		tracing.SpanAttrs("record", func() string { if mg == nil { return "" }; return mg.GetName() }(), "observe")...)
	defer span.End()

	cr, ok := mg.(*v1beta1.Record)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRecord)
	}

	// Record does not exist if we dont have an ID stored in external-name
	rid := meta.GetExternalName(cr)
	if rid == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalObservation{}, errors.New(errRecordNoZone)
	}

	rc := cloudflare.ZoneIdentifier(*cr.Spec.ForProvider.Zone)
	record, err := e.client.GetDNSRecord(ctx, rc, rid)

	if err != nil {
		return managed.ExternalObservation{},
			errors.Wrap(resource.Ignore(records.IsRecordNotFound, err), errRecordLookup)
	}

	cr.Status.AtProvider = records.GenerateObservation(record)

	cr.SetConditions(rtv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: records.LateInitialize(&cr.Spec.ForProvider, record),
		ResourceUpToDate:        records.UpToDate(&cr.Spec.ForProvider, record),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	_, span := tracing.StartSpan(ctx, "record.create",
		tracing.SpanAttrs("record", func() string { if mg == nil { return "" }; return mg.GetName() }(), "create")...)
	defer span.End()

	cr, ok := mg.(*v1beta1.Record)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRecord)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalCreation{},
			errors.Wrap(errors.New(errRecordNoZone), errRecordCreation)
	}

	if cr.Spec.ForProvider.TTL == nil {
		return managed.ExternalCreation{}, errors.New(errRecordCreation)
	}

	if cr.Spec.ForProvider.Type == nil {
		return managed.ExternalCreation{}, errors.New(errRecordCreation)
	}

	// Required for MX and URI records; unused by other record types.
	if cr.Spec.ForProvider.Priority == nil {
		switch *cr.Spec.ForProvider.Type {
		case "MX", "URI":
			return managed.ExternalCreation{}, errors.New(errRecordCreation)
		}
	}

	// SRV records require priority, weight, and port fields
	if *cr.Spec.ForProvider.Type == "SRV" {
		if cr.Spec.ForProvider.Priority == nil || cr.Spec.ForProvider.Weight == nil || cr.Spec.ForProvider.Port == nil {
			return managed.ExternalCreation{}, errors.New("SRV records require priority, weight, and port fields")
		}
	}

	cr.SetConditions(rtv1.Creating())

	ttl := int(*cr.Spec.ForProvider.TTL)
	var pri *uint16
	if cr.Spec.ForProvider.Priority != nil {
		val := uint16(*cr.Spec.ForProvider.Priority)
		pri = &val
	}

	rc := cloudflare.ZoneIdentifier(*cr.Spec.ForProvider.Zone)
	params := cloudflare.CreateDNSRecordParams{
		Type:    *cr.Spec.ForProvider.Type,
		Name:    cr.Spec.ForProvider.Name,
		Content: cr.Spec.ForProvider.Content,
		TTL:     ttl,
		Proxied: cr.Spec.ForProvider.Proxied,
	}
	if pri != nil {
		params.Priority = pri
	}

	// For SRV records, use the Data field instead of Priority/Content
	if *cr.Spec.ForProvider.Type == "SRV" {
		srvData := map[string]interface{}{
			"priority": int(*cr.Spec.ForProvider.Priority),
			"weight":   int(*cr.Spec.ForProvider.Weight),
			"port":     int(*cr.Spec.ForProvider.Port),
			"target":   cr.Spec.ForProvider.Content,
		}
		params.Data = srvData
		params.Priority = nil
		params.Content = ""
	}

	// For TLSA records, parse content and use Data field
	if *cr.Spec.ForProvider.Type == "TLSA" {
		tlsaData, err := parseTLSAContent(cr.Spec.ForProvider.Content)
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errRecordCreation)
		}
		params.Data = tlsaData
		params.Content = ""
	}
	
	res, err := e.client.CreateDNSRecord(ctx, rc, params)

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errRecordCreation)
	}

	cr.Status.AtProvider = records.GenerateObservation(res)

	// Update the external name with the ID of the new DNS Record
	meta.SetExternalName(cr, res.ID)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, span := tracing.StartSpan(ctx, "record.update",
		tracing.SpanAttrs("record", func() string { if mg == nil { return "" }; return mg.GetName() }(), "update")...)
	defer span.End()

	cr, ok := mg.(*v1beta1.Record)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRecord)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalUpdate{}, errors.Wrap(errors.New(errRecordNoZone), errRecordUpdate)
	}

	rid := meta.GetExternalName(cr)

	// Update should never be called on a nonexistent resource
	if rid == "" {
		return managed.ExternalUpdate{}, errors.New(errRecordUpdate)
	}

	return managed.ExternalUpdate{},
		errors.Wrap(
			records.UpdateRecord(ctx, e.client, *cr.Spec.ForProvider.Zone, rid, &cr.Spec.ForProvider),
			errRecordUpdate,
		)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	_, span := tracing.StartSpan(ctx, "record.delete",
		tracing.SpanAttrs("record", func() string { if mg == nil { return "" }; return mg.GetName() }(), "delete")...)
	defer span.End()

	cr, ok := mg.(*v1beta1.Record)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotRecord)
	}

	if cr.Spec.ForProvider.Zone == nil {
		return managed.ExternalDelete{}, errors.Wrap(errors.New(errRecordNoZone), errRecordDeletion)
	}

	rid := meta.GetExternalName(cr)

	// Delete should never be called on a nonexistent resource
	if rid == "" {
		return managed.ExternalDelete{}, errors.New(errRecordDeletion)
	}

	rc := cloudflare.ZoneIdentifier(*cr.Spec.ForProvider.Zone)
	err := e.client.DeleteDNSRecord(ctx, rc, meta.GetExternalName(cr))
	return managed.ExternalDelete{}, errors.Wrap(err, errRecordDeletion)
}

func (e *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
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
