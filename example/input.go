package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// This will be called from example.go
func RegisterInput() {
	RegisterInputSimple()
}

// ============
type InputSimple struct {
	engine.Node
}

// Bring value from imported iface to node output
func (node InputSimple) Imported() {
	val := node.Iface.(engine.Interface).Data["value"].(engine.GetterSetter)()
	if val != nil {
		log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mSaved data as output: %s\x1b[0m", val)
	}

	node.Output["Value"](val)
}

type InputSimpleIFace struct {
	engine.Interface
}

func (iface InputSimpleIFace) Changed(val interface{}) {
	// This node still being imported
	if iface.Importing != false {
		return
	}

	log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mThe input box have new value: %s\x1b[0m", val)

	node := iface.Node.(engine.Node)
	node.Output["Value"](val)

	// This will call every connected node
	node.Output["Changed"]()
}

func RegisterInputSimple() {
	Blackprint.RegisterNode("Example/Input/Simple", func(instance *engine.Instance) interface{} {
		node := InputSimple{
			Node: engine.Node{
				Instance: instance,

				// Node's Output Port Template
				TOutput: engine.NodePort{
					"Changed": types.Function,
					"Value":   types.String,
				},
			},
		}

		iface := node.SetInterface("BPIC/Example/Input").(*InputSimpleIFace)
		iface.Title = "Input"

		return &node
	})

	Blackprint.RegisterInterface("BPIC/Example/Input", func(node_ interface{}) interface{} {
		// node := node_.(InputSimple)
		value := "..."

		var iface InputSimpleIFace
		iface = InputSimpleIFace{
			Interface: engine.Interface{
				Data: engine.InterfaceData{
					"value": func(val ...interface{}) interface{} {
						if len(val) == 0 {
							return value
						}

						value = val[0].(string)
						iface.Changed(val)
						return nil
					},
				},
			},
		}

		return &iface
	})
}
