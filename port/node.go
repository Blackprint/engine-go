package port

type GetterSetter func(...any) any
type Node struct {
	Input    map[string]GetterSetter
	Output   map[string]GetterSetter
	Property map[string]GetterSetter
}
