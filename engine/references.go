package engine

type referencesShortcut struct {
	IInput  map[string]*Port
	Input   map[string]*portInputGetterSetter
	IOutput map[string]*Port
	Output  map[string]*portOutputGetterSetter
}
