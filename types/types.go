package types

import "reflect"

// type => reflect.Kind
const (
	Function = reflect.Func
	Array    = reflect.Array
	Object   = reflect.Map
	String   = reflect.String
	Bool     = reflect.Bool
	Int      = reflect.Int64
	Float    = reflect.Float64
	Any      = reflect.Interface
	Route    = 99999 // ToDo: do we need to fix this magic number?
)
