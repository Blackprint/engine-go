package example
import (
	"log"
	"reflect"
	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/engine"
)

// This will be called from example.go
func RegisterLogger(){
	RegisterLogger()
}


// ============
type LoggerNode struct {
	engine.Node
}
type LoggerIFace struct {
	engine.Interface
	log
}

func (iface LoggerIFace) Init() {
	refreshLogger := func (val interface{}) {
		if val == nil {
			val = 'nil'
			iface.Log(val)
		} else {
			types = reflect.TypeOf(val)

			if types == 'string' || types == 'int' {
				iface.Log(val)
			} else {
				// val = json_encode(val);
				iface.Log(val) // ToDo, convert any object to JSON string
			}
		}
	}

	// Let's show data after new cable was connected or disconnected
	iface.On('cable.connect cable.disconnect', func() {
		log.Printf("\x1b[1m\x1b[33mDisplay\Logger:\x1b[0m \x1b[33mA cable was changed on Logger, now refresing the input element\x1b[0m", val)
		refreshLogger(iface.Node.Input['Any']())
	})

	iface.Input['Any'].On('value', func(port engine.Port) {
		log.Printf("\x1b[1m\x1b[33mDisplay\Logger:\x1b[0m \x1b[33mI connected to %s (port %s), that have new value: %s\x1b[0m", port.Name, port.Iface.Title, port.Value)

		// Let's take all data from all connected nodes
		// Instead showing new single data. val
		refreshLogger(iface.Node.Input['Any']())
	})
}

func (iface LoggerIFace) Log(val interface{}) {
	if(val == nil) return iface.log

	iface.log = val
	log.Printf("\x1b[1m\x1b[33mLogger:\x1b[0m \x1b[33mLogger Data => %s\x1b[0m", val)
}

func RegisterLogger(){
	Blackprint.RegisterNode('Example/Logger/', func(instance engine.Instance) interface{} {
		node := LoggerNode {
			Node: engine.Node {
				Instance: instance,

				// Node's Input Port
				Input: engine.NodePort {
					'Any': port.ArrayOf(nil), // nil => Any
				}
			}
		}

		iface := node.SetInterface('BPIC/Example/Logger')
		iface.Title = "Logger"

		return node
	})

	Blackprint.RegisterInterface('BPIC/Example/Logger', func(node LoggerNode) interface{} {
		return LoggerIFace {
			Interface: engine.Interface{}
		}
	})
}