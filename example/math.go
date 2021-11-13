package example

import (
	"fmt"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
)

// This will be called from example.go
func RegisterMath() {
	RegisterMathMultiply()
	RegisterMathRandom()
}

// ============
type MathMultiple struct {
	engine.Node
}

// Your own processing mechanism
func (node *MathMultiple) Multiply() int {
	fmt.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mMultiplying %d with %d\x1b[0m\n", node.Input["A"]().(int), node.Input["B"]().(int))
	return node.Input["A"]().(int) * node.Input["B"]().(int)
}

// When any output value from other node are updated
// Let's immediately change current node result
func (node *MathMultiple) Update(cable engine.Cable) {
	node.Output["Result"](node.Multiply())
}

func RegisterMathMultiply() {
	Blackprint.RegisterNode("Example/Math/Multiply", func(instance *engine.Instance) interface{} {
		var node MathMultiple
		node = MathMultiple{
			Node: engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Exec": port.Trigger(func(...interface{}) {
						node.Output["Result"](node.Multiply())
					}),
					"A": types.Int,
					"B": port.Validator(types.Int, func(val interface{}) interface{} {
						fmt.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33m%s - Port B got input: %d\x1b[0m\n", node.Iface.(engine.Interface).Title, val)
						return val // ToDo: convert string to number
					}),
				},

				// Node's Output Port Template
				TOutput: engine.NodePort{
					"Result": types.Int,
				},
			},
		}

		iface := node.SetInterface().(*engine.Interface) // default interface
		iface.Title = "Multiply"

		node.On("cable.connect", func(event interface{}) {
			ev := event.(engine.CableEvent)
			fmt.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mCable connected from %s (%s) to %s (%s)\x1b[0m\n", ev.Port.Iface.Title, ev.Port.Name, ev.Target.Iface.Title, ev.Target.Name)
		})

		return &node
	})
}

// ============
type MathRandom struct {
	engine.Node
	Executed bool
}

// When the connected node is requesting for the output value
func (node *MathRandom) Request(port *engine.Port, iface_ interface{}) bool {
	// Only run once this node never been executed
	// Return false if no value was changed
	if node.Executed == true {
		return false
	}

	iface := iface_.(*engine.Interface)
	fmt.Printf("\x1b[1m\x1b[33mMath\\Random:\x1b[0m \x1b[33mValue request for port: %s, from node: %s\x1b[0m\n", port.Name, iface.Title)

	// Let's create the value for him
	node.Input["Re-seed"]()

	return true
}

func RegisterMathRandom() {
	Blackprint.RegisterNode("Example/Math/Random", func(instance *engine.Instance) interface{} {
		var node MathRandom
		node = MathRandom{
			Executed: false,
			Node: engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Re-seed": port.Trigger(func(...interface{}) {
						node.Executed = true
						node.Output["Out"]( /*random*/ )
					}),
				},

				// Node's Output Port Template
				TOutput: engine.NodePort{
					"Out": types.Int,
				},
			},
		}

		iface := node.SetInterface().(*engine.Interface) // default interface
		iface.Title = "Random"

		return &node
	})
}
