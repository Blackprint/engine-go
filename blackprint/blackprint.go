package blackprint

import (
	"strings"

	"github.com/blackprint/engine-go/engine"
)

// The constructor must return pointer (ex: &Node{})
func RegisterNode(namespace string, constructor func(*engine.Instance) any) {
	engine.QNodeList[namespace] = constructor
}

// The constructor must return pointer (ex: &any)
func RegisterInterface(namespace string, constructor func(any) any) {
	if strings.HasPrefix(namespace, "BPIC/") == false {
		panic(namespace + ": The first parameter of 'RegisterInterface' must be started with BPIC to avoid name conflict. Please name the interface similar with 'templatePrefix' for your module that you have set on 'blackprint.config.js'.")
	}

	engine.QInterfaceList[namespace] = constructor
}
