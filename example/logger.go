package example

import (
	"encoding/json"
	"log"
	"reflect"

	Blackprint "github.com/blackprint/engine-go/blackprint"
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/types"
)

// ============
type LoggerNode struct {
	*engine.EmbedNode
}
type LoggerIFace struct {
	*engine.EmbedInterface
	log string
}

func (this *LoggerIFace) Init() {
	refreshLogger := func(val any) {
		if val == nil {
			val = "nil"
			this.Log(val)
		} else {
			types := reflect.TypeOf(val).Kind()

			if types == reflect.String || types == reflect.Int64 || types == reflect.Float64 {
				this.Log(val)
			} else {
				byte_, _ := json.Marshal(val)
				this.Log(string(byte_)) // ToDo, convert any object to JSON string
			}
		}
	}

	node := this.Node

	// Let's show data after new cable was connected or disconnected
	this.Iface.On("cable.connect cable.disconnect", func(_cable any) {
		log.Printf("\x1b[1m\x1b[33mDisplay\\Logger:\x1b[0m \x1b[33mA cable was changed on Logger, now refresing the input element\x1b[0m\n")
		refreshLogger(node.Input["Any"].Get())
	})

	this.Iface.Input["Any"].On("value", func(_port any) {
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

func init() {
	Blackprint.RegisterNode("Example/Display/Logger", &engine.NodeRegister{
		Input: engine.PortTemplate{
			"Any": Blackprint.Port.ArrayOf(types.Any), // nil => Any
		},

		Constructor: func(node *engine.Node) {
			node.Embed = &LoggerNode{}

			iface := node.SetInterface("BPIC/Example/Display/Logger")
			iface.Title = "Logger"
		},
	})

	Blackprint.RegisterInterface("BPIC/Example/Display/Logger", &engine.InterfaceRegister{
		Constructor: func(iface *engine.Interface) {
			iface.Embed = &LoggerIFace{}
		},
	})
}
