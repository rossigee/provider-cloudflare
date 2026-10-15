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

type mockDomainClient struct {
	MockGetPagesDomain func(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error)
	MockPagesAddDomain func(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error)
}

func (m *mockDomainClient) GetPagesDomains(ctx context.Context, params cloudflare.PagesDomainsParameters) ([]cloudflare.PagesDomain, error) {
	return nil, nil
}

func (m *mockDomainClient) GetPagesDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	if m.MockGetPagesDomain != nil {
		return m.MockGetPagesDomain(ctx, params)
	}
	return cloudflare.PagesDomain{}, nil
}

func (m *mockDomainClient) PagesPatchDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	return cloudflare.PagesDomain{}, nil
}

func (m *mockDomainClient) PagesAddDomain(ctx context.Context, params cloudflare.PagesDomainParameters) (cloudflare.PagesDomain, error) {
	if m.MockPagesAddDomain != nil {
		return m.MockPagesAddDomain(ctx, params)
	}
	return cloudflare.PagesDomain{}, nil
}

func (m *mockDomainClient) PagesDeleteDomain(ctx context.Context, params cloudflare.PagesDomainParameters) error {
	return nil
}

func domain(mods ...func(*v1beta1.Domain)) *v1beta1.Domain {
	d := &v1beta1.Domain{
		Spec: v1beta1.DomainSpec{
			ForProvider: v1beta1.DomainParameters{
				AccountID:   "acct-123",
				ProjectName: "proj",
				Domain:      "example.com",
			},
		},
	}
	for _, m := range mods {
		m(d)
	}
	return d
}

func TestDomainObserve(t *testing.T) {
	cases := map[string]struct {
		service pages.CloudflarePagesDomainClient
		mg      resource.Managed
		want    managed.ExternalObservation
	}{
		"NoExternalName": {mg: domain(), want: managed.ExternalObservation{ResourceExists: false}},
		"Success": {
			service: &mockDomainClient{},
			mg:      domain(func(d *v1beta1.Domain) { meta.SetExternalName(d, "dom-1") }),
			want:    managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true, ConnectionDetails: managed.ConnectionDetails{}},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &pagesDomainExternal{service: tc.service}
			got, _ := e.Observe(context.Background(), tc.mg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s: %s", name, diff)
			}
		})
	}
}

func TestDomainCreate(t *testing.T) {
	e := &pagesDomainExternal{service: &mockDomainClient{}}
	_, err := e.Create(context.Background(), domain())
	if err != nil {
		t.Error(err)
	}
}

func TestDomainUpdateDelete(t *testing.T) {
	e := &pagesDomainExternal{service: &mockDomainClient{}}
	_, _ = e.Update(context.Background(), domain(func(d *v1beta1.Domain) { meta.SetExternalName(d, "d1") }))
	_, _ = e.Delete(context.Background(), domain(func(d *v1beta1.Domain) { meta.SetExternalName(d, "d1") }))
}
