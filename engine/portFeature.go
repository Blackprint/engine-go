package engine

import (
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

func portArrayOf_validate(source *Port, target *Port) bool {
	if source.Type == target.Type || source.Type == types.Any || target.Type == types.Any {
		return true
	}

	if source.Types != nil && utils.Contains(source.Types, target.Type) {
		return true
	}

	return false
}

func portUnion_validate(source *Port, target *Port) bool {
	if source.Types != nil && target.Types != nil {
		if len(source.Types) != len(target.Types) {
			return false
		}

		for _, type_ := range source.Types {
			if !utils.Contains(target.Types, type_) {
				return false
			}
		}

		return true
	}

	return target.Type == types.Any || utils.Contains(source.Types, target.Type)
}

func portStructOf_split(port *Port) {
	if port.Source == PortInput {
		panic("Port with feature 'StructOf' only supported for output port")
	}

	node := port.Iface.Node
	struct_ := &port.Struct

	for key, val := range *struct_ {
		name := port.Name + key
		newPort := node.CreatePort("output", name, val.Type)
		newPort.QParent = port
		newPort.QStructSplitted = true
	}

	port.Splitted = true
	port.DisconnectAll()

	portData := node.Output.(map[string]*PortOutputGetterSetter)[port.Name]
	if portData != nil {
		portStructOf_handle(port, portData)
	}
}

func portStructOf_unsplit(port *Port) {
	parent := port.QParent
	if parent == nil && port.Struct != nil {
		parent = port
	}

	parent.Splitted = false
	node := port.Iface.Node

	for key, _ := range parent.Struct {
		node.DeletePort("output", parent.Name+key)
	}
}

func portStructOf_handle(port *Port, data any) {
	output := port.Iface.Node.Output.(map[string]*PortOutputGetterSetter)

	if data != nil {
		for key, val := range port.Struct {
			if val.Field == "" {
				output[key].Set(utils.GetProperty(data, val.Field))
			} else {
				output[key].Set(val.Handle(data))
			}
		}
	} else {
		for key, _ := range port.Struct {
			output[key].Set(nil)
		}
	}
}
