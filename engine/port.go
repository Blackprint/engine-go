package engine

import (
	"reflect"

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
	Type        reflect.Kind
	Types       []reflect.Kind
	Cables      []*Cable
	Source      int
	Iface       any
	Default     any // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Value       any // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Func        func(any)
	Sync        bool
	Feature     int
	Struct      map[string]PortStructTemplate
	Splitted    bool
	AllowResync bool

	// Only in Golang we need to do this '-'
	RoutePort *RoutePort

	// Internal/Private property
	QCache          any
	QParent         *Port
	QStructSplitted bool
}

const (
	PortInput = iota + 1
	PortOutput
	// PortProperty
)

const (
	PortTypeArrayOf = iota + 1
	PortTypeDefault
	PortTypeSwitch
	PortTypeTrigger
	PortTypeUnion
	PortTypeValidator
)

// Port feature
type PortFeature struct {
	Id    int
	Type  reflect.Kind
	Types []reflect.Kind
	Value any
	Func  any
}

type PortInputGetterSetter struct {
	getterSetter
	port *Port
}

func (gs *PortInputGetterSetter) Set(val any) {
	panic("Can't set input port's value")
}

func (gs *PortInputGetterSetter) Call() {
	gs.port.Default.(func(*Port))(gs.port)
}

func (gs *PortInputGetterSetter) Get() any {
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
			utils.CallFunction(target.Iface.Node, "Request", &[]reflect.Value{
				reflect.ValueOf(target),
				reflect.ValueOf(port.Iface),
			})
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
			utils.CallFunction(target.Iface.Node, "Request", &[]reflect.Value{
				reflect.ValueOf(target),
				reflect.ValueOf(port.Iface),
			})
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

type PortOutputGetterSetter struct {
	getterSetter
	port *Port
}

func (gs *PortOutputGetterSetter) Set(val any) {
	port := gs.port

	if port.Source == PortInput {
		panic("Can't set data to input port")
	}

	// ToDo: do we need feature validation here?
	// fmt.Printf("3. %s = %s\n", port.Name, val)

	port.Value = val
	port.Emit("value", port)
	port.sync()
}

func (gs *PortOutputGetterSetter) Call() {
	var target *Port
	for _, cable := range gs.port.Cables {
		if cable.Owner == gs.port {
			target = cable.Target
		} else {
			target = cable.Owner
		}

		// fmt.Println(cable.String())
		target.Default.(func(*Port))(target)
	}
}

func (gs *PortOutputGetterSetter) Get() any {
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

func (port *Port) CreateLinker() getterSetter {
	if port.Source == PortInput {
		return &PortInputGetterSetter{port: port}
	}

	return &PortOutputGetterSetter{port: port}
}

func (port *Port) sync() {
	var target *Port
	for _, cable := range port.Cables {
		if cable.Owner == port {
			target = cable.Target
		} else {
			target = cable.Owner
		}

		if target.Iface.QRequesting == false {
			utils.CallFunction(target.Iface.Node, "Update", &[]reflect.Value{
				reflect.ValueOf(cable),
			})
		}

		target.Emit("value", port)
	}
}
