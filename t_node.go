package engine

import "github.com/blackprint/engine-go/parser"

type Node struct {
	ID        int
	Namespace parser.NodeName
	Pos       parser.NodePos
	Data      parser.NodeData

	Ports  []*Port
	Cables []*Cable
}

func NewNode(id int, ns parser.NodeName) *Node {
	return &Node{ID: id, Namespace: ns}
}

func (n *Node) addPort(p *Port) {
	n.Ports = append(n.Ports, p)
}
func (n *Node) addCable(cab *Cable) {
	n.Cables = append(n.Cables, cab)
}

func (n *Node) CableTo(on string, n2 *Node) *Cable {
	cab := NewCable(n, n2)
	port := NewPort(on, cab)
	cab.SetPort(port)
	n.addCable(cab)
	n2.addPort(port)
	return cab
}

func NewNodeFromParser(node *parser.Node) *Node {
	n := NewNode(node.Index, node.Name)
	return n.LoadParserData(node)
}

func (n *Node) LoadParserData(pn *parser.Node) *Node {
	n.ID = pn.Index
	n.Pos = pn.NodePos
	n.Data = pn.Data
	n.Namespace = pn.Name
	return n
}
