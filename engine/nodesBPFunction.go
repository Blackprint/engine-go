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
type BPFunctionNode struct { // Main function node -> BPI/F/{FunctionName}
	*EmbedNode
	Type string
}

func (b *BPFunctionNode) Init() {
	if b.Iface.Embed.(*FnMain).QImportOnce {
		b.Iface.QBpFnInit()
	}
}

func (b *BPFunctionNode) Imported(data map[string]any) {
	ins := b.Node.QFuncInstance
	ins.Used = append(ins.Used, b.Node.Iface)
}

func (b *BPFunctionNode) Update(cable *Cable) {
	iface := b.Iface.QProxyInput.Iface
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

func (b *BPFunctionNode) Destroy() {
	ins := b.Node.QFuncInstance
	utils.RemoveItem(ins.Used, b.Node.Iface)
}

// used for instance.createFunction
type BPFunction struct { // <= _funcInstance
	*CustomEvent
	Id           string
	Title        string
	Type         int
	Used         []*Interface
	Input        NodePortTemplate
	Output       NodePortTemplate
	Structure    SingleInstanceJSON
	Variables    map[string]*BPVariable
	PrivateVars  []string
	RootInstance *Instance
	Node         func(*Instance) *Node // Node constructor

	// for internal library use only
	QSyncing bool
}

func (b *BPFunction) QOnFuncChanges(eventName string, obj any, fromNode *Node) {
	for _, iface_ := range b.Used {
		if iface_.Node == fromNode {
			continue
		}

		nodeInstance := iface_.QBpInstance
		// nodeInstance.PendingRender = true // Force recalculation for cable position

		if eventName == "cable.connect" || eventName == "cable.disconnect" {
			cable := utils.GetProperty(obj, "Cable").(*Cable)
			input := cable.Input
			output := cable.Output
			ifaceList := fromNode.Iface.QBpInstance.IfaceList

			// Skip event that also triggered when deleting a node
			if input.Iface.QBpDestroy || output.Iface.QBpDestroy {
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
					if inputIface.QEnum == nodes.BPFnOutput {
						targetInput = inputIface.Embed.(*QBpFnInOut).AddPort(targetOutput, output.Name)
					} else {
						panic("Output port was not found")
					}
				}

				if targetOutput == nil {
					if outputIface.QEnum == nodes.BPFnInput {
						targetOutput = outputIface.Embed.(*QBpFnInOut).AddPort(targetInput, input.Name)
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

			index := utils.IndexOf(fromNode.Iface.QBpInstance.IfaceList, objIface)
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

// func (b *BPFunction) CreateNode(instance *Instance, options nodeConfig) (*Interface, []*Interface) {
// 	return instance.CreateNode(b.Node, options, nil)
// }

type FnVarOptions struct {
	Scope int
}

func (b *BPFunction) CreateVariable(id string, options FnVarOptions) *BPVariable {
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

type EvVariableNew struct {
	Id      string
	ScopeId int
}

func (b *BPFunction) AddPrivateVars(id string) {
	if utils.Contains(b.PrivateVars, id) {
		return
	}

	b.PrivateVars = append(b.PrivateVars, id)

	temp := &EvVariableNew{
		ScopeId: VarScopePrivate,
		Id:      id,
	}
	b.Emit("variable.new", temp)
	b.RootInstance.Emit("variable.new", temp)

	for _, iface := range b.Used {
		iface.QBpInstance.Variables[id] = &BPVariable{Id: id}
	}
}

func (b *BPFunction) RefreshPrivateVars(instance *Instance) {
	vars := instance.Variables
	for _, id := range b.PrivateVars {
		vars[id] = &BPVariable{Id: id}
	}
}

func (b *BPFunction) RenamePort(which string, fromName string, toName string) {
	var main NodePortTemplate
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
			temp = iface.QProxyOutput
		} else {
			temp = iface.QProxyInput
		}

		portList := utils.GetProperty(temp.Iface, proxyPort).(map[string]*Port)
		portList[fromName].Name_.Name = toName
		temp.RenamePort(proxyPort, fromName, toName)

		for _, proxyVar := range iface.QBpInstance.IfaceList {
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

func (b *BPFunction) Destroy() {
	for _, iface := range b.Used {
		iface.Node.Instance.DeleteNode(iface)
	}
}

type QNodeInput struct {
	*EmbedNode
}

func (n *QNodeInput) imported(data any) {
	input := n.Iface.QFuncMain.Node.QFuncInstance.Input

	for key, value := range input {
		n.Node.CreatePort("output", key, value)
	}
}

func (n *QNodeInput) request(cable *Cable) {
	name := cable.Output.Name

	// This will trigger the port to request from outside and assign to this node's port
	n.Node.Output[name].Set(n.Iface.QFuncMain.Node.Input[name].Get())
}

type QNodeOutput struct {
	*EmbedNode
}

func (n *QNodeOutput) imported(data map[string]any) {
	output := n.Iface.QFuncMain.Node.QFuncInstance.Output

	for key, value := range output {
		n.Node.CreatePort("input", key, value)
	}
}

func (n *QNodeOutput) update(cable *Cable) {
	iface := n.Iface.QFuncMain
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
	QImportOnce bool
	QSave       func(any, string, bool)
	QPortSw_    map[string]int
}

func (f *FnMain) QBpFnInit() {
	if f.QImportOnce {
		panic("Can't import function more than once")
	}

	f.QImportOnce = true
	node := f.Node

	f.Iface.QBpInstance = New()
	bpFunction := node.QFuncInstance

	newInstance := f.Iface.QBpInstance
	// newInstance.Variables = []; // private for one function
	newInstance.SharedVariables = bpFunction.Variables // shared between function
	newInstance.Functions = node.Instance.Functions
	newInstance.QFuncMain = f.Iface
	newInstance.QMainInstance = bpFunction.RootInstance

	bpFunction.RefreshPrivateVars(newInstance)

	// swallowCopy := make([]any, len(bpFunction.Structure))
	// copy(swallowCopy, bpFunction.Structure)

	f.Iface.QBpInstance.ImportJSONParsed(bpFunction.Structure)

	// Init port switches
	if f.QPortSw_ != nil {
		f.Iface.QInitPortSwitches(f.QPortSw_)
		f.QPortSw_ = nil

		InputIface := f.Iface.QProxyInput.Iface
		InputIface_ := InputIface.Embed.(*QBpFnInOut)

		if InputIface_.QPortSw_ != nil {
			InputIface.QInitPortSwitches(InputIface_.QPortSw_)
			InputIface_.QPortSw_ = nil
		}
	}

	f.QSave = func(ev any, eventName string, force bool) {
		if force || bpFunction.QSyncing {
			return
		}

		// ev.BpFunction = bpFunction
		newInstance.QMainInstance.Emit(eventName, ev)

		bpFunction.QSyncing = true
		bpFunction.QOnFuncChanges(eventName, ev, f.Node)
		bpFunction.QSyncing = false
	}

	f.Iface.QBpInstance.On("cable.connect cable.disconnect node.created node.delete node.id.changed", f.QSave)
}
func (f *FnMain) RenamePort(which string, fromName string, toName string) {
	f.Node.QFuncInstance.RenamePort(which, fromName, toName)
	f.QSave(false, "", true)
}

type QBpFnInOut struct {
	*EmbedInterface
	Type     string
	QPortSw_ map[string]int
}

type addPortRef struct {
	Node *Node
	Port *Port
}

func (b *QBpFnInOut) AddPort(port *Port, customName string) *Port {
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
		portType = Ports.Trigger(func(*Port) {
			reff.Node.Output[reff.Port.Name].Call()
		})
	} else {
		if port.Feature != 0 {
			portType = port.QGetPortFeature()
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

		nodeA = b.Iface.QFuncMain.Node
		nodeB = b.Node
		nodeA.QFuncInstance.Input[name] = portType
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
		nodeB = b.Iface.QFuncMain.Node
		nodeB.QFuncInstance.Output[name] = portType
	}

	outputPort := nodeB.CreatePort("output", name, portType)

	var inputPort *Port
	if portType == types.Function {
		inputPort = nodeA.CreatePort("input", name, Ports.Trigger(outputPort.QCallAll))
	} else {
		inputPort = nodeA.CreatePort("input", name, portType)
	}

	if reff != nil {
		reff.Node = nodeB
		reff.Port = inputPort
	}

	if b.Type == "bp-fn-input" {
		outputPort.Name_ = &RefPortName{Name: name} // When renaming port, this also need to be changed
		b.Iface.Emit("_add.{name}", outputPort)

		inputPort.On("value", func(ev PortValueEvent) {
			outputPort.Iface.Node.Output[outputPort.Name].Set(ev.Cable.Output.Value)
		})

		return outputPort
	}

	inputPort.Name_ = &RefPortName{Name: name} // When renaming port, this also need to be changed
	b.Iface.Emit("_add.{name}", inputPort)
	return inputPort
}

func (b *QBpFnInOut) renamePort(fromName string, toName string) {
	bpFunction := b.Iface.QFuncMain.Node.QFuncInstance
	// Main (input) -> Input (output)
	if b.Type == "bp-fn-input" {
		bpFunction.RenamePort("input", fromName, toName)
	} else { // Output (input) -> Main (output)
		bpFunction.RenamePort("output", fromName, toName)
	}
}

func (b *QBpFnInOut) deletePort(name string) {
	funcMainNode := b.Iface.QFuncMain.Node
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

func init() {
	QNodeList["BP/Fn/Input"] = &NodeRegister{
		Output: NodePortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &QNodeInput{}

			iface := node.SetInterface("BPIC/BP/Fn/Input")
			iface.QEnum = nodes.BPFnInput
			iface.QDynamicPort = true // Port is initialized dynamically

			iface.Title = "Input"
			iface.Embed.(*QBpFnInOut).Type = "bp-fn-input"
			iface.QFuncMain = node.Instance.QFuncMain
			iface.QFuncMain.QProxyInput = node
		},
	}

	QInterfaceList["BPIC/BP/Fn/Input"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &QBpFnInOut{}
		},
	}

	QNodeList["BP/Fn/Output"] = &NodeRegister{
		Input: NodePortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &bpVarGet{}

			iface := node.SetInterface("BPIC/BP/Fn/Output")
			iface.QEnum = nodes.BPFnOutput
			iface.QDynamicPort = true // Port is initialized dynamically

			iface.Title = "Output"
			iface.Embed.(*QBpFnInOut).Type = "bp-fn-output"
			iface.QFuncMain = node.Instance.QFuncMain
			iface.QFuncMain.QProxyOutput = node
		},
	}

	QInterfaceList["BPIC/BP/Fn/Output"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &QBpFnInOut{}
		},
	}

	QInterfaceList["BPIC/BP/Fn/Main"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &FnMain{}
		},
	}
}
