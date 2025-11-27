package errific

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestErrNew(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic error", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New()

		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}

		if !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected error message to contain 'test error', got: %s", err.Error())
		}

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New(io.EOF)

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error message to contain 'EOF', got: %s", err.Error())
		}
	})

	t.Run("with multiple wrapped errors", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Error("expected errors.Is to match io.ErrUnexpectedEOF")
		}
	})
}

func TestErrErrorf(t *testing.T) {
	Configure(OutputPretty)

	t.Run("formatted error", func(t *testing.T) {
		var ErrTest Err = "test error: %s %d"
		err := ErrTest.Errorf("hello", 42)

		if !strings.Contains(err.Error(), "hello") {
			t.Errorf("expected error message to contain 'hello', got: %s", err.Error())
		}

		if !strings.Contains(err.Error(), "42") {
			t.Errorf("expected error message to contain '42', got: %s", err.Error())
		}

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		var ErrTest Err = "test error: %w"
		err := ErrTest.Errorf(io.EOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}
	})
}

func TestErrWithf(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic withf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Withf("detail: %s", "info")

		if !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected error message to contain 'test error', got: %s", err.Error())
		}

		if !strings.Contains(err.Error(), "detail: info") {
			t.Errorf("expected error message to contain 'detail: info', got: %s", err.Error())
		}
	})

	t.Run("chained withf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Withf("first %d", 1).Withf("second %d", 2)

		msg := err.Error()
		if !strings.Contains(msg, "first 1") {
			t.Errorf("expected error message to contain 'first 1', got: %s", msg)
		}

		if !strings.Contains(msg, "second 2") {
			t.Errorf("expected error message to contain 'second 2', got: %s", msg)
		}
	})
}

func TestErrWrapf(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic wrapf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Wrapf("wrapped: %w", io.EOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !strings.Contains(err.Error(), "wrapped") {
			t.Errorf("expected error message to contain 'wrapped', got: %s", err.Error())
		}
	})

	t.Run("chained wrapf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Wrapf("first %d", 1).Wrapf("second %d", 2)

		msg := err.Error()
		if !strings.Contains(msg, "first 1") {
			t.Errorf("expected error message to contain 'first 1', got: %s", msg)
		}

		if !strings.Contains(msg, "second 2") {
			t.Errorf("expected error message to contain 'second 2', got: %s", msg)
		}
	})
}

func TestErrificJoin(t *testing.T) {
	Configure(OutputPretty)

	var ErrTest Err = "test error"
	err := ErrTest.New().Join(io.EOF, io.ErrUnexpectedEOF)

	if !errors.Is(err, io.EOF) {
		t.Error("expected errors.Is to match io.EOF")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Error("expected errors.Is to match io.ErrUnexpectedEOF")
	}
}

func TestConfigureCallerOption(t *testing.T) {
	t.Run("suffix", func(t *testing.T) {
		Configure(OutputPretty, Suffix)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should end with [location]
		if !strings.Contains(msg, "[") || !strings.HasSuffix(msg, "]") {
			t.Errorf("expected suffix format, got: %s", msg)
		}

		if strings.HasPrefix(msg, "[") {
			t.Errorf("expected suffix not prefix, got: %s", msg)
		}
	})

	t.Run("prefix", func(t *testing.T) {
		Configure(OutputPretty, Prefix)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should start with [location]
		if !strings.HasPrefix(msg, "[") {
			t.Errorf("expected prefix format, got: %s", msg)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		Configure(OutputPretty, Disabled)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should not contain brackets
		if strings.Contains(msg, "[") || strings.Contains(msg, "]") {
			t.Errorf("expected no caller info, got: %s", msg)
		}
	})
}

func TestConfigureLayoutOption(t *testing.T) {
	t.Run("newline", func(t *testing.T) {
		Configure(OutputPretty, Newline)

		var ErrTest Err = "test"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)
		msg := err.Error()

		// Should contain newlines
		if !strings.Contains(msg, "\n") {
			t.Errorf("expected newline layout, got: %s", msg)
		}
	})

	t.Run("inline", func(t *testing.T) {
		Configure(OutputPretty, Inline)

		var ErrTest Err = "test"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)
		msg := err.Error()

		// Should contain ↩ symbol
		if !strings.Contains(msg, "↩") {
			t.Errorf("expected inline layout with ↩, got: %s", msg)
		}

		// Should not contain newlines (except maybe in caller/stack)
		lines := strings.Split(msg, "\n")
		if len(lines) > 2 { // Allow for potential stack traces
			t.Errorf("expected inline layout with minimal newlines, got: %s", msg)
		}
	})
}

