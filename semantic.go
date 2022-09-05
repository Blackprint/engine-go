package engine

import (
	"fmt"

	"github.com/blackprint/engine-go/parser"
)

type nodeList []*Node
type nodeMap map[int]*Node
type parserNodeMapper map[int]*parser.Node

type semantic struct {
	pNodeMapper parserNodeMapper
	nodeList    nodeList
	nodeMap     nodeMap
	cables      []*Cable
}

func newSemantic() *semantic {
	sem := &semantic{
		pNodeMapper: parserNodeMapper{},
		nodeList:    nodeList{},
		nodeMap:     nodeMap{},
		cables:      []*Cable{},
	}
	return sem
}

func (s *semantic) registerParserNodes(ns parser.NodeName, nodes parser.NodeList, allowOverwrite bool) error {
	for _, node := range nodes {
		if _, exist := s.pNodeMapper[node.Index]; exist && !allowOverwrite {
			return fmt.Errorf("node id already exist, overwrite disallowed")
		}
		s.pNodeMapper[node.Index] = node
		n := NewNodeFromParser(node)
		n.Namespace = ns

		s.nodeMap[node.Index] = n
		s.nodeList = append(s.nodeList, n)
	}
	return nil
}
func (s *semantic) loadParserNodes(root parser.Root) (err error) {
	for ns, nodes := range root {
		if ns == "_" {
			continue
		}
		if err = s.registerParserNodes(ns, nodes, false); err != nil {
			return
		}
	}
	return
}
func (s *semantic) generateCablesAndPorts() (err error) {
	for _, node := range s.nodeList {
		meta := s.pNodeMapper[node.ID]
		for _, outputs := range meta.Output {
			for _, portOfNode2 := range outputs {
				node2, exist := s.nodeMap[portOfNode2.Index]
				if !exist {
					return fmt.Errorf("target dst cable not found: %+#v", portOfNode2)
				}
				// create cables
				cab := node.CableTo(portOfNode2.Name, node2)
				s.cables = append(s.cables, cab)
			}
		}
	}
	return
}

func (s *semantic) process(root parser.Root) (err error) {
	if err = s.loadParserNodes(root); err != nil {
		return
	}
	if err = s.generateCablesAndPorts(); err != nil {
		return
	}
	return
}
