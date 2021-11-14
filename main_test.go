package main

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/blackprint/engine-go/engine"
	"github.com/blackprint/engine-go/example"
)

// for benchmarking
var instance *engine.Instance
var input *example.InputSimpleIFace

func TestMain(m *testing.M) {
	log.SetFlags(0)
	// log.SetOutput(ioutil.Discard)

	example.RegisterAll()

	// === Import JSON after all nodes was registered ===
	// You can import the JSON to Blackprint Sketch if you want to view the nodes visually
	instance = engine.New()
	json := `{"Example/Math/Random":[{"i":0,"x":298,"y":73,"output":{"Out":[{"i":2,"name":"A"}]}},{"i":1,"x":298,"y":239,"output":{"Out":[{"i":2,"name":"B"}]}}],"Example/Math/Multiply":[{"i":2,"x":525,"y":155,"output":{"Result":[{"i":3,"name":"Any"}]}}],"Example/Display/Logger":[{"i":3,"id":"myLogger","x":763,"y":169}],"Example/Button/Simple":[{"i":4,"id":"myButton","x":41,"y":59,"output":{"Clicked":[{"i":2,"name":"Exec"}]}}],"Example/Input/Simple":[{"i":5,"id":"myInput","x":38,"y":281,"data":{"value":"saved input"},"output":{"Changed":[{"i":1,"name":"Re-seed"}],"Value":[{"i":3,"name":"Any"}]}}]}`
	instance.ImportJSON([]byte(json))

	// Because Golang lack of getter and setter, We need to get or set like calling a function
	// Anyway.. lets to run something :)
	button := instance.Iface["myButton"].(*example.ButtonSimpleIFace)

	log.Println("\n>> I'm clicking the button")
	button.Clicked(123)

	logger := instance.Iface["myLogger"].(*example.LoggerIFace)
	log.Println("\n>> I got the output value: " + logger.Log().(string))

	log.Println("\n>> I'm writing something to the input box")
	input = instance.Iface["myInput"].(*example.InputSimpleIFace)
	input.Data["value"]("hello wrold")

	// you can also use getNodes if you haven't set the ID
	myLogger := instance.GetNodes("Example/Display/Logger")[0].(*example.LoggerNode).Iface.(*example.LoggerIFace)
	log.Println("\n>> I got the output value: " + myLogger.Log().(string))
}

func BenchmarkInputBox(b *testing.B) {
	if instance == nil {
		TestMain(nil)
	}

	for i := 0; i < b.N; i++ {
		input.Data["value"]("hello wrold" + strconv.Itoa(time.Now().Nanosecond()))
	}
}
