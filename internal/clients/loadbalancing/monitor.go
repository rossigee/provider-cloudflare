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

package loadbalancing

import (
	"context"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1alpha1"
	clients "github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	errCreateMonitor = "failed to create load balancer monitor"
	errGetMonitor    = "failed to get load balancer monitor"
	errUpdateMonitor = "failed to update load balancer monitor"
	errDeleteMonitor = "failed to delete load balancer monitor"
)

// MonitorClient interface for Cloudflare Load Balancer Monitor operations
type MonitorClient interface {
	CreateMonitor(ctx context.Context, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	GetMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	UpdateMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error)
	DeleteMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) error
}

// NewMonitorClient creates a new Cloudflare Load Balancer Monitor client
func NewMonitorClient(cfg clients.Config, httpClient *http.Client) (MonitorClient, error) {
	api, err := clients.NewClient(cfg, httpClient)
	if err != nil {
		return nil, err
	}
	return &monitorClient{api: api}, nil
}

type monitorClient struct {
	api *cloudflare.API
}

// CreateMonitor creates a new Cloudflare load balancer monitor
func (c *monitorClient) CreateMonitor(ctx context.Context, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	monitor := cloudflare.LoadBalancerMonitor{
		Type: params.Type,
	}

	if params.Description != nil {
		monitor.Description = *params.Description
	}

	if params.Method != nil {
		monitor.Method = *params.Method
	}

	if params.Path != nil {
		monitor.Path = *params.Path
	}

	if params.Header != nil {
		monitor.Header = params.Header
	}

	if params.Timeout != nil {
		monitor.Timeout = *params.Timeout
	}

	if params.Retries != nil {
		monitor.Retries = *params.Retries
	}

	if params.Interval != nil {
		monitor.Interval = *params.Interval
	}

	if params.ConsecutiveUp != nil {
		monitor.ConsecutiveUp = *params.ConsecutiveUp
	}

	if params.ConsecutiveDown != nil {
		monitor.ConsecutiveDown = *params.ConsecutiveDown
	}

	if params.Port != nil {
		monitor.Port = uint16(*params.Port)
	}

	if params.ExpectedBody != nil {
		monitor.ExpectedBody = *params.ExpectedBody
	}

	if params.ExpectedCodes != nil {
		monitor.ExpectedCodes = *params.ExpectedCodes
	}

	if params.FollowRedirects != nil {
		monitor.FollowRedirects = *params.FollowRedirects
	}

	if params.AllowInsecure != nil {
		monitor.AllowInsecure = *params.AllowInsecure
	}

	if params.ProbeZone != nil {
		monitor.ProbeZone = *params.ProbeZone
	}

	createParams := cloudflare.CreateLoadBalancerMonitorParams{
		LoadBalancerMonitor: monitor,
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	result, err := c.api.CreateLoadBalancerMonitor(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateMonitor)
	}

	return &result, nil
}

// GetMonitor retrieves a Cloudflare load balancer monitor
func (c *monitorClient) GetMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	monitor, err := c.api.GetLoadBalancerMonitor(ctx, rc, monitorID)
	if err != nil {
		return nil, errors.Wrap(err, errGetMonitor)
	}

	return &monitor, nil
}

// UpdateMonitor updates a Cloudflare load balancer monitor
func (c *monitorClient) UpdateMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) (*cloudflare.LoadBalancerMonitor, error) {
	monitor := cloudflare.LoadBalancerMonitor{
		ID:   monitorID,
		Type: params.Type,
	}

	if params.Description != nil {
		monitor.Description = *params.Description
	}

	if params.Method != nil {
		monitor.Method = *params.Method
	}

	if params.Path != nil {
		monitor.Path = *params.Path
	}

	if params.Header != nil {
		monitor.Header = params.Header
	}

	if params.Timeout != nil {
		monitor.Timeout = *params.Timeout
	}

	if params.Retries != nil {
		monitor.Retries = *params.Retries
	}

	if params.Interval != nil {
		monitor.Interval = *params.Interval
	}

	if params.ConsecutiveUp != nil {
		monitor.ConsecutiveUp = *params.ConsecutiveUp
	}

	if params.ConsecutiveDown != nil {
		monitor.ConsecutiveDown = *params.ConsecutiveDown
	}

	if params.Port != nil {
		monitor.Port = uint16(*params.Port)
	}

	if params.ExpectedBody != nil {
		monitor.ExpectedBody = *params.ExpectedBody
	}

	if params.ExpectedCodes != nil {
		monitor.ExpectedCodes = *params.ExpectedCodes
	}

	if params.FollowRedirects != nil {
		monitor.FollowRedirects = *params.FollowRedirects
	}

	if params.AllowInsecure != nil {
		monitor.AllowInsecure = *params.AllowInsecure
	}

	if params.ProbeZone != nil {
		monitor.ProbeZone = *params.ProbeZone
	}

	updateParams := cloudflare.UpdateLoadBalancerMonitorParams{
		LoadBalancerMonitor: monitor,
	}

	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return nil, errors.New("either zone or account must be specified")
	}

	result, err := c.api.UpdateLoadBalancerMonitor(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateMonitor)
	}

	return &result, nil
}

