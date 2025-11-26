# errific Documentation

**Version**: 1.0.0 (Phase 1)
**Go Version**: 1.20+
**Keywords**: error handling, Go errors, structured logging, AI automation, retry logic, error codes, machine-readable errors, error context, error categories, JSON errors

## Documentation Index

### For AI Agents & RAG Systems

This documentation is optimized for Retrieval Augmented Generation (RAG) systems and AI agents. Each document is self-contained with complete context.

---

### 📖 [API Reference](./API_REFERENCE.md)

**Purpose**: Complete API documentation with examples, parameters, return values, and use cases.

**Best for**:
- Looking up specific method signatures
- Understanding parameter meanings
- Finding usage examples
- Extracting metadata from errors

**Key Sections**:
- Core Types (Err, Context, Category)
- Phase 1 Methods (WithContext, WithCode, etc.)
- Helper Functions (GetContext, IsRetryable, etc.)
- JSON Serialization format
- Configuration options
- Common patterns by operation type
- Troubleshooting Q&A

**RAG Query Examples**:
- "How do I add context to an error?"
- "What are the available error categories?"
- "How do I check if an error is retryable?"
- "What format does JSON serialization use?"

---

### 🧭 [Decision Guide](./DECISION_GUIDE.md)

**Purpose**: Decision trees and flowcharts for choosing the right error handling approach.

**Best for**:
- Deciding when to use errific
- Choosing error codes
- Selecting error categories
- Determining retry strategies
- Setting HTTP status codes
- Automated error handling patterns

**Key Sections**:
- When to use errific
- Error code naming conventions
- Category selection decision tree
- Retry decision trees (should retry, how long, how many)
- Context content decision trees
- HTTP status mapping
- Migration guides from other libraries
- AI agent automation patterns

**RAG Query Examples**:
- "Should I retry this network error?"
- "What category should I use for validation errors?"
- "How long should I wait before retrying?"
- "What HTTP status for a timeout error?"
- "What should I include in error context?"

---

## Quick Start by Use Case

### Use Case: API Error Handling

**Goal**: Return proper HTTP errors with structured data

