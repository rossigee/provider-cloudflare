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
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/r2/v1alpha1"
)

// MockR2BucketAPI implements the R2BucketAPI interface for testing
type MockR2BucketAPI struct {
	MockAccounts        func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error)
	MockCreateR2Bucket  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error)
	MockGetR2Bucket     func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error)
	MockDeleteR2Bucket  func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error
	MockListR2Buckets   func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error)
}

func (m *MockR2BucketAPI) Accounts(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
	if m.MockAccounts != nil {
		return m.MockAccounts(ctx, params)
	}
	return []cloudflare.Account{}, cloudflare.ResultInfo{}, nil
}

func (m *MockR2BucketAPI) CreateR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error) {
	if m.MockCreateR2Bucket != nil {
		return m.MockCreateR2Bucket(ctx, rc, params)
	}
	return cloudflare.R2Bucket{}, nil
}

func (m *MockR2BucketAPI) GetR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error) {
	if m.MockGetR2Bucket != nil {
		return m.MockGetR2Bucket(ctx, rc, bucketName)
	}
	return cloudflare.R2Bucket{}, nil
}

func (m *MockR2BucketAPI) DeleteR2Bucket(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error {
	if m.MockDeleteR2Bucket != nil {
		return m.MockDeleteR2Bucket(ctx, rc, bucketName)
	}
	return nil
}

func (m *MockR2BucketAPI) ListR2Buckets(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error) {
	if m.MockListR2Buckets != nil {
		return m.MockListR2Buckets(ctx, rc, params)
	}
	return []cloudflare.R2Bucket{}, nil
}

func TestGetAccountID(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client    *MockR2BucketAPI
		accountID string
	}

	type args struct {
		ctx context.Context
	}

	type want struct {
		accountID string
		err       error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetAccountIDCached": {
			reason: "getAccountID should return cached account ID when available",
			fields: fields{
				client:    &MockR2BucketAPI{},
				accountID: "cached-account-id",
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				accountID: "cached-account-id",
				err:       nil,
			},
		},
		"GetAccountIDSuccess": {
			reason: "getAccountID should return account ID when API call succeeds",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
				},
				accountID: "",
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				accountID: "test-account-id",
				err:       nil,
			},
		},
		"GetAccountIDNoAccounts": {
			reason: "getAccountID should return error when no accounts found",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{}, cloudflare.ResultInfo{}, nil
					},
				},
				accountID: "",
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				accountID: "",
				err:       errors.New("no accounts found"),
			},
		},
		"GetAccountIDAPIError": {
			reason: "getAccountID should return wrapped error when API call fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
				accountID: "",
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				accountID: "",
				err:       errors.Wrap(errBoom, "failed to list accounts"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := &BucketClient{
				client:    tc.fields.client,
				accountID: tc.fields.accountID,
			}
			got, err := client.getAccountID(tc.args.ctx)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\ngetAccountID(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.accountID, got); diff != "" {
				t.Errorf("\n%s\ngetAccountID(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client *MockR2BucketAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.BucketParameters
	}

	type want struct {
		obs *v1alpha1.BucketObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateR2BucketSuccess": {
			reason: "Create should create R2 bucket when API call succeeds",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.R2Bucket{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.R2Bucket{}, errors.New("wrong resource type")
						}
						if params.Name != "test-bucket" {
							return cloudflare.R2Bucket{}, errors.New("wrong bucket name")
						}
						if params.LocationHint != "ENAM" {
							return cloudflare.R2Bucket{}, errors.New("wrong location hint")
						}
						creationDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						return cloudflare.R2Bucket{
							Name:         params.Name,
							Location:     "ENAM",
							CreationDate: &creationDate,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name:         "test-bucket",
					LocationHint: ptr.To("ENAM"),
				},
			},
			want: want{
				obs: &v1alpha1.BucketObservation{
					Name:         "test-bucket",
					Location:     "ENAM",
					CreationDate: &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"CreateR2BucketMinimal": {
			reason: "Create should create R2 bucket with minimal parameters",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error) {
						return cloudflare.R2Bucket{
							Name:     params.Name,
							Location: "auto",
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name: "minimal-bucket",
				},
			},
			want: want{
				obs: &v1alpha1.BucketObservation{
					Name:     "minimal-bucket",
					Location: "auto",
				},
				err: nil,
			},
		},
		"CreateR2BucketAccountError": {
			reason: "Create should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name: "test-bucket",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"CreateR2BucketAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateR2BucketParameters) (cloudflare.R2Bucket, error) {
						return cloudflare.R2Bucket{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name: "test-bucket",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errCreateBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Create(tc.args.ctx, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nCreate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestGet(t *testing.T) {
	errBoom := errors.New("boom")
	bucketName := "test-bucket"

	type fields struct {
		client *MockR2BucketAPI
	}

	type args struct {
		ctx        context.Context
		bucketName string
	}

	type want struct {
		obs *v1alpha1.BucketObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetR2BucketSuccess": {
			reason: "Get should return R2 bucket when API call succeeds",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.R2Bucket{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.R2Bucket{}, errors.New("wrong resource type")
						}
						if bucketName != "test-bucket" {
							return cloudflare.R2Bucket{}, errors.New("wrong bucket name")
						}
						creationDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						return cloudflare.R2Bucket{
							Name:         bucketName,
							Location:     "ENAM",
							CreationDate: &creationDate,
						}, nil
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				obs: &v1alpha1.BucketObservation{
					Name:         "test-bucket",
					Location:     "ENAM",
					CreationDate: &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"GetR2BucketMinimal": {
			reason: "Get should return R2 bucket with minimal fields",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error) {
						return cloudflare.R2Bucket{
							Name:     bucketName,
							Location: "auto",
						}, nil
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: "minimal-bucket",
			},
			want: want{
				obs: &v1alpha1.BucketObservation{
					Name:     "minimal-bucket",
					Location: "auto",
				},
				err: nil,
			},
		},
		"GetR2BucketAccountError": {
			reason: "Get should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"GetR2BucketAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) (cloudflare.R2Bucket, error) {
						return cloudflare.R2Bucket{}, errBoom
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errGetBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.bucketName)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGet(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nGet(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")
	bucketName := "test-bucket"

	type fields struct {
		client *MockR2BucketAPI
	}

	type args struct {
		ctx        context.Context
		bucketName string
	}

	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"DeleteR2BucketSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error {
						if rc.Identifier != "test-account-id" {
							return errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return errors.New("wrong resource type")
						}
						if bucketName != "test-bucket" {
							return errors.New("wrong bucket name")
						}
						return nil
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteR2BucketNotFound": {
			reason: "Delete should succeed when R2 bucket is not found (already deleted)",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error {
						return errors.New("bucket not found")
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteR2BucketAccountError": {
			reason: "Delete should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"DeleteR2BucketAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteR2Bucket: func(ctx context.Context, rc *cloudflare.ResourceContainer, bucketName string) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:        context.Background(),
				bucketName: bucketName,
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteBucket),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.bucketName)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestList(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client *MockR2BucketAPI
	}

	type args struct {
		ctx context.Context
	}

	type want struct {
		obs []v1alpha1.BucketObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ListR2BucketsSuccess": {
			reason: "List should return R2 buckets when API call succeeds",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListR2Buckets: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error) {
						if rc.Identifier != "test-account-id" {
							return nil, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return nil, errors.New("wrong resource type")
						}
						creationDate1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						creationDate2 := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
						return []cloudflare.R2Bucket{
							{
								Name:         "bucket-1",
								Location:     "ENAM",
								CreationDate: &creationDate1,
							},
							{
								Name:         "bucket-2",
								Location:     "WNAM",
								CreationDate: &creationDate2,
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: []v1alpha1.BucketObservation{
					{
						Name:         "bucket-1",
						Location:     "ENAM",
						CreationDate: &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
					{
						Name:         "bucket-2",
						Location:     "WNAM",
						CreationDate: &metav1.Time{Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)},
					},
				},
				err: nil,
			},
		},
		"ListR2BucketsEmpty": {
			reason: "List should return empty list when no buckets exist",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListR2Buckets: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error) {
						return []cloudflare.R2Bucket{}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: []v1alpha1.BucketObservation{},
				err: nil,
			},
		},
		"ListR2BucketsAccountError": {
			reason: "List should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"ListR2BucketsAPIError": {
			reason: "List should return wrapped error when API call fails",
			fields: fields{
				client: &MockR2BucketAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListR2Buckets: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListR2BucketsParams) ([]cloudflare.R2Bucket, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errListBuckets),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.List(tc.args.ctx)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nList(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nList(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type fields struct {
		client *MockR2BucketAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.BucketParameters
		obs    v1alpha1.BucketObservation
	}

	type want struct {
		upToDate bool
		err      error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"IsUpToDateTrue": {
			reason: "IsUpToDate should return true when bucket names match",
			fields: fields{
				client: &MockR2BucketAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name: "test-bucket",
				},
				obs: v1alpha1.BucketObservation{
					Name:     "test-bucket",
					Location: "ENAM",
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalse": {
			reason: "IsUpToDate should return false when bucket names don't match",
			fields: fields{
				client: &MockR2BucketAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.BucketParameters{
					Name: "different-bucket",
				},
				obs: v1alpha1.BucketObservation{
					Name:     "test-bucket",
					Location: "ENAM",
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.IsUpToDate(tc.args.ctx, tc.args.params, tc.args.obs)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.upToDate, got); diff != "" {
				t.Errorf("\n%s\nIsUpToDate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertToObservation(t *testing.T) {
	type args struct {
		bucket cloudflare.R2Bucket
	}

	type want struct {
		obs v1alpha1.BucketObservation
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertFullBucket": {
			reason: "convertToObservation should convert all fields correctly",
			args: args{
				bucket: cloudflare.R2Bucket{
					Name:         "full-bucket",
					Location:     "ENAM",
					CreationDate: &time.Time{},
				},
			},
			want: want{
				obs: v1alpha1.BucketObservation{
					Name:         "full-bucket",
					Location:     "ENAM",
					CreationDate: &metav1.Time{Time: time.Time{}},
				},
			},
		},
		"ConvertMinimalBucket": {
			reason: "convertToObservation should handle minimal bucket",
			args: args{
				bucket: cloudflare.R2Bucket{
					Name:     "minimal-bucket",
					Location: "auto",
				},
			},
			want: want{
				obs: v1alpha1.BucketObservation{
					Name:     "minimal-bucket",
					Location: "auto",
				},
			},
		},
		"ConvertBucketWithCreationDate": {
			reason: "convertToObservation should handle creation date correctly",
			args: args{
				bucket: cloudflare.R2Bucket{
					Name:         "date-bucket",
					Location:     "WNAM",
					CreationDate: &time.Time{},
				},
			},
			want: want{
				obs: v1alpha1.BucketObservation{
					Name:         "date-bucket",
					Location:     "WNAM",
					CreationDate: &metav1.Time{Time: time.Time{}},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertToObservation(tc.args.bucket)
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nconvertToObservation(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestConvertToCloudflareParams(t *testing.T) {
	type args struct {
		params v1alpha1.BucketParameters
	}

	type want struct {
		cfParams cloudflare.CreateR2BucketParameters
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ConvertAllParameters": {
			reason: "convertToCloudflareParams should convert all parameters correctly",
			args: args{
				params: v1alpha1.BucketParameters{
					Name:         "test-bucket",
					LocationHint: ptr.To("ENAM"),
				},
			},
			want: want{
				cfParams: cloudflare.CreateR2BucketParameters{
					Name:         "test-bucket",
					LocationHint: "ENAM",
				},
			},
		},
		"ConvertMinimalParameters": {
			reason: "convertToCloudflareParams should handle minimal parameters",
			args: args{
				params: v1alpha1.BucketParameters{
					Name: "minimal-bucket",
				},
			},
			want: want{
				cfParams: cloudflare.CreateR2BucketParameters{
					Name: "minimal-bucket",
				},
			},
		},
		"ConvertWNAMParameters": {
			reason: "convertToCloudflareParams should handle WNAM location hint",
			args: args{
				params: v1alpha1.BucketParameters{
					Name:         "wnam-bucket",
					LocationHint: ptr.To("WNAM"),
				},
			},
			want: want{
				cfParams: cloudflare.CreateR2BucketParameters{
					Name:         "wnam-bucket",
					LocationHint: "WNAM",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := convertToCloudflareParams(tc.args.params)
			if diff := cmp.Diff(tc.want.cfParams, got); diff != "" {
				t.Errorf("\n%s\nconvertToCloudflareParams(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestIsBucketNotFound(t *testing.T) {
	type args struct {
		err error
	}

	type want struct {
		isNotFound bool
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"NilError": {
			reason: "IsBucketNotFound should return false for nil error",
			args: args{
				err: nil,
			},
			want: want{
				isNotFound: false,
			},
		},
		"BucketNotFoundError": {
			reason: "IsBucketNotFound should return true for 'bucket not found' error",
			args: args{
				err: errors.New("bucket not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"404Error": {
			reason: "IsBucketNotFound should return true for '404' error",
			args: args{
				err: errors.New("404"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"NotFoundError": {
			reason: "IsBucketNotFound should return true for 'Not found' error",
			args: args{
				err: errors.New("Not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "IsBucketNotFound should return false for other errors",
			args: args{
				err: errors.New("some other error"),
			},
			want: want{
				isNotFound: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsBucketNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.isNotFound, got); diff != "" {
				t.Errorf("\n%s\nIsBucketNotFound(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}