package engine

import (
	"encoding/json"
	"fmt"

	"github.com/blackprint/engine-go/utils"
)

type NodePort map[string]interface{}

type Instance struct {
	Iface     map[string]interface{} // Storing with node id if exist
	IfaceList map[int]interface{}    // Storing with node index
	settings  map[string]bool
}

func New() *Instance {
	return &Instance{
		Iface:     map[string]interface{}{},
		IfaceList: map[int]interface{}{},
		settings:  map[string]bool{},
	}
}

//
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

type SingleInstanceJSON map[string]nodeList
type metaValue map[string]string
type nodeList []nodeConfig
type nodeConfig struct {
	I      int                         `json:"i"`
	Id     string                      `json:"id"`
	Data   interface{}                 `json:"data"`
	Output map[string][]nodePortTarget `json:"output"`
}
type nodePortTarget struct {
	I    int    `json:"i"`
	Name string `json:"name"`
}

func (instance *Instance) ImportJSON(str []byte) {
	var data SingleInstanceJSON

	err := json.Unmarshal(str, &data)
	if err != nil {
		fmt.Println(err)
	}

	ifaceList := instance.IfaceList
	var nodes []interface{}

	// Prepare all ifaces based on the namespace
	// before we create cables for them
	for namespace, ifaces := range data {
		if namespace == "_" {
			// meta := ifaces.(metaValue)
			continue
		}

		list := ifaces //.(nodeList)

		// Every ifaces that using this namespace name
		for _, iface := range list {
			ifaceList[iface.I], nodes = instance.CreateNode(namespace, iface, nodes)
		}
	}

	// Create cable only from output and property
	// > Important to be separated from above, so the cable can reference to loaded ifaces
	for _, ifaces := range data {
		list := ifaces //.(nodeList)

		for _, iface := range list {
			current := ifaceList[iface.I]

			// If have output connection
			out := iface.Output
			if out != nil {
				Output := *utils.GetPropertyRef(current, "Output").(*map[string]*Port)

				// Every output port that have connection
				for portName, ports := range out {
					linkPortA := Output[portName]

					if linkPortA == nil {
						panic(fmt.Sprintf("Node port not found for iface (index: %d), with name: %s", iface.I, portName))
					}

					// Current output's available targets
					for _, target := range ports {
						targetNode := ifaceList[target.I]

						Input := *utils.GetPropertyRef(targetNode, "Input").(*map[string]*Port)
						linkPortB := Input[target.Name]

						if linkPortB == nil {
							targetTitle := utils.GetProperty(targetNode, "Title").(string)
							panic(fmt.Sprintf("Node port not found for %s with name: %s", targetTitle, target.Name))
						}

						// For Debugging ->
						// Title := utils.GetProperty(current, "Title").(string)
						// targetTitle := utils.GetProperty(targetNode, "Title").(string)
						// fmt.Printf("%s.%s => %s.%s\n", Title, linkPortA.Name, targetTitle, linkPortB.Name)
						// <- For Debugging

						cable := NewCable(linkPortA, linkPortB)
						linkPortA.Cables = append(linkPortA.Cables, &cable)
						linkPortB.Cables = append(linkPortB.Cables, &cable)

						cable.QConnected()
						// fmt.Println(cable.String())
					}
				}
			}
		}
	}

	// Call nodes init after creation processes was finished
	for _, val := range nodes {
		utils.CallFunction(val, "Init", utils.EmptyArgs)
	}
}

func (instance *Instance) Settings(id string, val ...bool) bool {
	if val == nil {
		return instance.settings[id]
	}

	temp := val[0]
	instance.settings[id] = temp
	return temp
}

func (instance *Instance) GetNode(id interface{}) interface{} {
	for _, val := range instance.IfaceList {
		iface := val.(Interface)
		if iface.Id == id || iface.I == id {
			return iface.Node
		}
	}
	return nil
}

func (instance *Instance) GetNodes(namespace string) []interface{} {
	var got []interface{} // interface = extends 'Node'

	for _, val := range instance.IfaceList {
		iface := val.(Interface)
		if iface.Namespace == namespace {
			got = append(got, val)
		}
	}

	return got
}

func (instance *Instance) CreateNode(namespace string, options nodeConfig, nodes []interface{}) (interface{}, []interface{}) {
	func_ := QNodeList[namespace]
	if func_ == nil {
		panic("Node nodes for " + namespace + " was not found, maybe .registerNode() haven't being called?")
	}

	// *node: extends engine.Node
	node := func_(instance) // func_ from registerNode(namespace, func_)
	if utils.IsPointer(node) == false {
		panic(namespace + ": .registerNode() must return pointer")
	}

	// *iface: extends engine.Interface
	iface := utils.GetProperty(node, "Iface")
	if iface == nil || utils.GetProperty(iface, "QInitialized").(bool) == false {
		panic(namespace + ": Node interface was not found, do you forget to call node->setInterface() ?")
	}

	utils.SetProperty(iface, "Node", node)

	// Assign the saved options if exist
	// Must be called here to avoid port trigger
	if options.Data != nil {
		data := utils.GetPropertyRef(iface, "Data").(*InterfaceData)
		if data != nil {
			deepMerge(data, options.Data.(map[string]interface{}))
		}
	}

	utils.SetProperty(iface, "Namespace", namespace)

	// Create the linker between the nodes and the iface
	utils.CallFunction(iface, "QPrepare", utils.EmptyArgs)

	if options.Id != "" {
		utils.SetProperty(iface, "Id", options.Id)
		instance.Iface[options.Id] = iface
	}

	utils.SetProperty(iface, "I", options.I)
	instance.IfaceList[options.I] = iface

	utils.SetProperty(iface, "Importing", false)
	utils.CallFunction(node, "Imported", utils.EmptyArgs)

	if nodes != nil {
		nodes = append(nodes, node)
	}

	utils.CallFunction(node, "Init", utils.EmptyArgs)
	utils.CallFunction(iface, "Init", utils.EmptyArgs)

	return iface, nodes
}

// Currently only one level
func deepMerge(real_ *InterfaceData, merge map[string]interface{}) {
	real := *real_
	for key, val := range merge {
		real[key](val)
	}
}
