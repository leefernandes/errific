# errific/otel - OpenTelemetry Integration

OpenTelemetry integration helpers for errific errors. This package provides convenience functions to seamlessly record errific errors to OpenTelemetry spans with all metadata preserved.

## Installation

```bash
go get github.com/leefernandes/errific/otel
```

## Features

- ✅ **One-liner error recording** - `otel.RecordError(span, err)`
- ✅ **Automatic metadata extraction** - All errific fields become span attributes
- ✅ **Standard compliant** - Follows OpenTelemetry semantic conventions
- ✅ **Zero configuration** - Works out of the box
- ✅ **Backward compatible** - Works with any error type, not just errific

## Quick Start

```go
package main

import (
    "context"
    "github.com/leefernandes/errific"
    "github.com/leefernandes/errific/otel"
    "go.opentelemetry.io/otel"
)

var ErrDatabase errific.Err = "database query failed"

func ProcessOrder(ctx context.Context, orderID string) error {
    tracer := otel.Tracer("order-service")
    ctx, span := tracer.Start(ctx, "ProcessOrder")
    defer span.End()
    
    if err := queryDatabase(orderID); err != nil {
        // One line records everything!
        otel.RecordError(span, err)
        return err
    }
    
    return nil
}

func queryDatabase(orderID string) error {
    // Simulate error with rich metadata
    return ErrDatabase.New().
        WithCode("DB_QUERY_001").
        WithCategory(errific.CategoryServer).
        WithCorrelationID("trace-abc-123").
        WithContext(errific.Context{
            "order_id": orderID,
            "query": "SELECT * FROM orders WHERE id = ?",
        })
}
```

## What Gets Recorded

When you call `otel.RecordError(span, err)`, the following happens automatically:

1. **Span status** → Set to `Error`
2. **Exception event** → Recorded with `RecordException(err)`
3. **Span attributes** → All errific metadata added:

| errific Field | OpenTelemetry Attribute | Example |
|--------------|------------------------|---------|
| Code | `error.code` | `"DB_QUERY_001"` |
| Category | `error.category` | `"server"` |
| CorrelationID | `correlation.id` | `"trace-abc-123"` |
| RequestID | `request.id` | `"req-456"` |
| UserID | `user.id` | `"user-789"` |
| SessionID | `session.id` | `"sess-abc"` |
| Retryable | `error.retryable` | `true` |
| RetryAfter | `error.retry_after` | `"5s"` |
| MaxRetries | `error.max_retries` | `3` |
| HTTPStatus | `http.status_code` | `500` |
| MCPCode | `mcp.error_code` | `-32000` |
| Tags | `error.tags` | `["database", "timeout"]` |
| Labels | `label.*` | `label.service="user-svc"` |
| Context | `context.*` | `context.query="SELECT..."` |

## API Reference

### RecordError

Records an error to a span with full metadata extraction.

```go
func RecordError(span trace.Span, err error)
```

**Example**:
```go
if err := doSomething(); err != nil {
    otel.RecordError(span, err)
    return err
}
```

### RecordErrorWithEvent

Records an error and adds a custom event with additional attributes.

```go
func RecordErrorWithEvent(span trace.Span, err error, eventName string, eventAttrs map[string]string)
```

**Example**:
```go
otel.RecordErrorWithEvent(span, err, "database_connection_failed", map[string]string{
    "pool_size": "10",
    "active_connections": "10",
    "wait_time_ms": "5000",
})
```

### AddErrorContext

Adds error metadata to span without marking it as failed. Useful for handled errors.

```go
func AddErrorContext(span trace.Span, err error)
```

**Example**:
```go
// Try primary source
if err := fetchFromPrimary(); err != nil {
    otel.AddErrorContext(span, err)  // Record attempt, don't fail
    
    // Try fallback (operation succeeds overall)
    return fetchFromFallback()
}
```

## Usage Patterns

### Pattern 1: Basic Error Recording

```go
func HandleRequest(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "HandleRequest")
    defer span.End()
    
    if err := processRequest(); err != nil {
        otel.RecordError(span, err)
        return err
    }
    return nil
}
```

### Pattern 2: Retry Logic with Tracing

```go
func CallExternalAPI(ctx context.Context, endpoint string) error {
    ctx, span := tracer.Start(ctx, "CallExternalAPI")
    defer span.End()
    
    var lastErr error
    for attempt := 1; attempt <= 3; attempt++ {
        err := httpClient.Get(endpoint)
        if err == nil {
            return nil  // Success
        }
        
        lastErr = err
        otel.RecordError(span, err)  // Record each attempt
        
        if !errific.IsRetryable(err) {
            break
        }
        
        time.Sleep(errific.GetRetryAfter(err))
    }
    
    return lastErr
}
```

### Pattern 3: Microservice Chain Tracing

