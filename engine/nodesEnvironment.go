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
	QListener func(any)
}

func (b *iEnvGet) Imported(data map[string]any) {
	b.bpEnvGetSet.Imported(data)

	b.QListener = func(v any) {
		ev := v.(*EnvironmentEvent)
		if ev.Key != b.Iface.Data["name"].Get().(string) {
			return
		}

		b.Ref.Output["Val"].Set(ev.Value)
	}

	Event.On("environment.changed environment.added", b.QListener)
	b.Ref.Output["Val"].Set(QEnvironment.Map[b.Iface.Data["name"].Get().(string)])
}

func (b *iEnvGet) Destroy() {
	if b.QListener == nil {
		return
	}

	Event.Off("environment.changed environment.added", b.QListener)
}

type iEnvSet struct {
	*bpEnvGetSet
}

func init() {
	QNodeList["BP/Env/Get"] = &QNodeRegister{
		Metadata: &NodeMetadata{
			Output: NodePortTemplate{
				"Val": types.String,
			},
		},
		Constructor: func(i *Instance) *Node {
			node := &Node{
				Instance: i,
				Embed:    &bpEnvGet{},
			}

			iface := node.SetInterface("BPIC/BP/Env/Get")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "EnvGet"
			iface.Embed.(*iEnvGet).Type = "bp-env-get"
			iface.QEnum = nodes.BPEnvGet

			return node
		},
	}

	QInterfaceList["BPIC/BP/Env/Get"] = func(node *Node) *Interface {
		return &Interface{
			Embed: &iEnvGet{
				bpEnvGetSet: &bpEnvGetSet{},
			},
		}
	}

	QNodeList["BP/Env/Set"] = &QNodeRegister{
		Metadata: &NodeMetadata{
			Input: NodePortTemplate{
				"Val": types.String,
			},
		},
		Constructor: func(i *Instance) *Node {
			node := &Node{
				Instance: i,
				Embed:    &bpEnvSet{},
			}

			iface := node.SetInterface("BPIC/BP/Env/Set")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = InterfaceData{
				"name": &GetterSetter{Value: ""},
			}

			iface.Title = "EnvSet"
			iface.Embed.(*iEnvSet).Type = "bp-env-set"
			iface.QEnum = nodes.BPEnvSet

			return node
		},
	}

	QInterfaceList["BPIC/BP/Env/Set"] = func(node *Node) *Interface {
		return &Interface{
			Embed: &iEnvSet{
				bpEnvGetSet: &bpEnvGetSet{},
			},
		}
	}
}
