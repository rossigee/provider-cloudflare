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

// Package apis contains Kubernetes API for the Template provider.
package apis

import (
	"github.com/rossigee/provider-cloudflare/apis/access/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/cache/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/device/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/dns/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/emailrouting/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/firewall/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/loadbalancing/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/logpush/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/originssl/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/r2/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/rulesets/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/security/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/spectrum/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/ssl/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/sslsaas/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/transform/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/tunnel/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/workers/v1beta1"
	"github.com/rossigee/provider-cloudflare/apis/zone/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		cloudflarev1beta1.SchemeBuilder.AddToScheme,
		cachev1beta1.SchemeBuilder.AddToScheme,
		dnsv1beta1.SchemeBuilder.AddToScheme,
		emailroutingv1beta1.SchemeBuilder.AddToScheme,
		sslsaasv1beta1.SchemeBuilder.AddToScheme,
		originsslv1beta1.SchemeBuilder.AddToScheme,
		spectrumv1beta1.SchemeBuilder.AddToScheme,
		zonev1beta1.SchemeBuilder.AddToScheme,
		firewallv1beta1.SchemeBuilder.AddToScheme,
		workersv1beta1.SchemeBuilder.AddToScheme,
		transformv1beta1.SchemeBuilder.AddToScheme,
		rulesetsv1beta1.SchemeBuilder.AddToScheme,
		securityv1beta1.SchemeBuilder.AddToScheme,
		sslv1beta1.SchemeBuilder.AddToScheme,
		loadbalancingv1beta1.SchemeBuilder.AddToScheme,
		logpushv1beta1.SchemeBuilder.AddToScheme,
		r2v1beta1.SchemeBuilder.AddToScheme,
		accessv1beta1.SchemeBuilder.AddToScheme,
		tunnelv1beta1.SchemeBuilder.AddToScheme,
		devicev1beta1.SchemeBuilder.AddToScheme,
	)
}

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}

// VerifySchemeRegistration verifies that all API types are properly registered
func VerifySchemeRegistration() error {
	// Create a test scheme
	scheme := runtime.NewScheme()

	// Try to add all types to the scheme
	if err := AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}
