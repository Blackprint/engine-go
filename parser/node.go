package parser

type NodePos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type NodeOutputList []*Port

type NodeOutputType string

type NodeOutput map[NodeOutputType]NodeOutputList

type NodeData map[string]any

type Node struct {
	Index int      `json:"i"`
	Name  NodeName `json:"name"`
	Data  NodeData `json:"data"`
	NodePos
	Output NodeOutput `json:"output"`
}
