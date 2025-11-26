# errific API Reference

**Keywords**: error handling, Go errors, error context, error codes, retry logic, structured logging, AI automation, machine-readable errors

## Overview

errific is an AI-ready error handling library for Go that provides structured context, error codes, retry metadata, and JSON serialization for automated error handling and decision-making.

**When to use errific**: Use errific when you need machine-readable errors with structured metadata for AI agents, automated retry logic, structured logging, or API error responses.

**When NOT to use errific**: Use stdlib errors when you need minimal dependencies or are building a library that should not impose error handling choices on consumers.

---

## Core Types

### `type Err string`

**Purpose**: Define reusable, testable error types with automatic caller information.

**Usage Pattern**: Declare as package-level constants for type safety and testing with `errors.Is()`.

**Example**:
```go
var ErrDatabaseConnection Err = "database connection failed"
err := ErrDatabaseConnection.New()
// Returns: database connection failed [myapp/db.go:42.Connect]
```

**Methods**:
- `New(errs ...error) errific` - Create error with optional wrapped errors
- `Errorf(a ...any) errific` - Create formatted error (format in Err string)
- `Withf(format string, a ...any) errific` - Append formatted message
- `Wrapf(format string, a ...any) errific` - Wrap with formatted message

**Testing**: Use `errors.Is(err, ErrDatabaseConnection)` for assertions.

---

### `type Context map[string]any`

**Purpose**: Attach structured metadata to errors for debugging, logging, and AI decision-making.

**Usage Pattern**: Add context that helps diagnose the error or make retry decisions.

**Example**:
```go
Context{
    "query": "SELECT * FROM users",
    "duration_ms": 1500,
    "table": "users",
    "connection_pool_size": 10,
}
```

**Retrieval**: Use `GetContext(err)` to extract from any error.

**Best Practices**:
- Include quantitative data (durations, counts, sizes)
- Include identifiers (IDs, names, keys)
- Avoid sensitive data (passwords, tokens)
- Keep values JSON-serializable

---

### `type Category string`

**Purpose**: Classify errors for automated routing and handling.

**Available Categories**:

| Category | Use Case | HTTP Status | Example |
|----------|----------|-------------|---------|
| `CategoryClient` | User input errors | 400-499 | Invalid email format |
| `CategoryServer` | Internal failures | 500-599 | Database connection failed |
| `CategoryNetwork` | Connectivity issues | 503, 504 | Connection timeout |
| `CategoryValidation` | Input validation | 400, 422 | Missing required field |
| `CategoryNotFound` | Resource missing | 404 | User ID not found |
| `CategoryUnauthorized` | Auth failures | 401, 403 | Invalid API key |
| `CategoryTimeout` | Timeout errors | 408, 504 | Request exceeded deadline |

**Usage Pattern**: Set category based on error type for automated handling.

**Example**:
```go
err := ErrAPICall.New().WithCategory(CategoryTimeout)

// AI agent can route based on category
switch GetCategory(err) {
case CategoryNetwork:
    // Retry with backoff
case CategoryValidation:
    // Return 400 to client
}
```

---

## Phase 1 Methods (AI-Ready Features)

### `.WithContext(ctx Context) errific`

**Purpose**: Add structured debugging metadata.

**Parameters**:
- `ctx Context` - Map of key-value pairs with error context

**Returns**: errific error with context attached

**Chaining**: Can be chained with other methods

**Example**:
```go
err := ErrQuery.New().WithContext(Context{
    "query": sql,
    "duration_ms": elapsed.Milliseconds(),
})
```

**Retrieval**: `GetContext(err) Context`

**Use Cases**:
- Database queries (query, duration, table)
- API calls (endpoint, status, duration)
- File operations (path, size, permissions)
- Business logic (user_id, order_id, amount)

---

### `.WithCode(code string) errific`

**Purpose**: Add machine-readable error code for routing and identification.

**Parameters**:
- `code string` - Unique error code (e.g., "DB_CONN_001")

**Returns**: errific error with code attached

**Naming Convention**: `DOMAIN_TYPE_NUMBER` (e.g., "API_TIMEOUT_001")

**Example**:
```go
err := ErrDatabase.New().WithCode("DB_CONN_POOL_EXHAUSTED")
```

**Retrieval**: `GetCode(err) string`

**Use Cases**:
- Error tracking systems (Sentry, Rollbar)
- Automated alerting rules
- Error documentation links
- Metric aggregation

