package engine

import (
	"strconv"

	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
	"github.com/blackprint/engine-go/utils"
)

type bpVarSet struct {
	*EmbedNode
}
type bpVarGet struct {
	*EmbedNode
}

func (b *bpVarSet) Update(c *Cable) {
	b.Iface._bpVarRef.Value.Set(b.Node.Input["Val"].Get())
}

type bpVarGetSet struct {
	*EmbedInterface
	bpVarGetSetIFace
	Type       string
	_bpVarRef  *BPVariable
	_onChanged func(*Port)
}

type bpVarGetSetIFace interface {
	_reinitPort() *Port
}

func (b *bpVarGetSet) Imported(data map[string]any) {
	if _, exist := data["scope"]; exist {
		panic("'scope' options is required for creating variable node")
	}

	if _, exist := data["name"]; exist {
		panic("'name' options is required for creating variable node")
	}

	b.ChangeVar(data["name"].(string), data["scope"].(int))
	b._bpVarRef.Used = append(b._bpVarRef.Used, b.Iface)
}

func (b *bpVarGetSet) ChangeVar(name string, scopeId int) map[string]*BPVariable {
	if _, exist := b.Iface.Data["name"]; exist {
		panic("Can't change variable node that already be initialized")
	}

	b.Iface.Data["name"] = &GetterSetter{Value: name}
	b.Iface.Data["scope"] = &GetterSetter{Value: scopeId}

	thisInstance := b.Node.Instance
	funcInstance := thisInstance._funcMain
	var bpFunc *bpFunction
	if funcInstance != nil {
		bpFunc = funcInstance.Node._funcInstance
	}

	var scope map[string]*BPVariable
	if scopeId == VarScopePublic {
		if funcInstance != nil {
			scope = bpFunc.RootInstance.Variables
		} else {
			scope = thisInstance.Variables
		}
	} else if scopeId == VarScopeShared {
		scope = bpFunc.Variables
	} else { // private
		scope = thisInstance.Variables
	}

	if _, exist := scope[name]; !exist {
		var scopeName string
		if scopeId == VarScopePublic {
			scopeName = "public"
		} else if scopeId == VarScopePrivate {
			scopeName = "private"
		} else if scopeId == VarScopeShared {
			scopeName = "shared"
		} else {
			scopeName = "unknown"
		}

		panic("'" + name + "' variable was not defined on the '" + scopeName + " (scopeId: " + strconv.Itoa(scopeId) + ")' instance")
	}

	return scope
}

func (b *bpVarGetSet) UseType(port *Port) bool {
	if b._bpVarRef.Type != 0 { // Type was set
		if port == nil {
			b._bpVarRef.Type = 0 // Type not set
		}
		return true
	}

	if port == nil {
		panic("Can't set type with null")
	}

	return false
}

func (b *bpVarGetSet) UseType_(port *Port, targetPort *Port) {
	b._bpVarRef.Type = port.Type
	targetPort.ConnectPort(port)

	// Also create port for other node that using $this variable
	for _, item := range b._bpVarRef.Used {
		item.Embed.(bpVarGetSetIFace)._reinitPort()
	}
}

func (b *bpVarGetSet) Destroy() {
	temp := b._bpVarRef
	if temp == nil {
		return
	}

	temp.Used = utils.RemoveItem(temp.Used, b.Iface)

	listener := b._bpVarRef.Listener
	if listener == nil {
		return
	}

	b._bpVarRef.Listener = utils.RemoveItem(listener, b.Iface)
}

type iVarSet struct {
	*bpVarGetSet
	_eventListen string
}

func (b *iVarSet) UseType(port *Port) {
	if !b.bpVarGetSet.UseType(port) {
		b.bpVarGetSet.UseType_(port, b._reinitPort())
	}
}

