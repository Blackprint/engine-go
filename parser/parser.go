package parser

import "encoding/json"

func ParseString(val string, dst *Root) (err error) {
	err = json.Unmarshal([]byte(val), dst)
	return
}
