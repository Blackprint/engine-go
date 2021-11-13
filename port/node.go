package port

type GetterSetter func(...interface{}) interface{}
type Node struct {
	Input    map[string]GetterSetter
	Output   map[string]GetterSetter
	Property map[string]GetterSetter
}
