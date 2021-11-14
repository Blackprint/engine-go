package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// This will be called from example.go
func RegisterButton() {
	RegisterButtonSimple()
}

// ============
type ButtonSimple struct {
	*engine.Node
}
type ButtonSimpleIFace struct {
	*engine.IFace
}

func (iface *ButtonSimpleIFace) Clicked(ev interface{}) {
	log.Printf("\x1b[1m\x1b[33mButton\\Simple:\x1b[0m \x1b[33mI got '%d', time to trigger to the other node\x1b[0m\n", ev)

	iface.Node.(*ButtonSimple).Output["Clicked"](ev)
}

func RegisterButtonSimple() {
	Blackprint.RegisterNode("Example/Button/Simple", func(instance *engine.Instance) interface{} {
		node := ButtonSimple{
			Node: &engine.Node{
				Instance: instance,

				// Node's Output Port Template
				TOutput: engine.NodePort{
					"Clicked": types.Function,
				},
			},
		}

		iface := node.SetInterface("BPIC/Example/Button").(*ButtonSimpleIFace)
		iface.Title = "Button"

		return &node
	})

	Blackprint.RegisterInterface("BPIC/Example/Button", func(node interface{}) interface{} {
		// node_ := node.(ButtonSimple)
		return &ButtonSimpleIFace{
			IFace: &engine.IFace{},
		}
	})
}
