package errific_test

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func ExampleNew() {
	Configure(Newline, Suffix)
	// w/out wrapping errors.
	var ErrExample Err = "example error"
	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [errific/example_test.go:15.ExampleNew]
	// true
}

func ExampleNewWrapError() {
	Configure(Newline, Suffix)
	// wrap an error.
	var ErrExample Err = "example error"
	err := ErrExample.New(io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/example_test.go:28.ExampleNewWrapError]
	// EOF
	// true
	// true
}

func ExampleNewWrapErrors() {
	Configure(Newline, Suffix)
	// wrap multiple errors.
	var ErrExample Err = "example error"
	err := ErrExample.New(io.ErrUnexpectedEOF, io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.ErrUnexpectedEOF))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/example_test.go:44.ExampleNewWrapErrors]
	// unexpected EOF
	// EOF
	// true
	// true
	// true
}

func ExampleWrapf() {
	Configure(Newline, Suffix)
	// wrap a formatted error.
	var ErrExample Err = "example error"
	err := ErrExample.Wrapf("formatted %d: %w", 1, io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error [errific/example_test.go:63.ExampleWrapf]
	// formatted 1: EOF
	// true
	// true
}

func ExampleErrorf() {
	Configure(Newline, Suffix)
	// format an error with parameters.
	var ErrExample Err = "formatted error: %s %w"
	err := ErrExample.Errorf("io error", io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// formatted error: io error EOF [errific/example_test.go:79.ExampleErrorf]
	// true
	// true
}

func ExampleWithf() {
	Configure(Newline, Suffix)
	var ErrExample Err = "example error"
	err := ErrExample.Withf("int (%d) string (%s): %w", 123, "yarn", io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// example error: int (123) string (yarn): EOF [errific/example_test.go:93.ExampleWithf]
	// true
	// true
}

func ExampleNewChain() {
	Configure(Newline, Suffix)
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
	// example error [errific/example_test.go:117.ExampleNewChain]
	// error 3 [errific/example_test.go:114.ExampleNewChain]
	// error 2 [errific/example_test.go:113.ExampleNewChain]
	// error 1 [errific/example_test.go:112.ExampleNewChain]
	// EOF
	// true
	// true
	// true
	// true
}

func ExampleWrapfChain() {
	Configure(Newline, Suffix)
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
	// error 2 [errific/example_test.go:143.ExampleWrapfChain]
	// format 1: error 1 [errific/example_test.go:142.ExampleWrapfChain]
	// format 0: EOF
	// true
	// true
	// true
}
