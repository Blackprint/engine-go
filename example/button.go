package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// ============
type ButtonSimple struct {
	*engine.Node
}
type ButtonSimpleIFace struct {
	*engine.Interface
}

func (iface *ButtonSimpleIFace) Clicked(ev any) {
	log.Printf("\x1b[1m\x1b[33mButton\\Simple:\x1b[0m \x1b[33mI got '%d', time to trigger to the other node\x1b[0m\n", ev)

	iface.Node.(*ButtonSimple).Output["Clicked"].Set(ev)
}

// This will be called from example.go
func RegisterButton() {
	ButtonSimpleOutput := &engine.NodePortTemplate{
		"Clicked": types.Function,
	}

	Blackprint.RegisterNode("Example/Button/Simple", func(instance *engine.Instance) any {
		node := &ButtonSimple{
			Node: &engine.Node{
				Instance: instance,

				// Node's Output Port Template
				TOutput: ButtonSimpleOutput,
			},
		}

		iface := node.SetInterface("BPIC/Example/Button").(*ButtonSimpleIFace)
		iface.Title = "Button"

		return node
	})

	Blackprint.RegisterInterface("BPIC/Example/Button", func(node any) any {
		// node_ := node.(ButtonSimple)
		return &ButtonSimpleIFace{
			Interface: &engine.Interface{},
		}
	})
}
