package engine

import (
	"fmt"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

// Main function node
type BPFunctionNode struct { // Main function node -> BPI/F/{FunctionName}
	Node *Node
}

func (b *BPFunctionNode) Init() {
	if b.Node.Iface.QImportOnce {
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
		thisInput := b.Input

		// Sync all port value
		for key, value := range iface.Output {
			if value.Type == types.Function {
				continue
			}

			Output[key].Set(thisInput[key]())
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
	Used         *Interface
	Input        *NodePortTemplate
	Output       *NodePortTemplate
	Structure    SingleInstanceJSON
	Variables    map[string]BPVariable
	PrivateVars  []string
	RootInstance *Instance
	Node         func(*Instance) *Node // Node constructor
}

func (b *BPFunction) QOnFuncChanges(eventName string, obj any, fromNode *Node) {
	for _, iface_ := range b.Used {
		if iface_.Node == fromNode {
			continue
		}

		nodeInstance := iface_.QBpInstance
		nodeInstance.PendingRender = true // Force recalculation for cable position

		if eventName == "cable.connect" || eventName == "cable.disconnect" {
			input := obj.Cable.Input
			output := obj.Cable.Output
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
						targetInput = inputIface.AddPort(targetOutput, output.Name)
					} else {
						panic("Output port was not found")
					}
				}

				if targetOutput == nil {
					if outputIface.QEnum == nodes.BPFnInput {
						targetOutput = outputIface.AddPort(targetInput, input.Name)
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
			iface := obj.Iface
			nodeInstance.CreateNode(iface.Namespace, map[string]any{
				"data": iface.Data,
			})
		} else if eventName == "node.delete" {
			index := utils.IndexOf(fromNode.Iface.QBpInstance.IfaceList, obj.Iface)
			if index == false {
				panic("Failed to get node index")
			}

			iface := nodeInstance.IfaceList[index]
			if iface.Namespace != obj.Iface.Namespace {
				fmt.Println(iface.Namespace + " " + obj.Iface.Namespace)
				panic("Failed to delete node from other function instance")
			}

			if eventName == "node.delete" {
				nodeInstance.DeleteNode(iface)
			}
		}
	}
}

func (b *BPFunction) CreateNode(instance *Instance, options map[string]any) (any, []any) {
	return instance.CreateNode(b.Node, options)
}

func (b *BPFunction) CreateVariable(id string, options map[string]any) *BPVariable {
	if _, exist := b.Variables[id]; exist {
		panic("Variable id already exist: id")
	}

	// deepProperty

	temp := &BPVariable{
		Id: id,
		// options,
	}
	temp.FuncInstance = b

	if options["scope"] == VarScopeShared {
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
		main = *b.Output
		proxyPort = "input"
	} else {
		main = *b.Input
		proxyPort = "output"
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

		temp.Iface[proxyPort][fromName].Name_.Name = toName
		temp.RenamePort(proxyPort, fromName, toName)

		for _, proxyVar := range iface.BpInstance.IfaceList {
			if (which == "output" && proxyVar.Namespace != "BP/FnVar/Output") || (which == "input" && proxyVar.Namespace != "BP/FnVar/Input") {
				continue
			}

			if proxyVar.Data.Name != fromName {
				continue
			}
			proxyVar.Data.Name = toName

			if which == "output" {
				proxyVar[proxyPort]["Val"].Name_.Name = toName
			}
		}
	}
}

func (b *BPFunction) Destroy() {
	for _, iface := range b.Used {
		iface.Node.Instance.DeleteNode(iface)
	}
}
