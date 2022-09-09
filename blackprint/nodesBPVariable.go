package blackprint

import (
	"strconv"

	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

type bpVarSet struct {
	*engine.Node
}
type bpVarGet struct {
	*engine.Node
}

func (b *bpVarSet) Update(c *engine.Cable) {
	b.Iface.QBpVarRef.Value.Set(b.Input["Val"].Get())
}

type bpVarGetSet struct {
	*engine.Interface
	QOnChanged func(*engine.Port)
}

func (b *bpVarGetSet) Imported(data map[string]any) {
	if _, exist := data["scope"]; exist {
		panic("'scope' options is required for creating variable node")
	}

	if _, exist := data["name"]; exist {
		panic("'name' options is required for creating variable node")
	}

	b.ChangeVar(data["name"].(string), data["scope"].(int))
	b.QBpVarRef.Used = append(b.QBpVarRef.Used, b)
}

func (b *bpVarGetSet) ChangeVar(name string, scopeId int) map[string]*engine.BPVariable {
	if _, exist := b.Data["name"]; exist {
		panic("Can't change variable node that already be initialized")
	}

	b.Data["name"] = &engine.GetterSetter{Value: name}
	b.Data["scope"] = &engine.GetterSetter{Value: scopeId}

	thisInstance := b.Node.Instance.(*engine.Instance)
	funcInstance := thisInstance.QFuncMain
	if funcInstance == nil {
		funcInstance.Node.QFuncInstance
	}

	var scope map[string]*engine.BPVariable
	if scopeId == engine.VarScopePublic {
		if funcInstance != nil {
			scope := funcInstance.Node.RootInstance.Variables
		} else {
			scope := thisInstance.Variables
		}
	} else if scopeId == engine.VarScopeShared {
		scope := funcInstance.Variables
	} else { // private
		scope := thisInstance.Variables
	}

	if _, exist := scope[name]; !exist {
		var scopeName string
		if scopeId == engine.VarScopePublic {
			scopeName = "public"
		} else if scopeId == engine.VarScopePrivate {
			scopeName = "private"
		} else if scopeId == engine.VarScopeShared {
			scopeName = "shared"
		} else {
			scopeName = "unknown"
		}

		panic("'" + name + "' variable was not defined on the '" + scopeName + " (scopeId: " + strconv.Itoa(scopeId) + ")' instance")
	}

	return scope
}

func (b *bpVarGetSet) UseType(port *engine.Port) bool {
	if b.QBpVarRef.Type != 0 { // Type was set
		if port == nil {
			b.QBpVarRef.Type = 0 // Type not set
		}
		return true
	}

	if port == nil {
		panic("Can't set type with null")
	}

	return false
}

func (b *bpVarGetSet) UseType_(port *engine.Port, targetPort *engine.Port) {
	b.QBpVarRef.Type = port.Type
	targetPort.ConnectPort(port)

	// Also create port for other node that using $this variable
	for _, item := range b.QBpVarRef.Used {
		item.QReinitPort()
	}
}

func (b *bpVarGetSet) Destroy() {
	temp := b.QBpVarRef
	if temp == nil {
		return
	}

	temp.Used = utils.RemoveItemAny(temp.Used, b)

	listener := b.QBpVarRef.Listener
	if listener == nil {
		return
	}

	b.QBpVarRef.Listener = utils.RemoveItemAny(listener, b)
}

type iVarSet struct {
	*bpVarGetSet
	QEventListen string
}

func (b *iVarSet) UseType(port *engine.Port) {
	if !b.bpVarGetSet.UseType(port) {
		b.bpVarGetSet.UseType_(port, b.QReinitPort())
	}
}

func (b *iVarSet) ChangeVar(name string, scopeId int) {
	if _, exist := b.Data["name"]; exist {
		panic("Can't change variable node that already be initialized")
	}

	if b.QOnChanged != nil && b.QBpVarRef != nil {
		b.QBpVarRef.Off("value", b.QOnChanged)
	}

	scope := b.bpVarGetSet.ChangeVar(name, scopeId)
	b.Title = "Get " + name

	temp := scope[b.Data["name"].Get().(string)]
	b.QBpVarRef = temp
	if temp.Type == 0 { // Type not set
		return
	}

	b.QReinitPort()
}

func (b *iVarSet) QReinitPort() *engine.Port {
	temp := b.QBpVarRef
	node := b.Node

	if b.Output["Val"] != nil {
		node.DeletePort("Val")
	}

	ref := b.Node.Output.(map[string]*engine.PortOutputGetterSetter)
	b.Node.CreatePort("Val", temp.Type)

	if temp.Type == types.Function {
		b.QEventListen = "call"
		b.QOnChanged = func(p *engine.Port) {
			ref["Val"].Call()
		}
	} else {
		b.QEventListen = "value"
		b.QOnChanged = func(p *engine.Port) {
			ref["Val"].Set(temp.Value.Get())
		}
	}

	temp.On(b.QEventListen, b.QOnChanged)
	return b.Output["Val"]
}

func (b *iVarSet) Destroy() {
	if b.QEventListen != "" {
		b.QBpVarRef.Off(b.QEventListen, b.QOnChanged)
	}

	b.bpVarGetSet.Destroy()
}

type iVarGet struct {
	*bpVarGetSet
}

func (b *iVarGet) UseType(port *engine.Port) {
	if !b.bpVarGetSet.UseType(port) {
		b.bpVarGetSet.UseType_(port, b.QReinitPort())
	}
}

func (b *iVarGet) ChangeVar(name string, scopeId int) {
	scope := b.bpVarGetSet.ChangeVar(name, scopeId)
	b.Title = "Set " + name

	temp := scope[b.Data["name"].Get().(string)]
	b.QBpVarRef = temp
	if temp.Type == 0 { // Type not set
		return
	}

	b.QReinitPort()
}

func (b *iVarGet) QReinitPort() *engine.Port {
	input := b.Input
	node := b.Node
	temp := b.QBpVarRef

	if _, exist := input["Val"]; exist {
		node.DeletePort("Input", "Val")
	}

	if temp.Type == types.Function {
		node.CreatePort("Input", "Val", engine.Ports.Trigger(func(p *engine.Port) {
			temp.Emit("call", nil)
		}))
	} else {
		node.CreatePort("Input", "Val", temp.Type)
	}

	return b.Input["Val"]
}

func init() {
	RegisterNode("BP/Var/Set", func(i *engine.Instance) any {
		node := &bpVarSet{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Var/Set").(*iVarSet)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name":  &engine.GetterSetter{Value: ""},
			"scope": &engine.GetterSetter{Value: engine.VarScopePublic},
		}

		iface.Title = "VarSet"
		iface.Type = "bp-var-set"
		iface.QEnum = nodes.BPVarSet
		iface.QDynamicPort = true

		return node
	})

	RegisterInterface("BPIC/BP/Var/Get", func(node any) any {
		return &iVarGet{
			bpVarGetSet: &bpVarGetSet{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})

	RegisterNode("BP/Var/Get", func(i *engine.Instance) any {
		node := &bpVarGet{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Var/Get").(*iVarGet)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name":  &engine.GetterSetter{Value: ""},
			"scope": &engine.GetterSetter{Value: engine.VarScopePublic},
		}

		iface.Title = "VarGet"
		iface.Type = "bp-var-get"
		iface.QEnum = nodes.BPVarGet
		iface.QDynamicPort = true

		return node
	})

	RegisterInterface("BPIC/BP/Var/Get", func(node any) any {
		return &iVarGet{
			bpVarGetSet: &bpVarGetSet{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})
}
