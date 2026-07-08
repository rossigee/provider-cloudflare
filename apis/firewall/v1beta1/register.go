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
	"firewall.cloudflare.m.crossplane.io"
	"firewallv1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
)

// Rule type metadata.
var (
	RuleKind             = reflect.TypeOf(Rule{}).Name()
	RuleGroupKind        = schema.GroupKind{Group: Group, Kind: RuleKind}.String()
	RuleKindAPIVersion   = RuleKind + "." + SchemeGroupVersion.String()
	RuleGroupVersionKind = SchemeGroupVersion.WithKind(RuleKind)
)

// Filter type metadata.
var (
	FilterKind             = reflect.TypeOf(Filter{}).Name()
	FilterGroupKind        = schema.GroupKind{Group: Group, Kind: FilterKind}.String()
	FilterKindAPIVersion   = FilterKind + "." + SchemeGroupVersion.String()
	FilterGroupVersionKind = SchemeGroupVersion.WithKind(FilterKind)
)

func addKnownTypes(s *runtime.Scheme) error {
	s.AddKnownTypes(SchemeGroupVersion,
		&Rule{},
		&RuleList{},
		&Filter{},
		&FilterList{},
	)
	return nil
}
