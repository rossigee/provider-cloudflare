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
	Group   = "workers.cloudflare.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Route type metadata.
var (
	RouteKind             = reflect.TypeOf(Route{}).Name()
	RouteGroupKind        = schema.GroupKind{Group: Group, Kind: RouteKind}.String()
	RouteKindAPIVersion   = RouteKind + "." + SchemeGroupVersion.String()
	RouteGroupVersionKind = SchemeGroupVersion.WithKind(RouteKind)
)

// Script type metadata.
var (
	ScriptKind             = reflect.TypeOf(Script{}).Name()
	ScriptGroupKind        = schema.GroupKind{Group: Group, Kind: ScriptKind}.String()
	ScriptKindAPIVersion   = ScriptKind + "." + SchemeGroupVersion.String()
	ScriptGroupVersionKind = SchemeGroupVersion.WithKind(ScriptKind)
)

// KVNamespace type metadata.
var (
	KVNamespaceKind             = reflect.TypeOf(KVNamespace{}).Name()
	KVNamespaceGroupKind        = schema.GroupKind{Group: Group, Kind: KVNamespaceKind}.String()
	KVNamespaceKindAPIVersion   = KVNamespaceKind + "." + SchemeGroupVersion.String()
	KVNamespaceGroupVersionKind = SchemeGroupVersion.WithKind(KVNamespaceKind)
)

// CronTrigger type metadata.
var (
	CronTriggerKind             = reflect.TypeOf(CronTrigger{}).Name()
	CronTriggerGroupKind        = schema.GroupKind{Group: Group, Kind: CronTriggerKind}.String()
	CronTriggerKindAPIVersion   = CronTriggerKind + "." + SchemeGroupVersion.String()
	CronTriggerGroupVersionKind = SchemeGroupVersion.WithKind(CronTriggerKind)
)

// Domain type metadata.
var (
	DomainKind             = reflect.TypeOf(Domain{}).Name()
	DomainGroupKind        = schema.GroupKind{Group: Group, Kind: DomainKind}.String()
	DomainKindAPIVersion   = DomainKind + "." + SchemeGroupVersion.String()
	DomainGroupVersionKind = SchemeGroupVersion.WithKind(DomainKind)
)

// Subdomain type metadata.
var (
	SubdomainKind             = reflect.TypeOf(Subdomain{}).Name()
	SubdomainGroupKind        = schema.GroupKind{Group: Group, Kind: SubdomainKind}.String()
	SubdomainKindAPIVersion   = SubdomainKind + "." + SchemeGroupVersion.String()
	SubdomainGroupVersionKind = SchemeGroupVersion.WithKind(SubdomainKind)
)

func init() {
	SchemeBuilder.Register(&Route{}, &RouteList{})
	SchemeBuilder.Register(&Script{}, &ScriptList{})
	SchemeBuilder.Register(&KVNamespace{}, &KVNamespaceList{})
	SchemeBuilder.Register(&CronTrigger{}, &CronTriggerList{})
	SchemeBuilder.Register(&Domain{}, &DomainList{})
	SchemeBuilder.Register(&Subdomain{}, &SubdomainList{})
}
