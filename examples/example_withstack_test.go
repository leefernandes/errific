package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func ExampleWithStack() {
	Configure(WithStack)
	var ErrExample Err = "example error"

	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [errific/examples/example_withstack_test.go:15.ExampleWithStack]
	//   /src/testing/run_example.go:63.runExample
	//   /src/testing/example.go:44.runExamples
	//   /src/testing/testing.go:1927.Run
	//   _testmain.go:75.main
	//   /src/runtime/proc.go:267.main
	//   /src/runtime/asm_amd64.s:1650.goexit
	// true
}

func ExampleWithStackBubbled() {
	Configure(WithStack)
	var ErrRoot Err = "root error"
	var ErrTop Err = "top error"

	err1 := ErrRoot.New(io.EOF)
	err2 := fmt.Errorf("fmt wrapped 1: %w", err1)
	err3 := Err("dynamic error").New(err2)
	err4 := fmt.Errorf("fmt wrapped 3: %w", err3)
	err5 := ErrTop.Withf("%w", err4)

	fmt.Println(err5)
	fmt.Println(errors.Is(err5, ErrRoot))

	// Output:
	// top error: fmt wrapped 3: dynamic error [errific/examples/example_withstack_test.go:37.ExampleWithStackBubbled]
	// fmt wrapped 1: root error [errific/examples/example_withstack_test.go:35.ExampleWithStackBubbled]
	// EOF [errific/examples/example_withstack_test.go:39.ExampleWithStackBubbled]
	//   /src/testing/run_example.go:63.runExample
	//   /src/testing/example.go:44.runExamples
	//   /src/testing/testing.go:1927.Run
	//   _testmain.go:75.main
	//   /src/runtime/proc.go:267.main
	//   /src/runtime/asm_amd64.s:1650.goexit
	// true
}
