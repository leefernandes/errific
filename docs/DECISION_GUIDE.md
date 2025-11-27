# errific Decision Guide

<!-- RAG: Decision trees and automation patterns for errific error handling library -->

**For AI Agents and Automated Systems**

## Quick Reference

<!-- RAG: Quick decision checklist for when to use errific -->

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

<!-- RAG: Complete retry decision logic for automated error handling -->

### Should this error be retryable?

**Why this decision matters**:
- **Prevents Retry Storms**: Retrying validation errors creates infinite loops and wastes resources
- **Enables Resilience**: Correctly marking transient errors allows automatic recovery from temporary failures
- **Saves Money**: Non-retryable errors fail fast, reducing API costs and resource usage
- **Improves UX**: Fast failure for user errors provides immediate feedback instead of delays

```
START: Analyze error characteristics
  │
  ├─ Is error caused by user input? → NO → Not retryable
  │   Example: Invalid email format, password too short
  │   Reason: Retrying won't fix bad user input
  │
  ├─ Is error a validation failure? → NO → Not retryable
  │   Example: Required field missing, value out of range
  │   Reason: Data won't magically become valid
  │
  ├─ Is error due to auth/permissions? → NO → Not retryable
  │   Example: Invalid API key, expired token, insufficient permissions
  │   Reason: Retrying without new credentials will fail
  │
  ├─ Is error "not found"? → NO → Not retryable
  │   Example: User not found, file doesn't exist, 404 error
  │   Reason: Resource won't appear on retry
  │
  ├─ Is error temporary/transient?
  │   ├─ Network timeout → YES → Retryable
  │   │   Example: Connection timeout after 30s
  │   │   Reason: Network might recover
  │   │   Delay: 5 seconds
  │   │
  │   ├─ Connection refused → YES → Retryable
  │   │   Example: Server not accepting connections
  │   │   Reason: Server might restart/recover
  │   │   Delay: 5-10 seconds
  │   │
  │   ├─ Rate limit → YES → Retryable (with delay)
  │   │   Example: HTTP 429, API quota exceeded
  │   │   Reason: Rate limit window will reset
  │   │   Delay: Use Retry-After header (30-60 seconds)
  │   │
  │   ├─ Service unavailable → YES → Retryable
  │   │   Example: HTTP 503, database temporarily down
  │   │   Reason: Service might recover
  │   │   Delay: 10-30 seconds
  │   │
  │   ├─ Resource exhausted → YES → Retryable (with backoff)
  │   │   Example: Connection pool full, memory limit
  │   │   Reason: Resources will be released
  │   │   Delay: 2-10 seconds with exponential backoff
  │   │
  │   └─ Temporary outage → YES → Retryable
  │       Example: Planned maintenance, rolling deployment
  │       Reason: Service will return
  │       Delay: 60-300 seconds
  │
  └─ Default → Analyze case-by-case
      Check error message, logs, and context for clues
```

### Decision Outcome Examples

**Example 1: NOT Retryable - Validation Error**
```go
// Use Case: User provided invalid email
// Decision: NOT retryable (user input error)
// Keywords: validation, non-retryable, user-error

var ErrInvalidEmail Err = "invalid email format"

err := ErrInvalidEmail.New().
    WithRetryable(false).              // ❌ User input won't fix itself
    WithCategory(CategoryValidation).
    WithHTTPStatus(400).
    WithContext(Context{
        "field": "email",
        "value": "bad-email",          // No @ sign
        "constraint": "must contain @",
    })

// AI Agent Decision:
if !IsRetryable(err) {
    // Return 400 immediately, don't waste time retrying
    w.WriteHeader(400)
    json.NewEncoder(w).Encode(err)
    return nil  // Done, no retry
}
```

**Example 2: Retryable - Network Timeout**
```go
// Use Case: API call timed out after 30 seconds
// Decision: Retryable (transient network issue)
// Keywords: network, timeout, retryable, transient

var ErrAPITimeout Err = "API request timeout"

err := ErrAPITimeout.New(netErr).
    WithRetryable(true).               // ✅ Network might recover
    WithRetryAfter(5 * time.Second).   // Wait 5s for network to stabilize
    WithMaxRetries(3).                 // Try up to 3 times
    WithCategory(CategoryTimeout).
    WithHTTPStatus(504)

// AI Agent Decision:
for attempt := 0; attempt < GetMaxRetries(err); attempt++ {
    err := callAPI()
    if err == nil {
        return nil  // Success!
    }
    if !IsRetryable(err) {
        return err  // Non-retryable, fail immediately
    }

    log.Info("Retrying after timeout",
        "attempt", attempt+1,
        "delay", GetRetryAfter(err))

    time.Sleep(GetRetryAfter(err))
}
return err  // Failed after max retries
```

