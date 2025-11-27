package errific

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/leefernandes/errific"
)

// ============================================================================
// MCP Error Serialization Tests
// ============================================================================

func TestToMCPError_WithNilError(t *testing.T) {
	Configure(OutputPretty)

	mcpErr := ToMCPError(nil)
	// ToMCPError returns zero MCPError for nil
	if mcpErr.Code != 0 {
		t.Errorf("Expected code 0, got %d", mcpErr.Code)
	}
	if mcpErr.Message != "" {
		t.Errorf("Expected empty message, got '%s'", mcpErr.Message)
	}
}

func TestToMCPError_WithStandardError(t *testing.T) {
	Configure(OutputPretty)

	stdErr := errors.New("standard error")
	mcpErr := ToMCPError(stdErr)

	if mcpErr.Code != MCPInternalError {
		t.Errorf("Expected code %d, got %d", MCPInternalError, mcpErr.Code)
	}
	if mcpErr.Message != "standard error" {
		t.Errorf("Expected message 'standard error', got '%s'", mcpErr.Message)
	}
}

func TestMCPError_AllCodes(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	codes := []int{
		MCPParseError,
		MCPInvalidRequest,
		MCPMethodNotFound,
		MCPInvalidParams,
		MCPInternalError,
		MCPToolError,
	}

	for _, code := range codes {
		t.Run(fmt.Sprintf("code_%d", code), func(t *testing.T) {
			err := ErrTest.New().WithMCPCode(code)
			mcpErr := ToMCPError(err)

			if mcpErr.Code != code {
				t.Errorf("Expected code %d, got %d", code, mcpErr.Code)
			}

			// Should be JSON serializable
			data, jsonErr := json.Marshal(mcpErr)
			if jsonErr != nil {
				t.Errorf("Failed to marshal MCP error: %v", jsonErr)
			}
			if len(data) == 0 {
				t.Error("Expected non-empty JSON data")
			}
		})
	}
}

func TestMCPCode_Validation(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	t.Run("valid MCP codes accepted", func(t *testing.T) {
		validCodes := []int{
			0,                 // Zero (unset) is valid
			MCPParseError,     // -32700
			MCPInvalidRequest, // -32600
			MCPMethodNotFound, // -32601
			MCPInvalidParams,  // -32602
			MCPInternalError,  // -32603
			MCPToolError,      // -32000
			-32099,            // Server error range max
			-32768,            // Reserved range min
		}

		for _, code := range validCodes {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Valid MCP code %d should not panic, got: %v", code, r)
					}
				}()
				_ = ErrTest.New().WithMCPCode(code)
			}()
		}
	})

	t.Run("invalid MCP codes rejected", func(t *testing.T) {
		invalidCodes := []int{
			1,       // Positive number
			100,     // Way out of range
			-1,      // Just outside range
			-31999,  // Just outside server range
			-32769,  // Just below min
			999999,  // Very large positive
			-999999, // Very large negative
		}

		for _, code := range invalidCodes {
			func() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Invalid MCP code %d should panic but didn't", code)
					} else {
						msg := r.(string)
						if !strings.Contains(msg, "invalid MCP code") {
							t.Errorf("Expected panic message to contain 'invalid MCP code', got: %v", r)
						}
						if !strings.Contains(msg, "JSON-RPC 2.0") {
							t.Errorf("Expected panic message to reference JSON-RPC 2.0 spec, got: %v", r)
						}
					}
				}()
				_ = ErrTest.New().WithMCPCode(code)
			}()
		}
	})

	t.Run("panic message format", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				msg := r.(string)
				// Should contain the invalid code
				if !strings.Contains(msg, "12345") {
					t.Errorf("Panic message should contain the invalid code, got: %s", msg)
				}
				// Should mention the valid range
				if !strings.Contains(msg, "-32768") || !strings.Contains(msg, "-32000") {
					t.Errorf("Panic message should mention valid range, got: %s", msg)
				}
			} else {
				t.Error("Should have panicked for invalid code 12345")
			}
		}()
		_ = ErrTest.New().WithMCPCode(12345)
	})
}

// ============================================================================
// JSON Serialization Tests
// ============================================================================

func TestMarshalJSON_WithAllFields(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "complete error"

	err := ErrTest.New().
		WithCode("ERR_001").
		WithCategory(CategoryServer).
		WithContext(Context{"key": "value"}).
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(500).
		WithMCPCode(MCPToolError).
		WithCorrelationID("corr-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithSessionID("sess-abc").
		WithHelp("help text").
		WithSuggestion("suggestion").
		WithDocs("https://example.com").
		WithTags("tag1", "tag2").
		WithLabels(map[string]string{"k1": "v1"}).
		WithTimestamp(time.Now()).
		WithDuration(100 * time.Millisecond)

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal: %v", jsonErr)
	}

	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify all fields are present
	requiredFields := []string{
		"error", "code", "category", "context", "retryable",
		"retry_after", "max_retries", "http_status", "mcp_code",
		"correlation_id", "request_id", "user_id", "session_id",
		"help", "suggestion", "docs", "tags", "labels", "timestamp", "duration",
	}

	for _, field := range requiredFields {
		if _, ok := decoded[field]; !ok {
			t.Errorf("Expected field '%s' in JSON output", field)
		}
	}
}

func TestMarshalJSON_WithSpecialCharacters(t *testing.T) {
	Configure(OutputPretty)
	var ErrTest Err = "test error"

	// Test with special characters that need escaping
	err := ErrTest.New().
		WithCode("ERR_\"QUOTE\"").
		WithContext(Context{
			"json":    `{"nested": "value"}`,
			"newline": "line1\nline2",
			"tab":     "col1\tcol2",
		}).
		WithHelp("Help with \"quotes\" and \n newlines").
		WithTags("tag-with-\"quotes\"", "tag\nwith\nnewlines")

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal: %v", jsonErr)
	}

	// Should be valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
}