```go
// Service A: API Gateway
func Gateway_HandleRequest(ctx context.Context, userID string) error {
    ctx, span := tracer.Start(ctx, "Gateway.HandleRequest")
    defer span.End()
    
    correlationID := uuid.New().String()
    
    user, err := userService.GetUser(ctx, userID, correlationID)
    if err != nil {
        otel.RecordError(span, err)  // Includes correlation_id
        return err
    }
    return nil
}

// Service B: User Service
func UserService_GetUser(ctx context.Context, userID, correlationID string) error {
    ctx, span := tracer.Start(ctx, "UserService.GetUser")
    defer span.End()
    
    err := database.Query(userID)
    if err != nil {
        // Error propagates with same correlation_id
        err = ErrUserQuery.New(err).
            WithCorrelationID(correlationID).
            WithLabel("service", "user-service")
        
        otel.RecordError(span, err)
        return err
    }
    return nil
}
```

### Pattern 4: Graceful Degradation

```go
func FetchData(ctx context.Context) ([]byte, error) {
    ctx, span := tracer.Start(ctx, "FetchData")
    defer span.End()
    
    // Try cache first
    data, err := cache.Get("key")
    if err != nil {
        // Record attempt but don't fail span
        otel.AddErrorContext(span, err)
        
        // Fallback to database
        data, err = database.Get("key")
        if err != nil {
            // Now actually fail
            otel.RecordError(span, err)
            return nil, err
        }
    }
    
    return data, nil
}
```

## Span Attributes in Action

Given this error:

```go
err := ErrDatabase.New().
    WithCode("DB_CONN_001").
    WithCategory(errific.CategoryServer).
    WithCorrelationID("trace-abc-123").
    WithRequestID("req-456").
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithContext(errific.Context{
        "query": "SELECT * FROM users",
        "duration_ms": 1500,
    }).
    WithTags("database", "connection", "timeout").
    WithLabel("service", "user-service")
```

Your span will have these attributes:

```json
{
  "span": {
    "name": "QueryDatabase",
    "status": "ERROR",
    "attributes": {
      "error.code": "DB_CONN_001",
      "error.category": "server",
      "correlation.id": "trace-abc-123",
      "request.id": "req-456",
      "error.retryable": true,
      "error.retry_after": "5s",
      "error.tags": ["database", "connection", "timeout"],
      "label.service": "user-service",
      "context.query": "SELECT * FROM users",
      "context.duration_ms": "1500"
    },
    "events": [
      {
        "name": "exception",
        "attributes": {
          "exception.type": "errific.errific",
          "exception.message": "database query failed"
        }
      }
    ]
  }
}
```

## Performance

The otel package adds minimal overhead:

```
BenchmarkRecordError          1,000,000    ~850 ns/op    512 B/op
BenchmarkRecordError_Minimal  2,000,000    ~420 ns/op    256 B/op
BenchmarkAddErrorContext      2,500,000    ~380 ns/op    192 B/op
```

- Sub-microsecond for most operations
- Negligible compared to network I/O or trace export
- No allocations for nil checks

## Best Practices

### ✅ DO: Use RecordError for actual failures

```go
if err := criticalOperation(); err != nil {
    otel.RecordError(span, err)  // Operation failed
    return err
}
```

### ✅ DO: Use AddErrorContext for handled errors

```go
if err := tryCache(); err != nil {
    otel.AddErrorContext(span, err)  // Informational only
    return tryDatabase()  // Succeeded with fallback
}
```

### ✅ DO: Record errors at the right level

```go
// Record at the operation level, not at every function
func HandleRequest(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "HandleRequest")
    defer span.End()
    
    if err := step1(); err != nil {
        otel.RecordError(span, err)  // ✅ Record here
        return err
    }
    return nil
}

func step1() error {
    // Don't create span here, just return error
    return ErrStep1.New()  // ✅ Error without span
}
```

### ❌ DON'T: Record the same error multiple times

```go
// ❌ BAD: Recording same error in multiple spans
func A() error {
    span := tracer.Start(ctx, "A")
    defer span.End()
    
    if err := B(); err != nil {
        otel.RecordError(span, err)  // ❌ Recorded here
        return err
    }
}

func B() error {
    span := tracer.Start(ctx, "B")
    defer span.End()
    
    if err := operation(); err != nil {
        otel.RecordError(span, err)  // ❌ Already recorded here
        return err
    }
}

// ✅ GOOD: Record once at the appropriate level
```

## Integration with Observability Platforms

The recorded attributes work seamlessly with:

- **Jaeger** - Full trace visualization with error attributes
- **Zipkin** - Error spans highlighted
- **Datadog APM** - Error tracking with custom tags
- **New Relic** - Error analytics with all metadata
- **Honeycomb** - Rich error context in traces
- **AWS X-Ray** - Error segments with annotations
- **Google Cloud Trace** - Error spans with labels

## Comparison: Before and After

### Before (Manual)

```go
if err := operation(); err != nil {
    span.SetStatus(codes.Error, err.Error())
    span.RecordException(err)
    span.SetAttributes(
        attribute.String("error.code", errific.GetCode(err)),
        attribute.String("error.category", string(errific.GetCategory(err))),
        attribute.String("correlation.id", errific.GetCorrelationID(err)),
        // ... 10+ more lines
    )
    return err
}
```

### After (One-liner)

```go
if err := operation(); err != nil {
    otel.RecordError(span, err)  // ✅ Everything automatic
    return err
}
```

## License

Same as errific (see main LICENSE file)

## Contributing

Contributions welcome! Please ensure:
- Tests pass: `go test ./...`
- Benchmarks don't regress
- Examples run successfully
