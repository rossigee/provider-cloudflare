package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Gateway type metadata.
var (
	GatewayKind             = reflect.TypeOf(Gateway{}).Name()
	GatewayGroupKind        = schema.GroupKind{Group: Group, Kind: GatewayKind}
	GatewayKindAPIVersion   = GatewayKind + "." + GroupVersion.String()
	GatewayGroupVersionKind = GroupVersion.WithKind(GatewayKind)
)
