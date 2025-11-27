package examples

import (
	"fmt"
	"github.com/leefernandes/errific"
)

// Example_outputFormats demonstrates the different output formats available.
func Example_outputFormats() {
	var ErrUserNotFound errific.Err = "user not found"

	// Pretty format (default) - shows all metadata
	errific.Configure(errific.OutputPretty, errific.VerbosityFull)
	err := ErrUserNotFound.
		WithCode("USER_404").
		WithCategory(errific.CategoryNotFound).
		WithContext(errific.Context{
			"user_id": "user-123",
			"source":  "database",
		}).
		WithHTTPStatus(404)

	fmt.Println("Pretty format:")
	fmt.Println(err)

	// Minimal verbosity
	errific.Configure(errific.VerbosityMinimal)
	err2 := ErrUserNotFound.
		WithCode("USER_404").
		WithContext(errific.Context{"user_id": "user-123"})

	fmt.Println("\nMinimal verbosity:")
	fmt.Println(err2)

	// JSON format
	errific.Configure(errific.OutputJSON)
	err3 := ErrUserNotFound.
		WithCode("USER_404").
		WithContext(errific.Context{"user_id": "user-123"})

	fmt.Println("\nJSON format:")
	fmt.Println(err3)

	// Compact format
	errific.Configure(errific.OutputCompact)
	err4 := ErrUserNotFound.
		WithCode("USER_404").
		WithContext(errific.Context{"user_id": "user-123"})

	fmt.Println("\nCompact format:")
	fmt.Println(err4)

	// Output varies based on file paths and caller info
}
