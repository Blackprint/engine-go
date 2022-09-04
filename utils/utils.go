package utils

import (
	"reflect"
)

var EmptyArgs = &[]reflect.Value{}

func IsPointer(obj any) bool {
	return reflect.ValueOf(obj).Kind() == reflect.Ptr
}

// obj must be a pointer
func SetProperty(obj any, key string, value any) {
	reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(value))
}

// obj must be a pointer
func GetProperty(obj any, key string) any {
	return reflect.ValueOf(obj).Elem().FieldByName(key).Interface()
}

func GetPropertyRef(obj any, key string) any {
	return reflect.ValueOf(obj).Elem().FieldByName(key).Addr().Interface()
}

// obj must be a pointer
func GetFunction(obj any, key string) any {
	return reflect.ValueOf(obj).MethodByName(key).Interface()
}

// obj must be a pointer
func CallFunction(obj any, key string, val *[]reflect.Value) any {
	return reflect.ValueOf(obj).MethodByName(key).Call(*val)
}

// obj must be a pointer
func CallFieldFunction(obj any, key string, val *[]reflect.Value) any {
	return reflect.ValueOf(obj).Elem().FieldByName(key).Elem().Call(*val)
}

// obj must be a pointer
func HasProperty(obj any, key string) bool {
	return reflect.ValueOf(obj).Elem().FieldByName(key).IsValid()
}
