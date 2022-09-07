package engine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

var Event = &CustomEvent{}

type NodePortTemplate map[string]any
type Instance struct {
	*CustomEvent
	Iface        map[string]any // Storing with node id if exist
	IfaceList    []any          // Storing with node index
	settings     map[string]bool
	DisablePorts bool
	PanicOnError bool

	Variables map[string]*BPVariable
	Functions map[string]*BPFunction
	Ref       map[string]*ReferencesShortcut

	QFuncMain // => *engine.Interface
}

func New() *Instance {
	return &Instance{
		Iface:        map[string]any{},
		IfaceList:    map[int]any{},
		settings:     map[string]bool{},
		PanicOnError: true,
	}
}

type Data struct {
	Value string `json:"value"`
}
type Namespace string
type NodeData struct {
	Data Data `json:"data,omitempty"`
}
type NodeOutput struct {
	Output []Node `json:"output"`
}
type NodeX struct {
	Name string  `json:"name"`
	I    *int64  `json:"i,omitempty"`
	ID   *string `json:"id,omitempty"`

	NodeData
	NodeOutput
}
type DataStructure map[Namespace][]NodeX

//

type SingleInstanceJSON map[string]any // any = nodeList | metadataValue
type metadataValue map[string]any
type nodeList []nodeConfig
type nodeConfig struct {
	I            int                         `json:"i"`
	Id           string                      `json:"id"`
	Data         any                         `json:"data"`
	Output       map[string][]nodePortTarget `json:"output"`
	InputDefault map[string]any              `json:"input_d"`
	OutputSwitch map[string]uint             `json:"output_sw"`
	Route        map[string]uint             `json:"route"`
}
type nodePortTarget struct {
	I    int    `json:"i"`
	Name string `json:"name"`
}

type getterSetter interface {
	Set(val any)
	Get() any
}

type GetterSetter struct {
	Value any
}

func (b *GetterSetter) Get() any {
	return b.Value
}

func (b *GetterSetter) Set(Value any) {
	b.Value = Value
}

type ImportOptions struct {
	AppendMode bool
}

func (instance *Instance) ImportJSON(str []byte, options ...ImportOptions) (inserted []any, err error) {
	var data SingleInstanceJSON

	err = json.Unmarshal(str, &data)
	if err != nil {
		return
	}

	hasOption := len(options) != 0
	options_ := options[0]

	if hasOption && options_.AppendMode == false {
		instance.ClearNodes()
	}

	// Do we need this?
	// instance.Emit("json.importing", {appendMode: options.appendMode, raw: json});

	ifaceList := instance.IfaceList
	var metadata metadataValue

	appendLength := 0
	if options_.AppendMode {
		appendLength = len(ifaceList)
	}

	var exist bool
	if metadata, exist = data["_"].(metadataValue); exist {
		if list, exist := metadata["env"]; exist {
			QEnvironment.Import(list.(map[string]string))
		}

		if list, exist := metadata["functions"].(map[string]any); exist {
			for key, options := range list {
				instance.CreateFunction(key, options)
			}
		}

		if list, exist := metadata["variables"].(map[string]any); exist {
			for key, options := range list {
				instance.CreateVariable(key, options)
			}
		}

		delete(data, "_")
	}

	// Prepare all ifaces based on the namespace
	// before we create cables for them
	for namespace, ifaces := range data {
		// Every ifaces that using this namespace name
		for _, iface := range ifaces.(nodeList) {
			iface.I += appendLength

			var temp any
			temp, inserted = instance.CreateNode(namespace, iface, inserted)

			ifaceList[iface.I] = temp
			utils.CallFunction(temp, "QBpFnInit", utils.EmptyArgs)
		}
	}

	// Create cable only from output and property
	// > Important to be separated from above, so the cable can reference to loaded ifaces
	for _, ifaces := range data {
		list := ifaces.(nodeList)

		for _, ifaceJSON := range list {
			iface := ifaceList[ifaceJSON.I]

			if val := ifaceJSON.Route; val != nil {
				iface.Node.Routes.RouteTo(ifaceList[val["i"]])
			}

			// If have output connection
			out := ifaceJSON.Output
			if out != nil {
				Output := *utils.GetPropertyRef(iface, "Output").(*map[string]*Port)

				// Every output port that have connection
				for portName, ports := range out {
					linkPortA := Output[portName]

					if linkPortA == nil {
						if iface.QEnum == nodes.BPFnInput {
							target := instance.QGetTargetPortType(iface.Node.Instance, "input", ports)
							linkPortA = iface.AddPort(target, portName)

							if linkPortA == nil {
								panic(fmt.Sprintf("Can't create output port (%s) for function (%s)", portName, iface.QFuncMain.Node.QFuncInstance.Id))
							}
						} else if iface.QEnum == nodes.BPVarGet {
							target := instance.QGetTargetPortType(instance, "input", ports)
							iface.UseType(target)
							linkPortA = iface.Output[portName]
						} else {
							panic(fmt.Sprintf("Node port not found for iface (index: %d), with name: %s", ifaceJSON.I, portName))
						}
					}

					// Current output's available targets
					for _, target := range ports {
						target.I += appendLength
						targetNode := ifaceList[target.I]

						// output can only meet input port
						Input := *utils.GetPropertyRef(targetNode, "Input").(*map[string]*Port)
						linkPortB := Input[target.Name]

						if linkPortB == nil {
							targetTitle := utils.GetProperty(targetNode, "Title").(string)

							if targetNode.QEnum == nodes.BPFnOutput {
								linkPortB = targetNode.AddPort(linkPortA, target)

								if linkPortB == nil {
									panic(fmt.Sprintf("Can't create output port (%s) for function (%s)", portName, targetNode.QFuncMain.Node.QFuncInstance.Id))
								}
							} else if targetNode.QEnum == nodes.BPVarGet {
								targetNode.UseType(target)
								linkPortB = targetNode.Input[target.Name]
							} else if targetNode.Type == types.Route {
								linkPortB = targetNode.Node.Routes
							} else {
								panic(fmt.Sprintf("Node port not found for %s with name: %s", targetTitle, target.Name))
							}
						}

						// For Debugging ->
						// Title := utils.GetProperty(iface, "Title").(string)
						// targetTitle := utils.GetProperty(targetNode, "Title").(string)
						// fmt.Printf("%s.%s => %s.%s\n", Title, linkPortA.Name, targetTitle, linkPortB.Name)
						// <- For Debugging

						linkPortA.ConnectPort(linkPortB)
						// fmt.Println(cable.String())
					}
				}
			}
		}
	}

	// Call nodes init after creation processes was finished
	for _, val := range inserted {
		utils.CallFunction(val, "Init", utils.EmptyArgs)
	}

	return
}

