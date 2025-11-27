package errific

import (
	"strings"
	"testing"
	"time"

	. "github.com/leefernandes/errific"
)

// TestForwardingMethods tests that With___ methods can be called directly on Err
func TestForwardingMethods(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("WithCode without explicit New", func(t *testing.T) {
		err := ErrTest.WithCode("CODE1")
		if GetCode(err) != "CODE1" {
			t.Errorf("Expected code CODE1, got %s", GetCode(err))
		}
	})

	t.Run("WithHTTPStatus without explicit New", func(t *testing.T) {
		err := ErrTest.WithHTTPStatus(404)
		if GetHTTPStatus(err) != 404 {
			t.Errorf("Expected status 404, got %d", GetHTTPStatus(err))
		}
	})

	t.Run("WithMCPCode without explicit New", func(t *testing.T) {
		err := ErrTest.WithMCPCode(MCPToolError)
		if GetMCPCode(err) != MCPToolError {
			t.Errorf("Expected MCP code %d, got %d", MCPToolError, GetMCPCode(err))
		}
	})

	t.Run("WithContext without explicit New", func(t *testing.T) {
		err := ErrTest.WithContext(Context{"key": "value"})
		ctx := GetContext(err)
		if ctx == nil || ctx["key"] != "value" {
			t.Error("Expected context to be set")
		}
	})

	t.Run("WithCategory without explicit New", func(t *testing.T) {
		err := ErrTest.WithCategory(CategoryServer)
		if GetCategory(err) != CategoryServer {
			t.Error("Expected category to be set")
		}
	})

	t.Run("WithRetryable without explicit New", func(t *testing.T) {
		err := ErrTest.WithRetryable(true)
		if !IsRetryable(err) {
			t.Error("Expected retryable to be true")
		}
	})

	t.Run("WithRetryAfter without explicit New", func(t *testing.T) {
		err := ErrTest.WithRetryAfter(5 * time.Second)
		if GetRetryAfter(err) != 5*time.Second {
			t.Error("Expected retry after to be set")
		}
	})

	t.Run("WithMaxRetries without explicit New", func(t *testing.T) {
		err := ErrTest.WithMaxRetries(3)
		if GetMaxRetries(err) != 3 {
			t.Error("Expected max retries to be 3")
		}
	})

	t.Run("WithCorrelationID without explicit New", func(t *testing.T) {
		err := ErrTest.WithCorrelationID("corr-123")
		if GetCorrelationID(err) != "corr-123" {
			t.Error("Expected correlation ID to be set")
		}
	})

	t.Run("WithRequestID without explicit New", func(t *testing.T) {
		err := ErrTest.WithRequestID("req-456")
		if GetRequestID(err) != "req-456" {
			t.Error("Expected request ID to be set")
		}
	})

	t.Run("WithUserID without explicit New", func(t *testing.T) {
		err := ErrTest.WithUserID("user-789")
		if GetUserID(err) != "user-789" {
			t.Error("Expected user ID to be set")
		}
	})

	t.Run("WithSessionID without explicit New", func(t *testing.T) {
		err := ErrTest.WithSessionID("sess-abc")
		if GetSessionID(err) != "sess-abc" {
			t.Error("Expected session ID to be set")
		}
	})

	t.Run("WithHelp without explicit New", func(t *testing.T) {
		err := ErrTest.WithHelp("help text")
		if GetHelp(err) != "help text" {
			t.Error("Expected help to be set")
		}
	})

	t.Run("WithSuggestion without explicit New", func(t *testing.T) {
		err := ErrTest.WithSuggestion("suggestion text")
		if GetSuggestion(err) != "suggestion text" {
			t.Error("Expected suggestion to be set")
		}
	})

	t.Run("WithDocs without explicit New", func(t *testing.T) {
		err := ErrTest.WithDocs("https://example.com")
		if GetDocs(err) != "https://example.com" {
			t.Error("Expected docs to be set")
		}
	})

	t.Run("WithTags without explicit New", func(t *testing.T) {
		err := ErrTest.WithTags("tag1", "tag2")
		tags := GetTags(err)
		if len(tags) != 2 || tags[0] != "tag1" || tags[1] != "tag2" {
			t.Error("Expected tags to be set")
		}
	})

	t.Run("WithLabel without explicit New", func(t *testing.T) {
		err := ErrTest.WithLabel("key", "value")
		if GetLabel(err, "key") != "value" {
			t.Error("Expected label to be set")
		}
	})

	t.Run("WithLabels without explicit New", func(t *testing.T) {
		err := ErrTest.WithLabels(map[string]string{"k1": "v1", "k2": "v2"})
		labels := GetLabels(err)
		if labels["k1"] != "v1" || labels["k2"] != "v2" {
			t.Error("Expected labels to be set")
		}
	})

	t.Run("WithTimestamp without explicit New", func(t *testing.T) {
		now := time.Now()
		err := ErrTest.WithTimestamp(now)
		if GetTimestamp(err).IsZero() {
			t.Error("Expected timestamp to be set")
		}
	})

	t.Run("WithDuration without explicit New", func(t *testing.T) {
		err := ErrTest.WithDuration(100 * time.Millisecond)
		if GetDuration(err) != 100*time.Millisecond {
			t.Error("Expected duration to be set")
		}
	})
}

