// Package otel provides OpenTelemetry integration helpers for errific errors.
//
// This package is completely optional and has no effect on the core errific package.
// It provides convenience functions for recording errific errors to OpenTelemetry spans.
//
// Usage:
//
//	import "github.com/leefernandes/errific/otel"
//
//	span := tracer.Start(ctx, "operation")
//	defer span.End()
//
//	if err := doSomething(); err != nil {
//	    otel.RecordError(span, err)  // One-liner!
//	    return err
//	}
package otel

import (
	"fmt"

	"github.com/leefernandes/errific"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// RecordError records an error to an OpenTelemetry span with full errific metadata.
//
// This function:
// - Sets the span status to Error
// - Records the exception event
// - Adds errific-specific attributes (code, category, correlation_id, etc.)
// - Adds structured context as span attributes
//
// If the error is not an errific error, it still works and records basic error information.
//
// Example:
//
//	span := tracer.Start(ctx, "ProcessOrder")
//	defer span.End()
//
//	if err := processOrder(orderID); err != nil {
//	    otel.RecordError(span, err)  // Automatically extracts all metadata
//	    return err
//	}
func RecordError(span trace.Span, err error) {
	if err == nil || span == nil {
		return
	}

	// Set span status to error
	span.SetStatus(codes.Error, err.Error())

	// Record exception event (OpenTelemetry standard)
	span.AddEvent("exception", trace.WithAttributes(
		attribute.String("exception.type", fmt.Sprintf("%T", err)),
		attribute.String("exception.message", err.Error()),
	))

	// Add errific-specific attributes if available
	attrs := make([]attribute.KeyValue, 0, 16)

	if code := errific.GetCode(err); code != "" {
		attrs = append(attrs, attribute.String("error.code", code))
	}

	if category := errific.GetCategory(err); category != "" {
		attrs = append(attrs, attribute.String("error.category", string(category)))
	}

	if correlationID := errific.GetCorrelationID(err); correlationID != "" {
		attrs = append(attrs, attribute.String("correlation.id", correlationID))
	}

	if requestID := errific.GetRequestID(err); requestID != "" {
		attrs = append(attrs, attribute.String("request.id", requestID))
	}

	if userID := errific.GetUserID(err); userID != "" {
		attrs = append(attrs, attribute.String("user.id", userID))
	}

	if sessionID := errific.GetSessionID(err); sessionID != "" {
		attrs = append(attrs, attribute.String("session.id", sessionID))
	}

	if errific.IsRetryable(err) {
		attrs = append(attrs, attribute.Bool("error.retryable", true))

		if retryAfter := errific.GetRetryAfter(err); retryAfter > 0 {
			attrs = append(attrs, attribute.String("error.retry_after", retryAfter.String()))
		}

		if maxRetries := errific.GetMaxRetries(err); maxRetries > 0 {
			attrs = append(attrs, attribute.Int("error.max_retries", maxRetries))
		}
	}

	if httpStatus := errific.GetHTTPStatus(err); httpStatus > 0 {
		attrs = append(attrs, attribute.Int("http.status_code", httpStatus))
	}

	if mcpCode := errific.GetMCPCode(err); mcpCode != 0 {
		attrs = append(attrs, attribute.Int("mcp.error_code", mcpCode))
	}

	// Add tags as array attribute
	if tags := errific.GetTags(err); len(tags) > 0 {
		attrs = append(attrs, attribute.StringSlice("error.tags", tags))
	}

	// Add labels as individual attributes with "label." prefix
	if labels := errific.GetLabels(err); len(labels) > 0 {
		for key, value := range labels {
			attrs = append(attrs, attribute.String("label."+key, value))
		}
	}

	// Add structured context as attributes with "context." prefix
	if context := errific.GetContext(err); len(context) > 0 {
		for key, value := range context {
			// Convert value to string for OpenTelemetry
			attrs = append(attrs, attribute.String("context."+key, fmt.Sprint(value)))
		}
	}

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
}

// RecordErrorWithEvent records an error to a span and adds a custom error event.
//
// This is useful when you want to add additional context beyond the standard span attributes.
//
// Example:
//
//	otel.RecordErrorWithEvent(span, err, "database_connection_failed", map[string]string{
//	    "pool_size": "10",
//	    "active_connections": "10",
//	})
func RecordErrorWithEvent(span trace.Span, err error, eventName string, eventAttrs map[string]string) {
	RecordError(span, err)

	if span == nil || eventName == "" {
		return
	}

	attrs := make([]attribute.KeyValue, 0, len(eventAttrs))
	for k, v := range eventAttrs {
		attrs = append(attrs, attribute.String(k, v))
	}

	span.AddEvent(eventName, trace.WithAttributes(attrs...))
}

// AddErrorContext adds errific error metadata to the current span without changing its status.
//
// This is useful when you want to add error context for debugging but the operation
// hasn't actually failed (e.g., handled errors, warnings, retried operations).
//
// Example:
//
//	if err := tryOperation(); err != nil {
//	    otel.AddErrorContext(span, err)  // Add context without marking as failed
//	    // Try alternative approach
//	    if err2 := alternativeOperation(); err2 == nil {
//	        return nil  // Succeeded with alternative, span status remains OK
//	    }
//	}
func AddErrorContext(span trace.Span, err error) {
	if err == nil || span == nil {
		return
	}

	attrs := make([]attribute.KeyValue, 0, 8)

	if code := errific.GetCode(err); code != "" {
		attrs = append(attrs, attribute.String("error.attempted.code", code))
	}

	if category := errific.GetCategory(err); category != "" {
		attrs = append(attrs, attribute.String("error.attempted.category", string(category)))
	}

	if correlationID := errific.GetCorrelationID(err); correlationID != "" {
		attrs = append(attrs, attribute.String("correlation.id", correlationID))
	}

	attrs = append(attrs, attribute.String("error.attempted.message", err.Error()))

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
}
