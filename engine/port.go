package engine

import (
	"reflect"

	portTypes "github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

type Port struct {
	customEvent
	Name    string
	Type    reflect.Kind
	Cables  []*Cable
	Source  int
	Iface   *Interface
	Default interface{} // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Value   interface{} // Dynamic data (depend on Type) for storing port value (int, string, map, etc..)
	Func    func(interface{})
	Sync    bool
	Feature int
}

func (port *Port) CreateLinker() portTypes.GetterSetter {
	if port.Type == types.Function {
		if port.Source == portTypes.Output {
			return func(data ...interface{}) interface{} {
				var target *Port
				for _, cable := range port.Cables {
					if cable.Owner == port {
						target = cable.Target
					} else {
						target = cable.Owner
					}

					args := []reflect.Value{}

					for _, v := range data {
						args = append(args, reflect.ValueOf(v))
					}

					// fmt.Println(cable.String() + " (args: " + strconv.Itoa(len(args)) + ")")
					utils.CallFieldFunction(target, "Default", &args)
				}

				return nil
			}
		}

		return func(data ...interface{}) interface{} {
			return port.Default
		}
	}

	return func(val ...interface{}) interface{} {
		// Getter value
		if len(val) == 0 {
			// This port must use values from connected output
			if port.Source == portTypes.Input {
				cableLen := len(port.Cables)
				if cableLen == 0 {
					if port.Feature == portTypes.TypeArrayOf {
						// ToDo: fix type to follow
						// the type from port.Type

						return [](interface{}){}
					}

					return port.Default
				}

				// Flag current iface is requesting value to other iface
				port.Iface.QRequesting = true

				// Return single data
				if cableLen == 1 {
					temp := port.Cables[0]
					var target *Port

					if temp.Owner == port {
						target = temp.Target
					} else {
						target = temp.Owner
					}

					utils.CallFunction(target.Iface.Node, "Request", &[]reflect.Value{
						reflect.ValueOf(target),
						reflect.ValueOf(port.Iface),
					})

					// fmt.Printf("1. %s -> %s (%s)\n", port.Name, target.Name, target.Value)

					port.Iface.QRequesting = false

					if port.Feature == portTypes.TypeArrayOf {
						var tempVal interface{}
						if target.Value == nil {
							tempVal = target.Default
						} else {
							tempVal = target.Value
						}

						return [](interface{}){tempVal}
					}
				}

				// Return multiple data as an array
				data := []interface{}{}
				for _, cable := range port.Cables {
					var target *Port
					if cable.Owner == port {
						target = cable.Target
					} else {
						target = cable.Owner
					}

					utils.CallFunction(target.Iface.Node, "Request", &[]reflect.Value{
						reflect.ValueOf(target),
						reflect.ValueOf(port.Iface),
					})

					// fmt.Printf("2. %s -> %s (%s)\n", port.Name, target.Name, target.Value)

					if target.Value == nil {
						data = append(data, target.Default)
					} else {
						data = append(data, target.Value)
					}
				}

				port.Iface.QRequesting = false
				if port.Feature != portTypes.TypeArrayOf {
					return data[0]
				}

				return data
			}

			if port.Feature == portTypes.TypeArrayOf {
				var tempVal interface{}
				if port.Value == nil {
					tempVal = port.Default
				} else {
					tempVal = port.Value
				}

				return [](interface{}){tempVal}
			}

			if port.Value == nil {
				return port.Default
			}

			return port.Value
		}
		// else setter (only for output port)

		if port.Source == portTypes.Input {
			panic("Can't set data to input port")
		}

		// ToDo: do we need feature validation here?

		_val := val[0]
		// fmt.Printf("3. %s = %s\n", port.Name, _val)

		port.Value = _val
		port.QTrigger("value", port)
		port.sync()

		return _val
	}
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

		target.QTrigger("value", port)
	}
}