**Example 3: Retryable with Backoff - Rate Limit**
```go
// Use Case: Hit API rate limit (HTTP 429)
// Decision: Retryable with specific delay from header
// Keywords: rate-limit, retry-after, backoff

var ErrRateLimit Err = "rate limit exceeded"

// Parse Retry-After header from response
retryAfterHeader := resp.Header.Get("Retry-After")
retryAfter, _ := time.ParseDuration(retryAfterHeader + "s")

err := ErrRateLimit.New().
    WithRetryable(true).                // ✅ Temporary limit
    WithRetryAfter(retryAfter).         // Use server's suggested delay
    WithMaxRetries(1).                  // Only retry once (avoid ban)
    WithHTTPStatus(429).
    WithContext(Context{
        "limit": "100 requests/hour",
        "reset_at": time.Now().Add(retryAfter).Unix(),
        "remaining": 0,
    })

// AI Agent Decision:
if IsRetryable(err) && GetMaxRetries(err) > 0 {
    delay := GetRetryAfter(err)
    log.Warn("Rate limit hit, waiting",
        "delay", delay,
        "reset_at", GetContext(err)["reset_at"])

    time.Sleep(delay)
    return retry()
}
```

**Example 4: NOT Retryable - Not Found**
```go
// Use Case: User ID doesn't exist in database
// Decision: NOT retryable (resource missing)
// Keywords: not-found, non-retryable, 404

var ErrUserNotFound Err = "user not found"

err := ErrUserNotFound.New().
    WithRetryable(false).              // ❌ User won't appear on retry
    WithCategory(CategoryNotFound).
    WithHTTPStatus(404).
    WithContext(Context{
        "user_id": "user-123",
        "checked_at": time.Now().Unix(),
    })

// AI Agent Decision:
if !IsRetryable(err) {
    // Resource doesn't exist, retrying is pointless
    return err
}
```

**Example 5: Retryable - Connection Pool Exhausted**
```go
// Use Case: Database connection pool is full
// Decision: Retryable with backoff (connections will free up)
// Keywords: resource-exhaustion, retryable, backoff

var ErrConnectionPoolFull Err = "connection pool exhausted"

err := ErrConnectionPoolFull.New(dbErr).
    WithRetryable(true).               // ✅ Connections will be released
    WithRetryAfter(2 * time.Second).   // Short delay for connection release
    WithMaxRetries(5).                 // More retries for resource contention
    WithCategory(CategoryServer).
    WithContext(Context{
        "pool_size": 100,
        "in_use": 100,
        "waiting": 15,
    })

// AI Agent Decision with Exponential Backoff:
for attempt := 0; attempt < GetMaxRetries(err); attempt++ {
    err := executeQuery()
    if err == nil {
        return nil
    }
    if !IsRetryable(err) {
        return err
    }

    // Exponential backoff: 2s, 4s, 8s, 16s, 32s
    delay := GetRetryAfter(err) * time.Duration(1<<attempt)
    log.Debug("Connection pool full, backing off",
        "attempt", attempt+1,
        "delay", delay)

    time.Sleep(delay)
}
```

---

### How long should retry delay be?

<!-- RAG: Retry delay recommendations by error type -->

**Why delay matters**:
- **Too short**: Retry storm, waste resources, risk bans
- **Too long**: Poor user experience, unnecessary waiting
- **Just right**: Balance between recovery time and responsiveness

```
Error Type              → Suggested Delay      → Reasoning

Transient glitch        → 1 second             → Quick retry, minimal impact
Network timeout         → 5 seconds            → Give network time to recover
Rate limit (known)      → Use Retry-After      → Respect server's guidance
Rate limit (unknown)    → 30-60 seconds        → Conservative to avoid ban
Service maintenance     → 5 minutes            → Wait for maintenance window
Resource exhaustion     → 2-10 seconds         → Resources free up quickly
Connection pool full    → 2 seconds            → Connections release fast
Database deadlock       → 100 milliseconds     → Retry immediately
Temporary file lock     → 500 milliseconds     → Lock releases quickly
```

