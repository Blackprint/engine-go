package utils

import (
	"reflect"
)

// obj must be a pointer
func SetProperty(obj interface{}, key string, value interface{}) {
	reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(value))
}

// obj must be a pointer
func GetProperty(obj interface{}, key string) interface{} {
	return reflect.ValueOf(obj).Elem().FieldByName(key).Interface()
}

// obj must be a pointer
func HasProperty(obj interface{}, key string) bool {
	return reflect.ValueOf(obj).Elem().FieldByName(key).IsValid()
}
