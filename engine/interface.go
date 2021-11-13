package engine

import (
	"reflect"

	portType "github.com/blackprint/engine-go/port"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

var portList = [3]string{"Input", "Output", "Property"}

type InterfaceData map[string]portType.GetterSetter
type Interface struct {
	*customEvent
	QInitialized bool // for internal library only

	Id    string
	I     int // index
	Title string
	// Interface string
	Namespace string

	Output   map[string]*Port
	Input    map[string]*Port
	Property map[string]*Port
	Data     InterfaceData
	Node     interface{} // interface = extends 'Node' data type

	QRequesting bool // private (to be used for internal library only)
	Importing   bool
}

// To be overriden
func (iface *Interface) Init() {}

var reflectKind = reflect.TypeOf(reflect.Int)

// Private (to be called for internal library only)
func (iface *Interface) QPrepare() {
	iface.customEvent = &customEvent{}

	node := iface.Node

	for i := 0; i < 3; i++ {
		which := portList[i]
		port := utils.GetProperty(node, "T"+which).(map[string]interface{}) // get value by property name

		if port == nil {
			continue
		}

		ifacePort := map[string]*Port{}
		utils.SetProperty(iface, which, ifacePort)

		upgradePort := map[string]portType.GetterSetter{}

		// name: string, config: PortFeature
		for name, config_ := range port {
			var config portType.PortFeature
			var type_ reflect.Kind
			var feature int

			var def interface{}
			if reflect.TypeOf(config_) == reflectKind {
				type_ = config_.(reflect.Kind)

				if type_ == types.Int {
					def = 0
				} else if type_ == types.Bool {
					def = false
				} else if type_ == types.String {
					def = ""
				} else if type_ == types.Array {
					def = [0]interface{}{} // ToDo: is this actually working?
				} else if type_ == types.Any { // Any
					// pass
				} else if type_ == types.Function {
					// pass
				} else {
					panic(iface.Namespace + ": '" + name + "' Port type(" + type_.String() + ") for initialization was not recognized")
				}
			} else {
				config = config_.(portType.PortFeature)
				type_ = config.Type
				feature = config.Feature

				if feature == portType.TypeTrigger {
					def = config.Func
					type_ = types.Function
				} else if feature == portType.TypeArrayOf {
					// pass
				} else {
					// panic(iface.Namespace + ": '" + name + "' Port feature(" + strconv.Itoa(feature) + ") for initialization was not recognized")
				}
			}

			var source int
			if which == "Input" {
				source = portType.Input
			} else if which == "Output" {
				source = portType.Output
			} else if which == "Property" {
				source = portType.Property
			}

			linkedPort := Port{
				Name:    name,
				Type:    type_,
				Default: def,
				Source:  source,
				Iface:   iface,
				Feature: feature,
			}

			ifacePort[name] = &linkedPort
			upgradePort[name] = linkedPort.CreateLinker()
		}

		utils.SetProperty(node, which, upgradePort)
	}
}
