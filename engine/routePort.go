package engine

import (
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/utils"
)

type routePort struct {
	*Port
	In         []*Cable
	Out        *Cable
	DisableOut bool
	Disabled   bool
	IsRoute    bool
	Iface      *Interface

	// for internal library use only
	_isPaused bool
}

func newRoutePort(iface *Interface) *routePort {
	temp := &routePort{
		Iface:   iface,
		IsRoute: true,
	}
	temp.Port = &Port{RoutePort: temp, Iface: iface}
	temp.Port.Cables = []*Cable{}
	temp.In = temp.Port.Cables
	return temp
}

// Connect other route port (this .out to other .in port)
func (r *routePort) RouteTo(iface *Interface) {
	if r.Out != nil {
		r.Out.Disconnect()
	}

	if iface == nil {
		cable := newCable(r.Port, nil)
		cable.IsRoute = true
		r.Out = cable
		return
	}

	port := iface.Node.Routes

	cable := newCable(r.Port, port.Port)
	cable.IsRoute = true
	cable.Output = r.Port
	r.Out = cable
	port.In = append(port.In, cable) // ToDo: check if this empty if the connected cable was disconnected

	cable._connected()
}

func (r *routePort) ConnectCable(cable *Cable) bool {
	if utils.Contains(r.In, cable) {
		return false
	}

	r.In = append(r.In, cable)
	cable.Input = r.Port
	cable.Target = r.Port
	cable._connected()

	return true
}

func (r *routePort) RouteIn(cable *Cable) {
	node := r.Iface.Node
	node.Update(cable)

	routes := node.Routes
	routes.RouteOut()
}

func (r *routePort) RouteOut() {
	if r.DisableOut {
		return
	}

	if r.Out == nil {
		if r.Iface._enum == nodes.BPFnOutput {
			node := r.Iface._funcMain.Node
			route := node.Routes
			route.RouteIn(nil)
		}

		return
	}

	targetRoute := r.Out.Input.RoutePort
	if targetRoute == nil {
		return
	}

	enum := targetRoute.Iface._enum

	if enum == 0 {
		targetRoute.RouteIn(r.Out)
	} else if enum == nodes.BPFnMain {
		routes := targetRoute.Iface._proxyInput.Routes
		routes.RouteIn(r.Out)
	} else if enum == nodes.BPFnOutput {
		node := targetRoute.Iface._funcMain.Node
		routes := node.Routes
		routes.RouteIn(r.Out)
	} else {
		targetRoute.RouteIn(r.Out)
	}
}
