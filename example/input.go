package example

import (
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// ============
type InputSimple struct {
	*engine.EmbedNode
}

// Bring value from imported iface to node output
func (this *InputSimple) Imported(map[string]any) {
	val := this.Iface.Data["value"].Get()
	if val != nil {
		log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mSaved data as output: %s\x1b[0m\n", val)
	}

	this.Node.Output["Value"].Set(val)
}

type InputSimpleIFace struct {
	*engine.EmbedInterface
}

func (this *InputSimpleIFace) Changed(val any) {
	// This node still being imported
	if this.Iface.Importing != false {
		return
	}

	log.Printf("\x1b[1m\x1b[33mInput\\Simple:\x1b[0m \x1b[33mThe input box have new value: %s\x1b[0m\n", val)

	node := this.Node
	node.Output["Value"].Set(val)

	// This will call every connected node
	node.Output["Changed"].Call()
}

type MyData struct {
	engine.GetterSetter
	Value any
}

func (gs *MyData) Set(val any) {
	gs.Value = val
	gs.Iface.Embed.(*InputSimpleIFace).Changed(gs.Value)
}

func (gs *MyData) Get() any {
	return gs.Value
}

// This will be called from example.go
func init() {
	Blackprint.RegisterNode("Example/Input/Simple", &engine.NodeRegister{
		Output: engine.PortTemplate{
			"Changed": types.Function,
			"Value":   types.String,
		},

		Constructor: func(node *engine.Node) {
			node.Embed = &InputSimple{}

			iface := node.SetInterface("BPIC/Example/Input")
			iface.Title = "Input"
		},
	})

	Blackprint.RegisterInterface("BPIC/Example/Input", &engine.InterfaceRegister{
		Constructor: func(iface *engine.Interface) {
			iface.Embed = &InputSimpleIFace{}
			iface.Data = engine.InterfaceData{
				"value": &MyData{Value: "..."},
			}
		},
	})
}
