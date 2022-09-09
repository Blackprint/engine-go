package engine

import (
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
)

type bpEnvGet struct {
	*EmbedNode
}

type bpEnvSet struct {
	*EmbedNode
}

func (b *bpEnvSet) Update(c *Cable) {
	QEnvironment.Set(b.Iface.Data["name"].Get().(string), c.GetValue().(string))
}

type bpEnvGetSet struct {
	*EmbedInterface
	Type string
}

func (b *bpEnvGetSet) Imported(data map[string]any) {
	if data["name"] == nil {
		panic("Parameter 'name' is required")
	}

	b.Iface.Data["name"].Set(data["name"])
	name := data["name"].(string)

	if _, exists := QEnvironment.Map[name]; !exists {
		QEnvironment.Set(name, "")
	}
}

type iEnvGet struct {
	*bpEnvGetSet
	_listener func(any)
}

func (b *iEnvGet) Imported(data map[string]any) {
	b.bpEnvGetSet.Imported(data)

	b._listener = func(v any) {
		ev := v.(*EnvironmentEvent)
		if ev.Key != b.Iface.Data["name"].Get().(string) {
			return
		}

		b.Ref.Output["Val"].Set(ev.Value)
	}

	Event.On("environment.changed environment.added", b._listener)
	b.Ref.Output["Val"].Set(QEnvironment.Map[b.Iface.Data["name"].Get().(string)])
}

func (b *iEnvGet) Destroy() {
	if b._listener == nil {
		return
	}

	Event.Off("environment.changed environment.added", b._listener)
}

type iEnvSet struct {
	*bpEnvGetSet
}

func init() {
	QNodeList["BP/Env/Get"] = &NodeRegister{
		Output: PortTemplate{
			"Val": types.String,
		},
		Constructor: func(node *Node) {
			node.Embed = &bpEnvGet{}

			iface := node.SetInterface("BPIC/BP/Env/Get")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "EnvGet"
			iface.Embed.(*iEnvGet).Type = "bp-env-get"
			iface._enum = nodes.BPEnvGet
		},
	}

	QInterfaceList["BPIC/BP/Env/Get"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &iEnvGet{
				bpEnvGetSet: &bpEnvGetSet{},
			}
		},
	}

	QNodeList["BP/Env/Set"] = &NodeRegister{
		Input: PortTemplate{
			"Val": types.String,
		},
		Constructor: func(node *Node) {
			node.Embed = &bpEnvSet{}

			iface := node.SetInterface("BPIC/BP/Env/Set")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "EnvSet"
			iface.Embed.(*iEnvSet).Type = "bp-env-set"
			iface._enum = nodes.BPEnvSet
		},
	}

	QInterfaceList["BPIC/BP/Env/Set"] = &InterfaceRegister{
		Constructor: func(iface *Interface) {
			iface.Embed = &iEnvSet{
				bpEnvGetSet: &bpEnvGetSet{},
			}
		},
	}
}
