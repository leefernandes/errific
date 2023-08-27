package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func ExampleNew() {
	Configure() // default configuration
	// w/out wrapping errors.
	var ErrExample Err = "example error"
	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [errific/examples/example_new_test.go:15.ExampleNew]
	// true
}

func ExampleNewWrapError() {
	Configure() // default configuration
	// wrap an error.
	var ErrExample Err = "example error"
	err := ErrExample.New(io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/examples/example_new_test.go:28.ExampleNewWrapError]
	// EOF
	// true
	// true
}

func ExampleNewWrapErrors() {
	Configure() // default configuration
	// wrap multiple errors.
	var ErrExample Err = "example error"
	err := ErrExample.New(io.ErrUnexpectedEOF, io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.ErrUnexpectedEOF))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/examples/example_new_test.go:44.ExampleNewWrapErrors]
	// unexpected EOF
	// EOF
	// true
	// true
	// true
}

func ExampleNewNest() {
	Configure() // default configuration
	// wrapped errific error chain.
	var (
		Err1 Err = "error 1"
		Err2 Err = "error 2"
		Err3 Err = "error 3"
	)
	err1 := Err1.New(io.EOF)
	err2 := Err2.New(err1)
	err3 := Err3.New(err2)

	var ErrExample Err = "example error"
	fmt.Println(ErrExample.New(err3))
	fmt.Println(errors.Is(err3, Err3))
	fmt.Println(errors.Is(err3, Err2))
	fmt.Println(errors.Is(err3, Err1))
	fmt.Println(errors.Is(err3, io.EOF))

	// Output:
	// example error [errific/examples/example_new_test.go:72.ExampleNewNest]
	// error 3 [errific/examples/example_new_test.go:69.ExampleNewNest]
	// error 2 [errific/examples/example_new_test.go:68.ExampleNewNest]
	// error 1 [errific/examples/example_new_test.go:67.ExampleNewNest]
	// EOF
	// true
	// true
	// true
	// true
}
