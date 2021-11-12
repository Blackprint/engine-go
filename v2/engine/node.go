package engine

type Namespace string

type NodeDataValue interface{}
type NodeData struct {
	Value NodeDataValue `json:"value"`
}

type NodeOutputHandler struct {
}
type NodeOutput struct {
	Clicked []NodeOutputHandler `json:"Clicked"`
	Out     []NodeOutputHandler `json:"Out"`
	Result  []NodeOutputHandler `json:"Result"`
	Value   []NodeOutputHandler `json:"Value"`
}
type Node struct {
	Index int64  `json:"i"`
	ID    string `json:"id"`
	X     int64  `json:"x"`
	Y     int64  `json:"y"`

	Data   *NodeData   `json:"data"`
	Output *NodeOutput `json:"output"`
}

type InputData map[Namespace][]*Node

//

func (n *Node) Connect(to *Node) (err error) {
	return
}
