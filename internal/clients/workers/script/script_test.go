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

package script

import (
	"context"
	"testing"
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	testAccountID = "test-account-id"
	testScriptName = "test-script"
	testScript = `
		addEventListener('fetch', event => {
			event.respondWith(new Response('Hello World!'))
		})
	`
)

var (
	testTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	testMetaTime = metav1.Time{Time: testTime}
)

func TestCreate(t *testing.T) {
	type args struct {
		params v1alpha1.ScriptParameters
	}
	type want struct {
		obs *v1alpha1.ScriptObservation
		err error
	}

	cases := map[string]struct {
		args        args
		mockClient  func() clients.ClientInterface
		want        want
	}{
		"CreateSuccess": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName: testScriptName,
					Script:     testScript,
					Module:     ptr.To(false),
					Logpush:    ptr.To(true),
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("UploadWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.CreateWorkerParams{
						ScriptName: testScriptName,
						Script:     testScript,
						Module:     false,
						Logpush:    ptr.To(true),
						Bindings:   map[string]cloudflare.WorkerBinding{},
						Tags:       []string{},
					},
				).Return(cloudflare.WorkerScriptResponse{
					WorkerScript: cloudflare.WorkerScript{
						WorkerMetaData: cloudflare.WorkerMetaData{
							ID:         "test-id",
							ETAG:       "test-etag",
							Size:       1024,
							CreatedOn:  testTime,
							ModifiedOn: testTime,
						},
						Script:     testScript,
						UsageModel: "standard",
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.ScriptObservation{
					ID:         "test-id",
					ETAG:       "test-etag", 
					Size:       1024,
					CreatedOn:  &testMetaTime,
					ModifiedOn: &testMetaTime,
					UsageModel: ptr.To("standard"),
				},
			},
		},
		"CreateWithBindings": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName: testScriptName,
					Script:     testScript,
					Bindings: []v1alpha1.WorkerBinding{
						{
							Type:        "kv_namespace",
							Name:        "MY_KV",
							NamespaceID: ptr.To("test-namespace-id"),
						},
						{
							Type: "text_blob",
							Name: "MY_TEXT",
							Text: ptr.To("Hello World"),
						},
					},
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("UploadWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.CreateWorkerParams{
						ScriptName: testScriptName,
						Script:     testScript,
						Bindings: map[string]cloudflare.WorkerBinding{
							"MY_KV": cloudflare.WorkerKvNamespaceBinding{
								NamespaceID: "test-namespace-id",
							},
							"MY_TEXT": cloudflare.WorkerPlainTextBinding{
								Text: "Hello World",
							},
						},
						Tags: []string{},
					},
				).Return(cloudflare.WorkerScriptResponse{
					WorkerScript: cloudflare.WorkerScript{
						WorkerMetaData: cloudflare.WorkerMetaData{
							ID:   "test-id",
							Size: 1024,
						},
						Script: testScript,
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.ScriptObservation{
					ID:   "test-id",
					Size: 1024,
				},
			},
		},
		"CreateError": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName: testScriptName,
					Script:     testScript,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("UploadWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.CreateWorkerParams{
						ScriptName: testScriptName,
						Script:     testScript,
						Bindings:   map[string]cloudflare.WorkerBinding{},
						Tags:       []string{},
					},
				).Return(cloudflare.WorkerScriptResponse{}, errors.New("api error"))
				return client
			},
			want: want{
				err: errors.New("cannot create worker script: api error"),
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
		scriptName string
	}
	type want struct {
		obs *v1alpha1.ScriptObservation
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"GetSuccess": {
			args: args{
				scriptName: testScriptName,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("GetWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(cloudflare.WorkerScriptResponse{
					WorkerScript: cloudflare.WorkerScript{
						Script:     testScript,
						UsageModel: "standard",
					},
				}, nil)
				client.On("GetWorkersScriptSettings", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(cloudflare.WorkerScriptSettingsResponse{
					WorkerMetaData: cloudflare.WorkerMetaData{
						ID:         "test-id",
						ETAG:       "test-etag",
						Size:       1024,
						CreatedOn:  testTime,
						ModifiedOn: testTime,
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.ScriptObservation{
					ID:         "test-id",
					ETAG:       "test-etag",
					Size:       1024,
					CreatedOn:  &testMetaTime,
					ModifiedOn: &testMetaTime,
					UsageModel: ptr.To("standard"),
				},
			},
		},
		"GetError": {
			args: args{
				scriptName: testScriptName,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("GetWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(cloudflare.WorkerScriptResponse{}, errors.New("not found"))
				return client
			},
			want: want{
				err: errors.New("cannot get worker script: not found"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Get(context.Background(), tc.args.scriptName)

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
		scriptName        string
		dispatchNamespace *string
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
				scriptName: testScriptName,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("DeleteWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.DeleteWorkerParams{
						ScriptName: testScriptName,
					},
				).Return(nil)
				return client
			},
			want: want{},
		},
		"DeleteWithDispatchNamespace": {
			args: args{
				scriptName:        testScriptName,
				dispatchNamespace: ptr.To("test-namespace"),
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("DeleteWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.DeleteWorkerParams{
						ScriptName:        testScriptName,
						DispatchNamespace: ptr.To("test-namespace"),
					},
				).Return(nil)
				return client
			},
			want: want{},
		},
		"DeleteError": {
			args: args{
				scriptName: testScriptName,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("DeleteWorker", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.DeleteWorkerParams{
						ScriptName: testScriptName,
					},
				).Return(errors.New("delete failed"))
				return client
			},
			want: want{
				err: errors.New("cannot delete worker script: delete failed"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			err := client.Delete(context.Background(), tc.args.scriptName, tc.args.dispatchNamespace)

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
		params v1alpha1.ScriptParameters
		obs    v1alpha1.ScriptObservation
	}
	type want struct {
		isUpToDate bool
		err        error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"UpToDate": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName:        testScriptName,
					Script:           testScript,
					Logpush:          ptr.To(true),
					CompatibilityDate: ptr.To("2023-01-01"),
				},
				obs: v1alpha1.ScriptObservation{
					ID: "test-id",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("GetWorkersScriptContent", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(testScript, nil)
				client.On("GetWorkersScriptSettings", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(cloudflare.WorkerScriptSettingsResponse{
					WorkerMetaData: cloudflare.WorkerMetaData{
						Logpush: ptr.To(true),
					},
				}, nil)
				return client
			},
			want: want{
				isUpToDate: true,
			},
		},
		"ScriptContentChanged": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName: testScriptName,
					Script:     testScript,
				},
				obs: v1alpha1.ScriptObservation{
					ID: "test-id",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("GetWorkersScriptContent", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return("different script content", nil)
				return client
			},
			want: want{
				isUpToDate: false,
			},
		},
		"LogpushChanged": {
			args: args{
				params: v1alpha1.ScriptParameters{
					ScriptName: testScriptName,
					Script:     testScript,
					Logpush:    ptr.To(true),
				},
				obs: v1alpha1.ScriptObservation{
					ID: "test-id",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("GetWorkersScriptContent", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(testScript, nil)
				client.On("GetWorkersScriptSettings", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testScriptName,
				).Return(cloudflare.WorkerScriptSettingsResponse{
					WorkerMetaData: cloudflare.WorkerMetaData{
						Logpush: ptr.To(false), // Different from desired
					},
				}, nil)
				return client
			},
			want: want{
				isUpToDate: false,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			isUpToDate, err := client.IsUpToDate(context.Background(), tc.args.params, tc.args.obs)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("IsUpToDate() error = %v, want %v", err, tc.want.err)
				}
				return
			}

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