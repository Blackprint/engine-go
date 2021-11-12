package engine

type Engine struct {
	Config

	model InputData
}

func New(opts ...Option) (e *Engine, err error) {
	e = &Engine{}
	if e.Config, err = scaffoldOptions(opts); err != nil {
		return
	}
	return
}
