package engine

import (
	"encoding/json"
)

func (e *Engine) LoadJSONBytes(b []byte) (err error) {
	return json.Unmarshal(b, &e.model)
}
