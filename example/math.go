package example
import (
	"log"
	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/engine"
)

// This will be called from example.go
func RegisterMath(){
	RegisterMathMultiply()
	RegisterMathRandom()
}

// ============
type MathMultiple struct {
	engine.Node
}

// Your own processing mechanism
func (node MathMultiple) Multiply() int {
	log.Printf("\x1b[1m\x1b[33mMath\Multiply:\x1b[0m \x1b[33mMultiplying %d with %d\x1b[0m", node.Input['A'], node.Input['B'])
	return node.Input['A'] * node.Input['B']
}

// When any output value from other node are updated
// Let's immediately change current node result
func (node MathMultiple) Update(cable engine.Cable) {
	node.Output['Result'](node.Multiply())
}

func RegisterMathMultiply() {
	Blackprint.RegisterNode('Example/Math/Random', func(instance engine.Instance) interface{} {
		node := MathMultiple {
			Node: engine.Node {
				Instance: instance,

				// Node's Input Port
				Input: engine.NodePort {
					'Exec': port.Trigger(func() {
						node.Output['Result'](node.Multiply())
					}),
					'A': types.Int,
					'B': port.Validator(types.Int, func(val interface{}) int {
						log.Printf("\x1b[1m\x1b[33mMath\Multiply:\x1b[0m \x1b[33m%s - Port B got input: %d\x1b[0m", node.Iface.Title, val)
						return val // ToDo: convert string to number
					}),
				},

				// Node's Output Port
				Output: engine.NodePort {
					'Result': types.Int,
				}
			}
		}

		iface := node.SetInterface() // default interface
		iface.Title = "Multiply"

		node.On('cable.connect', func (ev) {
			log.Printf("\x1b[1m\x1b[33mMath\Multiply:\x1b[0m \x1b[33mCable connected from %s (%s) to %s (%s)\x1b[0m", ev.Port.Node.Title, ev.Port.Name, ev.Target.Node.Title, ev.Target.Name)
		})

		return node
	})
}


// ============
type MathRandom struct {
	engine.Node
	Executed bool
}

// When the connected node is requesting for the output value
func (node MathRandom) Request(port engine.Port, iface interface{}) bool {
	// Only run once this node never been executed
	// Return false if no value was changed
	if Executed == true {
		return false
	}

	log.Printf("\x1b[1m\x1b[33mMath\Random:\x1b[0m \x1b[33mValue request for port: %s, from node: %s\x1b[0m", port.Name, iface.Title)

	// Let's create the value for him
	node.Input['Re-seed']()

	return true
}

func RegisterMathRandom() {
	Blackprint.RegisterNode('Example/Math/Random', func(instance engine.Instance) interface{} {
		node := MathRandom {
			Executed: false,
			Node: engine.Node {
				Instance: instance,

				// Node's Input Port
				Input: engine.NodePort {
					'Re-seed': port.Trigger(func() {
						node.Executed = true
						node.Output['Out'](/*random*/)
					}),
				},

				// Node's Output Port
				Output: engine.NodePort {
					'Out': types.Int,
				}
			}
		}

		iface := node.SetInterface(); // default interface
		iface.Title = "Random";

		return node
	})
}