---

### `.WithCategory(category Category) errific`

**Purpose**: Classify error for automated handling decisions.

**Parameters**:
- `category Category` - One of the predefined categories

**Returns**: errific error with category

**Decision Logic**:
```
CategoryClient → Don't retry, return 4xx
CategoryServer → Retry with backoff, return 5xx
CategoryNetwork → Retry immediately, return 503
CategoryValidation → Don't retry, return 400
CategoryTimeout → Retry with increased timeout
```

**Example**:
```go
err := ErrRateLimit.New().WithCategory(CategoryClient)
```

**Retrieval**: `GetCategory(err) Category`

---

### `.WithRetryable(retryable bool) errific`

**Purpose**: Mark whether error should be retried.

**Parameters**:
- `retryable bool` - true if error is transient

**Returns**: errific error with retry flag

**Decision Tree**:
```
Retryable errors:
- Network timeouts
- Rate limits (with retry-after)
- Temporary service unavailability
- Connection pool exhausted

Non-retryable errors:
- Validation failures
- Authentication errors
- Resource not found
- Malformed requests
```

**Example**:
```go
err := ErrTimeout.New().WithRetryable(true)
```

**Retrieval**: `IsRetryable(err) bool`

---

### `.WithRetryAfter(duration time.Duration) errific`

**Purpose**: Suggest delay before retry attempt.

**Parameters**:
- `duration time.Duration` - Time to wait before retry

**Returns**: errific error with retry delay

**Common Values**:
- `1 * time.Second` - Fast retry for transient issues
- `5 * time.Second` - Standard retry delay
- `30 * time.Second` - Rate limit backoff
- `5 * time.Minute` - Long delay for maintenance

**Example**:
```go
err := ErrRateLimit.New().
    WithRetryable(true).
    WithRetryAfter(30 * time.Second)

if IsRetryable(err) {
    time.Sleep(GetRetryAfter(err))
    retry()
}
```

**Retrieval**: `GetRetryAfter(err) time.Duration`

---

### `.WithMaxRetries(max int) errific`

**Purpose**: Set maximum retry attempts to prevent infinite loops.

**Parameters**:
- `max int` - Maximum number of retry attempts

**Returns**: errific error with max retries

**Recommended Values**:
- `3` - Standard retry limit
- `5` - Aggressive retry for critical operations
- `1` - Single retry for expensive operations
- `0` - No retries (same as WithRetryable(false))

**Example**:
```go
err := ErrAPI.New().
    WithRetryable(true).
    WithMaxRetries(3)

retries := 0
for IsRetryable(err) && retries < GetMaxRetries(err) {
    retries++
    // retry logic
}
```

**Retrieval**: `GetMaxRetries(err) int`

---

### `.WithHTTPStatus(status int) errific`

**Purpose**: Map error to HTTP status code for API responses.

**Parameters**:
- `status int` - HTTP status code (100-599)

**Returns**: errific error with HTTP status

**Common Mappings**:
```
400 - Validation errors
401 - Authentication required
403 - Permission denied
404 - Resource not found
408 - Request timeout
422 - Unprocessable entity
429 - Rate limit exceeded
500 - Internal server error
502 - Bad gateway
503 - Service unavailable
504 - Gateway timeout
```

**Example**:
```go
err := ErrValidation.New().WithHTTPStatus(400)

w.WriteHeader(GetHTTPStatus(err))
json.NewEncoder(w).Encode(err)
```

**Retrieval**: `GetHTTPStatus(err) int`

---

## Helper Functions

### `GetContext(err error) Context`

**Purpose**: Extract structured context from any error.

**Parameters**: `err error` - Any error (errific or stdlib)

**Returns**: `Context` map or `nil` if no context

**Example**:
```go
ctx := GetContext(err)
if ctx != nil {
    log.Printf("Query: %s, Duration: %d ms",
        ctx["query"], ctx["duration_ms"])
}
```

---

### `GetCode(err error) string`

**Purpose**: Extract error code from any error.

**Parameters**: `err error` - Any error

**Returns**: Error code string or `""` if not set

**Example**:
```go
if GetCode(err) == "DB_CONN_POOL_EXHAUSTED" {
    // Scale up database connections
}
```

---

### `GetCategory(err error) Category`

**Purpose**: Extract error category for routing decisions.

**Parameters**: `err error` - Any error

