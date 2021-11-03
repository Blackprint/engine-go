package engine

import (
    "strings"
)

type eventObj struct {
	callback interface{}
	once bool
}

type customEvent struct {
	events map[string] []eventObj {}
}

func (e *customEvent) On(evName string, callback interface{}, once bool) {
	evs := strings.Split(evName, ",")

	for _, name := range evs {
		list := e.events[name]

		// Only add when not exist
		exist := false
		for _, cb := range list {
			if cb.callback == callback {
				exist = true
				break
			}
		}

		if exist {
			break
		}

		e.events[name] = append(list, eventObj{
			callback: callback,
			once: once,
		})
	}
}

func (e *customEvent) once(evName string, callback interface{}) {
	e.On(evName, callback, true)
}

func (e *customEvent) Off(evName string, callback interface{}) {
	evs := strings.Split(evName, ",")

	for _, name := range evs {
		if callback == nil {
			e.events[name] = []
			break
		}

		list := e.events[name]
		for i, cb := range list {
			if cb.callback == callback {
				e.events[name] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}

func (e *customEvent) QTrigger(evName string, data interface{}) {
	list := e.events[name]
	for i, cb := range list {
		cb.callback(data)

		if cb.once {
			e.events[name] = append(list[:i], list[i+1:]...)
		}
	}
}