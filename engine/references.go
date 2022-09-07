package engine

type ReferencesShortcut struct {
	IInput  map[string]*Port
	Input   map[string]*PortInputGetterSetter
	IOutput map[string]*Port
	Output  map[string]*PortOutputGetterSetter
}
