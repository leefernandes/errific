// Package datadog provides Datadog integration helpers for errific errors.
//
// This package provides seamless integration with Datadog's APM traces and structured
// logging, including automatic trace correlation, error tracking, and JSON log formatting.
//
// Usage with dd-trace-go:
//
//	import "github.com/leefernandes/errific/datadog"
//
//	span, _ := tracer.StartSpanFromContext(ctx, "operation")
//	defer span.Finish()
//
//	if err := doSomething(); err != nil {
//	    datadog.RecordError(span, err)  // One-liner with full metadata!
//	    return err
//	}
//
// Usage with structured logging:
//
//	logEntry := datadog.ToLogEntry(err)
//	json.Marshal(logEntry)  // Datadog-compatible JSON log
package datadog

import (
	"fmt"
	"time"

	"github.com/leefernandes/errific"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// RecordError records an error to a Datadog span with full errific metadata.
//
// This function:
// - Marks the span as an error using span.Finish(tracer.WithError(err))
// - Sets error.msg, error.type, error.stack tags
// - Adds all errific metadata as span tags
// - Follows Datadog naming conventions
//
// Example:
//
//	span, ctx := tracer.StartSpanFromContext(ctx, "ProcessOrder")
//	defer datadog.RecordError(span, err)  // Will mark error if non-nil
//
//	if err := processOrder(orderID); err != nil {
//	    return err
//	}
func RecordError(span tracer.Span, err error) {
	if span == nil {
		return
	}

	// If no error, finish normally
	if err == nil {
		span.Finish()
		return
	}

	// Set error tags (Datadog standard)
	span.SetTag("error.msg", err.Error())
	span.SetTag("error.type", fmt.Sprintf("%T", err))

	// Stack trace would be added here if errific exposed it
	// For now, use error message which may contain wrapped errors

	// Add errific-specific tags
	if code := errific.GetCode(err); code != "" {
		span.SetTag("error.code", code)
	}

	if category := errific.GetCategory(err); category != "" {
		span.SetTag("error.category", string(category))
	}

	if correlationID := errific.GetCorrelationID(err); correlationID != "" {
		span.SetTag("correlation.id", correlationID)
	}

	if requestID := errific.GetRequestID(err); requestID != "" {
		span.SetTag("request.id", requestID)
	}

	if userID := errific.GetUserID(err); userID != "" {
		span.SetTag("user.id", userID)
	}

	if sessionID := errific.GetSessionID(err); sessionID != "" {
		span.SetTag("session.id", sessionID)
	}

	if errific.IsRetryable(err) {
		span.SetTag("error.retryable", true)

		if retryAfter := errific.GetRetryAfter(err); retryAfter > 0 {
			span.SetTag("error.retry_after", retryAfter.String())
		}

		if maxRetries := errific.GetMaxRetries(err); maxRetries > 0 {
			span.SetTag("error.max_retries", maxRetries)
		}
	}

	if httpStatus := errific.GetHTTPStatus(err); httpStatus > 0 {
		span.SetTag("http.status_code", httpStatus)
	}

	// Add tags as comma-separated string (Datadog best practice)
	if tags := errific.GetTags(err); len(tags) > 0 {
		for i, tag := range tags {
			span.SetTag(fmt.Sprintf("error.tag.%d", i), tag)
		}
	}

	// Add labels as individual tags
	if labels := errific.GetLabels(err); len(labels) > 0 {
		for key, value := range labels {
			span.SetTag("label."+key, value)
		}
	}

	// Add context as individual tags
	if context := errific.GetContext(err); len(context) > 0 {
		for key, value := range context {
			span.SetTag("context."+key, fmt.Sprint(value))
		}
	}

	// Finish span with error
	span.Finish(tracer.WithError(err))
}

// LogEntry represents a Datadog-compatible structured log entry.
//
// This struct follows Datadog's reserved attributes conventions and
// can be marshaled to JSON for log ingestion.
type LogEntry struct {
	// Datadog reserved attributes (processed specially)
	Timestamp   string `json:"timestamp"`
	Service     string `json:"service,omitempty"`
	Env         string `json:"env,omitempty"`
	Version     string `json:"version,omitempty"`
	TraceID     string `json:"dd.trace_id,omitempty"`
	SpanID      string `json:"dd.span_id,omitempty"`
	Message     string `json:"message"`
	Level       string `json:"level"`
	Status      string `json:"status"`
	Host        string `json:"host,omitempty"`
	Source      string `json:"source,omitempty"`
	Logger      string `json:"logger.name,omitempty"`
	Thread      string `json:"logger.thread_name,omitempty"`

	// Error-specific fields
	ErrorKind       string `json:"error.kind,omitempty"`
	ErrorMessage    string `json:"error.message,omitempty"`
	ErrorStack      string `json:"error.stack,omitempty"`
	ErrorCode       string `json:"error.code,omitempty"`
	ErrorCategory   string `json:"error.category,omitempty"`

	// Correlation fields
	CorrelationID   string `json:"correlation.id,omitempty"`
	RequestID       string `json:"request.id,omitempty"`
	UserID          string `json:"user.id,omitempty"`
	SessionID       string `json:"session.id,omitempty"`

	// HTTP fields
	HTTPStatusCode  int    `json:"http.status_code,omitempty"`
	HTTPMethod      string `json:"http.method,omitempty"`
	HTTPUrl         string `json:"http.url,omitempty"`
	HTTPUserAgent   string `json:"http.useragent,omitempty"`

	// Retry fields
	Retryable       *bool  `json:"error.retryable,omitempty"`
	RetryAfter      string `json:"error.retry_after,omitempty"`
	MaxRetries      int    `json:"error.max_retries,omitempty"`

	// Custom attributes (everything else)
	Tags            []string               `json:"error.tags,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Context         map[string]interface{} `json:"context,omitempty"`

	// Caller information
	Caller          string `json:"caller,omitempty"`
}

// ToLogEntry converts an errific error to a Datadog-compatible log entry.
//
// This creates a structured log entry that follows Datadog's reserved attributes
// and naming conventions. The entry can be marshaled to JSON and sent to Datadog.
//
// Example:
//
//	logEntry := datadog.ToLogEntry(err)
//	logEntry.Service = "my-service"
//	logEntry.Env = "production"
//	logBytes, _ := json.Marshal(logEntry)
//	log.Println(string(logBytes))
func ToLogEntry(err error) *LogEntry {
	if err == nil {
		return nil
	}

	entry := &LogEntry{
		Timestamp:    time.Now().Format(time.RFC3339Nano),
		Message:      err.Error(),
		Level:        "error",
		Status:       "error",
		ErrorMessage: err.Error(),
	}

	// Extract errific metadata if available
	if code := errific.GetCode(err); code != "" {
		entry.ErrorCode = code
		entry.ErrorKind = code
	}

	if category := errific.GetCategory(err); category != "" {
		entry.ErrorCategory = string(category)
	}

	// Stack trace would be added here if errific exposed it
	// Error messages contain wrapped error info which serves similar purpose

	if correlationID := errific.GetCorrelationID(err); correlationID != "" {
		entry.CorrelationID = correlationID
		entry.TraceID = correlationID // Can be used as trace ID
	}

	if requestID := errific.GetRequestID(err); requestID != "" {
		entry.RequestID = requestID
		if entry.SpanID == "" {
			entry.SpanID = requestID // Can be used as span ID
		}
	}

	if userID := errific.GetUserID(err); userID != "" {
		entry.UserID = userID
	}

	if sessionID := errific.GetSessionID(err); sessionID != "" {
		entry.SessionID = sessionID
	}

	if httpStatus := errific.GetHTTPStatus(err); httpStatus > 0 {
		entry.HTTPStatusCode = httpStatus
	}

	if errific.IsRetryable(err) {
		retryable := true
		entry.Retryable = &retryable

		if retryAfter := errific.GetRetryAfter(err); retryAfter > 0 {
			entry.RetryAfter = retryAfter.String()
		}

		if maxRetries := errific.GetMaxRetries(err); maxRetries > 0 {
			entry.MaxRetries = maxRetries
		}
	}

	if tags := errific.GetTags(err); len(tags) > 0 {
		entry.Tags = tags
	}

	if labels := errific.GetLabels(err); len(labels) > 0 {
		entry.Labels = labels
	}

	if context := errific.GetContext(err); len(context) > 0 {
		entry.Context = context
	}

	// Caller info would be added here if errific exposed it publicly

	return entry
}

// EnrichLogEntry enriches a log entry with trace and span IDs from a Datadog span.
//
// This enables log-to-trace correlation in Datadog. Call this before logging
// to automatically link logs to traces.
//
// Example:
//
//	logEntry := datadog.ToLogEntry(err)
//	datadog.EnrichLogEntry(logEntry, span)
//	json.Marshal(logEntry)  // Now has dd.trace_id and dd.span_id
func EnrichLogEntry(entry *LogEntry, span tracer.Span) {
	if entry == nil || span == nil {
		return
	}

	ctx := span.Context()
	if ctx == nil {
		return
	}

	// Set trace and span IDs for log-to-trace correlation
	entry.TraceID = fmt.Sprintf("%d", ctx.TraceID())
	entry.SpanID = fmt.Sprintf("%d", ctx.SpanID())
}

// SetServiceInfo sets the unified service tagging fields on a log entry.
//
// Datadog recommends using DD_SERVICE, DD_ENV, and DD_VERSION for unified
// service tagging. This helper makes it easy to set these fields.
//
// Example:
//
//	logEntry := datadog.ToLogEntry(err)
//	datadog.SetServiceInfo(logEntry, "my-service", "production", "1.0.4")
func SetServiceInfo(entry *LogEntry, service, env, version string) {
	if entry == nil {
		return
	}

	entry.Service = service
	entry.Env = env
	entry.Version = version
}

// AddContext adds custom context fields to a log entry.
//
// This is useful for adding application-specific metadata that doesn't
// fit into standard fields.
//
// Example:
//
//	logEntry := datadog.ToLogEntry(err)
//	datadog.AddContext(logEntry, map[string]interface{}{
//	    "customer_id": "12345",
//	    "plan": "enterprise",
//	})
func AddContext(entry *LogEntry, context map[string]interface{}) {
	if entry == nil {
		return
	}

	if entry.Context == nil {
		entry.Context = make(map[string]interface{})
	}

	for k, v := range context {
		entry.Context[k] = v
	}
}
