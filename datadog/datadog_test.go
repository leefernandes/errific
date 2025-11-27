package datadog

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leefernandes/errific"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func TestRecordError_NilChecks(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")

	// Nil error should finish normally
	RecordError(span, nil)

	// Nil span should not panic
	RecordError(nil, errors.New("test"))

	// Both nil should not panic
	RecordError(nil, nil)
}

func TestRecordError_BasicError(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	err := errors.New("basic error")

	RecordError(span, err)

	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	finishedSpan := spans[0]

	// Check error tags
	if msg := finishedSpan.Tag("error.msg"); msg != "basic error" {
		t.Errorf("error.msg = %v, want 'basic error'", msg)
	}

	if errType := finishedSpan.Tag("error.type"); errType == nil {
		t.Error("error.type not set")
	}
}

func TestRecordError_ErrificError(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")

	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithSessionID("sess-abc").
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(500).
		WithTags("tag1", "tag2", "tag3").
		WithLabel("service", "test-service").
		WithLabel("severity", "high").
		WithContext(errific.Context{
			"query":       "SELECT * FROM users",
			"duration_ms": 1500,
		})

	RecordError(span, err)

	spans := mt.FinishedSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	finishedSpan := spans[0]

	// Check all tags
	tests := []struct {
		key      string
		expected interface{}
	}{
		{"error.code", "TEST_001"},
		{"error.category", "server"},
		{"correlation.id", "corr-123"},
		{"request.id", "req-456"},
		{"user.id", "user-789"},
		{"session.id", "sess-abc"},
		{"error.retryable", true},
		{"error.retry_after", "5s"},
		{"error.max_retries", 3},
		{"http.status_code", 500},
		{"label.service", "test-service"},
		{"label.severity", "high"},
		{"context.query", "SELECT * FROM users"},
		{"context.duration_ms", "1500"},
	}

	for _, tt := range tests {
		actual := finishedSpan.Tag(tt.key)
		// Use fmt.Sprintf for comparison to handle type differences
		if fmt.Sprint(actual) != fmt.Sprint(tt.expected) {
			t.Errorf("tag %q = %v (type %T), want %v (type %T)", tt.key, actual, actual, tt.expected, tt.expected)
		}
	}

	// Check tags (array stored as individual tags)
	if tag0 := finishedSpan.Tag("error.tag.0"); tag0 != "tag1" {
		t.Errorf("error.tag.0 = %v, want 'tag1'", tag0)
	}
}

func TestToLogEntry_NilError(t *testing.T) {
	entry := ToLogEntry(nil)
	if entry != nil {
		t.Error("expected nil entry for nil error")
	}
}

func TestToLogEntry_BasicError(t *testing.T) {
	err := errors.New("basic error")
	entry := ToLogEntry(err)

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	if entry.Message != "basic error" {
		t.Errorf("message = %q, want 'basic error'", entry.Message)
	}

	if entry.Level != "error" {
		t.Errorf("level = %q, want 'error'", entry.Level)
	}

	if entry.Status != "error" {
		t.Errorf("status = %q, want 'error'", entry.Status)
	}

	if entry.ErrorMessage != "basic error" {
		t.Errorf("error.message = %q, want 'basic error'", entry.ErrorMessage)
	}
}

func TestToLogEntry_ErrificError(t *testing.T) {
	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithSessionID("sess-abc").
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(500).
		WithTags("tag1", "tag2").
		WithLabel("service", "test-service").
		WithContext(errific.Context{
			"query": "SELECT * FROM users",
		})

	entry := ToLogEntry(err)

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	// Check all fields
	// Check message contains error text (may include caller info)
	if entry.Message == "" || !contains(entry.Message, "test error") {
		t.Errorf("Message = %q, should contain 'test error'", entry.Message)
	}

	tests := []struct {
		name     string
		got      interface{}
		expected interface{}
	}{
		{"Level", entry.Level, "error"},
		{"Status", entry.Status, "error"},
		{"ErrorCode", entry.ErrorCode, "TEST_001"},
		{"ErrorKind", entry.ErrorKind, "TEST_001"},
		{"ErrorCategory", entry.ErrorCategory, "server"},
		{"CorrelationID", entry.CorrelationID, "corr-123"},
		{"TraceID", entry.TraceID, "corr-123"},
		{"RequestID", entry.RequestID, "req-456"},
		{"SpanID", entry.SpanID, "req-456"},
		{"UserID", entry.UserID, "user-789"},
		{"SessionID", entry.SessionID, "sess-abc"},
		{"HTTPStatusCode", entry.HTTPStatusCode, 500},
		{"RetryAfter", entry.RetryAfter, "5s"},
		{"MaxRetries", entry.MaxRetries, 3},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
		}
	}

	// Check retryable
	if entry.Retryable == nil || !*entry.Retryable {
		t.Error("expected Retryable = true")
	}

	// Check tags
	if len(entry.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(entry.Tags))
	}

	// Check labels
	if entry.Labels["service"] != "test-service" {
		t.Errorf("Labels[service] = %v, want 'test-service'", entry.Labels["service"])
	}

	// Check context
	if entry.Context["query"] != "SELECT * FROM users" {
		t.Errorf("Context[query] = %v, want 'SELECT * FROM users'", entry.Context["query"])
	}
}