func TestConfigureWithStack(t *testing.T) {
	t.Run("with stack", func(t *testing.T) {
		Configure(WithStack)

		var ErrTest Err = "test"

		// Create error in a helper function to ensure stack has frames
		err := helperFunctionForStack(ErrTest)
		msg := err.Error()

		// The error message should still be valid
		if !strings.Contains(msg, "test") {
			t.Errorf("expected error message to contain 'test', got: %s", msg)
		}

		// WithStack configuration should not cause errors
		if msg == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("without stack", func(t *testing.T) {
		Configure(OutputPretty) // Default is without stack

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should be a simple error message
		if msg == "" {
			t.Error("expected non-empty error message")
		}

		if !strings.Contains(msg, "test") {
			t.Errorf("expected error message to contain 'test', got: %s", msg)
		}
	})
}

func TestWithStackContents(t *testing.T) {
	Configure(OutputPretty, WithStack)

	t.Run("stack contains expected file and function", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := helperFunctionForStackTrace(ErrTest)
		msg := err.Error()

		// Should contain the helper function name
		if !strings.Contains(msg, "helperFunctionForStackTrace") {
			t.Errorf("expected stack to contain 'helperFunctionForStackTrace', got: %s", msg)
		}

		// Should contain the test file name
		if !strings.Contains(msg, "errific_test.go") {
			t.Errorf("expected stack to contain 'errific_test.go', got: %s", msg)
		}

		// Should NOT contain _testmain.go (it's filtered out)
		if strings.Contains(msg, "_testmain.go") {
			t.Errorf("expected stack to NOT contain '_testmain.go', got: %s", msg)
		}
	})

	t.Run("stack with bubbled errors", func(t *testing.T) {
		var ErrRoot Err = "root error"
		var ErrTop Err = "top error"

		err1 := helperCreateRootError(ErrRoot)
		err2 := helperWrapError(ErrTop, err1)
		msg := err2.Error()

		// Should contain both helper function names
		if !strings.Contains(msg, "helperCreateRootError") {
			t.Errorf("expected stack to contain 'helperCreateRootError', got: %s", msg)
		}

		if !strings.Contains(msg, "helperWrapError") {
			t.Errorf("expected stack to contain 'helperWrapError', got: %s", msg)
		}

		// Should contain file name
		if !strings.Contains(msg, "errific_test.go") {
			t.Errorf("expected stack to contain 'errific_test.go', got: %s", msg)
		}

		// Should NOT contain _testmain.go
		if strings.Contains(msg, "_testmain.go") {
			t.Errorf("expected stack to NOT contain '_testmain.go', got: %s", msg)
		}

		// Should contain both error messages
		if !strings.Contains(msg, "root error") {
			t.Errorf("expected message to contain 'root error', got: %s", msg)
		}

		if !strings.Contains(msg, "top error") {
			t.Errorf("expected message to contain 'top error', got: %s", msg)
		}
	})

	t.Run("stack trace format", func(t *testing.T) {
		var ErrTest Err = "format test"
		err := helperFunctionForStackTrace(ErrTest)
		msg := err.Error()

		// Stack should be indented with spaces
		if !strings.Contains(msg, "\n  ") {
			t.Errorf("expected stack frames to be indented, got: %s", msg)
		}

		// Stack should contain colon separator (file:line.function format)
		lines := strings.Split(msg, "\n")
		foundStackLine := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// Look for lines containing errific_test.go with line number
			if strings.Contains(trimmed, "errific_test.go:") && strings.Contains(trimmed, ".") {
				foundStackLine = true
				// Verify format: file:line.function (e.g., "errific/errific_test.go:350.func3")
				parts := strings.Split(trimmed, ":")
				if len(parts) < 2 {
					t.Errorf("expected stack line to have file:line format, got: %s", line)
				}
			}
		}

		if !foundStackLine {
			t.Errorf("expected to find stack trace line with errific_test.go, got: %s", msg)
		}
	})
}

// Helper functions for stack trace testing
func helperFunctionForStackTrace(e Err) errific {
	return e.New()
}

func helperCreateRootError(e Err) errific {
	return e.New(io.EOF)
}

func helperWrapError(e Err, wrapped error) errific {
	return e.New(wrapped)
}

// Helper function to create errors with a deeper stack
func helperFunctionForStack(e Err) errific {
	return e.New(io.EOF)
}

func TestConfigureTrimPrefixes(t *testing.T) {
	Configure(TrimPrefixes("/usr/local/go/", "/home/user/"))

	var ErrTest Err = "test"
	err := ErrTest.New()
	msg := err.Error()

	// Should not contain the trimmed prefixes
	if strings.Contains(msg, "/usr/local/go/") {
		t.Errorf("expected trimmed prefix, got: %s", msg)
	}

	if strings.Contains(msg, "/home/user/") {
		t.Errorf("expected trimmed prefix, got: %s", msg)
	}
}

func TestConfigureTrimCWD(t *testing.T) {
	Configure(TrimCWD)

	var ErrTest Err = "test"
	err := ErrTest.New()
	msg := err.Error()

	// Should have relative paths
	if !strings.Contains(msg, "errific") {
		t.Errorf("expected relative path, got: %s", msg)
	}
}

func TestConcurrentConfigure(t *testing.T) {
	// Test that concurrent Configure calls don't cause races
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			switch n % 4 {
			case 0:
				Configure(Suffix)
			case 1:
				Configure(Prefix)
			case 2:
				Configure(Disabled)
			case 3:
				Configure(Newline)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentErrorCreation(t *testing.T) {
	Configure(OutputPretty)

	var ErrTest Err = "test"
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := ErrTest.New(io.EOF)
			if !errors.Is(err, ErrTest) {
				t.Error("expected errors.Is to match ErrTest")
			}

			_ = err.Error()
		}()
	}

	wg.Wait()
}

func TestUnwrap(t *testing.T) {
	Configure(OutputPretty)

	var (
		Err1 Err = "error 1"
		Err2 Err = "error 2"
	)

	err1 := Err1.New(io.EOF)
	err2 := Err2.New(err1)

	// Test that unwrap chain works
	if !errors.Is(err2, Err2) {
		t.Error("expected errors.Is to match Err2")
	}

	if !errors.Is(err2, Err1) {
		t.Error("expected errors.Is to match Err1")
	}

	if !errors.Is(err2, io.EOF) {
		t.Error("expected errors.Is to match io.EOF")
	}
}

func TestCircularReferenceFixed(t *testing.T) {
	Configure(OutputPretty)

	var ErrTest Err = "test"
	err := ErrTest.Withf("detail %d", 1)

	// This should not cause infinite loop
	msg := err.Error()

	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// Make sure the error chain is valid
	// errific.Unwrap() returns []error, so we can't use errors.Unwrap
	// Instead, verify that errors.Is works properly
	if !errors.Is(err, ErrTest) {
		t.Error("expected errors.Is to match ErrTest")
	}
}

func BenchmarkErrNew(b *testing.B) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New()
	}
}

func BenchmarkErrNewWithWrap(b *testing.B) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New(io.EOF)
	}
}

