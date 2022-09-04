package engine

import (
	"strings"
)

type eventObj struct {
	callback any
	once     bool
}

type customEvent struct {
	events map[string][]*eventObj
}

func (e *customEvent) listen(evName string, callback any, once bool) {
	if e.events == nil {
		e.events = map[string][]*eventObj{}
	}

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

		e.events[name] = append(list, &eventObj{
			callback: callback,
			once:     once,
		})
	}
}

func (e *customEvent) On(evName string, callback any) {
	e.listen(evName, callback, false)
}

func (e *customEvent) Once(evName string, callback any) {
	e.listen(evName, callback, true)
}

func (e *customEvent) Off(evName string, callback any) {
	if e.events == nil {
		return
	}

	evs := strings.Split(evName, ",")
	for _, name := range evs {
		if callback == nil {
			e.events[name] = []*eventObj{}
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

func (e *customEvent) QTrigger(evName string, data any) {
	list := e.events[evName]
	for i, cb := range list {
		cb.callback.(func(any))(data)

		if cb.once {
			e.events[evName] = append(list[:i], list[i+1:]...)
		}
	}
}
