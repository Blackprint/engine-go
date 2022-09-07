package blackprint

import (
	"strconv"

	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/engine/nodes"
	portTypes "github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
)

type fnVarInput struct {
	*engine.Node
}

func (f *fnVarInput) Imported(data map[string]any) {
	if f.Routes != nil {
		f.Routes.Disabled = true
	}
}

func (f *fnVarInput) Request(cable *engine.Cable) {
	iface := f.Iface

	// This will trigger the port to request from outside and assign to this node's port
	f.Output["Val"].Set(iface.QParentFunc.Node.Input[iface.Data["name"].Get().(string)])
}

type fnVarOutput struct {
	*engine.Node
}

func (f *fnVarOutput) Update(c *engine.Cable) {
	id := f.Iface.Data["name"].Get()
	f.RefOutput[id].Set(f.Ref.Input["Val"].Get())
}

type bpFnVarInOut struct {
	*engine.Interface
	OnConnect func(*engine.Cable, *engine.Port)

	QParentFunc   any // => *engine.Interface
	QProxyIface   any
	QListener     func(any)
	QWaitPortInit func(*engine.Port)
}

func (f *bpFnVarInOut) Imported(data map[string]any) {
	if data["name"] == nil {
		panic("Parameter 'name' is required")
	}

	b.Data["name"].Set(data["name"])
	f.QParentFunc = f.Node.Instance.QFuncMain
}

type fnVarInputIface struct {
	*bpFnVarInOut
}

func (f *fnVarInputIface) Imported(data map[string]any) {
	f.bpFnVarInOut.Imported(data)
	ports := f.QParentFunc.Ref.IInput
	node := f.Node

	f.QProxyIface = f.QParentFunc.QProxyInput.Iface

	// Create temporary port if the main function doesn't have the port
	name := data["name"].(string)
	if _, exist := ports[name]; !exist {
		iPort := node.CreatePort("input", "Val", types.Any)
		proxyIface := f.QProxyIface

		// Run when $this node is being connected with other node
		iPort.OnConnect = func(cable *engine.Cable, port *engine.Port) {
			iPort.OnConnect = nil
			proxyIface.Off("_add."+name, iPort.QWaitPortInit)
			iPort.QWaitPortInit = nil

			cable.Disconnect()
			node.DeletePort("output", "Val")

			portName := &engine.RefPortName{Name: name}
			portType := getFnPortType(port, "input", f.QParentFunc, portName)
			newPort := node.CreatePort("output", "Val", portType)
			newPort.Name_ = portName
			newPort.ConnectPort(port)

			proxyIface.AddPort(port, name)
			f.QAddListener()
		}

		// Run when main node is the missing port
		iPort.QWaitPortInit = func(port *engine.Port) {
			iPort.OnConnect = nil
			iPort.QWaitPortInit = nil

			backup := []*engine.Port{}
			for _, val := range f.Output["Val"].Cables {
				backup = append(backup, val.Input)
			}

			node := f.Node
			node.DeletePort("output", "Val")

			portType := getFnPortType(port, "input", f.QParentFunc, port.Name_)

			newPort := node.CreatePort("output", "Val", portType)
			f.QAddListener()

			for _, val := range backup {
				newPort.ConnectPort(val)
			}
		}

		proxyIface.Once("_add."+name, iPort.QWaitPortInit)
	} else {
		if _, exist := f.Output["Val"]; !exist {
			port := ports[name]
			portType := getFnPortType(port, "input", f.QParentFunc, port.Name_)
			newPort := node.CreatePort("input", "Val", portType)
		}

		f.QAddListener()
	}
}

func (f *fnVarInputIface) QAddListener() {
	port := f.QProxyIface.Output[f.Data["name"].Get().(string)].(*engine.Port)

	if port.Feature == engine.PortTypeTrigger {
		f.QListener = func(p any) {
			f.Ref.Output["Val"].Call()
		}

		port.On("call", f.QListener)
	} else {
		f.QListener = func(ev any) {
			port := ev.(*engine.PortSelfEvent).Port
			if port.Iface.Node.Routes.Out == nil {
				val := f.Ref.IOutput["Val"]
				val.Value = port.Value // Change value without trigger node.update

				for _, temp := range val.Cables {
					// Clear connected cable's cache
					temp.Input.QCache = nil
				}
				return
			}

			f.Ref.Output["Val"].Set(port.Value)
		}

		port.On("value", f.QListener)
	}
}

func (f *fnVarInputIface) Destroy() {
	f.bpFnVarInOut.Destroy()

	if f.QListener == nil {
		return
	}

	port := f.QProxyIface.Output[f.Data["name"].Get().(string)].(*engine.Port)
	if port.Feature == engine.PortTypeTrigger {
		port.Off("call", f.QListener)
	} else {
		port.Off("value", f.QListener)
	}
}

