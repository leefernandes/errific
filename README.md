Errific
=========

<img src="./errific.png" width="340" alt="Errific Art"><br>

**AI-Ready Error Handling for Go** with caller metadata, clean error wrapping, structured context, error codes, retry metadata, and JSON serialization.

## ✨ Features

### Core Features
- 📍 **Automatic Caller Information** - File, line, and function automatically captured
- 🔗 **Clean Error Chaining** - Native `errors.Is` and `errors.As` support
- 🏷️ **Error Codes & Categories** - Machine-readable error classification
- 📊 **Structured Context** - Attach metadata for debugging and analytics
- 🔄 **Retry Metadata** - Built-in support for automated retry strategies
- 🌐 **HTTP Status Codes** - Direct mapping to HTTP responses
- 📦 **JSON Serialization** - Seamless integration with logging and APIs

### Phase 2A: MCP & RAG Integration
- 🔗 **MCP Error Format** - JSON-RPC 2.0 compatible error responses for MCP servers
- 🔍 **Correlation Tracking** - Correlation IDs, Request IDs, User IDs, Session IDs
- 💡 **Recovery Guidance** - Help text, suggestions, and documentation links for AI self-healing
- 🏷️ **Semantic Tags** - RAG-optimized tags for error categorization and search
- 📌 **Labels** - Key-value labels for filtering, grouping, and alerting
- ⏰ **Temporal Data** - Timestamps and duration tracking

### Quality
- 🧵 **Thread-Safe** - Concurrent configuration and error creation
- ⚡ **Lightweight** - Small footprint, high performance
- 🎯 **91% Test Coverage** - Comprehensive test suite with 88 test cases

## 🚀 Quick Start

### Basic Usage

```go
var ErrDatabaseQuery errific.Err = "database query failed"

err := ErrDatabaseQuery.New(sqlErr)
fmt.Println(err)
// Output: database query failed [myapp/db.go:42.QueryUsers]
// SQL error details...
```

### AI-Ready Error Handling

```go
var ErrAPITimeout errific.Err = "API request timeout"

err := ErrAPITimeout.New().
    WithCode("API_TIMEOUT_001").
    WithCategory(errific.CategoryTimeout).
    WithContext(errific.Context{
        "endpoint":    "/v1/users",
        "duration_ms": 30000,
        "retry_count": 2,
    }).
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithMaxRetries(3).
    WithHTTPStatus(504)

// AI agent can now automate responses
if errific.IsRetryable(err) {
    time.Sleep(errific.GetRetryAfter(err))
    // retry...
}

// Serialize for logging/monitoring
jsonBytes, _ := json.Marshal(err)
log.Info(string(jsonBytes))
```

### JSON Output

```json
{
  "error": "API request timeout",
  "code": "API_TIMEOUT_001",
  "category": "timeout",
  "caller": "myapp/api.go:123.CallExternalService",
  "context": {
    "endpoint": "/v1/users",
    "duration_ms": 30000,
    "retry_count": 2
  },
  "retryable": true,
  "retry_after": "5s",
  "max_retries": 3,
  "http_status": 504
}
```

### MCP Server Integration (Phase 2A)

**Scenario**: Your AI tool fails during execution and needs to return a proper MCP error response.

```go
var ErrToolExecution errific.Err = "search_database tool failed"

// Create rich error with MCP metadata
err := ErrToolExecution.New(dbErr).
    WithMCPCode(errific.MCPToolError).              // JSON-RPC 2.0 error code
    WithCorrelationID("trace-abc-123").              // Track across distributed calls
    WithRequestID("req-456").                        // Individual request tracking
    WithHelp("Database connection pool exhausted").  // Human-readable help
    WithSuggestion("Increase pool size to 50").      // Actionable recovery step
    WithDocs("https://docs.ai/errors/db-pool").     // Documentation link
    WithTags("database", "connection-pool", "retryable"). // RAG semantic tags
    WithLabel("tool_name", "search_database").       // Filter/group by tool
    WithRetryable(true).
    WithRetryAfter(5 * time.Second)

// Convert to MCP JSON-RPC 2.0 format
mcpErr := errific.ToMCPError(err)
json.NewEncoder(w).Encode(mcpErr)
```

