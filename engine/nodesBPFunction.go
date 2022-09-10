package engine

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

// Main function node
type bpFunctionNode struct { // Main function node -> BPI/F/{FunctionName}
	*EmbedNode
	Type string
}

func (b *bpFunctionNode) Init() {
	if b.Iface.Embed.(*FnMain)._importOnce {
		b.Iface._bpFnInit()
	}
}

func (b *bpFunctionNode) Imported(data map[string]any) {
	ins := b.Node._funcInstance
	ins.Used = append(ins.Used, b.Node.Iface)
}

func (b *bpFunctionNode) Update(cable *Cable) {
	iface := b.Iface._proxyInput.Iface
	Output := iface.Node.Output

	if cable == nil { // Triggered by port route
		thisInput := b.Node.Input

		// Sync all port value
		for key, value := range iface.Output {
			if value.Type == types.Function {
				continue
			}

			Output[key].Set(thisInput[key].Get())
		}

		return
	}

	// Update output value on the input node inside the function node
	Output[cable.Input.Name].Set(cable.GetValue())
}

func (b *bpFunctionNode) Destroy() {
	ins := b.Node._funcInstance
	utils.RemoveItem(ins.Used, b.Node.Iface)
}

// used for instance.createFunction
type bpFunction struct { // <= _funcInstance
	*CustomEvent
	Id           string
	Title        string
	Type         int
	Used         []*Interface
	Input        PortTemplate
	Output       PortTemplate
	Structure    singleInstanceJSON
	Variables    map[string]*BPVariable
	privateVars  []string
	RootInstance *Instance
	Node         func(*Instance) *Node // Node constructor

	// for internal library use only
	_syncing bool
}

func (b *bpFunction) _onFuncChanges(eventName string, obj any, fromNode *Node) {
	for _, iface_ := range b.Used {
		if iface_.Node == fromNode {
			continue
		}

		nodeInstance := iface_._bpInstance
		// nodeInstance.PendingRender = true // Force recalculation for cable position

		if eventName == "cable.connect" || eventName == "cable.disconnect" {
			cable := utils.GetProperty(obj, "Cable").(*Cable)
			input := cable.Input
			output := cable.Output
			ifaceList := fromNode.Iface._bpInstance.IfaceList

			// Skip event that also triggered when deleting a node
			if input.Iface._bpDestroy || output.Iface._bpDestroy {
				continue
			}

			inputIface := nodeInstance.IfaceList[utils.IndexOf(ifaceList, input.Iface)]
			if inputIface == nil {
				panic("Failed to get node input iface index")
			}

			outputIface := nodeInstance.IfaceList[utils.IndexOf(ifaceList, output.Iface)]
			if outputIface == nil {
				panic("Failed to get node output iface index")
			}

			if inputIface.Namespace != input.Iface.Namespace {
				fmt.Println(inputIface.Namespace + " !== " + input.Iface.Namespace)
				panic("Input iface namespace was different")
			}

			if outputIface.Namespace != output.Iface.Namespace {
				fmt.Println(outputIface.Namespace + " !== " + output.Iface.Namespace)
				panic("Output iface namespace was different")
			}

			if eventName == "cable.connect" {
				targetInput := inputIface.Input[input.Name]
				targetOutput := outputIface.Output[output.Name]

				if targetInput == nil {
					if inputIface._enum == nodes.BPFnOutput {
						targetInput = inputIface.Embed.(*qBpFnInOut).AddPort(targetOutput, output.Name)
					} else {
						panic("Output port was not found")
					}
				}

				if targetOutput == nil {
					if outputIface._enum == nodes.BPFnInput {
						targetOutput = outputIface.Embed.(*qBpFnInOut).AddPort(targetInput, input.Name)
					} else {
						panic("Input port was not found")
					}
				}

				targetInput.ConnectPort(targetOutput)
			} else if eventName == "cable.disconnect" {
				cables := inputIface.Input[input.Name].Cables
				outputPort := outputIface.Output[output.Name]

				for _, cable := range cables {
					if cable.Output == outputPort {
						cable.Disconnect()
						break
					}
				}
			}
		} else if eventName == "node.created" {
			iface := utils.GetProperty(obj, "Iface").(*Interface)
			nodeInstance.CreateNode(iface.Namespace, nodeConfig{
				Data: iface.Data,
			}, nil)
		} else if eventName == "node.delete" {
			objIface := utils.GetProperty(obj, "Iface").(*Interface)

			index := utils.IndexOf(fromNode.Iface._bpInstance.IfaceList, objIface)
			if index == -1 {
				panic("Failed to get node index")
			}

			iface := nodeInstance.IfaceList[index]
			if iface.Namespace != objIface.Namespace {
				fmt.Println(iface.Namespace + " " + objIface.Namespace)
				panic("Failed to delete node from other function instance")
			}

			if eventName == "node.delete" {
				nodeInstance.DeleteNode(iface)
			}
		}
	}
}

