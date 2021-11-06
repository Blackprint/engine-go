package port

type Node struct {
	Input    map[string]func(args ...interface{}) interface{}
	Output   map[string]func(args ...interface{}) interface{}
	Property map[string]func(args ...interface{}) interface{}
}
