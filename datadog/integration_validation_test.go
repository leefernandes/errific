package datadog

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/leefernandes/errific"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// TestDatadogIntegration_SpanTagsValidation validates all Datadog span tags
// are correctly mapped from errific errors according to Datadog APM conventions
func TestDatadogIntegration_SpanTagsValidation(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	// Create comprehensive errific error
	var ErrTest errific.Err = "test operation failed"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-abc-123").
		WithRequestID("req-xyz-456").
		WithUserID("user-789").
		WithSessionID("sess-def-012").
		WithRetryable(true).
		WithRetryAfter(10 * time.Second).
		WithMaxRetries(5).
		WithHTTPStatus(503).
		WithTags("database", "timeout", "critical").
		WithLabel("service", "user-service").
		WithLabel("region", "us-west-2").
		WithLabel("team", "platform").
		WithContext(errific.Context{
			"query":        "SELECT * FROM users WHERE id = ?",
			"duration_ms":  1500,
			"timeout_ms":   1000,
			"affected_ids": []int{1, 2, 3},
		})

	span := tracer.StartSpan("test.operation")
	RecordError(span, err)

	// Validate span was finished
	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	finishedSpan := spans[0]

	// Define expected tag mappings per Datadog conventions
	expectedTags := map[string]interface{}{
		// Standard Datadog error tags
		"error.msg":  "test operation failed [errific/datadog/integration_validation_test.go:16.TestDatadogIntegration_SpanTagsValidation]",
		"error.type": "errific.errific",

		// errific error metadata → Datadog tags
		"error.code":       "TEST_001",
		"error.category":   "server",
		"correlation.id":   "corr-abc-123",
		"request.id":       "req-xyz-456",
		"user.id":          "user-789",
		"session.id":       "sess-def-012",
		"error.retryable":  true,
		"error.retry_after": "10s",
		"error.max_retries": 5,
		"http.status_code":  503,

		// Tags array → numbered tags
		"error.tag.0": "database",
		"error.tag.1": "timeout",
		"error.tag.2": "critical",

		// Labels → prefixed tags
		"label.service": "user-service",
		"label.region":  "us-west-2",
		"label.team":    "platform",

		// Context → prefixed tags (converted to strings)
		"context.query":        "SELECT * FROM users WHERE id = ?",
		"context.duration_ms":  "1500",
		"context.timeout_ms":   "1000",
		"context.affected_ids": "[1 2 3]",
	}

	// Validate each expected tag
	for key, expected := range expectedTags {
		actual := finishedSpan.Tag(key)
		
		// Special handling for error.msg which includes caller info
		if key == "error.msg" {
			if actual == nil {
				t.Errorf("tag %q not found", key)
				continue
			}
			actualStr, ok := actual.(string)
			if !ok {
				t.Errorf("tag %q has wrong type: %T", key, actual)
				continue
			}
			expectedStr, _ := expected.(string)
			if actualStr != expectedStr {
				// Check if it at least contains the base message
				if len(actualStr) == 0 || !contains(actualStr, "test operation failed") {
					t.Errorf("tag %q = %q, should contain 'test operation failed'", key, actualStr)
				}
			}
			continue
		}

		// Use fmt.Sprint for comparison to handle type differences
		if fmt.Sprint(actual) != fmt.Sprint(expected) {
			t.Errorf("tag %q = %v (type %T), want %v (type %T)", key, actual, actual, expected, expected)
		}
	}

	t.Logf("✅ All %d Datadog span tags validated", len(expectedTags))
}