func BenchmarkErrError(b *testing.B) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"
	err := ErrTest.New(io.EOF)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrWithStack(b *testing.B) {
	Configure(WithStack)
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New()
	}
}

// ============================================================================
// Phase 1 Feature Tests: Context, Codes, Categories, Retry, JSON
// ============================================================================

func TestWithContext(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic context", func(t *testing.T) {
		var ErrTest Err = "test error"
		ctx := Context{
			"query": "SELECT * FROM users",
			"duration_ms": 1500,
		}
		err := ErrTest.New().WithContext(ctx)

		extractedCtx := GetContext(err)
		if extractedCtx == nil {
			t.Fatal("expected non-nil context")
		}

		if extractedCtx["query"] != "SELECT * FROM users" {
			t.Errorf("expected query in context, got: %v", extractedCtx)
		}

		if extractedCtx["duration_ms"] != 1500 {
			t.Errorf("expected duration_ms in context, got: %v", extractedCtx)
		}
	})

	t.Run("chained context", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().
			WithContext(Context{"key1": "value1"}).
			WithContext(Context{"key2": "value2"})

		ctx := GetContext(err)
		if ctx["key1"] != "value1" {
			t.Error("expected key1 in context")
		}
		if ctx["key2"] != "value2" {
			t.Error("expected key2 in context")
		}
	})

	t.Run("nil context on non-errific error", func(t *testing.T) {
		err := errors.New("standard error")
		ctx := GetContext(err)
		if ctx != nil {
			t.Error("expected nil context for standard error")
		}
	})
}

func TestWithCode(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic code", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithCode("DB_TIMEOUT")

		code := GetCode(err)
		if code != "DB_TIMEOUT" {
			t.Errorf("expected code 'DB_TIMEOUT', got: %s", code)
		}
	})

	t.Run("empty code on non-errific error", func(t *testing.T) {
		err := errors.New("standard error")
		code := GetCode(err)
		if code != "" {
			t.Error("expected empty code for standard error")
		}
	})
}

func TestWithCategory(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic category", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithCategory(CategoryNetwork)

		category := GetCategory(err)
		if category != CategoryNetwork {
			t.Errorf("expected category 'network', got: %s", category)
		}
	})

	t.Run("all categories", func(t *testing.T) {
		categories := []Category{
			CategoryClient,
			CategoryServer,
			CategoryNetwork,
			CategoryValidation,
			CategoryNotFound,
			CategoryUnauthorized,
			CategoryTimeout,
		}

		for _, cat := range categories {
			var ErrTest Err = "test"
			err := ErrTest.New().WithCategory(cat)
			if GetCategory(err) != cat {
				t.Errorf("expected category %s, got: %s", cat, GetCategory(err))
			}
		}
	})
}

func TestRetryMetadata(t *testing.T) {
	Configure(OutputPretty)

	t.Run("retryable", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithRetryable(true)

		if !IsRetryable(err) {
			t.Error("expected error to be retryable")
		}
	})

	t.Run("not retryable", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithRetryable(false)

		if IsRetryable(err) {
			t.Error("expected error to not be retryable")
		}
	})

	t.Run("retry after", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithRetryAfter(5 * time.Second)

		after := GetRetryAfter(err)
		if after != 5*time.Second {
			t.Errorf("expected retry after 5s, got: %v", after)
		}
	})

	t.Run("max retries", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithMaxRetries(3)

		max := GetMaxRetries(err)
		if max != 3 {
			t.Errorf("expected max retries 3, got: %d", max)
		}
	})

	t.Run("complete retry configuration", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().
			WithRetryable(true).
			WithRetryAfter(10 * time.Second).
			WithMaxRetries(5)

		if !IsRetryable(err) {
			t.Error("expected retryable")
		}
		if GetRetryAfter(err) != 10*time.Second {
			t.Error("expected retry after 10s")
		}
		if GetMaxRetries(err) != 5 {
			t.Error("expected max retries 5")
		}
	})
}

func TestWithHTTPStatus(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic http status", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().WithHTTPStatus(503)

		status := GetHTTPStatus(err)
		if status != 503 {
			t.Errorf("expected status 503, got: %d", status)
		}
	})

	t.Run("common http statuses", func(t *testing.T) {
		testCases := []struct {
			status int
			desc   string
		}{
			{400, "Bad Request"},
			{401, "Unauthorized"},
			{403, "Forbidden"},
			{404, "Not Found"},
			{500, "Internal Server Error"},
			{503, "Service Unavailable"},
		}

		for _, tc := range testCases {
			ErrTest := Err(tc.desc)
			err := ErrTest.New().WithHTTPStatus(tc.status)
			if GetHTTPStatus(err) != tc.status {
				t.Errorf("expected status %d for %s", tc.status, tc.desc)
			}
		}
	})
}

