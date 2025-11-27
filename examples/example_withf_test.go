package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func Example_withf() {
	Configure(OutputPretty) // default configuration
	var ErrExample Err = "example error"
	err := ErrExample.Withf("int (%d) string (%s): %w", 123, "yarn", io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error: int (123) string (yarn): EOF [errific/examples/example_withf_test.go:14.Example_withf]
	// true
	// true
}

func Example_withfNest() {
	Configure(OutputPretty) // default configuration
	var (
		Err1 Err = "error 1"
		Err2 Err = "error 2"
	)
	err1 := Err1.Withf("with format %d", 1).Join(io.EOF)
	err2 := Err2.Withf("with format %d", 2).Join(err1)

	fmt.Println(err2)
	fmt.Println(errors.Is(err2, Err2))
	fmt.Println(errors.Is(err2, Err1))
	fmt.Println(errors.Is(err2, io.EOF))

	// Output:
	// error 2: with format 2 [errific/examples/example_withf_test.go:32.Example_withfNest]
	// error 1: with format 1 [errific/examples/example_withf_test.go:31.Example_withfNest]
	// EOF
	// true
	// true
	// true
}

func Example_withfChain() {
	Configure(OutputPretty) // default configuration
	var ErrExample Err = "example error"

	err := ErrExample.
		Withf("first %d", 1).
		Withf("second %d", 2).
		Withf("third %d", 3).
		Join(io.EOF)

	fmt.Println(err)
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error: first 1: second 2: third 3 [errific/examples/example_withf_test.go:53.Example_withfChain]
	// EOF
	// true
}
