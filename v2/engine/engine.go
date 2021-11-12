package engine

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

type Engine struct {
	Config

	model InputData

	nodesByIndex map[int64]*Node
	mx           sync.RWMutex
}

func New(opts ...Option) (e *Engine, err error) {
	e = &Engine{
		model:        map[Namespace][]*Node{},
		nodesByIndex: map[int64]*Node{},
	}
	if e.Config, err = scaffoldOptions(opts); err != nil {
		return
	}
	return
}

func (e *Engine) reload() (err error) {
	e.mx.Lock()
	defer e.mx.Unlock()
	for ns, nodes := range e.model {
		if err = e.CreateNode(ns, nodes); err != nil {
			err = errors.Wrap(err, "reload")
			return
		}
	}
	return
}

//

func (e *Engine) GetNodes(ns Namespace) []*Node {
	nodes, ok := e.model[ns]
	if !ok {
		return nil
	}
	return nodes
}

func (e *Engine) CreateNode(ns Namespace, nodes []*Node) (err error) {
	for _, node := range nodes {
		if v, ok := e.nodesByIndex[node.Index]; ok {
			err = fmt.Errorf("node with index %d already exists as %v instead of %v",
				node.Index, v, node)
			return
		}
		e.nodesByIndex[node.Index] = node
	}
	return
}
