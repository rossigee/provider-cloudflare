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

package originssl

import (
	"testing"

	"github.com/rossigee/provider-cloudflare/apis/originssl/v1beta1"
)

// TestCertificate confirms Certificate type exists and is accessible.
func TestCertificate(t *testing.T) {
	r := &v1beta1.Certificate{
		Spec: v1beta1.CertificateSpec{
			ForProvider: v1beta1.CertificateParameters{},
		},
	}

	t.Logf("resource type %T instantiated successfully", r)
}