// func (b *bpFunction) CreateNode(instance *Instance, options nodeConfig) (*Interface, []*Interface) {
// 	return instance.CreateNode(b.Node, options, nil)
// }

type FnVarOptions struct {
	Scope int
}

func (b *bpFunction) CreateVariable(id string, options FnVarOptions) *BPVariable {
	if _, exist := b.Variables[id]; exist {
		panic("Variable id already exist: id")
	}

	// deepProperty

	temp := &BPVariable{
		Id: id,
		// options,
	}
	temp.FuncInstance = b

	if options.Scope == VarScopeShared {
		b.Variables[id] = temp
	} else {
		b.AddPrivateVars(id)
		return temp
	}

	b.Emit("variable.new", temp)
	b.RootInstance.Emit("variable.new", temp)
	return temp
}

type VariableNewEvent struct {
	Id      string
	ScopeId int
}

func (b *bpFunction) AddPrivateVars(id string) {
	if utils.Contains(b.privateVars, id) {
		return
	}

	b.privateVars = append(b.privateVars, id)

	temp := &VariableNewEvent{
		ScopeId: VarScopePrivate,
		Id:      id,
	}
	b.Emit("variable.new", temp)
	b.RootInstance.Emit("variable.new", temp)

	for _, iface := range b.Used {
		iface._bpInstance.Variables[id] = &BPVariable{Id: id}
	}
}

func (b *bpFunction) RefreshPrivateVars(instance *Instance) {
	vars := instance.Variables
	for _, id := range b.privateVars {
		vars[id] = &BPVariable{Id: id}
	}
}

func (b *bpFunction) RenamePort(which string, fromName string, toName string) {
	var main PortTemplate
	var proxyPort string
	if which == "output" {
		main = b.Output
		proxyPort = "Input"
	} else {
		main = b.Input
		proxyPort = "Output"
	}

	main[toName] = main[fromName]
	delete(main, fromName)

	for _, iface := range b.Used {
		iface.Node.RenamePort(which, fromName, toName)

		var temp *Node
		if which == "output" {
			temp = iface._proxyOutput
		} else {
			temp = iface._proxyInput
		}

		portList := utils.GetProperty(temp.Iface, proxyPort).(map[string]*Port)
		portList[fromName].Name_.Name = toName
		temp.RenamePort(proxyPort, fromName, toName)

		for _, proxyVar := range iface._bpInstance.IfaceList {
			if (which == "output" && proxyVar.Namespace != "BP/FnVar/Output") || (which == "input" && proxyVar.Namespace != "BP/FnVar/Input") {
				continue
			}

			if proxyVar.Data["name"].Get() != fromName {
				continue
			}
			proxyVar.Data["name"].Set(toName)

			if which == "output" {
				proxyVar.Input["Val"].Name_.Name = toName
			}
		}
	}
}

func (b *bpFunction) Destroy() {
	for _, iface := range b.Used {
		iface.Node.Instance.DeleteNode(iface)
	}
}

