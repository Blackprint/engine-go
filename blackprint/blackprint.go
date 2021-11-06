package blackprint

import (
	"strings"

	"github.com/blackprint/engine-go/engine"
)

func RegisterNode(namespace string, constructor func(*engine.Instance) interface{}) {
	engine.QNodeList[namespace] = constructor
}

func RegisterInterface(namespace string, constructor func(interface{}) interface{}) {
	if strings.HasPrefix(namespace, "BPIC/") == false {
		panic(namespace + ": The first parameter of 'RegisterInterface' must be started with BPIC to avoid name conflict. Please name the interface similar with 'templatePrefix' for your module that you have set on 'blackprint.config.js'.")
	}

	engine.QInterfaceList[namespace] = constructor
}