**MCP Response**:
```json
{
  "code": -32000,
  "message": "search_database tool failed",
  "data": {
    "error": "search_database tool failed",
    "code": "TOOL_001",
    "correlation_id": "trace-abc-123",
    "request_id": "req-456",
    "help": "Database connection pool exhausted",
    "suggestion": "Increase pool size to 50",
    "docs": "https://docs.ai/errors/db-pool",
    "tags": ["database", "connection-pool", "retryable"],
    "labels": {"tool_name": "search_database"},
    "retryable": true,
    "retry_after": "5s"
  }
}
```

**Why This Matters**:
- 🤖 AI agents can **self-heal** using help/suggestion fields
- 🔍 **Correlation tracking** across distributed MCP tool calls
- 📊 **RAG systems** can categorize and search errors by semantic tags
- 🎯 **Monitoring systems** can alert based on labels
- 🔄 **Automatic retry** logic from metadata

## 📖 Documentation

### Error Categories

```go
CategoryClient       // 4xx - client errors
CategoryServer       // 5xx - server errors
CategoryNetwork      // connectivity issues
CategoryValidation   // input validation
CategoryNotFound     // 404 errors
CategoryUnauthorized // 401/403 errors
CategoryTimeout      // timeout errors
```

### Key Methods

```go
// Structured context
.WithContext(Context{"key": "value"})

// Machine-readable codes
.WithCode("ERR_001")
.WithCategory(CategoryServer)

// Retry automation
.WithRetryable(true)
.WithRetryAfter(5 * time.Second)
.WithMaxRetries(3)

// HTTP integration
.WithHTTPStatus(503)

// Extract metadata
GetCode(err)        // → "ERR_001"
GetCategory(err)    // → CategoryServer
IsRetryable(err)    // → true
GetHTTPStatus(err)  // → 503
GetContext(err)     // → Context map
```

## 🎯 Use Cases

- **API Services** - Automatic HTTP status code mapping and JSON responses
- **Microservices** - Structured logging with correlation IDs and context
- **Retry Logic** - Built-in retry metadata for resilience patterns
- **AI Agents** - Machine-readable error codes and categories for automation
- **Monitoring** - JSON serialization for Datadog, ELK, Prometheus
- **Debugging** - Automatic caller information and stack traces

## 📊 More Examples

Check out the [comprehensive examples](https://github.com/leefernandes/errific/tree/main/examples) including:
- Context attachment
- Error codes and categories
- Retry metadata
- JSON serialization
- AI agent scenarios
- HTTP integration

Try it on the <a href="https://go.dev/play/p/N7asgc_1i-J"><img src="./gopher.png" height="14px" /></a> [playground](https://go.dev/play/p/N7asgc_1i-J)!

## 📚 RAG-Optimized Documentation

For AI agents and RAG systems, comprehensive documentation is available:

- **[API Reference](./docs/API_REFERENCE.md)** - Complete API documentation with examples, decision trees, and troubleshooting
- **[Decision Guide](./docs/DECISION_GUIDE.md)** - When to use each feature, error handling patterns, and automation guides
- **[Docs Index](./docs/README.md)** - Documentation overview with semantic tags and FAQ

Each document is self-contained with full context for RAG retrieval.

## 🤖 AI Agent Integration

errific is designed for AI-driven automation:

```go
// AI agents can automatically decide:
if IsRetryable(err) {
    delay := GetRetryAfter(err)      // How long to wait
    maxRetries := GetMaxRetries(err)  // How many attempts
    // ... implement retry logic
}

switch GetCategory(err) {
case CategoryValidation:
    // Return 400 to user
case CategoryNetwork:
    // Retry with backoff
case CategoryServer:
    // Alert ops team
}

// Serialize for monitoring/logging
jsonBytes, _ := json.Marshal(err)
sendToDatadog(jsonBytes)
```

## 📊 Coverage & Quality

- 90.4% test coverage
- Thread-safe (race detector clean)
- Zero external dependencies
- Comprehensive examples and documentation
