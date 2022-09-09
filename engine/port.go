package engine

import (
	"fmt"
	"reflect"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

type PortStructTemplate struct {
	Type   any
	Field  string
	Handle func(any) any
}

type Port struct {
	CustomEvent
	Name        string
	Name_       *refPortName // For bpFunction only, ToDo: fill alternate name, search in engine-php _name for hints
	Type        reflect.Kind
	Types       []reflect.Kind
	Cables      []*Cable
	Source      int
	Iface       *Interface
	Default     any // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Value       any // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Sync        bool
	Feature     int
	QFeature    *portFeature // For caching the configuration
	Struct      map[string]PortStructTemplate
	Splitted    bool
	AllowResync bool // Retrigger connected node's .update when the output value is similar

	// Only in Golang we need to do this '-'
	RoutePort *routePort

	// Internal/Private property
	QCache          any
	QParent         *Port
	QStructSplitted bool
	QGhost          bool
	QFunc           func(*Port)
	QCallAll        func(*Port)
	QOnConnect      func(*Cable, *Port) bool
	QWaitPortInit   func(*Port)
}

/** For internal library use only */
// For bpFunction only
type refPortName struct {
	Name string
}

const (
	PortInput = iota + 1
	PortOutput
	// PortProperty
)

const (
	PortTypeArrayOf = iota + 1
	PortTypeDefault
	PortTypeTrigger
	PortTypeUnion
	PortTypeStructOf
	PortTypeRoute
)

// Port feature
type portFeature struct {
	Id    int
	Type  reflect.Kind
	Types []reflect.Kind
	Value any
	Func  func(*Port)
}

func (port *Port) QGetPortFeature() *portFeature {
	return port.QFeature
}
func (port *Port) DisconnectAll() {
	hasRemote := port.Iface.Node.Instance.QRemote == nil
	for _, cable := range port.Cables {
		if hasRemote {
			cable.QEvDisconnected = true
		}

		cable.Disconnect()
	}
}

// ./portGetterSetter.go
func (port *Port) CreateLinker() getterSetter {
	if port.Source == PortInput {
		return &portInputGetterSetter{port: port}
	}

	return &portOutputGetterSetter{port: port}
}

func (port *Port) sync() {
	skipSync := port.Iface.Node.Routes.Out != nil

	for _, cable := range port.Cables {
		inp := cable.Input
		if inp == nil {
			continue
		}

		inp.QCache = nil

		temp := &PortValueEvent{
			Target: inp,
			Port:   port,
			Cable:  cable,
		}
		inpIface := inp.Iface

		inp.Emit("value", temp)
		inpIface.Emit("value", temp)

		if skipSync {
			continue
		}

		node := inpIface.Node
		if inpIface.QRequesting == false && len(node.Routes.In) == 0 {
			node.Update(cable)

			if inpIface.QEnum == nodes.BPFnMain {
				node.Routes.RouteOut()
			} else {
				inpIface.QProxyInput.Routes.RouteOut()
			}
		}
	}
}

func (port *Port) DisableCables(enable bool) {
	if enable {
		for _, cable := range port.Cables {
			cable.Disabled = 1
		}
	} else if !enable {
		for _, cable := range port.Cables {
			cable.Disabled = 0
		}
	} else {
		panic("Unhandled, please check engine-php's implementation")
	}
}

type CableErrorEvent struct {
	Cable    *Cable
	OldCable *Cable
	Iface    *Interface
	Port     *Port
	Target   *Port
	Message  string
}

func (port *Port) QCableConnectError(name string, obj *CableErrorEvent, severe bool) {
	msg := "Cable notify: " + name
	if obj.Iface != nil {
		msg += "\nIFace: " + obj.Iface.Namespace
	}

	if obj.Port != nil {
		msg += fmt.Sprintf("\nFrom port: %s (iface: %s)\n - Type: %d) (%d)", obj.Port.Name, obj.Port.Iface.Namespace, obj.Port.Source, obj.Port.Type)
	}

	if obj.Target != nil {
		msg += fmt.Sprintf("\nTo port: %s (iface: %s)\n - Type: %d) (%d)", obj.Target.Name, obj.Target.Iface.Namespace, obj.Target.Source, obj.Target.Type)
	}

	obj.Message = msg
	instance := port.Iface.Node.Instance

	if severe && instance.PanicOnError {
		panic(msg + "\n\n")
	}

	instance.Emit(name, obj)
}
func (port *Port) ConnectCable(cable *Cable) bool {
	if cable.IsRoute {
		port.QCableConnectError("cable.not_route_port", &CableErrorEvent{
			Cable:  cable,
			Port:   port,
			Target: cable.Owner,
		}, true)

		cable.Disconnect()
		return false
	}

	if cable.Owner == port { // It's referencing to same port
		cable.Disconnect()
		return false
	}

	if (port.QOnConnect != nil && port.QOnConnect(cable, cable.Owner)) || (cable.Owner.QOnConnect != nil && cable.Owner.QOnConnect(cable, port)) {
		return false
	}

	// Remove cable if ...
	if (cable.Source == PortOutput && port.Source != PortInput) /* Output source not connected to input */ || (cable.Source == PortInput && port.Source != PortOutput) /* Input source not connected to output */ {
		port.QCableConnectError("cable.wrong_pair", &CableErrorEvent{
			Cable:  cable,
			Port:   port,
			Target: cable.Owner,
		}, true)

		cable.Disconnect()
		return false
	}

	if cable.Owner.Source == PortOutput {
		if (port.Feature == PortTypeArrayOf && !portArrayOf_validate(port, cable.Owner)) || (port.Feature == PortTypeUnion && !portUnion_validate(port, cable.Owner)) {
			port.QCableConnectError("cable.wrong_type", &CableErrorEvent{
				Cable:  cable,
				Iface:  port.Iface,
				Port:   cable.Owner,
				Target: port,
			}, true)

			cable.Disconnect()
			return false
		}
	} else if port.Source == PortOutput {
		if (cable.Owner.Feature == PortTypeArrayOf && !portArrayOf_validate(cable.Owner, port)) || (cable.Owner.Feature == PortTypeUnion && !portUnion_validate(cable.Owner, port)) {
			port.QCableConnectError("cable.wrong_type", &CableErrorEvent{
				Cable:  cable,
				Iface:  port.Iface,
				Port:   port,
				Target: cable.Owner,
			}, true)

			cable.Disconnect()
			return false
		}
	}

	// Golang can't check by class instance or inheritance
	// ==========================================
	// ToDo: recheck why we need to check if the constructor is a function
	// isInstance = true
	// if cable.Owner.Type != port.Type && cable.Owner.Type == types.Function && port.Type == types.Function {
	// 	if cable.Owner.Source == PortOutput{
	// 		isInstance = cable.Owner.Type instanceof port.Type
	// 	} else {
	// 		isInstance =  port.Type instanceof cable.Owner.Type
	// 	}
	// }
	// ==========================================

	// Remove cable if type restriction
	// if !isInstance || (cable.Owner.Type == types.Function && port.Type != types.Function || cable.Owner.Type != types.Function && port.Type == types.Function) {
	if cable.Owner.Type == types.Function && port.Type != types.Function || cable.Owner.Type != types.Function && port.Type == types.Function {
		port.QCableConnectError("cable.wrong_type_pair", &CableErrorEvent{
			Cable:  cable,
			Port:   port,
			Target: cable.Owner,
		}, true)

		cable.Disconnect()
		return false
	}

	// Restrict connection between function input/output node with variable node
	// Connection to similar node function IO or variable node also restricted
	// These port is created on runtime dynamically
	if port.Iface.QDynamicPort && cable.Owner.Iface.QDynamicPort {
		port.QCableConnectError("cable.unsupported_dynamic_port", &CableErrorEvent{
			Cable:  cable,
			Port:   port,
			Target: cable.Owner,
		}, true)

		cable.Disconnect()
		return false
	}

	// Remove cable if there are similar connection for the ports
	for _, cable := range cable.Owner.Cables {
		if utils.Contains(port.Cables, cable) {
			port.QCableConnectError("cable.duplicate_removed", &CableErrorEvent{
				Cable:  cable,
				Port:   port,
				Target: cable.Owner,
			}, false)

			cable.Disconnect()
			return false
		}
	}

	// Put port reference to the cable
	cable.Target = port

	var inp *Port
	var out *Port
	if cable.Target.Source == PortInput {
		/** @var Port */
		inp = cable.Target
		out = cable.Owner
	} else {
		/** @var Port */
		inp = cable.Owner
		out = cable.Target
	}

	// Remove old cable if the port not support array
	if inp.Feature != PortTypeArrayOf && inp.Type != types.Function {
		cables := inp.Cables // Cables in input port

		if len(cables) != 0 {
			temp := cables[0]

			if temp == cable {
				temp = cables[1]
			}

			if temp != nil {
				inp.QCableConnectError("cable.replaced", &CableErrorEvent{
					Cable:    cable,
					OldCable: temp,
					Port:     inp,
					Target:   out,
				}, false)

				temp.Disconnect()
				return false
			}
		}
	}

	// Connect this cable into port's cable list
	port.Cables = append(port.Cables, cable)
	// cable.Connecting()
	cable.QConnected()

	return true
}
func (port *Port) ConnectPort(portTarget *Port) bool {
	cable := newCable(portTarget, port)
	if portTarget.QGhost {
		cable.QGhost = true
	}

	portTarget.Cables = append(portTarget.Cables, cable)
	return port.ConnectCable(cable)
}
