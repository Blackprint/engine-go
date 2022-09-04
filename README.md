<p align="center"><a href="#" target="_blank" rel="noopener noreferrer"><img width="150" src="https://user-images.githubusercontent.com/11073373/141421213-5decd773-a870-4324-8324-e175e83b0f55.png" alt="Blackprint"></a></p>

<h1 align="center">Blackprint Engine for Golang</h1>
<p align="center">Run exported Blackprint on Golang environment.</p>

<p align="center">
    <a href='https://github.com/Blackprint/Blackprint/blob/master/LICENSE'><img src='https://img.shields.io/badge/License-MIT-brightgreen.svg' height='20'></a>
</p>

## Example

![VEfiZCFQAi](https://user-images.githubusercontent.com/11073373/141679777-ddd9738f-4d09-42f1-b635-9d0c9a8c5952.png)

This repository provide an example with the JSON too, and you can try it on CLI:<br>

```sh
# Change your working directory into empty folder first
$ git clone --depth 1 https://github.com/Blackprint/engine-go .
$ go mod init
$ go test -v ./example
```

## Documentation

> Warning: This project haven't reach it stable version (semantic versioning at v1.0.0)<br>
> But please try to use it and help improve this project

### Defining Blackprint Node and Interface
Because Golang doesn't support Class Oriented programming, this engine will only support Node/Interface that declared with function as a constructor.

```go
package main
import (
    "log"

    Blackprint "github.com/blackprint/engine-go/blackprint"
    "github.com/blackprint/engine-go/engine"
    "github.com/blackprint/engine-go/types"
)

// class HelloNode extends engine.Node
type HelloNode struct {
    *engine.Node
}

// This will be registered as Node definition
func main() {
    Blackprint.RegisterNode("Example/Hello", func(instance *engine.Instance) any {
        node := HelloNode{
            Node: &engine.Node{ // Contruct the parent constructor
                Instance: instance,

                // Please remember to capitalize the port name

                // Set the output port structure for your node (Optional)
                TOutput: engine.NodePort{ // Port Template
                    "Changed": types.Function,
                    //> Callable: node.Output["Changed"](...args)

                    "Output": types.Number,
                    //> node.Output["Value"](246) - set Value to 246
                },

                // Set the input port structure for your node (Optional)
                TInput: engine.NodePort{ // Port Template
                    "Multiply":
                },
            },
        }

        // Set the Interface, let it empty if you want
        // to use default empty interface "setInterface()"
        iface := node.SetInterface("BPIC/Example/Hello").(*HelloIFace)
        iface.Title = "Hello"

        // This constructor function must return the pointer
        // After this node being processed by the engine
        // TOutput/TInput will be converted to Output/Input
        return &node
    })
}
```

Because Node is supposed to contain structure only it should be designed to be simple, the another complexity like calling system API or providing API for developer to interact with your node should be placed on Interface class.

```go
package main
import (
    "log"

    Blackprint "github.com/blackprint/engine-go/blackprint"
    "github.com/blackprint/engine-go/engine"
    "github.com/blackprint/engine-go/types"
)

// class HelloIFace extends engine.Interface
type HelloIFace struct {
    *engine.Interface
}

// Capitalize 'recalculate' to 'Recalculate' to make it public
// public function recalculate(){...}
func(iface *HelloIFace) Recalculate(){
    node := iface.Node.(*HelloNode)

    // Get value from input port (with GetterSetter)
    multiplyBy := node.Input["Multiply"]()

    // Assign new value to output port (with GetterSetter)
    // node.Output["Output"] = ...
    node.Output["Output"](iface.Data["Value"] * multiplyBy)
}

func main() {
    // Your Interface namespace must use "BPIC" as the prefix
    Blackprint.RegisterInterface("BPIC/Example/Hello", func(node any) any {
        // node := node_.(HelloNode)
        value := 123

        var iface HelloIFace
        iface = HelloIFace{
            Interface: &engine.Interface{  // Contruct the parent constructor
                Data: engine.InterfaceData{
                    "value": func(val ...any) any {
                        if len(val) == 0 { // Getter
                            return value
                        }
                        // else-> Setter
                        value = val.(int)

                        // Call recalculate() on HelloIFace
                        iface.Recalculate();
                        return nil
                    },
                },
            },
        }

        // This constructor function must return the pointer
        return &iface
    }
}
```

## Creating new Engine instance

```go
package main
import (
    "fmt"
    "github.com/blackprint/engine-go/engine"
)

func main(){
    // Create Blackprint Engine instance
    instance := engine.New()

    // You can import nodes with JSON
    // if the nodes haven't been registered, this will throw an error
    instance.ImportJSON([]byte(`{...}`));

    // You can also create the node dynamically
    iface_ := instance.CreateNode("Example/Hello", /* [..options..] */);

    // You must type cast it to target Interface
    iface := iface_.(*HelloIFace)

    // Change the default data 'Value' property
    iface.Data["Value"] = 123

    // --- Obtaining Node from Interface
    // You must type cast it to target Node
    node := iface.Node.(*HelloNode)

    // Assign the 'Multiply' input port = 2
    node.Input["Multiply"](2)

    // Get the value from 'Output' output port
    fmt.Println(node.Output["Output"]()) // 246
}
```

### Calling function on output port
For more detailed implementation, please `./example/button.go` and `./example/example_test.go`.

```go
package main
import (
    "fmt"
    "github.com/blackprint/engine-go/engine"
    "github.com/blackprint/engine-go/port"
    "github.com/blackprint/engine-go/types"
)

func main(){
    // For the example you was registered a Node
    // with callable output port somewhere

    // Blackprint.RegisterNode("Example/Button/Simple", ... {
        node := ButtonSimple{
            Node: &engine.Node{
                Instance: instance,

                // Node's Output Port Template
                TOutput: engine.NodePort{
                    "Clicked": types.Function,
                },
            },
        }
    // }

    iface := instance.CreateNode("Example/Button/Simple")
    node := iface.(*ButtonInterface).Node.(*ButtonSimple)

    // You can call it w/o parameter, and it will call
    // every connected Trigger port on every node
    node.Output["Clicked"]()

    // Note: Trigger port can only be placed on Input port
    // Example on `./example/math.go`
        TInput: engine.NodePort{
            "Clicked": port.Trigger(func(args ...any) {
                fmt.Println(args)
            }),
        },
}
```

## Note

This engine is focused for easy to use API, currently some implementation still not efficient because Golang doesn't support:

- Getter/setter (solution: use function as a getter/setter)
- Optional parameter (solution: use `...any`)

`any` or `any` still being used massively to make Blackprint works without too much complexity in Golang, so please don't expect too high for the performance. Some implementation also use `reflect` for obtaining and setting value by using dynamic field name. Current implementation is focused to be similar with PHP and JS.

## License

MIT