---

### How many retries?

<!-- RAG: Maximum retry count recommendations by operation type -->

**Why retry count matters**:
- **Too few**: Miss recovery opportunities
- **Too many**: Waste resources, delay failure detection
- **Just right**: Balance resilience with efficiency

```
Operation Type          → Max Retries  → Reasoning

Critical operation      → 5 retries    → Must succeed, worth extra attempts
Standard operation      → 3 retries    → Balance between reliability & cost
Expensive operation     → 1 retry      → Limit resource consumption
User-facing operation   → 2 retries    → Quick feedback, avoid frustration
Background job          → 10 retries   → Can wait, eventual consistency OK
Idempotent operation    → 5 retries    → Safe to retry multiple times
Non-idempotent operation→ 1 retry      → Risk of duplicate actions
Health check            → 0 retries    → Immediate status needed
```

---

## Context Decision Tree

<!-- RAG: Decision guide for what context metadata to include in errors for debugging and monitoring -->

### What should I include in context?

**Why context matters**:
- **Root Cause Analysis**: Preserves exact state when error occurred
- **Debugging**: Provides operation parameters without checking code
- **Monitoring**: Extract metrics (duration, size) from error context
- **AI Decision-Making**: Agents read context to decide actions
- **Audit Trails**: Track required compliance fields

```
START: What information helps debug this error?
  │
  ├─ Include ✅:
  │   ├─ Identifiers (user_id, order_id, request_id, correlation_id)
  │   ├─ Quantities (duration_ms, size_bytes, count, retry_attempt)
  │   ├─ Operation details (query, endpoint, file_path, command)
  │   ├─ State information (retry_count, pool_size, queue_depth)
  │   ├─ Diagnostic data (status_code, error_code, step)
  │   └─ Thresholds (max_size, timeout_ms, sla_threshold)
  │
  ├─ Exclude ❌:
  │   ├─ Sensitive data (passwords, tokens, API keys, credit cards)
  │   ├─ Large data (full request/response bodies, file contents)
  │   ├─ Non-JSON-serializable (channels, functions, interfaces)
  │   ├─ Redundant data (already in error message)
  │   └─ PII without hashing (emails, phone numbers, addresses)
  │
  └─ Decision checklist:
      1. Would this help me debug in production? → Include
      2. Is this sensitive/secret? → Exclude or hash
      3. Is this >1KB? → Summarize instead of full content
      4. Can this be serialized to JSON? → If no, exclude
      5. Does error message already say this? → Exclude (redundant)
```

### Context by Operation Type

**Database Operations**:
```go
// Use Case: Database query failed, need full debugging context
// Keywords: database, sql, performance, debugging

Context{
    "query": sql,                                    // What query failed
    "duration_ms": elapsed.Milliseconds(),          // How long it took
    "table": "users",                               // Which table
    "connection_id": connID,                        // Which connection
    "rows_affected": count,                         // Impact
    "pool_size": db.Stats().OpenConnections,       // Pool state
    "pool_in_use": db.Stats().InUse,              // Active connections
    "threshold_ms": 1000,                          // Expected max duration
}
```

**HTTP/API Calls**:
```go
// Use Case: External API call failed
// Keywords: http, api, external-service, timeout

Context{
    "endpoint": url,                         // Which API
    "method": "POST",                        // HTTP method
    "status_code": resp.StatusCode,         // Response status
    "duration_ms": elapsed.Milliseconds(),   // Latency
    "retry_count": attempt,                  // Which retry attempt
    "request_id": reqID,                    // Request tracking
    "timeout_ms": 30000,                    // Configured timeout
    "response_size_bytes": len(body),       // Response size
}
```

**File Operations**:
```go
// Use Case: File operation failed
// Keywords: filesystem, io, permissions

Context{
    "path": filePath,                           // Which file
    "operation": "read",                        // What operation
    "size_bytes": fileSize,                    // File size
    "permissions": fileMode.String(),          // File permissions
    "disk_space_available_mb": diskSpace / 1024 / 1024,
}
```

