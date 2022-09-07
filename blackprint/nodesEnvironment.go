package blackprint

import (
	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/engine/nodes"
)

type bpEnvGet struct {
	*engine.Node
}

type bpEnvSet struct {
	*engine.Node
}

func (b *bpEnvSet) Update(c *engine.Cable) {
	Environment.Set(b.Iface.Data["name"].Get().(string), c.GetValue().(string))
}

type bpEnvGetSet struct {
	*engine.Interface
}

func (b *bpEnvGetSet) Imported(data map[string]any) {
	if data["name"] == nil {
		panic("Parameter 'name' is required")
	}

	b.Data["name"].Set(data["name"])
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
		if ev.Key != b.Data["name"].Get().(string) {
			return
		}

		b.Ref.Output["Val"].Set(ev.Value)
	}

	Event.On("environment.changed environment.added", b.QListener)
	b.Ref.Output["Val"].Set(Environment.Map[b.Data["name"].Get().(string)])
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

func registerEnvNode() {
	RegisterNode("BP/Env/Get", func(i *engine.Instance) any {
		node := &bpEnvGet{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Env/Get").(*iEnvGet)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name": &engine.GetterSetter{Value: ""},
		}

		iface.Title = "EnvGet"
		iface.Type = "bp-env-get"
		iface.QEnum = nodes.BPEnvGet

		return node
	})

	RegisterInterface("BPIC/BP/Env/Get", func(node any) any {
		return &iEnvGet{
			bpEnvGetSet: &bpEnvGetSet{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})

	RegisterNode("BP/Env/Set", func(i *engine.Instance) any {
		node := &bpEnvSet{
			Node: &engine.Node{
				Instance: i,
			},
		}

		iface := node.SetInterface("BPIC/BP/Env/Set").(*iEnvSet)

		// Specify data field from here to make it enumerable and exportable
		iface.Data = engine.InterfaceData{
			"name": &engine.GetterSetter{Value: ""},
		}

		iface.Title = "EnvSet"
		iface.Type = "bp-env-set"
		iface.QEnum = nodes.BPEnvSet

		return node
	})

	RegisterInterface("BPIC/BP/Env/Set", func(node any) any {
		return &iEnvSet{
			bpEnvGetSet: &bpEnvGetSet{
				Interface: &engine.Interface{
					Node: node,
				},
			},
		}
	})
}
