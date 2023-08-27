package examples

import (
	"errors"
	"fmt"

	. "github.com/leefernandes/errific"
)

func ExampleWithStack() {
	Configure(WithStack)
	var ErrExample Err = "example error"
	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [errific/examples/example_withstack_test.go:13.ExampleWithStack]
	//   /src/testing/run_example.go:63.runExample
	//   /src/testing/example.go:44.runExamples
	//   /src/testing/testing.go:1927.Run
	//   _testmain.go:73.main
	//   /src/runtime/proc.go:267.main
	//   /src/runtime/asm_amd64.s:1650.goexit
	// true
}
