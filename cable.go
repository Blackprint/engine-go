package engine

type Cable struct {
	A    *Node
	B    *Node
	Port *Port
}

func NewCable(a, b *Node) *Cable {
	return &Cable{
		A: a,
		B: b,
	}
}

func (cab *Cable) SetPort(p *Port) {
	cab.Port = p
}
