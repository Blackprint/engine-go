package engine

import (
	"fmt"
	"sync"
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
	for namespace, nodes := range e.model {
		_ = namespace
		for _, node := range nodes {
			if v, ok := e.nodesByIndex[node.Index]; ok {
				err = fmt.Errorf("node with index %d already exists as %v instead of %v",
					node.Index, v, node)
				return
			}
		}
	}
	return
}
