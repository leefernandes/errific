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
	// Stack trace is appended but we don't test exact output here
	// See TestWithStack for detailed stack trace validation
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
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

	// Stack trace is appended but we don't test exact output here
	// See TestWithStackBubbled for detailed stack trace validation
	fmt.Println(errors.Is(err5, ErrRoot))

	// Output:
	// true
}
