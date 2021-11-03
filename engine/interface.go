package engine
import (
	"reflect"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/port"
)

var portList = [3]string {"input", "output", "property"}

type InterfaceData map[string] func(interface {}) interface{}
type Interface struct {
	customEvent

	Id string
	I int // index
	Title string
	// Interface string
	Namespace string

	Output map[string] Port
	Input map[string] Port
	Property map[string] Port
	Data InterfaceData
	Node interface{} // interface = extends 'Node'

	QRequesting bool // private (to be used for internal library only)
	Importing bool
}

// To be overriden
func (iface *Interface) Init() {
	return nil
}

func setProperty(obj interface{}, key string, value interface{}) {
	reflect.ValueOf(obj).Elem().FieldByName(key).Set(reflect.ValueOf(value))
}

// Private (to be called for internal library only)
func (iface *Interface) QPrepare() {
	node := iface.Node

	for i := 0; i < 3; i++ {
		which := portList[i]
		port := node[which]

		if port == nil {
			continue
		}

		ifacePort := map[string] Port{}
		setProperty(iface, which, ifacePort)

		// name: string, config: PortFeature
		for name, config := range port {
			var def interface{}

			type_ := config.Type
			feature := config.Feature

			if feature == port.PortFeatureConst.Trigger {
				def = config.Func
				type_ = types.Function
			} else if feature == port.PortFeatureConst.ArrayOf {
				// pass
			} else if type_ == types.Int {
				def = 0
			} else if type_ == types.Bool {
				def = false
			} else if type_ == types.String {
				def = ""
			} else if type_ == types.Array {
				def = [0]interface{} {} // ToDo: is this actually working?
			} else if type_ == nil { // Any
				// pass
			} else if type_ == types.Func {
				// pass
			} else if feature == nil && type_ == nil {
				panic("Port for initialization must be a types")
			}

			linkedPort := Port{
				Name: name,
				Type: type_,
				Default: def,
				Source: which,
				Iface: iface,
				Feature: feature,
			}

			setProperty(iface, ifacePort, linkedPort.CreateLinker())
		}
	}

	return nil
}