package engine

import "github.com/blackprint/engine-go/parser"

type Engine struct {
}

func New() *Engine {
	eng := &Engine{}
	return eng
}

func (eng *Engine) ImportJSON(val string) (err error) {
	var root parser.Root
	err = parser.ParseString(val, &root)

	return
}

func (eng *Engine) Register(root parser.Root) {

}
