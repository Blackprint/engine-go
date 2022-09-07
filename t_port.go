package engine

type Port struct {
	Name  string
	Cable *Cable
}

func NewPort(name string, cab *Cable) *Port {
	return &Port{Name: name, Cable: cab}
}

func (p *Port) Emit(ev Event) (err error) {
	if cab := p.Cable; cab != nil {
		return cab.Emit(ev)
	}
	return
}
