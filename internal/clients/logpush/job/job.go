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

package job

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/logpush/v1alpha1"
)

// LogpushJobAPI defines the interface for Logpush Job operations
type LogpushJobAPI interface {
	Accounts(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error)
	CreateLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error)
	GetLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error)
	UpdateLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLogpushJobParams) error
	DeleteLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error
	ListLogpushJobs(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error)
}

const (
	errCreateJob = "cannot create logpush job"
	errUpdateJob = "cannot update logpush job"
	errGetJob    = "cannot get logpush job"
	errDeleteJob = "cannot delete logpush job"
	errListJobs  = "cannot list logpush jobs"
)

// JobClient provides operations for Logpush Jobs.
type JobClient struct {
	client    LogpushJobAPI
	accountID string
}

// NewClient creates a new Logpush Job client.
func NewClient(client LogpushJobAPI) *JobClient {
	return &JobClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
	}
}

// getAccountID gets the account ID from the Cloudflare API
func (c *JobClient) getAccountID(ctx context.Context) (string, error) {
	if c.accountID != "" {
		return c.accountID, nil
	}
	
	// Get account ID from Cloudflare API by listing accounts
	// Most users have access to only one account, so we'll use the first one
	accounts, _, err := c.client.Accounts(ctx, cloudflare.AccountsListParams{})
	if err != nil {
		return "", errors.Wrap(err, "failed to list accounts")
	}
	
	if len(accounts) == 0 {
		return "", errors.New("no accounts found")
	}
	
	// Use the first account (most common case for users)
	c.accountID = accounts[0].ID
	return c.accountID, nil
}

// convertToObservation converts cloudflare-go logpush job to Crossplane observation.
func convertToObservation(job cloudflare.LogpushJob) v1alpha1.JobObservation {
	obs := v1alpha1.JobObservation{
		ID:              &job.ID,
		Dataset:         job.Dataset,
		Name:            job.Name,
		DestinationConf: job.DestinationConf,
	}

	if job.Enabled {
		obs.Enabled = &job.Enabled
	}

	if job.Kind != "" {
		obs.Kind = &job.Kind
	}

	if job.LogpullOptions != "" {
		obs.LogpullOptions = &job.LogpullOptions
	}

	if job.OutputOptions != nil {
		obs.OutputOptions = convertOutputOptions(job.OutputOptions)
	}

	if job.OwnershipChallenge != "" {
		obs.OwnershipChallenge = &job.OwnershipChallenge
	}

	if job.LastComplete != nil {
		obs.LastComplete = &metav1.Time{Time: *job.LastComplete}
	}

	if job.LastError != nil {
		obs.LastError = &metav1.Time{Time: *job.LastError}
	}

	if job.ErrorMessage != "" {
		obs.ErrorMessage = &job.ErrorMessage
	}

	if job.Frequency != "" {
		obs.Frequency = &job.Frequency
	}

	if job.Filter != nil {
		obs.Filter = convertJobFilters(job.Filter)
	}

	if job.MaxUploadBytes > 0 {
		obs.MaxUploadBytes = &job.MaxUploadBytes
	}

	if job.MaxUploadRecords > 0 {
		obs.MaxUploadRecords = &job.MaxUploadRecords
	}

	if job.MaxUploadIntervalSeconds > 0 {
		obs.MaxUploadIntervalSeconds = &job.MaxUploadIntervalSeconds
	}

	return obs
}

// convertOutputOptions converts cloudflare-go output options to Crossplane output options.
func convertOutputOptions(opts *cloudflare.LogpushOutputOptions) *v1alpha1.OutputOptions {
	if opts == nil {
		return nil
	}

	result := &v1alpha1.OutputOptions{}

	if len(opts.FieldNames) > 0 {
		result.FieldNames = opts.FieldNames
	}

	if opts.OutputType != "" {
		result.OutputType = &opts.OutputType
	}

	if opts.BatchPrefix != "" {
		result.BatchPrefix = &opts.BatchPrefix
	}

	if opts.BatchSuffix != "" {
		result.BatchSuffix = &opts.BatchSuffix
	}

	if opts.RecordPrefix != "" {
		result.RecordPrefix = &opts.RecordPrefix
	}

	if opts.RecordSuffix != "" {
		result.RecordSuffix = &opts.RecordSuffix
	}

	if opts.RecordTemplate != "" {
		result.RecordTemplate = &opts.RecordTemplate
	}

	if opts.RecordDelimiter != "" {
		result.RecordDelimiter = &opts.RecordDelimiter
	}

	if opts.FieldDelimiter != "" {
		result.FieldDelimiter = &opts.FieldDelimiter
	}

	if opts.TimestampFormat != "" {
		result.TimestampFormat = &opts.TimestampFormat
	}

	if opts.SampleRate != 0 {
		sampleRateStr := fmt.Sprintf("%f", opts.SampleRate)
		result.SampleRate = &sampleRateStr
	}

	return result
}

