# errific API Reference

<!-- RAG: Complete API documentation for errific Go error handling library -->

**Keywords**: error handling, Go errors, error context, error codes, retry logic, structured logging, AI automation, machine-readable errors, MCP integration, distributed tracing

## Overview

errific is an AI-ready error handling library for Go that provides structured context, error codes, retry metadata, and JSON serialization for automated error handling and decision-making.

**When to use errific**: Use errific when you need machine-readable errors with structured metadata for AI agents, automated retry logic, structured logging, or API error responses.

---

## Core Types

<!-- RAG: Core type definitions - Err, Context, Category -->

### `type Err string`

<!-- RAG: Base error type for defining reusable, typed errors with automatic caller information -->

**Purpose**: Define reusable, testable error types with automatic caller information.

**Why this matters**:
- **Type Safety**: Use with `errors.Is()` for reliable error checking across your codebase
- **Reusability**: Define once at package level, use everywhere consistently
- **Testability**: Mock and assert error types easily without string comparison
- **Debugging**: Automatic caller information (file:line.function) captured at error creation
- **Zero Allocation**: Defining errors as constants has no runtime allocation overhead

**Usage Pattern**: Declare as package-level constants for type safety and testing with `errors.Is()`.

**When to use**:
- ✅ Defining package-level or application-level errors
- ✅ When you need `errors.Is()` compatibility for error checking
- ✅ When caller information is valuable for debugging
- ✅ When you want consistent error messages across your codebase

**Example - Basic**:
```go
// Use Case: Database connection error
// Keywords: database, connection, typed-error

var ErrDatabaseConnection Err = "database connection failed"

// Style 1: Explicit .New() (use when wrapping errors or caller info matters)
err := ErrDatabaseConnection.New(sqlErr)

// Style 2: Concise (recommended for new code without wrapped errors)
err := ErrDatabaseConnection.WithCode("DB_001").WithHTTPStatus(500)

// Returns: database connection failed [myapp/db.go:42.Connect]
```

**Example - With Wrapped Error**:
```go
// Use Case: Wrapping underlying database driver errors
// Keywords: error-wrapping, error-chain

sqlErr := sql.Open("postgres", connString)
if sqlErr != nil {
    return ErrDatabaseConnection.New(sqlErr)
}
// Returns: database connection failed: pq: connection refused [myapp/db.go:42.Connect]
```

**Example - Testing with errors.Is()**:
```go
// Use Case: Reliable error type checking in tests
// Keywords: testing, error-checking, errors-is

func TestDatabaseError(t *testing.T) {
    err := connectDatabase()

    // ✅ Correct: Type-safe error checking
    if !errors.Is(err, ErrDatabaseConnection) {
        t.Error("expected database connection error")
    }

    // ❌ Incorrect: String comparison (fragile)
    if err.Error() != "database connection failed" {
        // Breaks if caller info changes
    }
}
```

**Methods**:
- `New(errs ...error) errific` - Create error with optional wrapped errors
- `Errorf(a ...any) errific` - Create formatted error (format in Err string)
- `Withf(format string, a ...any) errific` - Append formatted message
- `Wrapf(format string, a ...any) errific` - Wrap with formatted message


**Forwarding Methods**: All `With()` methods (WithCode, WithHTTPStatus, etc.) can now be called directly on `Err` without calling `.New()` first. This provides a more concise API while maintaining full backwards compatibility.

**How It Works**: The first `With()` method automatically calls `.New()` internally, then returns an `errific` instance. Subsequent methods in the chain operate on that instance directly, so `.New()` is only called once per error.

**Testing**: Use `errors.Is(err, ErrDatabaseConnection)` for assertions.

---

### `type Context map[string]any`

<!-- RAG: Structured metadata map for attaching debugging context to errors -->

**Purpose**: Attach structured metadata to errors for debugging, logging, and AI decision-making.

**Why this matters**:
- **Debugging**: See exact parameters, state, and conditions that caused the error
- **Monitoring**: Extract metrics from error context (duration, size, count) for dashboards
- **AI Decision-Making**: Agents can read context values to decide actions (e.g., retry if duration > threshold)
- **Compliance**: Track required audit fields (user_id, transaction_id, ip_address)
- **Root Cause Analysis**: Context preserves state snapshot at error time
- **JSON Serialization**: Context maps directly to JSON for logging systems

**Usage Pattern**: Add context that helps diagnose the error or make retry decisions.

**When to include**:
- ✅ **Operation parameters**: query, endpoint, file_path, command
- ✅ **Identifiers**: user_id, request_id, transaction_id, correlation_id, session_id
- ✅ **Measurements**: duration_ms, size_bytes, retry_count, timeout_ms
- ✅ **State information**: current_step, pool_size, queue_depth, connection_count
- ✅ **Diagnostic data**: status_code, error_code, response_time

**When to exclude**:
- ❌ **Sensitive data**: passwords, tokens, API keys, credit cards, PII
- ❌ **Large data**: full request/response bodies (>1KB), binary data, images
- ❌ **Non-JSON types**: channels, functions, unsafe pointers, goroutines
- ❌ **Redundant data**: already in error message or obvious from error type

**Example - Database Query**:
```go
// Use Case: Debugging slow database queries
// Keywords: database, query, performance, debugging

Context{
    "query": "SELECT * FROM users WHERE status = ?",
    "duration_ms": 1534,           // Slow! Helps identify performance issue
    "table": "users",
    "connection_id": "conn-123",
    "rows_affected": 0,            // Shows query returned nothing
    "params": []interface{}{"active"},
}
```

**Example - API Call**:
```go
// Use Case: Debugging API timeout errors
// Keywords: api, http, timeout, monitoring

Context{
    "endpoint": "https://api.example.com/v1/users",
    "method": "POST",
    "status_code": 504,            // Gateway timeout
    "duration_ms": 30000,          // 30 seconds = timeout threshold
    "retry_count": 2,              // Already retried twice
    "request_id": "req-abc-123",   // Trace across logs
    "response_size": 0,            // No response received
}
```

**Example - AI Agent Reading Context**:
```go
// Use Case: AI agent makes decision based on context
// Keywords: ai-automation, decision-making, context-analysis

if ctx := GetContext(err); ctx != nil {
    // Agent: "Query took 1534ms, this is a slow query issue"
    if duration, ok := ctx["duration_ms"].(int); ok && duration > 1000 {
        log.Warn("slow query detected",
            "duration", duration,
            "query", ctx["query"])

        // AI decision: Add to slow query monitoring
        slowQueryMonitor.Track(ctx)
    }

    // Agent: "API call failed after 2 retries, don't retry again"
    if retryCount, ok := ctx["retry_count"].(int); ok && retryCount >= 2 {
        // Don't retry, max attempts reached
        return err
    }
}
```

**Example - File Operations**:
```go
// Use Case: File operation errors with context
// Keywords: file-io, file-operations, permissions

Context{
    "path": "/data/uploads/file.txt",
    "operation": "read",
    "size_bytes": 1048576,         // 1MB file
    "permissions": "0644",
    "owner": "www-data",
    "exists": true,                // File exists but can't read
}
```

**Retrieval**: Use `GetContext(err)` to extract from any error.

**Best Practices**:
- Include quantitative data (durations, counts, sizes)
- Include identifiers (IDs, names, keys)
- Avoid sensitive data (passwords, tokens)
- Keep values JSON-serializable
- Use consistent key names across your application
- Limit context size to <100 keys per error

---

### `type Category string`

<!-- RAG: Error classification categories for automated error routing and handling -->

**Purpose**: Classify errors for automated routing and handling.

**Why this matters**:
- **Automated Routing**: AI agents can route errors to appropriate handlers without hardcoding error types
- **HTTP Mapping**: Automatically map errors to correct HTTP status codes in API responses
- **Retry Decisions**: Categories indicate whether errors are retryable (network=yes, validation=no)
- **Logging Severity**: Different categories map to different log levels (server=error, validation=warn)
- **Alerting**: Alert different teams based on category (server→ops, validation→product)
- **Monitoring**: Group errors by category in dashboards for better insights

**Available Categories**:

| Category | Use Case | HTTP Status | Retryable | Example |
|----------|----------|-------------|-----------|---------|
| `CategoryClient` | User input errors | 400-499 | ❌ No | Invalid email format |
| `CategoryServer` | Internal failures | 500-599 | ✅ Maybe | Database connection failed |
| `CategoryNetwork` | Connectivity issues | 503, 504 | ✅ Yes | Connection timeout |
| `CategoryValidation` | Input validation | 400, 422 | ❌ No | Missing required field |
| `CategoryNotFound` | Resource missing | 404 | ❌ No | User ID not found |
| `CategoryUnauthorized` | Auth failures | 401, 403 | ❌ No | Invalid API key |
| `CategoryTimeout` | Timeout errors | 408, 504 | ✅ Yes | Request exceeded deadline |

**Usage Pattern**: Set category based on error type for automated handling.

**When to use each category**:

**CategoryValidation** - Use when:
- ✅ User provided invalid input (email, phone, format)
- ✅ Required field is missing
- ✅ Value doesn't match constraints (min/max, regex)
- ✅ Business rule violation (duplicate username)

**CategoryClient** - Use when:
- ✅ General client-side error not covered by other categories
- ✅ Malformed request
- ✅ Unsupported media type
- ✅ Request too large

**CategoryUnauthorized** - Use when:
- ✅ Missing or invalid authentication token
- ✅ Insufficient permissions
- ✅ Expired session
- ✅ API key revoked

**CategoryNotFound** - Use when:
- ✅ Resource doesn't exist (user, order, file)
- ✅ Endpoint doesn't exist (404)
- ✅ Record deleted

**CategoryTimeout** - Use when:
- ✅ Operation exceeded deadline
- ✅ Client timeout
- ✅ Gateway timeout
- ✅ Context deadline exceeded

**CategoryNetwork** - Use when:
- ✅ Connection refused
- ✅ DNS resolution failed
- ✅ Network unreachable
- ✅ Service temporarily unavailable

**CategoryServer** - Use when:
- ✅ Internal server error
- ✅ Database query failed
- ✅ File system error
- ✅ Panic recovered

**Example - Basic Usage**:
```go
// Use Case: Categorize API timeout for automated handling
// Keywords: category, timeout, api-error, automation

err := ErrAPICall.WithCategory(CategoryTimeout)

// AI agent can route based on category
switch GetCategory(err) {
case CategoryNetwork:
    // Retry immediately (network might recover)
    time.Sleep(1 * time.Second)
    return retry()

case CategoryTimeout:
    // Retry with longer timeout
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()
    return retryWithContext(ctx)

case CategoryValidation:
    // Don't retry, return 400 to client
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(err)
    return nil
}
```

**Example - HTTP Status Mapping**:
```go
// Use Case: Automatically map category to HTTP status
// Keywords: http-mapping, api-error, status-code

err := ErrUserNotFound.New().WithCategory(CategoryNotFound)

// Get HTTP status based on category
status := GetHTTPStatus(err)
if status == 0 {
    // No explicit status, use category default
    switch GetCategory(err) {
    case CategoryNotFound:
        status = 404
    case CategoryUnauthorized:
        status = 401
    case CategoryValidation:
        status = 400
    case CategoryTimeout:
        status = 504
    default:
        status = 500
    }
}

w.WriteHeader(status)
```

**Example - Logging Severity**:
```go
// Use Case: Set log level based on error category
// Keywords: logging, severity, monitoring

func logError(err error) {
    var level string
    switch GetCategory(err) {
    case CategoryValidation, CategoryClient:
        level = "warn"  // User errors, not critical
    case CategoryServer, CategoryNetwork:
        level = "error" // System errors, needs attention
    default:
        level = "info"
    }

    logger.Log(level, err.Error(),
        "category", GetCategory(err),
        "code", GetCode(err))
}
```

---

## API Styles

errific supports **two equivalent API styles** for creating errors:

### Style 1: Explicit `.New()` (Traditional)

**When to use**:
- Wrapping other errors: `ErrDatabase.New(sqlErr)`
- When caller information is critical for debugging
- When porting from existing code

**Example**:
```go
err := ErrDatabase.New(sqlErr).
    WithCode("DB_001").
    WithHTTPStatus(500)
```

### Style 2: Concise (Forwarding Methods)

**When to use** (recommended):
- Creating new errors without wrapping
- When code brevity is preferred
- New code and modern Go projects

**Example**:
```go
err := ErrDatabase.
    WithCode("DB_001").
    WithHTTPStatus(500)
```

**Key Difference**: The concise style automatically calls `.New()` on the first `With()` method. Both styles produce equivalent errors with the same metadata.

**Performance**: `.New()` is called exactly **once** regardless of how many methods are chained. Subsequent methods operate directly on the `errific` instance with zero overhead.

---

## Phase 1 Methods (AI-Ready Features)

<!-- RAG: Core methods for adding metadata - context, codes, categories, retry logic, HTTP status -->

### `.WithContext(ctx Context) errific`

<!-- RAG: Add structured debugging context to errors -->

**Purpose**: Add structured debugging metadata.

**Why this matters**:
- **Root Cause Analysis**: Preserve exact state and parameters when error occurred
- **Metrics Extraction**: Pull duration, count, size from errors for monitoring dashboards
- **AI Decisions**: Agents read context to make intelligent decisions (retry if slow, alert if large)
- **Audit Trails**: Track user_id, transaction_id, request_id for compliance
- **Performance Debugging**: Identify slow operations by analyzing duration_ms in context

**Parameters**:
- `ctx Context` - Map of key-value pairs with error context

**Returns**: errific error with context attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Database operations (include query, duration, table name)
- ✅ API calls (include endpoint, method, status code, duration)
- ✅ File operations (include path, operation, size, permissions)
- ✅ Business logic (include user_id, order_id, transaction_id)
- ✅ Any operation where state/parameters help debugging

**When NOT to use**:
- ❌ For sensitive data (use sanitized versions or omit)
- ❌ For large payloads (summarize instead)
- ❌ For obvious information (already in error message)

**Example (both styles work)**:
```go
// Use Case: Database query with performance tracking
// Keywords: database, context, performance, debugging

// Explicit style
err := ErrQuery.New().WithContext(Context{
    "query": sql,
    "duration_ms": elapsed.Milliseconds(),
    "table": "users",
    "rows_affected": count,
})

// Concise style (recommended)
err := ErrQuery.WithContext(Context{
    "query": sql,
    "duration_ms": elapsed.Milliseconds(),
})
```

**Example - API Call with Retry Context**:
```go
// Use Case: Track API call failures for retry decisions
// Keywords: api, http, retry-context, monitoring

err := ErrAPICall.New(httpErr).WithContext(Context{
    "endpoint": "https://api.example.com/v1/payment",
    "method": "POST",
    "status_code": resp.StatusCode,
    "duration_ms": elapsed.Milliseconds(),
    "retry_attempt": attempt,        // Track which retry this is
    "timeout_ms": 30000,
    "request_id": reqID,
})

// AI Agent Decision Based on Context:
if ctx := GetContext(err); ctx != nil {
    if attempt, ok := ctx["retry_attempt"].(int); ok && attempt >= 3 {
        // Already retried 3 times, don't retry again
        return err
    }
    if duration, ok := ctx["duration_ms"].(int); ok && duration > 25000 {
        // Slow response, might need longer timeout next time
        return retryWithTimeout(60 * time.Second)
    }
}
```

**Retrieval**: `GetContext(err) Context`

**Use Cases**:
- Database queries (query, duration, table)
- API calls (endpoint, status, duration)
- File operations (path, size, permissions)
- Business logic (user_id, order_id, amount)

---

### `.WithCode(code string) errific`