// DeleteMonitor deletes a Cloudflare load balancer monitor
func (c *monitorClient) DeleteMonitor(ctx context.Context, monitorID string, params v1alpha1.LoadBalancerMonitorParameters) error {
	var rc *cloudflare.ResourceContainer
	if params.Zone != nil {
		rc = cloudflare.ZoneIdentifier(*params.Zone)
	} else if params.Account != nil {
		rc = cloudflare.AccountIdentifier(*params.Account)
	} else {
		return errors.New("either zone or account must be specified")
	}

	err := c.api.DeleteLoadBalancerMonitor(ctx, rc, monitorID)
	if err != nil {
		return errors.Wrap(err, errDeleteMonitor)
	}

	return nil
}

// IsMonitorNotFound checks if error indicates monitor not found
func IsMonitorNotFound(err error) bool {
	if err == nil {
		return false
	}
	// Check for Cloudflare API not found errors
	if cfErr := (*cloudflare.Error)(nil); errors.As(err, &cfErr) {
		return cfErr.StatusCode == 404
	}
	return false
}

// GenerateMonitorObservation creates observation from Cloudflare load balancer monitor
func GenerateMonitorObservation(monitor *cloudflare.LoadBalancerMonitor) v1alpha1.LoadBalancerMonitorObservation {
	observation := v1alpha1.LoadBalancerMonitorObservation{
		ID: monitor.ID,
	}

	if monitor.CreatedOn != nil {
		createdOn := monitor.CreatedOn.String()
		observation.CreatedOn = &createdOn
	}

	if monitor.ModifiedOn != nil {
		modifiedOn := monitor.ModifiedOn.String()
		observation.ModifiedOn = &modifiedOn
	}

	return observation
}

// IsMonitorUpToDate determines if the Cloudflare load balancer monitor is up to date
func IsMonitorUpToDate(params *v1alpha1.LoadBalancerMonitorParameters, monitor *cloudflare.LoadBalancerMonitor) bool {
	if params.Type != monitor.Type {
		return false
	}

	if params.Description != nil && *params.Description != monitor.Description {
		return false
	}

	if params.Description == nil && monitor.Description != "" {
		return false
	}

	if params.Method != nil && *params.Method != monitor.Method {
		return false
	}

	if params.Path != nil && *params.Path != monitor.Path {
		return false
	}

	if params.Timeout != nil && *params.Timeout != monitor.Timeout {
		return false
	}

	if params.Retries != nil && *params.Retries != monitor.Retries {
		return false
	}

	if params.Interval != nil && *params.Interval != monitor.Interval {
		return false
	}

	if params.ConsecutiveUp != nil && *params.ConsecutiveUp != monitor.ConsecutiveUp {
		return false
	}

	if params.ConsecutiveDown != nil && *params.ConsecutiveDown != monitor.ConsecutiveDown {
		return false
	}

	if params.Port != nil && uint16(*params.Port) != monitor.Port {
		return false
	}

	if params.ExpectedBody != nil && *params.ExpectedBody != monitor.ExpectedBody {
		return false
	}

	if params.ExpectedCodes != nil && *params.ExpectedCodes != monitor.ExpectedCodes {
		return false
	}

	if params.FollowRedirects != nil && *params.FollowRedirects != monitor.FollowRedirects {
		return false
	}

	if params.AllowInsecure != nil && *params.AllowInsecure != monitor.AllowInsecure {
		return false
	}

	if params.ProbeZone != nil && *params.ProbeZone != monitor.ProbeZone {
		return false
	}

	return true
}