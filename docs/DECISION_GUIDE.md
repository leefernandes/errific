# errific Decision Guide

**For AI Agents and Automated Systems**

## Quick Reference

### When to Use errific

✅ **Use errific when you need:**
- Automated error handling and retry logic
- Structured logging with context
- Machine-readable error codes and categories
- HTTP API error responses
- Error tracking and monitoring integration
- AI agent decision-making based on errors
- Debugging with automatic caller information

❌ **Don't use errific when:**
- Building a library (don't impose error handling on consumers)
- Minimal dependencies are required
- Error handling is already standardized in your ecosystem
- Simple stdlib errors are sufficient

---

## Error Code Decision Tree

### Should I add an error code?

```
START
  │
  ├─ Will this error be tracked/monitored? → YES → Add code
  ├─ Will AI/automation handle this error? → YES → Add code
  ├─ Does this error map to documentation? → YES → Add code
  ├─ Do you need error-specific alerts? → YES → Add code
  └─ Is this a one-off error? → YES → No code needed
```

### Code Naming Convention

```
Format: DOMAIN_TYPE_NUMBER

Examples:
- DB_CONN_001        (Database connection error #1)
- API_TIMEOUT_001    (API timeout error #1)
- VAL_EMAIL_001      (Email validation error #1)
- AUTH_TOKEN_001     (Auth token error #1)
- FILE_READ_001      (File read error #1)
```

---

## Category Decision Tree

### Which category should I use?

```
START: Analyze the error cause
  │
  ├─ User provided bad input?
  │   └─ CategoryValidation or CategoryClient
  │
  ├─ Resource doesn't exist?
  │   └─ CategoryNotFound
  │
  ├─ Authentication/permission issue?
  │   └─ CategoryUnauthorized
  │
  ├─ Network connectivity problem?
  │   └─ CategoryNetwork
  │
  ├─ Operation took too long?
  │   └─ CategoryTimeout
  │
  ├─ Internal system failure?
  │   └─ CategoryServer
  │
  └─ Unsure?
      └─ Check: Can user fix it? → CategoryClient
         Check: System issue? → CategoryServer
```

### Category → HTTP Status Mapping

| Category | Default HTTP | When to Use |
|----------|-------------|-------------|
| CategoryValidation | 400 | Input validation failed |
| CategoryClient | 400 | General client error |
| CategoryUnauthorized | 401/403 | Auth/permission denied |
| CategoryNotFound | 404 | Resource missing |
| CategoryTimeout | 408/504 | Request/gateway timeout |
| CategoryServer | 500 | Internal server error |
| CategoryNetwork | 503 | Service unavailable |

---

## Retry Decision Tree

### Should this error be retryable?

```
START: Analyze error characteristics
  │
  ├─ Is error caused by user input? → NO → Not retryable
  │
  ├─ Is error a validation failure? → NO → Not retryable
  │
  ├─ Is error due to auth/permissions? → NO → Not retryable
  │
  ├─ Is error "not found"? → NO → Not retryable
  │
  ├─ Is error temporary/transient?
  │   ├─ Network timeout → YES → Retryable
  │   ├─ Connection refused → YES → Retryable
  │   ├─ Rate limit → YES → Retryable (with delay)
  │   ├─ Service unavailable → YES → Retryable
  │   ├─ Resource exhausted → YES → Retryable (with backoff)
  │   └─ Temporary outage → YES → Retryable
  │
  └─ Default → Analyze case-by-case
```

### How long should retry delay be?

```
Error Type              → Suggested Delay

Transient glitch        → 1 second
Network timeout         → 5 seconds
Rate limit (known)      → Use Retry-After header value
Rate limit (unknown)    → 30 seconds
Service maintenance     → 5 minutes
Resource exhaustion     → 10 seconds
Connection pool full    → 2 seconds
```

### How many retries?

```
Operation Type          → Max Retries

Critical operation      → 5 retries
Standard operation      → 3 retries
Expensive operation     → 1 retry
User-facing operation   → 2 retries
Background job          → 10 retries
```

---

## Context Decision Tree

### What should I include in context?

```
Include:
  ✅ Identifiers (user_id, order_id, request_id)
  ✅ Quantities (duration_ms, size_bytes, count)
  ✅ Operation details (query, endpoint, file_path)
  ✅ State information (retry_count, pool_size)
  ✅ Diagnostic data (status_code, error_code)

Exclude:
  ❌ Sensitive data (passwords, tokens, API keys)
  ❌ Large data (full request/response bodies)
  ❌ Non-JSON-serializable values (channels, functions)
  ❌ Redundant data (already in error message)
```

### Context by Operation Type

**Database Operations**:
```go
Context{
    "query": sql,
    "duration_ms": elapsed.Milliseconds(),
    "table": "users",
    "connection_id": connID,
    "rows_affected": count,
}
```

**HTTP/API Calls**:
```go
Context{
    "endpoint": url,
    "method": "POST",
    "status_code": resp.StatusCode,
    "duration_ms": elapsed.Milliseconds(),
    "retry_count": attempt,
    "request_id": reqID,
}
```

**File Operations**:
```go
Context{
    "path": filePath,
    "operation": "read",
    "size_bytes": fileSize,
    "permissions": fileMode.String(),
}
```

**Business Logic**:
```go
Context{
    "user_id": userID,
    "order_id": orderID,
    "amount": amount,
    "currency": "USD",
    "state": currentState,
}
```

---

## HTTP Status Decision Tree

### What HTTP status should I set?

```
START: Analyze error category
  │
  ├─ CategoryValidation → 400 Bad Request
  │   Exception: Specific field errors → 422 Unprocessable Entity
  │
  ├─ CategoryClient → 400 Bad Request
  │
  ├─ CategoryUnauthorized
  │   ├─ Missing credentials → 401 Unauthorized
  │   └─ Insufficient permissions → 403 Forbidden
  │
  ├─ CategoryNotFound → 404 Not Found
  │
  ├─ CategoryTimeout
  │   ├─ Client timeout → 408 Request Timeout
  │   └─ Upstream timeout → 504 Gateway Timeout
  │
  ├─ CategoryNetwork → 503 Service Unavailable
  │   Exception: Bad Gateway → 502 Bad Gateway
  │
  └─ CategoryServer → 500 Internal Server Error
```

### Special Cases

| Scenario | Status | Category |
|----------|--------|----------|
| Rate limit exceeded | 429 | CategoryClient |
| Method not allowed | 405 | CategoryClient |
| Conflict (duplicate) | 409 | CategoryClient |
| Gone (deleted) | 410 | CategoryNotFound |
| Payload too large | 413 | CategoryValidation |
| URI too long | 414 | CategoryValidation |
| Unsupported media | 415 | CategoryValidation |
| Service in maintenance | 503 | CategoryServer |

---

## JSON Serialization Decision

### Should I serialize this error to JSON?

✅ **Serialize when:**
- Returning error in API response
- Writing to structured logs (JSON Lines, ELK)
- Sending to monitoring system (Datadog, Sentry)
- Storing error for later analysis
- Passing error between services

❌ **Don't serialize when:**
- Writing to simple text logs
- Displaying error to end user directly
- Error is temporary/debugging only
- Performance is critical (hot path)

---

## Migration Decision Guide

### From stdlib errors

```go
// Before: stdlib
return errors.New("database failed")

// After: errific basic
var ErrDatabase Err = "database failed"
return ErrDatabase.New()

// After: errific with metadata
return ErrDatabase.New().
    WithCode("DB_001").
    WithCategory(CategoryServer).
    WithRetryable(true)
```

### From pkg/errors

```go
// Before: pkg/errors
return errors.Wrap(err, "failed to query")

// After: errific
var ErrQuery Err = "failed to query"
return ErrQuery.New(err)

// With context
return ErrQuery.New(err).WithContext(Context{
    "query": sql,
})
```

### From github.com/cockroachdb/errors

```go
// Before: cockroachdb/errors
return errors.WithSecondaryError(err, secondaryErr)

// After: errific
return ErrPrimary.New(err, secondaryErr)

// With metadata
return ErrPrimary.New(err, secondaryErr).
    WithCode("ERR_001").
    WithContext(Context{"details": "..."})
```

---

## AI Agent Automation Patterns

### Pattern 1: Automatic Retry

```go
func callWithRetry(fn func() error) error {
    var lastErr error

    for attempt := 0; attempt < 10; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }

        lastErr = err

        // AI decision: Should retry?
        if !IsRetryable(err) {
            return err
        }

        // AI decision: How long to wait?
        delay := GetRetryAfter(err)
        if delay == 0 {
            // Exponential backoff
            delay = time.Duration(math.Pow(2, float64(attempt))) * time.Second
        }

        // AI decision: Max retries?
        maxRetries := GetMaxRetries(err)
        if maxRetries > 0 && attempt >= maxRetries {
            return err
        }

        time.Sleep(delay)
    }

    return lastErr
}
```

### Pattern 2: Automatic HTTP Response

```go
func handleError(w http.ResponseWriter, err error) {
    // AI decision: What status code?
    status := GetHTTPStatus(err)
    if status == 0 {
        // Fallback based on category
        switch GetCategory(err) {
        case CategoryClient, CategoryValidation:
            status = http.StatusBadRequest
        case CategoryUnauthorized:
            status = http.StatusUnauthorized
        case CategoryNotFound:
            status = http.StatusNotFound
        case CategoryTimeout:
            status = http.StatusGatewayTimeout
        default:
            status = http.StatusInternalServerError
        }
    }

    // AI decision: What to log?
    logError(err, status)

    // AI decision: What to return?
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(err)
}
```

### Pattern 3: Automatic Logging

```go
func logError(err error, fields ...interface{}) {
    // AI decision: What severity?
    level := "error"
    switch GetCategory(err) {
    case CategoryValidation, CategoryClient:
        level = "warn"
    case CategoryServer, CategoryNetwork:
        level = "error"
    }

    // AI decision: What metadata?
    entry := logger.WithFields(logrus.Fields{
        "error_code": GetCode(err),
        "category":   string(GetCategory(err)),
        "retryable":  IsRetryable(err),
    })

    // Add context
    if ctx := GetContext(err); ctx != nil {
        for k, v := range ctx {
            entry = entry.WithField(k, v)
        }
    }

    // Log at appropriate level
    entry.Log(logrus.Level(level), err.Error())
}
```

### Pattern 4: Automatic Alerting

```go
func checkAlert(err error) {
    code := GetCode(err)

    // AI decision: Alert based on code?
    if strings.HasPrefix(code, "DB_") {
        // Database errors - alert DBA team
        alertTeam("dba", err)
    }

    // AI decision: Alert based on category?
    if GetCategory(err) == CategoryServer {
        // Server errors - alert ops team
        alertTeam("ops", err)
    }

    // AI decision: Alert based on context?
    if ctx := GetContext(err); ctx != nil {
        if duration, ok := ctx["duration_ms"].(int); ok && duration > 5000 {
            // Slow operations - alert performance team
            alertTeam("performance", err)
        }
    }
}
```

---

## Best Practices Summary

1. **Always use typed errors** (`var ErrX Err = "..."`) for testability
2. **Add codes to errors** that need tracking or automation
3. **Use categories** for all errors that AI/automation will handle
4. **Set retryable** explicitly for all transient errors
5. **Include context** with operation-specific details
6. **Set HTTP status** for all API-facing errors
7. **Test with `errors.Is()`** not string comparison
8. **Serialize to JSON** for structured logging
9. **Use helper functions** to extract metadata
10. **Document error codes** in a central registry
