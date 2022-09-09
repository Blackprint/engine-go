package engine

import (
	"github.com/blackprint/engine-go/types"
)

type portInputGetterSetter struct {
	getterSetter
	port *Port
}

func (gs *portInputGetterSetter) Set(val any) {
	panic("Can't set input port's value")
}

func (gs *portInputGetterSetter) Call() {
	gs.port.QFunc(gs.port)
	gs.port.Iface.Node.Routes.RouteOut()
}

func (gs *portInputGetterSetter) Get() any {
	port := gs.port

	// This port must use values from connected output
	cableLen := len(port.Cables)

	if cableLen == 0 {
		if port.Feature == PortTypeArrayOf {
			// ToDo: fix type to follow
			// the type from port.Type

			return [](any){}
		}

		return port.Default
	}

	// Flag current iface is requesting value to other iface

	// Return single data
	if cableLen == 1 {
		temp := port.Cables[0]
		var target *Port

		if temp.Owner == port {
			target = temp.Target
		} else {
			target = temp.Owner
		}

		if target.Value == nil {
			port.Iface.QRequesting = true
			target.Iface.Node.Request(temp)
			port.Iface.QRequesting = false
		}

		// fmt.Printf("1. %s -> %s (%s)\n", port.Name, target.Name, target.Value)

		if port.Feature == PortTypeArrayOf {
			var tempVal any
			if target.Value == nil {
				tempVal = target.Default
			} else {
				tempVal = target.Value
			}

			return [](any){tempVal}
		}

		if target.Value == nil {
			return target.Default
		} else {
			return target.Value
		}
	}

	// Return multiple data as an array
	data := []any{}
	for _, cable := range port.Cables {
		var target *Port
		if cable.Owner == port {
			target = cable.Target
		} else {
			target = cable.Owner
		}

		if target.Value == nil {
			port.Iface.QRequesting = true
			target.Iface.Node.Request(cable)
			port.Iface.QRequesting = false
		}

		// fmt.Printf("2. %s -> %s (%s)\n", port.Name, target.Name, target.Value)

		if target.Value == nil {
			data = append(data, target.Default)
		} else {
			data = append(data, target.Value)
		}
	}

	if port.Feature != PortTypeArrayOf {
		return data[0]
	}

	return data
}

type portOutputGetterSetter struct {
	getterSetter
	port *Port
}

func (gs *portOutputGetterSetter) Set(val any) {
	port := gs.port

	if port.Source == PortInput {
		panic("Can't set data to input port")
	}

	// ToDo: do we need feature validation here?
	// fmt.Printf("3. %s = %s\n", port.Name, val)

	port.Value = val
	port.Emit("value", &PortValueEvent{
		Port: port,
	})
	port.sync()
}

// createCallablePort
// createCallableRoutePort
func (gs *portOutputGetterSetter) Call() {
	if gs.port.Type == types.Route {
		cable := gs.port.Cables[0]
		if cable == nil {
			return
		}

		cable.Input.RoutePort.RouteIn(cable)
	} else {
		for _, cable := range gs.port.Cables {
			target := cable.Input
			if target == nil {
				continue
			}

			// fmt.Println(cable.String())
			if target.Name_ != nil {
				target.Iface.QParentFunc.Node.Output[target.Name_.Name].Call()
			} else {
				target.Iface.Node.Input[target.Name].Call()
			}
		}

		gs.port.Emit("call", nil)
	}
}

func (gs *portOutputGetterSetter) Get() any {
	port := gs.port

	if port.Feature == PortTypeArrayOf {
		var tempVal any
		if port.Value == nil {
			tempVal = port.Default
		} else {
			tempVal = port.Value
		}

		return [](any){tempVal}
	}

	if port.Value == nil {
		return port.Default
	}

	return port.Value
}
