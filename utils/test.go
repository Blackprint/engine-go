package utils

import "fmt"

type GetterSetter interface {
	Set(val any)
	Get() any
}

type MyGetterSetter struct {
	val any
	aa  []*map[string]GetterSetter
}

func (gs *MyGetterSetter) Set(val any) {
	gs.val = val
}

func (gs *MyGetterSetter) Get() any {
	return gs.aa == nil
	// return gs.val
}

func main() {
	var aa map[string]GetterSetter
	bb := []*map[string]GetterSetter{nil}

	aa = map[string]GetterSetter{
		"Result": &MyGetterSetter{val: 123, aa: bb},
	}
	bb[0] = &aa

	aa["Result"].Set("Hello")
	fmt.Println(aa["Result"].Get().(bool))
}
