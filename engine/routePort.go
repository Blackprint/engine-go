package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/utils"
)

type RoutePort struct {
	*Port
	In         []*Cable
	Out        *Cable
	DisableOut bool
	Disabled   bool
	IsRoute    bool
	Iface      any // any = extends *engine.Interface

	// for internal library use only
	QIsPaused bool
}

func newRoutePort(iface any) *RoutePort {
	temp := &RoutePort{
		Iface:   iface,
		IsRoute: true,
	}
	temp.Port = &Port{RoutePort: temp, Iface: iface}
	temp.Port.Cables = []*Cable{}
	temp.In = temp.Port.Cables
	return temp
}

// Connect other route port (this .out to other .in port)
func (r *RoutePort) RouteTo(iface any) {
	if r.Out != nil {
		r.Out.Disconnect()
	}

	if iface == nil {
		cable := NewCable(r.Port, nil)
		cable.IsRoute = true
		r.Out = cable
		return
	}

	port := utils.GetProperty(utils.GetProperty(iface, "Node"), "Routes").(*RoutePort)

	cable := NewCable(r.Port, port.Port)
	cable.IsRoute = true
	cable.Output = r.Port
	r.Out = cable
	port.In = append(port.In, cable) // ToDo: check if this empty if the connected cable was disconnected

	cable.QConnected()
}

func (r *RoutePort) ConnectCable(cable *Cable) bool {
	if utils.Contains(r.In, cable) {
		return false
	}

	r.In = append(r.In, cable)
	cable.Input = r.Port
	cable.Target = r.Port
	cable.QConnected()

	return true
}

func (r *RoutePort) RouteIn(cable *Cable) {
	node := utils.GetProperty(r.Iface, "Node")
	utils.CallFunction(node, "Update", &[]reflect.Value{
		reflect.ValueOf(cable),
	})

	routes := utils.GetProperty(node, "Routes")
	utils.CallFunction(routes, "RouteOut", utils.EmptyArgs)
}

func (r *RoutePort) RouteOut() {
	if r.DisableOut {
		return
	}

	if r.Out == nil {
		if utils.GetProperty(r.Iface, "QEnum").(int) == nodes.BPFnOutput {
			node := utils.GetProperty(utils.GetProperty(r.Iface, "QFuncMain"), "Node")
			route := utils.GetProperty(node, "Routes").(*RoutePort)
			utils.CallFunction(route, "RouteIn", utils.EmptyArgs)
		}

		return
	}

	targetRoute := r.Out.Input.RoutePort
	if targetRoute == nil {
		return
	}

	enum := utils.GetProperty(targetRoute.Iface, "QEnum").(int)

	if enum == 0 {
		targetRoute.RouteIn(r.Out)
	} else if enum == nodes.BPFnMain {
		routes := utils.GetProperty(utils.GetProperty(targetRoute.Iface, "QProxyInput"), "Routes").(*RoutePort)
		routes.RouteIn(r.Out)
	} else if enum == nodes.BPFnOutput {
		node := utils.GetProperty(utils.GetProperty(targetRoute.Iface, "QFuncMain"), "Node")
		routes := utils.GetProperty(node, "Routes").(*RoutePort)
		routes.RouteIn(r.Out)
	} else {
		targetRoute.RouteIn(r.Out)
	}
}