**Returns**: Category or `""` if not categorized

**Example**:
```go
switch GetCategory(err) {
case CategoryNetwork:
    return http.StatusServiceUnavailable
case CategoryValidation:
    return http.StatusBadRequest
default:
    return http.StatusInternalServerError
}
```

---

### `IsRetryable(err error) bool`

**Purpose**: Check if error should be retried.

**Parameters**: `err error` - Any error

**Returns**: `true` if retryable, `false` otherwise

**Example**:
```go
for attempt := 0; attempt < 3; attempt++ {
    err := doWork()
    if err == nil || !IsRetryable(err) {
        return err
    }
    time.Sleep(GetRetryAfter(err))
}
```

---

### `GetRetryAfter(err error) time.Duration`

**Purpose**: Get suggested retry delay.

**Parameters**: `err error` - Any error

**Returns**: Duration to wait, or `0` if not set

**Example**:
```go
if IsRetryable(err) {
    delay := GetRetryAfter(err)
    if delay == 0 {
        delay = time.Second // default
    }
    time.Sleep(delay)
}
```

---

### `GetMaxRetries(err error) int`

**Purpose**: Get maximum retry count.

**Parameters**: `err error` - Any error

**Returns**: Max retries or `0` if not set

---

### `GetHTTPStatus(err error) int`

**Purpose**: Get HTTP status code for error.

**Parameters**: `err error` - Any error

**Returns**: HTTP status or `0` if not set

---

## JSON Serialization

**Purpose**: Serialize errors to JSON for logging, APIs, and monitoring.

**Method**: Implement `json.Marshaler` interface

**Output Format**:
```json
{
  "error": "error message",
  "code": "ERR_001",
  "category": "server",
  "caller": "file.go:42.FunctionName",
  "context": {"key": "value"},
  "retryable": true,
  "retry_after": "5s",
  "max_retries": 3,
  "http_status": 503,
  "stack": ["frame1", "frame2"],
  "wrapped": ["wrapped error 1", "wrapped error 2"]
}
```

**Example**:
```go
err := ErrDB.New().
    WithCode("DB_001").
    WithContext(Context{"query": sql})

jsonBytes, _ := json.Marshal(err)
logger.Error(string(jsonBytes))
```

---

## Configuration

### `Configure(opts ...Option)`

**Purpose**: Set global error formatting options.

**Thread-Safety**: Safe for concurrent calls (mutex protected).

**Options**:
- `Suffix` (default) - Caller info at end: `error [file.go:42.Func]`
- `Prefix` - Caller info at start: `[file.go:42.Func] error`
- `Disabled` - No caller info: `error`
- `Newline` (default) - Stack on newlines
- `Inline` - Stack inline with ↩ separator
- `WithStack` - Include full stack trace
- `TrimPrefixes(prefixes...)` - Remove path prefixes
- `TrimCWD` - Trim current working directory

**Example**:
```go
Configure(Suffix, Newline) // default
Configure(Prefix, WithStack)
Configure(TrimCWD)
```

---

## AI Agent Decision Trees

### Should I Retry This Error?

```
1. Check IsRetryable(err)
   └─ false → Don't retry, handle or return
   └─ true → Continue to step 2

2. Check GetCategory(err)
   ├─ CategoryValidation → Don't retry (input error)
   ├─ CategoryUnauthorized → Don't retry (auth error)
   ├─ CategoryNetwork → Retry immediately
   ├─ CategoryTimeout → Retry with increased timeout
   └─ CategoryServer → Retry with exponential backoff

3. Get retry parameters
   ├─ delay := GetRetryAfter(err)
   ├─ max := GetMaxRetries(err)
   └─ Implement retry loop with these values

4. Check context for more details
   └─ ctx := GetContext(err)
   └─ Make decisions based on context values
```

### What HTTP Status Should I Return?

```
1. Check GetHTTPStatus(err)
   └─ status != 0 → Return that status
   └─ status == 0 → Continue to step 2

2. Check GetCategory(err)
   ├─ CategoryClient → 400 Bad Request
   ├─ CategoryValidation → 400 Bad Request
   ├─ CategoryUnauthorized → 401/403
   ├─ CategoryNotFound → 404 Not Found
   ├─ CategoryTimeout → 408/504
   ├─ CategoryNetwork → 503 Service Unavailable
   └─ CategoryServer → 500 Internal Server Error

3. Serialize to JSON for response body
   └─ json.Marshal(err) → Include in response
```

