package engine

import (
	"strings"

	"github.com/blackprint/engine-go/utils"
)

type eventObj struct {
	callback any
	once     bool
}

type CustomEvent struct {
	events map[string][]*eventObj
}

func (e *CustomEvent) listen(evName string, callback any, once bool) {
	if e.events == nil {
		e.events = map[string][]*eventObj{}
	}

	evs := strings.Split(evName, " ")

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
			continue
		}

		e.events[name] = append(list, &eventObj{
			callback: callback,
			once:     once,
		})
	}
}

func (e *CustomEvent) On(evName string, callback any) {
	e.listen(evName, callback, false)
}

func (e *CustomEvent) Once(evName string, callback any) {
	e.listen(evName, callback, true)
}

func (e *CustomEvent) Off(evName string, callback any) {
	if e.events == nil {
		return
	}

	evs := strings.Split(evName, " ")
	for _, name := range evs {
		if callback == nil {
			e.events[name] = []*eventObj{}
			continue
		}

		list := e.events[name]
		if list == nil {
			continue
		}

		for i, cb := range list {
			if cb.callback == callback {
				e.events[name] = utils.RemoveItemAtIndex(list, i)
				continue
			}
		}
	}
}

func (e *CustomEvent) Emit(evName string, data any) {
	list := e.events[evName]
	for i, cb := range list {
		cb.callback.(func(any))(data)

		if cb.once {
			e.events[evName] = utils.RemoveItemAtIndex(list, i)
		}
	}
}
