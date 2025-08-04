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

package route

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
	testZoneID  = "test-zone-id"
	testRouteID = "test-route-id"
)

func TestCreate(t *testing.T) {
	type args struct {
		zoneID string
		params v1alpha1.RouteParameters
	}
	type want struct {
		obs *v1alpha1.RouteObservation
		err error
	}

	script := "test-script"

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"CreateSuccessWithScript": {
			args: args{
				zoneID: testZoneID,
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
					Script:  &script,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("CreateWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.CreateWorkerRouteParams{
						Pattern: "example.com/*",
						Script:  "test-script",
					},
				).Return(cloudflare.WorkerRouteResponse{
					WorkerRoute: cloudflare.WorkerRoute{
						ID:         testRouteID,
						Pattern:    "example.com/*",
						ScriptName: "test-script",
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.RouteObservation{},
			},
		},
		"CreateSuccessWithoutScript": {
			args: args{
				zoneID: testZoneID,
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("CreateWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.CreateWorkerRouteParams{
						Pattern: "example.com/*",
						Script:  "",
					},
				).Return(cloudflare.WorkerRouteResponse{
					WorkerRoute: cloudflare.WorkerRoute{
						ID:      testRouteID,
						Pattern: "example.com/*",
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.RouteObservation{},
			},
		},
		"CreateError": {
			args: args{
				zoneID: testZoneID,
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
					Script:  &script,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("CreateWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.CreateWorkerRouteParams{
						Pattern: "example.com/*",
						Script:  "test-script",
					},
				).Return(nil, errors.New("api error"))
				return client
			},
			want: want{
				err: errors.New("cannot create worker route: api error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Create(context.Background(), tc.args.zoneID, tc.args.params)

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
		zoneID  string
		routeID string
	}
	type want struct {
		obs *v1alpha1.RouteObservation
		err error
	}

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"GetSuccess": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
						{
							ID:         "other-route-id",
							Pattern:    "other.com/*",
							ScriptName: "other-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.RouteObservation{},
			},
		},
		"GetNotFound": {
			args: args{
				zoneID:  testZoneID,
				routeID: "nonexistent-route-id",
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				err: clients.NewNotFoundError("worker route not found"),
			},
		},
		"GetListError": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(nil, errors.New("list error"))
				return client
			},
			want: want{
				err: errors.New("cannot get worker route: list error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Get(context.Background(), tc.args.zoneID, tc.args.routeID)

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
		zoneID  string
		routeID string
		params  v1alpha1.RouteParameters
	}
	type want struct {
		obs *v1alpha1.RouteObservation
		err error
	}

	script := "updated-script"

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"UpdateSuccess": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
				params: v1alpha1.RouteParameters{
					Pattern: "updated.example.com/*",
					Script:  &script,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("UpdateWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.UpdateWorkerRouteParams{
						ID:      testRouteID,
						Pattern: "updated.example.com/*",
						Script:  "updated-script",
					},
				).Return(cloudflare.WorkerRouteResponse{}, nil)
				// Mock the Get call that happens after update
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "updated.example.com/*",
							ScriptName: "updated-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				obs: &v1alpha1.RouteObservation{},
			},
		},
		"UpdateError": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
				params: v1alpha1.RouteParameters{
					Pattern: "updated.example.com/*",
					Script:  &script,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("UpdateWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.UpdateWorkerRouteParams{
						ID:      testRouteID,
						Pattern: "updated.example.com/*",
						Script:  "updated-script",
					},
				).Return(nil, errors.New("update error"))
				return client
			},
			want: want{
				err: errors.New("cannot update worker route: update error"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			obs, err := client.Update(context.Background(), tc.args.zoneID, tc.args.routeID, tc.args.params)

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
		zoneID  string
		routeID string
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
				zoneID:  testZoneID,
				routeID: testRouteID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("DeleteWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					testRouteID,
				).Return(cloudflare.WorkerRouteResponse{}, nil)
				return client
			},
			want: want{
				err: nil,
			},
		},
		"DeleteError": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("DeleteWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					testRouteID,
				).Return(nil, errors.New("delete error"))
				return client
			},
			want: want{
				err: errors.New("cannot delete worker route: delete error"),
			},
		},
		"DeleteNotFoundIgnored": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("DeleteWorkerRoute", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					testRouteID,
				).Return(nil, errors.New("Worker Route not found"))
				return client
			},
			want: want{
				err: nil, // Not found errors should be ignored
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			err := client.Delete(context.Background(), tc.args.zoneID, tc.args.routeID)

			if tc.want.err != nil {
				if err == nil || err.Error() != tc.want.err.Error() {
					t.Errorf("Delete() error = %v, want %v", err, tc.want.err)
				}
				return
			}

			if err != nil {
				t.Errorf("Delete() unexpected error = %v", err)
				return
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		zoneID  string
		routeID string
		params  v1alpha1.RouteParameters
	}
	type want struct {
		upToDate bool
		err      error
	}

	script1 := "test-script"
	script2 := "different-script"

	cases := map[string]struct {
		args       args
		mockClient func() clients.ClientInterface
		want       want
	}{
		"UpToDateMatching": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
					Script:  &script1,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				upToDate: true,
			},
		},
		"UpToDateDifferentPattern": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
				params: v1alpha1.RouteParameters{
					Pattern: "different.com/*",
					Script:  &script1,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateDifferentScript": {
			args: args{
				zoneID:  testZoneID,
				routeID: testRouteID,
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
					Script:  &script2,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				upToDate: false,
			},
		},
		"UpToDateRouteNotFound": {
			args: args{
				zoneID:  testZoneID,
				routeID: "nonexistent-route-id",
				params: v1alpha1.RouteParameters{
					Pattern: "example.com/*",
					Script:  &script1,
				},
			},
			mockClient: func() clients.ClientInterface {
				client := clients.NewMockClient()
				client.On("ListWorkerRoutes", 
					context.Background(), 
					cloudflare.ZoneIdentifier(testZoneID),
					cloudflare.ListWorkerRoutesParams{},
				).Return(cloudflare.WorkerRoutesResponse{
					Routes: []cloudflare.WorkerRoute{
						{
							ID:         testRouteID,
							Pattern:    "example.com/*",
							ScriptName: "test-script",
						},
					},
				}, nil)
				return client
			},
			want: want{
				upToDate: false,
				err:      clients.NewNotFoundError("worker route not found"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			client := NewClient(tc.mockClient())
			upToDate, err := client.IsUpToDate(context.Background(), tc.args.zoneID, tc.args.routeID, tc.args.params)

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

			if upToDate != tc.want.upToDate {
				t.Errorf("IsUpToDate() = %v, want %v", upToDate, tc.want.upToDate)
			}
		})
	}
}

func TestIsRouteNotFound(t *testing.T) {
	cases := map[string]struct {
		reason string
		err    error
		want   bool
	}{
		"NilError": {
			reason: "Should return false for nil error",
			err:    nil,
			want:   false,
		},
		"RouteNotFoundError": {
			reason: "Should return true for route not found error",
			err:    errors.New(errRouteNotFound),
			want:   true,
		},
		"404Error": {
			reason: "Should return true for 404 error",
			err:    errors.New("404"),
			want:   true,
		},
		"NotFoundError": {
			reason: "Should return true for Not found error",
			err:    errors.New("Not found"),
			want:   true,
		},
		"OtherError": {
			reason: "Should return false for other errors",
			err:    errors.New("other error"),
			want:   false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRouteNotFound(tc.err)

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("\n%s\nIsRouteNotFound(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}