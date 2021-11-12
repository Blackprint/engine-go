package main

import "github.com/blackprint/engine-go/v2/engine"

func main() {
	data := []byte(`{
		"Example/Math/Random": [{
			"i": 0,
			"x": 298,
			"y": 73,
			"output": {
				"Out": [{
					"i": 2,
					"name": "A"
				}]
			}
		}, {
			"i": 1,
			"x": 298,
			"y": 239,
			"output": {
				"Out": [{
					"i": 2,
					"name": "B"
				}]
			}
		}],
		"Example/Math/Multiply": [{
			"i": 2,
			"x": 525,
			"y": 155,
			"output": {
				"Result": [{
					"i": 3,
					"name": "Any"
				}]
			}
		}],
		"Example/Display/Logger": [{
			"i": 3,
			"id": "myLogger",
			"x": 763,
			"y": 169
		}],
		"Example/Button/Simple": [{
			"i": 4,
			"id": "myButton",
			"x": 41,
			"y": 59,
			"output": {
				"Clicked": [{
					"i": 2,
					"name": "Exec"
				}]
			}
		}],
		"Example/Input/Simple": [{
			"i": 5,
			"id": "myInput",
			"x": 38,
			"y": 281,
			"data": {
				"value": "saved input"
			},
			"output": {
				"Changed": [{
					"i": 1,
					"name": "Re-seed"
				}],
				"Value": [{
					"i": 3,
					"name": "Any"
				}]
			}
		}]
	}`)

	engine, err := engine.New()
	if err != nil {
		panic(err)
	}
	if err = engine.LoadJSONBytes(data); err != nil {
		panic(err)
	}
}
