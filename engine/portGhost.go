package engine

var fakeIface = &Interface{
	Title:   "Blackprint.PortGhost",
	IsGhost: true,
	Node: &Node{
		Instance: &Instance{},
	},
}

func OutputPort(type_ any) *Port {
	port := fakeIface.QCreatePort("Output", "Blackprint.OutputPort", type_)
	return port
}

func InputPort(type_ any) *Port {
	port := fakeIface.QCreatePort("Input", "Blackprint.InputPort", type_)
	return port
}
