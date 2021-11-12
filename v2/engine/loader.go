package engine

import (
	"encoding/json"
)

func (e *Engine) LoadJSONBytes(b []byte) (err error) {
	err = json.Unmarshal(b, &e.model)
	if err == nil {
		return e.reload()
	}
	return
}