<!-- RAG: Add machine-readable error code for tracking, alerting, and metrics -->

**Purpose**: Add machine-readable error code for routing and identification.

**Why this matters**:
- **Error Tracking**: Unique codes make errors searchable in Sentry, Rollbar, Datadog
- **Automated Alerting**: Route alerts to specific teams based on error code prefix (DB_*, API_*, etc.)
- **Metric Aggregation**: Group errors by code for dashboards ("API_TIMEOUT_001: 1,247 today")
- **Documentation Links**: Map error codes to documentation pages automatically
- **Debugging**: Filter logs by error code to find all occurrences of specific issue
- **API Consistency**: Return consistent error codes to API clients for reliable error handling

**Parameters**:
- `code string` - Unique error code (e.g., "DB_CONN_001")
- Empty string is ignored (no-op)

**Returns**: errific error with code attached

**Naming Convention**: `DOMAIN_TYPE_NUMBER` (e.g., "API_TIMEOUT_001")

**When to use**:
- ✅ Errors that need tracking/monitoring in error tracking systems
- ✅ Errors that trigger specific team alerts
- ✅ Errors that map to documentation pages
- ✅ Public API errors (consistent client experience)
- ✅ Errors used for metrics/dashboards

**When NOT to use**:
- ❌ One-off internal errors with no tracking
- ❌ Overly generic codes (e.g., "ERROR_001")
- ❌ Test-only errors

**Example - Database Error with Code**:
```go
// Use Case: Track database connection pool exhaustion for alerts
// Keywords: error-code, database, tracking, alerting, monitoring

err := ErrDatabase.WithCode("DB_POOL_EXHAUSTED").
    WithCategory(CategoryServer).
    WithHTTPStatus(503)

// Monitoring system can alert based on code
if GetCode(err) == "DB_POOL_EXHAUSTED" {
    alert.Send("dba-team", "Database pool exhausted", err)
}
```

**Common Code Prefixes**:
```
DB_*     → Database errors (DB_CONN_001, DB_QUERY_TIMEOUT_001)
API_*    → External API errors (API_TIMEOUT_001, API_AUTH_FAILED_001)
VAL_*    → Validation errors (VAL_EMAIL_INVALID_001, VAL_REQUIRED_FIELD_001)
AUTH_*   → Authentication errors (AUTH_TOKEN_EXPIRED_001, AUTH_INVALID_CREDENTIALS_001)
```

**Retrieval**: `GetCode(err) string`

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

<!-- RAG: Mark error as retryable or non-retryable for automated retry logic -->

**Purpose**: Mark whether error should be retried.

**Why this matters**:
- **Prevents Infinite Loops**: Marking validation errors as non-retryable prevents retry storms
- **Improves Resilience**: Transient errors (network, timeout) marked retryable enable automatic recovery
- **Saves Resources**: Non-retryable errors fail fast instead of wasting retries
- **AI Automation**: Agents can implement retry logic without hardcoding error types
- **Circuit Breaker Integration**: Retryable flag helps circuit breakers decide when to open

**Parameters**:
- `retryable bool` - true if error is transient and can be retried

**Returns**: errific error with retry flag

**When to set `true` (retryable)**:
- ✅ Network timeouts (connection timeout, read timeout)
- ✅ Rate limits (HTTP 429, with retry-after)
- ✅ Temporary service unavailability (HTTP 503)
- ✅ Connection pool exhausted (will free up)
- ✅ Deadlock detected (might succeed on retry)
- ✅ Transient database errors (connection refused)

