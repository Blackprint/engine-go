package engine

import (
	"reflect"
)

type Cable struct {
	Type   reflect.Kind
	Owner  *Port
	Target *Port
}

type CableEvent struct {
	Cable  *Cable
	Port   *Port
	Target *Port
}

func NewCable(owner *Port, target *Port) Cable {
	return Cable{
		Type:   owner.Type,
		Owner:  owner,
		Target: target,
	}
}

func (c *Cable) QConnected() {
	c.Owner.QTrigger("cable.connect", CableEvent{
		Cable:  c,
		Port:   c.Owner,
		Target: c.Target,
	})

	c.Target.QTrigger("cable.connect", CableEvent{
		Cable:  c,
		Port:   c.Target,
		Target: c.Owner,
	})

	var inp, out *Port
	if c.Owner.Source == "input" {
		inp = c.Target
		out = c.Owner
	} else {
		inp = c.Owner
		out = c.Target
	}

	if out.Value != nil {
		inp.QTrigger("value", out)
	}
}

// For debugging
func (c *Cable) String() string {
	return "\nCable: " + c.Owner.Iface.Title + "." + c.Owner.Name + " . " + c.Target.Name + "." + c.Target.Iface.Title
}