**Business Logic**:
```go
// Use Case: Business operation failed (payment, order, etc.)
// Keywords: business-logic, transaction, state-machine

Context{
    "user_id": userID,                     // Who
    "order_id": orderID,                   // What
    "amount": amount,                      // How much
    "currency": "USD",                     // Currency
    "current_state": "payment_pending",    // State machine
    "previous_state": "cart_confirmed",    // Where we came from
    "step": 3,                            // Which step failed
    "total_steps": 5,                     // Total steps
}
```

**MCP Tool Execution**:
```go
// Use Case: LLM tool execution failed
// Keywords: mcp, llm, tool-server, parameters

Context{
    "tool_name": "search_database",
    "tool_params": params,                     // What params LLM provided
    "expected_params": expectedParams,         // What we expected
    "llm_request_id": requestID,
    "execution_time_ms": duration.Milliseconds(),
}
```

**Microservices / Distributed Tracing**:
```go
// Use Case: Error in microservice chain
// Keywords: microservices, distributed-tracing, service-mesh

Context{
    "service_name": "user-service",
    "correlation_id": traceID,          // Trace through all services
    "request_id": requestID,            // This specific request
    "user_id": userID,                  // Who
    "upstream_service": "api-gateway",  // Where request came from
    "downstream_service": "database",   // Where we were calling
    "hop_count": 3,                    // How many services deep
}
```

---

## MCP Error Code Decision Tree

<!-- RAG: Decision guide for choosing MCP error codes for LLM tool servers -->

### Which MCP error code should I use?

**Why MCP codes matter**:
- **LLM Understanding**: Standard codes that all LLMs recognize
- **Automated Recovery**: LLMs use codes to decide recovery strategy
- **JSON-RPC Compliance**: Follow official JSON-RPC 2.0 specification
- **Tool Server Development**: Build servers that work with Claude and other LLMs

```
START: What went wrong in your MCP tool?
  │
  ├─ JSON parsing failed? → MCPParseError (-32700)
  │   Example: Invalid JSON in request
  │   LLM Action: Fix JSON syntax and retry
  │
  ├─ Request structure invalid? → MCPInvalidRequest (-32600)
  │   Example: Missing "jsonrpc" field, wrong version
  │   LLM Action: Fix request format
  │
  ├─ Tool/method doesn't exist? → MCPMethodNotFound (-32601)
  │   Example: LLM requested "search_web" but only "search_db" exists
  │   LLM Action: Use list_tools to find available tools
  │
  ├─ Parameters are invalid? → MCPInvalidParams (-32602)
  │   Example: "limit" must be number but got string "unlimited"
  │   LLM Action: Read suggestion and fix parameter types/values
  │
  ├─ Tool execution failed (internal error)? → MCPInternalError (-32603)
  │   Example: Unexpected crash, null pointer, panic
  │   LLM Action: Report to developers, don't retry
  │
  ├─ Tool execution failed (expected failure)?
  │   ├─ Resource not available → MCPResourceError (-32001)
  │   │   Example: Database offline, API quota exceeded
  │   │   LLM Action: Wait and retry if retryable
  │   │
  │   ├─ Operation timed out → MCPTimeoutError (-32002)
  │   │   Example: Query took >30s
  │   │   LLM Action: Retry with simpler query or longer timeout
  │   │
  │   ├─ Authentication failed → MCPAuthError (-32003)
  │   │   Example: Invalid API key
  │   │   LLM Action: Prompt user for valid credentials
  │   │
  │   └─ General tool error → MCPToolError (-32000)
  │       Example: Any other expected failure
  │       LLM Action: Read help/suggestion for guidance
  │
  └─ Custom error codes: -32000 to -32099 (Server Error range)
      Use for application-specific errors
```

### MCP Code Examples

