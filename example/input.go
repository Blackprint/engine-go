package example
import (
	"log"
	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/engine"
)

// This will be called from example.go
func RegisterInput(){
	RegisterInputSimple()
}


// ============
type InputSimple struct {
	engine.Node
}

// Bring value from imported iface to node output
func (node InputSimple) Imported() {
	val := node.Iface.Data['value']();
	if val != nil {
		log.Printf("\x1b[1m\x1b[33mInput\Simple:\x1b[0m \x1b[33mSaved data as output: %s\x1b[0m", val)
	}

	node.Output['Value']($val);
}

type InputSimpleIFace struct {
	engine.Interface
}
func (iface InputSimpleIFace) Changed(val interface{}) {
	// This node still being imported
	if iface.Importing != false {
		return
	}

	log.Printf("\x1b[1m\x1b[33mInput\Simple:\x1b[0m \x1b[33mThe input box have new value: %s\x1b[0m", val)

	iface.Node.Output['Value'](val);

	// This will call every connected node
	iface.Node.Output['Changed']();
}

func RegisterInputSimple(){
	Blackprint.RegisterNode('Example/Input/Simple', func(instance engine.Instance) interface{} {
		node := InputSimple {
			Node: engine.Node {
				Instance: instance,

				// Node's Output Port
				Output: engine.NodePort {
					'Changed': types.Function,
					'Value': types.String,
				}
			}
		}

		iface := node.SetInterface('BPIC/Example/Input')
		iface.Title = "Input"

		return node
	})

	Blackprint.RegisterInterface('BPIC/Example/Input', func(node InputSimple) interface{} {
		value := '...';

		iface := InputSimpleIFace {
			Interface: engine.Interface {
				Data: engine.InterfaceData {
					'value': func(val interface{}) interface{} {
						if val == nil {
							return value
						}

						value = val
						iface.Changed(val)
					}
				}
			},
		}

		return iface
	})
}