package examples

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	. "github.com/leefernandes/errific"
)

func ExampleContext() {
	Configure(OutputPretty, VerbosityMinimal)
	// Add structured context to errors for better debugging
	var ErrDatabaseQuery Err = "database query failed"
	err := ErrDatabaseQuery.New(io.EOF).WithContext(Context{
		"query":       "SELECT * FROM users WHERE id = ?",
		"duration_ms": 1500,
		"table":       "users",
	})

	// Extract context for logging
	ctx := GetContext(err)
	fmt.Println("Error:", err)
	fmt.Printf("Table: %v\n", ctx["table"])
	fmt.Printf("Duration: %v ms\n", ctx["duration_ms"])

	// Output:
	// Error: database query failed [errific/examples/example_phase1_test.go:17.ExampleContext]
	// EOF
	// Table: users
	// Duration: 1500 ms
}

func Example_errorCode() {
	Configure(OutputPretty)
	// Use error codes for machine-readable identification
	var ErrAPITimeout Err = "API request timeout"
	err := ErrAPITimeout.New().
		WithCode("API_TIMEOUT_001").
		WithCategory(CategoryTimeout)

	code := GetCode(err)
	category := GetCategory(err)

	fmt.Printf("Code: %s\n", code)
	fmt.Printf("Category: %s\n", category)
	fmt.Println(errors.Is(err, ErrAPITimeout))

	// Output:
	// Code: API_TIMEOUT_001
	// Category: timeout
	// true
}

func Example_retryable() {
	Configure(OutputPretty)
	// Mark errors as retryable with suggested retry strategy
	var ErrRateLimit Err = "rate limit exceeded"
	err := ErrRateLimit.New().
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3)

	// AI agents can now automate retry logic
	if IsRetryable(err) {
		retryAfter := GetRetryAfter(err)
		maxRetries := GetMaxRetries(err)
		fmt.Printf("Retryable: %v\n", IsRetryable(err))
		fmt.Printf("Retry after: %v\n", retryAfter)
		fmt.Printf("Max retries: %d\n", maxRetries)
	}

	// Output:
	// Retryable: true
	// Retry after: 5s
	// Max retries: 3
}

func Example_httpStatus() {
	Configure(OutputPretty)
	// Set HTTP status codes for automatic response handling
	var ErrValidation Err = "validation failed"
	err := ErrValidation.New().
		WithCode("VAL_001").
		WithCategory(CategoryValidation).
		WithHTTPStatus(400)

	status := GetHTTPStatus(err)
	fmt.Printf("HTTP Status: %d\n", status)
	fmt.Printf("Category: %s\n", GetCategory(err))

	// Output:
	// HTTP Status: 400
	// Category: validation
}

func Example_json() {
	Configure(OutputPretty)
	// Serialize errors to JSON for logging and APIs
	var ErrDatabase Err = "database connection failed"
	err := ErrDatabase.New(io.EOF).
		WithCode("DB_CONN_001").
		WithCategory(CategoryNetwork).
		WithContext(Context{
			"host": "localhost",
			"port": 5432,
		}).
		WithRetryable(true).
		WithHTTPStatus(503)

	// Marshal to JSON
	jsonBytes, _ := json.MarshalIndent(err, "", "  ")
	fmt.Println(string(jsonBytes))

	// Output:
	// {
	//   "error": "database connection failed",
	//   "code": "DB_CONN_001",
	//   "category": "network",
	//   "caller": "errific/examples/example_phase1_test.go:103.Example_json",
	//   "context": {
	//     "host": "localhost",
	//     "port": 5432
	//   },
	//   "retryable": true,
	//   "http_status": 503,
	//   "wrapped": [
	//     "EOF"
	//   ]
	// }
}

func Example_aiAgentScenario() {
	Configure(OutputPretty)
	// Complete example for AI agent automated error handling
	var ErrServiceCall Err = "external service call failed"
	err := ErrServiceCall.New().
		WithCode("SVC_TIMEOUT").
		WithCategory(CategoryTimeout).
		WithContext(Context{
			"service":     "payment-api",
			"endpoint":    "/v1/charge",
			"request_id":  "req_abc123",
			"duration_ms": 30000,
		}).
		WithRetryable(true).
		WithRetryAfter(10 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(504)

	// AI agent decision logic
	fmt.Printf("Error Code: %s\n", GetCode(err))
	fmt.Printf("Category: %s\n", GetCategory(err))
	fmt.Printf("Should retry: %v\n", IsRetryable(err))
	fmt.Printf("Retry after: %v\n", GetRetryAfter(err))
	fmt.Printf("HTTP Status: %d\n", GetHTTPStatus(err))

	ctx := GetContext(err)
	fmt.Printf("Service: %v\n", ctx["service"])

	// Output:
	// Error Code: SVC_TIMEOUT
	// Category: timeout
	// Should retry: true
	// Retry after: 10s
	// HTTP Status: 504
	// Service: payment-api
}

func Example_chainedMethods() {
	Configure(OutputPretty)
	// Chain all Phase 1 methods together
	var ErrProcessing Err = "processing failed"
	err := ErrProcessing.New(io.EOF).
		WithCode("PROC_001").
		WithCategory(CategoryServer).
		WithContext(Context{
			"file":  "data.csv",
			"line":  42,
			"bytes": 1024,
		}).
		WithRetryable(false).
		WithHTTPStatus(500)

	fmt.Printf("Code: %s\n", GetCode(err))
	fmt.Printf("Retryable: %v\n", IsRetryable(err))

	ctx := GetContext(err)
	fmt.Printf("File: %v\n", ctx["file"])
	fmt.Printf("Line: %v\n", ctx["line"])

	// Output:
	// Code: PROC_001
	// Retryable: false
	// File: data.csv
	// Line: 42
}
