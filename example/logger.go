package example

import (
	"encoding/json"
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
	*engine.Node
}
type LoggerIFace struct {
	*engine.Interface
	log string
}

func (iface *LoggerIFace) Init() {
	refreshLogger := func(val any) {
		if val == nil {
			val = "nil"
			iface.Log(val)
		} else {
			types := reflect.TypeOf(val).Kind()

			if types == reflect.String || types == reflect.Int64 || types == reflect.Float64 {
				iface.Log(val)
			} else {
				byte_, _ := json.Marshal(val)
				iface.Log(string(byte_)) // ToDo, convert any object to JSON string
			}
		}
	}

	node := iface.Node.(*LoggerNode)

	// Let's show data after new cable was connected or disconnected
	iface.On("cable.connect cable.disconnect", func(_cable any) {
		log.Printf("\x1b[1m\x1b[33mDisplay\\Logger:\x1b[0m \x1b[33mA cable was changed on Logger, now refresing the input element\x1b[0m\n")
		refreshLogger(node.Input["Any"].Get())
	})

	iface.Input["Any"].On("value", func(_port any) {
		port := _port.(*engine.Port)
		log.Printf("\x1b[1m\x1b[33mDisplay\\Logger:\x1b[0m \x1b[33mI connected to %s (port %s), that have new value: %v\x1b[0m\n", port.Iface.Title, port.Name, port.Value)

		// Let's take all data from all connected nodes
		// Instead showing new single data. val
		refreshLogger(node.Input["Any"].Get())
	})
}

func (iface *LoggerIFace) Log(val ...any) any {
	if len(val) == 0 {
		return iface.log
	}

	iface.log = val[0].(string)
	log.Printf("\x1b[1m\x1b[33mLogger:\x1b[0m \x1b[33mLogger Data => %s\x1b[0m\n", iface.log)
	return nil
}

func RegisterLogger() {
	Blackprint.RegisterNode("Example/Display/Logger", func(instance *engine.Instance) any {
		node := &LoggerNode{
			Node: &engine.Node{
				Instance: instance,

				// Node's Input Port Template
				TInput: engine.NodePort{
					"Any": port.ArrayOf(reflect.Interface), // nil => Any
				},
			},
		}

		iface := node.SetInterface("BPIC/Example/Display/Logger").(*LoggerIFace)
		iface.Title = "Logger"

		return node
	})

	Blackprint.RegisterInterface("BPIC/Example/Display/Logger", func(node_ any) any {
		// node := node_.(LoggerNode)
		return &LoggerIFace{
			Interface: &engine.Interface{},
		}
	})
}
