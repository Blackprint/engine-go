package engine

type Cable struct {
	A *Node
	B *Node
}

func NewCable(a, b *Node) *Cable {
	return &Cable{a, b}
}