type qNodeInput struct {
	*EmbedNode
}

func (n *qNodeInput) Imported(data map[string]any) {
	input := n.Iface._funcMain.Node._funcInstance.Input

	for key, value := range input {
		n.Node.CreatePort("output", key, value)
	}
}

func (n *qNodeInput) Request(cable *Cable) {
	name := cable.Output.Name

	// This will trigger the port to request from outside and assign to this node's port
	n.Node.Output[name].Set(n.Iface._funcMain.Node.Input[name].Get())
}

type qNodeOutput struct {
	*EmbedNode
}

func (n *qNodeOutput) Imported(data map[string]any) {
	output := n.Iface._funcMain.Node._funcInstance.Output

	for key, value := range output {
		n.Node.CreatePort("input", key, value)
	}
}

func (n *qNodeOutput) Update(cable *Cable) {
	iface := n.Iface._funcMain
	if cable == nil { // Triggered by port route
		IOutput := iface.Output
		Output := iface.Node.Output
		thisInput := n.Node.Input

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
	*EmbedInterface
	_importOnce bool
	_save       func(any, string, bool)
	_portSw_    map[string]int
}

func (f *FnMain) _bpFnInit() {
	if f._importOnce {
		panic("Can't import function more than once")
	}

	f._importOnce = true
	node := f.Node

	f.Iface._bpInstance = New()
	bpFunction := node._funcInstance

	newInstance := f.Iface._bpInstance
	// newInstance.Variables = []; // private for one function
	newInstance.sharedVariables = bpFunction.Variables // shared between function
	newInstance.Functions = node.Instance.Functions
	newInstance._funcMain = f.Iface
	newInstance._mainInstance = bpFunction.RootInstance

	bpFunction.RefreshPrivateVars(newInstance)

	// swallowCopy := make([]any, len(bpFunction.Structure))
	// copy(swallowCopy, bpFunction.Structure)

	f.Iface._bpInstance.importParsed(bpFunction.Structure)

	// Init port switches
	if f._portSw_ != nil {
		f.Iface._initPortSwitches(f._portSw_)
		f._portSw_ = nil

		InputIface := f.Iface._proxyInput.Iface
		InputIface_ := InputIface.Embed.(*qBpFnInOut)

		if InputIface_._portSw_ != nil {
			InputIface._initPortSwitches(InputIface_._portSw_)
			InputIface_._portSw_ = nil
		}
	}

	f._save = func(ev any, eventName string, force bool) {
		if force || bpFunction._syncing {
			return
		}

		// ev.BpFunction = bpFunction
		newInstance._mainInstance.Emit(eventName, ev)

		bpFunction._syncing = true
		bpFunction._onFuncChanges(eventName, ev, f.Node)
		bpFunction._syncing = false
	}

	f.Iface._bpInstance.On("cable.connect cable.disconnect node.created node.delete node.id.changed", f._save)
}
func (f *FnMain) RenamePort(which string, fromName string, toName string) {
	f.Node._funcInstance.RenamePort(which, fromName, toName)
	f._save(false, "", true)
}

type qBpFnInOut struct {
	*EmbedInterface
	Type     string
	_portSw_ map[string]int
}

type addPortRef struct {
	Node *Node
	Port *Port
}

func (b *qBpFnInOut) AddPort(port *Port, customName string) *Port {
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
	if port.Feature == PortTypeTrigger {
		reff = &addPortRef{}
		portType = QPorts.Trigger(func(*Port) {
			reff.Node.Output[reff.Port.Name].Call()
		})
	} else {
		if port.Feature != 0 {
			portType = port._getPortFeature()
		} else {
			portType = port.Type
		}
	}

	var nodeA *Node
	var nodeB *Node
	// nodeA, nodeB; // Main (input) -> Input (output), Output (input) -> Main (output)
	if b.Type == "bp-fn-input" { // Main (input) -> Input (output)
		inc := 1
		for true {
			_, exist := b.Iface.Output[name]
			if !exist {
				break
			}

			name += strconv.Itoa(inc)
			inc++
		}

		nodeA = b.Iface._funcMain.Node
		nodeB = b.Node
		nodeA._funcInstance.Input[name] = portType
	} else { // Output (input) -> Main (output)
		inc := 1
		for true {
			_, exist := b.Iface.Input[name]
			if !exist {
				break
			}

			name += strconv.Itoa(inc)
			inc++
		}

		nodeA = b.Node
		nodeB = b.Iface._funcMain.Node
		nodeB._funcInstance.Output[name] = portType
	}

	outputPort := nodeB.CreatePort("output", name, portType)

	var inputPort *Port
	if portType == types.Function {
		inputPort = nodeA.CreatePort("input", name, QPorts.Trigger(outputPort._callAll))
	} else {
		inputPort = nodeA.CreatePort("input", name, portType)
	}

	if reff != nil {
		reff.Node = nodeB
		reff.Port = inputPort
	}

	if b.Type == "bp-fn-input" {
		outputPort.Name_ = &refPortName{Name: name} // When renaming port, this also need to be changed
		b.Iface.Emit("_add.{name}", outputPort)

		inputPort.On("value", func(ev PortValueEvent) {
			outputPort.Iface.Node.Output[outputPort.Name].Set(ev.Cable.Output.Value)
		})

		return outputPort
	}

	inputPort.Name_ = &refPortName{Name: name} // When renaming port, this also need to be changed
	b.Iface.Emit("_add.{name}", inputPort)
	return inputPort
}

func (b *qBpFnInOut) RenamePort(fromName string, toName string) {
	bpFunction := b.Iface._funcMain.Node._funcInstance
	// Main (input) -> Input (output)
	if b.Type == "bp-fn-input" {
		bpFunction.RenamePort("input", fromName, toName)
	} else { // Output (input) -> Main (output)
		bpFunction.RenamePort("output", fromName, toName)
	}
}

func (b *qBpFnInOut) DeletePort(name string) {
	funcMainNode := b.Iface._funcMain.Node
	if b.Type == "bp-fn-input" { // Main (input) -> Input (output)
		funcMainNode.DeletePort("input", name)
		b.Node.DeletePort("output", name)
		delete(funcMainNode._funcInstance.Input, name)
	} else { // Output (input) -> Main (output)
		funcMainNode.DeletePort("output", name)
		b.Node.DeletePort("input", name)
		delete(funcMainNode._funcInstance.Output, name)
	}
}

func init() {
	QNodeList["BP/Fn/Input"] = &NodeRegister{
		Output: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &qNodeInput{}

			iface := node.SetInterface("BPIC/BP/Fn/Input")
			iface._enum = nodes.BPFnInput
			iface._dynamicPort = true // Port is initialized dynamically

			iface.Title = "Input"
			iface.Embed.(*qBpFnInOut).Type = "bp-fn-input"
			iface._funcMain = node.Instance._funcMain
			iface._funcMain._proxyInput = node
		},
	}

	QInterfaceList["BPIC/BP/Fn/Input"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &qBpFnInOut{}
		},
	}

	QNodeList["BP/Fn/Output"] = &NodeRegister{
		Input: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &bpVarGet{}

			iface := node.SetInterface("BPIC/BP/Fn/Output")
			iface._enum = nodes.BPFnOutput
			iface._dynamicPort = true // Port is initialized dynamically

			iface.Title = "Output"
			iface.Embed.(*qBpFnInOut).Type = "bp-fn-output"
			iface._funcMain = node.Instance._funcMain
			iface._funcMain._proxyOutput = node
		},
	}

	QInterfaceList["BPIC/BP/Fn/Output"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &qBpFnInOut{}
		},
	}

	QInterfaceList["BPIC/BP/Fn/Main"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &FnMain{}
		},
	}
}
