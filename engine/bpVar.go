package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/utils"
)

type bpVarValue struct {
	*CustomEvent
	val any
}

func (b *bpVarValue) Get() any {
	return b.val
}

func (b *bpVarValue) Set(val any) {
	b.val = val
	b.Emit("Value", nil)
}

// used for instance.CreateVariable
type BPVariable struct {
	*CustomEvent
	Id       string
	Title    string
	Type     reflect.Kind
	Used     []*Interface
	Value    bpVarValue
	Listener []*Interface
}

func (b *BPVariable) Destroy() {
	for _, iface := range b.Used {
		ins := utils.GetProperty((utils.GetProperty(iface, "Node")), "Instance").(*Instance)
		utils.CallFunction(ins, "DeleteNode", &[]reflect.Value{
			reflect.ValueOf(iface),
		})
	}

	b.Used = b.Used[:0]
}
