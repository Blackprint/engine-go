package engine

import (
	"fmt"
)

type Cable struct {
	Type   string
	Owner  *Port
	Target *Port
}

type CableEvent struct {
	Cable  Cable
	Port   *Port
	Target *Port
}

func NewCable(owner *Port, target *Port) *Cable {
	return &Cable{
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

	if c.Owner.Source == "input" {
		inp := c.Target
		out := c.Owner
	} else {
		inp := c.Owner
		out := c.Target
	}

	if out.value != nil {
		inp.QTrigger("value", out)
	}
}

// For debugging
func (c *Cable) QPrint() {
	fmt.Println("\nCable: " + c.owner.iface.title + "." + c.owner.name + " . " + c.target.name + "." + c.target.iface.title)
}
