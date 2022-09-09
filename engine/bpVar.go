package engine

import (
	"reflect"
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
	b.Emit("value", nil)
}

// used for instance.CreateVariable
type BPVariable struct {
	*CustomEvent
	Id           string
	Title        string
	Type         reflect.Kind
	Used         []*Interface
	Value        bpVarValue
	Listener     []*Interface
	FuncInstance *BPFunction // for shared function variables
}

func (b *BPVariable) Destroy() {
	for _, iface := range b.Used {
		ins := iface.Node.Instance
		ins.DeleteNode(iface)
	}

	b.Used = b.Used[:0]
}
