package engine

import (
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
)

type fnVarInput struct {
	*EmbedNode
}

func (f *fnVarInput) Imported(data map[string]any) {
	if f.Node.Routes != nil {
		f.Node.Routes.Disabled = true
	}
}

func (f *fnVarInput) Request(cable *Cable) {
	iface := f.Iface

	// This will trigger the port to request from outside and assign to this node's port
	f.Node.Output["Val"].Set(iface._parentFunc.Node.Input[iface.Data["name"].Get().(string)])
}

type fnVarOutput struct {
	*EmbedNode
}

func (f *fnVarOutput) Update(c *Cable) {
	id := f.Iface.Data["name"].Get().(string)
	f.Node.RefOutput[id].Set(f.Ref.Input["Val"].Get())
}

type bpFnVarInOut struct {
	*EmbedInterface
	_onConnect func(*Cable, *Port)

	_parentFunc   *Interface
	_proxyIface   *Interface
	_listener     func(any)
	_waitPortInit func(*Port)
	Type          string
}

func (f *bpFnVarInOut) Imported(data map[string]any) {
	if data["name"] == nil {
		panic("Parameter 'name' is required")
	}

	f.Iface.Data["name"].Set(data["name"])
	f._parentFunc = f.Node.Instance._funcMain
}

type fnVarInputIface struct {
	*bpFnVarInOut
}

func (f *fnVarInputIface) Imported(data map[string]any) {
	f.bpFnVarInOut.Imported(data)
	ports := f._parentFunc.Ref.IInput
	node := f.Node

	f._proxyIface = f._parentFunc._proxyInput.Iface

	// Create temporary port if the main function doesn't have the port
	name := data["name"].(string)
	if _, exist := ports[name]; !exist {
		iPort := node.CreatePort("input", "Val", types.Any)
		proxyIface := f._proxyIface

		// Run when $this node is being connected with other node
		iPort._onConnect = func(cable *Cable, port *Port) bool {
			iPort._onConnect = nil
			proxyIface.Off("_add."+name, iPort._waitPortInit)
			iPort._waitPortInit = nil

			cable.Disconnect()
			node.DeletePort("output", "Val")

			portName := &refPortName{Name: name}
			portType := getFnPortType(port, "input", f._parentFunc, portName)
			newPort := node.CreatePort("output", "Val", portType)
			newPort.Name_ = portName
			newPort.ConnectPort(port)

			proxyIface.Embed.(*qBpFnInOut).AddPort(port, name)
			f._addListener()

			return true
		}

		// Run when main node is the missing port
		iPort._waitPortInit = func(port *Port) {
			iPort._onConnect = nil
			iPort._waitPortInit = nil

			backup := []*Port{}
			for _, val := range f.Iface.Output["Val"].Cables {
				backup = append(backup, val.Input)
			}

			node := f.Node
			node.DeletePort("output", "Val")

			portType := getFnPortType(port, "input", f._parentFunc, port.Name_)
			newPort := node.CreatePort("output", "Val", portType)
			f._addListener()

			for _, val := range backup {
				newPort.ConnectPort(val)
			}
		}

		proxyIface.Once("_add."+name, iPort._waitPortInit)
	} else {
		if _, exist := f.Iface.Output["Val"]; !exist {
			port := ports[name]
			portType := getFnPortType(port, "input", f._parentFunc, port.Name_)
			node.CreatePort("input", "Val", portType)
		}

		f._addListener()
	}
}

func (f *fnVarInputIface) _addListener() {
	port := f._proxyIface.Output[f.Iface.Data["name"].Get().(string)]

	if port.Feature == PortTypeTrigger {
		f._listener = func(p any) {
			f.Ref.Output["Val"].Call()
		}

		port.On("call", f._listener)
	} else {
		f._listener = func(ev any) {
			port := ev.(*PortValueEvent).Port
			if port.Iface.Node.Routes.Out == nil {
				val := f.Ref.IOutput["Val"]
				val.Value = port.Value // Change value without trigger node.update

				for _, temp := range val.Cables {
					// Clear connected cable's cache
					temp.Input._cache = nil
				}
				return
			}

			f.Ref.Output["Val"].Set(port.Value)
		}

		port.On("value", f._listener)
	}
}

