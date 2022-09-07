package engine

import (
	"reflect"

	"github.com/blackprint/engine-go/utils"
)

type Cable struct {
	Type            reflect.Kind
	Owner           *Port
	Target          *Port
	Input           *Port
	Output          *Port
	Disabled        bool
	IsRoute         bool
	Connected       bool
	QEvDisconnected bool
}

type CableEvent struct {
	Cable  *Cable
	Port   *Port
	Target *Port
}

type PortSelfEvent struct {
	Port *Port
}

func NewCable(owner *Port, target *Port) *Cable {
	var input *Port
	var output *Port

	if owner.Source == PortInput {
		input = owner
		output = target
	} else {
		output = owner
		input = target
	}

	return &Cable{
		Type:   owner.Type,
		Owner:  owner,
		Target: target,
		Input:  input,
		Output: output,
	}
}

func (c *Cable) QConnected() {
	c.Connected = true

	ownerEv := &CableEvent{
		Cable:  c,
		Port:   c.Owner,
		Target: c.Target,
	}
	c.Owner.Emit("cable.connect", ownerEv)
	c.Owner.Iface.Emit("cable.connect", ownerEv)

	targetEv := &CableEvent{
		Cable:  c,
		Port:   c.Target,
		Target: c.Owner,
	}
	c.Target.Emit("cable.connect", targetEv)
	c.Target.Iface.Emit("cable.connect", targetEv)

	if c.Output.Value == nil {
		return
	}

	inputEv := &PortSelfEvent{
		Port: c.Output,
	}
	c.Input.Emit("value", inputEv)
	c.Input.Iface.Emit("value", inputEv)

	utils.CallFunction(c.Input.Iface.Node, "Update", &[]reflect.Value{
		reflect.ValueOf(c),
	})
}

// For debugging
func (c *Cable) String() string {
	return "\nCable: " + c.Output.Iface.Title + "." + c.Output.Name + " <=> " + c.Input.Name + "." + c.Input.Iface.Title
}

func (c *Cable) GetValue() any {
	return c.Output.Value
}

func (c *Cable) Disconnect(which_ ...*Port) { // which = port
	if c.IsRoute { // ToDo: simplify, use 'which' instead of check all
		if c.Output.Cables != nil {
			c.Output.Cables = c.Output.Cables[:0]
		} else if c.Output.Out == c {
			c.Output.Out = nil
		} else if c.Input.Out == c {
			c.Input.Out = nil
		}

		c.Output.In = utils.RemoveItem(c.Output.In, c)
		c.Input.In = utils.RemoveItem(c.Input.In, c)

		c.Connected = false
		return
	}

	hasWhich := len(which_) == 0
	which := which_[0]
	alreadyEmitToInstance := false

	if c.Input != nil {
		c.Input.QCache = nil
	}

	if c.Owner != nil && (!hasWhich || which == c.Owner) {
		c.Owner.Cables = utils.RemoveItem(c.Owner.Cables, c)

		if c.Connected {
			temp := &CableEvent{
				Cable:  c,
				Port:   c.Owner,
				Target: c.Target,
			}

			c.Owner.Emit("disconnect", temp)
			c.Owner.Iface.Emit("cable.disconnect", temp)
			ins := utils.GetPropertyRef(c.Owner.Iface.Node, "Instance").(*Instance)
			ins.Emit("cable.disconnect", temp)

			alreadyEmitToInstance = true
		} else {
			c.Owner.Iface.Emit("cable.cancel", &CableEvent{
				Cable:  c,
				Port:   c.Owner,
				Target: nil,
			})
		}
	}

	if c.Target != nil && c.Connected && (!hasWhich || which == c.Target) {
		c.Target.Cables = utils.RemoveItem(c.Target.Cables, c)

		temp := &CableEvent{
			Cable:  c,
			Port:   c.Target,
			Target: c.Owner,
		}

		c.Target.Emit("disconnect", temp)
		c.Target.Iface.Emit("cable.disconnect", temp)

		if alreadyEmitToInstance == false {
			ins := utils.GetPropertyRef(c.Target.Iface.Node, "Instance").(*Instance)
			ins.Emit("cable.disconnect", temp)
		}
	}
}
