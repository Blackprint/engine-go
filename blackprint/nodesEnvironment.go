package blackprint

import (
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/engine/nodes"
	"github.com/blackprint/engine-go/types"
)

type bpEnvGet struct {
	*engine.EmbedNode
}

type bpEnvSet struct {
	*engine.EmbedNode
}

func (b *bpEnvSet) Update(c *engine.Cable) {
	Environment.Set(b.Iface.Data["name"].Get().(string), c.GetValue().(string))
}

type bpEnvGetSet struct {
	*engine.EmbedInterface
	Type string
}

func (b *bpEnvGetSet) Imported(data map[string]any) {
	if data["name"] == nil {
		panic("Parameter 'name' is required")
	}

	b.Iface.Data["name"].Set(data["name"])
	name := data["name"].(string)

	if _, exists := Environment.Map[name]; !exists {
		Environment.Set(name, "")
	}
}

type iEnvGet struct {
	*bpEnvGetSet
	QListener func(any)
}

func (b *iEnvGet) Imported(data map[string]any) {
	b.bpEnvGetSet.Imported(data)

	b.QListener = func(v any) {
		ev := v.(*engine.EnvironmentEvent)
		if ev.Key != b.Iface.Data["name"].Get().(string) {
			return
		}

		b.Ref.Output["Val"].Set(ev.Value)
	}

	Event.On("environment.changed environment.added", b.QListener)
	b.Ref.Output["Val"].Set(Environment.Map[b.Iface.Data["name"].Get().(string)])
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
	RegisterNode("BP/Env/Get", &engine.NodeMetadata{
		Output: engine.NodePortTemplate{
			"Val": types.String,
		},
	},
		func(i *engine.Instance) *engine.Node {
			node := &engine.Node{
				Instance: i,
				Embed:    &bpEnvGet{},
			}

			iface := node.SetInterface("BPIC/BP/Env/Get")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = engine.InterfaceData{
				"name": &engine.GetterSetter{Value: ""},
			}

			iface.Title = "EnvGet"
			iface.Embed.(*iEnvGet).Type = "bp-env-get"
			iface.QEnum = nodes.BPEnvGet

			return node
		})

	RegisterInterface("BPIC/BP/Env/Get",
		func(node *engine.Node) *engine.Interface {
			return &engine.Interface{
				Node: node,
				Embed: &iEnvGet{
					bpEnvGetSet: &bpEnvGetSet{},
				},
			}
		})

	RegisterNode("BP/Env/Set", &engine.NodeMetadata{
		Input: engine.NodePortTemplate{
			"Val": types.String,
		},
	},
		func(i *engine.Instance) *engine.Node {
			node := &engine.Node{
				Instance: i,
				Embed:    &bpEnvSet{},
			}

			iface := node.SetInterface("BPIC/BP/Env/Set")

			// Specify data field from here to make it enumerable and exportable
			iface.Data = engine.InterfaceData{
				"name": &engine.GetterSetter{Value: ""},
			}

			iface.Title = "EnvSet"
			iface.Embed.(*iEnvSet).Type = "bp-env-set"
			iface.QEnum = nodes.BPEnvSet

			return node
		})

	RegisterInterface("BPIC/BP/Env/Set",
		func(node *engine.Node) *engine.Interface {
			return &engine.Interface{
				Node: node,
				Embed: &iEnvSet{
					bpEnvGetSet: &bpEnvGetSet{},
				},
			}
		})
}
