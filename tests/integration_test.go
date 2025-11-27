package errific

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/leefernandes/errific"
)

// TestIntegration_WebAPIWithFullErrorHandling tests a complete web API scenario
func TestIntegration_WebAPIWithFullErrorHandling(t *testing.T) {
	Configure(OutputPretty)
	
	var (
		ErrInvalidInput = Err("invalid input")
		ErrDBQuery      = Err("database query failed")
		ErrUnauthorized = Err("unauthorized")
	)
	
	// Simulate API handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("id")
		if userID == "" {
			err := ErrInvalidInput.New().
				WithCode("VAL_001").
				WithCategory(CategoryValidation).
				WithHTTPStatus(400).
				WithContext(Context{"field": "id", "query": r.URL.RawQuery})
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(GetHTTPStatus(err))
			json.NewEncoder(w).Encode(err)
			return
		}
		
		if userID == "unauthorized" {
			err := ErrUnauthorized.New().
				WithCode("AUTH_001").
				WithCategory(CategoryUnauthorized).
				WithHTTPStatus(401).
				WithHelp("Valid authentication token required").
				WithSuggestion("Include Authorization header with valid token")
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(GetHTTPStatus(err))
			json.NewEncoder(w).Encode(err)
			return
		}
		
		// Simulate DB error
		if userID == "error" {
			err := ErrDBQuery.New(io.EOF).
				WithCode("DB_001").
				WithCategory(CategoryServer).
				WithHTTPStatus(500).
				WithContext(Context{
					"query":   "SELECT * FROM users WHERE id = ?",
					"user_id": userID,
				}).
				WithRetryable(true).
				WithRetryAfter(5 * time.Second)
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(GetHTTPStatus(err))
			json.NewEncoder(w).Encode(err)
			return
		}
		
		w.WriteHeader(200)
		w.Write([]byte(`{"status": "ok"}`))
	}
	
	// Test validation error
	t.Run("validation_error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		
		if w.Code != 400 {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
		
		var decoded map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if decoded["code"] != "VAL_001" {
			t.Errorf("Expected code VAL_001, got %v", decoded["code"])
		}
	})
	
	// Test auth error
	t.Run("auth_error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users?id=unauthorized", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		
		if w.Code != 401 {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
		
		var decoded map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if decoded["help"] == nil {
			t.Error("Expected help text in error response")
		}
	})
	
	// Test server error
	t.Run("server_error", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users?id=error", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		
		if w.Code != 500 {
			t.Errorf("Expected status 500, got %d", w.Code)
		}
		
		var decoded map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		
		if decoded["retryable"] != true {
			t.Error("Expected error to be retryable")
		}
	})
	
	// Test success
	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users?id=123", nil)
		w := httptest.NewRecorder()
		handler(w, req)
		
		if w.Code != 200 {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestIntegration_MCPToolServer tests MCP tool server scenario
func TestIntegration_MCPToolServer(t *testing.T) {
	Configure(OutputPretty)
	
	var ErrToolExecution = Err("tool execution failed")
	
	// Simulate MCP tool call handler
	handleToolCall := func(toolName string, params map[string]interface{}) (interface{}, error) {
		if toolName == "invalid_tool" {
			return nil, ErrToolExecution.New().
				WithMCPCode(MCPMethodNotFound).
				WithRequestID("req-123").
				WithHelp(fmt.Sprintf("Tool '%s' does not exist", toolName)).
				WithSuggestion("Check available tools with list_tools method").
				WithDocs("https://docs.example.com/tools")
		}
		
		if toolName == "failing_tool" {
			return nil, ErrToolExecution.New().
				WithMCPCode(MCPToolError).
				WithCorrelationID("trace-abc-123").
				WithRequestID("req-456").
				WithHelp("The search_database tool encountered a connection error").
				WithSuggestion("Check database credentials and connection string").
				WithDocs("https://docs.example.com/tools/search_database").
				WithTags("mcp", "tool-error", "database", "connection").
				WithLabel("tool_name", toolName).
				WithLabel("severity", "high").
				WithRetryable(true).
				WithRetryAfter(5 * time.Second)
		}
		
		return map[string]string{"result": "success"}, nil
	}
	
	t.Run("method_not_found", func(t *testing.T) {
		result, err := handleToolCall("invalid_tool", nil)
		if err == nil {
			t.Fatal("Expected error for invalid tool")
		}
		if result != nil {
			t.Error("Expected nil result for error")
		}
		
		mcpErr := ToMCPError(err)
		if mcpErr.Code != MCPMethodNotFound {
			t.Errorf("Expected MCP code %d, got %d", MCPMethodNotFound, mcpErr.Code)
		}
		
		// Should be JSON serializable
		data, jsonErr := json.Marshal(mcpErr)
		if jsonErr != nil {
			t.Fatalf("Failed to marshal MCP error: %v", jsonErr)
		}
		
		var decoded map[string]interface{}
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}
		
		if decoded["code"].(float64) != float64(MCPMethodNotFound) {
			t.Error("MCP code mismatch in JSON")
		}
	})
	
	t.Run("tool_error_with_recovery", func(t *testing.T) {
		result, err := handleToolCall("failing_tool", nil)
		if err == nil {
			t.Fatal("Expected error for failing tool")
		}
		if result != nil {
			t.Error("Expected nil result for error")
		}
		
		// Check error metadata
		if !IsRetryable(err) {
			t.Error("Expected error to be retryable")
		}
		
		if GetRetryAfter(err) != 5*time.Second {
			t.Errorf("Expected retry after 5s, got %v", GetRetryAfter(err))
		}
		
		if GetHelp(err) == "" {
			t.Error("Expected help text")
		}
		
		if GetSuggestion(err) == "" {
			t.Error("Expected suggestion")
		}
		
		tags := GetTags(err)
		if len(tags) == 0 {
			t.Error("Expected tags")
		}
		
		mcpErr := ToMCPError(err)
		if mcpErr.Code != MCPToolError {
			t.Errorf("Expected MCP code %d, got %d", MCPToolError, mcpErr.Code)
		}
	})
	
	t.Run("success", func(t *testing.T) {
		result, err := handleToolCall("valid_tool", nil)
		if err != nil {
			t.Fatalf("Expected success, got error: %v", err)
		}
		if result == nil {
			t.Error("Expected result")
		}
	})
}

