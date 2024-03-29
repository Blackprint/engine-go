package example

import (
	"crypto/rand"
	"encoding/binary"
	"log"
	"reflect"
	"strconv"

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
	log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mMultiplying %d with %d\x1b[0m\n", node.Input["A"].Get().(int), node.Input["B"].Get().(int))
	return node.Input["A"].Get().(int) * node.Input["B"].Get().(int)
}

// When any output value from other node are updated
// Let's immediately change current node result
func (node *MathMultiple) Update(cable *engine.Cable) {
	node.Output["Result"].Set(node.Multiply())
}

func RegisterMathMultiply() {
	Blackprint.RegisterNode("Example/Math/Multiply", func(instance *engine.Instance) any {
		var node MathMultiple
		node = MathMultiple{
			Node: engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Exec": port.Trigger(func(port *engine.Port) {
						node.Output["Result"].Set(node.Multiply())
						log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mResult has been set: %d\x1b[0m\n", node.Output["Result"].Get())
					}),
					"A": types.Int,
					"B": port.Validator(types.Int, func(val any) any {
						log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33m%s - Port B got input: %d\x1b[0m\n", node.IFace.(*engine.Interface).Title, val)

						// Convert string to number
						if reflect.ValueOf(val).Kind() == reflect.String {
							num, _ := strconv.Atoi(val.(string))
							return num
						}

						return val
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

		node.On("cable.connect", func(event any) {
			ev := event.(engine.CableEvent)
			log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mCable connected from %s (%s) to %s (%s)\x1b[0m\n", ev.Port.Iface.Title, ev.Port.Name, ev.Target.Iface.Title, ev.Target.Name)
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
func (node *MathRandom) Request(port *engine.Port, iface_ any) bool {
	// Only run once this node never been executed
	// Return false if no value was changed
	if node.Executed == true {
		return false
	}

	iface := iface_.(*engine.Interface)
	log.Printf("\x1b[1m\x1b[33mMath\\Random:\x1b[0m \x1b[33mValue request for port: %s, from node: %s\x1b[0m\n", port.Name, iface.Title)

	// Let's create the value for him
	node.Input["Re-seed"].Call()

	return true
}

func RegisterMathRandom() {
	Blackprint.RegisterNode("Example/Math/Random", func(instance *engine.Instance) any {
		var node MathRandom
		node = MathRandom{
			Executed: false,
			Node: engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Re-seed": port.Trigger(func(port *engine.Port) {
						node.Executed = true
						byt := make([]byte, 2)
						rand.Read(byt)
						node.Output["Out"].Set(int(binary.BigEndian.Uint16(byt[:])) % 100)
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
