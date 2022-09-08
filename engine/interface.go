package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

var portList = [3]string{"Input", "Output", "Property"}

type InterfaceData map[string]getterSetter
type Interface struct {
	*CustomEvent

	Id        string
	I         int // index
	Title     string
	Namespace string

	// Property map[string]*Port
	Output map[string]*Port
	Input  map[string]*Port
	Data   InterfaceData
	Node   any // any = extends *engine.Node

	Ref     *ReferencesShortcut
	IsGhost bool

	Importing bool

	// for internal library use only
	QInitialized bool
	QRequesting  bool
	QFuncMain    *NodesFunctionMain
	QDynamicPort bool
	QEnum        int
	QBpVarRef    *BPVariable
}

// To be overriden
func (iface *Interface) Init()             {}
func (iface *Interface) Destroy()          {}
func (iface *Interface) Imported(data any) {}

// Internal blackprint function node initialization
func (iface *Interface) QBpFnInit() {}

var reflectKind = reflect.TypeOf(reflect.Int)

// Private (to be called for internal library only)
func (iface *Interface) QPrepare() {
	iface.CustomEvent = &CustomEvent{}
	ref := &ReferencesShortcut{}

	node := iface.Node
	utils.SetProperty(node, "Ref", ref)
	iface.Ref = ref

	utils.SetProperty(node, "Route", &RoutePort{Iface: iface})

	for i := 0; i < 3; i++ {
		which := portList[i]
		port := *utils.GetPropertyRef(node, "T"+which).(*map[string]any) // get value by property name

		if port == nil {
			continue
		}

		ifacePort := map[string]*Port{}
		utils.SetProperty(iface, which, ifacePort)

		var inputUpgradePort map[string]*PortInputGetterSetter
		var outputUpgradePort map[string]*PortOutputGetterSetter

		if which == "Input" {
			inputUpgradePort = map[string]*PortInputGetterSetter{}
			ref.Input = inputUpgradePort
			ref.IInput = ifacePort

			utils.SetProperty(node, which, inputUpgradePort)
		} else {
			outputUpgradePort = map[string]*PortOutputGetterSetter{}
			ref.Output = outputUpgradePort
			ref.IOutput = ifacePort

			utils.SetProperty(node, which, outputUpgradePort)
		}

		// name: string
		for name, config_ := range port {
			linkedPort := iface.QCreatePort(which, name, config_)
			ifacePort[name] = linkedPort

			// CreateLinker()
			if which == "Input" {
				inputUpgradePort[name] = &PortInputGetterSetter{port: linkedPort}
			} else {
				outputUpgradePort[name] = &PortOutputGetterSetter{port: linkedPort}
			}
		}
	}
}

func (iface *Interface) QCreatePort(which string, name string, config_ any) *Port {
	var config *PortFeature
	var type_ reflect.Kind
	var types_ []reflect.Kind
	var feature int

	var def any
	var qfunc func(*Port)
	if reflect.TypeOf(config_) == reflectKind {
		type_ = config_.(reflect.Kind)

		if type_ == types.Int {
			def = 0
		} else if type_ == types.Bool {
			def = false
		} else if type_ == types.String {
			def = ""
		} else if type_ == types.Array {
			def = [0]any{} // ToDo: is this actually working?
		} else if type_ == types.Any { // Any
			// pass
		} else if type_ == types.Function {
			// pass
		} else if type_ == types.Route {
			// pass
		} else {
			panic(iface.Namespace + ": '" + name + "' Port type(" + type_.String() + ") for initialization was not recognized")
		}
	} else {
		config = config_.(*PortFeature)
		type_ = config.Type
		feature = config.Id

		if feature == PortTypeTrigger {
			qfunc = config.Func
			type_ = types.Function
		} else if feature == PortTypeArrayOf {
			if type_ != types.Any {
				def = &[]any{}
			}
		} else if feature == PortTypeUnion {
			types_ = config.Types
		} else if feature == PortTypeDefault {
			def = config.Value
		} else {
			// panic(iface.Namespace + ": '" + name + "' Port feature(" + strconv.Itoa(feature) + ") for initialization was not recognized")
		}
	}

	var source int
	if which == "Input" {
		source = PortInput
	} else if which == "Output" {
		source = PortOutput
	}
	// else if which == "Property" {
	// 	source = PortProperty
	// }

	port := &Port{
		Name:     name,
		Type:     type_,
		Types:    types_,
		Default:  def,
		QFunc:    qfunc,
		Source:   source,
		Iface:    iface,
		Feature:  feature,
		QFeature: config,
	}

	return port
}

func (iface *Interface) QInitPortSwitches(portSwitches map[string]int) {
	for key, val := range portSwitches {
		if (val | 1) == 1 {
			portStructOf_split(iface.Output[key])
		}

		if (val | 2) == 2 {
			iface.Output[key].AllowResync = true
		}
	}
}

// Load saved port data value
func (iface *Interface) QImportInputs(ports map[string]*Port) {
	for key, val := range ports {
		iface.Input[key].Default = val
	}
}
