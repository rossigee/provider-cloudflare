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

package ssl

import (
	"testing"

	"github.com/rossigee/provider-cloudflare/apis/ssl/v1beta1"
)

// TestUniversalSSL confirms UniversalSSL type exists and is accessible.
func TestUniversalSSL(t *testing.T) {
	r := &v1beta1.UniversalSSL{
		Spec: v1beta1.UniversalSSLSpec{
			ForProvider: v1beta1.UniversalSSLParameters{},
		},
	}

	t.Logf("resource type %T instantiated successfully", r)
}