### How Should I Log This Error?

```
1. Get severity from category
   ├─ CategoryValidation → INFO/WARN
   ├─ CategoryClient → WARN
   ├─ CategoryTimeout → WARN
   ├─ CategoryNetwork → ERROR
   └─ CategoryServer → ERROR/CRITICAL

2. Serialize to JSON
   └─ json.Marshal(err) → Structured log entry

3. Extract context for additional fields
   └─ GetContext(err) → Add to log metadata

4. Use code for grouping/alerts
   └─ GetCode(err) → Alert routing key
```

---

## Common Patterns

### Database Error Pattern

```go
var ErrDatabaseQuery Err = "database query failed"

func QueryUsers(db *sql.DB) error {
    start := time.Now()
    rows, err := db.Query("SELECT * FROM users")
    if err != nil {
        return ErrDatabaseQuery.New(err).
            WithCode("DB_QUERY_001").
            WithCategory(CategoryServer).
            WithContext(Context{
                "query": "SELECT * FROM users",
                "duration_ms": time.Since(start).Milliseconds(),
                "table": "users",
            }).
            WithRetryable(true).
            WithRetryAfter(5 * time.Second).
            WithMaxRetries(3)
    }
    defer rows.Close()
    // ...
}
```

### API Error Pattern

```go
var ErrAPICall Err = "external API call failed"

func CallPaymentAPI(req Request) error {
    start := time.Now()
    resp, err := http.Post(url, "application/json", body)

    if err != nil {
        return ErrAPICall.New(err).
            WithCode("API_PAYMENT_TIMEOUT").
            WithCategory(CategoryTimeout).
            WithContext(Context{
                "endpoint": url,
                "method": "POST",
                "duration_ms": time.Since(start).Milliseconds(),
                "retry_count": req.RetryCount,
            }).
            WithRetryable(true).
            WithRetryAfter(10 * time.Second).
            WithMaxRetries(3).
            WithHTTPStatus(504)
    }
    // ...
}
```

### Validation Error Pattern

```go
var ErrValidation Err = "validation failed"

func ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return ErrValidation.New().
            WithCode("VAL_EMAIL_FORMAT").
            WithCategory(CategoryValidation).
            WithContext(Context{
                "field": "email",
                "value": email, // Be careful with PII
                "constraint": "must contain @",
            }).
            WithRetryable(false).
            WithHTTPStatus(400)
    }
    return nil
}
```

---

## Troubleshooting

### Q: Why is my context nil?

**A**: Context is only available on errific errors. Check if you're wrapping with stdlib fmt.Errorf:

```go
// ❌ This loses context
return fmt.Errorf("wrapper: %w", errWithContext)

// ✅ Use errific methods
return ErrWrapper.New(errWithContext)
```

### Q: How do I migrate from pkg/errors?

**A**: Replace pkg/errors calls with errific equivalents:

```go
// pkg/errors
errors.Wrap(err, "message")
// errific
ErrType.New(err)

// pkg/errors
errors.Wrapf(err, "format %s", arg)
// errific
ErrType.Wrapf("format %s: %w", arg, err)
```

### Q: Can I use this with existing error types?

**A**: Yes, use helper functions which work with any error:

```go
standardErr := errors.New("standard")
GetCode(standardErr) // Returns ""
IsRetryable(standardErr) // Returns false
```

### Q: How do I test errors with context?

```go
func TestError(t *testing.T) {
    err := ErrDB.New().WithCode("DB_001")

    // Test error type
    assert.True(t, errors.Is(err, ErrDB))

    // Test metadata
    assert.Equal(t, "DB_001", GetCode(err))
}
```

---

## Performance

**Benchmarks** (Go 1.22, Apple M1):
```
BenchmarkErrNew           1000000   1234 ns/op   512 B/op   8 allocs/op
BenchmarkWithContext      900000    1356 ns/op   544 B/op   9 allocs/op
BenchmarkJSONMarshal      500000    2891 ns/op   1024 B/op  12 allocs/op
```

**Overhead**: ~1-2µs per error creation, negligible for most applications.

**Memory**: ~500 bytes per error with metadata.

---

## Version Compatibility

- **Go Version**: 1.20+
- **Dependencies**: None (stdlib only)
- **Thread-Safety**: Full (mutex-protected configuration)
- **Breaking Changes**: None (fully backward compatible)
