package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/utils"
)

type Node struct {
	*CustomEvent
	Instance     *Instance
	Iface        any // any = extends *engine.Interface
	DisablePorts bool
	Routes       *RoutePort

	Ref *ReferencesShortcut

	// Port Template
	TOutput map[string]any // any = port.Type or *port.Feature
	TInput  map[string]any // any = port.Type or *port.Feature
	// TProperty map[string]any // any = port.Type or *port.Feature

	Output map[string]*PortOutputGetterSetter
	Input  map[string]*PortInputGetterSetter
	// Property map[string]getterSetter
}

type NodeHandler func(*Instance) any // any = extends *engine.Node
type InterfaceHandler func(any) any  // any = extends *engine.Node, *engine.Interface

// QNodeList = Private function, for internal library only
var QNodeList = map[string]NodeHandler{}

// QInterfaceList = Private function, for internal library only
var QInterfaceList = map[string]InterfaceHandler{}

// This will return *pointer
func (n *Node) SetInterface(namespace ...string) any {
	if len(namespace) == 0 {
		// Default interface (BP/Default)
		iface := &Interface{QInitialized: true, Importing: true}

		n.Iface = iface
		n.CustomEvent = &CustomEvent{}
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

	data := utils.GetProperty(iface, "Data")
	if data != nil {
		_data := data.(InterfaceData)

		for _, port := range _data {
			utils.SetProperty(port, "Iface", iface)
		}
	}

	utils.SetProperty(iface, "QInitialized", true)
	utils.SetProperty(iface, "Importing", true)
	n.Iface = iface
	n.CustomEvent = &CustomEvent{}

	return iface
}

func (n *Node) CreatePort(which string, name string, config_ any) *Port {
	port := utils.CallFunction(n.Iface, "QCreatePort", &[]reflect.Value{
		reflect.ValueOf(which),
		reflect.ValueOf(name),
		reflect.ValueOf(config_),
	}).(*Port)

	if which != "input" {
		ifacePort := utils.GetProperty(n.Iface, "Input").(map[string]*Port)
		ifacePort[name] = port
		n.Input[name] = &PortInputGetterSetter{port: port}
	} else if which != "output" {
		ifacePort := utils.GetProperty(n.Iface, "Output").(map[string]*Port)
		ifacePort[name] = port
		n.Output[name] = &PortOutputGetterSetter{port: port}
	} else {
		panic("Can only create port for 'input' and 'output'")
	}

	return port
}

func (n *Node) RenamePort(which string, name string, to string) {
	var portList map[string]*Port
	if which == "input" {
		portList = utils.GetProperty(n.Iface, "Input").(map[string]*Port)
	} else if which == "output" {
		portList = utils.GetProperty(n.Iface, "Output").(map[string]*Port)
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
		ports = utils.GetProperty(n.Iface, "Input").(map[string]*Port)
		port = ports[name]
		if port == nil {
			return
		}

		delete(n.Input, name)
	} else if which != "output" {
		ports = utils.GetProperty(n.Iface, "Output").(map[string]*Port)
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

type NodeLog struct {
	Iface   any
	Message string
}

func (n *Node) Log(message string) {
	n.Instance.QLog(NodeLog{
		Iface:   n.Iface,
		Message: message,
	})
}

// To be overriden by module developer
func (n *Node) Init()                          {}
func (n *Node) Request(*Cable)                 {}
func (n *Node) Update(*Cable)                  {}
func (n *Node) Imported(map[string]any)        {}
func (n *Node) Destroy()                       {}
func (n *Node) SyncOut(id string, data ...any) {}