**Documentation Path**:
1. Read [API_REFERENCE.md#HTTP Status](./API_REFERENCE.md#withttpstatus-status-int-errific)
2. Read [DECISION_GUIDE.md#HTTP Status Tree](./DECISION_GUIDE.md#what-http-status-should-i-set)
3. See [API_REFERENCE.md#API Error Pattern](./API_REFERENCE.md#api-error-pattern)

**Quick Answer**:
```go
err := ErrAPI.New(httpErr).
    WithCategory(CategoryTimeout).
    WithHTTPStatus(504).
    WithContext(Context{"endpoint": url})

w.WriteHeader(GetHTTPStatus(err))
json.NewEncoder(w).Encode(err)
```

---

### Use Case: Automated Retry Logic

**Goal**: Let AI agents decide when and how to retry

**Documentation Path**:
1. Read [API_REFERENCE.md#Retry Metadata](./API_REFERENCE.md#withretryable-retryable-bool-errific)
2. Read [DECISION_GUIDE.md#Retry Decision Tree](./DECISION_GUIDE.md#should-this-error-be-retryable)
3. See [DECISION_GUIDE.md#Pattern 1: Automatic Retry](./DECISION_GUIDE.md#pattern-1-automatic-retry)

**Quick Answer**:
```go
err := ErrNetwork.New().
    WithRetryable(true).
    WithRetryAfter(5 * time.Second).
    WithMaxRetries(3)

if IsRetryable(err) {
    time.Sleep(GetRetryAfter(err))
    // retry
}
```

---

### Use Case: Structured Logging

**Goal**: Log errors with rich context for debugging

**Documentation Path**:
1. Read [API_REFERENCE.md#WithContext](./API_REFERENCE.md#withcontext-ctx-context-errific)
2. Read [API_REFERENCE.md#JSON Serialization](./API_REFERENCE.md#json-serialization)
3. See [API_REFERENCE.md#Database Error Pattern](./API_REFERENCE.md#database-error-pattern)

**Quick Answer**:
```go
err := ErrDatabase.New(sqlErr).
    WithCode("DB_001").
    WithContext(Context{
        "query": sql,
        "duration_ms": elapsed,
    })

jsonBytes, _ := json.Marshal(err)
logger.Error(string(jsonBytes))
```

---

### Use Case: Error Monitoring & Alerting

**Goal**: Track errors by code and trigger alerts

**Documentation Path**:
1. Read [API_REFERENCE.md#WithCode](./API_REFERENCE.md#withcode-code-string-errific)
2. Read [DECISION_GUIDE.md#Error Code Decision Tree](./DECISION_GUIDE.md#should-i-add-an-error-code)
3. See [DECISION_GUIDE.md#Pattern 4: Automatic Alerting](./DECISION_GUIDE.md#pattern-4-automatic-alerting)

**Quick Answer**:
```go
err := ErrCritical.New().
    WithCode("SYS_DISK_FULL").
    WithCategory(CategoryServer)

if GetCode(err) == "SYS_DISK_FULL" {
    alertOps(err)
}
```

---

### Use Case: Migration from stdlib errors

**Goal**: Upgrade from stdlib errors to errific

**Documentation Path**:
1. Read [DECISION_GUIDE.md#From stdlib errors](./DECISION_GUIDE.md#from-stdlib-errors)
2. Read [API_REFERENCE.md#Core Types](./API_REFERENCE.md#type-err-string)

**Quick Answer**:
```go
// Before
return errors.New("database failed")

// After
var ErrDatabase Err = "database failed"
return ErrDatabase.New()
```

---

## Semantic Tags for RAG

### Error Handling Concepts
`error-handling`, `error-wrapping`, `error-chaining`, `error-types`, `error-codes`, `error-categories`, `error-context`, `structured-errors`

### Automation & AI
`ai-agents`, `automated-retry`, `retry-logic`, `machine-readable`, `decision-making`, `error-routing`, `self-healing`

### Observability
`structured-logging`, `json-logging`, `error-monitoring`, `error-tracking`, `debugging`, `stack-traces`, `caller-information`

### Web & API
`http-errors`, `api-errors`, `status-codes`, `json-responses`, `rest-api`, `error-responses`

### Go Ecosystem
`golang`, `go-errors`, `errors-package`, `error-interface`, `errors-is`, `errors-as`

### Operations
`retry-strategies`, `exponential-backoff`, `circuit-breaker`, `resilience`, `fault-tolerance`, `error-recovery`

---

## FAQ for RAG Systems

### Q: What is the main purpose of errific?

**A**: errific provides AI-ready error handling for Go with structured context, machine-readable error codes, automated retry metadata, and JSON serialization. It's designed for systems where AI agents or automation needs to make decisions based on errors.

### Q: How is this different from stdlib errors?

**A**: stdlib errors provide basic error wrapping. errific adds:
- Automatic caller information (file:line.function)
- Structured context (Context maps)
- Error codes and categories for classification
- Retry metadata (retryable, retry_after, max_retries)
- HTTP status codes
- JSON serialization
All while maintaining compatibility with stdlib errors.Is() and errors.As().

### Q: Is this thread-safe?

**A**: Yes. Configuration is protected by sync.RWMutex. Error creation and modification are concurrent-safe. All helper functions are safe to call from multiple goroutines.

### Q: What's the performance overhead?

**A**: ~1-2 microseconds per error creation, ~500 bytes memory per error with metadata. Negligible for most applications. See benchmarks in API_REFERENCE.md.

### Q: Can I use this with existing error types?

**A**: Yes. All helper functions (GetCode, GetContext, IsRetryable, etc.) work with any error type. They return zero values if the error is not an errific error.

### Q: How do I migrate from pkg/errors?

**A**: See [DECISION_GUIDE.md#Migration Decision Guide](./DECISION_GUIDE.md#migration-decision-guide) for step-by-step migration patterns.

### Q: What Go version is required?

**A**: Go 1.20+. No external dependencies, stdlib only.

### Q: How do I test errors with errific?

**A**: Use `errors.Is()` for type checking. Use helper functions (GetCode, GetCategory) for metadata assertions:
```go
assert.True(t, errors.Is(err, ErrDatabase))
assert.Equal(t, "DB_001", GetCode(err))
assert.Equal(t, CategoryServer, GetCategory(err))
```

### Q: Can I serialize errors with context that contains custom types?

**A**: Yes, as long as the custom types implement `json.Marshaler` or are JSON-serializable primitive types (string, int, bool, etc.). Avoid channels, functions, and other non-serializable types.

### Q: How do I handle rate limits?

**A**: Use WithRetryAfter() with the delay from the Retry-After header:
```go
err := ErrRateLimit.New().
    WithRetryable(true).
    WithRetryAfter(retryAfter).
    WithHTTPStatus(429)
```

### Q: What if I want to add custom metadata fields beyond Context?

**A**: Context maps support any value (`map[string]any`), so you can add custom fields:
```go
Context{
    "custom_field": customValue,
    "nested": map[string]interface{}{
        "deep": "value",
    },
}
```

---

## Document Metadata

### Last Updated
2024-01-XX

### Target Audience
- AI Agents
- RAG Systems
- Automated Error Handling Systems
- Go Developers
- DevOps Engineers
- SRE Teams

### Related Libraries
- stdlib `errors`
- `github.com/pkg/errors`
- `github.com/cockroachdb/errors`
- `github.com/rotisserie/eris`

### Version History
- **v1.0.0** (Phase 1): Context, Codes, Categories, Retry Metadata, JSON Serialization
- **Future** (Phase 2): HTTP Helpers, OpenTelemetry, Correlation IDs
- **Future** (Phase 3): Metrics, Severity Levels, Error Fingerprinting
