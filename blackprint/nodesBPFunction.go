package blackprint

import (
	"strconv"
	"strings"

	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
)

type nodeInput struct {
	*engine.Node
}

func (n *nodeInput) imported(data any) {
	input := n.Iface.QFuncMain.Node.QFuncInstance.Input

	for key, value := range input {
		n.CreatePort("output", key, value)
	}
}

func (n *nodeInput) request(cable *engine.Cable) {
	name := cable.Output.Name

	// This will trigger the port to request from outside and assign to this node's port
	n.Output[name].Set(n.Iface.QFuncMain.Node.Input[name].Get())
}

type nodeOutput struct {
	*engine.Node
}

func (n *nodeOutput) imported(data map[string]any) {
	output := n.Iface.QFuncMain.Node.QFuncInstance.Output

	for key, value := range output {
		n.CreatePort("input", key, value)
	}
}

func (n *nodeOutput) update(cable *engine.Cable) {
	iface := n.Iface.QFuncMain
	if cable == nil { // Triggered by port route
		IOutput := iface.Output
		Output := iface.Node.Output
		thisInput := n.Input

		// Sync all port value
		for key, value := range IOutput {
			if value.Type == types.Function {
				continue
			}
			Output[key].Set(thisInput[key].Get())
		}

		return
	}

	iface.Node.Output[cable.Input.Name].Set(cable.GetValue())
}

type FnMain struct {
	*engine.Interface
	QImportOnce bool
	QSave       func(any, string, bool)
	QPortSw_    any
}

func (f *FnMain) QBpFnInit() {
	if f.QImportOnce {
		panic("Can't import function more than once")
	}

	f.QImportOnce = true
	node := f.Node

	f.QBpInstance = engine.New()
	bpFunction := node.QFuncInstance

	newInstance := f.QBpInstance
	// newInstance.Variables = []; // private for one function
	newInstance.SharedVariables = bpFunction.Variables // shared between function
	newInstance.Functions = node.Instance.Functions
	newInstance.QFuncMain = f
	newInstance.QMainInstance = bpFunction.RootInstance

	bpFunction.RefreshPrivateVars(newInstance)

	swallowCopy := make([]any, len(bpFunction.Structure))
	copy(swallowCopy, bpFunction.Structure)

	f.QBpInstance.ImportJSON(swallowCopy)

	// Init port switches
	if f.QPortSw_ != nil {
		f.QInitPortSwitches(f.QPortSw_)
		f.QPortSw_ = nil

		InputIface := f.QProxyInput.Iface
		if InputIface.QPortSw_ != nil {
			InputIface.QInitPortSwitches(InputIface.QPortSw_)
			InputIface.QPortSw_ = nil
		}
	}

	f.QSave = func(ev any, eventName string, force bool) {
		if force || bpFunction.QSyncing {
			return
		}

		ev.BpFunction = bpFunction
		newInstance.QMainInstance.Emit(eventName, ev)

		bpFunction.QSyncing = true
		bpFunction.QOnFuncChanges(eventName, ev, f.Node)
		bpFunction.QSyncing = false
	}

	f.QBpInstance.On("cable.connect cable.disconnect node.created node.delete node.id.changed", f.QSave)
}
func (f *FnMain) RenamePort(which string, fromName string, toName string) {
	f.Node.QFuncInstance.RenamePort(which, fromName, toName)
	f.QSave(false, "", true)
}

type bpFnInOut struct {
	*engine.Interface
}

type addPortRef struct {
	Node *engine.Node
	Port *engine.Port
}

