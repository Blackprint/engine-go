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

// https://stackoverflow.com/a/71184501/6563200
func IndexOf[T comparable](collection []T, el T) int {
	for i, x := range collection {
		if x == el {
			return i
		}
	}
	return -1
}

func IndexOfAny(collection []any, el any) int {
	pointer := reflect.ValueOf(el).Pointer()
	for i, x := range collection {
		if reflect.ValueOf(x).Pointer() == pointer {
			return i
		}
	}
	return -1
}

func Contains[T comparable](collection []T, el T) bool {
	for _, x := range collection {
		if x == el {
			return true
		}
	}
	return false
}

func ContainsAny(collection []any, el any) bool {
	pointer := reflect.ValueOf(el).Pointer()
	for _, x := range collection {
		if reflect.ValueOf(x).Pointer() == pointer {
			return true
		}
	}
	return false
}

func RemoveItem[T comparable](collection []T, el T) []T {
	i := IndexOf(collection, el)
	if i == -1 {
		return collection
	}

	return append(collection[:i], collection[i+1:]...)
}

func RemoveItemAny(collection []any, el any) []any {
	i := IndexOfAny(collection, el)
	if i == -1 {
		return collection
	}

	return append(collection[:i], collection[i+1:]...)
}

func RemoveItemAtIndex[T comparable](collection []T, i int) []T {
	return append(collection[:i], collection[i+1:]...)
}

func RemoveItemAtIndexAny(collection []any, i int) []any {
	return append(collection[:i], collection[i+1:]...)
}

func ClearMap[T any](collection map[string]T) {
	for key := range collection {
		delete(collection, key)
	}
}
