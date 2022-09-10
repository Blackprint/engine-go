package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
)

// ============
type ButtonSimple struct {
	*engine.EmbedNode
}
type ButtonSimpleIFace struct {
	*engine.EmbedInterface
}

func (iface *ButtonSimpleIFace) Clicked(ev any) {
	log.Printf("\x1b[1m\x1b[33mButton\\Simple:\x1b[0m \x1b[33mI got '%d', time to trigger to the other node\x1b[0m\n", ev)

	iface.Node.Output["Clicked"].Set(ev)
}

// This will be called from example.go
func init() {
	Blackprint.RegisterNode("Example/Button/Simple", &engine.NodeRegister{
		Output: engine.PortTemplate{},
		Input:  engine.PortTemplate{},

		Constructor: func(node *engine.Node) {
			node.Embed = &ButtonSimple{}

			iface := node.SetInterface("BPIC/Example/Button")
			iface.Title = "Button"
		},
	})

	Blackprint.RegisterInterface("BPIC/Example/Button", &engine.InterfaceRegister{
		Constructor: func(iface *engine.Interface) {
			iface.Embed = &ButtonSimpleIFace{}
		},
	})
}
