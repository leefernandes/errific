package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func Example_wrapf() {
	Configure(OutputPretty) // default configuration
	// wrap a formatted error.
	var ErrExample Err = "example error"
	err := ErrExample.Wrapf("formatted %d: %w", 1, io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/examples/example_wrapf_test.go:15.Example_wrapf]
	// formatted 1: EOF
	// true
	// true
}

func Example_wrapfNest() {
	Configure(OutputPretty) // default configuration
	// wrapped & formatted errific error chain.
	var (
		Err1 Err = "error 1"
		Err2 Err = "error 2"
	)
	err1 := Err1.Wrapf("format %d: %w", 0, io.EOF)
	err2 := Err2.Wrapf("format %d: %w", 1, err1)

	fmt.Println(err2)
	fmt.Println(errors.Is(err2, Err2))
	fmt.Println(errors.Is(err2, Err1))
	fmt.Println(errors.Is(err2, io.EOF))

	// Output:
	// error 2 [errific/examples/example_wrapf_test.go:35.Example_wrapfNest]
	// format 1: error 1 [errific/examples/example_wrapf_test.go:34.Example_wrapfNest]
	// format 0: EOF
	// true
	// true
	// true
}

func Example_wrapfChain() {
	Configure(OutputPretty) // default configuration
	var ErrExample Err = "example error"

	err := ErrExample.
		Wrapf("first %d", 1).
		Wrapf("second %d", 2).
		Wrapf("third %d", 3).
		Join(io.EOF)

	fmt.Println(err)
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/examples/example_wrapf_test.go:56.Example_wrapfChain]
	// first 1
	// second 2
	// third 3
	// EOF
	// true
}
