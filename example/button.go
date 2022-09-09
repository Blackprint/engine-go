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
	Blackprint.RegisterNode("Example/Button/Simple", &engine.NodeMetadata{
		Output: engine.NodePortTemplate{},
		Input:  engine.NodePortTemplate{},
	},
		func(instance *engine.Instance) *engine.Node {
			node := &engine.Node{
				Embed: &ButtonSimple{},
			}

			iface := node.SetInterface("BPIC/Example/Button")
			iface.Title = "Button"

			return node
		})

	Blackprint.RegisterInterface("BPIC/Example/Button",
		func(node *engine.Node) *engine.Interface {
			return &engine.Interface{
				Embed: &ButtonSimpleIFace{},
			}
		})
}
