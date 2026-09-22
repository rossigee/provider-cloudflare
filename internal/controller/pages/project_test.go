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
	"time"

	"github.com/cloudflare/cloudflare-go"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"github.com/crossplane/crossplane-runtime/v2/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/rossigee/provider-cloudflare/apis/pages/v1beta1"
	"github.com/rossigee/provider-cloudflare/internal/clients/pages"
)

type mockPagesProjectClient struct {
	MockCreatePagesProject func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesProjectParams) (cloudflare.PagesProject, error)
	MockGetPagesProject    func(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) (cloudflare.PagesProject, error)
	MockUpdatePagesProject func(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdatePagesProjectParams) (cloudflare.PagesProject, error)
	MockDeletePagesProject func(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) error
}

func (m *mockPagesProjectClient) CreatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.CreatePagesProjectParams) (cloudflare.PagesProject, error) {
	if m.MockCreatePagesProject != nil {
		return m.MockCreatePagesProject(ctx, rc, params)
	}
	return cloudflare.PagesProject{ID: "proj-123", Name: params.Name}, nil
}

func (m *mockPagesProjectClient) GetPagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) (cloudflare.PagesProject, error) {
	if m.MockGetPagesProject != nil {
		return m.MockGetPagesProject(ctx, rc, projectName)
	}
	now := time.Now()
	p := cloudflare.PagesProject{ID: projectName, Name: projectName}
	p.BuildConfig.BuildCommand = "npm run build"
	p.BuildConfig.DestinationDir = "dist"
	p.BuildConfig.RootDir = "/"
	p.CreatedOn = &now
	return p, nil
}

func (m *mockPagesProjectClient) UpdatePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.UpdatePagesProjectParams) (cloudflare.PagesProject, error) {
	if m.MockUpdatePagesProject != nil {
		return m.MockUpdatePagesProject(ctx, rc, params)
	}
	return cloudflare.PagesProject{}, nil
}

func (m *mockPagesProjectClient) DeletePagesProject(ctx context.Context, rc *cloudflare.ResourceContainer, projectName string) error {
	if m.MockDeletePagesProject != nil {
		return m.MockDeletePagesProject(ctx, rc, projectName)
	}
	return nil
}

func (m *mockPagesProjectClient) ListPagesProjects(ctx context.Context, rc *cloudflare.ResourceContainer, params cloudflare.ListPagesProjectsParams) ([]cloudflare.PagesProject, cloudflare.ResultInfo, error) {
	return nil, cloudflare.ResultInfo{}, nil
}

func pagesProject(mods ...func(*v1beta1.Project)) *v1beta1.Project {
	p := &v1beta1.Project{
		Spec: v1beta1.ProjectSpec{
			ForProvider: v1beta1.ProjectParameters{
				AccountID: "acct-123",
				Name:      "test-project",
			},
		},
	}
	for _, m := range mods {
		m(p)
	}
	return p
}

func TestPagesProjectObserve(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesProjectClient
		mg      resource.Managed
		want    managed.ExternalObservation
		wantErr error
	}{
		"NoExternalName": {
			mg:   pagesProject(),
			want: managed.ExternalObservation{ResourceExists: false},
		},
		"Success": {
			service: &mockPagesProjectClient{},
			mg:      pagesProject(func(p *v1beta1.Project) { meta.SetExternalName(p, "proj-123") }),
			want: managed.ExternalObservation{
				ResourceExists:    true,
				ResourceUpToDate:  true,
				ConnectionDetails: managed.ConnectionDetails{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesProjectExternal{service: tc.service}
			got, err := e.Observe(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s: -wantErr +gotErr\n%s", name, diff)
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPagesProjectCreate(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesProjectClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPagesProjectClient{},
			mg:      pagesProject(),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesProjectExternal{service: tc.service}
			_, err := e.Create(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPagesProjectUpdate(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesProjectClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPagesProjectClient{},
			mg:      pagesProject(func(p *v1beta1.Project) { meta.SetExternalName(p, "proj-123") }),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesProjectExternal{service: tc.service}
			_, err := e.Update(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s: -want +got\n%s", name, diff)
			}
		})
	}
}

func TestPagesProjectDelete(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesProjectClient
		mg      resource.Managed
		wantErr error
	}{
		"Success": {
			service: &mockPagesProjectClient{},
			mg:      pagesProject(func(p *v1beta1.Project) { meta.SetExternalName(p, "proj-123") }),
			wantErr: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesProjectExternal{service: tc.service}
			_, err := e.Delete(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("%s: -want +got\n%s", name, diff)
			}
		})
	}
}