// convertJobFilters converts cloudflare-go job filters to Crossplane job filters.
func convertJobFilters(filters *cloudflare.LogpushJobFilters) *v1alpha1.JobFilters {
	if filters == nil {
		return nil
	}

	return &v1alpha1.JobFilters{
		Where: convertJobFilter(&filters.Where),
	}
}

// convertJobFilter converts cloudflare-go job filter to Crossplane job filter.
func convertJobFilter(filter *cloudflare.LogpushJobFilter) *v1alpha1.JobFilter {
	if filter == nil {
		return nil
	}

	result := &v1alpha1.JobFilter{}

	if filter.Key != "" {
		result.Key = &filter.Key
	}

	if filter.Operator != "" {
		operatorStr := string(filter.Operator)
		result.Operator = &operatorStr
	}

	if filter.Value != nil {
		valueStr := fmt.Sprintf("%v", filter.Value)
		result.Value = &valueStr
	}

	return result
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.JobParameters) cloudflare.CreateLogpushJobParams {
	cfParams := cloudflare.CreateLogpushJobParams{
		Dataset:         params.Dataset,
		Name:            params.Name,
		DestinationConf: params.DestinationConf,
	}

	if params.Enabled != nil {
		cfParams.Enabled = *params.Enabled
	}

	if params.Kind != nil {
		cfParams.Kind = *params.Kind
	}

	if params.LogpullOptions != nil {
		cfParams.LogpullOptions = *params.LogpullOptions
	}

	if params.OutputOptions != nil {
		cfParams.OutputOptions = convertToCloudflareOutputOptions(params.OutputOptions)
	}

	if params.Frequency != nil {
		cfParams.Frequency = *params.Frequency
	}

	if params.Filter != nil {
		cfParams.Filter = convertToCloudflareJobFilters(params.Filter)
	}

	if params.MaxUploadBytes != nil {
		cfParams.MaxUploadBytes = *params.MaxUploadBytes
	}

	if params.MaxUploadRecords != nil {
		cfParams.MaxUploadRecords = *params.MaxUploadRecords
	}

	if params.MaxUploadIntervalSeconds != nil {
		cfParams.MaxUploadIntervalSeconds = *params.MaxUploadIntervalSeconds
	}

	return cfParams
}

// convertToCloudflareOutputOptions converts Crossplane output options to cloudflare-go output options.
func convertToCloudflareOutputOptions(opts *v1alpha1.OutputOptions) *cloudflare.LogpushOutputOptions {
	if opts == nil {
		return nil
	}

	result := &cloudflare.LogpushOutputOptions{}

	if len(opts.FieldNames) > 0 {
		result.FieldNames = opts.FieldNames
	}

	if opts.OutputType != nil {
		result.OutputType = *opts.OutputType
	}

	if opts.BatchPrefix != nil {
		result.BatchPrefix = *opts.BatchPrefix
	}

	if opts.BatchSuffix != nil {
		result.BatchSuffix = *opts.BatchSuffix
	}

	if opts.RecordPrefix != nil {
		result.RecordPrefix = *opts.RecordPrefix
	}

	if opts.RecordSuffix != nil {
		result.RecordSuffix = *opts.RecordSuffix
	}

	if opts.RecordTemplate != nil {
		result.RecordTemplate = *opts.RecordTemplate
	}

	if opts.RecordDelimiter != nil {
		result.RecordDelimiter = *opts.RecordDelimiter
	}

	if opts.FieldDelimiter != nil {
		result.FieldDelimiter = *opts.FieldDelimiter
	}

	if opts.TimestampFormat != nil {
		result.TimestampFormat = *opts.TimestampFormat
	}

	if opts.SampleRate != nil {
		if sampleRate, err := strconv.ParseFloat(*opts.SampleRate, 64); err == nil {
			result.SampleRate = sampleRate
		}
	}

	return result
}

// convertToCloudflareJobFilters converts Crossplane job filters to cloudflare-go job filters.
func convertToCloudflareJobFilters(filters *v1alpha1.JobFilters) *cloudflare.LogpushJobFilters {
	if filters == nil {
		return nil
	}

	result := &cloudflare.LogpushJobFilters{}

	if filters.Where != nil {
		result.Where = *convertToCloudflareJobFilter(filters.Where)
	}

	return result
}