func (b *iVarSet) ChangeVar(name string, scopeId int) {
	if _, exist := b.Iface.Data["name"]; exist {
		panic("Can't change variable node that already be initialized")
	}

	if b._onChanged != nil && b._bpVarRef != nil {
		b._bpVarRef.Off("value", b._onChanged)
	}

	scope := b.bpVarGetSet.ChangeVar(name, scopeId)
	b.Iface.Title = "Get " + name

	temp := scope[b.Iface.Data["name"].Get().(string)]
	b._bpVarRef = temp
	if temp.Type == 0 { // Type not set
		return
	}

	b._reinitPort()
}

func (b *iVarSet) _reinitPort() *Port {
	temp := b._bpVarRef
	node := b.Node

	if b.Iface.Output["Val"] != nil {
		node.DeletePort("output", "Val")
	}

	ref := b.Node.Output
	b.Node.CreatePort("output", "Val", temp.Type)

	if temp.Type == types.Function {
		b._eventListen = "call"
		b._onChanged = func(p *Port) {
			ref["Val"].Call()
		}
	} else {
		b._eventListen = "value"
		b._onChanged = func(p *Port) {
			ref["Val"].Set(temp.Value.Get())
		}
	}

	temp.On(b._eventListen, b._onChanged)
	return b.Iface.Output["Val"]
}

func (b *iVarSet) Destroy() {
	if b._eventListen != "" {
		b._bpVarRef.Off(b._eventListen, b._onChanged)
	}

	b.bpVarGetSet.Destroy()
}

type iVarGet struct {
	*bpVarGetSet
}

func (b *iVarGet) UseType(port *Port) {
	if !b.bpVarGetSet.UseType(port) {
		b.bpVarGetSet.UseType_(port, b._reinitPort())
	}
}

func (b *iVarGet) ChangeVar(name string, scopeId int) {
	scope := b.bpVarGetSet.ChangeVar(name, scopeId)
	b.Iface.Title = "Set " + name

	temp := scope[b.Iface.Data["name"].Get().(string)]
	b._bpVarRef = temp
	if temp.Type == 0 { // Type not set
		return
	}

	b._reinitPort()
}

func (b *iVarGet) _reinitPort() *Port {
	input := b.Iface.Input
	node := b.Node
	temp := b._bpVarRef

	if _, exist := input["Val"]; exist {
		node.DeletePort("Input", "Val")
	}

	if temp.Type == types.Function {
		node.CreatePort("Input", "Val", QPorts.Trigger(func(p *Port) {
			temp.Emit("call", nil)
		}))
	} else {
		node.CreatePort("Input", "Val", temp.Type)
	}

	return input["Val"]
}

func init() {
	QNodeList["BP/Var/Set"] = &NodeRegister{
		Input: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &bpVarSet{}

			iface := node.SetInterface("BPIC/BP/Var/Set")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name":  &GetterSetter{Value: ""},
				"scope": &GetterSetter{Value: VarScopePublic},
			}

			iface.Title = "VarSet"
			iface.Embed.(*iVarSet).Type = "bp-var-set"
			iface._enum = nodes.BPVarSet
			iface._dynamicPort = true
		},
	}

	QInterfaceList["BPIC/BP/Var/Set"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &iVarSet{
				bpVarGetSet: &bpVarGetSet{},
			}
		},
	}

	QNodeList["BP/Var/Get"] = &NodeRegister{
		Output: PortTemplate{},
		Constructor: func(node *Node) {
			node.Embed = &bpVarGet{}

			iface := node.SetInterface("BPIC/BP/Var/Get")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name":  &GetterSetter{Value: ""},
				"scope": &GetterSetter{Value: VarScopePublic},
			}

			iface.Title = "VarGet"
			iface.Embed.(*iVarGet).Type = "bp-var-get"
			iface._enum = nodes.BPVarGet
			iface._dynamicPort = true
		},
	}

	QInterfaceList["BPIC/BP/Var/Get"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &iVarGet{
				bpVarGetSet: &bpVarGetSet{},
			}
		},
	}
}