func (instance *Instance) QGetTargetPortType(ins *Instance, which string, targetNodes []nodePortTarget) *Port {
	target := targetNodes[0] // ToDo: check all target in case if it's supporting Union type
	targetIface := ins.IfaceList[target.I]

	if which == "input" {
		return targetIface.Input[target.Name]
	} else {
		return targetIface.Output[target.Name]
	}
}

type EvNodeDelete struct {
	Iface any
}

func (instance *Instance) DeleteNode(iface any) {
	i := utils.IndexOfAny(instance.IfaceList, iface)
	if i == -1 {
		panic("Node to be deleted was not found")
	}

	instance.IfaceList = utils.RemoveItemAtIndexAny(instance.IfaceList, i)

	eventData := &EvNodeDelete{
		Iface: iface,
	}
	instance.Emit("node.delete", eventData)

	iface.Node.Destroy()
	iface.Destroy()

	for _, port := range iface.Output {
		port.DisconnectAll(instance.QRemote == nil)
	}

	routes := iface.Node.Routes
	for _, cable := range routes.In {
		cable.Disconnect()
	}

	if routes.Out != nil {
		routes.Out.Disconnect()
	}

	// Delete reference
	delete(instance.Iface, iface.Id)
	delete(instance.Ref, iface.Id)

	parent := iface.Node.Instance.QFuncMain
	if parent != nil {
		delete(parent.Ref, iface.Id)
	}

	instance.Emit("node.deleted", eventData)
}

func (instance *Instance) ClearNodes() {
	for _, iface := range instance.IfaceList {
		iface.Node.Destroy()
		iface.Destroy()
	}

	instance.IfaceList = instance.IfaceList[:0]
	utils.ClearMap(instance.Iface)
	utils.ClearMap(instance.Ref)
}

func (instance *Instance) Settings(id string, val ...bool) bool {
	if val == nil {
		return instance.settings[id]
	}

	temp := val[0]
	instance.settings[id] = temp
	return temp
}

func (instance *Instance) GetNode(id any) any {
	for _, val := range instance.IfaceList {
		temp := reflect.ValueOf(val).Elem()
		if temp.FieldByName("Id").Interface().(string) == id || temp.FieldByName("I").Interface().(int) == id {
			return utils.GetProperty(val, "Node")
		}
	}
	return nil
}

func (instance *Instance) GetNodes(namespace string) []any {
	var got []any // any = extends 'engine.Node'

	for _, val := range instance.IfaceList {
		if utils.GetProperty(val, "Namespace").(string) == namespace {
			got = append(got, utils.GetProperty(val, "Node"))
		}
	}

	return got
}