type fnVarOutputIface struct {
	*bpFnVarInOut
}

func (f *fnVarOutputIface) Imported(data map[string]any) {
	f.bpFnVarInOut.Imported(data)
	ports := f.QParentFunc.Ref.IOutput
	node := f.Node

	node.RefOutput = f.QParentFunc.Ref.Output

	// Create temporary port if the main function doesn't have the port
	name := data["name"].(string)
	if _, exist := ports[name]; !exist {
		iPort := node.CreatePort("input", "Val", types.Any)
		proxyIface := f.QParentFunc.QProxyOutput.Iface

		// Run when $this node is being connected with other node
		iPort.OnConnect = func(cable *engine.Cable, port *engine.Port) {
			iPort.OnConnect = nil
			proxyIface.Off("_add."+name, iPort.QWaitPortInit)
			iPort.QWaitPortInit = nil

			cable.Disconnect()
			node.DeletePort("input", "Val")

			portName := &engine.RefPortName{Name: name}
			portType := getFnPortType(port, "output", f.QParentFunc, portName)
			newPort := node.CreatePort("input", "Val", portType)
			newPort.Name_ = portName
			newPort.ConnectPort(port)

			proxyIface.AddPort(port, name)
		}

		// Run when main node is the missing port
		iPort.QWaitPortInit = func(port *engine.Port) {
			iPort.OnConnect = nil
			iPort.QWaitPortInit = nil

			backup := []*engine.Port{}
			for _, val := range f.Output["Val"].Cables {
				backup = append(backup, val.Input)
			}

			node := f.Node
			node.DeletePort("input", "Val")

			portType := getFnPortType(port, "output", f.QParentFunc, port.Name_)

			newPort := node.CreatePort("input", "Val", portType)
			f.QAddListener()

			for _, val := range backup {
				newPort.ConnectPort(val)
			}
		}

		proxyIface.Once("_add."+name, iPort.QWaitPortInit)
	} else {
		if _, exist := f.Output["Val"]; !exist {
			port := ports[name]
			portType := getFnPortType(port, "output", f.QParentFunc, port.Name_)
			newPort := node.CreatePort("input", "Val", portType)
		}
	}
}

func getFnPortType(port *engine.Port, which string, parentNode any, ref *engine.ReferencesShortcut) any {
	if port.Feature == engine.PortTypeTrigger {
		if which == "input" { // Function Input (has output port inside, and input port on main node)
			return types.Function
		} else {
			return types.Function
		}
	} else {
		if port.Feature != 0 {
			if port.Feature == engine.PortTypeArrayOf {
				return portTypes.ArrayOf(port.Type)
			} else if port.Feature == engine.PortTypeDefault {
				return portTypes.Default(port.Type, port.Default)
			} else if port.Feature == engine.PortTypeTrigger {
				return portTypes.Trigger(port.Func)
			} else if port.Feature == engine.PortTypeUnion {
				return portTypes.Union(port.Types)
			} else if port.Feature == engine.PortTypeStructOf {
				return portTypes.StructOf(port.Type, port.Struct)
			} else if port.Feature == engine.PortTypeRoute {
				return portTypes.Route()
			} else {
				panic("Port feature was not found for: " + strconv.Itoa(port.Feature))
			}
		} else {
			return port.Type
		}
	}
}

func registerFnVarNode() {
	RegisterNode("BP/FnVar/Input", func(i *engine.Instance) any {
		node := &fnVarInput{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/FnVar/Input").(*fnVarInputIface)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name": &engine.GetterSetter{Value: ""},
		}

		iface.Title = "FnInput"
		iface.Type = "bp-fnvar-input"
		iface.QEnum = nodes.BPFnVarInput
		iface.QDynamicPort = true

		return node
	})

	RegisterInterface("BPIC/BP/FnVar/Input", func(node any) any {
		return &fnVarInputIface{
			bpFnVarInOut: &bpFnVarInOut{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})

	RegisterNode("BP/FnVar/Output", func(i *engine.Instance) any {
		node := &fnVarOutput{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/FnVar/Output").(*fnVarOutputIface)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name": &engine.GetterSetter{Value: ""},
		}

		iface.Title = "FnOutput"
		iface.Type = "bp-fnvar-output"
		iface.QEnum = nodes.BPFnVarOutput
		iface.QDynamicPort = true

		return node
	})

	RegisterInterface("BPIC/BP/FnVar/Output", func(node any) any {
		return &fnVarOutputIface{
			bpFnVarInOut: &bpFnVarInOut{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})
}
