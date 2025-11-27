package errific

import (
	"math"
	"strings"
	"testing"
	"time"

	. "github.com/leefernandes/errific"
)

// ============================================================================
// HTTP Status Validation Tests
// ============================================================================

func TestHTTPStatus_Validation(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("valid HTTP status codes accepted", func(t *testing.T) {
		validCodes := []int{
			0,   // Zero (unset) is valid
			100, // Informational
			200, // Success
			300, // Redirection
			400, // Client error
			404, // Not found
			500, // Server error
			503, // Service unavailable
			599, // Max valid code
		}

		for _, code := range validCodes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Valid HTTP status %d should not panic, got: %v", code, r)
					}
				}()
				err := ErrTest.New().WithHTTPStatus(code)
				if GetHTTPStatus(err) != code {
					t.Errorf("Expected HTTP status %d, got %d", code, GetHTTPStatus(err))
				}
			}()
		}
	})

	t.Run("invalid HTTP status codes rejected", func(t *testing.T) {
		invalidCodes := []int{
			-1,   // Negative
			99,   // Below minimum
			600,  // Above maximum
			999,  // Way out of range
			1000, // Large number
		}

		for _, code := range invalidCodes {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Invalid HTTP status %d should panic but didn't", code)
					} else {
						msg := r.(string)
						if !strings.Contains(msg, "invalid HTTP status code") {
							t.Errorf("Expected panic message to contain 'invalid HTTP status code', got: %v", r)
						}
					}
				}()
				_ = ErrTest.New().WithHTTPStatus(code)
			}()
		}
	})

	t.Run("HTTP status boundaries", func(t *testing.T) {
		// Test exact boundaries
		validBoundaries := []int{100, 599}
		for _, status := range validBoundaries {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Boundary value %d should be valid, got panic: %v", status, r)
					}
				}()
				_ = ErrTest.New().WithHTTPStatus(status)
			}()
		}

		// Test just outside boundaries
		invalidBoundaries := []int{99, 600}
		for _, status := range invalidBoundaries {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Value %d just outside boundary should panic", status)
					}
				}()
				_ = ErrTest.New().WithHTTPStatus(status)
			}()
		}
	})
}

// ============================================================================
// Retry Metadata Validation Tests
// ============================================================================

func TestMaxRetries_Validation(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("non-negative values accepted", func(t *testing.T) {
		validValues := []int{0, 1, 3, 10, 100, 1000}

		for _, val := range validValues {
			err := ErrTest.New().WithMaxRetries(val)
			if GetMaxRetries(err) != val {
				t.Errorf("Expected max retries %d, got %d", val, GetMaxRetries(err))
			}
		}
	})

	t.Run("negative values treated as zero", func(t *testing.T) {
		negativeValues := []int{-1, -5, -100, -999}

		for _, val := range negativeValues {
			err := ErrTest.New().WithMaxRetries(val)
			result := GetMaxRetries(err)
			if result != 0 {
				t.Errorf("Negative max retries %d should be treated as 0, got %d", val, result)
			}
		}
	})

	t.Run("very large negative retry values", func(t *testing.T) {
		err := ErrTest.New().WithMaxRetries(math.MinInt)
		if GetMaxRetries(err) != 0 {
			t.Errorf("MinInt should be normalized to 0, got %d", GetMaxRetries(err))
		}
	})

	t.Run("very large positive values are preserved", func(t *testing.T) {
		largeRetries := math.MaxInt
		err := ErrTest.New().WithMaxRetries(largeRetries)
		if GetMaxRetries(err) != largeRetries {
			t.Errorf("Large positive retries should be preserved, got %d", GetMaxRetries(err))
		}
	})
}

func TestRetryAfter_Validation(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("non-negative durations accepted", func(t *testing.T) {
		validDurations := []time.Duration{
			0,
			time.Millisecond,
			time.Second,
			5 * time.Second,
			time.Minute,
			time.Hour,
		}

		for _, dur := range validDurations {
			err := ErrTest.New().WithRetryAfter(dur)
			if GetRetryAfter(err) != dur {
				t.Errorf("Expected retry after %v, got %v", dur, GetRetryAfter(err))
			}
		}
	})

	t.Run("negative durations treated as zero", func(t *testing.T) {
		negativeDurations := []time.Duration{
			-1,
			-time.Millisecond,
			-time.Second,
			-5 * time.Second,
			-time.Minute,
		}

		for _, dur := range negativeDurations {
			err := ErrTest.New().WithRetryAfter(dur)
			result := GetRetryAfter(err)
			if result != 0 {
				t.Errorf("Negative duration %v should be treated as 0, got %v", dur, result)
			}
		}
	})

	t.Run("very large negative duration", func(t *testing.T) {
		err := ErrTest.New().WithRetryAfter(time.Duration(math.MinInt64))
		if GetRetryAfter(err) != 0 {
			t.Errorf("MinInt64 duration should be normalized to 0, got %v", GetRetryAfter(err))
		}
	})
}

// ============================================================================
// Empty String Validation Tests
// ============================================================================