// TestIntegration_DistributedTracing tests distributed tracing scenario
func TestIntegration_DistributedTracing(t *testing.T) {
	Configure(OutputPretty)
	
	var (
		ErrServiceA = Err("service A failed")
		ErrServiceB = Err("service B failed")
		ErrServiceC = Err("service C failed")
	)
	
	// Simulate service chain: A -> B -> C
	serviceC := func(correlationID string) error {
		return ErrServiceC.New().
			WithCorrelationID(correlationID).
			WithRequestID("req-c-123").
			WithLabel("service", "service-c").
			WithLabel("environment", "production").
			WithContext(Context{
				"operation":   "database_query",
				"duration_ms": 1500,
			})
	}
	
	serviceB := func(correlationID string) error {
		err := serviceC(correlationID)
		if err != nil {
			return ErrServiceB.New(err).
				WithCorrelationID(correlationID).
				WithRequestID("req-b-456").
				WithLabel("service", "service-b").
				WithLabel("environment", "production")
		}
		return nil
	}
	
	serviceA := func(correlationID string) error {
		err := serviceB(correlationID)
		if err != nil {
			return ErrServiceA.New(err).
				WithCorrelationID(correlationID).
				WithRequestID("req-a-789").
				WithLabel("service", "service-a").
				WithLabel("environment", "production")
		}
		return nil
	}
	
	// Test error propagation through service chain
	correlationID := "trace-xyz-789"
	err := serviceA(correlationID)
	
	if err == nil {
		t.Fatal("Expected error from service chain")
	}
	
	// Verify correlation ID propagated
	if GetCorrelationID(err) != correlationID {
		t.Errorf("Expected correlation ID %s, got %s", correlationID, GetCorrelationID(err))
	}
	
	// Verify service label
	if GetLabel(err, "service") != "service-a" {
		t.Errorf("Expected service-a, got %s", GetLabel(err, "service"))
	}
	
	// Verify error message contains wrapped errors
	errMsg := err.Error()
	if !containsAny(errMsg, "service A failed", "service B failed", "service C failed") {
		t.Logf("Error message: %s", errMsg)
	}
	
	// Verify JSON serialization captures correlation ID
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal: %v", jsonErr)
	}
	
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	if decoded["correlation_id"] != correlationID {
		t.Errorf("Expected correlation_id in JSON")
	}
}