func TestJSONSerialization(t *testing.T) {
	Configure(OutputPretty)

	t.Run("basic json", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New()

		jsonBytes, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Fatalf("failed to marshal error: %v", jsonErr)
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(jsonBytes, &result); jsonErr != nil {
			t.Fatalf("failed to unmarshal JSON: %v", jsonErr)
		}

		if result["error"] != "test error" {
			t.Errorf("expected error message in JSON, got: %v", result)
		}
	})

	t.Run("json with all metadata", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New(io.EOF).
			WithCode("TEST_001").
			WithCategory(CategoryServer).
			WithContext(Context{"key": "value"}).
			WithRetryable(true).
			WithRetryAfter(5 * time.Second).
			WithMaxRetries(3).
			WithHTTPStatus(503)

		jsonBytes, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Fatalf("failed to marshal error: %v", jsonErr)
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(jsonBytes, &result); jsonErr != nil {
			t.Fatalf("failed to unmarshal JSON: %v", jsonErr)
		}

		// Check all fields
		if result["code"] != "TEST_001" {
			t.Errorf("expected code in JSON, got: %v", result["code"])
		}

		if result["category"] != "server" {
			t.Errorf("expected category in JSON, got: %v", result["category"])
		}

		if result["retryable"] != true {
			t.Errorf("expected retryable in JSON, got: %v", result["retryable"])
		}

		if result["max_retries"] != float64(3) {
			t.Errorf("expected max_retries in JSON, got: %v", result["max_retries"])
		}

		if result["http_status"] != float64(503) {
			t.Errorf("expected http_status in JSON, got: %v", result["http_status"])
		}

		// Check context
		ctx, ok := result["context"].(map[string]interface{})
		if !ok || ctx["key"] != "value" {
			t.Errorf("expected context in JSON, got: %v", result["context"])
		}

		// Check wrapped errors
		wrapped, ok := result["wrapped"].([]interface{})
		if !ok || len(wrapped) == 0 {
			t.Errorf("expected wrapped errors in JSON, got: %v", result["wrapped"])
		}
	})

	t.Run("json pretty print", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New().
			WithCode("ERR_001").
			WithContext(Context{"request_id": "abc123"})

		jsonBytes, _ := json.MarshalIndent(err, "", "  ")
		jsonStr := string(jsonBytes)

		if !strings.Contains(jsonStr, "ERR_001") {
			t.Error("expected code in pretty JSON")
		}

		if !strings.Contains(jsonStr, "abc123") {
			t.Error("expected request_id in pretty JSON")
		}
	})
}

func TestAIAgentScenario(t *testing.T) {
	Configure(OutputPretty)

	t.Run("database timeout scenario", func(t *testing.T) {
		// Simulate a database timeout error with full AI-agent metadata
		var ErrDBTimeout Err = "database query timeout"
		err := ErrDBTimeout.New(io.EOF).
			WithCode("DB_TIMEOUT_001").
			WithCategory(CategoryNetwork).
			WithContext(Context{
				"query":       "SELECT * FROM large_table",
				"duration_ms": 30000,
				"table":       "large_table",
			}).
			WithRetryable(true).
			WithRetryAfter(5 * time.Second).
			WithMaxRetries(3).
			WithHTTPStatus(503)

		// AI agent can now make decisions
		if IsRetryable(err) {
			retryAfter := GetRetryAfter(err)
			maxRetries := GetMaxRetries(err)
			
			// AI knows to retry after 5 seconds, max 3 times
			if retryAfter != 5*time.Second {
				t.Error("AI should know to wait 5 seconds")
			}
			if maxRetries != 3 {
				t.Error("AI should know max 3 retries")
			}
		} else {
			t.Error("AI should know this is retryable")
		}

		// AI can extract context for logging
		ctx := GetContext(err)
		if ctx["table"] != "large_table" {
			t.Error("AI should know which table failed")
		}

		// AI can respond with correct HTTP status
		status := GetHTTPStatus(err)
		if status != 503 {
			t.Error("AI should return 503 Service Unavailable")
		}

		// AI can serialize for monitoring
		jsonBytes, _ := json.Marshal(err)
		if len(jsonBytes) == 0 {
			t.Error("AI should be able to serialize error")
		}
	})

	t.Run("validation error scenario", func(t *testing.T) {
		var ErrValidation Err = "validation failed"
		err := ErrValidation.New().
			WithCode("VAL_EMAIL_INVALID").
			WithCategory(CategoryValidation).
			WithContext(Context{
				"field": "email",
				"value": "invalid",
			}).
			WithRetryable(false).
			WithHTTPStatus(400)

		// AI knows not to retry validation errors
		if IsRetryable(err) {
			t.Error("AI should not retry validation errors")
		}

		// AI returns 400 Bad Request
		if GetHTTPStatus(err) != 400 {
			t.Error("AI should return 400 for validation")
		}

		// AI can tell user which field failed
		ctx := GetContext(err)
		if ctx["field"] != "email" {
			t.Error("AI should know email field failed")
		}
	})
}

// Phase 2A: MCP & RAG Integration Tests

func TestPhase2A_CorrelationID(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("with correlation ID", func(t *testing.T) {
		err := ErrTest.New().WithCorrelationID("corr-12345")

		id := GetCorrelationID(err)
		if id != "corr-12345" {
			t.Errorf("expected correlation ID 'corr-12345', got '%s'", id)
		}
	})

	t.Run("without correlation ID", func(t *testing.T) {
		err := ErrTest.New()

		id := GetCorrelationID(err)
		if id != "" {
			t.Errorf("expected empty correlation ID, got '%s'", id)
		}
	})

	t.Run("nil error", func(t *testing.T) {
		id := GetCorrelationID(nil)
		if id != "" {
			t.Errorf("expected empty correlation ID for nil, got '%s'", id)
		}
	})
}

func TestPhase2A_RequestID(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().WithRequestID("req-67890")

	id := GetRequestID(err)
	if id != "req-67890" {
		t.Errorf("expected request ID 'req-67890', got '%s'", id)
	}
}

func TestPhase2A_UserID(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().WithUserID("user-123")

	id := GetUserID(err)
	if id != "user-123" {
		t.Errorf("expected user ID 'user-123', got '%s'", id)
	}
}

func TestPhase2A_SessionID(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().WithSessionID("sess-456")

	id := GetSessionID(err)
	if id != "sess-456" {
		t.Errorf("expected session ID 'sess-456', got '%s'", id)
	}
}

func TestPhase2A_Help(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	helpText := "Check your configuration and try again"
	err := ErrTest.New().WithHelp(helpText)

	got := GetHelp(err)
	if got != helpText {
		t.Errorf("expected help text '%s', got '%s'", helpText, got)
	}
}

func TestPhase2A_Suggestion(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	suggestion := "Increase timeout to 30 seconds"
	err := ErrTest.New().WithSuggestion(suggestion)

	got := GetSuggestion(err)
	if got != suggestion {
		t.Errorf("expected suggestion '%s', got '%s'", suggestion, got)
	}
}

