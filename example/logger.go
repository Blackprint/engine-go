package example

import (
	"log"
	"reflect"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/port"
)

// This will be called from example.go
func RegisterDisplay() {
	RegisterLogger()
}

// ============
type LoggerNode struct {
	engine.Node
}
type LoggerIFace struct {
	engine.Interface
	log string
}

func (iface *LoggerIFace) Init() {
	refreshLogger := func(val interface{}) {
		if val == nil {
			val = "nil"
			iface.Log(val)
		} else {
			types := reflect.TypeOf(val).Kind()

			if types == reflect.String || types == reflect.Int64 || types == reflect.Float64 {
				iface.Log(val)
			} else {
				// val = json_encode(val);
				iface.Log(val) // ToDo, convert any object to JSON string
			}
		}
	}

	node := iface.Node.(*LoggerNode)

	// Let's show data after new cable was connected or disconnected
	iface.On("cable.connect cable.disconnect", func(_cable interface{}) {
		log.Printf("\x1b[1m\x1b[33mDisplay\\Logger:\x1b[0m \x1b[33mA cable was changed on Logger, now refresing the input element\x1b[0m")
		refreshLogger(node.Input["Any"]())
	})

	iface.Input["Any"].On("value", func(_port interface{}) {
		port := _port.(engine.Port)
		log.Printf("\x1b[1m\x1b[33mDisplay\\Logger:\x1b[0m \x1b[33mI connected to %s (port %s), that have new value: %s\x1b[0m", port.Name, port.Iface.Title, port.Value)

		// Let's take all data from all connected nodes
		// Instead showing new single data. val
		refreshLogger(node.Input["Any"]())
	})
}

func (iface *LoggerIFace) Log(val interface{}) interface{} {
	if val == nil {
		return iface.log
	}

	iface.log = val.(string)
	log.Printf("\x1b[1m\x1b[33mLogger:\x1b[0m \x1b[33mLogger Data => %s\x1b[0m", val)
	return nil
}

func RegisterLogger() {
	Blackprint.RegisterNode("Example/Display/Logger", func(instance *engine.Instance) interface{} {
		node := LoggerNode{
			Node: engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Any": port.ArrayOf(reflect.Interface), // nil => Any
				},
			},
		}

		iface := node.SetInterface("BPIC/Example/Display/Logger").(*LoggerIFace)
		iface.Title = "Logger"

		return &node
	})

	Blackprint.RegisterInterface("BPIC/Example/Display/Logger", func(node_ interface{}) interface{} {
		// node := node_.(LoggerNode)
		return &LoggerIFace{
			Interface: engine.Interface{},
		}
	})
}
