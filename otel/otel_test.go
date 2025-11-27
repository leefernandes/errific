package otel

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/leefernandes/errific"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// MockSpan implements trace.Span for testing
type MockSpan struct {
	noop.Span
	attributes   map[string]interface{}
	status       codes.Code
	statusDesc   string
	events       []string
	exceptionSet bool
}

func NewMockSpan() *MockSpan {
	return &MockSpan{
		attributes: make(map[string]interface{}),
	}
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	for _, attr := range attrs {
		m.attributes[string(attr.Key)] = attr.Value.AsInterface()
	}
}

func (m *MockSpan) SetStatus(code codes.Code, desc string) {
	m.status = code
	m.statusDesc = desc
}

func (m *MockSpan) AddEvent(name string, opts ...oteltrace.EventOption) {
	m.events = append(m.events, name)
}

func (m *MockSpan) RecordException(err error, opts ...oteltrace.EventOption) {
	m.exceptionSet = true
	m.events = append(m.events, "exception")
}

func TestRecordError_NilChecks(t *testing.T) {
	span := NewMockSpan()
	err := errific.Err("test error").New()

	// Nil error
	RecordError(span, nil)
	if span.status != 0 {
		t.Error("expected no status change for nil error")
	}

	// Nil span
	RecordError(nil, err)
	// Should not panic

	// Both nil
	RecordError(nil, nil)
	// Should not panic
}

func TestRecordError_BasicError(t *testing.T) {
	span := NewMockSpan()
	err := errors.New("basic error")

	RecordError(span, err)

	if span.status != codes.Error {
		t.Errorf("expected Error status, got %v", span.status)
	}

	if span.statusDesc != "basic error" {
		t.Errorf("expected status desc 'basic error', got %v", span.statusDesc)
	}

	// Check exception event was added
	found := false
	for _, event := range span.events {
		if event == "exception" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected exception event to be added")
	}
}

func TestRecordError_ErrificError(t *testing.T) {
	span := NewMockSpan()

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
		WithMCPCode(-32000).
		WithTags("tag1", "tag2", "tag3").
		WithLabel("service", "test-service").
		WithLabel("severity", "high").
		WithContext(errific.Context{
			"query":       "SELECT * FROM users",
			"duration_ms": 1500,
		})

	RecordError(span, err)

	// Check status
	if span.status != codes.Error {
		t.Errorf("expected Error status, got %v", span.status)
	}

	// Check exception event was added
	found := false
	for _, event := range span.events {
		if event == "exception" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected exception event to be added")
	}

	// Check all attributes
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
		{"error.max_retries", int64(3)},
		{"http.status_code", int64(500)},
		{"mcp.error_code", int64(-32000)},
		{"label.service", "test-service"},
		{"label.severity", "high"},
		{"context.query", "SELECT * FROM users"},
		{"context.duration_ms", "1500"},
	}

	for _, tt := range tests {
		if val, ok := span.attributes[tt.key]; !ok {
			t.Errorf("attribute %q not found", tt.key)
		} else if fmt.Sprint(val) != fmt.Sprint(tt.expected) {
			t.Errorf("attribute %q = %v, expected %v", tt.key, val, tt.expected)
		}
	}

	// Check tags (array)
	if tags, ok := span.attributes["error.tags"]; !ok {
		t.Error("tags attribute not found")
	} else {
		tagSlice := tags.([]string)
		if len(tagSlice) != 3 {
			t.Errorf("expected 3 tags, got %d", len(tagSlice))
		}
	}
}

func TestRecordError_MinimalErrificError(t *testing.T) {
	span := NewMockSpan()

	var ErrMinimal errific.Err = "minimal error"
	err := ErrMinimal.New()

	RecordError(span, err)

	// Should still work with minimal error
	if span.status != codes.Error {
		t.Errorf("expected Error status, got %v", span.status)
	}

	// Should have few or no custom attributes
	if len(span.attributes) > 2 {
		t.Logf("attributes: %+v", span.attributes)
		// Some attributes are okay, but shouldn't have many
	}
}

func TestRecordErrorWithEvent(t *testing.T) {
	span := NewMockSpan()

	var ErrDB errific.Err = "database error"
	err := ErrDB.New().WithCode("DB_001")

	eventAttrs := map[string]string{
		"pool_size":    "10",
		"connections":  "10",
		"wait_time_ms": "5000",
	}

	RecordErrorWithEvent(span, err, "db_connection_failed", eventAttrs)

	// Check error recorded
	if span.status != codes.Error {
		t.Error("expected error status")
	}

	// Check event added
	found := false
	for _, event := range span.events {
		if event == "db_connection_failed" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected custom event to be added")
	}
}

func TestAddErrorContext(t *testing.T) {
	span := NewMockSpan()

	var ErrAttempt errific.Err = "attempt failed"
	err := ErrAttempt.New().
		WithCode("ATTEMPT_001").
		WithCategory(errific.CategoryNetwork).
		WithCorrelationID("corr-123")

	AddErrorContext(span, err)

	// Status should NOT be set to error
	if span.status == codes.Error {
		t.Error("status should not be Error for AddErrorContext")
	}

	// Should have attempted attributes
	if _, ok := span.attributes["error.attempted.code"]; !ok {
		t.Error("expected error.attempted.code attribute")
	}

	if _, ok := span.attributes["error.attempted.category"]; !ok {
		t.Error("expected error.attempted.category attribute")
	}

	if _, ok := span.attributes["error.attempted.message"]; !ok {
		t.Error("expected error.attempted.message attribute")
	}

	// Correlation ID should still be present
	if _, ok := span.attributes["correlation.id"]; !ok {
		t.Error("expected correlation.id attribute")
	}
}

func TestAddErrorContext_NilChecks(t *testing.T) {
	span := NewMockSpan()
	err := errific.Err("test").New()

	// Nil error
	AddErrorContext(span, nil)
	if len(span.attributes) > 0 {
		t.Error("expected no attributes for nil error")
	}

	// Nil span
	AddErrorContext(nil, err)
	// Should not panic

	// Both nil
	AddErrorContext(nil, nil)
	// Should not panic
}

func BenchmarkRecordError(b *testing.B) {
	span := NewMockSpan()

	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer).
		WithCorrelationID("corr-123").
		WithContext(errific.Context{
			"key1": "value1",
			"key2": "value2",
		})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordError(span, err)
		span.attributes = make(map[string]interface{}) // Reset for next iteration
	}
}

func BenchmarkRecordError_MinimalError(b *testing.B) {
	span := NewMockSpan()
	err := errors.New("basic error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RecordError(span, err)
		span.attributes = make(map[string]interface{})
	}
}

func BenchmarkAddErrorContext(b *testing.B) {
	span := NewMockSpan()

	var ErrTest errific.Err = "test error"
	err := ErrTest.New().
		WithCode("TEST_001").
		WithCategory(errific.CategoryServer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddErrorContext(span, err)
		span.attributes = make(map[string]interface{})
	}
}
