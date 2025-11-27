package examples

import (
	"errors"
	"fmt"
	"os"

	. "github.com/leefernandes/errific"
)

func ExampleTrimPrefixes() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	Configure(TrimPrefixes(wd + "/"))
	var ErrExample Err = "example error"
	err = ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [example_trimprefixes_test.go:18.ExampleTrimPrefixes]
	// true
}

func ExampleTrimCWD() {
	Configure(TrimCWD)
	var ErrExample Err = "example error"
	err := ErrExample.New()
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrExample))

	// Output:
	// example error [example_trimprefixes_test.go:30.ExampleTrimCWD]
	// true
}
