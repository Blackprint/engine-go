package engine

import (
	"reflect"
)

type portObject struct{}

var Ports *portObject

func init() {
	Ports = &portObject{}
}

/* This port can contain multiple cable as input
 * and the value will be array of 'type'
 * it's only one type, not union
 * for union port, please split it to different port to handle it
 */
func (*portObject) ArrayOf(type_ reflect.Kind) *PortFeature {
	return &PortFeature{
		Id:   PortTypeArrayOf,
		Type: type_,
	}
}

/* This port can have default value if no cable was connected
 * type = Type Data that allowed for the Port
 * value = default value for the port
 */
func (*portObject) Default(type_ reflect.Kind, val any) *PortFeature {
	return &PortFeature{
		Id:    PortTypeDefault,
		Type:  type_,
		Value: val,
	}
}

/* This port will be used as a trigger or callable input port
 * func (*portObject) = callback when the port was being called as a function
 */
func (*portObject) Trigger(callback func(*Port)) *PortFeature {
	return &PortFeature{
		Id:   PortTypeTrigger,
		Func: callback,
	}
}

/* This port can allow multiple different types
 * like an 'any' port, but can only contain one value
 */
func (*portObject) Union(types []reflect.Kind) *PortFeature {
	return &PortFeature{
		Id:    PortTypeUnion,
		Types: types,
	}
}

/* This port can allow multiple different types
 * like an 'any' port, but can only contain one value
 */
func (*portObject) StructOf(type_ reflect.Kind, structure map[string]PortStructTemplate) *PortFeature {
	return &PortFeature{
		Id:    PortTypeStructOf,
		Type:  type_,
		Value: structure,
	}
}

func (*portObject) Route() *PortFeature {
	return &PortFeature{
		Id: PortTypeRoute,
	}
}
