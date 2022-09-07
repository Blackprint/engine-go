package example

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/blackprint/engine-go/engine"
)

// for benchmarking
var instance *engine.Instance
var input *InputSimpleIFace

func TestMain(m *testing.M) {
	log.SetFlags(0)
	// log.SetOutput(ioutil.Discard)

	// Nodes already be registered from ./example.go -> init()

	// === Import JSON after all nodes was registered ===
	// You can import the JSON to Blackprint Sketch if you want to view the nodes visually
	instance = engine.New()
	json := `{"Example/Math/Random":[{"i":0,"x":298,"y":73,"output":{"Out":[{"i":2,"name":"A"}]}},{"i":1,"x":298,"y":239,"output":{"Out":[{"i":2,"name":"B"}]}}],"Example/Math/Multiply":[{"i":2,"x":525,"y":155,"output":{"Result":[{"i":3,"name":"Any"}]}}],"Example/Display/Logger":[{"i":3,"id":"myLogger","x":763,"y":169}],"Example/Button/Simple":[{"i":4,"id":"myButton","x":41,"y":59,"output":{"Clicked":[{"i":2,"name":"Exec"}]}}],"Example/Input/Simple":[{"i":5,"id":"myInput","x":38,"y":281,"data":{"value":"saved input"},"output":{"Changed":[{"i":1,"name":"Re-seed"}],"Value":[{"i":3,"name":"Any"}]}}]}`
	// json := `{"_":{"moduleJS":["http://localhost:6789/dist/nodes-example.mjs"],"functions":{"Test":{"id":"Test","title":"Test","description":"No description","vars":["shared"],"privateVars":["private"],"structure":{"BP/Fn/Input":[{"i":0,"x":389,"y":100,"z":3,"output":{"A":[{"i":2,"name":"A"}],"Exec":[{"i":2,"name":"Exec"}]}}],"BP/Fn/Output":[{"i":1,"x":973,"y":228,"z":14}],"Example/Math/Multiply":[{"i":2,"x":656,"y":99,"z":8,"output":{"Result":[{"i":3,"name":"Val"},{"i":9,"name":"Val"}]}},{"i":10,"x":661,"y":289,"z":4,"output":{"Result":[{"i":5,"name":"Val"},{"i":1,"name":"Result1"}]}}],"BP/Var/Set":[{"i":3,"x":958,"y":142,"z":9,"data":{"name":"shared","scope":2}},{"i":5,"x":971,"y":333,"z":2,"data":{"name":"private","scope":1},"route":{"i":1}}],"BP/Var/Get":[{"i":4,"x":387,"y":461,"z":5,"data":{"name":"shared","scope":2},"output":{"Val":[{"i":8,"name":"Any"}]}},{"i":6,"x":389,"y":524,"z":0,"data":{"name":"private","scope":1},"output":{"Val":[{"i":8,"name":"Any"}]}}],"BP/FnVar/Input":[{"i":7,"x":387,"y":218,"z":7,"data":{"name":"B"},"output":{"Val":[{"i":2,"name":"B"}]}},{"i":11,"x":386,"y":301,"z":6,"data":{"name":"Exec"},"output":{"Val":[{"i":10,"name":"Exec"}]}},{"i":12,"x":386,"y":370,"z":10,"data":{"name":"A"},"output":{"Val":[{"i":10,"name":"A"},{"i":10,"name":"B"}]}}],"Example/Display/Logger":[{"i":8,"x":661,"y":474,"z":11}],"BP/FnVar/Output":[{"i":9,"x":956,"y":69,"z":1,"data":{"name":"Result"}},{"i":14,"x":969,"y":629,"z":13,"data":{"name":"Clicked"}}],"Example/Button/Simple":[{"i":13,"x":634,"y":616,"z":12,"output":{"Clicked":[{"i":14,"name":"Val"}]}}]}}}},"Example/Math/Random":[{"i":0,"x":512,"y":76,"z":0,"output":{"Out":[{"i":5,"name":"A"}]},"route":{"i":5}},{"i":1,"x":512,"y":242,"z":1,"output":{"Out":[{"i":5,"name":"B"}]}}],"Example/Display/Logger":[{"i":2,"x":986,"y":282,"z":2,"id":"myLogger"}],"Example/Button/Simple":[{"i":3,"x":244,"y":64,"z":6,"id":"myButton","output":{"Clicked":[{"i":5,"name":"Exec"}]}}],"Example/Input/Simple":[{"i":4,"x":238,"y":279,"z":4,"id":"myInput","data":{"value":"saved input"},"output":{"Changed":[{"i":1,"name":"Re-seed"}],"Value":[{"i":2,"name":"Any"}]}}],"BPI/F/Test":[{"i":5,"x":738,"y":138,"z":5,"output":{"Result1":[{"i":2,"name":"Any"}],"Result":[{"i":2,"name":"Any"}],"Clicked":[{"i":6,"name":"Exec"}]},"route":{"i":6}}],"Example/Math/Multiply":[{"i":6,"x":1032,"y":143,"z":3}]}`
	instance.ImportJSON([]byte(json))

	// Because Golang lack of getter and setter, We need to get or set like calling a function
	// Anyway.. lets to run something :)
	button := instance.Iface["myButton"].(*ButtonSimpleIFace)

	log.Println("\n>> I'm clicking the button")
	button.Clicked(123)

	logger := instance.Iface["myLogger"].(*LoggerIFace)
	log.Println("\n>> I got the output value: " + logger.Log().(string))

	log.Println("\n>> I'm writing something to the input box")
	input = instance.Iface["myInput"].(*InputSimpleIFace)
	input.Data["value"].Set("hello wrold")

	// you can also use getNodes if you haven't set the ID
	myLogger := instance.GetNodes("Example/Display/Logger")[0].(*LoggerNode).Iface.(*LoggerIFace)
	log.Println("\n>> I got the output value: " + myLogger.Log().(string))
}

func BenchmarkInputBox(b *testing.B) {
	if instance == nil {
		TestMain(nil)
	}

	for i := 0; i < b.N; i++ {
		input.Data["value"].Set("hello wrold" + strconv.Itoa(time.Now().Nanosecond()))
	}
}
