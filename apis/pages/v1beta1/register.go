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

var (
	ProjectKind             = reflect.TypeOf(Project{}).Name()
	ProjectGroupKind        = schema.GroupKind{Group: Group, Kind: ProjectKind}
	ProjectKindAPIVersion   = ProjectKind + "." + GroupVersion.String()
	ProjectGroupVersionKind = GroupVersion.WithKind(ProjectKind)

	DeploymentKind             = reflect.TypeOf(Deployment{}).Name()
	DeploymentGroupKind        = schema.GroupKind{Group: Group, Kind: DeploymentKind}
	DeploymentKindAPIVersion   = DeploymentKind + "." + GroupVersion.String()
	DeploymentGroupVersionKind = GroupVersion.WithKind(DeploymentKind)

	DomainKind             = reflect.TypeOf(Domain{}).Name()
	DomainGroupKind        = schema.GroupKind{Group: Group, Kind: DomainKind}
	DomainKindAPIVersion   = DomainKind + "." + GroupVersion.String()
	DomainGroupVersionKind = GroupVersion.WithKind(DomainKind)
)