func (b *bpFnInOut) AddPort(port *engine.Port, customName string) *engine.Port {
	if port == nil {
		panic("Can't set type with nil")
	}

	if strings.HasPrefix(port.Iface.Namespace, "BP/Fn") {
		panic("Function Input can't be connected directly to Output")
	}

	var name string
	if port.Name_ != nil {
		name = port.Name_.Name
	} else if customName != "" {
		name = customName
	} else {
		name = port.Name
	}

	var reff *addPortRef
	var portType any
	if port.Feature == engine.PortTypeTrigger {
		reff = &addPortRef{}
		portType = engine.Ports.Trigger(func(*engine.Port) {
			reff.Node.Output[reff.Port.Name].Call()
		})
	} else {
		if port.Feature != 0 {
			portType = port.QGetPortFeature()
		} else {
			portType = port.Type
		}
	}

	var nodeA *engine.Node
	var nodeB *engine.Node
	// nodeA, nodeB; // Main (input) -> Input (output), Output (input) -> Main (output)
	if b.Type == "bp-fn-input" { // Main (input) -> Input (output)
		inc := 1
		for true {
			_, exist := b.Output[name]
			if !exist {
				break
			}

			name += strconv.Itoa(inc)
			inc++
		}

		nodeA = b.QFuncMain.Node
		nodeB = b.Node
		nodeA.QFuncInstance.Input[name] = portType
	} else { // Output (input) -> Main (output)
		inc := 1
		for true {
			_, exist := b.Input[name]
			if !exist {
				break
			}

			name += strconv.Itoa(inc)
			inc++
		}

		nodeA = b.Node
		nodeB = b.QFuncMain.Node
		nodeB.QFuncInstance.Output[name] = portType
	}

	outputPort := nodeB.CreatePort("output", name, portType)

	var inputPort *engine.Port
	if portType == types.Function {
		inputPort = nodeA.CreatePort("input", name, engine.Ports.Trigger(outputPort.QCallAll))
	} else {
		inputPort = nodeA.CreatePort("input", name, portType)
	}

	if reff != nil {
		reff.Node = nodeB
		reff.Port = inputPort
	}

	if b.Type == "bp-fn-input" {
		outputPort.Name_ = &RefPortName{Name: name} // When renaming port, this also need to be changed
		b.Emit("_add.{name}", outputPort)

		inputPort.On("value", func(ev engine.PortValueEvent) {
			outputPort.Iface.Node.Output[outputPort.Name](ev.Cable.Output.Value)
		})

		return outputPort
	}

	inputPort.Name_ = &RefPortName{Name: name} // When renaming port, this also need to be changed
	b.Emit("_add.{name}", inputPort)
	return inputPort
}

func (b *bpFnInOut) renamePort(fromName string, toName string) {
	bpFunction := b.QFuncMain.Node.QFuncInstance
	// Main (input) -> Input (output)
	if b.Type == "bp-fn-input" {
		bpFunction.RenamePort("input", fromName, toName)
	} else { // Output (input) -> Main (output)
		bpFunction.RenamePort("output", fromName, toName)
	}
}

func (b *bpFnInOut) deletePort(name string) {
	funcMainNode := b.QFuncMain.Node
	if b.Type == "bp-fn-input" { // Main (input) -> Input (output)
		funcMainNode.DeletePort("input", name)
		b.Node.DeletePort("output", name)
		delete(funcMainNode.QFuncInstance.Input, name)
	} else { // Output (input) -> Main (output)
		funcMainNode.DeletePort("output", name)
		b.Node.DeletePort("input", name)
		delete(funcMainNode.QFuncInstance.Output, name)
	}
}

func registerBpFuncNode() {
	RegisterNode("BP/Fn/Input", func(i *engine.Instance) any {
		node := &nodeInput{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Fn/Input").(*fnInput)
		iface.QEnum = nodes.BPFnInput
		iface.QProxyInput = true // Port is initialized dynamically
		iface.QDynamicPort = true

		iface.Title = "Input"
		iface.Type = "bp-fn-input"
		iface.QFuncMain = i.QFuncMain
		i.QFuncMain.QProxyInput = node

		return node
	})

	RegisterInterface("BPIC/BP/Fn/Input", func(node any) any {
		return &bpFnInOut{
			Interface: &engine.Interface{
				Node: node,
			},
		}
	})

	RegisterNode("BP/Fn/Output", func(i *engine.Instance) any {
		node := &bpVarGet{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Fn/Output").(*fnOutput)
		iface.QEnum = nodes.BPFnOutput
		iface.QDynamicPort = true // Port is initialized dynamically

		iface.Title = "Output"
		iface.Type = "bp-fn-output"
		iface.QFuncMain = i.QFuncMain
		i.QFuncMain.QProxyOutput = node

		return node
	})

	RegisterInterface("BPIC/BP/Fn/Output", func(node any) any {
		return &bpFnInOut{
			Interface: &engine.Interface{
				Node: node,
			},
		}
	})

	RegisterInterface("BPIC/BP/Fn/Main", func(node any) any {
		return &FnMain{
			Interface: &engine.Interface{
				Node: node,
			},
		}
	})
}
