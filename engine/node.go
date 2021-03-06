package engine

import (
	"github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/utils"
)

type Node struct {
	*customEvent
	Instance *Instance
	IFace    interface{} // interface = extends *engine.Interface

	// Port Template
	TOutput   map[string]interface{} // interface = port.Type or *port.Feature
	TInput    map[string]interface{} // interface = port.Type or *port.Feature
	TProperty map[string]interface{} // interface = port.Type or *port.Feature

	Output   map[string]port.GetterSetter
	Input    map[string]port.GetterSetter
	Property map[string]port.GetterSetter
}

type NodeHandler func(*Instance) interface{}        // interface = extends *engine.Node
type InterfaceHandler func(interface{}) interface{} // interface = extends *engine.Node, *engine.Interface

// QNodeList = Private function, for internal library only
var QNodeList = map[string]NodeHandler{}

// QInterfaceList = Private function, for internal library only
var QInterfaceList = map[string]InterfaceHandler{}

// This will return *pointer
func (n *Node) SetInterface(namespace ...string) interface{} {
	if len(namespace) == 0 {
		// Default interface (BP/Default)
		iface := &Interface{QInitialized: true, Importing: true}

		n.IFace = iface
		n.customEvent = &customEvent{}
		return iface
	}

	name := namespace[0]
	class := QInterfaceList[name]
	if class == nil {
		panic("Node interface for '" + name + "' was not found, maybe .registerInterface() haven't being called?")
	}

	iface := class(n)
	if utils.IsPointer(iface) == false {
		panic(".registerInterface() must return pointer")
	}

	utils.SetProperty(iface, "QInitialized", true)
	utils.SetProperty(iface, "Importing", true)
	n.IFace = iface
	n.customEvent = &customEvent{}

	return iface
}

// To be overriden
func (n *Node) Init()                      {}
func (n *Node) Request(*Port, interface{}) {} // interface{} => extends engine.Interface
func (n *Node) Update(*Cable)              {}
func (n *Node) Imported()                  {}