```go
// Use Case: LLM requested non-existent tool
// Keywords: mcp, tool-not-found, json-rpc

err := ErrToolNotFound.New().
    WithMCPCode(errific.MCPMethodNotFound).  // -32601
    WithHelp("Tool 'search_web' is not available").
    WithSuggestion("Use 'list_tools' to see available tools")

// Use Case: LLM provided wrong parameter type
// Keywords: mcp, invalid-params, validation

err := ErrInvalidParams.New().
    WithMCPCode(errific.MCPInvalidParams).   // -32602
    WithHelp("Parameter 'limit' must be number 1-100").
    WithSuggestion("Change 'limit' to a number like 10 or 50")

// Use Case: Database unavailable during tool execution
// Keywords: mcp, resource-error, retryable

err := ErrDatabaseUnavailable.New().
    WithMCPCode(errific.MCPResourceError).   // -32001
    WithRetryable(true).
    WithRetryAfter(10 * time.Second).
    WithHelp("Database is temporarily unavailable").
    WithSuggestion("Retry in 10 seconds")
```

---

## Distributed Tracing Decision Tree

<!-- RAG: Decision guide for when to use correlation IDs, request IDs, user IDs, and session IDs -->

### Which tracing ID should I use?

**Why tracing IDs matter**:
- **Distributed Tracing**: Track requests across multiple services
- **Log Aggregation**: Find all logs for a specific request/user/session
- **Root Cause Analysis**: Trace errors back to origin
- **Customer Support**: Search logs by user or session

```
START: What are you trying to track?
  │
  ├─ Track single request across multiple services?
  │   └─→ Use WithCorrelationID()
  │       When: Microservices, service mesh, distributed systems
  │       Example: Gateway → Auth → Database (same correlation ID)
  │       Value: Trace ID from OpenTelemetry, or generated UUID
  │
  ├─ Track individual HTTP request/API call?
  │   └─→ Use WithRequestID()
  │       When: REST APIs, HTTP servers, API gateways
  │       Example: Each HTTP request gets unique ID
  │       Value: X-Request-ID header, or generated UUID
  │
  ├─ Track specific user (authenticated)?
  │   └─→ Use WithUserID()
  │       When: User-facing features, support investigations
  │       Example: Find all errors for user "user-123"
  │       Value: Internal user ID (not email for PII reasons)
  │
  ├─ Track anonymous user session (unauthenticated)?
  │   └─→ Use WithSessionID()
  │       When: Guest checkout, signup flows, session debugging
  │       Example: Track cart errors before user logs in
  │       Value: Session cookie ID
  │
  └─ Track multiple?
      └─→ Use multiple methods together
          Example: .WithCorrelationID(traceID).
                  WithRequestID(requestID).
                  WithUserID(userID)
```

### Tracing ID Comparison

| ID Type | Scope | Lifetime | Use Case |
|---------|-------|----------|----------|
| **CorrelationID** | Multi-service | Entire request chain | Distributed tracing |
| **RequestID** | Single request | One HTTP request | API debugging |
| **UserID** | User-specific | User's lifetime | Support, analytics |
| **SessionID** | Session-specific | Browser session | Guest users, session bugs |

### When to Use Each ID

**Use CorrelationID when**:
- ✅ You have microservices (request touches multiple services)
- ✅ Using OpenTelemetry, Datadog, or distributed tracing
- ✅ Need to trace request from edge to database
- ✅ Debugging multi-service workflows

**Use RequestID when**:
- ✅ Single HTTP request needs tracking
- ✅ API gateway generates request IDs
- ✅ Load balancer correlation needed
- ✅ Idempotency keys (prevent duplicate charges)

**Use UserID when**:
- ✅ User is authenticated
- ✅ Support needs to find user's errors
- ✅ A/B testing or feature flags (track which users affected)
- ✅ Compliance/audit trails

**Use SessionID when**:
- ✅ User is NOT authenticated yet
- ✅ Guest checkout or signup flows
- ✅ Session replay tools (FullStory, LogRocket)
- ✅ Multi-tab debugging

### Practical Examples

**Example 1: E-commerce Checkout**
```go
// Guest user (not logged in) checking out
err := ErrCheckout.New(paymentErr).
    WithSessionID(sessionID).      // Track guest session
    WithRequestID(requestID).      // Track this API call
    WithCorrelationID(traceID)     // Track through services
```

**Example 2: Authenticated API Call**
```go
// Logged-in user making API request
err := ErrAPI.New(apiErr).
    WithUserID(userID).            // Who the user is
    WithRequestID(requestID).      // This specific request
    WithCorrelationID(traceID)     // Trace through microservices
```

