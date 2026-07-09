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

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// CertificatePack type metadata.
var (
	CertificatePackKind             = reflect.TypeOf(CertificatePack{}).Name()
	CertificatePackGroupKind        = schema.GroupKind{Group: Group, Kind: CertificatePackKind}
	CertificatePackKindAPIVersion   = CertificatePackKind + "." + SchemeGroupVersion.String()
	CertificatePackGroupVersionKind = SchemeGroupVersion.WithKind(CertificatePackKind)

	TotalTLSKind             = reflect.TypeOf(TotalTLS{}).Name()
	TotalTLSGroupKind        = schema.GroupKind{Group: Group, Kind: TotalTLSKind}
	TotalTLSKindAPIVersion   = TotalTLSKind + "." + SchemeGroupVersion.String()
	TotalTLSGroupVersionKind = SchemeGroupVersion.WithKind(TotalTLSKind)

	UniversalSSLKind             = reflect.TypeOf(UniversalSSL{}).Name()
	UniversalSSLGroupKind        = schema.GroupKind{Group: Group, Kind: UniversalSSLKind}
	UniversalSSLKindAPIVersion   = UniversalSSLKind + "." + SchemeGroupVersion.String()
	UniversalSSLGroupVersionKind = SchemeGroupVersion.WithKind(UniversalSSLKind)
)
