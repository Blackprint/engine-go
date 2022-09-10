package engine

import "regexp"

type environment struct {
	_noEvent bool
	Map      map[string]string
}

var QEnvironment = &environment{
	Map: map[string]string{},
}

// arr = ["KEY": "value"]
func (e *environment) Import(arr map[string]string) {
	e._noEvent = true
	for key, val := range arr {
		e.Set(key, val)
	}
	e._noEvent = false
	Event.Emit("environment.imported", nil)
}

var envSetRegx = regexp.MustCompile(`[^A-Z_][^A-Z0-9_]`)

type EnvironmentEvent struct {
	Key   string
	Value string
}

func (e *environment) Set(key string, val string) {
	if len(envSetRegx.FindStringIndex(key)) > 0 {
		panic("Environment must be uppercase and not contain any symbol except underscore, and not started by a number. But got: " + key)
	}

	e.Map[key] = val

	if !e._noEvent {
		Event.Emit("environment.added", &EnvironmentEvent{
			Key:   key,
			Value: val,
		})
	}
}

func (e *environment) Delete(key string) {
	delete(e.Map, key)
	Event.Emit("environment.deleted", &EnvironmentEvent{
		Key:   key,
		Value: "",
	})
}
