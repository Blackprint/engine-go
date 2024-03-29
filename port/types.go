package port

import (
	"reflect"

	engine "github.com/blackprint/engine-go/engine"
)

/* This port can contain multiple cable as input
 * and the value will be array of 'type'
 * it's only one type, not union
 * for union port, please split it to different port to handle it
 */
func ArrayOf(type_ reflect.Kind) *engine.PortFeature {
	return &engine.PortFeature{
		Id:   engine.PortTypeArrayOf,
		Type: type_,
	}
}

/* This port can have default value if no cable was connected
 * type = Type Data that allowed for the Port
 * value = default value for the port
 */
func Default(type_ reflect.Kind, val any) *engine.PortFeature {
	return &engine.PortFeature{
		Id:    engine.PortTypeDefault,
		Type:  type_,
		Value: val,
	}
}

/* Allow many cable connected to a port
 * But only the last value that will used as value
 */
func Switch(type_ reflect.Kind) *engine.PortFeature {
	return &engine.PortFeature{
		Id:   engine.PortTypeSwitch,
		Type: type_,
	}
}

/* This port will be used as a trigger or callable input port
 * func = callback when the port was being called as a function
 */
func Trigger(callback func(*engine.Port)) *engine.PortFeature {
	return &engine.PortFeature{
		Id:   engine.PortTypeTrigger,
		Func: callback,
	}
}

/* This port can allow multiple different types
 * like an 'any' port, but can only contain one value
 */
func Union(types []reflect.Kind) *engine.PortFeature {
	return &engine.PortFeature{
		Id:    engine.PortTypeUnion,
		Types: types,
	}
}

/* This port will allow any value to be passed to a function
 * then you can write custom data validation in the function
 * the value returned by your function will be used as the input value
 */
func Validator(type_ reflect.Kind, callback func(any) any) *engine.PortFeature {
	return &engine.PortFeature{
		Id:   engine.PortTypeValidator,
		Type: type_,
		Func: callback,
	}
}
