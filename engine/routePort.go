package engine

import (
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
	Iface      *engine.Interface

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

	port := iface.Node.Routes.(*RoutePort)

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
	node := r.Iface.Node
	node.Update(cable)

	routes := node.Routes
	routes.RouteOut()
}

func (r *RoutePort) RouteOut() {
	if r.DisableOut {
		return
	}

	if r.Out == nil {
		if r.Iface.QEnum.(int) == nodes.BPFnOutput {
			node := r.Iface.QFuncMain.Node
			route := node.Routes.(*RoutePort)
			route.RouteIn()
		}

		return
	}

	targetRoute := r.Out.Input.RoutePort
	if targetRoute == nil {
		return
	}

	enum := targetRoute.Iface.QEnum.(int)

	if enum == 0 {
		targetRoute.RouteIn(r.Out)
	} else if enum == nodes.BPFnMain {
		routes := targetRoute.Iface.QProxyInput.Routes.(*RoutePort)
		routes.RouteIn(r.Out)
	} else if enum == nodes.BPFnOutput {
		node := targetRoute.Iface.QFuncMain.Node
		routes := node.Routes.(*RoutePort)
		routes.RouteIn(r.Out)
	} else {
		targetRoute.RouteIn(r.Out)
	}
}
