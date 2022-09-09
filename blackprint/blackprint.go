package blackprint

import (
	"strings"

	"github.com/blackprint/engine-go/engine"
)

// The constructor must return pointer (ex: &Node{})
func RegisterNode(namespace string, meta *engine.NodeMetadata, constructor engine.NodeConstructor) {
	engine.QNodeList[namespace] = &engine.QNodeRegister{
		Metadata:    meta,
		Constructor: constructor,
	}
}

// The constructor must return pointer (ex: &any)
func RegisterInterface(namespace string, constructor engine.InterfaceConstructor) {
	if strings.HasPrefix(namespace, "BPIC/") == false {
		panic(namespace + ": The first parameter of 'RegisterInterface' must be started with BPIC to avoid name conflict. Please name the interface similar with 'templatePrefix' for your module that you have set on 'blackprint.config.js'.")
	}

	engine.QInterfaceList[namespace] = constructor
}

var Event = engine.Event
var Environment = engine.QEnvironment
