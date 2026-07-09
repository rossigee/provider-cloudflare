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

// Route type metadata.
var (
	RouteKind             = reflect.TypeOf(Route{}).Name()
	RouteGroupKind        = schema.GroupKind{Group: Group, Kind: RouteKind}
	RouteKindAPIVersion   = RouteKind + "." + GroupVersion.String()
	RouteGroupVersionKind = GroupVersion.WithKind(RouteKind)

	CronTriggerKind             = reflect.TypeOf(CronTrigger{}).Name()
	CronTriggerGroupKind        = schema.GroupKind{Group: Group, Kind: CronTriggerKind}
	CronTriggerKindAPIVersion   = CronTriggerKind + "." + GroupVersion.String()
	CronTriggerGroupVersionKind = GroupVersion.WithKind(CronTriggerKind)

	DomainKind             = reflect.TypeOf(Domain{}).Name()
	DomainGroupKind        = schema.GroupKind{Group: Group, Kind: DomainKind}
	DomainKindAPIVersion   = DomainKind + "." + GroupVersion.String()
	DomainGroupVersionKind = GroupVersion.WithKind(DomainKind)

	KVNamespaceKind             = reflect.TypeOf(KVNamespace{}).Name()
	KVNamespaceGroupKind        = schema.GroupKind{Group: Group, Kind: KVNamespaceKind}
	KVNamespaceKindAPIVersion   = KVNamespaceKind + "." + GroupVersion.String()
	KVNamespaceGroupVersionKind = GroupVersion.WithKind(KVNamespaceKind)

	ScriptKind             = reflect.TypeOf(Script{}).Name()
	ScriptGroupKind        = schema.GroupKind{Group: Group, Kind: ScriptKind}
	ScriptKindAPIVersion   = ScriptKind + "." + GroupVersion.String()
	ScriptGroupVersionKind = GroupVersion.WithKind(ScriptKind)

	SubdomainKind             = reflect.TypeOf(Subdomain{}).Name()
	SubdomainGroupKind        = schema.GroupKind{Group: Group, Kind: SubdomainKind}
	SubdomainKindAPIVersion   = SubdomainKind + "." + GroupVersion.String()
	SubdomainGroupVersionKind = GroupVersion.WithKind(SubdomainKind)
)