**Example 3: Microservice Chain**
```go
// Gateway → User Service → Database
// All share same correlation ID, different request IDs

// Gateway
gatewayErr := ErrGateway.New().
    WithCorrelationID("trace-abc-123").  // Same through all services
    WithRequestID("req-gateway-456")

// User Service (receives correlation ID)
userErr := ErrUserService.New().
    WithCorrelationID("trace-abc-123").  // SAME as gateway
    WithRequestID("req-userservice-789")  // Different request ID

// Database (receives correlation ID)
dbErr := ErrDatabase.New().
    WithCorrelationID("trace-abc-123").  // SAME as gateway & user service
    WithRequestID("req-db-012")           // Different request ID
```

---

## Labels vs Tags vs Context Decision Tree

<!-- RAG: Decision guide for choosing between labels, tags, and context for error metadata -->

### Labels, Tags, or Context - Which should I use?

**Why this matters**:
- **Prometheus Integration**: Labels map to metric labels
- **Semantic Search**: Tags enable AI/RAG search
- **Debugging**: Context provides detailed information
- **Performance**: Labels have cardinality limits

```
START: What kind of metadata do you have?
  │
  ├─ Low-cardinality key-value pairs for metrics?
  │   └─→ Use WithLabel() or WithLabels()
  │       Examples: region, environment, severity, team
  │       Cardinality: <100 unique values
  │       Use for: Prometheus, OpenTelemetry, Datadog metrics
  │
  │       ✅ DO: WithLabel("region", "us-east-1")        // ~20 regions
  │       ✅ DO: WithLabel("severity", "critical")        // 5 severities
  │       ❌ DON'T: WithLabel("user_id", userID)          // Millions of users!
  │
  ├─ Freeform semantic tags for filtering/search?
  │   └─→ Use WithTags()
  │       Examples: "payment", "database", "critical", "external-api"
  │       Use for: Log filtering, semantic search, RAG systems, alert routing
  │
  │       ✅ DO: WithTags("payment", "stripe", "timeout")
  │       ✅ DO: WithTags("database", "postgres", "slow-query")
  │       ❌ DON'T: WithTags(userID)                     // High-cardinality!
  │
  ├─ Detailed debugging information (any value)?
  │   └─→ Use WithContext()
  │       Examples: SQL query, duration, user_id, order details
  │       Use for: Debugging, logging, root cause analysis
  │
  │       ✅ DO: WithContext(Context{
  │               "query": sql,                    // Full SQL query
  │               "duration_ms": 5000,            // Performance data
  │               "user_id": "user-abc-123",      // High-cardinality OK here
  │           })
  │
  └─ Multiple types?
      └─→ Use all three together!
          Example:
          err := ErrDatabase.New(dbErr).
              WithLabels(map[string]string{
                  "region": "us-east-1",           // Prometheus label
                  "severity": "high",              // Alert routing
              }).
              WithTags("database", "postgres", "timeout").  // Searchable tags
              WithContext(Context{
                  "query": sql,                    // Detailed context
                  "duration_ms": 5000,
                  "user_id": userID,
              })
```

### Decision Matrix

| Metadata Type | Use Labels | Use Tags | Use Context |
|---------------|------------|----------|-------------|
| Region (us-east-1, eu-west-1) | ✅ Yes | ✅ Yes | ❌ No |
| Severity (critical, warning) | ✅ Yes | ✅ Yes | ❌ No |
| User ID (millions of users) | ❌ No | ❌ No | ✅ Yes |
| SQL Query (unbounded strings) | ❌ No | ❌ No | ✅ Yes |
| Error Category (7 categories) | ✅ Yes | ❌ No | ❌ No (use WithCategory) |
| Service Name (~10 services) | ✅ Yes | ✅ Yes | ✅ Yes |
| Duration (milliseconds) | ❌ No | ❌ No | ✅ Yes |
| Semantic Keywords | ❌ No | ✅ Yes | ❌ No |

### Cardinality Guidelines

**Low Cardinality** (<100 unique values) → **Labels**
- region, environment, severity, team, service
- Exported to Prometheus metrics

**Medium Cardinality** (100-10,000 values) → **Tags or Context**
- endpoint paths, error codes, tool names
- Use tags for filtering, context for exact values

