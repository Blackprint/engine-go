package engine

import (
	"github.com/buger/jsonparser"
	"fmt"
	"log"
	"reflect"
)

type Instance struct{}
type NodePort map[string]reflect.Kind

type engine struct {
	Iface     map[string]interface{} // Storing with node id if exist
	IfaceList map[int]interface{}    // Storing with node index
	settings  map[string]bool
}

func New() engine {
	return engine{}
}

type SingleInstanceJSON map[string]interface{}
type metaValue map[string]string
type nodeList map[string][]map[string]nodeConfig
type nodeConfig struct {
	I      int                         `json:"i"`
	Id     int                         `json:"id"`
	Data   int                         `json:"data"`
	Output map[string][]nodePortTarget `json:"output"`
}
type nodePortTarget struct {
	I    int    `json:"i"`
	Name string `json:"name"`
}

func (instance *engine) ImportJSON(str []byte) {
	var data SingleInstanceJSON

	err := json.Unmarshal(str, &data)
	if err != nil {
		fmt.Println(err)
	}

	IfaceList := instance.IfaceList
	var nodes []interface{}

	// Prepare all ifaces based on the namespace
	// before we create cables for them
	for namespace, ifaces := range data {
		// Every ifaces that using this namespace name
		for _, iface := range ifaces {
			i := iface["i"]
			IfaceList[i] = instance.CreateNode(namespace, map[string]interface{}{
				"id":   iface["id"],
				"i":    i,
				"data": iface["data"],
			}, nodes)
		}
	}

	// Create cable only from output and property
	// > Important to be separated from above, so the cable can reference to loaded ifaces
	for namespace, ifaces := range data {
		for _, iface := range ifaces {
			current := ifaceList[iface["i"]]

			// If have output connection
			out := iface["output"]
			if out != nil {
				// Every output port that have connection
				for portName, ports := range out {
					linkPortA := current.Output[portName]
					if linkPortA == nil {
						panic(fmt.Sprintf("Node port not found for iface (index: %d), with name: %s", iface["i"], portName))
					}

					// Current output's available targets
					for _, target := range ports {
						targetNode := ifaceList[target["i"]]
						linkPortB := current.Input[target["name"]]

						if linkPortB == nil {
							panic(fmt.Sprintf("Node port not found for %s with name: %s", targetNode.Title, target["name"]))
						}

						log.Printf("%s.%s => %s.%s", current.Title, linkPortA.Name, targetNode.Title, linkPortB.Name)

						cable := NewCable(linkPortA, linkPortB)
						append(linkPortA.Cables, cable)
						append(linkPortB.Cables, cable)

						cable.QConnected()
						cable.QPrint()
					}
				}
			}
		}
	}

	// Call nodes init after creation processes was finished
	for _, val := range nodes {
		val.Init()
	}
}

func (instance *engine) Settings(id string, val bool) {
	if val == nil {
		return instance.settings[id]
	}

	instance.settings[id] = val
}

func (instance *engine) GetNode(id string) {
	for _, val := range instance.IfaceList {
		if val.Id == id || val.I == id {
			return val.Node
		}
	}
}

func (instance *engine) GetNodes(namespace string) []interface{} {
	var got []interface{} // interface = extends 'Node'

	for _, val := range instance.IfaceList {
		if val.Namespace == namespace {
			got = append(got, val)
		}
	}

	return got
}

func (instance *engine) CreateNode(namespace string, options string, nodes *map[string]interface{}) interface{} {
	func_ := QNodeList[namespace]
	if func_ == nil {
		panic("Node nodes for " + namespace + " was not found, maybe .registerNode() haven't being called?")
	}

	node := func_(instance) // from registerNode(namespace, func_)
	iface := node.Iface

	if iface == nil {
		panic("Node interface was not found, do you forget to call node->setInterface() ?")
	}

	// Assign the saved options if exist
	// Must be called here to avoid port trigger
	if options.data != nil {
		if iface.Data != nil {
			deepMerge(iface.Data, options.data)
		} else {
			iface.Data = options.data
		}
	}

	// Create the linker between the nodes and the iface
	iface.QPrepare()

	iface.Namespace = namespace
	if options.id != nil {
		iface.Id = options.id
		instance.Iface[options.id] = iface
	}

	if options.i != nil {
		iface.I = options.i
		instance.IfaceList[options.i] = iface
	} else {
		instance.IfaceList = append(instance.IfaceList[options.i], iface)
	}

	iface.Importing = false
	node.Imported()

	if nodes != nil {
		nodes = append(nodes, node)
	}

	node.Init()
	iface.Init()

	return iface
}

type objKeyVal map[string]interface{}

func deepMerge(real objKeyVal, merge objKeyVal) {
	for key, val := range merge {
		if reflect.KindOf(val) == "map" {
			deepMerge(real[key], val)
			continue
		}

		real[key](val)
	}
}
