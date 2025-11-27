package otel_test

import (
	"context"
	"fmt"
	"time"

	"github.com/leefernandes/errific"
	"github.com/leefernandes/errific/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

// Example_basicUsage demonstrates the simplest way to use otel.RecordError
func Example_basicUsage() {
	tracer := noop.NewTracerProvider().Tracer("example")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "ProcessOrder")
	defer span.End()

	// Your business logic
	err := processOrder("order-123")
	if err != nil {
		// One-liner to record error with all metadata
		otel.RecordError(span, err)
		fmt.Println("Error recorded to span")
		return
	}
}

// Example_withRetryLogic shows how otel integration works with retry metadata
func Example_withRetryLogic() {
	tracer := noop.NewTracerProvider().Tracer("example")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "CallExternalAPI")
	defer span.End()

	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		err := callExternalAPI("https://api.example.com/users")
		if err == nil {
			fmt.Println("Success")
			return
		}

		lastErr = err

		// Record error to span (includes retry metadata)
		otel.RecordError(span, err)

		if !errific.IsRetryable(err) {
			break
		}

		delay := errific.GetRetryAfter(err)
		fmt.Printf("Retrying after %v\n", delay)
		time.Sleep(delay)
	}

	fmt.Printf("Failed after retries: %v\n", lastErr)
}

// Example_handledError shows using AddErrorContext for errors that don't fail the operation
func Example_handledError() {
	tracer := noop.NewTracerProvider().Tracer("example")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "FetchData")
	defer span.End()

	// Try primary data source
	err := fetchFromPrimary()
	if err != nil {
		// Add context but don't mark span as failed
		otel.AddErrorContext(span, err)

		// Try fallback
		err = fetchFromFallback()
		if err != nil {
			// This one actually failed
			otel.RecordError(span, err)
			fmt.Println("All sources failed")
			return
		}
	}

	fmt.Println("Data fetched successfully")
}

// Example_customEvent demonstrates adding a custom event with additional context
func Example_customEvent() {
	tracer := noop.NewTracerProvider().Tracer("example")
	ctx := context.Background()

	_, span := tracer.Start(ctx, "DatabaseOperation")
	defer span.End()

	err := connectToDatabase()
	if err != nil {
		// Record error with a custom event
		otel.RecordErrorWithEvent(span, err, "database_connection_failed", map[string]string{
			"pool_size":          "10",
			"active_connections": "10",
			"wait_time_ms":       "5000",
		})
		fmt.Println("Database connection failed with details")
		return
	}
}

// Example_microserviceChain shows error tracking across microservices
func Example_microserviceChain() {
	tracer := noop.NewTracerProvider().Tracer("example")
	ctx := context.Background()

	correlationID := "trace-abc-123"

	// Service A: Gateway
	_, spanA := tracer.Start(ctx, "Gateway.HandleRequest")
	defer spanA.End()

	err := callUserService(correlationID)
	if err != nil {
		otel.RecordError(spanA, err)
		fmt.Printf("Gateway error with correlation_id: %s\n", errific.GetCorrelationID(err))
		return
	}
}

// Helper functions for examples

var (
	ErrOrderNotFound   errific.Err = "order not found"
	ErrAPITimeout      errific.Err = "API timeout"
	ErrPrimaryDown     errific.Err = "primary source unavailable"
	ErrFallbackDown    errific.Err = "fallback source unavailable"
	ErrDatabaseConn    errific.Err = "database connection failed"
	ErrUserServiceDown errific.Err = "user service unavailable"
)

func processOrder(orderID string) error {
	return ErrOrderNotFound.New().
		WithCode("ORD_NOT_FOUND").
		WithCategory(errific.CategoryNotFound).
		WithHTTPStatus(404).
		WithContext(errific.Context{"order_id": orderID})
}

func callExternalAPI(endpoint string) error {
	return ErrAPITimeout.New().
		WithCode("API_TIMEOUT_001").
		WithCategory(errific.CategoryTimeout).
		WithRetryable(true).
		WithRetryAfter(2 * time.Second).
		WithMaxRetries(3).
		WithContext(errific.Context{"endpoint": endpoint})
}

func fetchFromPrimary() error {
	return ErrPrimaryDown.New().
		WithCode("PRIMARY_UNAVAILABLE").
		WithCategory(errific.CategoryNetwork)
}

func fetchFromFallback() error {
	// Simulate success
	return nil
}

func connectToDatabase() error {
	return ErrDatabaseConn.New().
		WithCode("DB_CONN_POOL_EXHAUSTED").
		WithCategory(errific.CategoryServer).
		WithContext(errific.Context{
			"pool_size":          10,
			"active_connections": 10,
		})
}

func callUserService(correlationID string) error {
	return ErrUserServiceDown.New().
		WithCode("USER_SVC_DOWN").
		WithCategory(errific.CategoryNetwork).
		WithCorrelationID(correlationID).
		WithLabel("service", "user-service")
}

// MockSpan removed - examples use noop spans which is sufficient