// convertToCloudflareJobFilter converts Crossplane job filter to cloudflare-go job filter.
func convertToCloudflareJobFilter(filter *v1alpha1.JobFilter) *cloudflare.LogpushJobFilter {
	if filter == nil {
		return nil
	}

	result := &cloudflare.LogpushJobFilter{}

	if filter.Key != nil {
		result.Key = *filter.Key
	}

	if filter.Operator != nil {
		result.Operator = cloudflare.Operator(*filter.Operator)
	}

	if filter.Value != nil {
		result.Value = *filter.Value
	}

	return result
}

// Create creates a new Logpush Job.
func (c *JobClient) Create(ctx context.Context, params v1alpha1.JobParameters) (*v1alpha1.JobObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	createParams := convertToCloudflareParams(params)
	
	job, err := c.client.CreateLogpushJob(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateJob)
	}

	obs := convertToObservation(*job)
	return &obs, nil
}

// Get retrieves a Logpush Job.
func (c *JobClient) Get(ctx context.Context, jobID int) (*v1alpha1.JobObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	job, err := c.client.GetLogpushJob(ctx, rc, jobID)
	if err != nil {
		return nil, errors.Wrap(err, errGetJob)
	}

	obs := convertToObservation(job)
	return &obs, nil
}

// Update updates an existing Logpush Job.
func (c *JobClient) Update(ctx context.Context, jobID int, params v1alpha1.JobParameters) (*v1alpha1.JobObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	updateParams := cloudflare.UpdateLogpushJobParams{
		ID:              jobID,
		Dataset:         params.Dataset,
		Name:            params.Name,
		DestinationConf: params.DestinationConf,
	}

	if params.Enabled != nil {
		updateParams.Enabled = *params.Enabled
	}

	if params.Kind != nil {
		updateParams.Kind = *params.Kind
	}

	if params.LogpullOptions != nil {
		updateParams.LogpullOptions = *params.LogpullOptions
	}

	if params.OutputOptions != nil {
		updateParams.OutputOptions = convertToCloudflareOutputOptions(params.OutputOptions)
	}

	if params.Frequency != nil {
		updateParams.Frequency = *params.Frequency
	}

	if params.Filter != nil {
		updateParams.Filter = convertToCloudflareJobFilters(params.Filter)
	}

	if params.MaxUploadBytes != nil {
		updateParams.MaxUploadBytes = *params.MaxUploadBytes
	}

	if params.MaxUploadRecords != nil {
		updateParams.MaxUploadRecords = *params.MaxUploadRecords
	}

	if params.MaxUploadIntervalSeconds != nil {
		updateParams.MaxUploadIntervalSeconds = *params.MaxUploadIntervalSeconds
	}

	err = c.client.UpdateLogpushJob(ctx, rc, updateParams)
	if err != nil {
		return nil, errors.Wrap(err, errUpdateJob)
	}

	// Get the updated job to return the observation
	return c.Get(ctx, jobID)
}

// Delete removes a Logpush Job.
func (c *JobClient) Delete(ctx context.Context, jobID int) error {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	err = c.client.DeleteLogpushJob(ctx, rc, jobID)
	if err != nil && !IsJobNotFound(err) {
		return errors.Wrap(err, errDeleteJob)
	}

	return nil
}

// List retrieves all Logpush Jobs.
func (c *JobClient) List(ctx context.Context) ([]v1alpha1.JobObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	jobs, err := c.client.ListLogpushJobs(ctx, rc, cloudflare.ListLogpushJobsParams{})
	if err != nil {
		return nil, errors.Wrap(err, errListJobs)
	}

	observations := make([]v1alpha1.JobObservation, len(jobs))
	for i, job := range jobs {
		observations[i] = convertToObservation(job)
	}

	return observations, nil
}

// IsUpToDate checks if the Logpush Job is up to date.
func (c *JobClient) IsUpToDate(ctx context.Context, params v1alpha1.JobParameters, obs v1alpha1.JobObservation) (bool, error) {
	// Compare key fields to determine if update is needed
	if obs.Name != params.Name ||
		obs.Dataset != params.Dataset ||
		obs.DestinationConf != params.DestinationConf {
		return false, nil
	}

	if params.Enabled != nil && (obs.Enabled == nil || *obs.Enabled != *params.Enabled) {
		return false, nil
	}

	return true, nil
}

// IsJobNotFound returns true if the error indicates the job was not found
func IsJobNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "job not found" ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}

// ParseJobID parses a string job ID to int
func ParseJobID(jobIDStr string) (int, error) {
	return strconv.Atoi(jobIDStr)
}