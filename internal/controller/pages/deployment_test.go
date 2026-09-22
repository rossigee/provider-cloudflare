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

package pages

import (
	"context"
	"testing"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/pages/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/pages"
)

type mockDeploymentClient struct {
	MockGetPagesDeploymentInfo func(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error)
	MockCreatePagesDeployment  func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesDeploymentParams) (cloudflare.PagesProjectDeployment, error)
}

func (m *mockDeploymentClient) GetPagesDeploymentInfo(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	if m.MockGetPagesDeploymentInfo != nil {
		return m.MockGetPagesDeploymentInfo(ctx, rc, projectName, deploymentID)
	}
	return cloudflare.PagesProjectDeployment{ID: deploymentID}, nil
}

func (m *mockDeploymentClient) CreatePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesDeploymentParams) (cloudflare.PagesProjectDeployment, error) {
	if m.MockCreatePagesDeployment != nil {
		return m.MockCreatePagesDeployment(ctx, rc, params)
	}
	return cloudflare.PagesProjectDeployment{}, nil
}

func (m *mockDeploymentClient) DeletePagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.DeletePagesDeploymentParams) error {
	return nil
}

func (m *mockDeploymentClient) ListPagesDeployments(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesDeploymentsParams) ([]cloudflare.PagesProjectDeployment, *cloudflare.ResultInfo, error) {
	return nil, nil, nil
}

func (m *mockDeploymentClient) RetryPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	return cloudflare.PagesProjectDeployment{}, nil
}

func (m *mockDeploymentClient) RollbackPagesDeployment(ctx context.Context, rc *cloudflare.ResourceContainer, projectName, deploymentID string) (cloudflare.PagesProjectDeployment, error) {
	return cloudflare.PagesProjectDeployment{}, nil
}

func (m *mockDeploymentClient) GetPagesDeploymentLogs(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.GetPagesDeploymentLogsParams) (cloudflare.PagesDeploymentLogs, error) {
	return cloudflare.PagesDeploymentLogs{}, nil
}

func deployment(mods ...func(*v1beta1.Deployment)) *v1beta1.Deployment {
	d := &v1beta1.Deployment{
		Spec: v1beta1.DeploymentSpec{
			ForProvider: v1beta1.DeploymentParameters{
				AccountID:   "acct-123",
				ProjectName: "proj",
			},
		},
	}
	for _, m := range mods {
		m(d)
	}
	return d
}

func TestDeploymentObserve(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesDeploymentClient
		mg      resource.Managed
		want    managed.ExternalObservation
	}{
		"NoExternalName": {mg: deployment(), want: managed.ExternalObservation{ResourceExists: false}},
		"Success": {
			service: &mockDeploymentClient{},
			mg:      deployment(func(d *v1beta1.Deployment) { meta.SetExternalName(d, "dep-1") }),
			want:    managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ConnectionDetails: managed.ConnectionDetails{}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesDeploymentExternal{service: tc.service}
			got, _ := e.Observe(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s: %s", name, diff)
			}
		})
	}
}

func TestDeploymentCreate(t *testing.T) {
	e := &pagesDeploymentExternal{service: &mockDeploymentClient{}}
	_, err := e.Create(context.Background(), deployment())
	if err != nil {
		t.Error(err)
	}
}

func TestDeploymentUpdateDelete(t *testing.T) {
	e := &pagesDeploymentExternal{service: &mockDeploymentClient{}}
	_, _ = e.Update(context.Background(), deployment(func(d *v1beta1.Deployment) { meta.SetExternalName(d, "d1") }))
	_, _ = e.Delete(context.Background(), deployment(func(d *v1beta1.Deployment) { meta.SetExternalName(d, "d1") }))
}
