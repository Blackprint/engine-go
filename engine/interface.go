package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

var portList = [3]string{"Input", "Output", "Property"}

type embedInterface interface {
	Init()
	Request(*Cable)
	Update(*Cable)
	Imported(map[string]any)
	Destroy()
	SyncIn(id string, data ...any)
}

type EmbedInterface struct {
	embedInterface
	Node  *Node
	Iface *Interface
	Ref   *referencesShortcut
}

// To be overriden by module developer
func (iface *EmbedInterface) Init()                        {}
func (iface *EmbedInterface) Destroy()                     {}
func (iface *EmbedInterface) Imported(data map[string]any) {}

type InterfaceData map[string]getterSetter
type Interface struct {
	*CustomEvent

	Id        string
	uniqId    int // For bpFunction only
	I         int // index
	Title     string
	Namespace string

	// Property map[string]*Port
	Output map[string]*Port
	Input  map[string]*Port
	Data   InterfaceData
	Node   *Node
	Embed  embedInterface

	Ref     *referencesShortcut
	IsGhost bool

	Importing bool

	// for internal library use only
	_initialized bool
	_requesting  bool
	_funcMain    *Interface
	_dynamicPort bool
	_enum        int
	_bpVarRef    *BPVariable
	_proxyInput  *Node
	_proxyOutput *Node
	_parentFunc  *Interface
	_bpInstance  *Instance
	_bpDestroy   bool
}

// To be overriden
func (i *Interface) Init()                        { i.Embed.Init() }
func (i *Interface) Destroy()                     { i.Embed.Destroy() }
func (i *Interface) Imported(data map[string]any) { i.Embed.Imported(data) }

// Internal blackprint function node initialization
func (iface *Interface) _bpFnInit() {}

var reflectKind = reflect.TypeOf(reflect.Int)

// Private (to be called for internal library only)
func (iface *Interface) _prepare(meta *NodeRegister) {
	iface.CustomEvent = &CustomEvent{}
	ref := &referencesShortcut{}

	node := iface.Node
	node.Ref = ref
	iface.Ref = ref

	node.Routes = &routePort{Iface: iface}

	for i := 0; i < 3; i++ {
		which := portList[i]
		port := utils.GetProperty(meta, which).(PortTemplate) // get value by property name

		if port == nil {
			continue
		}

		ifacePort := map[string]*Port{}

		var inputUpgradePort map[string]*portInputGetterSetter
		var outputUpgradePort map[string]*portOutputGetterSetter

		if which == "Input" {
			inputUpgradePort = map[string]*portInputGetterSetter{}
			ref.Input = inputUpgradePort
			ref.IInput = ifacePort

			iface.Input = ifacePort
			node.Input = inputUpgradePort
		} else {
			outputUpgradePort = map[string]*portOutputGetterSetter{}
			ref.Output = outputUpgradePort
			ref.IOutput = ifacePort

			iface.Output = ifacePort
			node.Output = outputUpgradePort
		}

		// name: string
		for name, config_ := range port {
			linkedPort := iface._createPort(which, name, config_)
			ifacePort[name] = linkedPort

			// CreateLinker()
			if which == "Input" {
				inputUpgradePort[name] = &portInputGetterSetter{port: linkedPort}
			} else {
				outputUpgradePort[name] = &portOutputGetterSetter{port: linkedPort}
			}
		}
	}
}

func (iface *Interface) _createPort(which string, name string, config_ any) *Port {
	var config *portFeature
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
		config = config_.(*portFeature)
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
		_func:    qfunc,
		Source:   source,
		Iface:    iface,
		Feature:  feature,
		_feature: config,
	}

	return port
}

func (iface *Interface) _initPortSwitches(portSwitches map[string]int) {
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
func (iface *Interface) _importInputs(ports map[string]any) {
	for key, val := range ports {
		iface.Input[key].Default = val
	}
}