// TestDatadogIntegration_LogEntryStructure validates the log entry structure
// matches Datadog's reserved attributes and JSON schema requirements
func TestDatadogIntegration_LogEntryStructure(t *testing.T) {
	var ErrTest errific.Err = "database connection failed"
	err := ErrTest.New().
		WithCode("DB_CONN_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("trace-abc-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithHTTPStatus(500).
		WithContext(errific.Context{
			"pool_size":   10,
			"retry_count": 3,
		})

	logEntry := ToLogEntry(err)
	SetServiceInfo(logEntry, "test-service", "testing", "1.0.0")

	// Serialize to JSON
	jsonBytes, jsonErr := json.MarshalIndent(logEntry, "", "  ")
	if jsonErr != nil {
		t.Fatalf("JSON marshal failed: %v", jsonErr)
	}

	// Deserialize to validate structure
	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(jsonBytes, &result); unmarshalErr != nil {
		t.Fatalf("JSON unmarshal failed: %v", unmarshalErr)
	}

	// Datadog reserved attributes that MUST be present
	requiredFields := []string{
		"timestamp",        // ISO 8601 timestamp
		"message",          // Log message
		"level",            // Log level
		"status",           // Status (error/ok)
		"service",          // Service name
		"env",              // Environment
		"version",          // Version
		"dd.trace_id",      // Trace ID (for correlation)
		"error.code",       // Error code
		"error.category",   // Error category
		"correlation.id",   // Correlation ID
		"request.id",       // Request ID
		"user.id",          // User ID
		"http.status_code", // HTTP status
	}

	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("❌ Missing required Datadog field: %q", field)
		} else {
			t.Logf("✅ Field %q present", field)
		}
	}

	// Validate field types
	typeValidations := map[string]string{
		"timestamp":        "string",
		"message":          "string",
		"level":            "string",
		"status":           "string",
		"service":          "string",
		"env":              "string",
		"version":          "string",
		"http.status_code": "float64", // JSON numbers are float64
	}

	for field, expectedType := range typeValidations {
		if val, ok := result[field]; ok {
			actualType := fmt.Sprintf("%T", val)
			if actualType != expectedType {
				t.Errorf("❌ Field %q has type %s, expected %s", field, actualType, expectedType)
			}
		}
	}

	// Validate timestamp format (ISO 8601)
	if ts, ok := result["timestamp"].(string); ok {
		_, parseErr := time.Parse(time.RFC3339Nano, ts)
		if parseErr != nil {
			t.Errorf("❌ Timestamp not in ISO 8601 format: %v", parseErr)
		} else {
			t.Log("✅ Timestamp in valid ISO 8601 format")
		}
	}

	// Validate level is "error"
	if level, ok := result["level"].(string); ok {
		if level != "error" {
			t.Errorf("❌ Level = %q, expected 'error'", level)
		} else {
			t.Log("✅ Level is 'error'")
		}
	}

	// Validate status is "error"
	if status, ok := result["status"].(string); ok {
		if status != "error" {
			t.Errorf("❌ Status = %q, expected 'error'", status)
		} else {
			t.Log("✅ Status is 'error'")
		}
	}

	// Validate context is a map
	if context, ok := result["context"].(map[string]interface{}); ok {
		if len(context) == 0 {
			t.Error("❌ Context is empty")
		} else {
			t.Logf("✅ Context has %d fields", len(context))
		}
	}

	t.Log("✅ Log entry structure validated")
	t.Logf("JSON output:\n%s", string(jsonBytes))
}

// TestDatadogIntegration_LogTraceCorrelation validates that logs can be
// correlated with traces using dd.trace_id and dd.span_id
func TestDatadogIntegration_LogTraceCorrelation(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test.operation")

	var ErrTest errific.Err = "operation failed"
	err := ErrTest.New().WithCode("TEST_001")

	// Create log entry
	logEntry := ToLogEntry(err)

	// Before enrichment, trace fields should be empty
	if logEntry.TraceID != "" {
		t.Errorf("TraceID should be empty before enrichment, got %q", logEntry.TraceID)
	}

	// Enrich with trace info
	EnrichLogEntry(logEntry, span)
	span.Finish()

	// After enrichment, trace fields should be populated
	if logEntry.TraceID == "" {
		t.Error("❌ TraceID not set after enrichment")
	} else {
		t.Logf("✅ TraceID set: %s", logEntry.TraceID)
	}

	if logEntry.SpanID == "" {
		t.Error("❌ SpanID not set after enrichment")
	} else {
		t.Logf("✅ SpanID set: %s", logEntry.SpanID)
	}

	// Validate JSON structure
	jsonBytes, _ := json.Marshal(logEntry)
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	if _, ok := result["dd.trace_id"]; !ok {
		t.Error("❌ dd.trace_id not in JSON output")
	} else {
		t.Log("✅ dd.trace_id present in JSON")
	}

	if _, ok := result["dd.span_id"]; !ok {
		t.Error("❌ dd.span_id not in JSON output")
	} else {
		t.Log("✅ dd.span_id present in JSON")
	}

	t.Log("✅ Log-to-trace correlation validated")
}