**High Cardinality** (>10,000 values) → **Context Only**
- user IDs, order IDs, SQL queries, file paths
- Never use as labels (Prometheus explosion!)

---

## Help & Suggestion Decision Tree

<!-- RAG: Decision guide for when to add help text and suggestions to errors -->

### Should I add help text or suggestions?

**Why help/suggestion matters**:
- **LLM Automation**: AI agents read help to understand and suggestions to act
- **Self-Service**: Users fix issues without contacting support
- **Reduced Support Load**: Clear guidance prevents support tickets
- **Faster Resolution**: Immediate recovery guidance reduces downtime

```
START: Who will see this error?
  │
  ├─ LLM/AI agent will handle this error?
  │   └─→ Use WithHelp() + WithSuggestion() + WithDocs()
  │       Example: MCP tool servers, automated systems
  │
  │       WithHelp("The 'limit' parameter must be 1-100, but you provided 'unlimited'")
  │       WithSuggestion("Change 'limit' to a number like 10, 50, or 100")
  │       WithDocs("https://docs.example.com/api/parameters#limit")
  │
  ├─ End user will see this error?
  │   └─→ Use WithHelp() + WithSuggestion()
  │       Example: Web app, mobile app, CLI tool
  │
  │       WithHelp("Your API rate limit of 1000 requests/hour has been exceeded")
  │       WithSuggestion("Wait 30 minutes or upgrade to Pro for 100,000 requests/hour")
  │
  ├─ Developer will debug this error?
  │   └─→ Use WithHelp() (explain what went wrong)
  │       Example: Internal services, background jobs
  │
  │       WithHelp("Database connection pool exhausted (100/100 connections active)")
  │
  ├─ Error message is already clear?
  │   └─→ Skip help/suggestion (avoid redundancy)
  │       Example: "user not found" is self-explanatory
  │
  └─ Complex failure with multiple causes?
      └─→ Use WithHelp() to explain + WithSuggestion() for each cause
          Example: Configuration error

          WithHelp("S3 bucket name 'my bucket' contains spaces, which AWS doesn't allow")
          WithSuggestion("Change bucket name to 'my-bucket' in config.yaml line 23")
          WithDocs("https://docs.aws.amazon.com/s3/bucket-naming-rules")
```

### Help Text Guidelines

**Good Help Text**:
- ✅ Explains WHAT went wrong clearly
- ✅ Explains WHY it's a problem
- ✅ Includes relevant numbers/thresholds
- ✅ Uses plain language (not jargon for user-facing)

**Bad Help Text**:
- ❌ Repeats error message ("An error occurred")
- ❌ Too vague ("Something went wrong")
- ❌ Too technical for users ("ECONNREFUSED on fd 42")
- ❌ Missing context (no numbers, no details)

**Examples**:
```go
// ✅ GOOD
WithHelp("The file size is 50MB, which exceeds the 10MB upload limit")

// ❌ BAD
WithHelp("File too large")  // Redundant with error message

// ✅ GOOD
WithHelp("Database connection pool is full (100/100 connections active). Too many concurrent requests.")

// ❌ BAD
WithHelp("Database error")  // Vague, no details
```

### Suggestion Guidelines

**Good Suggestions**:
- ✅ Specific actionable steps
- ✅ Multiple options when applicable
- ✅ Exact values/examples to use
- ✅ Tells WHEN to retry if relevant

**Bad Suggestions**:
- ❌ Vague ("Fix the configuration")
- ❌ Impossible ("Contact administrator" when no admin)
- ❌ Repeats help text (no new information)
- ❌ Missing timing ("Retry later" - when?)

**Examples**:
```go
// ✅ GOOD
WithSuggestion("Wait 15 minutes until 3:00 PM when rate limit resets, or upgrade to Pro plan")

// ❌ BAD
WithSuggestion("Try again later")  // When is "later"?

// ✅ GOOD
WithSuggestion("Compress the file below 10MB, split into multiple files, or upgrade to Pro for 100MB uploads")

// ❌ BAD
WithSuggestion("Reduce file size")  // How? To what size?

// ✅ GOOD
WithSuggestion("Add port number to connection string in config.yaml line 15. Example: 'postgres://localhost:5432/mydb'")

// ❌ BAD
WithSuggestion("Fix the connection string")  // Where? How?
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