**When to set `false` (non-retryable)**:
- ❌ Validation failures (won't fix themselves)
- ❌ Authentication errors (need new credentials)
- ❌ Authorization failures (permissions won't change)
- ❌ Resource not found (404)
- ❌ Malformed requests (400)
- ❌ Business logic violations

**Example - Retryable Network Error**:
```go
// Use Case: Network timeout should be retried
// Keywords: retry, network, timeout, transient

err := ErrTimeout.New(netErr).
    WithRetryable(true).              // ✅ Network errors are transient
    WithRetryAfter(5 * time.Second).  // Wait 5s for network to recover
    WithMaxRetries(3).                // Try up to 3 times
    WithCategory(CategoryNetwork)

// AI Agent Usage:
if IsRetryable(err) {
    delay := GetRetryAfter(err)
    if delay == 0 {
        delay = time.Second  // Default delay
    }
    time.Sleep(delay)
    return retry()
}
```

**Example - Non-Retryable Validation Error**:
```go
// Use Case: Validation errors should NOT be retried
// Keywords: validation, non-retryable, user-error

err := ErrInvalidEmail.New().
    WithRetryable(false).              // ❌ User input won't fix itself
    WithCategory(CategoryValidation).
    WithHTTPStatus(400).
    WithContext(Context{
        "field": "email",
        "value": "invalid-email",      // Missing @ sign
        "constraint": "must contain @",
    })

// AI Agent Usage:
if !IsRetryable(err) {
    // Don't retry, return error to user immediately
    w.WriteHeader(GetHTTPStatus(err))
    json.NewEncoder(w).Encode(err)
    return
}
```

**Example - Rate Limit (Retryable with Delay)**:
```go
// Use Case: Rate limit should be retried after delay
// Keywords: rate-limit, retry-after, backoff

err := ErrRateLimit.New().
    WithRetryable(true).                // ✅ Temporary, will reset
    WithRetryAfter(30 * time.Second).   // Wait for rate limit window
    WithMaxRetries(1).                  // Only retry once (avoid ban)
    WithHTTPStatus(429).
    WithContext(Context{
        "limit": "100/hour",
        "reset_at": time.Now().Add(30*time.Minute).Unix(),
    })
```

**Decision Tree**:
```
Is error retryable?
├─ User input error? → NO (validation won't fix itself)
├─ Auth/permission error? → NO (credentials won't change)
├─ Not found error? → NO (resource won't appear)
├─ Network error? → YES (network might recover)
├─ Timeout error? → YES (might succeed with more time)
├─ Rate limit? → YES (will reset after window)
├─ Resource exhausted? → YES (resources will free up)
└─ Server error? → MAYBE (depends on cause)
```

**Common Mistakes**:
```go
// ❌ DON'T: Retry validation errors (infinite loop)
err := ErrInvalidInput.New().WithRetryable(true)  // WRONG!

// ✅ DO: Mark validation as non-retryable
err := ErrInvalidInput.New().WithRetryable(false)

// ❌ DON'T: Retry without delay (spam)
err := ErrNetwork.New().WithRetryable(true)  // Missing RetryAfter!

// ✅ DO: Include retry delay
err := ErrNetwork.New().
    WithRetryable(true).
    WithRetryAfter(5 * time.Second)
```

**Retrieval**: `IsRetryable(err) bool`

---

### `.WithRetryAfter(duration time.Duration) errific`

<!-- RAG: Set delay duration before retry attempt for rate limits and backoff -->

**Purpose**: Suggest delay before retry attempt.

**Why this matters**:
- **Respects Rate Limits**: Honor server's Retry-After header to avoid bans
- **Prevents Retry Storms**: Delays prevent overwhelming recovering services
- **Improves Success Rate**: Waiting gives transient issues time to resolve
- **Circuit Breaker Integration**: Delays help circuit breakers stay open long enough
- **Resource Management**: Prevents wasting resources on immediate re-failure

**Parameters**:
- `duration time.Duration` - Time to wait before retry
- Negative values are normalized to 0

**Returns**: errific error with retry delay

**When to use**:
- ✅ Rate limit errors (use Retry-After header value)
- ✅ Network timeouts (give network time to recover)
- ✅ Service unavailable (wait for service to restart)
- ✅ Resource exhaustion (wait for resources to free up)
- ✅ Any retryable error with WithRetryable(true)

**When NOT to use**:
- ❌ Non-retryable errors (validation, not found, auth)
- ❌ When using external backoff library (conflicts)

**Common Values**:
- `1 * time.Second` - Fast retry for transient issues
- `5 * time.Second` - Standard retry delay
- `30 * time.Second` - Rate limit backoff
- `5 * time.Minute` - Long delay for maintenance

**Example - Rate Limit with Retry-After Header**:
```go
// Use Case: Respect server's rate limit window
// Keywords: rate-limit, retry-after, http-429, backoff

// Parse Retry-After header from HTTP 429 response
retryAfterSec := resp.Header.Get("Retry-After")
retryAfter, _ := time.ParseDuration(retryAfterSec + "s")

err := ErrRateLimit.New().
    WithRetryable(true).
    WithRetryAfter(retryAfter).  // Use server's suggested delay
    WithMaxRetries(1).            // Only retry once (avoid ban)
    WithHTTPStatus(429)

// AI Agent automatically waits
if IsRetryable(err) {
    delay := GetRetryAfter(err)
    log.Info("Rate limited, waiting", "delay", delay)
    time.Sleep(delay)
    return retry()
}
```

**Example - Exponential Backoff**:
```go
// Use Case: Retry with increasing delays for transient failures
// Keywords: exponential-backoff, retry-strategy, resilience

func retryWithBackoff(operation func() error) error {
    baseDelay := 2 * time.Second

    for attempt := 0; attempt < 5; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        if !IsRetryable(err) {
            return err
        }

        // Exponential backoff: 2s, 4s, 8s, 16s, 32s
        delay := GetRetryAfter(err)
        if delay == 0 {
            delay = baseDelay * time.Duration(1<<attempt)
        }

        log.Info("Retrying with backoff",
            "attempt", attempt+1,
            "delay", delay)
        time.Sleep(delay)
    }
    return err
}
```

**Example - Service Unavailable**:
```go
// Use Case: Wait for service to recover from deployment
// Keywords: service-unavailable, deployment, recovery

err := ErrServiceUnavailable.New().
    WithRetryable(true).
    WithRetryAfter(10 * time.Second).  // Wait for service restart
    WithMaxRetries(6).                 // Try for 1 minute total
    WithHTTPStatus(503)

// Output: Will retry after 10s, up to 6 times
```

**Common Mistakes**:
```go
// ❌ DON'T: Set retryable without delay (immediate retry)
err := ErrNetwork.New().WithRetryable(true)  // Missing RetryAfter!

// ✅ DO: Always include delay with retryable errors
err := ErrNetwork.New().
    WithRetryable(true).
    WithRetryAfter(5 * time.Second)

// ❌ DON'T: Use negative duration (confusing)
err := ErrTimeout.New().WithRetryAfter(-5 * time.Second)  // Normalized to 0

// ✅ DO: Use positive duration
err := ErrTimeout.New().WithRetryAfter(5 * time.Second)
```

**Retrieval**: `GetRetryAfter(err) time.Duration`

---

### `.WithMaxRetries(max int) errific`

<!-- RAG: Set maximum retry attempts to prevent infinite loops and resource waste -->

**Purpose**: Set maximum retry attempts to prevent infinite loops.

**Why this matters**:
- **Prevents Infinite Loops**: Cap retries so errors eventually fail instead of retry forever
- **Resource Protection**: Limit wasted compute/network resources on failing operations
- **Fast Failure Detection**: Fail after reasonable attempts instead of hanging indefinitely
- **Cost Control**: Prevent runaway API costs from excessive retry attempts
- **User Experience**: Bound retry time so users don't wait forever

**Parameters**:
- `max int` - Maximum number of retry attempts (0 = no retries)
- Negative values are normalized to 0

**Returns**: errific error with max retries

**When to use**:
- ✅ All retryable errors (always set a limit)
- ✅ Critical operations (higher limit = 5)
- ✅ Standard operations (3 retries is typical)
- ✅ Expensive operations (1 retry to limit cost)
- ✅ Background jobs (higher limit = 10)

**When NOT to use**:
- ❌ Non-retryable errors (set WithRetryable(false) instead)
- ❌ Health checks (use 0, need immediate result)

**Recommended Values**:
- `0` - No retries (health checks, immediate failure needed)
- `1` - Single retry (expensive operations, non-idempotent)
- `3` - Standard retry limit (most operations)
- `5` - Aggressive retry (critical operations, idempotent)
- `10` - Background jobs (can wait, eventual consistency)

**Example - Standard API Call**:
```go
// Use Case: Retry API call with reasonable limit
// Keywords: api-retry, retry-limit, standard-operation

err := ErrAPI.New(httpErr).
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithMaxRetries(3).  // Try 3 times max
    WithHTTPStatus(504)

// AI Agent implements retry loop with limit
for attempt := 0; attempt < GetMaxRetries(err); attempt++ {
    err := callAPI()
    if err == nil {
        return nil  // Success!
    }
    if !IsRetryable(err) || attempt >= GetMaxRetries(err)-1 {
        return err  // Failed or max retries reached
    }
    time.Sleep(GetRetryAfter(err))
}
```

**Example - Critical Operation with Higher Limit**:
```go
// Use Case: Payment processing must succeed if possible
// Keywords: critical-operation, payment, high-retry-limit

err := ErrPaymentGateway.New(gatewayErr).
    WithRetryable(true).
    WithRetryAfter(10 * time.Second).
    WithMaxRetries(5).  // Higher limit for critical operation
    WithContext(Context{
        "transaction_id": txID,
        "amount": amount,
        "idempotency_key": idempotencyKey,  // Safe to retry
    })

// Output: Will retry up to 5 times with 10s delay
```

**Example - Expensive Operation with Low Limit**:
```go
// Use Case: ML model inference is expensive, limit retries
// Keywords: expensive-operation, ml-inference, cost-control

err := ErrMLInference.New(inferenceErr).
    WithRetryable(true).
    WithRetryAfter(2 * time.Second).
    WithMaxRetries(1).  // Only retry once (expensive)
    WithContext(Context{
        "model": "gpt-4",
        "tokens": 5000,
        "cost_usd": 0.50,
    })

// Output: Will only retry once to limit costs
```

**Example - Background Job with High Limit**:
```go
// Use Case: Background job can retry many times
// Keywords: background-job, eventual-consistency, high-retry

err := ErrEmailSend.New(smtpErr).
    WithRetryable(true).
    WithRetryAfter(30 * time.Second).
    WithMaxRetries(10).  // Background job, can wait
    WithContext(Context{
        "email": recipient,
        "template": "welcome",
    })

// Output: Will retry up to 10 times over 5 minutes
```

**Common Mistakes**:
```go
// ❌ DON'T: Set retryable without max retries (unbounded)
err := ErrNetwork.New().WithRetryable(true)  // Missing MaxRetries!

// ✅ DO: Always set a limit
err := ErrNetwork.New().
    WithRetryable(true).
    WithMaxRetries(3)

// ❌ DON'T: Set max retries too high for user-facing ops
err := ErrAPICall.New().WithMaxRetries(100)  // User waits forever!

// ✅ DO: Use reasonable limits for user-facing operations
err := ErrAPICall.New().WithMaxRetries(2)  // 2 retries = ~15s max

// ❌ DON'T: Set max retries on non-retryable errors (confusing)
err := ErrValidation.New().
    WithRetryable(false).
    WithMaxRetries(3)  // Ignored, but confusing

// ✅ DO: Only set max retries on retryable errors
err := ErrValidation.New().WithRetryable(false)
```

**Retrieval**: `GetMaxRetries(err) int`

---

### `.WithHTTPStatus(status int) errific`

<!-- RAG: Map error to HTTP status code for automatic API response handling -->

**Purpose**: Map error to HTTP status code for API responses.

**Why this matters**:
- **Automatic Response Mapping**: Errors carry their own HTTP status for consistent API responses
- **Client Understanding**: Proper status codes help clients handle errors correctly
- **API Compliance**: Meet HTTP specification and RESTful API conventions
- **Error Categorization**: Status codes group errors (4xx = client, 5xx = server)
- **Monitoring**: Track API errors by status code in dashboards
- **Retry Decisions**: Clients know which errors to retry based on 5xx vs 4xx

**Parameters**:
- `status int` - HTTP status code (100-599, 0 = not set)
- Panics if code is outside valid range (except 0)

**Returns**: errific error with HTTP status

**When to use**:
- ✅ All errors in HTTP/REST APIs
- ✅ All errors in web services
- ✅ Any error that might be returned to HTTP client
- ✅ When building API middleware or handlers

**When NOT to use**:
- ❌ Internal errors never exposed via HTTP
- ❌ CLI applications (no HTTP involved)
- ❌ Background jobs not triggered by HTTP

**Common Mappings**:
```
4xx - Client Errors (don't retry):
400 - Validation errors, bad request
401 - Authentication required
403 - Permission denied, forbidden
404 - Resource not found
408 - Client request timeout
409 - Conflict (duplicate, version mismatch)
422 - Unprocessable entity (semantic validation)
429 - Rate limit exceeded (retry with delay)

5xx - Server Errors (retry possible):
500 - Internal server error
502 - Bad gateway (upstream failure)
503 - Service unavailable (temporary)
504 - Gateway timeout (upstream timeout)
```

**Example - Validation Error**:
```go
// Use Case: Return 400 for invalid user input
// Keywords: validation, bad-request, api-error, http-400

err := ErrValidation.New().
    WithHTTPStatus(400).
    WithCategory(CategoryValidation).
    WithCode("VAL_EMAIL_INVALID").
    WithContext(Context{
        "field": "email",
        "value": "invalid-email",
    })

// Automatic HTTP response
w.WriteHeader(GetHTTPStatus(err))  // 400
json.NewEncoder(w).Encode(err)
```

**Example - Not Found**:
```go
// Use Case: Return 404 when resource doesn't exist
// Keywords: not-found, http-404, resource-missing

err := ErrUserNotFound.New().
    WithHTTPStatus(404).
    WithCategory(CategoryNotFound).
    WithContext(Context{
        "user_id": userID,
    })

// Client knows resource doesn't exist, won't retry
```

**Example - Rate Limit**:
```go
// Use Case: Return 429 with Retry-After header
// Keywords: rate-limit, http-429, retry-after

retryAfter := 30 * time.Second

err := ErrRateLimit.New().
    WithHTTPStatus(429).
    WithRetryable(true).
    WithRetryAfter(retryAfter).
    WithContext(Context{
        "limit": "100/hour",
        "reset_at": time.Now().Add(retryAfter).Unix(),
    })

// Set both status and Retry-After header
w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
w.WriteHeader(GetHTTPStatus(err))  // 429
json.NewEncoder(w).Encode(err)
```

**Example - Server Error with Retry**:
```go
// Use Case: Database failure, return 503 so clients retry
// Keywords: service-unavailable, http-503, server-error

err := ErrDatabaseDown.New(dbErr).
    WithHTTPStatus(503).
    WithCategory(CategoryServer).
    WithRetryable(true).
    WithRetryAfter(10 * time.Second).
    WithCode("DB_UNAVAILABLE")

// Clients see 503 and know to retry
```

**Example - Automatic Status from Category**:
```go
// Use Case: Fallback to category-based status if not set explicitly
// Keywords: automatic-mapping, category-to-status

err := ErrValidation.New().
    WithCategory(CategoryValidation)
    // No WithHTTPStatus() call

// Middleware can use category as fallback
status := GetHTTPStatus(err)
if status == 0 {
    switch GetCategory(err) {
    case CategoryValidation:
        status = 400
    case CategoryNotFound:
        status = 404
    case CategoryUnauthorized:
        status = 401
    case CategoryServer:
        status = 500
    default:
        status = 500
    }
}
w.WriteHeader(status)
```

**Common Mistakes**:
```go
// ❌ DON'T: Use 200 for errors (confusing)
err := ErrFailed.New().WithHTTPStatus(200)  // WRONG!

// ✅ DO: Use appropriate error status (4xx or 5xx)
err := ErrFailed.New().WithHTTPStatus(500)

// ❌ DON'T: Use invalid status codes
err := ErrTest.New().WithHTTPStatus(999)  // Panics!

// ✅ DO: Use valid HTTP status codes (100-599)
err := ErrTest.New().WithHTTPStatus(400)

// ❌ DON'T: Use 5xx for client errors
err := ErrInvalidInput.New().WithHTTPStatus(500)  // Wrong category!

// ✅ DO: Use 4xx for client errors, 5xx for server errors
err := ErrInvalidInput.New().WithHTTPStatus(400)
```

**Retrieval**: `GetHTTPStatus(err) int`

---

## Phase 2A Methods (MCP, Tracing & AI Guidance)

<!-- RAG: Methods for MCP integration, distributed tracing, AI guidance, and semantic tagging -->

### `.WithMCPCode(code int) errific`

<!-- RAG: Add JSON-RPC 2.0 / MCP error code for LLM tool integration -->

**Purpose**: Set MCP (Model Context Protocol) error code for LLM tool servers.

**Why this matters**:
- **LLM Integration**: Standard error codes that LLMs understand and can act on
- **JSON-RPC 2.0 Compliance**: Maps to official JSON-RPC 2.0 error codes
- **Tool Server Development**: Build MCP servers that communicate clearly with Claude and other LLMs
- **Error Classification**: LLMs can distinguish between method not found vs invalid params vs execution error
- **Automatic Recovery**: LLMs use MCP codes to decide recovery actions (retry, fix params, abort)
- **Debugging**: Trace errors through LLM → Tool Server → Backend with consistent codes

**Parameters**:
- `code int` - MCP error code (see MCP constants below)

**Returns**: errific error with MCP code attached

**Chaining**: Can be chained with other methods

**MCP Error Codes** (from JSON-RPC 2.0 spec):
```go
// Standard JSON-RPC 2.0 codes
MCPParseError     = -32700  // Invalid JSON
MCPInvalidRequest = -32600  // Invalid request structure
MCPMethodNotFound = -32601  // Method doesn't exist
MCPInvalidParams  = -32602  // Invalid method parameters
MCPInternalError  = -32603  // Internal server error

// MCP-specific codes (Server Error range: -32000 to -32099)
MCPToolError      = -32000  // Tool execution failed
MCPResourceError  = -32001  // Resource access failed
MCPTimeoutError   = -32002  // Operation timeout
MCPAuthError      = -32003  // Authentication failed
```

**When to use**:
- ✅ Building MCP tool servers for Claude or other LLMs
- ✅ When returning errors in JSON-RPC 2.0 format
- ✅ When LLMs need to distinguish error types programmatically
- ✅ For tools that need automatic retry/recovery logic

**When NOT to use**:
- ❌ In standard REST APIs (use WithHTTPStatus instead)
- ❌ In internal services not exposed to LLMs
- ❌ When you're not following JSON-RPC 2.0 protocol

**Example 1 - Method Not Found**:
```go
// Use Case: LLM requested a tool that doesn't exist
// Keywords: mcp, tool-not-found, json-rpc, llm-integration

err := ErrToolNotFound.New().
    WithMCPCode(errific.MCPMethodNotFound).  // -32601
    WithCategory(CategoryNotFound).
    WithHelp("The requested tool is not available in this server").
    WithSuggestion("Use the 'list_tools' method to see available tools").
    WithDocs("https://docs.example.com/mcp/available-tools").
    WithContext(Context{
        "requested_tool": "search_web",
        "available_tools": []string{"search_db", "send_email"},
    })

// Converts to JSON-RPC 2.0 format
mcpErr := errific.ToMCPError(err)
// {
//   "code": -32601,
//   "message": "tool not found",
//   "data": {
//     "help": "The requested tool is not available...",
//     "suggestion": "Use the 'list_tools' method...",
//     ...
//   }
// }
```

**Example 2 - Invalid Parameters**:
```go
// Use Case: LLM provided parameters that don't match tool schema
// Keywords: mcp, invalid-params, validation, schema

err := ErrInvalidToolParams.New().
    WithMCPCode(errific.MCPInvalidParams).  // -32602
    WithCategory(CategoryValidation).
    WithHTTPStatus(400).
    WithHelp("The 'query' parameter is required but was not provided").
    WithSuggestion("Include a 'query' parameter with your search term").
    WithDocs("https://docs.example.com/mcp/tools/search#parameters").
    WithContext(Context{
        "tool": "search",
        "required_params": []string{"query", "limit"},
        "provided_params": []string{"limit"},  // Missing 'query'
        "schema_url": "https://example.com/schema/search.json",
    })

// LLM reads this and fixes the request automatically
```

**Example 3 - Tool Execution Error with Retry**:
```go
// Use Case: Tool execution failed due to transient database issue
// Keywords: mcp, tool-error, retryable, database

err := ErrToolExecution.New(dbErr).
    WithMCPCode(errific.MCPToolError).       // -32000
    WithCategory(CategoryServer).
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithMaxRetries(3).
    WithHelp("Database connection pool is exhausted. This is temporary.").
    WithSuggestion("Retry in 5 seconds when connections are released").
    WithCorrelationID(traceID).
    WithContext(Context{
        "tool": "search_database",
        "database": "users_db",
        "pool_size": 100,
        "active_connections": 100,
    })

// AI Agent Decision Logic:
mcpErr := errific.ToMCPError(err)
if mcpErr.Data["retryable"] == true {
    delay := mcpErr.Data["retry_after"].(string)  // "5s"
    // LLM waits 5s and retries automatically
}
```

**Example 4 - Authentication Error (Non-Retryable)**:
```go
// Use Case: LLM provided invalid API key for tool
// Keywords: mcp, auth-error, non-retryable, security

err := ErrAuthFailed.New().
    WithMCPCode(errific.MCPAuthError).       // -32003
    WithCategory(CategoryUnauthorized).
    WithHTTPStatus(401).
    WithRetryable(false).                    // Don't retry, credentials won't change
    WithHelp("The API key provided is invalid or expired").
    WithSuggestion("Check that you're using a valid API key from your account settings").
    WithDocs("https://docs.example.com/authentication#api-keys").
    WithContext(Context{
        "auth_method": "api_key",
        "key_prefix": "sk_test_...",  // Partial key for debugging
    })

// LLM knows not to retry and should prompt user for new credentials
```

**Common Mistakes**:
```go
// ❌ DON'T: Use HTTP status codes as MCP codes
err := ErrNotFound.New().WithMCPCode(404)  // Wrong! Use MCPMethodNotFound

// ✅ DO: Use MCP constants
err := ErrNotFound.New().WithMCPCode(errific.MCPMethodNotFound)

// ❌ DON'T: Use MCP codes in REST APIs
func restHandler(w http.ResponseWriter, r *http.Request) {
    err := ErrFailed.New().WithMCPCode(errific.MCPToolError)  // Wrong context!
    // Use WithHTTPStatus for REST
}

// ✅ DO: Use MCP codes only for JSON-RPC 2.0 / MCP servers
func mcpHandler(req *MCPRequest) *MCPResponse {
    err := ErrFailed.New().WithMCPCode(errific.MCPToolError)
    return &MCPResponse{Error: errific.ToMCPError(err)}
}

// ❌ DON'T: Forget to include help and suggestions with MCP errors
err := ErrTool.New().WithMCPCode(errific.MCPToolError)  // LLM can't recover!

// ✅ DO: Include help, suggestion, and docs for LLM recovery
err := ErrTool.New().
    WithMCPCode(errific.MCPToolError).
    WithHelp("What went wrong").
    WithSuggestion("How to fix it").
    WithDocs("Where to learn more")
```

**Retrieval**: `GetMCPCode(err) int`

**See Also**:
- `WithHelp()` - Add human-readable help message for LLMs
- `WithSuggestion()` - Add actionable recovery suggestion
- `WithDocs()` - Add documentation URL
- `ToMCPError()` - Convert errific error to MCP format

---

### `.WithCorrelationID(id string) errific`

<!-- RAG: Add distributed tracing correlation ID for tracking errors across services -->

**Purpose**: Add correlation/trace ID for distributed tracing across microservices.

**Why this matters**:
- **Distributed Tracing**: Track a single request across multiple services (Gateway → Auth → Database → Cache)
- **Log Aggregation**: Group all logs from one user request using the correlation ID
- **Root Cause Analysis**: Trace error back through entire service chain to find origin
- **OpenTelemetry/Datadog Integration**: Correlation ID links errific errors to traces in APM tools
- **Customer Support**: Search all logs for specific customer interaction using their correlation ID
- **Performance Analysis**: Measure total latency of request across all services

**Parameters**:
- `id string` - Correlation/trace ID (often from OpenTelemetry, request headers, or generated UUID)

**Returns**: errific error with correlation ID attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Microservices architecture (track requests across services)
- ✅ When using OpenTelemetry, Datadog, or other tracing systems
- ✅ Multi-step workflows (payment processing, order fulfillment)
- ✅ Debugging distributed systems
- ✅ Customer support investigations

**When NOT to use**:
- ❌ Single monolithic application with no tracing
- ❌ When you already use WithRequestID (don't duplicate)
- ❌ For internal function-level errors (too granular)

**Example 1 - Microservices Request Chain**:
```go
// Use Case: Track error through Gateway → User Service → Database
// Keywords: microservices, distributed-tracing, correlation-id, opentelemetry

// Gateway receives request with trace ID
correlationID := r.Header.Get("X-Correlation-ID")
if correlationID == "" {
    correlationID = uuid.New().String()
}

// Gateway calls User Service
userResp, err := userService.GetUser(ctx, userID)
if err != nil {
    return ErrUserServiceFailed.New(err).
        WithCorrelationID(correlationID).  // Pass through chain
        WithRequestID(r.Header.Get("X-Request-ID")).
        WithHTTPStatus(503).
        WithRetryable(true).
        WithContext(Context{
            "service": "user-service",
            "user_id": userID,
            "gateway_host": "api-gw-01",
        })
}

// User Service propagates to Database
dbErr := ErrDatabaseQuery.New(sqlErr).
    WithCorrelationID(correlationID).  // Same ID through entire chain
    WithContext(Context{
        "query": sql,
        "service": "database",
    })

// Later: Search logs for correlationID to see full request path
// Gateway (200ms) → User Service (150ms) → Database ERROR
```

**Example 2 - OpenTelemetry Integration**:
```go
// Use Case: Link errific errors to OpenTelemetry traces
// Keywords: opentelemetry, tracing, spans, distributed-tracing

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func processOrder(ctx context.Context, orderID string) error {
    // Extract trace ID from OpenTelemetry context
    span := trace.SpanFromContext(ctx)
    traceID := span.SpanContext().TraceID().String()

    // Create payment
    err := paymentService.Charge(ctx, orderID)
    if err != nil {
        return ErrPaymentFailed.New(err).
            WithCorrelationID(traceID).  // Link to OTel trace
            WithContext(Context{
                "order_id": orderID,
                "span_id": span.SpanContext().SpanID().String(),
                "trace_id": traceID,
            })
    }

    // Now you can:
    // 1. Find error in errific logs using correlation_id
    // 2. Search Datadog/Jaeger for same trace_id
    // 3. See full distributed trace with error context
    return nil
}
```

**Example 3 - Customer Support Investigation**:
```go
// Use Case: Customer reports error, support needs to find all related logs
// Keywords: customer-support, debugging, log-aggregation

// Customer sees error ID: "corr-abc-123-def"
// Support engineer searches logs

err := ErrCheckoutFailed.New().
    WithCorrelationID("corr-abc-123-def").  // Customer's session trace ID
    WithUserID("user-789").
    WithSessionID("sess-456").
    WithContext(Context{
        "cart_total": 299.99,
        "payment_method": "credit_card",
        "shipping_address": "CA, USA",
    })

// Support can now:
// 1. Search logs for "corr-abc-123-def"
// 2. See all services involved (cart, payment, shipping, email)
// 3. Find exact failure point (payment gateway timeout at 14:23:45)
// 4. Trace request timeline across all microservices
```

**Example 4 - Multi-Step Workflow**:
```go
// Use Case: Track long-running workflow (order processing)
// Keywords: workflow, saga-pattern, compensation, long-running

correlationID := uuid.New().String()

// Step 1: Reserve inventory
if err := inventoryService.Reserve(ctx, items); err != nil {
    return ErrInventoryReservation.New(err).
        WithCorrelationID(correlationID).
        WithContext(Context{
            "step": "reserve_inventory",
            "workflow": "order_processing",
        })
}

// Step 2: Process payment
if err := paymentService.Charge(ctx, total); err != nil {
    // Compensate: unreserve inventory
    inventoryService.Unreserve(ctx, items)

    return ErrPaymentProcessing.New(err).
        WithCorrelationID(correlationID).  // Same ID for entire saga
        WithContext(Context{
            "step": "process_payment",
            "workflow": "order_processing",
            "compensation": "unreserved_inventory",
        })
}

// Step 3: Schedule shipping
if err := shippingService.Schedule(ctx, address); err != nil {
    // Compensate: refund payment, unreserve inventory
    return ErrShippingSchedule.New(err).
        WithCorrelationID(correlationID).  // Track compensation actions
        WithContext(Context{
            "step": "schedule_shipping",
            "workflow": "order_processing",
        })
}

// All errors in this saga share correlationID for debugging
```

**Common Mistakes**:
```go
// ❌ DON'T: Generate new correlation ID at each service
err1 := ErrGateway.New().WithCorrelationID(uuid.New().String())
err2 := ErrService.New().WithCorrelationID(uuid.New().String())  // Can't trace!

// ✅ DO: Propagate same correlation ID through entire request chain
correlationID := extractOrGenerateTraceID(r)
err1 := ErrGateway.New().WithCorrelationID(correlationID)
err2 := ErrService.New().WithCorrelationID(correlationID)  // Same ID

// ❌ DON'T: Use correlation ID for non-distributed systems
// (Simple monolith doesn't need correlation ID)
err := ErrSimple.New().WithCorrelationID(uuid.New().String())  // Overkill

// ✅ DO: Use correlation ID only for distributed/multi-service systems
// (In microservices)
err := ErrService.New().WithCorrelationID(traceID)
```

**Retrieval**: `GetCorrelationID(err) string`

**See Also**:
- `WithRequestID()` - For individual HTTP request tracking
- `WithSessionID()` - For user session tracking
- `WithUserID()` - For user identification

---

### `.WithRequestID(id string) errific`

<!-- RAG: Add unique request ID for tracking individual HTTP requests or API calls -->

**Purpose**: Add unique request ID for tracking individual HTTP requests.

**Why this matters**:
- **Request Tracing**: Track single HTTP request from receipt to response
- **API Debugging**: Find all logs for specific API call using request ID
- **Load Balancer Correlation**: Match errors to load balancer logs using request ID
- **Rate Limiting**: Track request patterns and identify abusive clients
- **Idempotency**: Ensure duplicate requests with same ID are handled consistently
- **API Gateway Integration**: Request IDs from Kong, Nginx, AWS API Gateway automatically included

**Parameters**:
- `id string` - Unique request identifier (often from X-Request-ID header or generated UUID)

**Returns**: errific error with request ID attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ HTTP/REST API servers
- ✅ When using API gateways (Kong, Nginx, AWS ALB)
- ✅ For request-response debugging
- ✅ When implementing idempotency
- ✅ For rate limiting and abuse detection

**When NOT to use**:
- ❌ Background jobs (use job ID in context instead)
- ❌ WebSocket connections (use connection ID)
- ❌ Batch processing (use batch ID)

**Example 1 - HTTP Request Tracking**:
```go
// Use Case: Track HTTP request through middleware chain
// Keywords: http, request-id, middleware, api-gateway

func requestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get or generate request ID
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        // Add to response headers for client correlation
        w.Header().Set("X-Request-ID", requestID)

        // Store in context for downstream use
        ctx := context.WithValue(r.Context(), "request_id", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
    requestID := r.Context().Value("request_id").(string)

    // Business logic
    user, err := getUserFromDB(r.Context(), userID)
    if err != nil {
        apiErr := ErrUserNotFound.New(err).
            WithRequestID(requestID).  // Track this specific request
            WithHTTPStatus(404).
            WithContext(Context{
                "user_id": userID,
                "endpoint": r.URL.Path,
                "method": r.Method,
            })

        // Client can use request ID to report issues
        http.Error(w, apiErr.Error(), 404)
        return
    }
}

// Client sees: "user not found [X-Request-ID: req-abc-123]"
// Support searches logs: grep "req-abc-123" and finds entire request path
```

**Example 2 - Idempotent Payment Processing**:
```go
// Use Case: Prevent duplicate payment charges using request ID
// Keywords: idempotency, payment, duplicate-prevention, request-id

func processPayment(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("Idempotency-Key")  // Client provides
    if requestID == "" {
        http.Error(w, "Idempotency-Key required", 400)
        return
    }

    // Check if this request was already processed
    if result, found := idempotencyCache.Get(requestID); found {
        // Return same result (already charged)
        w.Write(result)
        return
    }

    // Process payment
    err := paymentGateway.Charge(amount)
    if err != nil {
        paymentErr := ErrPaymentFailed.New(err).
            WithRequestID(requestID).  // Track idempotency key
            WithRetryable(false).       // Don't auto-retry (already have idempotency)
            WithContext(Context{
                "amount": amount,
                "currency": "USD",
                "idempotency_key": requestID,
                "duplicate_check": "passed",
            })

        // Store error in cache to return same error if client retries
        idempotencyCache.Set(requestID, paymentErr)
        http.Error(w, paymentErr.Error(), 500)
        return
    }

    // Cache successful result
    idempotencyCache.Set(requestID, result)
}
```

**Example 3 - Load Balancer Log Correlation**:
```go
// Use Case: Correlate application errors with load balancer logs
// Keywords: load-balancer, aws-alb, nginx, logging

// AWS ALB adds X-Amzn-Trace-Id header
// Nginx adds X-Request-ID header

func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Extract request ID from various sources
    requestID := r.Header.Get("X-Request-ID")           // Nginx
    if requestID == "" {
        requestID = r.Header.Get("X-Amzn-Trace-Id")    // AWS ALB
    }
    if requestID == "" {
        requestID = uuid.New().String()                 // Generate
    }

    err := processBusinessLogic(r.Context())
    if err != nil {
        appErr := ErrProcessing.New(err).
            WithRequestID(requestID).  // Match load balancer logs
            WithHTTPStatus(500).
            WithContext(Context{
                "client_ip": r.RemoteAddr,
                "user_agent": r.UserAgent(),
                "load_balancer": "alb-prod-01",
            })

        // Operations team can:
        // 1. Find error in app logs: grep "requestID"
        // 2. Find in ALB logs: grep "X-Amzn-Trace-Id"
        // 3. See client IP, timing, routing info from ALB
        // 4. Correlate with WAF logs if security issue

        http.Error(w, appErr.Error(), 500)
        return
    }
}
```

**Common Mistakes**:
```go
// ❌ DON'T: Use request ID for background jobs
func backgroundJob() {
    err := ErrJob.New().WithRequestID(uuid.New().String())  // Wrong context!
}

// ✅ DO: Use context for background jobs
func backgroundJob() {
    jobID := uuid.New().String()
    err := ErrJob.New().WithContext(Context{"job_id": jobID})
}

// ❌ DON'T: Generate new request ID instead of using existing one
requestID := uuid.New().String()  // Ignores X-Request-ID from gateway!

// ✅ DO: Extract from headers, generate only if missing
requestID := r.Header.Get("X-Request-ID")
if requestID == "" {
    requestID = uuid.New().String()
}
```

**Retrieval**: `GetRequestID(err) string`

**See Also**:
- `WithCorrelationID()` - For distributed tracing across services
- `WithSessionID()` - For user session tracking

---

### `.WithUserID(id string) errific`

<!-- RAG: Add user identifier for tracking errors by user -->

**Purpose**: Add user ID to track which user encountered the error.

**Why this matters**:
- **User-Specific Debugging**: Find all errors for a specific user when they report issues
- **Support Tickets**: Quickly search logs by user ID when customer contacts support
- **Abuse Detection**: Identify users with abnormal error rates (bots, attackers)
- **Feature Rollout**: Track errors during A/B testing or gradual feature rollouts
- **Compliance**: Required for audit trails (GDPR, HIPAA, SOC 2)
- **User Impact Analysis**: Determine how many users are affected by a bug

**Parameters**:
- `id string` - User identifier (user ID, email, username, or external ID)

**Returns**: errific error with user ID attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Any error in user-facing features
- ✅ Authentication/authorization errors
- ✅ When user reports a bug
- ✅ For audit logging
- ✅ During A/B testing or feature flags

**When NOT to use**:
- ❌ System/background errors not tied to a user
- ❌ When user is not authenticated (use session ID instead)
- ❌ For sensitive PII (hash or pseudonymize first)

**Example 1 - User Support Investigation**:
```go
// Use Case: User reports checkout error, support needs to debug
// Keywords: user-support, debugging, customer-service

func processCheckout(w http.ResponseWriter, r *http.Request) {
    userID := GetAuthenticatedUserID(r)  // From JWT/session

    err := paymentService.Charge(total)
    if err != nil {
        checkoutErr := ErrCheckoutFailed.New(err).
            WithUserID(userID).  // Support can search logs by user_id
            WithRequestID(requestID).
            WithContext(Context{
                "cart_total": total,
                "payment_method": paymentMethod,
                "cart_items": len(items),
            })

        // User reports: "Checkout failed at 2pm"
        // Support searches: grep "user_id: user-123" logs.json
        // Finds: All checkout attempts, cart data, payment gateway responses

        http.Error(w, checkoutErr.Error(), 500)
        return
    }
}
```

**Example 2 - Abuse Detection**:
```go
// Use Case: Detect users generating excessive errors (bots, scrapers)
// Keywords: abuse-detection, rate-limiting, security, bot-detection

func handleAPIRequest(w http.ResponseWriter, r *http.Request) {
    userID := GetUserIDFromAPIKey(r)

    // Track error rate per user
    err := processRequest(r)
    if err != nil {
        apiErr := ErrAPIRequest.New(err).
            WithUserID(userID).
            WithHTTPStatus(429).
            WithContext(Context{
                "endpoint": r.URL.Path,
                "api_key_prefix": apiKey[:8],
            })

        // Monitor: Count errors per user_id in last 5 minutes
        // If user_id "bot-456" has >1000 errors → Block API key
        // If user_id "user-789" has 2 errors → Normal usage

        errorRateTracker.Record(userID, apiErr)
        if errorRateTracker.IsAbusing(userID) {
            // Block user
            return
        }

        http.Error(w, apiErr.Error(), 429)
        return
    }
}
```

**Example 3 - A/B Testing Impact Analysis**:
```go
// Use Case: New feature causes errors for some users, track impact
// Keywords: ab-testing, feature-flags, gradual-rollout, impact-analysis

func handleFeatureRequest(w http.ResponseWriter, r *http.Request) {
    userID := GetUserID(r)

    // Feature flag: 10% of users get new algorithm
    if featureFlags.IsEnabled("new-search-v2", userID) {
        err := newSearchAlgorithmV2(query)
        if err != nil {
            return ErrSearch.New(err).
                WithUserID(userID).  // Track which users hit errors
                WithContext(Context{
                    "feature_flag": "new-search-v2",
                    "rollout_percentage": 10,
                    "algorithm_version": "v2",
                    "query": query,
                })

            // Analysis: Search for user_ids in "new-search-v2" cohort
            // Result: 50 unique user_ids affected = 50 users hit bug
            // Decision: Roll back feature (too many errors)
        }
    }
}
```

**Example 4 - GDPR Compliance Audit Trail**:
```go
// Use Case: Track data access and errors for compliance audit
// Keywords: gdpr, compliance, audit-trail, data-privacy

func exportUserData(w http.ResponseWriter, r *http.Request) {
    userID := GetUserID(r)
    adminID := GetAdminID(r)  // Who initiated export

    // User data export (GDPR Right to Data Portability)
    data, err := database.ExportUserData(userID)
    if err != nil {
        exportErr := ErrDataExport.New(err).
            WithUserID(userID).  // Which user's data
            WithContext(Context{
                "admin_id": adminID,          // Who tried to export
                "export_type": "gdpr_request",
                "data_size_mb": 0,            // Failed, no data
                "ip_address": r.RemoteAddr,   // Where request came from
                "timestamp": time.Now().Unix(),
            })

        // Audit log must track:
        // - Which user's data was accessed/failed (user_id)
        // - Who accessed it (admin_id)
        // - When (timestamp)
        // - Result (error or success)

        auditLog.Record(exportErr)
        http.Error(w, "Export failed", 500)
        return
    }
}
```

**Common Mistakes**:
```go
// ❌ DON'T: Store email addresses directly (PII risk)
err := ErrFailed.New().WithUserID("user@example.com")  // PII!

// ✅ DO: Use internal user ID or hash
err := ErrFailed.New().WithUserID("user-abc-123")

// ❌ DON'T: Add user ID to system errors
err := ErrDatabaseMigration.New().WithUserID(userID)  // System error, no user

// ✅ DO: Add user ID only to user-initiated operations
err := ErrUserProfile.New().WithUserID(userID)

// ❌ DON'T: Use user ID when user is not authenticated
err := ErrPublicAPI.New().WithUserID("unknown")  // Meaningless

// ✅ DO: Use session ID for unauthenticated users
err := ErrPublicAPI.New().WithSessionID(sessionID)
```

**Retrieval**: `GetUserID(err) string`

**See Also**:
- `WithSessionID()` - For unauthenticated user tracking
- `WithCorrelationID()` - For tracing requests across services

---

### `.WithSessionID(id string) errific`

<!-- RAG: Add session identifier for tracking unauthenticated users and user sessions -->

**Purpose**: Add session ID for tracking unauthenticated users and user sessions.

**Why this matters**:
- **Anonymous User Tracking**: Track errors for users who aren't logged in
- **Session Debugging**: Debug issues within a specific browsing session
- **Conversion Funnel Analysis**: Track errors through signup/checkout flows
- **Session Replay**: Link errors to session replay tools (FullStory, LogRocket)
- **Bot Detection**: Identify bot sessions vs human sessions
- **Multi-Tab Issues**: Debug errors when user has multiple tabs open

**Parameters**:
- `id string` - Session identifier (from cookie, JWT, or generated)

**Returns**: errific error with session ID attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Unauthenticated users (before login)
- ✅ Signup/registration flows
- ✅ Guest checkout processes
- ✅ Public-facing pages with errors
- ✅ When integrating with session replay tools

**When NOT to use**:
- ❌ When you already have user ID (use WithUserID instead)
- ❌ API-only services without sessions
- ❌ Background jobs

**Example 1 - Guest Checkout Debugging**:
```go
// Use Case: Track checkout errors for users who aren't logged in
// Keywords: guest-checkout, unauthenticated, e-commerce, conversion-funnel

func guestCheckout(w http.ResponseWriter, r *http.Request) {
    sessionID := GetSessionID(r)  // From cookie

    // Guest user (not logged in) tries to checkout
    err := processGuestCheckout(cart, shippingInfo)
    if err != nil {
        checkoutErr := ErrGuestCheckout.New(err).
            WithSessionID(sessionID).  // Track anonymous user's session
            WithContext(Context{
                "cart_total": cart.Total,
                "cart_items": len(cart.Items),
                "shipping_country": shippingInfo.Country,
                "user_type": "guest",
            })

        // Support can:
        // 1. Search logs by session_id
        // 2. See full checkout flow: cart → shipping → payment → ERROR
        // 3. Identify if error affects multiple sessions (systemic issue)

        http.Error(w, checkoutErr.Error(), 500)
        return
    }
}
```

**Example 2 - Signup Flow Analysis**:
```go
// Use Case: Track errors during user registration process
// Keywords: signup-flow, registration, user-onboarding, conversion

func handleSignup(w http.ResponseWriter, r *http.Request) {
    sessionID := GetSessionID(r)  // Session started when user visited landing page

    // Validate email
    if !isValidEmail(email) {
        return ErrInvalidEmail.New().
            WithSessionID(sessionID).  // Track which signup session
            WithHTTPStatus(400).
            WithRetryable(false).
            WithContext(Context{
                "step": "email_validation",
                "signup_source": "google_ads",  // How they found us
                "referrer": r.Referer(),
            })
    }

    // Create account
    err := createUserAccount(email, password)
    if err != nil {
        return ErrSignupFailed.New(err).
            WithSessionID(sessionID).  // Same session through whole flow
            WithContext(Context{
                "step": "account_creation",
                "email_domain": getDomain(email),
            })
    }

    // Analysis: Count errors by session_id in signup funnel
    // Landing page → Email form → Password form → ERROR
    // Find: 30% of sessions with .edu emails fail at password step
    // Fix: Improve password requirements UI
}
```

**Example 3 - Session Replay Integration**:
```go
// Use Case: Link errors to FullStory/LogRocket session recordings
// Keywords: session-replay, fullstory, logrocket, user-experience

func handleAction(w http.ResponseWriter, r *http.Request) {
    sessionID := GetSessionID(r)
    replayURL := GetSessionReplayURL(sessionID)  // From FullStory SDK

    err := performAction(r)
    if err != nil {
        actionErr := ErrActionFailed.New(err).
            WithSessionID(sessionID).
            WithContext(Context{
                "session_replay_url": replayURL,  // Link to video
                "user_agent": r.UserAgent(),
                "viewport_width": r.Header.Get("X-Viewport-Width"),
            })

        // Support engineer:
        // 1. Sees error in logs with session_id
        // 2. Clicks session_replay_url
        // 3. Watches video of user's exact actions leading to error
        // 4. Sees UI state, network requests, console errors

        http.Error(w, actionErr.Error(), 500)
        return
    }
}
```

**Example 4 - Multi-Tab Session Debugging**:
```go
// Use Case: User has multiple tabs open, causing conflicts
// Keywords: multi-tab, session-conflicts, race-conditions

func updateCart(w http.ResponseWriter, r *http.Request) {
    sessionID := GetSessionID(r)
    tabID := r.Header.Get("X-Tab-ID")  // Track which browser tab

    // User opened 2 tabs, both trying to modify cart simultaneously
    err := cartService.UpdateItem(sessionID, itemID, quantity)
    if err != nil {
        return ErrCartConflict.New(err).
            WithSessionID(sessionID).  // Same session
            WithContext(Context{
                "tab_id": tabID,           // Different tabs!
                "conflict_type": "concurrent_modification",
                "item_id": itemID,
            })

        // Debug logs show:
        // session_id: sess-123, tab_id: tab-A → Updated item X to qty 2
        // session_id: sess-123, tab_id: tab-B → Updated item X to qty 5 (CONFLICT!)
        // Fix: Add optimistic locking or merge tab changes
    }
}
```

**Common Mistakes**:
```go
// ❌ DON'T: Use session ID when you have user ID
func authenticatedAction(r *http.Request) error {
    sessionID := GetSessionID(r)
    userID := GetUserID(r)  // User is logged in!

    err := ErrAction.New().WithSessionID(sessionID)  // Should use userID!
}

// ✅ DO: Use user ID for authenticated users
func authenticatedAction(r *http.Request) error {
    userID := GetUserID(r)
    err := ErrAction.New().WithUserID(userID)
}

// ❌ DON'T: Generate new session ID on every request
sessionID := uuid.New().String()  // Can't track across requests!

// ✅ DO: Use persistent session ID from cookie/JWT
sessionID := GetSessionIDFromCookie(r)

// ✅ ALTERNATIVE: Include both if useful
err := ErrAction.New().
    WithUserID(userID).      // Who the user is
    WithSessionID(sessionID) // Which login session
```

**Retrieval**: `GetSessionID(err) string`

**See Also**:
- `WithUserID()` - For authenticated user tracking
- `WithRequestID()` - For individual request tracking

---

### `.WithHelp(message string) errific`

<!-- RAG: Add human-readable help message explaining what went wrong for LLMs and users -->

**Purpose**: Add human-readable explanation of what went wrong and why.

**Why this matters**:
- **LLM Understanding**: AI agents read help text to understand error context without parsing code
- **User Communication**: Clear explanation improves user experience (avoid cryptic errors)
- **Automated Recovery**: LLMs use help text to decide recovery strategy
- **Reduced Support Tickets**: Good help text answers user questions before they contact support
- **Faster Debugging**: Developers understand issue immediately without checking code
- **Self-Service**: Users can often fix issues themselves with good help text

**Parameters**:
- `message string` - Clear, user-friendly explanation of what went wrong

**Returns**: errific error with help message attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ MCP tool servers (LLMs need context)
- ✅ User-facing errors (improve UX)
- ✅ Complex failure scenarios (explain non-obvious causes)
- ✅ Resource exhaustion (explain why limit was hit)
- ✅ Configuration errors (explain what's misconfigured)

**When NOT to use**:
- ❌ When error message is already clear
- ❌ For internal errors users won't see
- ❌ When it duplicates the error message

**Example 1 - Database Connection Pool Exhausted**:
```go
// Use Case: Explain resource exhaustion to LLM/user
// Keywords: database, connection-pool, resource-exhaustion, help

err := ErrDatabaseConnection.New(dbErr).
    WithHelp("The database connection pool is full because too many requests are running simultaneously. All 100 connections are in use.").
    WithSuggestion("Wait a few seconds for connections to be released, or reduce the number of concurrent requests.").
    WithRetryable(true).
    WithRetryAfter(10 * time.Second).
    WithContext(Context{
        "pool_size": 100,
        "active_connections": 100,
        "waiting_requests": 50,
    })

// LLM reads help and understands:
// - Problem: Pool is full (not a network issue, not credentials)
// - Cause: Too many concurrent requests
// - Action: Wait for connections to free up (from suggestion)
```

**Example 2 - API Rate Limit for LLM**:
```go
// Use Case: Explain rate limit to LLM tool caller
// Keywords: rate-limit, api-quota, llm-integration, mcp

err := ErrAPIRateLimit.New().
    WithMCPCode(errific.MCPResourceError).
    WithHelp("You've exceeded the API rate limit of 1000 requests per hour. The limit resets at the top of each hour.").
    WithSuggestion("Wait 45 minutes until the rate limit resets at 3:00 PM, or upgrade to a higher tier plan for more requests.").
    WithDocs("https://docs.example.com/api/rate-limits").
    WithRetryable(true).
    WithRetryAfter(45 * time.Minute).
    WithContext(Context{
        "rate_limit": 1000,
        "requests_made": 1000,
        "reset_time": "2024-01-15T15:00:00Z",
        "current_tier": "free",
    })

// LLM Decision:
// - Reads help: "Rate limit exceeded, resets at 3:00 PM"
// - Reads suggestion: "Wait 45 minutes OR upgrade plan"
// - Decides: Inform user and wait (don't spam retries)
```

**Example 3 - Configuration Error**:
```go
// Use Case: Explain misconfiguration clearly
// Keywords: configuration, setup-error, validation

err := ErrInvalidConfig.New().
    WithHelp("The S3 bucket name 'my bucket' contains spaces, which is not allowed by AWS. Bucket names must only contain lowercase letters, numbers, and hyphens.").
    WithSuggestion("Change the bucket name to 'my-bucket' in your configuration file (config.yaml, line 23).").
    WithDocs("https://docs.aws.amazon.com/s3/bucket-naming-rules").
    WithRetryable(false).
    WithContext(Context{
        "config_file": "config.yaml",
        "config_line": 23,
        "invalid_bucket_name": "my bucket",
        "suggested_bucket_name": "my-bucket",
    })

// Developer reads help and immediately knows:
// - What's wrong: Spaces in bucket name
// - Why it's wrong: AWS doesn't allow it
// - How to fix: Replace with hyphens (from suggestion)
// - Where to fix: config.yaml line 23 (from context)
```

**Example 4 - Authentication Failure**:
```go
// Use Case: Explain auth failure without exposing security details
// Keywords: authentication, security, user-error

err := ErrAuthFailed.New().
    WithHelp("Your API key is invalid or has expired. API keys are valid for 90 days from creation.").
    WithSuggestion("Generate a new API key from your account dashboard at https://example.com/dashboard/api-keys").
    WithDocs("https://docs.example.com/authentication").
    WithHTTPStatus(401).
    WithRetryable(false).
    WithContext(Context{
        "auth_method": "api_key",
        "key_age_days": 95,  // Expired
    })

// User sees helpful error without security risk:
// - Clear problem: Key expired (not "invalid credentials")
// - Clear action: Generate new key
// - No sensitive info: Doesn't reveal valid key format or DB details
```

**Best Practices**:
```go
// ✅ DO: Explain the problem clearly
WithHelp("The payment gateway timed out after 30 seconds waiting for a response")

// ❌ DON'T: Repeat the error message
WithHelp("Payment timeout")  // Already in error message!

// ✅ DO: Provide context about WHY it failed
WithHelp("The file upload failed because the file size (50MB) exceeds the 10MB limit")

// ❌ DON'T: Be vague or generic
WithHelp("An error occurred")  // Useless!

// ✅ DO: Include relevant numbers/thresholds
WithHelp("Database connection pool is exhausted (100/100 connections active)")

// ❌ DON'T: Include technical jargon for user-facing errors
WithHelp("ECONNREFUSED on socket descriptor 42")  // Too technical for users

// ✅ DO: Explain in plain language
WithHelp("Unable to connect to the database server. The server may be down or unreachable.")
```

**Retrieval**: `GetHelp(err) string`

**See Also**:
- `WithSuggestion()` - Add actionable recovery steps
- `WithDocs()` - Add link to documentation
- `WithMCPCode()` - For LLM tool integration

---

### `.WithSuggestion(message string) errific`

<!-- RAG: Add actionable recovery suggestion for LLMs and users to fix the error -->

**Purpose**: Add actionable suggestion for how to fix or recover from the error.

**Why this matters**:
- **Automated Recovery**: LLMs read suggestions and take action automatically
- **Reduced Downtime**: Clear recovery steps enable faster resolution
- **Self-Service**: Users fix issues themselves without contacting support
- **Developer Productivity**: Immediate guidance saves debugging time
- **Reduced Support Load**: Good suggestions prevent support tickets
- **AI Decision-Making**: LLMs use suggestions to choose between retry, fix params, or abort

**Parameters**:
- `message string` - Specific, actionable steps to resolve the error

**Returns**: errific error with suggestion attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ When there's a clear recovery action
- ✅ For user-correctable errors (invalid input, config issues)
- ✅ MCP tool servers (LLMs need actionable guidance)
- ✅ When you want to guide users to self-resolution
- ✅ For common problems with known solutions

**When NOT to use**:
- ❌ When there's no recovery action (permanent failures)
- ❌ For internal errors users can't fix
- ❌ When the action isn't clear or specific

**Example 1 - Invalid API Parameters (LLM Can Fix)**:
```go
// Use Case: LLM provided wrong parameter type, guide it to fix
// Keywords: mcp, parameter-validation, llm-recovery, automated-fix

err := ErrInvalidParams.New().
    WithMCPCode(errific.MCPInvalidParams).
    WithHelp("The 'limit' parameter must be a number between 1 and 100, but you provided 'unlimited'.").
    WithSuggestion("Change the 'limit' parameter to a number like 10, 50, or 100.").
    WithDocs("https://docs.example.com/api/parameters#limit").
    WithContext(Context{
        "parameter": "limit",
        "provided_value": "unlimited",
        "expected_type": "integer",
        "valid_range": "1-100",
    })

// LLM reads suggestion and:
// 1. Understands it needs to change "unlimited" to a number
// 2. Picks a reasonable default (e.g., 50)
// 3. Retries request with {"limit": 50}
// 4. Succeeds automatically without human intervention
```

**Example 2 - Rate Limit with Specific Action**:
```go
// Use Case: Tell LLM exactly when to retry
// Keywords: rate-limit, retry-timing, specific-action

err := ErrRateLimit.New().
    WithHelp("You've made 1000 API requests in the last hour, exceeding your limit.").
    WithSuggestion("Wait 15 minutes until 3:00 PM when your rate limit resets, then retry this request.").
    WithRetryable(true).
    WithRetryAfter(15 * time.Minute).
    WithContext(Context{
        "requests_made": 1000,
        "rate_limit": 1000,
        "reset_time": time.Now().Add(15 * time.Minute).Format(time.RFC3339),
    })

// LLM reads suggestion and:
// 1. Knows to wait (not retry immediately)
// 2. Knows exact time to retry (3:00 PM)
// 3. Can inform user: "I'll retry in 15 minutes when limit resets"
```

**Example 3 - Missing Required Field**:
```go
// Use Case: Guide user to provide missing data
// Keywords: validation, missing-field, user-input

err := ErrMissingField.New().
    WithHelp("The 'email' field is required but was not provided.").
    WithSuggestion("Please provide your email address in the 'email' field.").
    WithHTTPStatus(400).
    WithRetryable(false).  // Can't retry without fixing input
    WithContext(Context{
        "missing_field": "email",
        "required_fields": []string{"email", "password", "name"},
        "provided_fields": []string{"password", "name"},
    })

// User sees:
// - Help: "email field is required"
// - Suggestion: "provide your email address"
// - Context: Shows they provided password and name but not email
// → User adds email and retries successfully
```

**Example 4 - File Too Large**:
```go
// Use Case: Guide user to compress or split file
// Keywords: file-upload, size-limit, compression

err := ErrFileTooLarge.New().
    WithHelp("The file you're uploading is 50MB, which exceeds the 10MB limit.").
    WithSuggestion("Compress the file to reduce its size below 10MB, or split it into smaller files.").
    WithHTTPStatus(413).
    WithRetryable(false).
    WithContext(Context{
        "file_size_mb": 50,
        "max_size_mb": 10,
        "file_name": "presentation.pptx",
        "suggested_action": "compress",
    })

// User reads suggestion and has options:
// 1. Compress the PowerPoint file
// 2. Split into multiple files
// 3. Knows exact limit (10MB) for reference
```

**Example 5 - Configuration Fix Location**:
```go
// Use Case: Tell developer exactly where and how to fix config
// Keywords: configuration, developer-guidance, specific-fix

err := ErrInvalidConfig.New().
    WithHelp("The database connection string is missing the port number.").
    WithSuggestion("Add the port number to the connection string in config.yaml line 15. Example: 'postgres://localhost:5432/mydb'").
    WithDocs("https://docs.example.com/configuration/database").
    WithContext(Context{
        "config_file": "config.yaml",
        "config_line": 15,
        "current_value": "postgres://localhost/mydb",
        "expected_format": "postgres://localhost:5432/mydb",
    })

// Developer knows:
// - What to fix: Add port number
// - Where to fix: config.yaml line 15
// - How to fix: Example provided ("localhost:5432")
// - Can copy/paste the example format
```

**Best Practices**:
```go
// ✅ DO: Be specific and actionable
WithSuggestion("Increase the 'max_connections' setting to 200 in postgresql.conf")

// ❌ DON'T: Be vague
WithSuggestion("Fix the database configuration")  // HOW?

// ✅ DO: Provide multiple options when applicable
WithSuggestion("Either compress the file below 10MB, split it into smaller files, or upgrade to Pro plan for 100MB uploads")

// ❌ DON'T: Suggest impossible actions
WithSuggestion("Contact the administrator")  // User may not have admin access

// ✅ DO: Include examples when helpful
WithSuggestion("Use ISO 8601 format for dates. Example: '2024-01-15T10:30:00Z'")

// ❌ DON'T: Repeat the help message
WithHelp("Rate limit exceeded")
WithSuggestion("You exceeded the rate limit")  // Duplicate!

// ✅ DO: Tell WHEN to retry if relevant
WithSuggestion("Retry in 5 seconds after connections are released")

// ✅ DO: Provide links to self-service actions
WithSuggestion("Generate a new API key at https://example.com/dashboard/api-keys")
```

**Retrieval**: `GetSuggestion(err) string`

**See Also**:
- `WithHelp()` - Explain what went wrong
- `WithDocs()` - Link to documentation
- `WithRetryable()` - Indicate if retry is appropriate

---

### `.WithDocs(url string) errific`

<!-- RAG: Add documentation URL for detailed information and troubleshooting -->

**Purpose**: Add link to documentation for detailed information about the error.

**Why this matters**:
- **Self-Service Support**: Users can read docs instead of contacting support
- **LLM Context Expansion**: AI agents can fetch and read docs to understand complex issues
- **Onboarding**: New developers learn about features through error-linked docs
- **Comprehensive Guidance**: Docs provide more detail than error messages can include
- **Updated Information**: Docs can be updated without code changes
- **SEO and Discoverability**: Error docs help users find solutions via search engines

**Parameters**:
- `url string` - URL to documentation page related to this error

**Returns**: errific error with documentation URL attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Complex features that need detailed explanation
- ✅ MCP tool servers (LLMs can fetch docs)
- ✅ API errors (link to API reference)
- ✅ Configuration errors (link to config guide)
- ✅ Rate limits and quotas (link to pricing/limits page)

**When NOT to use**:
- ❌ When no relevant documentation exists
- ❌ For internal errors (docs won't help)
- ❌ When linking to generic homepage (be specific)

**Example 1 - API Authentication Documentation**:
```go
// Use Case: Link to auth docs when API key is invalid
// Keywords: authentication, api-docs, self-service

err := ErrInvalidAPIKey.New().
    WithHelp("Your API key is invalid or has expired.").
    WithSuggestion("Generate a new API key from your dashboard at https://example.com/dashboard").
    WithDocs("https://docs.example.com/authentication/api-keys").  // Detailed auth guide
    WithHTTPStatus(401).
    WithRetryable(false).
    WithContext(Context{
        "auth_method": "api_key",
        "key_format": "sk_live_...",
    })

// User clicks docs link and finds:
// - How API keys work
// - How to generate new keys
// - Key rotation best practices
// - Security recommendations
// - Troubleshooting common issues
```

**Example 2 - MCP Tool Error with Docs**:
```go
// Use Case: LLM can fetch docs to understand tool usage
// Keywords: mcp, llm-integration, tool-documentation

err := ErrToolNotFound.New().
    WithMCPCode(errific.MCPMethodNotFound).
    WithHelp("The tool 'search_web' is not available in this server.").
    WithSuggestion("Use the 'list_tools' method to see all available tools.").
    WithDocs("https://docs.example.com/mcp/tools").  // List of all tools
    WithContext(Context{
        "requested_tool": "search_web",
        "available_tools": []string{"search_db", "send_email", "create_calendar_event"},
    })

// LLM can:
// 1. Read the help message (tool not found)
// 2. Fetch docs URL to see all available tools
// 3. Find "search_db" tool as alternative
// 4. Use search_db instead of search_web
// 5. Succeed automatically
```

**Example 3 - Rate Limit Documentation**:
```go
// Use Case: Link to rate limits and pricing page
// Keywords: rate-limit, pricing, quota, upgrade

err := ErrRateLimit.New().
    WithHelp("You've exceeded the free tier limit of 1000 requests per day.").
    WithSuggestion("Upgrade to Pro plan for 100,000 requests per day at https://example.com/pricing").
    WithDocs("https://docs.example.com/api/rate-limits").  // Detailed rate limit guide
    WithHTTPStatus(429).
    WithRetryable(true).
    WithRetryAfter(24 * time.Hour).
    WithContext(Context{
        "current_tier": "free",
        "daily_limit": 1000,
        "requests_today": 1000,
        "upgrade_url": "https://example.com/pricing",
    })

// Docs page explains:
// - Rate limits for each tier (Free, Pro, Enterprise)
// - How limits are calculated (per day, per hour, per endpoint)
// - What happens when you exceed limits
// - How to upgrade to higher tier
// - Best practices for staying within limits
```

**Example 4 - Configuration Error with Schema Docs**:
```go
// Use Case: Link to configuration schema documentation
// Keywords: configuration, schema, yaml, validation

err := ErrInvalidConfig.New().
    WithHelp("The configuration file has invalid YAML syntax at line 23.").
    WithSuggestion("Check the YAML syntax and ensure all quotes are closed.").
    WithDocs("https://docs.example.com/configuration/schema").  // Config schema reference
    WithContext(Context{
        "config_file": "config.yaml",
        "error_line": 23,
        "error_column": 15,
        "syntax_error": "unclosed string",
    })

// Docs page provides:
// - Complete configuration schema
// - Example config files
// - Description of each field
// - Validation rules
// - Common configuration mistakes
```

**Example 5 - Feature-Specific Error**:
```go
// Use Case: Link to feature documentation for complex feature
// Keywords: feature-docs, onboarding, learning

err := ErrWebhookValidation.New().
    WithHelp("The webhook signature validation failed. The signature in the X-Webhook-Signature header doesn't match the computed signature.").
    WithSuggestion("Ensure you're using the correct webhook secret from your dashboard and following the signature algorithm described in the docs.").
    WithDocs("https://docs.example.com/webhooks/signature-validation").  // Detailed webhook guide
    WithContext(Context{
        "webhook_id": "wh_123",
        "signature_header": "X-Webhook-Signature",
        "algorithm": "HMAC-SHA256",
    })

// Docs explain:
// - How webhook signatures work
// - Step-by-step signature computation
// - Code examples in multiple languages
// - Common signature validation mistakes
// - How to test webhooks locally
```

**Best Practices**:
```go
// ✅ DO: Link to specific, relevant docs
WithDocs("https://docs.example.com/api/authentication/api-keys#rotation")

// ❌ DON'T: Link to generic homepage
WithDocs("https://example.com")  // Not helpful!

// ✅ DO: Use deep links to exact section
WithDocs("https://docs.example.com/errors#rate-limit-exceeded")

// ❌ DON'T: Link to 404 pages
WithDocs("https://docs.example.com/old-page")  // Check links!

// ✅ DO: Include anchor links for long pages
WithDocs("https://docs.example.com/configuration#database-connection-pool")

// ✅ DO: Version docs links if API is versioned
WithDocs("https://docs.example.com/v2/api/errors")  // Not v1

// ✅ DO: Use stable URLs that won't break
WithDocs("https://docs.example.com/permanent/api-keys")  // Permanent path

// ❌ DON'T: Link to docs behind authentication
WithDocs("https://internal.example.com/docs")  // Users can't access!
```

**Retrieval**: `GetDocs(err) string`

**See Also**:
- `WithHelp()` - Explain the problem
- `WithSuggestion()` - Provide actionable steps
- `WithMCPCode()` - For LLM tool integration

---

### `.WithTags(tags ...string) errific`

<!-- RAG: Add semantic tags for filtering, categorizing, and RAG system retrieval -->

**Purpose**: Add semantic tags for error categorization and RAG system retrieval.

**Why this matters**:
- **RAG Optimization**: Tags improve error searchability in RAG/AI systems
- **Log Filtering**: Filter logs by tags (e.g., "show all 'payment' errors")
- **Metric Aggregation**: Group errors by tags for dashboards (count by "database" tag)
- **Alert Routing**: Route specific tagged errors to specialized teams
- **Semantic Search**: Find related errors using semantic similarity of tags
- **Multi-Dimensional Categorization**: Tags provide flexible categorization beyond single category field

**Parameters**:
- `tags ...string` - Variable number of semantic tags (e.g., "payment", "critical", "external-api")

**Returns**: errific error with tags attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ For multi-dimensional error classification
- ✅ When building RAG/AI systems that search errors
- ✅ For flexible log filtering and aggregation
- ✅ When routing errors to different teams/systems
- ✅ For semantic search across errors

**When NOT to use**:
- ❌ For single-dimension classification (use WithCategory instead)
- ❌ For structured data (use WithContext instead)
- ❌ When you need exact key-value pairs (use WithLabels instead)

**Example 1 - Multi-Dimensional Classification**:
```go
// Use Case: Tag error with multiple relevant dimensions
// Keywords: tagging, classification, filtering, rag

err := ErrPaymentFailed.New(gatewayErr).
    WithTags("payment", "external-api", "critical", "retryable", "user-facing").
    WithCategory(CategoryServer).  // Primary category
    WithHTTPStatus(503).
    WithRetryable(true).
    WithContext(Context{
        "gateway": "stripe",
        "amount": 99.99,
    })

// Now searchable by:
// - "payment" tag → Finds all payment errors
// - "external-api" tag → Finds all third-party API errors
// - "critical" tag → Finds high-priority errors
// - "retryable" tag → Finds errors that can be retried
// - Multiple tags → "payment AND critical" → Payment errors that are critical
```

**Example 2 - Team Routing**:
```go
// Use Case: Route errors to appropriate teams based on tags
// Keywords: alert-routing, team-ownership, on-call

err := ErrDatabaseQuery.New(sqlErr).
    WithTags("database", "postgresql", "performance", "backend-team").
    WithContext(Context{
        "query": sql,
        "duration_ms": 5000,  // Slow query
        "table": "orders",
    })

// Alert router reads tags:
if hasTag(err, "database") && hasTag(err, "backend-team") {
    alertSystem.NotifyTeam("backend-oncall", err)
}

// Alternative: Frontend error
err2 := ErrUIRender.New().
    WithTags("frontend", "react", "rendering", "frontend-team").
    WithContext(Context{"component": "CheckoutForm"})

if hasTag(err2, "frontend-team") {
    alertSystem.NotifyTeam("frontend-oncall", err2)
}
```

**Example 3 - RAG System Semantic Search**:
```go
// Use Case: Enable semantic search across errors for AI agents
// Keywords: rag, semantic-search, ai-retrieval, embeddings

// Error 1: Payment gateway timeout
err1 := ErrPaymentTimeout.New().
    WithTags("payment", "timeout", "stripe", "network", "checkout").
    WithHelp("Payment gateway did not respond within 30 seconds")

// Error 2: Database connection timeout
err2 := ErrDatabaseTimeout.New().
    WithTags("database", "timeout", "postgres", "network", "connection-pool").
    WithHelp("Database connection timed out after 5 seconds")

// RAG System Query: "Find all timeout errors"
// Returns: Both err1 and err2 (both have "timeout" tag)

// RAG System Query: "Find payment-related errors"
// Returns: err1 only (has "payment" tag)

// RAG System Query: "Find network issues"
// Returns: Both err1 and err2 (both have "network" tag)
```

**Example 4 - Dashboard Metrics**:
```go
// Use Case: Aggregate error counts by tags for monitoring dashboard
// Keywords: metrics, monitoring, dashboard, aggregation

// Various errors with tags
err1 := ErrAPICall.New().WithTags("api", "external", "retryable")
err2 := ErrValidation.New().WithTags("validation", "user-input", "non-retryable")
err3 := ErrDatabase.New().WithTags("database", "internal", "retryable")

// Dashboard queries:
// COUNT(errors WHERE tag = "retryable") = 2
// COUNT(errors WHERE tag = "external") = 1
// COUNT(errors WHERE tag = "user-input") = 1

// Advanced dashboard:
// - "Retryable errors by hour" (filter by "retryable" tag)
// - "External API errors" (filter by "external" tag)
// - "User-facing errors" (filter by "user-input" tag)
```

**Example 5 - MCP Tool Server Tagging**:
```go
// Use Case: Tag MCP tool errors for LLM categorization
// Keywords: mcp, llm, tool-server, semantic-tagging

err := ErrToolExecution.New(execErr).
    WithMCPCode(errific.MCPToolError).
    WithTags("mcp", "tool-execution", "database", "search", "transient").
    WithHelp("Database search tool failed due to connection timeout").
    WithSuggestion("Retry the search in a few seconds").
    WithRetryable(true).
    WithContext(Context{
        "tool_name": "search_database",
        "search_query": "find users",
    })

// LLM reads tags and understands:
// - "mcp" → This is an MCP tool error
// - "tool-execution" → Failed during execution (not parameter validation)
// - "database" → Related to database operations
// - "search" → Specifically a search operation
// - "transient" → Temporary issue (can retry)
```

**Best Practices**:
```go
// ✅ DO: Use lowercase, hyphenated tags
WithTags("user-input", "rate-limit", "external-api")

// ❌ DON'T: Use mixed case or spaces
WithTags("User Input", "RateLimit")  // Inconsistent!

// ✅ DO: Use multiple specific tags
WithTags("payment", "stripe", "timeout", "checkout")

// ❌ DON'T: Use single vague tag
WithTags("error")  // Useless!

// ✅ DO: Include team/ownership tags
WithTags("database", "backend-team", "postgres")

// ✅ DO: Include severity/priority tags when relevant
WithTags("critical", "high-priority", "user-facing")

// ❌ DON'T: Duplicate information from other fields
WithTags("retryable")  // Already have WithRetryable(true)!
// Better: Use tags for additional dimensions

// ✅ DO: Use consistent tag vocabulary across codebase
// Define standard tags: "payment", "database", "network", "validation"
```

**Retrieval**: `GetTags(err) []string`

**See Also**:
- `WithLabel()` / `WithLabels()` - For key-value pairs
- `WithCategory()` - For primary error classification
- `WithContext()` - For structured metadata

---

### `.WithLabel(key, value string) errific`

<!-- RAG: Add single key-value label for structured error metadata -->

**Purpose**: Add single key-value label for structured error classification.

**Why this matters**:
- **Structured Queries**: Query errors by exact key-value pairs (e.g., "severity=high")
- **Prometheus/OpenTelemetry**: Labels map directly to metric labels
- **Datadog/APM Integration**: Labels become tags in monitoring systems
- **Faceted Search**: Filter errors by multiple label dimensions
- **Cardinality Control**: Labels are better than tags for high-cardinality data
- **Type Safety**: Key-value structure enforces consistent labeling

**Parameters**:
- `key string` - Label key (e.g., "severity", "team", "region")
- `value string` - Label value (e.g., "high", "backend", "us-east-1")

**Returns**: errific error with label attached

**Chaining**: Can be chained with other methods (call multiple times for multiple labels)

**When to use**:
- ✅ For key-value metadata (severity, region, team, version)
- ✅ When integrating with Prometheus, OpenTelemetry, Datadog
- ✅ For structured filtering and aggregation
- ✅ When you need exact key-value matching

**When NOT to use**:
- ❌ For freeform tags (use WithTags instead)
- ❌ For complex structured data (use WithContext instead)
- ❌ For temporary debugging info (use WithContext instead)

**Example 1 - Error Severity Labeling**:
```go
// Use Case: Label errors by severity for alerting
// Keywords: severity, priority, alerting, filtering

// Critical error
err1 := ErrPaymentGatewayDown.New().
    WithLabel("severity", "critical").
    WithLabel("team", "payments").
    WithLabel("region", "us-east-1").
    WithRetryable(false)

// Warning error
err2 := ErrSlowQuery.New().
    WithLabel("severity", "warning").
    WithLabel("team", "backend").
    WithLabel("query_type", "analytics")

// Alert system:
if GetLabelValue(err, "severity") == "critical" {
    pagerduty.Alert(GetLabelValue(err, "team"))
}
```

**Example 2 - Prometheus Metrics Integration**:
```go
// Use Case: Export error metrics to Prometheus with labels
// Keywords: prometheus, metrics, observability, monitoring

err := ErrAPICall.New(httpErr).
    WithLabel("endpoint", "/api/users").
    WithLabel("method", "GET").
    WithLabel("status", "500").
    WithLabel("region", "us-west-2").
    WithContext(Context{
        "duration_ms": 1500,
        "user_id": "user-123",
    })

// Prometheus metric:
// api_errors_total{endpoint="/api/users",method="GET",status="500",region="us-west-2"} 1

// Prometheus query examples:
// rate(api_errors_total{status="500"}[5m])  → 500 errors per second
// sum by (endpoint) (api_errors_total)      → Errors grouped by endpoint
// api_errors_total{region="us-west-2"}      → Errors in specific region
```

**Example 3 - Multi-Tenant Error Tracking**:
```go
// Use Case: Track errors per tenant/customer
// Keywords: multi-tenant, saas, customer-tracking

err := ErrQuotaExceeded.New().
    WithLabel("tenant_id", "tenant-abc-123").
    WithLabel("plan", "free").
    WithLabel("resource", "api_requests").
    WithContext(Context{
        "quota_limit": 1000,
        "current_usage": 1000,
    })

// Support query: Find all errors for tenant "tenant-abc-123"
// Billing query: Count errors by "plan" label
// Resource query: Find quota errors for "api_requests" resource
```

**Example 4 - Deployment Version Tracking**:
```go
// Use Case: Track errors by deployment version for rollback decisions
// Keywords: deployment, version-tracking, rollback, canary

err := ErrFeatureExecution.New(featureErr).
    WithLabel("version", "v2.5.0").
    WithLabel("deployment", "canary").
    WithLabel("feature_flag", "new-checkout-flow").
    WithContext(Context{
        "deployed_at": deployTime,
        "commit_sha": "abc123",
    })

// Deployment dashboard:
// - Count errors in v2.5.0 vs v2.4.0
// - Compare canary deployment vs production
// - Decide: Too many errors in v2.5.0 → Rollback!
```

**Common Mistakes**:
```go
// ❌ DON'T: Use labels for high-cardinality data
err := ErrAPI.New().WithLabel("user_id", "user-123-456-789")  // Millions of users!

// ✅ DO: Use labels for low-cardinality dimensions
err := ErrAPI.New().WithLabel("region", "us-east-1")  // Only ~20 regions

// ❌ DON'T: Duplicate context data in labels
err := ErrDB.New().
    WithContext(Context{"query": sql}).
    WithLabel("query", sql)  // Duplicate!

// ✅ DO: Use labels for classification, context for details
err := ErrDB.New().
    WithContext(Context{"query": sql}).       // Detailed query
    WithLabel("query_type", "analytics")      // Classification

// ❌ DON'T: Use inconsistent label names
err1 := ErrAPI.New().WithLabel("severity", "high")
err2 := ErrAPI.New().WithLabel("priority", "high")  // Different key!

// ✅ DO: Use consistent label keys across codebase
err1 := ErrAPI.New().WithLabel("severity", "high")
err2 := ErrDB.New().WithLabel("severity", "critical")  // Same key
```

**Retrieval**: `GetLabelValue(err, key string) string`, `GetLabels(err) map[string]string`

**See Also**:
- `WithLabels()` - Add multiple labels at once
- `WithTags()` - For freeform semantic tags
- `WithContext()` - For detailed structured data

---

### `.WithLabels(labels map[string]string) errific`

<!-- RAG: Add multiple key-value labels at once for structured error metadata -->

**Purpose**: Add multiple key-value labels at once (batch version of WithLabel).

**Why this matters**:
- **Convenience**: Set multiple labels in one call
- **Consistency**: Ensures all related labels are set together
- **Prometheus/APM**: Matches metric label pattern (map of key-values)
- **Template Reuse**: Define label sets once and reuse across errors
- **Structured Queries**: Enable complex multi-dimensional filtering

**Parameters**:
- `labels map[string]string` - Map of label key-value pairs

**Returns**: errific error with all labels attached

**Chaining**: Can be chained with other methods

**Example 1 - Standard Label Template**:
```go
// Use Case: Reuse common label sets across application
// Keywords: template, consistency, reusability

// Define standard label templates
var standardLabels = map[string]string{
    "service": "api-gateway",
    "environment": "production",
    "region": "us-east-1",
    "version": "v2.1.0",
}

// Apply to all errors
err := ErrAPICall.New(httpErr).
    WithLabels(standardLabels).  // Batch apply all labels
    WithLabel("endpoint", "/users").  // Add specific label
    WithHTTPStatus(500)

// All errors now have consistent base labels
```

**Example 2 - OpenTelemetry Span Labels**:
```go
// Use Case: Match OpenTelemetry span attributes in errors
// Keywords: opentelemetry, tracing, observability

func handleRequest(ctx context.Context) error {
    span := trace.SpanFromContext(ctx)

    // Extract span attributes as labels
    spanLabels := map[string]string{
        "trace_id": span.SpanContext().TraceID().String(),
        "span_id": span.SpanContext().SpanID().String(),
        "service_name": "user-service",
        "operation": "get_user",
    }

    err := database.GetUser(userID)
    if err != nil {
        return ErrDatabaseQuery.New(err).
            WithLabels(spanLabels).  // Link error to trace
            WithContext(Context{
                "user_id": userID,
                "query": sql,
            })
    }
}
```

**Example 3 - Multi-Dimensional Monitoring**:
```go
// Use Case: Complex filtering across multiple dimensions
// Keywords: monitoring, filtering, dimensions, metrics

err := ErrCheckout.New(checkoutErr).
    WithLabels(map[string]string{
        "severity": "high",
        "team": "payments",
        "component": "checkout",
        "customer_tier": "enterprise",
        "payment_method": "credit_card",
        "region": "eu-west-1",
    }).
    WithContext(Context{
        "order_id": orderID,
        "amount": 9999.99,
    })

// Query examples:
// - severity=high AND team=payments
// - component=checkout AND region=eu-west-1
// - customer_tier=enterprise  (prioritize enterprise customer issues)
```

**Retrieval**: `GetLabels(err) map[string]string`

**See Also**:
- `WithLabel()` - Add single label
- `WithTags()` - For freeform tags
- `WithContext()` - For detailed structured data

---

### `.WithTimestamp(t time.Time) errific`

<!-- RAG: Add explicit timestamp for when error occurred (useful for async/batch processing) -->

**Purpose**: Add explicit timestamp for when the error occurred.

**Why this matters**:
- **Async Processing**: Track when error occurred vs when it was logged (batch jobs, queues)
- **Time-Series Analysis**: Accurate error timing for performance analysis
- **Distributed Systems**: Consistent timestamps across services with clock skew
- **Replay Scenarios**: Preserve original error time when replaying events
- **Audit Trails**: Exact error occurrence time for compliance
- **Latency Tracking**: Measure time between error occurrence and detection

**Parameters**:
- `t time.Time` - Timestamp when error occurred

**Returns**: errific error with timestamp attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Batch/async processing (error time ≠ log time)
- ✅ Queue processing with delays
- ✅ Event replay systems
- ✅ Distributed systems with clock skew
- ✅ When you need precise error timing

**When NOT to use**:
- ❌ Synchronous request-response (error timestamp = now)
- ❌ When clock sync is not important

**Example 1 - Batch Processing**:
```go
// Use Case: Process batch of events, preserve original error times
// Keywords: batch-processing, async, queue, timing

func processBatch(events []Event) {
    for _, event := range events {
        err := processEvent(event)
        if err != nil {
            // Error occurred now, but event is from 10 minutes ago
            batchErr := ErrEventProcessing.New(err).
                WithTimestamp(event.CreatedAt).  // Original event time
                WithContext(Context{
                    "event_id": event.ID,
                    "batch_id": batchID,
                    "processing_delay_seconds": time.Since(event.CreatedAt).Seconds(),
                    "processed_at": time.Now(),
                })

            // Log shows:
            // - Event created: 2:00 PM (from timestamp)
            // - Processing failed: 2:10 PM (from processed_at context)
            // - Delay: 10 minutes
        }
    }
}
```

**Example 2 - Message Queue Processing**:
```go
// Use Case: SQS/RabbitMQ message processing with delays
// Keywords: message-queue, sqs, rabbitmq, delay

func processMessage(msg *sqs.Message) error {
    // Message was queued 5 minutes ago
    enqueuedAt := time.Unix(msg.Attributes["SentTimestamp"], 0)

    err := processMessageContent(msg.Body)
    if err != nil {
        return ErrMessageProcessing.New(err).
            WithTimestamp(enqueuedAt).  // When message was created
            WithContext(Context{
                "message_id": msg.MessageId,
                "queue_time_seconds": time.Since(enqueuedAt).Seconds(),
                "attempt_count": msg.Attributes["ApproximateReceiveCount"],
            })
    }
}
```

**Example 3 - Distributed System Clock Skew**:
```go
// Use Case: Consistent timestamps across services with clock differences
// Keywords: distributed-systems, clock-skew, ntp, timing

// Service A (clock is 2 minutes fast)
errTime := time.Now()  // 3:02 PM (local clock)

err := ErrServiceA.New().
    WithTimestamp(errTime).
    WithCorrelationID(traceID)

// Service B (clock is accurate)
// Receives error from Service A
// Can see exact time error occurred on Service A (3:02 PM)
// Even though Service B's clock shows 3:00 PM

// Analysis: Can properly order events across services despite clock skew
```

**Example 4 - Event Replay**:
```go
// Use Case: Replay historical events while preserving original timestamps
// Keywords: event-sourcing, replay, time-travel, debugging

func replayEvents(events []HistoricalEvent) {
    for _, event := range events {
        // Replaying event from last week
        err := reprocessEvent(event)
        if err != nil {
            replayErr := ErrEventReplay.New(err).
                WithTimestamp(event.OriginalTimestamp).  // Last week
                WithContext(Context{
                    "event_id": event.ID,
                    "original_time": event.OriginalTimestamp,
                    "replay_time": time.Now(),
                    "time_difference_days": time.Since(event.OriginalTimestamp).Hours() / 24,
                })

            // Logs show both:
            // - When error originally happened (last week)
            // - When replay happened (today)
        }
    }
}
```

**Retrieval**: `GetTimestamp(err) time.Time`

**See Also**:
- `WithDuration()` - Track operation duration
- `WithContext()` - For multiple temporal fields

---

### `.WithDuration(d time.Duration) errific`

<!-- RAG: Add operation duration for performance tracking and timeout analysis -->

**Purpose**: Track how long an operation ran before failing.

**Why this matters**:
- **Performance Analysis**: Identify slow operations that fail
- **Timeout Debugging**: See if errors are timeout-related
- **SLA Monitoring**: Track operations exceeding SLA thresholds
- **Optimization Targets**: Find slow operations that need optimization
- **Latency Distribution**: Understand error latency patterns
- **Alerting**: Alert when operations consistently take too long before failing

**Parameters**:
- `d time.Duration` - How long the operation ran before failing

**Returns**: errific error with duration attached

**Chaining**: Can be chained with other methods

**When to use**:
- ✅ Database queries (track slow queries)
- ✅ API calls (measure latency)
- ✅ File operations (track I/O time)
- ✅ Any operation with time limits/SLAs
- ✅ When debugging timeouts

**When NOT to use**:
- ❌ Instant validation errors (duration is meaningless)
- ❌ When operation time is irrelevant

**Example 1 - Slow Database Query**:
```go
// Use Case: Track database query duration to find slow queries
// Keywords: database, performance, slow-query, optimization

func queryUsers(ctx context.Context, query string) error {
    start := time.Now()

    rows, err := db.QueryContext(ctx, query)
    duration := time.Since(start)

    if err != nil {
        return ErrDatabaseQuery.New(err).
            WithDuration(duration).  // Track how long it took to fail
            WithContext(Context{
                "query": query,
                "duration_ms": duration.Milliseconds(),
                "threshold_ms": 1000,  // Expected max 1 second
            })
    }

    // Analysis: If duration > 1s, query is slow and needs optimization
    // Even if query succeeds, log slow queries for monitoring
    if duration > time.Second {
        log.Warn("Slow query detected", "duration", duration, "query", query)
    }

    return nil
}
```

**Example 2 - API Timeout Analysis**:
```go
// Use Case: Determine if errors are timeout-related
// Keywords: api, timeout, latency, performance

func callExternalAPI(ctx context.Context, endpoint string) error {
    client := &http.Client{Timeout: 30 * time.Second}
    start := time.Now()

    resp, err := client.Get(endpoint)
    duration := time.Since(start)

    if err != nil {
        isTimeout := errors.Is(err, context.DeadlineExceeded)

        return ErrAPICall.New(err).
            WithDuration(duration).  // 30+ seconds if timeout
            WithRetryable(isTimeout).
            WithContext(Context{
                "endpoint": endpoint,
                "duration_ms": duration.Milliseconds(),
                "timeout_ms": 30000,
                "is_timeout": isTimeout,
                "duration_vs_timeout": duration.Seconds() / 30.0,  // 100% = hit timeout
            })
    }

    // Success but slow (warning)
    if duration > 10*time.Second {
        log.Warn("Slow API call", "duration", duration, "endpoint", endpoint)
    }

    return nil
}
```

**Example 3 - SLA Monitoring**:
```go
// Use Case: Track SLA violations for errors
// Keywords: sla, monitoring, performance, alerting

func processOrder(ctx context.Context, order Order) error {
    start := time.Now()
    slaThreshold := 5 * time.Second  // Orders must complete in 5s

    err := orderService.Process(ctx, order)
    duration := time.Since(start)

    if err != nil {
        slaViolation := duration > slaThreshold

        orderErr := ErrOrderProcessing.New(err).
            WithDuration(duration).
            WithContext(Context{
                "order_id": order.ID,
                "duration_ms": duration.Milliseconds(),
                "sla_threshold_ms": slaThreshold.Milliseconds(),
                "sla_violation": slaViolation,
                "sla_percentage": (duration.Seconds() / slaThreshold.Seconds()) * 100,
            })

        // Alert on SLA violations
        if slaViolation {
            alerting.NotifySLA("Order processing exceeded 5s SLA", orderErr)
        }

        return orderErr
    }

    return nil
}
```

**Example 4 - Performance Comparison**:
```go
// Use Case: Compare operation duration across different implementations
// Keywords: performance, comparison, ab-testing, optimization

func searchUsers(query string, useNewAlgorithm bool) error {
    start := time.Now()

    var err error
    if useNewAlgorithm {
        err = searchUsersV2(query)  // New optimized algorithm
    } else {
        err = searchUsersV1(query)  // Old algorithm
    }

    duration := time.Since(start)

    if err != nil {
        return ErrSearch.New(err).
            WithDuration(duration).
            WithContext(Context{
                "query": query,
                "algorithm": map[bool]string{true: "v2", false: "v1"}[useNewAlgorithm],
                "duration_ms": duration.Milliseconds(),
            })
    }

    // Metrics: Compare v1 vs v2 duration
    // Result: v2 is 3x faster (500ms vs 1500ms average)
    metrics.RecordDuration("search_duration", duration, map[string]string{
        "algorithm": map[bool]string{true: "v2", false: "v1"}[useNewAlgorithm],
    })

    return nil
}
```

**Best Practices**:
```go
// ✅ DO: Include duration in context as milliseconds for easy querying
err := ErrSlow.New().
    WithDuration(duration).
    WithContext(Context{
        "duration_ms": duration.Milliseconds(),  // Easy to query/graph
    })

// ✅ DO: Compare duration to thresholds
WithContext(Context{
    "duration_ms": 1500,
    "threshold_ms": 1000,
    "exceeded_by_ms": 500,
    "exceeded_by_percent": 50,  // 50% over threshold
})

// ✅ DO: Track both success and error durations for comparison
// Success: 200ms average
// Error: 5000ms average → Errors take 25x longer! (timeout?)

// ❌ DON'T: Use duration for instant errors
err := ErrInvalidEmail.New().
    WithDuration(50 * time.Nanosecond)  // Meaningless!

// ✅ DO: Use duration only for operations that take measurable time
err := ErrDatabaseQuery.New().
    WithDuration(1500 * time.Millisecond)  // Meaningful!
```

**Retrieval**: `GetDuration(err) time.Duration`

**See Also**:
- `WithTimestamp()` - Track when error occurred
- `WithRetryAfter()` - Specify retry delay
- `WithContext()` - For multiple performance metrics

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

## 📚 Complete Examples

### Example 1: API Service with Full Error Handling

**Scenario**: Building a REST API with consistent error responses

```go
package main

import (
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/leefernandes/errific"
)

// Define application errors
var (
    ErrInvalidInput  errific.Err = "invalid input"
    ErrUnauthorized  errific.Err = "unauthorized"
    ErrDBQuery       errific.Err = "database query failed"
    ErrNotFound      errific.Err = "resource not found"
)

type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
}

// API Handler
func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("id")
    
    user, err := getUser(userID)
    if err != nil {
        respondError(w, err)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}

// Business Logic with errific
func getUser(id string) (*User, error) {
    // Validation
    if id == "" {
        return nil, ErrInvalidInput.New().
            WithCode("VAL_USER_ID_REQUIRED").
            WithCategory(errific.CategoryValidation).
            WithHTTPStatus(400).
            WithContext(errific.Context{
                "field": "id",
                "message": "user ID is required",
            })
    }
    
    // Authorization
    if !hasPermission(id, "users:read") {
        return nil, ErrUnauthorized.New().
            WithCode("AUTH_USER_ACCESS_DENIED").
            WithCategory(errific.CategoryUnauthorized).
            WithHTTPStatus(403).
            WithContext(errific.Context{
                "user_id": id,
                "required_permission": "users:read",
            })
    }
    
    // Database Query
    user, err := db.QueryUser(id)
    if err == sql.ErrNoRows {
        return nil, ErrNotFound.New().
            WithCode("USER_NOT_FOUND").
            WithCategory(errific.CategoryNotFound).
            WithHTTPStatus(404).
            WithContext(errific.Context{"user_id": id})
    }
    if err != nil {
        return nil, ErrDBQuery.New(err).
            WithCode("DB_QUERY_USER_FAILED").
            WithCategory(errific.CategoryServer).
            WithHTTPStatus(500).
            WithContext(errific.Context{
                "query": "SELECT * FROM users WHERE id = ?",
                "user_id": id,
            }).
            WithRetryable(true).
            WithRetryAfter(5 * time.Second)
    }
    
    return user, nil
}

// Error Response Handler
func respondError(w http.ResponseWriter, err error) {
    status := errific.GetHTTPStatus(err)
    if status == 0 {
        status = 500  // Default
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": err,  // errific implements json.Marshaler
    })
}
```

**Example Responses**:

```json
// 400 Bad Request
{
  "error": {
    "error": "invalid input",
    "code": "VAL_USER_ID_REQUIRED",
    "category": "validation",
    "caller": "api/users.go:45.getUser",
    "context": {
      "field": "id",
      "message": "user ID is required"
    },
    "http_status": 400
  }
}

// 500 Internal Server Error
{
  "error": {
    "error": "database query failed: connection timeout",
    "code": "DB_QUERY_USER_FAILED",
    "category": "server",
    "caller": "api/users.go:78.getUser",
    "context": {
      "query": "SELECT * FROM users WHERE id = ?",
      "user_id": "user-123"
    },
    "retryable": true,
    "retry_after": "5s",
    "http_status": 500
  }
}
```

---

### Example 2: MCP Tool Server with AI-Ready Errors

**Scenario**: Building an MCP server that LLMs can interact with

```go
package main

import (
    "encoding/json"
    "fmt"
    
    "github.com/leefernandes/errific"
)

// MCP Tool Errors
var (
    ErrToolNotFound   errific.Err = "tool not found"
    ErrInvalidParams  errific.Err = "invalid tool parameters"
    ErrToolExecution  errific.Err = "tool execution failed"
)

type MCPRequest struct {
    JSONRPC       string                 `json:"jsonrpc"`
    ID            string                 `json:"id"`
    Method        string                 `json:"method"`
    Params        map[string]interface{} `json:"params"`
    CorrelationID string                 `json:"correlation_id,omitempty"`
}

type MCPResponse struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      string                 `json:"id"`
    Result  interface{}            `json:"result,omitempty"`
    Error   *errific.MCPError      `json:"error,omitempty"`
}

// MCP Server Handler
func HandleMCPRequest(r *MCPRequest) *MCPResponse {
    // Validate method exists
    if !toolRegistry.Has(r.Method) {
        err := ErrToolNotFound.New().
            WithMCPCode(errific.MCPMethodNotFound).
            WithHelp(fmt.Sprintf("Tool '%s' is not available", r.Method)).
            WithSuggestion("Use the 'list_tools' method to see available tools").
            WithDocs("https://docs.example.com/mcp/tools").
            WithTags("mcp", "tool-not-found", "validation")
        
        mcpErr := errific.ToMCPError(err)
        return &MCPResponse{
            JSONRPC: "2.0",
            ID:      r.ID,
            Error:   &mcpErr,
        }
    }
    
    // Validate parameters
    if err := validateToolParams(r.Method, r.Params); err != nil {
        toolErr := ErrInvalidParams.New(err).
            WithMCPCode(errific.MCPInvalidParams).
            WithHelp("The parameters provided do not match the tool schema").
            WithSuggestion("Check the tool documentation for required parameters").
            WithDocs(fmt.Sprintf("https://docs.example.com/mcp/tools/%s", r.Method)).
            WithTags("mcp", "invalid-params", "validation").
            WithContext(errific.Context{
                "tool": r.Method,
                "provided_params": r.Params,
            })
        
        mcpErr := errific.ToMCPError(toolErr)
        return &MCPResponse{
            JSONRPC: "2.0",
            ID:      r.ID,
            Error:   &mcpErr,
        }
    }
    
    // Execute tool
    result, err := toolRegistry.Execute(r.Method, r.Params)
    if err != nil {
        // Tool execution failed - create rich error
        execErr := ErrToolExecution.New(err).
            WithMCPCode(errific.MCPToolError).
            WithCorrelationID(r.CorrelationID).
            WithRequestID(r.ID).
            WithHelp(getToolHelp(r.Method, err)).
            WithSuggestion(getToolSuggestion(r.Method, err)).
            WithDocs(getToolDocs(r.Method)).
            WithTags("mcp", "tool-error", getToolCategory(r.Method)).
            WithLabels(map[string]string{
                "tool_name": r.Method,
                "error_type": classifyError(err),
                "severity": calculateSeverity(err),
            }).
            WithRetryable(isRetryable(err)).
            WithRetryAfter(getRetryDelay(err))
        
        mcpErr := errific.ToMCPError(execErr)
        return &MCPResponse{
            JSONRPC: "2.0",
            ID:      r.ID,
            Error:   &mcpErr,
        }
    }
    
    // Success
    return &MCPResponse{
        JSONRPC: "2.0",
        ID:      r.ID,
        Result:  result,
    }
}

// Helper functions
func getToolHelp(toolName string, err error) string {
    // Return context-specific help based on error
    switch {
    case isDatabaseError(err):
        return "Database connection pool exhausted. The database is under heavy load."
    case isNetworkError(err):
        return "Network connectivity issue. Unable to reach external service."
    default:
        return fmt.Sprintf("Tool '%s' encountered an unexpected error", toolName)
    }
}

func getToolSuggestion(toolName string, err error) string {
    switch {
    case isDatabaseError(err):
        return "Retry in 10 seconds when connections are released, or simplify your query."
    case isNetworkError(err):
        return "Check network connectivity and retry in 30 seconds."
    default:
        return "Contact support if the issue persists."
    }
}
```

**LLM Receives** (for database error):

```json
{
  "jsonrpc": "2.0",
  "id": "req-789",
  "error": {
    "code": -32000,
    "message": "tool execution failed: database connection failed",
    "data": {
      "error": "tool execution failed",
      "code": "TOOL_DB_CONN_FAILED",
      "correlation_id": "trace-abc-123",
      "request_id": "req-789",
      "help": "Database connection pool exhausted. The database is under heavy load.",
      "suggestion": "Retry in 10 seconds when connections are released, or simplify your query.",
      "docs": "https://docs.example.com/mcp/tools/search_database#errors",
      "tags": ["mcp", "tool-error", "database"],
      "labels": {
        "tool_name": "search_database",
        "error_type": "connection",
        "severity": "high"
      },
      "retryable": true,
      "retry_after": "10s",
      "caller": "tools/search.go:45.Execute"
    }
  }
}
```

**LLM Decision Making**:
1. Read `help` → Explain to user: "The database is too busy"
2. Read `suggestion` → Take action: Wait 10s and retry
3. Check `retryable` → Decide: Yes, safe to retry
4. Read `docs` → Provide link to user
5. Check `labels.severity` → Know: This is high priority

---

### Example 3: Distributed Microservices with Correlation Tracking

**Scenario**: Tracing errors across multiple microservices

```go
package main

import (
    "context"
    "github.com/google/uuid"
    "github.com/leefernandes/errific"
)

// Service-specific errors
var (
    // Gateway errors
    ErrGatewayAuth    errific.Err = "gateway authentication failed"
    
    // User service errors
    ErrUserNotFound   errific.Err = "user not found"
    ErrUserQuery      errific.Err = "user query failed"
    
    // Database service errors
    ErrDBConnection   errific.Err = "database connection failed"
    ErrDBQuery        errific.Err = "database query failed"
)

// ============================================================
// Service A: API Gateway
// ============================================================

func (gw *Gateway) HandleRequest(w http.ResponseWriter, r *http.Request) {
    // Create correlation ID for entire request chain
    correlationID := uuid.New().String()
    requestID := uuid.New().String()
    
    ctx := context.WithValue(r.Context(), "correlation_id", correlationID)
    ctx = context.WithValue(ctx, "request_id", requestID)
    
    userID := r.Header.Get("X-User-ID")
    
    // Call user service
    user, err := gw.userService.GetUser(ctx, userID)
    if err != nil {
        // Log with correlation tracking
        log.Error("request failed",
            "correlation_id", errific.GetCorrelationID(err),
            "request_id", errific.GetRequestID(err),
            "service_chain", "gateway → user-service → db-service",
            "error", err)
        
        respondError(w, err)
        return
    }
    
    json.NewEncoder(w).Encode(user)
}

// ============================================================
// Service B: User Service
// ============================================================

func (us *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    correlationID := ctx.Value("correlation_id").(string)
    requestID := ctx.Value("request_id").(string)
    
    // Query database service
    user, err := us.dbService.QueryUser(ctx, userID)
    if err != nil {
        // Wrap error with service context
        return nil, ErrUserQuery.New(err).
            WithCorrelationID(correlationID).
            WithRequestID(requestID).
            WithLabel("service", "user-service").
            WithLabel("operation", "get_user").
            WithContext(errific.Context{
                "user_id": userID,
            })
    }
    
    return user, nil
}

// ============================================================
// Service C: Database Service
// ============================================================

func (db *DatabaseService) QueryUser(ctx context.Context, userID string) (*User, error) {
    correlationID := ctx.Value("correlation_id").(string)
    requestID := ctx.Value("request_id").(string)
    
    query := "SELECT id, email, name FROM users WHERE id = ?"
    
    var user User
    err := db.conn.QueryRowContext(ctx, query, userID).Scan(&user.ID, &user.Email, &user.Name)
    
    if err == sql.ErrNoRows {
        return nil, ErrUserNotFound.New().
            WithCorrelationID(correlationID).
            WithRequestID(requestID).
            WithLabel("service", "db-service").
            WithLabel("operation", "query_user").
            WithHTTPStatus(404).
            WithContext(errific.Context{
                "query": query,
                "user_id": userID,
            })
    }
    
    if err != nil {
        return nil, ErrDBQuery.New(err).
            WithCorrelationID(correlationID).  // Same correlation ID!
            WithRequestID(requestID).
            WithLabel("service", "db-service").
            WithLabel("operation", "query_user").
            WithHTTPStatus(500).
            WithRetryable(true).
            WithRetryAfter(5 * time.Second).
            WithContext(errific.Context{
                "query": query,
                "user_id": userID,
            })
    }
    
    return &user, nil
}
```

**Log Output** (with correlation tracking):

```json
// All logs from the same request have the same correlation_id
{
  "level": "error",
  "service": "gateway",
  "correlation_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "request_id": "req-abc-123",
  "message": "request failed",
  "error": {
    "error": "user query failed: database query failed: connection timeout",
    "correlation_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "request_id": "req-abc-123",
    "labels": {
      "service": "user-service",
      "operation": "get_user"
    }
  }
}
```

**Benefits**:
- ✅ Single correlation ID traces through all services
- ✅ Each service adds its own context
- ✅ Easy to find all logs for a single request in log aggregation
- ✅ Service labels enable filtering by service in monitoring

---

## Performance

**Benchmarks** (Go 1.23, Apple M1, 14 cores):
```
BenchmarkErrNew                    2,245,603    ~523 ns/op     680 B/op   9 allocs/op
BenchmarkErrError                 10,321,795    ~115 ns/op     192 B/op   6 allocs/op
BenchmarkWithContext               1,953,436    ~616 ns/op     696 B/op   9 allocs/op
BenchmarkJSONMarshal               1,391,380    ~862 ns/op    1105 B/op   7 allocs/op
BenchmarkWithCorrelationID         2,121,558    ~567 ns/op     688 B/op   8 allocs/op
BenchmarkWithTags                  2,084,160    ~575 ns/op     720 B/op   9 allocs/op
BenchmarkToMCPError                1,596,895    ~743 ns/op    1730 B/op   6 allocs/op
BenchmarkCompleteErrorChain        1,758,771    ~680 ns/op     720 B/op   9 allocs/op
```

**Overhead**: Sub-microsecond for most operations, negligible for error handling.

**Memory**: ~680-720 bytes per error with metadata, ~1KB with MCP conversion.

**Thread Safety**: All operations are thread-safe with minimal lock contention.

---

## Version Compatibility

- **Go Version**: 1.20+
- **Dependencies**: None (stdlib only)
- **Thread-Safety**: Full (mutex-protected configuration)
- **Breaking Changes**: None (fully backward compatible)