// ToDo: sync with JS, when creating function node this still broken
func (instance *Instance) CreateNode(namespace string, options nodeConfig, nodes []any) (any, []any) {
	func_ := QNodeList[namespace]
	var node any // *node: extends engine.Node
	var isFuncNode bool
	if func_ == nil {
		if strings.HasPrefix(namespace, "BPI/F/") {
			temp := instance.Functions[namespace]
			if temp != nil {
				node = temp.QBuilder()
			}

			isFuncNode = true
		}

		if node == nil {
			panic("Node nodes for " + namespace + " was not found, maybe .registerNode() haven't being called?")
		}
	} else {
		node = func_(instance) // func_ from registerNode(namespace, func_)
	}

	// Disable data flow on any node ports
	if instance.DisablePorts {
		node.DisablePorts = true
	}

	if utils.IsPointer(node) == false {
		panic(namespace + ": .registerNode() must return pointer")
	}

	// *iface: extends engine.Interface
	iface := utils.GetProperty(node, "Iface")
	if iface == nil || utils.GetProperty(iface, "QInitialized").(bool) == false {
		panic(namespace + ": Node interface was not found, do you forget to call node->setInterface() ?")
	}

	utils.SetProperty(iface, "Node", node)
	utils.SetProperty(iface, "Namespace", namespace)

	// Create the linker between the nodes and the iface
	if isFuncNode == false {
		utils.CallFunction(iface, "QPrepare", utils.EmptyArgs)
	}

	if options.Id != "" {
		utils.SetProperty(iface, "Id", options.Id)
		instance.Iface[options.Id] = iface
		instance.Ref[options.Id] = iface.Ref

		parent := iface.Node.Instance.QFuncMain
		if parent != nil {
			parent.Ref[options.Id] = iface.Ref
		}
	}

	utils.SetProperty(iface, "I", options.I)
	instance.IfaceList[options.I] = iface

	if options.InputDefault != nil {
		iface.QImportInputs(options.InputDefault)
	}

	savedData := &[]reflect.Value{reflect.ValueOf(options.Data)}

	if options.OutputSwitch != nil {
		for key, val := range options.OutputSwitch {
			if (val | 1) == 1 {
				portStructOf_split(iface.Output[key])
			}

			if (val | 2) == 2 {
				iface.Output[key].AllowResync = true
			}
		}
	}

	utils.SetProperty(iface, "Importing", false)

	utils.CallFunction(iface, "Imported", savedData)
	utils.CallFunction(node, "Imported", savedData)

	utils.CallFunction(iface, "Init", utils.EmptyArgs)
	if nodes != nil {
		nodes = append(nodes, node)
	} else {
		// Init now if not A batch creation
		utils.CallFunction(node, "Init", utils.EmptyArgs)
	}

	return iface, nodes
}

var createBPVariableRegx = regexp.MustCompile(`[` + "`" + `~!@#$%^&*()\-_+={}\[\]:"|;\'\\\\,.\/<>?]+`)

type varOptions struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

func (instance *Instance) CreateVariable(id string, options any) *BPVariable {
	if old, exist := instance.Variables[id]; exist {
		old.Destroy()
		delete(instance.Variables, id)
	}

	// options_ = options.(varOptions)

	temp := &BPVariable{
		Id:    id,
		Title: id,
		Type:  0, // Type not set
	}
	instance.Variables[id] = temp
	instance.Emit("variable.new", temp)

	return temp
}

type funcOptions struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Vars        []string `json:"vars"`
	PrivateVars []string `json:"privateVars"`
	Structure   nodeList `json:"structure"`
}

func (instance *Instance) CreateFunction(id string, options any) *BPFunction {
	if old, exist := instance.Functions[id]; exist {
		old.Destroy()
		delete(instance.Functions, id)
	}

	temp := &BPFunction{
		Id:    id,
		Title: id,
		Type:  0, // Type not set
	}
	instance.Functions[id] = temp

	options_ := options.(funcOptions)

	for _, val := range options_.Vars {
		temp.CreateVariable(val, bpFnVarOptions{
			Scope: val,
		})
	}

	for _, val := range options_.PrivateVars {
		temp.AddPrivateVars(val)
	}

	instance.Emit("function.new", temp)
	return temp
}

type NodeLogEvent struct {
	Instance   *Instance
	Iface      any
	IfaceTitle string
	Message    string
}

func (instance *Instance) QLog(iface any, message string) {
	evData := NodeLogEvent{
		Instance:   instance,
		Iface:      iface,
		IfaceTitle: utils.GetProperty(iface, "Title").(string),
		Message:    message,
	}

	if instance.QMainInstance != nil {
		instance.QMainInstance.Emit("log", evData)
	} else {
		instance.Emit("log", evData)
	}
}

func (instance *Instance) Destroy() {
	instance.ClearNodes()
}

// Currently only one level
func deepMerge(real_ *InterfaceData, merge map[string]any) {
	real := *real_
	for key, val := range merge {
		real[key].Set(val)
	}
}