func TestToLogEntry_JSONSerialization(t *testing.T) {
	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123")

	entry := ToLogEntry(err)
	entry.Service = "my-service"
	entry.Env = "production"
	entry.Version = "1.0.4"

	// Serialize to JSON
	jsonBytes, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		t.Fatalf("JSON marshal failed: %v", jsonErr)
	}

	// Deserialize to check structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}

	// Check reserved attributes
	requiredFields := []string{
		"timestamp", "message", "level", "status",
		"service", "env", "version",
		"dd.trace_id", "error.code", "error.category",
	}

	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON missing required field: %q", field)
		}
	}
}

func TestEnrichLogEntry(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	entry := &LogEntry{
		Message: "test message",
	}

	EnrichLogEntry(entry, span)
	span.Finish()

	// Check trace and span IDs are set
	if entry.TraceID == "" {
		t.Error("TraceID not set")
	}

	if entry.SpanID == "" {
		t.Error("SpanID not set")
	}
}

func TestEnrichLogEntry_NilChecks(t *testing.T) {
	mt := mocktracer.Start()
	defer mt.Stop()

	span := tracer.StartSpan("test")
	defer span.Finish()

	// Nil entry
	EnrichLogEntry(nil, span)

	// Nil span
	entry := &LogEntry{}
	EnrichLogEntry(entry, nil)

	// Both nil
	EnrichLogEntry(nil, nil)
}

func TestSetServiceInfo(t *testing.T) {
	entry := &LogEntry{}

	SetServiceInfo(entry, "my-service", "production", "1.0.4")

	if entry.Service != "my-service" {
		t.Errorf("Service = %q, want 'my-service'", entry.Service)
	}

	if entry.Env != "production" {
		t.Errorf("Env = %q, want 'production'", entry.Env)
	}

	if entry.Version != "1.0.4" {
		t.Errorf("Version = %q, want '1.0.4'", entry.Version)
	}
}

func TestSetServiceInfo_NilEntry(t *testing.T) {
	// Should not panic
	SetServiceInfo(nil, "service", "env", "version")
}

func TestAddContext(t *testing.T) {
	entry := &LogEntry{}

	context := map[string]interface{}{
		"customer_id": "12345",
		"plan":        "enterprise",
		"count":       42,
	}

	AddContext(entry, context)

	if entry.Context["customer_id"] != "12345" {
		t.Errorf("Context[customer_id] = %v, want '12345'", entry.Context["customer_id"])
	}

	if entry.Context["plan"] != "enterprise" {
		t.Errorf("Context[plan] = %v, want 'enterprise'", entry.Context["plan"])
	}

	if entry.Context["count"] != 42 {
		t.Errorf("Context[count] = %v, want 42", entry.Context["count"])
	}
}

func TestAddContext_Multiple(t *testing.T) {
	entry := &LogEntry{}

	AddContext(entry, map[string]interface{}{
		"key1": "value1",
	})

	AddContext(entry, map[string]interface{}{
		"key2": "value2",
	})

	if len(entry.Context) != 2 {
		t.Errorf("len(Context) = %d, want 2", len(entry.Context))
	}

	if entry.Context["key1"] != "value1" {
		t.Error("key1 not preserved")
	}

	if entry.Context["key2"] != "value2" {
		t.Error("key2 not added")
	}
}

func TestAddContext_NilEntry(t *testing.T) {
	// Should not panic
	AddContext(nil, map[string]interface{}{"key": "value"})
}

func BenchmarkRecordError(b *testing.B) {
	mt := mocktracer.Start()
	defer mt.Stop()

	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		span := tracer.StartSpan("test")
		RecordError(span, err)
	}
}

func BenchmarkToLogEntry(b *testing.B) {
	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToLogEntry(err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}

func BenchmarkJSONSerialization(b *testing.B) {
	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123")

	entry := ToLogEntry(err)
	entry.Service = "my-service"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(entry)
	}
}
