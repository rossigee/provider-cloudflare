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
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/rossigee/provider-cloudflare/apis/logpush/v1alpha1"
)

// MockLogpushJobAPI implements the LogpushJobAPI interface for testing
type MockLogpushJobAPI struct {
	MockAccounts           func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error)
	MockCreateLogpushJob   func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error)
	MockGetLogpushJob      func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error)
	MockUpdateLogpushJob   func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLogpushJobParams) error
	MockDeleteLogpushJob   func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error
	MockListLogpushJobs    func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error)
}

func (m *MockLogpushJobAPI) Accounts(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
	if m.MockAccounts != nil {
		return m.MockAccounts(ctx, params)
	}
	return []cloudflare.Account{}, cloudflare.ResultInfo{}, nil
}

func (m *MockLogpushJobAPI) CreateLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error) {
	if m.MockCreateLogpushJob != nil {
		return m.MockCreateLogpushJob(ctx, rc, params)
	}
	return &cloudflare.LogpushJob{}, nil
}

func (m *MockLogpushJobAPI) GetLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error) {
	if m.MockGetLogpushJob != nil {
		return m.MockGetLogpushJob(ctx, rc, jobID)
	}
	return cloudflare.LogpushJob{}, nil
}

func (m *MockLogpushJobAPI) UpdateLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLogpushJobParams) error {
	if m.MockUpdateLogpushJob != nil {
		return m.MockUpdateLogpushJob(ctx, rc, params)
	}
	return nil
}

