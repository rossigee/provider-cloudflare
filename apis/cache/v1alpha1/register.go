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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	CRDGroup   = "cache.cloudflare.crossplane.io"
	CRDVersion = "v1alpha1"
)

var (
	// CRDGroupVersion is the API Group Version used to register the objects
	CRDGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}
)

// CacheRule type metadata.
var (
	CacheRuleKind             = reflect.TypeOf(CacheRule{}).Name()
	CacheRuleGroupKind        = schema.GroupKind{Group: CRDGroup, Kind: CacheRuleKind}.String()
	CacheRuleKindAPIVersion   = CacheRuleKind + "." + CRDGroupVersion.String()
	CacheRuleGroupVersionKind = CRDGroupVersion.WithKind(CacheRuleKind)
)

var (
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme is a global function that registers this API group & version to a scheme
	AddToScheme = SchemeBuilder.AddToScheme
)

// addKnownTypes adds the set of types defined in this package to the supplied scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&CacheRule{},
		&CacheRuleList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
}