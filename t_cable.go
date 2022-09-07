package engine

type Cable struct {
	Source      *Node
	Destination *Node

	Port *Port
}

func NewCable(a, b *Node) *Cable {
	return &Cable{
		Source:      a,
		Destination: b,
	}
}

func (cab *Cable) SetPort(p *Port) {
	cab.Port = p
}

func (cab *Cable) Emit(ev Event) (err error) {
	return nil
}
