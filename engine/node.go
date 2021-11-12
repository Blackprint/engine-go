package engine

type Node struct {
	customEvent
	Instance *Instance
	Iface    interface{}

	Output   map[string]interface{} // interface = port value
	Input    map[string]interface{} // interface = port value
	Property map[string]interface{} // interface = port value
}

type INode interface {
	SetInterface(namespace ...string) interface{}
	Init()
	Request(*Port, *Interface)
	Update(Cable)
	Imported()
}
type INodeInternal interface {
	INode
	Obj() *Node
}

type NodeHandler func(*Instance) interface{}
type InterfaceHandler func(interface{}) interface{}

// QNodeList = Private function, for internal library only
var QNodeList = map[string]NodeHandler{} // interface = extends 'Node'

// QInterfaceList = Private function, for internal library only
var QInterfaceList = map[string]InterfaceHandler{}

func (n *Node) Obj() *Node {
	return n
}

func (n *Node) SetInterface(namespace ...string) interface{} {
	if len(namespace) == 0 {
		return Interface{}
	}

	name := namespace[0]
	class := QInterfaceList[name]
	if class == nil {
		panic("Node interface for '" + name + "' was not found, maybe .registerInterface() haven't being called?")
	}

	iface := class(n).(Interface)
	iface.Obj().QInitialized = true
	n.Iface = iface

	return &iface
}

// To be overriden
func (n *Node) Init()                     {}
func (n *Node) Request(*Port, *Interface) {}
func (n *Node) Update(Cable)              {}
func (n *Node) Imported()                 {}