func (m *MockLogpushJobAPI) DeleteLogpushJob(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error {
	if m.MockDeleteLogpushJob != nil {
		return m.MockDeleteLogpushJob(ctx, rc, jobID)
	}
	return nil
}

func (m *MockLogpushJobAPI) ListLogpushJobs(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error) {
	if m.MockListLogpushJobs != nil {
		return m.MockListLogpushJobs(ctx, rc, params)
	}
	return []cloudflare.LogpushJob{}, nil
}

func TestGetAccountID(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client    *MockLogpushJobAPI
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
				client:    &MockLogpushJobAPI{},
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
				client: &MockLogpushJobAPI{
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
				client: &MockLogpushJobAPI{
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
				client: &MockLogpushJobAPI{
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
			client := &JobClient{
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
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.JobParameters
	}

	type want struct {
		obs *v1alpha1.JobObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"CreateLogpushJobSuccess": {
			reason: "Create should create logpush job when API call succeeds",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error) {
						if rc.Identifier != "test-account-id" {
							return nil, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return nil, errors.New("wrong resource type")
						}
						if params.Dataset != "http_requests" {
							return nil, errors.New("wrong dataset")
						}
						if params.Name != "test-job" {
							return nil, errors.New("wrong name")
						}
						if params.DestinationConf != "s3://bucket/path" {
							return nil, errors.New("wrong destination")
						}
						lastComplete := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						return &cloudflare.LogpushJob{
							ID:              123,
							Dataset:         params.Dataset,
							Name:            params.Name,
							DestinationConf: params.DestinationConf,
							Enabled:         params.Enabled,
							Kind:            params.Kind,
							LogpullOptions:  params.LogpullOptions,
							Frequency:       params.Frequency,
							LastComplete:    &lastComplete,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(true),
					Kind:            ptr.To("edge"),
					LogpullOptions:  ptr.To("fields=RayID,EdgeStartTimestamp"),
					Frequency:       ptr.To("high"),
				},
			},
			want: want{
				obs: &v1alpha1.JobObservation{
					ID:              ptr.To(123),
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(true),
					Kind:            ptr.To("edge"),
					LogpullOptions:  ptr.To("fields=RayID,EdgeStartTimestamp"),
					Frequency:       ptr.To("high"),
					LastComplete:    &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
				},
				err: nil,
			},
		},
		"CreateLogpushJobMinimal": {
			reason: "Create should create logpush job with minimal parameters",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error) {
						return &cloudflare.LogpushJob{
							ID:              456,
							Dataset:         params.Dataset,
							Name:            params.Name,
							DestinationConf: params.DestinationConf,
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "dns_logs",
					Name:            "minimal-job",
					DestinationConf: "gcs://bucket/path",
				},
			},
			want: want{
				obs: &v1alpha1.JobObservation{
					ID:              ptr.To(456),
					Dataset:         "dns_logs",
					Name:            "minimal-job",
					DestinationConf: "gcs://bucket/path",
				},
				err: nil,
			},
		},
		"CreateLogpushJobAccountError": {
			reason: "Create should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"CreateLogpushJobAPIError": {
			reason: "Create should return wrapped error when API call fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockCreateLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreateLogpushJobParams) (*cloudflare.LogpushJob, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errCreateJob),
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
	jobID := 123

	type fields struct {
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx   context.Context
		jobID int
	}

	type want struct {
		obs *v1alpha1.JobObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"GetLogpushJobSuccess": {
			reason: "Get should return logpush job when API call succeeds",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error) {
						if rc.Identifier != "test-account-id" {
							return cloudflare.LogpushJob{}, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return cloudflare.LogpushJob{}, errors.New("wrong resource type")
						}
						if jobID != 123 {
							return cloudflare.LogpushJob{}, errors.New("wrong job ID")
						}
						lastComplete := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						lastError := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)
						return cloudflare.LogpushJob{
							ID:                        123,
							Dataset:                   "http_requests",
							Name:                      "test-job",
							DestinationConf:           "s3://bucket/path",
							Enabled:                   true,
							Kind:                      "edge",
							LogpullOptions:            "fields=RayID,EdgeStartTimestamp",
							Frequency:                 "high",
							LastComplete:              &lastComplete,
							LastError:                 &lastError,
							ErrorMessage:              "test error",
							MaxUploadBytes:            1000000,
							MaxUploadRecords:          1000,
							MaxUploadIntervalSeconds:  300,
							OwnershipChallenge:        "challenge-token",
						}, nil
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				obs: &v1alpha1.JobObservation{
					ID:                        ptr.To(123),
					Dataset:                   "http_requests",
					Name:                      "test-job",
					DestinationConf:           "s3://bucket/path",
					Enabled:                   ptr.To(true),
					Kind:                      ptr.To("edge"),
					LogpullOptions:            ptr.To("fields=RayID,EdgeStartTimestamp"),
					Frequency:                 ptr.To("high"),
					LastComplete:              &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
					LastError:                 &metav1.Time{Time: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)},
					ErrorMessage:              ptr.To("test error"),
					MaxUploadBytes:            ptr.To(1000000),
					MaxUploadRecords:          ptr.To(1000),
					MaxUploadIntervalSeconds:  ptr.To(300),
					OwnershipChallenge:        ptr.To("challenge-token"),
				},
				err: nil,
			},
		},
		"GetLogpushJobMinimal": {
			reason: "Get should return logpush job with minimal fields",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error) {
						return cloudflare.LogpushJob{
							ID:              456,
							Dataset:         "dns_logs",
							Name:            "minimal-job",
							DestinationConf: "gcs://bucket/path",
						}, nil
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: 456,
			},
			want: want{
				obs: &v1alpha1.JobObservation{
					ID:              ptr.To(456),
					Dataset:         "dns_logs",
					Name:            "minimal-job",
					DestinationConf: "gcs://bucket/path",
				},
				err: nil,
			},
		},
		"GetLogpushJobAccountError": {
			reason: "Get should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"GetLogpushJobAPIError": {
			reason: "Get should return wrapped error when API call fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockGetLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error) {
						return cloudflare.LogpushJob{}, errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errGetJob),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Get(tc.args.ctx, tc.args.jobID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nGet(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nGet(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")
	jobID := 123

	type fields struct {
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx    context.Context
		jobID  int
		params v1alpha1.JobParameters
	}

	type want struct {
		obs *v1alpha1.JobObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"UpdateLogpushJobSuccess": {
			reason: "Update should update logpush job when API call succeeds",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockUpdateLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLogpushJobParams) error {
						if rc.Identifier != "test-account-id" {
							return errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return errors.New("wrong resource type")
						}
						if params.ID != 123 {
							return errors.New("wrong job ID")
						}
						if params.Name != "updated-job" {
							return errors.New("wrong name")
						}
						return nil
					},
					MockGetLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) (cloudflare.LogpushJob, error) {
						return cloudflare.LogpushJob{
							ID:              123,
							Dataset:         "http_requests",
							Name:            "updated-job",
							DestinationConf: "s3://updated-bucket/path",
							Enabled:         false,
						}, nil
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "updated-job",
					DestinationConf: "s3://updated-bucket/path",
					Enabled:         ptr.To(false),
				},
			},
			want: want{
				obs: &v1alpha1.JobObservation{
					ID:              ptr.To(123),
					Dataset:         "http_requests",
					Name:            "updated-job",
					DestinationConf: "s3://updated-bucket/path",
				},
				err: nil,
			},
		},
		"UpdateLogpushJobAccountError": {
			reason: "Update should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "updated-job",
					DestinationConf: "s3://updated-bucket/path",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"UpdateLogpushJobAPIError": {
			reason: "Update should return wrapped error when API call fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockUpdateLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdateLogpushJobParams) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "updated-job",
					DestinationConf: "s3://updated-bucket/path",
				},
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errUpdateJob),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			got, err := client.Update(tc.args.ctx, tc.args.jobID, tc.args.params)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.obs, got); diff != "" {
				t.Errorf("\n%s\nUpdate(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")
	jobID := 123

	type fields struct {
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx   context.Context
		jobID int
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
		"DeleteLogpushJobSuccess": {
			reason: "Delete should succeed when API call succeeds",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error {
						if rc.Identifier != "test-account-id" {
							return errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return errors.New("wrong resource type")
						}
						if jobID != 123 {
							return errors.New("wrong job ID")
						}
						return nil
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteLogpushJobNotFound": {
			reason: "Delete should succeed when logpush job is not found (already deleted)",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error {
						return errors.New("job not found")
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				err: nil,
			},
		},
		"DeleteLogpushJobAccountError": {
			reason: "Delete should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return nil, cloudflare.ResultInfo{}, errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				err: errors.Wrap(errors.Wrap(errBoom, "failed to list accounts"), "failed to get account ID"),
			},
		},
		"DeleteLogpushJobAPIError": {
			reason: "Delete should return wrapped error when API call fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockDeleteLogpushJob: func(ctx context.Context, rc *cloudflare.ResourceContainer, jobID int) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx:   context.Background(),
				jobID: jobID,
			},
			want: want{
				err: errors.Wrap(errBoom, errDeleteJob),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.fields.client)
			err := client.Delete(tc.args.ctx, tc.args.jobID)
			
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("\n%s\nDelete(...): -want error, +got error:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestList(t *testing.T) {
	errBoom := errors.New("boom")

	type fields struct {
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx context.Context
	}

	type want struct {
		obs []v1alpha1.JobObservation
		err error
	}

	cases := map[string]struct {
		reason string
		fields fields
		args   args
		want   want
	}{
		"ListLogpushJobsSuccess": {
			reason: "List should return logpush jobs when API call succeeds",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListLogpushJobs: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error) {
						if rc.Identifier != "test-account-id" {
							return nil, errors.New("wrong account ID")
						}
						if rc.Type != cloudflare.AccountType {
							return nil, errors.New("wrong resource type")
						}
						lastComplete1 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						lastComplete2 := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
						return []cloudflare.LogpushJob{
							{
								ID:              123,
								Dataset:         "http_requests",
								Name:            "job-1",
								DestinationConf: "s3://bucket1/path",
								Enabled:         true,
								LastComplete:    &lastComplete1,
							},
							{
								ID:              456,
								Dataset:         "dns_logs",
								Name:            "job-2",
								DestinationConf: "gcs://bucket2/path",
								Enabled:         false,
								LastComplete:    &lastComplete2,
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: []v1alpha1.JobObservation{
					{
						ID:              ptr.To(123),
						Dataset:         "http_requests",
						Name:            "job-1",
						DestinationConf: "s3://bucket1/path",
						Enabled:         ptr.To(true),
						LastComplete:    &metav1.Time{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
					{
						ID:              ptr.To(456),
						Dataset:         "dns_logs",
						Name:            "job-2",
						DestinationConf: "gcs://bucket2/path",
						LastComplete:    &metav1.Time{Time: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)},
					},
				},
				err: nil,
			},
		},
		"ListLogpushJobsEmpty": {
			reason: "List should return empty list when no jobs exist",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListLogpushJobs: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error) {
						return []cloudflare.LogpushJob{}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: []v1alpha1.JobObservation{},
				err: nil,
			},
		},
		"ListLogpushJobsAccountError": {
			reason: "List should return wrapped error when account lookup fails",
			fields: fields{
				client: &MockLogpushJobAPI{
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
		"ListLogpushJobsAPIError": {
			reason: "List should return wrapped error when API call fails",
			fields: fields{
				client: &MockLogpushJobAPI{
					MockAccounts: func(ctx context.Context, params cloudflare.AccountsListParams) ([]cloudflare.Account, cloudflare.ResultInfo, error) {
						return []cloudflare.Account{
							{ID: "test-account-id", Name: "Test Account"},
						}, cloudflare.ResultInfo{}, nil
					},
					MockListLogpushJobs: func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListLogpushJobsParams) ([]cloudflare.LogpushJob, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				obs: nil,
				err: errors.Wrap(errBoom, errListJobs),
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
		client *MockLogpushJobAPI
	}

	type args struct {
		ctx    context.Context
		params v1alpha1.JobParameters
		obs    v1alpha1.JobObservation
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
			reason: "IsUpToDate should return true when all key fields match",
			fields: fields{
				client: &MockLogpushJobAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(true),
				},
				obs: v1alpha1.JobObservation{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(true),
				},
			},
			want: want{
				upToDate: true,
				err:      nil,
			},
		},
		"IsUpToDateFalseName": {
			reason: "IsUpToDate should return false when name doesn't match",
			fields: fields{
				client: &MockLogpushJobAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "updated-job",
					DestinationConf: "s3://bucket/path",
				},
				obs: v1alpha1.JobObservation{
					Dataset:         "http_requests",
					Name:            "original-job",
					DestinationConf: "s3://bucket/path",
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseDataset": {
			reason: "IsUpToDate should return false when dataset doesn't match",
			fields: fields{
				client: &MockLogpushJobAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "dns_logs",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
				},
				obs: v1alpha1.JobObservation{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseDestination": {
			reason: "IsUpToDate should return false when destination doesn't match",
			fields: fields{
				client: &MockLogpushJobAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "gcs://different-bucket/path",
				},
				obs: v1alpha1.JobObservation{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
				},
			},
			want: want{
				upToDate: false,
				err:      nil,
			},
		},
		"IsUpToDateFalseEnabled": {
			reason: "IsUpToDate should return false when enabled status doesn't match",
			fields: fields{
				client: &MockLogpushJobAPI{},
			},
			args: args{
				ctx: context.Background(),
				params: v1alpha1.JobParameters{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(false),
				},
				obs: v1alpha1.JobObservation{
					Dataset:         "http_requests",
					Name:            "test-job",
					DestinationConf: "s3://bucket/path",
					Enabled:         ptr.To(true),
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

func TestIsJobNotFound(t *testing.T) {
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
			reason: "IsJobNotFound should return false for nil error",
			args: args{
				err: nil,
			},
			want: want{
				isNotFound: false,
			},
		},
		"JobNotFoundError": {
			reason: "IsJobNotFound should return true for 'job not found' error",
			args: args{
				err: errors.New("job not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"404Error": {
			reason: "IsJobNotFound should return true for '404' error",
			args: args{
				err: errors.New("404"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"NotFoundError": {
			reason: "IsJobNotFound should return true for 'Not found' error",
			args: args{
				err: errors.New("Not found"),
			},
			want: want{
				isNotFound: true,
			},
		},
		"OtherError": {
			reason: "IsJobNotFound should return false for other errors",
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
			got := IsJobNotFound(tc.args.err)
			if diff := cmp.Diff(tc.want.isNotFound, got); diff != "" {
				t.Errorf("\n%s\nIsJobNotFound(...): -want, +got:\n%s\n", tc.reason, diff)
			}
		})
	}
}

func TestParseJobID(t *testing.T) {
	type args struct {
		jobIDStr string
	}

	type want struct {
		jobID int
		err   error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"ParseJobIDSuccess": {
			reason: "ParseJobID should parse valid job ID string",
			args: args{
				jobIDStr: "123",
			},
			want: want{
				jobID: 123,
				err:   nil,
			},
		},
		"ParseJobIDZero": {
			reason: "ParseJobID should parse zero job ID",
			args: args{
				jobIDStr: "0",
			},
			want: want{
				jobID: 0,
				err:   nil,
			},
		},
		"ParseJobIDInvalid": {
			reason: "ParseJobID should return error for invalid job ID string",
			args: args{
				jobIDStr: "invalid",
			},
			want: want{
				jobID: 0,
				err:   nil, // We'll check for error existence separately
			},
		},
		"ParseJobIDEmpty": {
			reason: "ParseJobID should return error for empty job ID string",
			args: args{
				jobIDStr: "",
			},
			want: want{
				jobID: 0,
				err:   nil, // We'll check for error existence separately
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseJobID(tc.args.jobIDStr)
			
			// For invalid cases, just check that an error occurred
			if tc.args.jobIDStr == "invalid" || tc.args.jobIDStr == "" {
				if err == nil {
					t.Errorf("\n%s\nParseJobID(...): expected error but got none\n", tc.reason)
				}
				if got != tc.want.jobID {
					t.Errorf("\n%s\nParseJobID(...): got %d, want %d\n", tc.reason, got, tc.want.jobID)
				}
			} else {
				if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
					t.Errorf("\n%s\nParseJobID(...): -want error, +got error:\n%s\n", tc.reason, diff)
				}
				if diff := cmp.Diff(tc.want.jobID, got); diff != "" {
					t.Errorf("\n%s\nParseJobID(...): -want, +got:\n%s\n", tc.reason, diff)
				}
			}
		})
	}
}