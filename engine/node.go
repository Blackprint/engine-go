package engine

type Node struct {
	customEvent
	Instance Instance
	IFace Interface

	Output map[string] interface{} {} // interface = port value
	Input map[string] interface{} {} // interface = port value
	Property map[string] interface{} {} // interface = port value
}

// QNodeList = Private function, for internal library only
var QNodeList := map[string] interface{} {}; // interface = extends 'Node'
func NewNode(namespace string) interface{} {
	class := QNodeList[namespace]
	if(class == nil)
		panic("Node for '"+namespace+"' was not found, maybe .registerNode() haven't being called?");

	return &class{}
}

// QInterfaceList = Private function, for internal library only
var QInterfaceList := map[string] interface{} {};
func (n *Node) SetInterface(namespace string) interface{} {
	class := QInterfaceList[namespace]
	if(class == nil)
		panic("Node interface for '"+namespace+"' was not found, maybe .registerInterface() haven't being called?");

	iface = class{Node: n};
	n.IFace = iface;

	return &iface;
}

// To be overriden
func (n *Node) Init(namespace string) interface{} {
	return nil
}

func (n *Node) Request(namespace string) interface{} {
	return nil
}

func (n *Node) Update(namespace string) interface{} {
	return nil
}

func (n *Node) Imported(namespace string) interface{} {
	return nil
}