package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// ============
type InputSimple struct {
	*engine.Node
}

// Bring value from imported iface to node output
func (node *InputSimple) Imported() {
	val := node.IFace.(*InputSimpleIFace).Data["value"].Get()
	if val != nil {
		log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mSaved data as output: %s\x1b[0m\n", val)
	}

	node.Output["Value"].Set(val)
}

type InputSimpleIFace struct {
	*engine.Interface
}

func (iface *InputSimpleIFace) Changed(val any) {
	// This node still being imported
	if iface.Importing != false {
		return
	}

	log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mThe input box have new value: %s\x1b[0m\n", val)

	node := iface.Node.(*InputSimple)
	node.Output["Value"].Set(val)

	// This will call every connected node
	node.Output["Changed"].Call()
}

type MyData struct {
	IFace any
	val   any
}

func (gs *MyData) Set(val any) {
	gs.val = val
	gs.IFace.(*InputSimpleIFace).Changed(gs.val)
}

func (gs *MyData) Get() any {
	return gs.val
}

// This will be called from example.go
func RegisterInput() {
	Blackprint.RegisterNode("Example/Input/Simple", func(instance *engine.Instance) any {
		node := &InputSimple{
			Node: &engine.Node{
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

		return node
	})

	Blackprint.RegisterInterface("BPIC/Example/Input", func(node_ any) any {
		// node := node_.(InputSimple)

		var iface *InputSimpleIFace
		iface = &InputSimpleIFace{
			Interface: &engine.Interface{
				Data: engine.InterfaceData{
					"value": &MyData{val: "..."},
				},
			},
		}

		return iface
	})
}
