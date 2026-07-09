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

// SessionAffinityAttributes type metadata.
var (
	SessionAffinityAttributesKind             = reflect.TypeOf(SessionAffinityAttributes{}).Name()
	SessionAffinityAttributesGroupKind        = schema.GroupKind{Group: Group, Kind: SessionAffinityAttributesKind}
	SessionAffinityAttributesKindAPIVersion   = SessionAffinityAttributesKind + "." + SchemeGroupVersion.String()
	SessionAffinityAttributesGroupVersionKind = SchemeGroupVersion.WithKind(SessionAffinityAttributesKind)

	LoadBalancerKind             = reflect.TypeOf(LoadBalancer{}).Name()
	LoadBalancerGroupKind        = schema.GroupKind{Group: Group, Kind: LoadBalancerKind}
	LoadBalancerKindAPIVersion   = LoadBalancerKind + "." + SchemeGroupVersion.String()
	LoadBalancerGroupVersionKind = SchemeGroupVersion.WithKind(LoadBalancerKind)

	LoadBalancerMonitorKind             = reflect.TypeOf(LoadBalancerMonitor{}).Name()
	LoadBalancerMonitorGroupKind        = schema.GroupKind{Group: Group, Kind: LoadBalancerMonitorKind}
	LoadBalancerMonitorKindAPIVersion   = LoadBalancerMonitorKind + "." + SchemeGroupVersion.String()
	LoadBalancerMonitorGroupVersionKind = SchemeGroupVersion.WithKind(LoadBalancerMonitorKind)

	LoadBalancerPoolKind             = reflect.TypeOf(LoadBalancerPool{}).Name()
	LoadBalancerPoolGroupKind        = schema.GroupKind{Group: Group, Kind: LoadBalancerPoolKind}
	LoadBalancerPoolKindAPIVersion   = LoadBalancerPoolKind + "." + SchemeGroupVersion.String()
	LoadBalancerPoolGroupVersionKind = SchemeGroupVersion.WithKind(LoadBalancerPoolKind)
)
