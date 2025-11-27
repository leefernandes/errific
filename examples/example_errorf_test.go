package examples

import (
	"errors"
	"fmt"
	"io"

	. "github.com/leefernandes/errific"
)

func ExampleErrorf() {
	Configure(OutputPretty) // default configuration
	// format an error with parameters.
	var ErrExample Err = "formatted error: %s %w"
	err := ErrExample.Errorf("io error", io.EOF)
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))
	fmt.Println(errors.Is(err, io.EOF))

	// Output:
	// formatted error: io error EOF [errific/examples/example_errorf_test.go:15.ExampleErrorf]
	// true
	// true
}
