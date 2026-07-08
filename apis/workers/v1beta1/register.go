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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Package type metadata.
const (
	ScriptKind      = "Script"
	RouteKind       = "Route"
	CronTriggerKind = "CronTrigger"
	DomainKind      = "Domain"
	KVNamespaceKind = "KVNamespace"
	SubdomainKind   = "Subdomain"
)

var (
	ScriptKindAPIVersion   = ScriptKind + "." + GroupVersion.String()
	ScriptGroupKind        = schema.GroupKind{Group: Group, Kind: ScriptKind}.String()
	ScriptGroupVersionKind = GroupVersion.WithKind(ScriptKind)

	RouteKindAPIVersion   = RouteKind + "." + GroupVersion.String()
	RouteGroupKind        = schema.GroupKind{Group: Group, Kind: RouteKind}.String()
	RouteGroupVersionKind = GroupVersion.WithKind(RouteKind)

	CronTriggerKindAPIVersion   = CronTriggerKind + "." + GroupVersion.String()
	CronTriggerGroupKind        = schema.GroupKind{Group: Group, Kind: CronTriggerKind}.String()
	CronTriggerGroupVersionKind = GroupVersion.WithKind(CronTriggerKind)

	DomainKindAPIVersion   = DomainKind + "." + GroupVersion.String()
	DomainGroupKind        = schema.GroupKind{Group: Group, Kind: DomainKind}.String()
	DomainGroupVersionKind = GroupVersion.WithKind(DomainKind)

	KVNamespaceKindAPIVersion   = KVNamespaceKind + "." + GroupVersion.String()
	KVNamespaceGroupKind        = schema.GroupKind{Group: Group, Kind: KVNamespaceKind}.String()
	KVNamespaceGroupVersionKind = GroupVersion.WithKind(KVNamespaceKind)

	SubdomainKindAPIVersion   = SubdomainKind + "." + GroupVersion.String()
	SubdomainGroupKind        = schema.GroupKind{Group: Group, Kind: SubdomainKind}.String()
	SubdomainGroupVersionKind = GroupVersion.WithKind(SubdomainKind)
)
