package port

import (
	"reflect"
)

const (
	Input = iota + 1
	Output
	Property
)

const (
	TypeArrayOf = iota + 1
	TypeDefault
	TypeSwitch
	TypeTrigger
	TypeUnion
	TypeValidator
)

// Port feature
type Feature struct {
	Id    int
	Type  reflect.Kind
	Types []reflect.Kind
	Value any
	Func  any
}

/* This port can contain multiple cable as input
 * and the value will be array of 'type'
 * it's only one type, not union
 * for union port, please split it to different port to handle it
 */
func ArrayOf(type_ reflect.Kind) *Feature {
	return &Feature{
		Id:   TypeArrayOf,
		Type: type_,
	}
}

/* This port can have default value if no cable was connected
 * type = Type Data that allowed for the Port
 * value = default value for the port
 */
func Default(type_ reflect.Kind, val any) *Feature {
	return &Feature{
		Id:    TypeDefault,
		Type:  type_,
		Value: val,
	}
}

/* Allow many cable connected to a port
 * But only the last value that will used as value
 */
func Switch(type_ reflect.Kind) *Feature {
	return &Feature{
		Id:   TypeSwitch,
		Type: type_,
	}
}

/* This port will be used as a trigger or callable input port
 * func = callback when the port was being called as a function
 */
func Trigger(callback func(...any)) *Feature {
	return &Feature{
		Id:   TypeTrigger,
		Func: callback,
	}
}

/* This port can allow multiple different types
 * like an 'any' port, but can only contain one value
 */
func Union(types []reflect.Kind) *Feature {
	return &Feature{
		Id:    TypeUnion,
		Types: types,
	}
}

/* This port will allow any value to be passed to a function
 * then you can write custom data validation in the function
 * the value returned by your function will be used as the input value
 */
func Validator(type_ reflect.Kind, callback func(any) any) *Feature {
	return &Feature{
		Id:   TypeValidator,
		Type: type_,
		Func: callback,
	}
}