// TestDatadogIntegration_UnifiedServiceTagging validates unified service
// tagging with service, env, version fields
func TestDatadogIntegration_UnifiedServiceTagging(t *testing.T) {
	var ErrTest errific.Err = "test error"
	err := ErrTest.New()

	logEntry := ToLogEntry(err)

	// Before SetServiceInfo
	if logEntry.Service != "" || logEntry.Env != "" || logEntry.Version != "" {
		t.Error("❌ Service info should be empty before SetServiceInfo")
	}

	// Set service info
	SetServiceInfo(logEntry, "payment-service", "production", "2.1.3")

	// Validate fields
	if logEntry.Service != "payment-service" {
		t.Errorf("❌ Service = %q, want 'payment-service'", logEntry.Service)
	} else {
		t.Log("✅ Service set correctly")
	}

	if logEntry.Env != "production" {
		t.Errorf("❌ Env = %q, want 'production'", logEntry.Env)
	} else {
		t.Log("✅ Env set correctly")
	}

	if logEntry.Version != "2.1.3" {
		t.Errorf("❌ Version = %q, want '2.1.3'", logEntry.Version)
	} else {
		t.Log("✅ Version set correctly")
	}

	// Validate JSON output
	jsonBytes, _ := json.Marshal(logEntry)
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	requiredFields := []string{"service", "env", "version"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("❌ Field %q missing in JSON", field)
		}
	}

	t.Log("✅ Unified service tagging validated")
}

// TestDatadogIntegration_ErrorTrackingCompatibility validates that errors
// are formatted correctly for Datadog Error Tracking
func TestDatadogIntegration_ErrorTrackingCompatibility(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	// Create error that should be grouped by error.code
	var ErrPayment errific.Err = "payment declined"
	err := ErrPayment.New().
		WithCode("PAYMENT_DECLINED").
		WithCategory(errific.CategoryClient).
		WithUserID("user-12345").
		WithHTTPStatus(402).
		WithContext(errific.Context{
			"amount":       99.99,
			"currency":     "USD",
			"decline_code": "insufficient_funds",
		})

	// Record to span
	span := tracer.StartSpan("payment.process")
	RecordError(span, err)

	// Validate span has error.code for Error Tracking grouping
	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	finishedSpan := spans[0]

	// Error Tracking requires these fields
	errorTrackingFields := map[string]interface{}{
		"error.code":     "PAYMENT_DECLINED", // For grouping
		"error.category": "client",           // For classification
		"error.msg":      nil,                // Must exist (actual value varies)
		"user.id":        "user-12345",       // For impact tracking
	}

	for field, expected := range errorTrackingFields {
		actual := finishedSpan.Tag(field)
		if actual == nil {
			t.Errorf("❌ Error Tracking field %q missing", field)
			continue
		}

		if expected != nil && actual != expected {
			// Special case for error.msg which includes caller
			if field == "error.msg" && contains(actual.(string), "payment declined") {
				t.Logf("✅ Field %q present with message", field)
				continue
			}
			t.Errorf("❌ Field %q = %v, want %v", field, actual, expected)
		} else {
			t.Logf("✅ Error Tracking field %q present", field)
		}
	}

	// Create log entry
	logEntry := ToLogEntry(err)
	EnrichLogEntry(logEntry, span)
	SetServiceInfo(logEntry, "payment-service", "production", "1.0.0")

	// Validate log entry has error tracking fields
	jsonBytes, _ := json.Marshal(logEntry)
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	logErrorTrackingFields := []string{
		"error.code",     // For grouping
		"error.kind",     // Alternative to error.code
		"error.message",  // Error message
		"error.category", // Error category
		"dd.trace_id",    // Link to trace
		"user.id",        // User impact
	}

	for _, field := range logErrorTrackingFields {
		if _, ok := result[field]; !ok {
			t.Errorf("❌ Log Error Tracking field %q missing", field)
		} else {
			t.Logf("✅ Log Error Tracking field %q present", field)
		}
	}

	t.Log("✅ Error Tracking compatibility validated")
}