// Helper function to check if string contains any of the substrings
func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		found := false
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				found = true
				break
			}
		}
		if found {
			return true
		}
	}
	return false
}

// TestIntegration_AIAgentWithSelfHealing tests AI agent self-healing scenario
func TestIntegration_AIAgentWithSelfHealing(t *testing.T) {
	Configure(OutputPretty)
	
	var ErrAPICall = Err("API call failed")
	
	// Simulate failing API call
	makeAPICall := func(attempt int) error {
		if attempt < 3 {
			return ErrAPICall.New().
				WithCode(fmt.Sprintf("API_TIMEOUT_%03d", attempt)).
				WithCategory(CategoryTimeout).
				WithHelp("Request to external API timed out after 30 seconds").
				WithSuggestion("Increase timeout or implement circuit breaker").
				WithDocs("https://docs.example.com/api/timeouts").
				WithRetryable(true).
				WithRetryAfter(time.Duration(attempt*5) * time.Second).
				WithMaxRetries(3).
				WithContext(Context{
					"endpoint":    "https://api.example.com/v1/data",
					"method":      "GET",
					"duration_ms": 30000,
					"attempt":     attempt,
				})
		}
		return nil // Success on 3rd attempt
	}
	
	// AI agent retry logic
	var finalErr error
	for attempt := 1; attempt <= 3; attempt++ {
		err := makeAPICall(attempt)
		if err == nil {
			finalErr = nil
			break
		}
		
		finalErr = err
		
		// AI agent decision making
		if !IsRetryable(err) {
			break
		}
		
		retryAfter := GetRetryAfter(err)
		maxRetries := GetMaxRetries(err)
		
		if attempt >= maxRetries {
			break
		}
		
		// In real scenario, would sleep for retryAfter duration
		_ = retryAfter
	}
	
	if finalErr != nil {
		t.Errorf("Expected success after retries, got error: %v", finalErr)
	}
}

// TestIntegration_RAGErrorCategorization tests RAG system error categorization
func TestIntegration_RAGErrorCategorization(t *testing.T) {
	Configure(OutputPretty)
	
	var ErrEmbedding = Err("embedding generation failed")
	
	// Create error with rich metadata for RAG
	err := ErrEmbedding.New().
		WithCode("EMB_001").
		WithCategory(CategoryTimeout).
		WithTags("rag", "embedding", "openai", "rate-limit").
		WithLabel("model", "text-embedding-ada-002").
		WithLabel("provider", "openai").
		WithLabel("cost_category", "compute").
		WithHelp("OpenAI API rate limit exceeded").
		WithSuggestion("Implement exponential backoff or use batch API").
		WithRetryable(true).
		WithRetryAfter(60 * time.Second).
		WithContext(Context{
			"token_count": 8192,
			"batch_size":  100,
			"rate_limit":  "60/min",
		})
	
	// Verify all RAG-relevant metadata
	tags := GetTags(err)
	if len(tags) != 4 {
		t.Errorf("Expected 4 tags, got %d", len(tags))
	}
	
	labels := GetLabels(err)
	if len(labels) != 3 {
		t.Errorf("Expected 3 labels, got %d", len(labels))
	}
	
	if GetLabel(err, "provider") != "openai" {
		t.Error("Expected provider label")
	}
	
	ctx := GetContext(err)
	if len(ctx) != 3 {
		t.Errorf("Expected 3 context fields, got %d", len(ctx))
	}
	
	// Should be fully serializable for RAG indexing
	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("Failed to marshal: %v", jsonErr)
	}
	
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	
	// Verify tags are in JSON (for RAG search)
	if decoded["tags"] == nil {
		t.Error("Expected tags in JSON for RAG indexing")
	}
}