func (f *fnVarInputIface) Destroy() {
	f.bpFnVarInOut.Destroy()

	if f._listener == nil {
		return
	}

	port := f._proxyIface.Output[f.Iface.Data["name"].Get().(string)]
	if port.Feature == PortTypeTrigger {
		port.Off("call", f._listener)
	} else {
		port.Off("value", f._listener)
	}
}

type fnVarOutputIface struct {
	*bpFnVarInOut
}

func (f *fnVarOutputIface) Imported(data map[string]any) {
	f.bpFnVarInOut.Imported(data)
	ports := f._parentFunc.Ref.IOutput
	node := f.Node

	node.RefOutput = f._parentFunc.Ref.Output

	// Create temporary port if the main function doesn't have the port
	name := data["name"].(string)
	if _, exist := ports[name]; !exist {
		iPort := node.CreatePort("input", "Val", types.Any)
		proxyIface := f._parentFunc._proxyOutput.Iface

		// Run when $this node is being connected with other node
		iPort._onConnect = func(cable *Cable, port *Port) bool {
			iPort._onConnect = nil
			proxyIface.Off("_add."+name, iPort._waitPortInit)
			iPort._waitPortInit = nil

			cable.Disconnect()
			node.DeletePort("input", "Val")

			portName := &refPortName{Name: name}
			portType := getFnPortType(port, "output", f._parentFunc, portName)
			newPort := node.CreatePort("input", "Val", portType)
			newPort.Name_ = portName
			newPort.ConnectPort(port)

			proxyIface.Embed.(*qBpFnInOut).AddPort(port, name)
			return true
		}

		// Run when main node is the missing port
		iPort._waitPortInit = func(port *Port) {
			iPort._onConnect = nil
			iPort._waitPortInit = nil

			backup := []*Port{}
			for _, val := range f.Iface.Output["Val"].Cables {
				backup = append(backup, val.Input)
			}

			node := f.Node
			node.DeletePort("input", "Val")

			portType := getFnPortType(port, "output", f._parentFunc, port.Name_)
			newPort := node.CreatePort("input", "Val", portType)

			for _, val := range backup {
				newPort.ConnectPort(val)
			}
		}

		proxyIface.Once("_add."+name, iPort._waitPortInit)
	} else {
		if _, exist := f.Iface.Output["Val"]; !exist {
			port := ports[name]
			portType := getFnPortType(port, "output", f._parentFunc, port.Name_)
			node.CreatePort("input", "Val", portType)
		}
	}
}

func getFnPortType(port *Port, which string, parentNode *Interface, ref *refPortName) any {
	if port.Feature == PortTypeTrigger {
		if which == "input" { // Function Input (has output port inside, and input port on main node)
			return types.Function
		} else {
			return QPorts.Trigger(parentNode.Output[ref.Name]._callAll)
		}
	} else {
		if port.Feature != 0 {
			return port._getPortFeature()
		} else {
			return port.Type
		}
	}
}

func init() {
	QNodeList["BP/FnVar/Input"] = &NodeRegister{
		Output: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &fnVarInput{}

			iface := node.SetInterface("BPIC/BP/FnVar/Input")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "FnInput"
			iface.Embed.(*fnVarInputIface).Type = "bp-fnvar-input"
			iface._enum = nodes.BPFnVarInput
			iface._dynamicPort = true
		},
	}

	QInterfaceList["BPIC/BP/FnVar/Input"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &fnVarInputIface{
				bpFnVarInOut: &bpFnVarInOut{},
			}
		},
	}

	QNodeList["BP/FnVar/Output"] = &NodeRegister{
		Input: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &fnVarOutput{}

			iface := node.SetInterface("BPIC/BP/FnVar/Output")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "FnOutput"
			iface.Embed.(*fnVarOutputIface).Type = "bp-fnvar-output"
			iface._enum = nodes.BPFnVarOutput
			iface._dynamicPort = true
		},
	}

	QInterfaceList["BPIC/BP/FnVar/Output"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &fnVarOutputIface{
				bpFnVarInOut: &bpFnVarInOut{},
			}
		},
	}
}