// TestDatadogIntegration_RetryableErrorMetadata validates retry-specific
// metadata is properly recorded
func TestDatadogIntegration_RetryableErrorMetadata(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	var ErrTimeout errific.Err = "operation timeout"
	err := ErrTimeout.New().
		WithCode("TIMEOUT_001").
		WithCategory(errific.CategoryTimeout).
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3)

	span := tracer.StartSpan("test.operation")
	RecordError(span, err)

	spans := mt.FinishedSpans()
	finishedSpan := spans[0]

	// Validate retry metadata in span
	if val := finishedSpan.Tag("error.retryable"); val == nil {
		t.Error("❌ error.retryable not set")
	} else if fmt.Sprint(val) != fmt.Sprint(true) {
		t.Errorf("❌ error.retryable = %v, want true", val)
	} else {
		t.Log("✅ error.retryable = true")
	}

	if val := finishedSpan.Tag("error.retry_after"); val == nil {
		t.Error("❌ error.retry_after not set")
	} else if fmt.Sprint(val) != fmt.Sprint("5s") {
		t.Errorf("❌ error.retry_after = %v, want '5s'", val)
	} else {
		t.Log("✅ error.retry_after = 5s")
	}

	if val := finishedSpan.Tag("error.max_retries"); val == nil {
		t.Error("❌ error.max_retries not set")
	} else if fmt.Sprint(val) != fmt.Sprint(3) {
		t.Errorf("❌ error.max_retries = %v, want 3", val)
	} else {
		t.Log("✅ error.max_retries = 3")
	}

	// Validate retry metadata in log
	logEntry := ToLogEntry(err)
	jsonBytes, _ := json.Marshal(logEntry)
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	if val, ok := result["error.retryable"].(bool); !ok || !val {
		t.Errorf("❌ Log error.retryable = %v, want true", val)
	} else {
		t.Log("✅ Log error.retryable = true")
	}

	if val, ok := result["error.retry_after"].(string); !ok || val != "5s" {
		t.Errorf("❌ Log error.retry_after = %v, want '5s'", val)
	} else {
		t.Log("✅ Log error.retry_after = 5s")
	}

	t.Log("✅ Retryable error metadata validated")
}

// TestDatadogIntegration_CompleteWorkflow validates a complete real-world
// workflow with both span and log recording
func TestDatadogIntegration_CompleteWorkflow(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	// Simulate real-world error scenario
	var ErrDatabase errific.Err = "database query failed"
	err := ErrDatabase.New().
		WithCode("DB_QUERY_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("req-abc-123").
		WithRequestID("req-xyz-789").
		WithUserID("user-456").
		WithSessionID("sess-def-012").
		WithRetryable(true).
		WithRetryAfter(10 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(503).
		WithLabel("service", "user-service").
		WithLabel("database", "postgres").
		WithContext(errific.Context{
			"query":       "SELECT * FROM users WHERE id = $1",
			"duration_ms": 5000,
			"timeout_ms":  3000,
		})

	// 1. Record to span
	span := tracer.StartSpan("database.query")
	RecordError(span, err)

	// 2. Create log entry
	logEntry := ToLogEntry(err)

	// 3. Enrich with trace info
	EnrichLogEntry(logEntry, span)

	// 4. Set service info
	SetServiceInfo(logEntry, "user-service", "production", "2.3.1")

	// 5. Add custom context
	AddContext(logEntry, map[string]interface{}{
		"host":     "db-primary-1",
		"pool_id":  5,
		"trace_on": true,
	})

	// Validate span
	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	// Validate log entry
	jsonBytes, _ := json.MarshalIndent(logEntry, "", "  ")
	var result map[string]interface{}
	json.Unmarshal(jsonBytes, &result)

	// Check all critical fields are present
	criticalFields := []string{
		// Datadog reserved
		"timestamp", "service", "env", "version",
		"dd.trace_id", "dd.span_id",
		"message", "level", "status",
		// Error specific
		"error.code", "error.category", "error.message",
		// Correlation
		"correlation.id", "request.id", "user.id", "session.id",
		// HTTP
		"http.status_code",
		// Retry
		"error.retryable", "error.retry_after", "error.max_retries",
		// Custom
		"labels", "context",
	}

	missingFields := []string{}
	for _, field := range criticalFields {
		if _, ok := result[field]; !ok {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		t.Errorf("❌ Missing %d critical fields: %v", len(missingFields), missingFields)
	} else {
		t.Logf("✅ All %d critical fields present", len(criticalFields))
	}

	t.Log("✅ Complete workflow validated")
	t.Logf("Complete JSON output:\n%s", string(jsonBytes))
}
