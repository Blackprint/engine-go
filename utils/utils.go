package utils

import (
	"reflect"
)

var EmptyArgs = []reflect.Value{}

func IsPointer(obj interface{}) bool {
	return reflect.ValueOf(obj).Kind() == reflect.Ptr
}

// obj must be a pointer
func SetProperty(obj interface{}, key string, value interface{}) {
	reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(value))
}

// obj must be a pointer
func GetProperty(obj interface{}, key string) interface{} {
	return reflect.ValueOf(obj).Elem().FieldByName(key).Interface()
}

// obj must be a pointer
func GetFunction(obj interface{}, key string) interface{} {
	return reflect.ValueOf(obj).MethodByName(key).Interface()
}

// obj must be a pointer
func CallFunction(obj interface{}, key string, val []reflect.Value) interface{} {
	return reflect.ValueOf(obj).MethodByName(key).Call(val)
}

// obj must be a pointer
func HasProperty(obj interface{}, key string) bool {
	return reflect.ValueOf(obj).Elem().FieldByName(key).IsValid()
}
