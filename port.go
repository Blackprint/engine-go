package engine

type Port struct {
	Name  string
	Cable *Cable
}

func NewPort(name string, cab *Cable) *Port {
	return &Port{Name: name, Cable: cab}
}
