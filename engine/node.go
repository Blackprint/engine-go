package engine

import (
	"github.com/blackprint/engine-go/utils"
)

type embedNode interface {
	Init()
	Request(*Cable)
	Update(*Cable)
	Imported(map[string]any)
	Destroy()
	SyncIn(id string, data ...any)
}

type EmbedNode struct {
	embedNode
	Node  *Node
	Iface *Interface
	Ref   *referencesShortcut
}

// To be overriden by module developer
func (n *EmbedNode) Init()                         {}
func (n *EmbedNode) Request(*Cable)                {}
func (n *EmbedNode) Update(*Cable)                 {}
func (n *EmbedNode) Imported(map[string]any)       {}
func (n *EmbedNode) Destroy()                      {}
func (n *EmbedNode) SyncIn(id string, data ...any) {}

type Node struct {
	Instance     *Instance
	Iface        *Interface
	DisablePorts bool
	Routes       *routePort
	Embed        embedNode

	Ref *referencesShortcut

	Output map[string]*portOutputGetterSetter
	Input  map[string]*portInputGetterSetter
	// Property map[string]getterSetter

	// For internal library use only
	QFuncInstance *bpFunction
	RefOutput     map[string]*portOutputGetterSetter
	// RefInput      map[string]*portInputGetterSetter
}

// Proxies, for library only
func (n *Node) Init()                         { n.Embed.Init() }
func (n *Node) Request(c *Cable)              { n.Embed.Request(c) }
func (n *Node) Update(c *Cable)               { n.Embed.Update(c) }
func (n *Node) Imported(d map[string]any)     { n.Embed.Imported(d) }
func (n *Node) Destroy()                      { n.Embed.Destroy() }
func (n *Node) SyncIn(id string, data ...any) { n.Embed.SyncIn(id, data) }

type NodeRegister struct {
	// Port Template
	Output PortTemplate
	Input  PortTemplate
	// Property *PortTemplate

	Constructor nodeConstructor
}

type InterfaceRegister struct {
	Constructor interfaceConstructor
}

type nodeConstructor func(*Node)
type interfaceConstructor func(*Interface)

// QNodeList = Private function, for internal library only
var QNodeList = map[string]*NodeRegister{}

// QInterfaceList = Private function, for internal library only
var QInterfaceList = map[string]*InterfaceRegister{}

func (n *Node) SetInterface(namespace ...string) *Interface {
	iface := &Interface{QInitialized: true, Importing: true}

	// Default interface (BP/Default)
	if len(namespace) == 0 {
		n.Iface = iface
		return iface
	}

	name := namespace[0]
	class := QInterfaceList[name]
	if class == nil {
		panic("Node interface for '" + name + "' was not found, maybe .registerInterface() haven't being called?")
	}

	class.Constructor(iface)
	for _, val := range iface.Data {
		utils.SetProperty(val, "Iface", iface)
	}

	iface.QInitialized = true
	iface.Importing = true
	n.Iface = iface

	return iface
}

func (n *Node) CreatePort(which string, name string, config_ any) *Port {
	port := n.Iface.QCreatePort(which, name, config_)

	if which != "input" {
		ifacePort := n.Iface.Input
		ifacePort[name] = port
		n.Input[name] = &portInputGetterSetter{port: port}
	} else if which != "output" {
		ifacePort := n.Iface.Output
		ifacePort[name] = port
		n.Output[name] = &portOutputGetterSetter{port: port}
	} else {
		panic("Can only create port for 'input' and 'output'")
	}

	return port
}

func (n *Node) RenamePort(which string, name string, to string) {
	var portList map[string]*Port
	if which == "input" {
		portList = n.Iface.Input
	} else if which == "output" {
		portList = n.Iface.Output
	} else {
		panic("Can only rename port for 'input' and 'output'")
	}

	port := portList[name]
	if port == nil {
		panic(which + " port with name '" + name + "' was not found")
	}

	if portList[to] != nil {
		panic(which + " port with name '" + to + "' already exist")
	}

	portList[to] = port
	delete(portList, name)

	port.Name = to

	if which == "input" {
		n.Input[to] = n.Input[name]
		delete(n.Input, name)
	} else if which == "output" {
		n.Output[to] = n.Output[name]
		delete(n.Output, name)
	}
}

func (n *Node) DeletePort(which string, name string) {
	var ports map[string]*Port
	var port *Port

	if which != "input" {
		ports = n.Iface.Input
		port = ports[name]
		if port == nil {
			return
		}

		delete(n.Input, name)
	} else if which != "output" {
		ports = n.Iface.Output
		port = ports[name]
		if port == nil {
			return
		}

		delete(n.Output, name)
	} else {
		panic("Can only delete port for 'input' and 'output'")
	}

	port.DisconnectAll()
	delete(ports, name)
}

func (n *Node) Log(message string) {
	n.Instance.QLog(n.Iface, message)
}