func TestPhase2A_Docs(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	docsURL := "https://docs.example.com/errors/timeout"
	err := ErrTest.New().WithDocs(docsURL)

	got := GetDocs(err)
	if got != docsURL {
		t.Errorf("expected docs URL '%s', got '%s'", docsURL, got)
	}
}

func TestPhase2A_Tags(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("with tags", func(t *testing.T) {
		err := ErrTest.New().WithTags("network", "timeout", "retryable")

		tags := GetTags(err)
		if len(tags) != 3 {
			t.Errorf("expected 3 tags, got %d", len(tags))
		}

		expected := []string{"network", "timeout", "retryable"}
		for i, tag := range expected {
			if tags[i] != tag {
				t.Errorf("expected tag[%d] = '%s', got '%s'", i, tag, tags[i])
			}
		}
	})

	t.Run("without tags", func(t *testing.T) {
		err := ErrTest.New()

		tags := GetTags(err)
		if tags != nil {
			t.Errorf("expected nil tags, got %v", tags)
		}
	})
}

func TestPhase2A_Labels(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("with labels map", func(t *testing.T) {
		labels := map[string]string{
			"severity": "high",
			"team":     "backend",
		}
		err := ErrTest.New().WithLabels(labels)

		got := GetLabels(err)
		if len(got) != 2 {
			t.Errorf("expected 2 labels, got %d", len(got))
		}
		if got["severity"] != "high" {
			t.Errorf("expected severity 'high', got '%s'", got["severity"])
		}
		if got["team"] != "backend" {
			t.Errorf("expected team 'backend', got '%s'", got["team"])
		}
	})

	t.Run("with individual label", func(t *testing.T) {
		err := ErrTest.New().WithLabel("env", "production")

		val := GetLabel(err, "env")
		if val != "production" {
			t.Errorf("expected label 'env' = 'production', got '%s'", val)
		}
	})

	t.Run("chained labels", func(t *testing.T) {
		err := ErrTest.New().
			WithLabel("region", "us-east-1").
			WithLabel("instance", "i-12345")

		labels := GetLabels(err)
		if len(labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(labels))
		}
		if labels["region"] != "us-east-1" {
			t.Errorf("expected region 'us-east-1', got '%s'", labels["region"])
		}
		if labels["instance"] != "i-12345" {
			t.Errorf("expected instance 'i-12345', got '%s'", labels["instance"])
		}
	})
}

func TestPhase2A_Timestamp(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	now := time.Now()
	err := ErrTest.New().WithTimestamp(now)

	ts := GetTimestamp(err)
	if !ts.Equal(now) {
		t.Errorf("expected timestamp %v, got %v", now, ts)
	}
}

func TestPhase2A_Duration(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	duration := 5 * time.Second
	err := ErrTest.New().WithDuration(duration)

	d := GetDuration(err)
	if d != duration {
		t.Errorf("expected duration %v, got %v", duration, d)
	}
}

func TestPhase2A_MCPCode(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("with MCP code", func(t *testing.T) {
		err := ErrTest.New().WithMCPCode(MCPInvalidParams)

		code := GetMCPCode(err)
		if code != MCPInvalidParams {
			t.Errorf("expected MCP code %d, got %d", MCPInvalidParams, code)
		}
	})

	t.Run("without MCP code", func(t *testing.T) {
		err := ErrTest.New()

		code := GetMCPCode(err)
		if code != 0 {
			t.Errorf("expected MCP code 0, got %d", code)
		}
	})

	t.Run("nil error", func(t *testing.T) {
		code := GetMCPCode(nil)
		if code != 0 {
			t.Errorf("expected MCP code 0 for nil, got %d", code)
		}
	})
}

func TestPhase2A_ToMCPError(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("with explicit MCP code", func(t *testing.T) {
		err := ErrTest.New().
			WithMCPCode(MCPInvalidParams).
			WithContext(Context{"param": "invalid"})

		var e errific
		if !errors.As(err, &e) {
			t.Fatal("expected errific error")
		}
		mcpErr := e.ToMCPError()

		if mcpErr.Code != MCPInvalidParams {
			t.Errorf("expected code %d, got %d", MCPInvalidParams, mcpErr.Code)
		}
		if mcpErr.Message != "test error" {
			t.Errorf("expected message 'test error', got '%s'", mcpErr.Message)
		}
		if len(mcpErr.Data) == 0 {
			t.Error("expected data to be populated")
		}
	})

	t.Run("without MCP code defaults to internal error", func(t *testing.T) {
		err := ErrTest.New()

		var e errific
		if !errors.As(err, &e) {
			t.Fatal("expected errific error")
		}
		mcpErr := e.ToMCPError()

		if mcpErr.Code != MCPInternalError {
			t.Errorf("expected default code %d, got %d", MCPInternalError, mcpErr.Code)
		}
	})

	t.Run("MCP error is JSON serializable", func(t *testing.T) {
		err := ErrTest.New().WithMCPCode(MCPToolError)

		var e errific
		if !errors.As(err, &e) {
			t.Fatal("expected errific error")
		}
		mcpErr := e.ToMCPError()

		jsonBytes, jsonErr := json.Marshal(mcpErr)
		if jsonErr != nil {
			t.Errorf("failed to marshal MCP error: %v", jsonErr)
		}
		if len(jsonBytes) == 0 {
			t.Error("expected JSON bytes")
		}

		// Verify JSON structure
		var decoded map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
			t.Errorf("failed to unmarshal JSON: %v", err)
		}
		if decoded["code"].(float64) != float64(MCPToolError) {
			t.Errorf("expected code %d in JSON, got %v", MCPToolError, decoded["code"])
		}
	})
}

func TestPhase2A_MCPErrorType(t *testing.T) {
	t.Run("MCPError implements error interface", func(t *testing.T) {
		mcpErr := MCPError{
			Code:    MCPInvalidRequest,
			Message: "invalid request",
		}

		errStr := mcpErr.Error()
		if errStr != "MCP error -32600: invalid request" {
			t.Errorf("unexpected error string: %s", errStr)
		}
	})
}