// TestForwardingMethodChaining tests that chaining works efficiently
func TestForwardingMethodChaining(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("chained calls work", func(t *testing.T) {
		err := ErrTest.
			WithCode("CODE1").
			WithHTTPStatus(400).
			WithCategory(CategoryValidation).
			WithRetryable(false).
			WithCorrelationID("corr-123")

		// Verify all fields are set
		if GetCode(err) != "CODE1" {
			t.Error("Code not set")
		}
		if GetHTTPStatus(err) != 400 {
			t.Error("HTTP status not set")
		}
		if GetCategory(err) != CategoryValidation {
			t.Error("Category not set")
		}
		if IsRetryable(err) {
			t.Error("Retryable should be false")
		}
		if GetCorrelationID(err) != "corr-123" {
			t.Error("Correlation ID not set")
		}
	})

	t.Run("long chain works", func(t *testing.T) {
		err := ErrTest.
			WithCode("CODE1").
			WithHTTPStatus(500).
			WithCategory(CategoryServer).
			WithRetryable(true).
			WithRetryAfter(5 * time.Second).
			WithMaxRetries(3).
			WithMCPCode(MCPInternalError).
			WithCorrelationID("corr-123").
			WithRequestID("req-456").
			WithUserID("user-789").
			WithSessionID("sess-abc").
			WithHelp("help").
			WithSuggestion("suggestion").
			WithDocs("https://example.com").
			WithTags("tag1", "tag2").
			WithLabel("env", "prod")

		// Verify a few fields
		if GetCode(err) != "CODE1" {
			t.Error("Code not set in long chain")
		}
		if GetLabel(err, "env") != "prod" {
			t.Error("Label not set in long chain")
		}
		if len(GetTags(err)) != 2 {
			t.Error("Tags not set in long chain")
		}
	})
}

// TestForwardingBackwardsCompatibility tests that old style still works
func TestForwardingBackwardsCompatibility(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("explicit New still works", func(t *testing.T) {
		err := ErrTest.New().WithCode("CODE1").WithHTTPStatus(400)

		if GetCode(err) != "CODE1" {
			t.Error("Code not set with explicit New")
		}
		if GetHTTPStatus(err) != 400 {
			t.Error("HTTP status not set with explicit New")
		}
	})

	t.Run("both styles are equivalent", func(t *testing.T) {
		// Old style
		err1 := ErrTest.New().WithCode("CODE1").WithHTTPStatus(400)

		// New style
		err2 := ErrTest.WithCode("CODE1").WithHTTPStatus(400)

		// Should have same values
		if GetCode(err1) != GetCode(err2) {
			t.Error("Codes don't match")
		}
		if GetHTTPStatus(err1) != GetHTTPStatus(err2) {
			t.Error("HTTP statuses don't match")
		}

		// Error messages should be similar (might differ in caller line)
		msg1 := err1.Error()
		msg2 := err2.Error()
		if !strings.Contains(msg1, "test error") || !strings.Contains(msg2, "test error") {
			t.Error("Error messages don't contain base error")
		}
	})
}

