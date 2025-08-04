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

package kvnamespace

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/rossigee/provider-cloudflare/apis/workers/v1alpha1"
	"github.com/rossigee/provider-cloudflare/internal/clients"
)

const (
	testAccountID     = "test-account-id"
	testNamespaceID   = "test-namespace-id"
	testNamespaceTitle = "Test KV Namespace"
)

func TestCreate(t *testing.T) {
	type args struct {
		params v1alpha1.KVNamespaceParameters
	}
	type want struct {
		obs *v1alpha1.KVNamespaceObservation
		err error
	}

	cases := map[string]struct {
		args        args
		mockClient  func() clients.ClientInterface
		want        want
	}{
		"CreateSuccess": {
			args: args{
				params: v1alpha1.KVNamespaceParameters{
					Title: testNamespaceTitle,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("CreateWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.CreateWorkersKVNamespaceParams{
						Title: testNamespaceTitle,
					},
				).Return(cloudflare.WorkersKVNamespaceResponse{
					Result: cloudflare.WorkersKVNamespace{
						ID:    testNamespaceID,
						Title: testNamespaceTitle,
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.KVNamespaceObservation{
					ID:    testNamespaceID,
					Title: testNamespaceTitle,
				},
			},
		},
		"CreateError": {
			args: args{
				params: v1alpha1.KVNamespaceParameters{
					Title: testNamespaceTitle,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("CreateWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.CreateWorkersKVNamespaceParams{
						Title: testNamespaceTitle,
					},
				).Return(cloudflare.WorkersKVNamespaceResponse{}, errors.New("api error"))
				return client
			},
			want: want{
				err: errors.New("cannot create workers kv namespace: api error"),
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
		namespaceID string
	}
	type want struct {
		obs *v1alpha1.KVNamespaceObservation
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"GetSuccess": {
			args: args{
				namespaceID: testNamespaceID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkersKVNamespaces", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkersKVNamespacesParams{},
				).Return([]cloudflare.WorkersKVNamespace{
					{
						ID:    testNamespaceID,
						Title: testNamespaceTitle,
					},
					{
						ID:    "other-namespace-id",
						Title: "Other Namespace",
					},
				}, &cloudflare.ResultInfo{}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.KVNamespaceObservation{
					ID:    testNamespaceID,
					Title: testNamespaceTitle,
				},
			},
		},
		"GetNotFound": {
			args: args{
				namespaceID: testNamespaceID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkersKVNamespaces", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkersKVNamespacesParams{},
				).Return([]cloudflare.WorkersKVNamespace{
					{
						ID:    "other-namespace-id",
						Title: "Other Namespace",
					},
				}, &cloudflare.ResultInfo{}, nil)
				return client
			},
			want: want{
				err: clients.NewNotFoundError("kv namespace not found"),
			},
		},
		"GetListError": {
			args: args{
				namespaceID: testNamespaceID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("ListWorkersKVNamespaces", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkersKVNamespacesParams{},
				).Return([]cloudflare.WorkersKVNamespace{}, &cloudflare.ResultInfo{}, errors.New("list error"))
				return client
			},
			want: want{
				err: errors.New("cannot get workers kv namespace: list error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Get(context.Background(), tc.args.namespaceID)

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

func TestUpdate(t *testing.T) {
	type args struct {
		namespaceID string
		params      v1alpha1.KVNamespaceParameters
	}
	type want struct {
		obs *v1alpha1.KVNamespaceObservation
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"UpdateSuccess": {
			args: args{
				namespaceID: testNamespaceID,
				params: v1alpha1.KVNamespaceParameters{
					Title: "Updated Title",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("UpdateWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkersKVNamespaceParams{
						NamespaceID: testNamespaceID,
						Title:       "Updated Title",
					},
				).Return(cloudflare.Response{}, nil)
				// Mock the Get call after update
				client.On("ListWorkersKVNamespaces", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.ListWorkersKVNamespacesParams{},
				).Return([]cloudflare.WorkersKVNamespace{
					{
						ID:    testNamespaceID,
						Title: "Updated Title",
					},
				}, &cloudflare.ResultInfo{}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.KVNamespaceObservation{
					ID:    testNamespaceID,
					Title: "Updated Title",
				},
			},
		},
		"UpdateError": {
			args: args{
				namespaceID: testNamespaceID,
				params: v1alpha1.KVNamespaceParameters{
					Title: "Updated Title",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("UpdateWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					cloudflare.UpdateWorkersKVNamespaceParams{
						NamespaceID: testNamespaceID,
						Title:       "Updated Title",
					},
				).Return(cloudflare.Response{}, errors.New("update error"))
				return client
			},
			want: want{
				err: errors.New("cannot update workers kv namespace: update error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Update(context.Background(), tc.args.namespaceID, tc.args.params)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("Update() error = %v, want %v", err, tc.want.err)
				}
				return
			}

			if err != nil {
				t.Errorf("Update() unexpected error = %v", err)
				return
			}

			if diff := cmp.Diff(tc.want.obs, obs); diff != "" {
				t.Errorf("Update() -want +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		namespaceID string
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
				namespaceID: testNamespaceID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("DeleteWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testNamespaceID,
				).Return(cloudflare.Response{}, nil)
				return client
			},
			want: want{},
		},
		"DeleteError": {
			args: args{
				namespaceID: testNamespaceID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("GetAccountID").Return(testAccountID)
				client.On("DeleteWorkersKVNamespace", 
					context.Background(), 
					cloudflare.AccountIdentifier(testAccountID),
					testNamespaceID,
				).Return(cloudflare.Response{}, errors.New("delete error"))
				return client
			},
			want: want{
				err: errors.New("cannot delete workers kv namespace: delete error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			err := client.Delete(context.Background(), tc.args.namespaceID)

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
		params v1alpha1.KVNamespaceParameters
		obs    v1alpha1.KVNamespaceObservation
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
				params: v1alpha1.KVNamespaceParameters{
					Title: testNamespaceTitle,
				},
				obs: v1alpha1.KVNamespaceObservation{
					ID:    testNamespaceID,
					Title: testNamespaceTitle,
				},
			},
			want: want{
				isUpToDate: true,
			},
		},
		"NotUpToDate": {
			args: args{
				params: v1alpha1.KVNamespaceParameters{
					Title: "New Title",
				},
				obs: v1alpha1.KVNamespaceObservation{
					ID:    testNamespaceID,
					Title: testNamespaceTitle,
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