func TestPhase2A_JSONSerialization(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("Phase 2A fields in JSON", func(t *testing.T) {
		now := time.Now()
		err := ErrTest.New().
			WithCorrelationID("corr-123").
			WithRequestID("req-456").
			WithUserID("user-789").
			WithSessionID("sess-abc").
			WithHelp("Check configuration").
			WithSuggestion("Increase timeout").
			WithDocs("https://docs.example.com").
			WithTags("network", "timeout").
			WithLabel("severity", "high").
			WithTimestamp(now).
			WithDuration(5 * time.Second).
			WithMCPCode(MCPToolError)

		jsonBytes, marshalErr := json.Marshal(err)
		if marshalErr != nil {
			t.Fatalf("failed to marshal: %v", marshalErr)
		}

		var decoded map[string]interface{}
		if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		// Verify Phase 2A fields
		if decoded["correlation_id"] != "corr-123" {
			t.Errorf("expected correlation_id 'corr-123', got '%v'", decoded["correlation_id"])
		}
		if decoded["request_id"] != "req-456" {
			t.Errorf("expected request_id 'req-456', got '%v'", decoded["request_id"])
		}
		if decoded["user_id"] != "user-789" {
			t.Errorf("expected user_id 'user-789', got '%v'", decoded["user_id"])
		}
		if decoded["session_id"] != "sess-abc" {
			t.Errorf("expected session_id 'sess-abc', got '%v'", decoded["session_id"])
		}
		if decoded["help"] != "Check configuration" {
			t.Errorf("expected help text, got '%v'", decoded["help"])
		}
		if decoded["suggestion"] != "Increase timeout" {
			t.Errorf("expected suggestion, got '%v'", decoded["suggestion"])
		}
		if decoded["docs"] != "https://docs.example.com" {
			t.Errorf("expected docs URL, got '%v'", decoded["docs"])
		}
		if decoded["mcp_code"].(float64) != float64(MCPToolError) {
			t.Errorf("expected mcp_code %d, got '%v'", MCPToolError, decoded["mcp_code"])
		}

		// Verify tags
		tags := decoded["tags"].([]interface{})
		if len(tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(tags))
		}

		// Verify labels
		labels := decoded["labels"].(map[string]interface{})
		if labels["severity"] != "high" {
			t.Errorf("expected severity 'high', got '%v'", labels["severity"])
		}

		// Verify timestamp and duration
		if decoded["timestamp"] == nil {
			t.Error("expected timestamp in JSON")
		}
		if decoded["duration"] != "5s" {
			t.Errorf("expected duration '5s', got '%v'", decoded["duration"])
		}
	})
}

func TestPhase2A_MCPIntegration(t *testing.T) {
	Configure(OutputPretty)

	t.Run("MCP tool error scenario", func(t *testing.T) {
		var ErrToolExecution Err = "tool execution failed"

		err := ErrToolExecution.New().
			WithMCPCode(MCPToolError).
			WithCorrelationID("mcp-corr-123").
			WithRequestID("mcp-req-456").
			WithHelp("The tool encountered an error during execution").
			WithSuggestion("Check the tool parameters and retry").
			WithDocs("https://docs.mcp.ai/errors/tool-execution").
			WithTags("mcp", "tool-error", "retryable").
			WithLabel("tool_name", "search_database").
			WithTimestamp(time.Now()).
			WithRetryable(true).
			WithRetryAfter(2 * time.Second)

		// Verify MCP code
		if GetMCPCode(err) != MCPToolError {
			t.Error("expected MCP tool error code")
		}

		// Verify correlation tracking
		if GetCorrelationID(err) != "mcp-corr-123" {
			t.Error("MCP should track correlation ID")
		}

		// Verify recovery guidance
		if GetHelp(err) == "" {
			t.Error("MCP should provide help text")
		}
		if GetSuggestion(err) == "" {
			t.Error("MCP should provide suggestion")
		}

		// Verify RAG semantic tags
		tags := GetTags(err)
		if len(tags) != 3 {
			t.Errorf("expected 3 semantic tags for RAG, got %d", len(tags))
		}

		// Convert to MCP format
		var e errific
		if !errors.As(err, &e) {
			t.Fatal("expected errific error")
		}
		mcpErr := e.ToMCPError()
		if mcpErr.Code != MCPToolError {
			t.Error("MCP error conversion failed")
		}

		// Verify JSON serialization for MCP response
		jsonBytes, _ := json.Marshal(mcpErr)
		if len(jsonBytes) == 0 {
			t.Error("MCP error should be JSON serializable")
		}
	})

	t.Run("MCP invalid params scenario", func(t *testing.T) {
		var ErrInvalidParams Err = "invalid parameters"

		err := ErrInvalidParams.New().
			WithMCPCode(MCPInvalidParams).
			WithContext(Context{
				"expected": "string",
				"received": "number",
				"param":    "query",
			}).
			WithHelp("Parameter 'query' must be a string").
			WithRetryable(false)

		var e errific
		if !errors.As(err, &e) {
			t.Fatal("expected errific error")
		}
		mcpErr := e.ToMCPError()
		if mcpErr.Code != MCPInvalidParams {
			t.Errorf("expected code %d, got %d", MCPInvalidParams, mcpErr.Code)
		}

		// Should not be retryable (param validation errors)
		if IsRetryable(err) {
			t.Error("param validation errors should not be retryable")
		}
	})
}

// Phase 2A: Additional Edge Case Tests

