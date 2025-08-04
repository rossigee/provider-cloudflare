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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	CRDGroup   = "ssl.cloudflare.crossplane.io"
	CRDVersion = "v1alpha1"
)

var (
	// CRDGroupVersion is the API Group Version used to register the objects
	CRDGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: CRDGroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// UniversalSSL type metadata.
var (
	UniversalSSLKind             = reflect.TypeOf(UniversalSSL{}).Name()
	UniversalSSLGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UniversalSSLKind}
	UniversalSSLKindAPIVersion   = UniversalSSLKind + "." + CRDGroupVersion.String()
	UniversalSSLGroupVersionKind = CRDGroupVersion.WithKind(UniversalSSLKind)
)

// TotalTLS type metadata.
var (
	TotalTLSKind             = reflect.TypeOf(TotalTLS{}).Name()
	TotalTLSGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TotalTLSKind}
	TotalTLSKindAPIVersion   = TotalTLSKind + "." + CRDGroupVersion.String()
	TotalTLSGroupVersionKind = CRDGroupVersion.WithKind(TotalTLSKind)
)

// CertificatePack type metadata.
var (
	CertificatePackKind             = reflect.TypeOf(CertificatePack{}).Name()
	CertificatePackGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: CertificatePackKind}
	CertificatePackKindAPIVersion   = CertificatePackKind + "." + CRDGroupVersion.String()
	CertificatePackGroupVersionKind = CRDGroupVersion.WithKind(CertificatePackKind)
)

func init() {
	SchemeBuilder.Register(&UniversalSSL{}, &UniversalSSLList{}, &TotalTLS{}, &TotalTLSList{}, &CertificatePack{}, &CertificatePackList{})
}