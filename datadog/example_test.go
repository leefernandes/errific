package datadog_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/leefernandes/errific"
	"github.com/leefernandes/errific/datadog"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Example_basicTracing demonstrates basic error recording to Datadog spans
func Example_basicTracing() {
	// Initialize Datadog tracer (normally done at app startup)
	tracer.Start()
	defer tracer.Stop()

	// Create a span
	span := tracer.StartSpan("process.order")

	// Your business logic
	err := processOrder("order-123")

	// Record error with full metadata (one line!)
	datadog.RecordError(span, err)

	fmt.Println("Error recorded to Datadog span")
}

// Example_structuredLogging shows how to create Datadog-compatible logs
func Example_structuredLogging() {
	var ErrDatabase errific.Err = "database connection failed"

	err := ErrDatabase.New().
		WithCode("DB_CONN_001").
		WithCategory(errific.CategoryServer).
		WithContext(errific.Context{
			"pool_size":   10,
			"retry_count": 3,
		})

	// Convert to Datadog log entry
	logEntry := datadog.ToLogEntry(err)

	// Set service info (unified service tagging)
	datadog.SetServiceInfo(logEntry, "checkout-api", "production", "1.0.4")

	// Serialize to JSON
	logBytes, _ := json.MarshalIndent(logEntry, "", "  ")
	fmt.Println(string(logBytes))
}

// Example_logTraceCorrelation shows how to correlate logs with traces
func Example_logTraceCorrelation() {
	tracer.Start()
	defer tracer.Stop()

	span, ctx := tracer.StartSpanFromContext(context.Background(), "api.request")

	err := handleRequest(ctx)
	if err != nil {
		// Create log entry
		logEntry := datadog.ToLogEntry(err)

		// Enrich with trace info for correlation
		datadog.EnrichLogEntry(logEntry, span)

		// Set service info
		datadog.SetServiceInfo(logEntry, "api-gateway", "production", "2.1.0")

		// Log it (will have dd.trace_id and dd.span_id)
		logBytes, _ := json.Marshal(logEntry)
		log.Println(string(logBytes))
	}

	datadog.RecordError(span, err)
	fmt.Println("Log correlated with trace")
}

// Example_microserviceChain demonstrates error tracking across services
func Example_microserviceChain() {
	tracer.Start()
	defer tracer.Stop()

	// Gateway service
	gatewaySpan := tracer.StartSpan("gateway.handle_request")
	correlationID := "trace-abc-123"

	// Call downstream service
	err := callUserService(correlationID)
	if err != nil {
		// Log with correlation ID
		logEntry := datadog.ToLogEntry(err)
		datadog.EnrichLogEntry(logEntry, gatewaySpan)
		datadog.SetServiceInfo(logEntry, "gateway", "production", "1.0.0")

		// Correlation ID propagated automatically
		fmt.Printf("Gateway error with correlation_id: %s\n", logEntry.CorrelationID)
	}

	datadog.RecordError(gatewaySpan, err)
}

// Example_retryableErrors shows retry metadata in logs and traces
func Example_retryableErrors() {
	tracer.Start()
	defer tracer.Stop()

	span := tracer.StartSpan("external.api_call")

	var ErrAPITimeout errific.Err = "API timeout"
	err := ErrAPITimeout.New().
		WithCode("API_TIMEOUT_001").
		WithCategory(errific.CategoryTimeout).
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3).
		WithContext(errific.Context{
			"endpoint": "https://api.example.com/users",
			"timeout":  "30s",
		})

	// Record to span (includes retry metadata)
	datadog.RecordError(span, err)

	// Also log it
	logEntry := datadog.ToLogEntry(err)
	datadog.EnrichLogEntry(logEntry, span)

	fmt.Printf("Retryable: %v, Retry After: %s\n", *logEntry.Retryable, logEntry.RetryAfter)
}

// Example_errorTracking demonstrates Datadog Error Tracking integration
func Example_errorTracking() {
	tracer.Start()
	defer tracer.Stop()

	span := tracer.StartSpan("payment.process")

	var ErrPayment errific.Err = "payment processing failed"
	err := ErrPayment.New().
		WithCode("PAYMENT_DECLINED").
		WithCategory(errific.CategoryClient).
		WithHTTPStatus(402).
		WithContext(errific.Context{
			"amount":        "99.99",
			"currency":      "USD",
			"card_last4":    "1234",
			"decline_code":  "insufficient_funds",
		}).
		WithLabel("payment_gateway", "stripe").
		WithLabel("merchant_id", "merch_123")

	// Record error (will appear in Datadog Error Tracking)
	datadog.RecordError(span, err)

	// Also create structured log for Error Tracking
	logEntry := datadog.ToLogEntry(err)
	datadog.EnrichLogEntry(logEntry, span)
	datadog.SetServiceInfo(logEntry, "payment-service", "production", "3.2.1")

	logBytes, _ := json.Marshal(logEntry)
	log.Println(string(logBytes))

	fmt.Println("Error tracked in Datadog")
}

// Example_customContext shows adding application-specific context
func Example_customContext() {
	var ErrOrder errific.Err = "order validation failed"

	err := ErrOrder.New().
		WithCode("ORDER_INVALID").
		WithCategory(errific.CategoryValidation)

	logEntry := datadog.ToLogEntry(err)

	// Add custom business context
	datadog.AddContext(logEntry, map[string]interface{}{
		"customer_id":   "CUST-12345",
		"customer_tier": "gold",
		"order_value":   1599.99,
		"items_count":   5,
		"shipping_zip":  "94107",
	})

	datadog.SetServiceInfo(logEntry, "order-service", "production", "1.5.2")

	logBytes, _ := json.Marshal(logEntry)
	fmt.Println(string(logBytes))
}

// Helper functions for examples

var (
	ErrOrderNotFound errific.Err = "order not found"
	ErrUserService   errific.Err = "user service unavailable"
	ErrRequest       errific.Err = "request failed"
)

func processOrder(orderID string) error {
	return ErrOrderNotFound.New().
		WithCode("ORD_NOT_FOUND").
		WithCategory(errific.CategoryNotFound).
		WithHTTPStatus(404).
		WithContext(errific.Context{"order_id": orderID})
}

func handleRequest(ctx context.Context) error {
	return ErrRequest.New().
		WithCode("REQ_FAILED").
		WithCategory(errific.CategoryServer)
}

func callUserService(correlationID string) error {
	return ErrUserService.New().
		WithCode("USER_SVC_DOWN").
		WithCategory(errific.CategoryNetwork).
		WithCorrelationID(correlationID).
		WithLabel("service", "user-service").
		WithLabel("region", "us-west-2")
}