func TestMCPErrorCodeConstants(t *testing.T) {
	// Verify JSON-RPC 2.0 specification compliance
	tests := []struct {
		name string
		code int
		want int
	}{
		{"Parse Error", MCPParseError, -32700},
		{"Invalid Request", MCPInvalidRequest, -32600},
		{"Method Not Found", MCPMethodNotFound, -32601},
		{"Invalid Params", MCPInvalidParams, -32602},
		{"Internal Error", MCPInternalError, -32603},
		{"Tool Error", MCPToolError, -32000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.want {
				t.Errorf("%s: got %d, want %d", tt.name, tt.code, tt.want)
			}
		})
	}
}

func TestToMCPError_EdgeCases(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("nil error returns zero MCPError", func(t *testing.T) {
		mcpErr := ToMCPError(nil)
		if mcpErr.Code != 0 {
			t.Errorf("expected code 0, got %d", mcpErr.Code)
		}
		if mcpErr.Message != "" {
			t.Errorf("expected empty message, got '%s'", mcpErr.Message)
		}
	})

	t.Run("stdlib error uses MCPInternalError", func(t *testing.T) {
		err := errors.New("standard library error")
		mcpErr := ToMCPError(err)

		if mcpErr.Code != MCPInternalError {
			t.Errorf("expected code %d, got %d", MCPInternalError, mcpErr.Code)
		}
		if mcpErr.Message != "standard library error" {
			t.Errorf("expected message 'standard library error', got '%s'", mcpErr.Message)
		}
		if mcpErr.Data != nil {
			t.Error("stdlib errors should not have data field")
		}
	})

	t.Run("errific error with MCP code", func(t *testing.T) {
		err := ErrTest.New().WithMCPCode(MCPInvalidParams)
		mcpErr := ToMCPError(err)

		if mcpErr.Code != MCPInvalidParams {
			t.Errorf("expected code %d, got %d", MCPInvalidParams, mcpErr.Code)
		}
		if len(mcpErr.Data) == 0 {
			t.Error("errific errors should have data field populated")
		}
	})

	t.Run("errific error without MCP code defaults to internal", func(t *testing.T) {
		err := ErrTest.New()
		mcpErr := ToMCPError(err)

		if mcpErr.Code != MCPInternalError {
			t.Errorf("expected default code %d, got %d", MCPInternalError, mcpErr.Code)
		}
	})
}

func TestMCPError_ErrorFormat(t *testing.T) {
	tests := []struct {
		code int
		msg  string
		want string
	}{
		{-32600, "invalid request", "MCP error -32600: invalid request"},
		{-32000, "tool failed", "MCP error -32000: tool failed"},
		{-32603, "internal error", "MCP error -32603: internal error"},
	}

	for _, tt := range tests {
		mcpErr := MCPError{Code: tt.code, Message: tt.msg}
		got := mcpErr.Error()
		if got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
}

func TestPhase2A_LabelMerging(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("WithLabels then WithLabel merges", func(t *testing.T) {
		err := ErrTest.New().
			WithLabels(map[string]string{"a": "1", "b": "2"}).
			WithLabel("c", "3").
			WithLabel("a", "10") // Overwrites 'a'

		labels := GetLabels(err)
		if len(labels) != 3 {
			t.Errorf("expected 3 labels, got %d", len(labels))
		}
		if labels["a"] != "10" {
			t.Errorf("expected a=10, got a=%s", labels["a"])
		}
		if labels["b"] != "2" {
			t.Errorf("expected b=2, got b=%s", labels["b"])
		}
		if labels["c"] != "3" {
			t.Errorf("expected c=3, got c=%s", labels["c"])
		}
	})

	t.Run("multiple WithLabels calls merge", func(t *testing.T) {
		err := ErrTest.New().
			WithLabels(map[string]string{"a": "1"}).
			WithLabels(map[string]string{"b": "2"})

		labels := GetLabels(err)
		if len(labels) != 2 {
			t.Errorf("expected 2 labels, got %d", len(labels))
		}
	})
}

func TestPhase2A_WithLabelsNil(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().WithLabels(nil)
	labels := GetLabels(err)

	// Should not panic, should return nil or empty map
	if len(labels) != 0 {
		t.Errorf("WithLabels(nil) should result in nil or empty map, got %v", labels)
	}
}

func TestPhase2A_EmptyTags(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().WithTags()
	tags := GetTags(err)

	// Empty variadic should result in empty or nil slice
	if len(tags) != 0 {
		t.Errorf("WithTags() should result in nil or empty slice, got %v", tags)
	}
}

func TestPhase2A_HelpersWithStdlibErrors(t *testing.T) {
	err := errors.New("stdlib error")

	// All helpers should return zero values for non-errific errors
	if GetMCPCode(err) != 0 {
		t.Error("GetMCPCode should return 0 for stdlib errors")
	}
	if GetCorrelationID(err) != "" {
		t.Error("GetCorrelationID should return empty for stdlib errors")
	}
	if GetRequestID(err) != "" {
		t.Error("GetRequestID should return empty for stdlib errors")
	}
	if GetUserID(err) != "" {
		t.Error("GetUserID should return empty for stdlib errors")
	}
	if GetSessionID(err) != "" {
		t.Error("GetSessionID should return empty for stdlib errors")
	}
	if GetHelp(err) != "" {
		t.Error("GetHelp should return empty for stdlib errors")
	}
	if GetSuggestion(err) != "" {
		t.Error("GetSuggestion should return empty for stdlib errors")
	}
	if GetDocs(err) != "" {
		t.Error("GetDocs should return empty for stdlib errors")
	}
	if GetTags(err) != nil {
		t.Error("GetTags should return nil for stdlib errors")
	}
	if GetLabels(err) != nil {
		t.Error("GetLabels should return nil for stdlib errors")
	}
	if GetLabel(err, "any") != "" {
		t.Error("GetLabel should return empty for stdlib errors")
	}
	if !GetTimestamp(err).IsZero() {
		t.Error("GetTimestamp should return zero time for stdlib errors")
	}
	if GetDuration(err) != 0 {
		t.Error("GetDuration should return 0 for stdlib errors")
	}
}

func TestPhase2A_TimestampEdgeCases(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("no timestamp returns zero time", func(t *testing.T) {
		err := ErrTest.New()
		ts := GetTimestamp(err)

		if !ts.IsZero() {
			t.Errorf("expected zero time, got %v", ts)
		}
	})

	t.Run("future timestamp accepted", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)
		err := ErrTest.New().WithTimestamp(future)
		ts := GetTimestamp(err)

		if !ts.Equal(future) {
			t.Errorf("expected %v, got %v", future, ts)
		}
	})

	t.Run("past timestamp accepted", func(t *testing.T) {
		past := time.Now().Add(-24 * time.Hour)
		err := ErrTest.New().WithTimestamp(past)
		ts := GetTimestamp(err)

		if !ts.Equal(past) {
			t.Errorf("expected %v, got %v", past, ts)
		}
	})
}

