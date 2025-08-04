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

package bucket

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/r2/v1alpha1"
)

// R2BucketAPI defines the interface for R2 Bucket operations
type R2BucketAPI interface {
	Accounts(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error)
	CreateR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error)
	GetR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error)
	DeleteR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error
	ListR2Buckets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error)
}

const (
	errCreateBucket = "cannot create R2 bucket"
	errUpdateBucket = "cannot update R2 bucket"
	errGetBucket    = "cannot get R2 bucket"
	errDeleteBucket = "cannot delete R2 bucket"
	errListBuckets  = "cannot list R2 buckets"
)

// BucketClient provides operations for R2 Buckets.
type BucketClient struct {
	client    R2BucketAPI
	accountID string
}

// NewClient creates a new R2 Bucket client.
func NewClient(client R2BucketAPI) *BucketClient {
	return &BucketClient{
		client:    client,
		accountID: "", // Account ID will be retrieved when needed
	}
}

// getAccountID gets the account ID from the Cloudflare API
func (c *BucketClient) getAccountID(ctx context.Context) (string, error) {
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

// convertToObservation converts cloudflare-go R2 bucket to Crossplane observation.
func convertToObservation(bucket cloudflare.R2Bucket) v1alpha1.BucketObservation {
	obs := v1alpha1.BucketObservation{
		Name:     bucket.Name,
		Location: bucket.Location,
	}

	if bucket.CreationDate != nil {
		obs.CreationDate = &metav1.Time{Time: *bucket.CreationDate}
	}

	return obs
}

// convertToCloudflareParams converts Crossplane parameters to cloudflare-go parameters.
func convertToCloudflareParams(params v1alpha1.BucketParameters) cloudflare.CreateR2BucketParameters {
	cfParams := cloudflare.CreateR2BucketParameters{
		Name: params.Name,
	}

	if params.LocationHint != nil {
		cfParams.LocationHint = *params.LocationHint
	}

	return cfParams
}

// Create creates a new R2 Bucket.
func (c *BucketClient) Create(ctx context.Context, params v1alpha1.BucketParameters) (*v1alpha1.BucketObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)
	
	createParams := convertToCloudflareParams(params)
	
	bucket, err := c.client.CreateR2Bucket(ctx, rc, createParams)
	if err != nil {
		return nil, errors.Wrap(err, errCreateBucket)
	}

	obs := convertToObservation(bucket)
	return &obs, nil
}

// Get retrieves an R2 Bucket.
func (c *BucketClient) Get(ctx context.Context, bucketName string) (*v1alpha1.BucketObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	bucket, err := c.client.GetR2Bucket(ctx, rc, bucketName)
	if err != nil {
		return nil, errors.Wrap(err, errGetBucket)
	}

	obs := convertToObservation(bucket)
	return &obs, nil
}

// Delete removes an R2 Bucket.
func (c *BucketClient) Delete(ctx context.Context, bucketName string) error {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	err = c.client.DeleteR2Bucket(ctx, rc, bucketName)
	if err != nil && !IsBucketNotFound(err) {
		return errors.Wrap(err, errDeleteBucket)
	}

	return nil
}

// List retrieves all R2 Buckets.
func (c *BucketClient) List(ctx context.Context) ([]v1alpha1.BucketObservation, error) {
	accountID, err := c.getAccountID(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get account ID")
	}
	rc := cloudflare.AccountIdentifier(accountID)

	buckets, err := c.client.ListR2Buckets(ctx, rc, cloudflare.ListR2BucketsParams{})
	if err != nil {
		return nil, errors.Wrap(err, errListBuckets)
	}

	observations := make([]v1alpha1.BucketObservation, len(buckets))
	for i, bucket := range buckets {
		observations[i] = convertToObservation(bucket)
	}

	return observations, nil
}

// IsUpToDate checks if the R2 Bucket is up to date.
func (c *BucketClient) IsUpToDate(ctx context.Context, params v1alpha1.BucketParameters, obs v1alpha1.BucketObservation) (bool, error) {
	// R2 buckets don't have many updatable properties
	// Main check is if the bucket exists with the correct name
	return obs.Name == params.Name, nil
}

// IsBucketNotFound returns true if the error indicates the bucket was not found
func IsBucketNotFound(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "bucket not found" ||
		err.Error() == "404" ||
		err.Error() == "Not found"
}