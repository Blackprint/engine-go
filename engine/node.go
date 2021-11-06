package engine

type Node struct {
	customEvent
	Instance *Instance
	Iface    interface{}

	Output   map[string]interface{} // interface = port value
	Input    map[string]interface{} // interface = port value
	Property map[string]interface{} // interface = port value
}

// QNodeList = Private function, for internal library only
var QNodeList map[string]func(*Instance) interface{} // interface = extends 'Node'

// QInterfaceList = Private function, for internal library only
var QInterfaceList map[string]func(interface{}) interface{}

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
	iface.QInitialized = true
	n.Iface = iface

	return &iface
}

// To be overriden
func (n *Node) Init()                     {}
func (n *Node) Request(*Port, *Interface) {}
func (n *Node) Update(Cable)              {}
func (n *Node) Imported()                 {}