func TestPhase2A_DurationEdgeCases(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("no duration returns zero", func(t *testing.T) {
		err := ErrTest.New()
		d := GetDuration(err)

		if d != 0 {
			t.Errorf("expected 0, got %v", d)
		}
	})

	t.Run("zero duration", func(t *testing.T) {
		err := ErrTest.New().WithDuration(0)
		d := GetDuration(err)

		if d != 0 {
			t.Errorf("expected 0, got %v", d)
		}
	})

	t.Run("negative duration accepted", func(t *testing.T) {
		negative := -5 * time.Second
		err := ErrTest.New().WithDuration(negative)
		d := GetDuration(err)

		if d != negative {
			t.Errorf("expected %v, got %v", negative, d)
		}
	})
}

func TestPhase2A_SpecialCharacters(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().
		WithHelp("Help with \"quotes\" and \n newlines").
		WithSuggestion("Suggestion with <html> & special chars").
		WithDocs("https://example.com?param=value&foo=bar").
		WithCorrelationID("id-with-dashes-123")

	// Verify JSON serialization handles special chars
	jsonBytes, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("failed to marshal: %v", marshalErr)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify special characters are preserved
	if !strings.Contains(decoded["help"].(string), "\"quotes\"") {
		t.Error("quotes should be preserved")
	}
	if !strings.Contains(decoded["suggestion"].(string), "<html>") {
		t.Error("HTML should be preserved")
	}
	if !strings.Contains(decoded["docs"].(string), "?param=value&foo=bar") {
		t.Error("URL parameters should be preserved")
	}
}

func TestPhase2A_JSONZeroValues(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	// Create error with no Phase 2A fields set
	err := ErrTest.New()

	jsonBytes, _ := json.Marshal(err)
	var decoded map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Verify omitempty works - Phase 2A fields should not appear
	phase2AFields := []string{
		"mcp_code", "correlation_id", "request_id", "user_id", "session_id",
		"help", "suggestion", "docs", "tags", "labels", "timestamp", "duration",
	}

	for _, field := range phase2AFields {
		if _, ok := decoded[field]; ok {
			t.Errorf("zero value field %q should be omitted from JSON", field)
		}
	}
}

func TestPhase2A_ChainAllMethods(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	now := time.Now()
	duration := 5 * time.Second

	err := ErrTest.New().
		WithMCPCode(MCPToolError).
		WithCorrelationID("corr").
		WithRequestID("req").
		WithUserID("user").
		WithSessionID("sess").
		WithHelp("help text").
		WithSuggestion("suggestion text").
		WithDocs("https://docs.example.com").
		WithTags("tag1", "tag2").
		WithLabels(map[string]string{"a": "1"}).
		WithLabel("b", "2").
		WithTimestamp(now).
		WithDuration(duration)

	// Verify all fields are set
	if GetMCPCode(err) != MCPToolError {
		t.Error("MCP code not set")
	}
	if GetCorrelationID(err) != "corr" {
		t.Error("CorrelationID not set")
	}
	if GetRequestID(err) != "req" {
		t.Error("RequestID not set")
	}
	if GetUserID(err) != "user" {
		t.Error("UserID not set")
	}
	if GetSessionID(err) != "sess" {
		t.Error("SessionID not set")
	}
	if GetHelp(err) != "help text" {
		t.Error("Help not set")
	}
	if GetSuggestion(err) != "suggestion text" {
		t.Error("Suggestion not set")
	}
	if GetDocs(err) != "https://docs.example.com" {
		t.Error("Docs not set")
	}
	if len(GetTags(err)) != 2 {
		t.Error("Tags not set")
	}
	if len(GetLabels(err)) != 2 {
		t.Error("Labels not set")
	}
	if !GetTimestamp(err).Equal(now) {
		t.Error("Timestamp not set")
	}
	if GetDuration(err) != duration {
		t.Error("Duration not set")
	}
}

func TestPhase2A_LabelKeyEdgeCases(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	err := ErrTest.New().
		WithLabel("", "empty_key").
		WithLabel("key with spaces", "value1").
		WithLabel("key-with-dashes", "value2").
		WithLabel("key.with.dots", "value3")

	labels := GetLabels(err)

	if len(labels) != 4 {
		t.Errorf("expected 4 labels, got %d", len(labels))
	}

	if labels[""] != "empty_key" {
		t.Error("empty key should be stored")
	}
	if labels["key with spaces"] != "value1" {
		t.Error("keys with spaces should be stored")
	}
	if labels["key-with-dashes"] != "value2" {
		t.Error("keys with dashes should be stored")
	}
	if labels["key.with.dots"] != "value3" {
		t.Error("keys with dots should be stored")
	}
}

func BenchmarkWithContext(b *testing.B) {
	Configure(OutputPretty)
	var ErrTest Err = "test"
	ctx := Context{"key": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New().WithContext(ctx)
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	Configure(OutputPretty)
	var ErrTest Err = "test"
	err := ErrTest.New().
		WithCode("ERR_001").
		WithCategory(CategoryServer).
		WithContext(Context{"key": "value"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(err)
	}
}