func TestEmptyString_Validation(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("empty code ignored", func(t *testing.T) {
		err := ErrTest.New().WithCode("")
		if GetCode(err) != "" {
			t.Error("Empty code should be ignored")
		}
	})

	t.Run("empty correlation ID ignored", func(t *testing.T) {
		err := ErrTest.New().WithCorrelationID("")
		if GetCorrelationID(err) != "" {
			t.Error("Empty correlation ID should be ignored")
		}
	})

	t.Run("empty request ID ignored", func(t *testing.T) {
		err := ErrTest.New().WithRequestID("")
		if GetRequestID(err) != "" {
			t.Error("Empty request ID should be ignored")
		}
	})

	t.Run("empty user ID ignored", func(t *testing.T) {
		err := ErrTest.New().WithUserID("")
		if GetUserID(err) != "" {
			t.Error("Empty user ID should be ignored")
		}
	})

	t.Run("empty session ID ignored", func(t *testing.T) {
		err := ErrTest.New().WithSessionID("")
		if GetSessionID(err) != "" {
			t.Error("Empty session ID should be ignored")
		}
	})

	t.Run("empty help ignored", func(t *testing.T) {
		err := ErrTest.New().WithHelp("")
		if GetHelp(err) != "" {
			t.Error("Empty help should be ignored")
		}
	})

	t.Run("empty suggestion ignored", func(t *testing.T) {
		err := ErrTest.New().WithSuggestion("")
		if GetSuggestion(err) != "" {
			t.Error("Empty suggestion should be ignored")
		}
	})

	t.Run("empty docs URL ignored", func(t *testing.T) {
		err := ErrTest.New().WithDocs("")
		if GetDocs(err) != "" {
			t.Error("Empty docs URL should be ignored")
		}
	})

	t.Run("non-empty values still set", func(t *testing.T) {
		err := ErrTest.New().
			WithCode("TEST_001").
			WithCorrelationID("corr-123").
			WithRequestID("req-456").
			WithUserID("user-789").
			WithSessionID("sess-abc").
			WithHelp("Help text").
			WithSuggestion("Suggestion text").
			WithDocs("https://example.com")

		if GetCode(err) != "TEST_001" {
			t.Error("Non-empty code should be set")
		}
		if GetCorrelationID(err) != "corr-123" {
			t.Error("Non-empty correlation ID should be set")
		}
		if GetRequestID(err) != "req-456" {
			t.Error("Non-empty request ID should be set")
		}
		if GetUserID(err) != "user-789" {
			t.Error("Non-empty user ID should be set")
		}
		if GetSessionID(err) != "sess-abc" {
			t.Error("Non-empty session ID should be set")
		}
		if GetHelp(err) != "Help text" {
			t.Error("Non-empty help should be set")
		}
		if GetSuggestion(err) != "Suggestion text" {
			t.Error("Non-empty suggestion should be set")
		}
		if GetDocs(err) != "https://example.com" {
			t.Error("Non-empty docs URL should be set")
		}
	})
}

// ============================================================================
// Chained Method Call Tests
// ============================================================================

func TestChainedMethods_LastWins(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("multiple WithMCPCode calls - last wins", func(t *testing.T) {
		err := ErrTest.New().
			WithMCPCode(MCPInternalError).
			WithMCPCode(MCPToolError)

		if GetMCPCode(err) != MCPToolError {
			t.Errorf("Expected last MCP code to win, got %d", GetMCPCode(err))
		}
	})

	t.Run("multiple WithCode calls - last wins", func(t *testing.T) {
		err := ErrTest.New().
			WithCode("CODE1").
			WithCode("CODE2")

		if GetCode(err) != "CODE2" {
			t.Errorf("Expected 'CODE2', got %s", GetCode(err))
		}
	})

	t.Run("empty string doesn't override previous value", func(t *testing.T) {
		err := ErrTest.New().
			WithCode("CODE1").
			WithCode("") // Should be ignored

		if GetCode(err) != "CODE1" {
			t.Errorf("Expected 'CODE1' to be preserved, got %s", GetCode(err))
		}
	})
}

// ============================================================================
// Boundary Value Tests
// ============================================================================

func TestBoundaryValues_Extremes(t *testing.T) {
	Configure()
	var ErrTest Err = "test error"

	t.Run("MCP code boundaries", func(t *testing.T) {
		// Test exact boundaries
		validBoundaries := []int{-32768, -32000}
		for _, code := range validBoundaries {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Boundary value %d should be valid, got panic: %v", code, r)
					}
				}()
				_ = ErrTest.New().WithMCPCode(code)
			}()
		}

		// Test just outside boundaries
		invalidBoundaries := []int{-32769, -31999}
		for _, code := range invalidBoundaries {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Value %d just outside boundary should panic", code)
					}
				}()
				_ = ErrTest.New().WithMCPCode(code)
			}()
		}
	})

	t.Run("MaxInt and MinInt values", func(t *testing.T) {
		// MaxInt should panic for both MCP and HTTP
		defer func() {
			if r := recover(); r == nil {
				t.Error("MaxInt should panic")
			}
		}()
		_ = ErrTest.New().WithMCPCode(math.MaxInt)
	})

	t.Run("MinInt should panic for MCP", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MinInt should panic for MCP code")
			}
		}()
		_ = ErrTest.New().WithMCPCode(math.MinInt)
	})
}
