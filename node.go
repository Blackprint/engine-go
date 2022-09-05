package engine

import "github.com/blackprint/engine-go/parser"

type Node struct {
	ID        int
	Namespace string
	Pos       parser.NodePos
	Data      parser.NodeData

	Cables []*Cable
}

func NewNode() *Node {
	return &Node{}
}

func (n *Node) CableTo(n2 *Node) *Cable {
	cab := NewCable(n, n2)
	n.Cables = append(n.Cables, cab)
	return cab
}

func NewNodeFromParser(node *parser.Node) *Node {
	n := NewNode()
	return n.LoadParserData(node)
}

func (n *Node) LoadParserData(pn *parser.Node) *Node {
	n.ID = pn.Index
	n.Pos = pn.NodePos
	n.Data = pn.Data
	return n
}
