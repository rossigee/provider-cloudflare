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

package crontrigger

import (
	"context"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	testAccountID   = "test-account-id"
	testScriptName  = "test-script"
	testCronExpr    = "0 0 * * *"
	testNewCronExpr = "*/5 * * * *"
)

var (
	testTime     = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testMetaTime = metav1.Time{Time: testTime}
)

func TestCreate(t *testing.T) {
	type args struct {
		params v1alpha1.CronTriggerParameters
	}
	type want struct {
		obs *v1alpha1.CronTriggerObservation
		err error
	}

	cases := map[string]struct {
		args        args
		mockClient  func() clients.ClientInterface
		want        want
	}{
		"CreateSuccess": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers").Return([]cloudflare.WorkerCronTrigger{}, nil)
				client.On("UpdateWorkerCronTriggers").Return([]cloudflare.WorkerCronTrigger{
					{
						Cron:       testCronExpr,
						CreatedOn:  &testTime,
						ModifiedOn: &testTime,
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
					CreatedOn:  &testMetaTime,
					ModifiedOn: &testMetaTime,
				},
			},
		},
		"CreateWithExistingTriggers": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: "*/10 * * * *"},
				}, nil)
				client.On("UpdateWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkerCronTriggersParams{
						ScriptName: testScriptName,
						Crons: []cloudflare.WorkerCronTrigger{
							{Cron: "*/10 * * * *"},
							{Cron: testCronExpr},
						},
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: "*/10 * * * *"},
					{Cron: testCronExpr, CreatedOn: &testTime},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
					CreatedOn:  &testMetaTime,
				},
			},
		},
		"CreateListError": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{}, errors.New("list error"))
				return client
			},
			want: want{
				err: errors.New("cannot list workers cron triggers: list error"),
			},
		},
		"CreateUpdateError": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{}, nil)
				client.On("UpdateWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkerCronTriggersParams{
						ScriptName: testScriptName,
						Crons: []cloudflare.WorkerCronTrigger{
							{Cron: testCronExpr},
						},
					},
				).Return([]cloudflare.WorkerCronTrigger{}, errors.New("update error"))
				return client
			},
			want: want{
				err: errors.New("cannot create workers cron trigger: update error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Create(context.Background(), tc.args.params)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("Create() error = %v, want %v", err, tc.want.err)
				}
				return
			}

			if err != nil {
				t.Errorf("Create() unexpected error = %v", err)
				return
			}

			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Create() -want +got:\n%s", diff)
			}
		})
	}
}

func TestGet(t *testing.T) {
	type args struct {
		scriptName     string
		cronExpression string
	}
	type want struct {
		obs *v1alpha1.CronTriggerObservation
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"GetSuccess": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{
						Cron:       testCronExpr,
						CreatedOn:  &testTime,
						ModifiedOn: &testTime,
					},
					{
						Cron: "*/5 * * * *",
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
					CreatedOn:  &testMetaTime,
					ModifiedOn: &testMetaTime,
				},
			},
		},
		"GetNotFound": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: "*/5 * * * *"},
				}, nil)
				return client
			},
			want: want{
				err: errors.New("cron trigger not found"),
			},
		},
		"GetListError": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{}, errors.New("list error"))
				return client
			},
			want: want{
				err: errors.New("cannot get workers cron trigger: list error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Get(context.Background(), tc.args.scriptName, tc.args.cronExpression)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("Get() error = %v, want %v", err, tc.want.err)
				}
				return
			}

			if err != nil {
				t.Errorf("Get() unexpected error = %v", err)
				return
			}

			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Get() -want +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		scriptName     string
		cronExpression string
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"DeleteSuccess": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: testCronExpr},
					{Cron: "*/5 * * * *"},
				}, nil)
				client.On("UpdateWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkerCronTriggersParams{
						ScriptName: testScriptName,
						Crons: []cloudflare.WorkerCronTrigger{
							{Cron: "*/5 * * * *"},
						},
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: "*/5 * * * *"},
				}, nil)
				return client
			},
			want: want{},
		},
		"DeleteNotFound": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: "*/5 * * * *"},
				}, nil)
				return client
			},
			want: want{}, // Not found is OK for delete
		},
		"DeleteError": {
			args: args{
				scriptName:     testScriptName,
				cronExpression: testCronExpr,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkerCronTriggersParams{
						ScriptName: testScriptName,
					},
				).Return([]cloudflare.WorkerCronTrigger{
					{Cron: testCronExpr},
				}, nil)
				client.On("UpdateWorkerCronTriggers", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkerCronTriggersParams{
						ScriptName: testScriptName,
						Crons:      []cloudflare.WorkerCronTrigger{},
					},
				).Return([]cloudflare.WorkerCronTrigger{}, errors.New("update error"))
				return client
			},
			want: want{
				err: errors.New("cannot delete workers cron trigger: update error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			err := client.Delete(context.Background(), tc.args.scriptName, tc.args.cronExpression)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("Delete() error = %v, want %v", err, tc.want.err)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		params v1alpha1.CronTriggerParameters
		obs    v1alpha1.CronTriggerObservation
	}
	type want struct {
		isUpToDate bool
	}

	cases := map[string]struct {
		args args
		want want
	}{
		"UpToDate": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
				obs: v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"CronChanged": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: testScriptName,
					Cron:       testNewCronExpr,
				},
				obs: v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
		"ScriptChanged": {
			args: args{
				params: v1alpha1.CronTriggerParameters{
					ScriptName: "different-script",
					Cron:       testCronExpr,
				},
				obs: v1alpha1.CronTriggerObservation{
					ScriptName: testScriptName,
					Cron:       testCronExpr,
				},
			},
			want: want{
				isUpToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(&clients.MockClient{})
			isUpToDate, err := client.IsUpToDate(context.Background(), tc.args.params, tc.args.obs)

			if err != nil {
				t.Errorf("IsUpToDate() unexpected error = %v", err)
				return
			}

			if isUpToDate != tc.want.isUpToDate {
				t.Errorf("IsUpToDate() = %v, want %v", isUpToDate, tc.want.isUpToDate)
			}
		})
	}
}