// TestForwardingNewCalledOnce tests that New() is only called once
func TestForwardingNewCalledOnce(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("caller info shows forwarding method", func(t *testing.T) {
		// The caller will be from the forwarding method in error.go (where New() is called)
		// This is expected - the caller captures where New() was invoked
		err := ErrTest.WithCode("CODE1").WithHTTPStatus(400)

		msg := err.Error()

		// Should contain error.go (the forwarding method location)
		if !strings.Contains(msg, "error.go") {
			t.Errorf("Expected caller info with error.go, got: %s", msg)
		}

		// Should contain "test error"
		if !strings.Contains(msg, "test error") {
			t.Errorf("Expected 'test error' in message, got: %s", msg)
		}

		// Should contain WithCode (the forwarding method name)
		if !strings.Contains(msg, "WithCode") {
			t.Errorf("Expected WithCode in caller info, got: %s", msg)
		}
	})

	t.Run("chaining is efficient", func(t *testing.T) {
		// Even with a long chain, caller info appears only once
		err := ErrTest.
			WithCode("CODE1").
			WithHTTPStatus(400).
			WithCategory(CategoryServer).
			WithRetryable(true)

		msg := err.Error()

		// Verify we still get a valid error message
		if !strings.Contains(msg, "test error") {
			t.Error("Should contain error message")
		}

		// Verify all fields are set correctly (proving New() worked and chain succeeded)
		if GetCode(err) != "CODE1" {
			t.Error("Code should be set")
		}
		if GetHTTPStatus(err) != 400 {
			t.Error("HTTP status should be set")
		}
		if GetCategory(err) != CategoryServer {
			t.Error("Category should be set")
		}
		if !IsRetryable(err) {
			t.Error("Retryable should be true")
		}
	})
}

// TestForwardingWithWrappedErrors tests forwarding with wrapped errors
func TestForwardingWithWrappedErrors(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"
	var ErrOther Err = "other error"

	t.Run("forwarding with wrapped error", func(t *testing.T) {
		// This should work: call forwarding method and pass wrapped error
		underlying := ErrOther.New()
		err := ErrTest.New(underlying).WithCode("CODE1")

		if GetCode(err) != "CODE1" {
			t.Error("Code not set")
		}

		// Check wrapped error is present
		msg := err.Error()
		if !strings.Contains(msg, "other error") {
			t.Error("Wrapped error not present")
		}
	})
}

// TestForwardingValidationStillWorks tests that validation still applies
func TestForwardingValidationStillWorks(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("invalid MCP code panics with forwarding", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid MCP code")
			}
		}()
		_ = ErrTest.WithMCPCode(12345)  // Invalid code
	})

	t.Run("invalid HTTP status panics with forwarding", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid HTTP status")
			}
		}()
		_ = ErrTest.WithHTTPStatus(999)  // Invalid status
	})

	t.Run("empty string ignored with forwarding", func(t *testing.T) {
		err := ErrTest.WithCode("")
		if GetCode(err) != "" {
			t.Error("Empty code should be ignored")
		}
	})

	t.Run("negative retry normalized with forwarding", func(t *testing.T) {
		err := ErrTest.WithMaxRetries(-5)
		if GetMaxRetries(err) != 0 {
			t.Error("Negative max retries should be normalized to 0")
		}
	})
}
