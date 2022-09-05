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
func (s *semantic) generateCables() (err error) {
	for _, node := range s.nodeList {
		meta := s.pNodeMapper[node.ID]
		for _, outputs := range meta.Output {
			for _, out := range outputs {
				node2, exist := s.nodeMap[out.Index]
				if !exist {
					return fmt.Errorf("target dst cable not found: %+#v", out)
				}
				// create cables
				cab1 := node.CableTo(node2)
				cab2 := node2.CableTo(node)
				s.cables = append(s.cables, cab1)
				s.cables = append(s.cables, cab2)
			}
		}
	}
	return
}

func (s *semantic) process(root parser.Root) (err error) {
	if err = s.loadParserNodes(root); err != nil {
		return
	}
	if err = s.generateCables(); err != nil {
		return
	}
	return
}
