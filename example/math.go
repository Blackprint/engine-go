package example

import (
	"crypto/rand"
	"encoding/binary"
	"log"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// ============ MathMultiple Node ============
type MathMultiple struct {
	engine.EmbedNode
}

// Your own processing mechanism
func (this *MathMultiple) Multiply() int {
	log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mMultiplying %d with %d\x1b[0m\n", this.Node.Input["A"].Get().(int), this.Node.Input["B"].Get().(int))
	return this.Node.Input["A"].Get().(int) * this.Node.Input["B"].Get().(int)
}

// When any output value from other node are updated
// Let's immediately change current node result
func (this *MathMultiple) Update(cable *engine.Cable) {
	this.Node.Output["Result"].Set(this.Multiply())
}

func init() {
	Blackprint.RegisterNode("Example/Math/Multiply", &engine.NodeMetadata{
		Input: engine.NodePortTemplate{
			"Exec": engine.Ports.Trigger(func(port *engine.Port) {
				port.Iface.Node.Output["Result"].Set(port.Iface.Node.Embed.(*MathMultiple).Multiply())
				log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mResult has been set: %d\x1b[0m\n", port.Iface.Node.Output["Result"].Get())
			}),
			"A": types.Int,
			"B": types.Any,
		},

		Output: engine.NodePortTemplate{
			"Result": types.Int,
		},
	},
		func(instance *engine.Instance) *engine.Node {
			node := &engine.Node{
				Instance: instance,
				Embed:    &MathMultiple{},
			}

			iface := node.SetInterface()
			iface.Title = "Multiply"

			iface.On("cable.connect", func(event any) {
				ev := event.(engine.CableEvent)
				log.Printf("\x1b[1m\x1b[33mMath\\Multiply:\x1b[0m \x1b[33mCable connected from %s (%s) to %s (%s)\x1b[0m\n", ev.Port.Iface.Title, ev.Port.Name, ev.Target.Iface.Title, ev.Target.Name)
			})

			return node
		})
}

// ============ MathRandom Node ============
type MathRandom struct {
	engine.EmbedNode
	Executed bool
}

// When the connected node is requesting for the output value
func (this *MathRandom) Request(cable *engine.Cable) {
	// Only run once this node never been executed
	// Return false if no value was changed
	if this.Executed == true {
		return
	}

	log.Printf("\x1b[1m\x1b[33mMath\\Random:\x1b[0m \x1b[33mValue request for port: %s, from node: %s\x1b[0m\n", cable.Output.Name, cable.Input.Iface.Title)

	// Let's create the value for him
	this.Node.Input["Re-seed"].Call()

	return
}

func init() {
	Blackprint.RegisterNode("Example/Math/Random", &engine.NodeMetadata{
		Input: engine.NodePortTemplate{
			"Re-seed": engine.Ports.Trigger(func(port *engine.Port) {
				node := port.Iface.Node
				node.Embed.(*MathRandom).Executed = true

				byt := make([]byte, 2)
				rand.Read(byt)
				node.Output["Out"].Set(int(binary.BigEndian.Uint16(byt[:])) % 100)
			}),
		},

		Output: engine.NodePortTemplate{
			"Out": types.Int,
		},
	},
		func(instance *engine.Instance) *engine.Node {
			node := &engine.Node{
				Instance: instance,
				Embed:    &MathRandom{},
			}

			iface := node.SetInterface()
			iface.Title = "Random"

			return node
		})
}
