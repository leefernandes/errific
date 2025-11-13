package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func Example_withStack() {
	Configure(WithStack)
	var ErrExample Err = "example error"

	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [errific/examples/example_withstack_test.go:15.Example_withStack]
	//   _testmain.go:73.main
	// true
}

func Example_withStackBubbled() {
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
	// top error: fmt wrapped 3: dynamic error [errific/examples/example_withstack_test.go:32.Example_withStackBubbled]
	// fmt wrapped 1: root error [errific/examples/example_withstack_test.go:30.Example_withStackBubbled]
	// EOF [errific/examples/example_withstack_test.go:34.Example_withStackBubbled]
	//   _testmain.go:73.main
	// true
}
