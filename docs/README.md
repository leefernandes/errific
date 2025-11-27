# errific Documentation Navigator

<!-- RAG: Documentation index and navigation guide for finding the right errific documentation -->

**Version**: 1.0.0 (Phase 1 & 2A Complete)
**Go Version**: 1.20+
**Keywords**: error handling, Go errors, structured logging, AI automation, retry logic, error codes, machine-readable errors, error context, error categories, JSON errors, MCP, distributed tracing

---

## 🎯 Choose Your Path

<!-- RAG: Quick navigation flowchart to help users find relevant documentation based on their needs -->

**New to errific?** → Start with [Quick Start Examples](#quick-start-by-use-case) below

**Looking for a specific method?** → See [API Reference](./API_REFERENCE.md)

**Need to make a decision?** → See [Decision Guide](./DECISION_GUIDE.md)

**Building for AI/LLMs?** → See [MCP & AI Integration](#use-case-6-mcp-tool-server-for-llms)

**Debugging an issue?** → See [Troubleshooting](./API_REFERENCE.md#troubleshooting)

**Migrating from another library?** → See [Migration Guide](./DECISION_GUIDE.md#migration-decision-guide)

---

### Visual Decision Tree

```
What do you want to do?
│
├─ 🆕 Learn errific basics
│   └─→ Start here: README.md (root) Quick Start
│       Then: Use Cases 1-5 below
│
├─ 🔍 Find a specific method/function
│   └─→ Go to: API_REFERENCE.md
│       Search for: .WithXXX() or GetXXX()
│
├─ 🤔 Make a decision
│   ├─ Should I retry this error?
│   │   └─→ DECISION_GUIDE.md > Retry Decision Tree
│   │
│   ├─ What category should I use?
│   │   └─→ DECISION_GUIDE.md > Category Decision Tree
│   │
│   ├─ What HTTP status should I return?
│   │   └─→ DECISION_GUIDE.md > HTTP Status Decision Tree
│   │
│   └─ What should I include in context?
│       └─→ DECISION_GUIDE.md > Context Decision Tree
│
├─ 🤖 Build AI/LLM integration
│   ├─ MCP Tool Server
│   │   └─→ Use Case 6 below
│   │       Then: API_REFERENCE.md > WithMCPCode()
│   │
│   ├─ Automated Retry Logic
│   │   └─→ Use Case 2 below
│   │       Then: API_REFERENCE.md > WithRetryable()
│   │
│   └─ AI-Readable Error Messages
│       └─→ Use Case 7 below
│           Then: API_REFERENCE.md > WithHelp(), WithSuggestion()
│
├─ 📊 Add monitoring/observability
│   ├─ Distributed Tracing
│   │   └─→ Use Case 8 below
│   │       Then: API_REFERENCE.md > WithCorrelationID()
│   │
│   ├─ Prometheus Metrics
│   │   └─→ Use Case 9 below
│   │       Then: API_REFERENCE.md > WithLabels()
│   │
│   └─ Performance Tracking
│       └─→ Use Case 10 below
│           Then: API_REFERENCE.md > WithDuration()
│
└─ 🔧 Debug or troubleshoot
    └─→ API_REFERENCE.md > Troubleshooting section
        Then: DECISION_GUIDE.md for decision help
```

---

## 📚 Documentation Files

<!-- RAG: Overview of all documentation files and their purposes -->

### 1. [API Reference](./API_REFERENCE.md) - "The Manual"

**Purpose**: Complete API documentation with exhaustive examples and method signatures.

**What's inside**:
- ✅ Core Types (Err, Context, Category)
- ✅ Phase 1 Methods (WithContext, WithCode, WithCategory, WithRetryable, WithRetryAfter, WithMaxRetries, WithHTTPStatus)
- ✅ Phase 2A Methods (WithMCPCode, WithCorrelationID, WithRequestID, WithUserID, WithSessionID, WithHelp, WithSuggestion, WithDocs, WithTags, WithLabel, WithLabels, WithTimestamp, WithDuration)
- ✅ Helper Functions (GetContext, IsRetryable, GetCode, etc.)
- ✅ JSON Serialization format
- ✅ Configuration options
- ✅ Complete code examples (10+ full examples)
- ✅ Troubleshooting Q&A

**Best for**:
- Looking up specific method signatures
- Understanding parameters and return values
- Finding detailed code examples
- Learning about all available methods

**When to read**:
- You know what method you want but need to see how to use it
- You want exhaustive examples with all options
- You're implementing a feature and need exact syntax

**File size**: ~4,400 lines (comprehensive reference)

---

### 2. [Decision Guide](./DECISION_GUIDE.md) - "The Helper"

**Purpose**: Decision trees and flowcharts for choosing the right approach.

**What's inside**:
- ✅ When to use errific vs stdlib errors
- ✅ Error code naming conventions
- ✅ Category selection decision tree
- ✅ Retry decision trees (should retry? how long? how many?)
- ✅ Context content decision trees
- ✅ HTTP status mapping guide
- ✅ Migration guides from other libraries
- ✅ AI agent automation patterns

**Best for**:
- Making decisions about error handling
- Choosing between options (retry vs not, category selection)
- Understanding trade-offs
- Migration from other libraries

**When to read**:
- You're unsure which approach to use
- You need to decide retry strategy
- You want to know best practices
- You're migrating from pkg/errors or stdlib

**File size**: ~800 lines (decision-focused)

---

### 3. Root README.md - "The Hook"

**Purpose**: Get users excited and started quickly (5-minute quickstart).

**What's inside**:
- Installation instructions
- 5 real-world scenarios with code
- Quick reference table
- Feature highlights

**Best for**:
- First-time users
- Quick evaluation of errific
- Copy-paste examples to get started

**When to read**:
- You're evaluating errific for your project
- You want to see what errific can do quickly
- You need a simple example to start with

---

## 🚀 Quick Start by Use Case

<!-- RAG: Use-case-driven documentation paths with code examples and links to detailed docs -->

Below are common use cases with **quick examples** and **documentation paths** for deeper learning.

---

### Use Case 1: REST API Error Handling

<!-- RAG: Handle REST API errors with proper HTTP status codes and JSON responses -->

**Goal**: Return proper HTTP errors with structured JSON and correct status codes

**Scenario**: You're building a REST API and need to return errors that clients can parse

**Quick Example**:
```go
// Use Case: API endpoint returns 404 for missing user
// Keywords: rest-api, http-status, json-response, not-found

import (
    "net/http"
    "encoding/json"
    "github.com/leefernandes/errific"
)

var ErrUserNotFound errific.Err = "user not found"

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    user, err := database.GetUser(userID)
    if err != nil {
        apiErr := ErrUserNotFound.New(err).
            WithCategory(errific.CategoryNotFound).
            WithHTTPStatus(404).
            WithContext(errific.Context{
                "user_id": userID,
                "endpoint": r.URL.Path,
            })

        // Automatic HTTP status and JSON response
        w.WriteHeader(errific.GetHTTPStatus(apiErr))  // 404
        json.NewEncoder(w).Encode(apiErr)
        return
    }

    json.NewEncoder(w).Encode(user)
}
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → Scenario 2: HTTP API Errors](../README.md#scenario-2-http-api-errors-with-status-codes)
2. 📖 **Complete reference**: [API_REFERENCE.md → WithHTTPStatus()](./API_REFERENCE.md#withhttpstatus-status-int-errific)
3. 🧭 **Decision help**: [DECISION_GUIDE.md → HTTP Status Decision Tree](./DECISION_GUIDE.md#what-http-status-should-i-set)
4. 📖 **Full example**: [API_REFERENCE.md → Complete Examples → API Server](./API_REFERENCE.md#complete-examples)

**What you'll learn**:
- How to map errors to HTTP status codes
- How to return JSON error responses
- How to add request context to errors
- Best practices for API error handling

---

### Use Case 2: Automated Retry Logic

<!-- RAG: Enable AI agents and automation to make intelligent retry decisions -->

**Goal**: Let AI agents or automation decide when and how to retry operations

**Scenario**: Network calls fail sometimes; you want automatic retry with exponential backoff

**Quick Example**:
```go
// Use Case: Retry network call with exponential backoff
// Keywords: retry-logic, exponential-backoff, automated-retry, network-resilience

var ErrAPITimeout errific.Err = "API call timed out"

func callExternalAPI(endpoint string) error {
    err := makeHTTPCall(endpoint)
    if err != nil {
        return ErrAPITimeout.New(err).
            WithRetryable(true).                    // Can be retried
            WithRetryAfter(5 * time.Second).       // Wait 5s before retry
            WithMaxRetries(3).                     // Try max 3 times
            WithContext(errific.Context{
                "endpoint": endpoint,
                "timeout_ms": 30000,
            })
    }
    return nil
}

// AI Agent automatically retries:
func retryableOperation() error {
    var err error
    for attempt := 0; attempt < 5; attempt++ {
        err = callExternalAPI("https://api.example.com")

        // AI reads retry metadata
        if err == nil || !errific.IsRetryable(err) {
            return err  // Success or non-retryable
        }

        delay := errific.GetRetryAfter(err)
        maxRetries := errific.GetMaxRetries(err)

        if attempt >= maxRetries {
            return err  // Exceeded max retries
        }

        time.Sleep(delay)  // Wait before retry
    }
    return err
}
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → Scenario 3: Automated Retry](../README.md#scenario-3-automated-retry-logic)
2. 📖 **Complete reference**: [API_REFERENCE.md → WithRetryable()](./API_REFERENCE.md#withretryable-retryable-bool-errific)
3. 🧭 **Decision help**: [DECISION_GUIDE.md → Retry Decision Tree](./DECISION_GUIDE.md#should-this-error-be-retryable)
4. 📖 **Full example**: [API_REFERENCE.md → Retry with Exponential Backoff](./API_REFERENCE.md#withretryafter-delay-timeduration-errific)

**What you'll learn**:
- How to mark errors as retryable or non-retryable
- How to set retry delays (exponential backoff)
- How to limit retry attempts
- Best practices for retry decision-making

---

### Use Case 3: Structured Logging & Debugging

<!-- RAG: Add rich context to errors for debugging and log aggregation -->

**Goal**: Log errors with rich context for debugging and monitoring

**Scenario**: You need detailed error logs with query parameters, timing, and state information

**Quick Example**:
```go
// Use Case: Database query error with full debugging context
// Keywords: structured-logging, debugging, context, json-logging

var ErrDatabaseQuery errific.Err = "database query failed"

func queryUsers(sql string) ([]User, error) {
    start := time.Now()

    rows, err := db.Query(sql)
    duration := time.Since(start)

    if err != nil {
        dbErr := ErrDatabaseQuery.New(err).
            WithCode("DB_QUERY_001").
            WithCategory(errific.CategoryServer).
            WithDuration(duration).  // Track how long it took
            WithContext(errific.Context{
                "query": sql,
                "duration_ms": duration.Milliseconds(),
                "table": "users",
                "connection_pool_size": db.Stats().OpenConnections,
            })

        // JSON logging (works with any logger)
        jsonBytes, _ := json.Marshal(dbErr)
        logger.Error(string(jsonBytes))
        /*
        Logs:
        {
          "error": "database query failed: connection timeout",
          "code": "DB_QUERY_001",
          "category": "server",
          "duration": "5s",
          "context": {
            "query": "SELECT * FROM users WHERE age > 30",
            "duration_ms": 5000,
            "table": "users",
            "connection_pool_size": 10
          },
          "caller": "database.go:45.queryUsers"
        }
        */

        return nil, dbErr
    }

    return scanUsers(rows)
}
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → Scenario 1: Database with Context](../README.md#scenario-1-database-errors-with-context)
2. 📖 **Complete reference**: [API_REFERENCE.md → WithContext()](./API_REFERENCE.md#withcontext-ctx-context-errific)
3. 🧭 **Decision help**: [DECISION_GUIDE.md → Context Decision Tree](./DECISION_GUIDE.md#what-should-i-include-in-context)
4. 📖 **JSON format**: [API_REFERENCE.md → JSON Serialization](./API_REFERENCE.md#json-serialization)

**What you'll learn**:
- How to add structured context to errors
- What to include in error context
- How to serialize errors to JSON
- Best practices for debugging with context

---

### Use Case 4: Error Monitoring & Alerting

<!-- RAG: Track errors by code and category for monitoring dashboards and alerts -->

**Goal**: Track errors by code/category and trigger alerts for critical issues

**Scenario**: You want to monitor error rates and alert on-call engineers for critical errors

**Quick Example**:
```go
// Use Case: Critical system error with alerting
// Keywords: monitoring, alerting, error-codes, critical-errors

var ErrDiskFull errific.Err = "disk space exhausted"

func writeToFile(data []byte) error {
    err := os.WriteFile(filename, data, 0644)
    if err != nil {
        // Check if disk full
        if strings.Contains(err.Error(), "no space left") {
            criticalErr := ErrDiskFull.New(err).
                WithCode("SYS_DISK_FULL").
                WithCategory(errific.CategoryServer).
                WithContext(errific.Context{
                    "disk_path": "/var/data",
                    "file_size_mb": len(data) / 1024 / 1024,
                })

            // Alert monitoring system
            if errific.GetCode(criticalErr) == "SYS_DISK_FULL" {
                monitoring.Alert("Critical: Disk Full", criticalErr)
                pagerduty.NotifyOnCall(criticalErr)
            }

            return criticalErr
        }

        // Non-critical file error
        return errific.Err("file write failed").New(err).
            WithCode("FILE_WRITE_001").
            WithContext(errific.Context{"filename": filename})
    }

    return nil
}

// Monitoring dashboard queries:
// - Count errors by code: GROUP BY code
// - Alert on SYS_DISK_FULL
// - Track error rate trends
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → Scenario 4: Monitoring](../README.md#scenario-4-error-monitoring--alerting)
2. 📖 **Complete reference**: [API_REFERENCE.md → WithCode()](./API_REFERENCE.md#withcode-code-string-errific)
3. 🧭 **Decision help**: [DECISION_GUIDE.md → Error Code Decision](./DECISION_GUIDE.md#should-i-add-an-error-code)
4. 📖 **Full example**: [API_REFERENCE.md → Monitoring Example](./API_REFERENCE.md#withcode-code-string-errific)

**What you'll learn**:
- How to assign error codes
- How to categorize errors for routing
- How to integrate with monitoring systems
- Best practices for alerting

---

### Use Case 5: Error Categories & Routing

<!-- RAG: Classify errors by category for automatic routing and handling decisions -->

**Goal**: Classify errors into categories for automatic routing and handling

**Scenario**: Different error types need different handling (retry vs fail, 4xx vs 5xx)

**Quick Example**:
```go
// Use Case: Route errors to different handlers based on category
// Keywords: error-categories, routing, classification, error-handling

var (
    ErrInvalidEmail errific.Err = "invalid email format"
    ErrUserExists   errific.Err = "user already exists"
    ErrDatabase     errific.Err = "database error"
    ErrNetwork      errific.Err = "network timeout"
)

func createUser(email string) error {
    // Validation error (client's fault)
    if !isValidEmail(email) {
        return ErrInvalidEmail.New().
            WithCategory(errific.CategoryValidation).  // Client error
            WithHTTPStatus(400).
            WithRetryable(false)  // Don't retry validation errors
    }

    // Check if exists (conflict)
    exists, err := database.UserExists(email)
    if err != nil {
        // Database error (server's fault)
        return ErrDatabase.New(err).
            WithCategory(errific.CategoryServer).  // Server error
            WithHTTPStatus(500).
            WithRetryable(true)  // Can retry server errors
    }

    if exists {
        return ErrUserExists.New().
            WithCategory(errific.CategoryClient).  // Client error (duplicate)
            WithHTTPStatus(409)
    }

    // Create user...
    return nil
}

// Router uses categories:
func errorHandler(err error) int {
    switch errific.GetCategory(err) {
    case errific.CategoryValidation:
        return 400  // Bad Request
    case errific.CategoryNotFound:
        return 404  // Not Found
    case errific.CategoryNetwork:
        return 503  // Service Unavailable (can retry)
    case errific.CategoryServer:
        return 500  // Internal Server Error
    default:
        return 500
    }
}
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → Categories](../README.md)
2. 📖 **Complete reference**: [API_REFERENCE.md → type Category](./API_REFERENCE.md#type-category-string)
3. 🧭 **Decision help**: [DECISION_GUIDE.md → Category Decision Tree](./DECISION_GUIDE.md#what-category-should-i-use)
4. 📖 **Full example**: [API_REFERENCE.md → Category Examples](./API_REFERENCE.md#type-category-string)

**What you'll learn**:
- Available error categories (7 categories)
- When to use each category
- How to map categories to HTTP status codes
- How to route errors based on category

---

### Use Case 6: MCP Tool Server for LLMs

<!-- RAG: Build MCP tool servers that communicate errors clearly to Claude and other LLMs -->

**Goal**: Build MCP tool servers with errors that LLMs can understand and act on

**Scenario**: You're building tools for Claude/LLMs and need structured error responses

**Quick Example**:
```go
// Use Case: MCP tool with LLM-friendly error messages
// Keywords: mcp, llm-integration, tool-server, json-rpc, ai-agents

import "github.com/leefernandes/errific"

var (
    ErrToolNotFound  errific.Err = "tool not found"
    ErrInvalidParams errific.Err = "invalid tool parameters"
)

// MCP Request Handler
func handleMCPRequest(req *MCPRequest) *MCPResponse {
    // Tool doesn't exist
    if !toolRegistry.Has(req.Method) {
        err := ErrToolNotFound.New().
            WithMCPCode(errific.MCPMethodNotFound).  // -32601
            WithHelp("The requested tool is not available on this server.").
            WithSuggestion("Use the 'list_tools' method to see available tools.").
            WithDocs("https://docs.example.com/mcp/tools").
            WithTags("mcp", "tool-not-found", "validation")

        return &MCPResponse{
            JSONRPC: "2.0",
            ID:      req.ID,
            Error:   errific.ToMCPError(err),
        }
    }

    // Invalid parameters
    if err := validateParams(req.Method, req.Params); err != nil {
        paramErr := ErrInvalidParams.New(err).
            WithMCPCode(errific.MCPInvalidParams).  // -32602
            WithHelp("The 'limit' parameter must be a number between 1 and 100.").
            WithSuggestion("Change the 'limit' parameter to a number like 10 or 50.").
            WithDocs("https://docs.example.com/tools/" + req.Method).
            WithContext(errific.Context{
                "tool": req.Method,
                "provided_params": req.Params,
                "expected_params": getExpectedParams(req.Method),
            })

        return &MCPResponse{
            JSONRPC: "2.0",
            ID:      req.ID,
            Error:   errific.ToMCPError(paramErr),
        }
    }

    // Execute tool...
    result, _ := toolRegistry.Execute(req.Method, req.Params)
    return &MCPResponse{JSONRPC: "2.0", ID: req.ID, Result: result}
}

/*
LLM receives:
{
  "jsonrpc": "2.0",
  "id": "req-123",
  "error": {
    "code": -32602,
    "message": "invalid tool parameters",
    "data": {
      "help": "The 'limit' parameter must be a number between 1 and 100.",
      "suggestion": "Change the 'limit' parameter to a number like 10 or 50.",
      "docs": "https://docs.example.com/tools/search"
    }
  }
}

LLM reads "suggestion" and fixes the request automatically!
*/
```

**Documentation Path**:
1. 📖 **Simple intro**: [Root README.md → MCP Integration](../README.md)
2. 📖 **Complete reference**: [API_REFERENCE.md → WithMCPCode()](./API_REFERENCE.md#withmcpcode-code-int-errific)
3. 📖 **Help/Suggestion**: [API_REFERENCE.md → WithHelp()](./API_REFERENCE.md#withhelp-message-string-errific)
4. 📖 **Full example**: [API_REFERENCE.md → MCP Tool Server Example](./API_REFERENCE.md#complete-examples)

**What you'll learn**:
- How to use MCP error codes (JSON-RPC 2.0)
- How to add help text for LLMs
- How to suggest recovery actions
- Best practices for LLM-friendly errors

---

### Use Case 7: AI-Readable Error Messages

<!-- RAG: Add human-readable help and actionable suggestions for AI agents and users -->

**Goal**: Add help text and suggestions that AI agents can read and act on

**Scenario**: You want errors that explain what went wrong and how to fix it

**Quick Example**:
```go
// Use Case: Self-documenting errors with help and suggestions
// Keywords: ai-readable, help-text, error-recovery, self-service

var ErrRateLimit errific.Err = "API rate limit exceeded"

func apiHandler(w http.ResponseWriter, r *http.Request) {
    if rateLimiter.IsExceeded(userID) {
        resetTime := rateLimiter.GetResetTime(userID)
        retryAfter := time.Until(resetTime)

        err := ErrRateLimit.New().
            WithHTTPStatus(429).
            WithRetryable(true).
            WithRetryAfter(retryAfter).
            WithHelp("You've made 1000 API requests in the last hour, exceeding your rate limit of 1000/hour.").
            WithSuggestion("Wait 15 minutes until 3:00 PM when your rate limit resets, or upgrade to Pro plan for 100,000 requests/hour.").
            WithDocs("https://docs.example.com/api/rate-limits").
            WithContext(errific.Context{
                "rate_limit": 1000,
                "requests_made": 1000,
                "reset_time": resetTime.Format(time.RFC3339),
                "current_tier": "free",
            })

        w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
        w.WriteHeader(429)
        json.NewEncoder(w).Encode(err)
        return
    }

    // Process request...
}

/*
User/AI sees:
{
  "error": "API rate limit exceeded",
  "help": "You've made 1000 API requests in the last hour...",
  "suggestion": "Wait 15 minutes until 3:00 PM...",
  "docs": "https://docs.example.com/api/rate-limits",
  "retryable": true,
  "retry_after": "15m"
}

AI agent:
1. Reads "help" → Understands the problem
2. Reads "suggestion" → Knows to wait 15 minutes
3. Reads "retry_after" → Waits exactly 15 minutes
4. Retries automatically
*/
```

**Documentation Path**:
1. 📖 **Complete reference**: [API_REFERENCE.md → WithHelp()](./API_REFERENCE.md#withhelp-message-string-errific)
2. 📖 **Suggestion reference**: [API_REFERENCE.md → WithSuggestion()](./API_REFERENCE.md#withsuggestion-message-string-errific)
3. 📖 **Docs reference**: [API_REFERENCE.md → WithDocs()](./API_REFERENCE.md#withdocs-url-string-errific)

**What you'll learn**:
- How to write helpful error messages
- How to suggest actionable recovery steps
- How to link to documentation
- Best practices for AI-readable errors

---

### Use Case 8: Distributed Tracing & Correlation

<!-- RAG: Track errors across microservices using correlation IDs and trace IDs -->

**Goal**: Track errors across multiple microservices with correlation/trace IDs

**Scenario**: Your system has multiple services; you need to trace errors through the entire chain

**Quick Example**:
```go
// Use Case: Trace error through Gateway → User Service → Database
// Keywords: distributed-tracing, microservices, correlation-id, opentelemetry

import (
    "github.com/google/uuid"
    "go.opentelemetry.io/otel/trace"
)

// Gateway Service
func gatewayHandler(w http.ResponseWriter, r *http.Request) {
    // Get or generate correlation ID
    correlationID := r.Header.Get("X-Correlation-ID")
    if correlationID == "" {
        correlationID = uuid.New().String()
    }

    requestID := r.Header.Get("X-Request-ID")

    // Call User Service
    user, err := userService.GetUser(ctx, userID)
    if err != nil {
        gatewayErr := errific.Err("user service failed").New(err).
            WithCorrelationID(correlationID).  // Trace through services
            WithRequestID(requestID).           // Track this specific request
            WithHTTPStatus(503).
            WithContext(errific.Context{
                "service": "user-service",
                "user_id": userID,
                "gateway": "api-gw-01",
            })

        w.WriteHeader(503)
        json.NewEncoder(w).Encode(gatewayErr)
        return
    }

    json.NewEncoder(w).Encode(user)
}

// User Service
func (s *UserService) GetUser(ctx context.Context, userID string) (*User, error) {
    correlationID := ctx.Value("correlation_id").(string)

    user, err := database.QueryUser(userID)
    if err != nil {
        return nil, errific.Err("database query failed").New(err).
            WithCorrelationID(correlationID).  // Same ID through chain
            WithContext(errific.Context{
                "service": "database",
                "query": "SELECT * FROM users WHERE id = $1",
                "user_id": userID,
            })
    }

    return user, nil
}

/*
Later: Search logs for correlation_id="abc-123"
Finds:
1. Gateway: 10:00:00.100 - Received request
2. User Service: 10:00:00.150 - Querying database
3. Database: 10:00:00.200 - ERROR: connection timeout

Full trace of error through all services!
*/
```

**Documentation Path**:
1. 📖 **Complete reference**: [API_REFERENCE.md → WithCorrelationID()](./API_REFERENCE.md#withcorrelationid-id-string-errific)
2. 📖 **Request ID**: [API_REFERENCE.md → WithRequestID()](./API_REFERENCE.md#withrequestid-id-string-errific)
3. 📖 **Full example**: [API_REFERENCE.md → Distributed Microservices Example](./API_REFERENCE.md#complete-examples)

**What you'll learn**:
- How to use correlation IDs for tracing
- How to track individual requests
- How to integrate with OpenTelemetry
- Best practices for distributed systems

---

### Use Case 9: Prometheus Metrics & Labels

<!-- RAG: Export error metrics to Prometheus with labels for dashboards and alerting -->

**Goal**: Export error metrics to Prometheus with labels for filtering and aggregation

**Scenario**: You want error dashboards in Grafana showing errors by endpoint, region, status

**Quick Example**:
```go
// Use Case: Error metrics with Prometheus labels
// Keywords: prometheus, metrics, monitoring, labels, grafana

import "github.com/prometheus/client_golang/prometheus"

var errorCounter = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "api_errors_total",
        Help: "Total number of API errors",
    },
    []string{"endpoint", "method", "status", "region"},
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
    err := processRequest(r)
    if err != nil {
        apiErr := errific.Err("API request failed").New(err).
            WithLabel("endpoint", r.URL.Path).
            WithLabel("method", r.Method).
            WithLabel("status", "500").
            WithLabel("region", "us-west-2").
            WithHTTPStatus(500)

        // Export to Prometheus
        labels := errific.GetLabels(apiErr)
        errorCounter.WithLabelValues(
            labels["endpoint"],
            labels["method"],
            labels["status"],
            labels["region"],
        ).Inc()

        w.WriteHeader(500)
        json.NewEncoder(w).Encode(apiErr)
        return
    }

    // Success...
}

/*
Prometheus metrics:
api_errors_total{endpoint="/api/users",method="GET",status="500",region="us-west-2"} 42

Grafana queries:
- rate(api_errors_total[5m])  → Errors per second
- sum by (endpoint) (api_errors_total)  → Errors by endpoint
- api_errors_total{region="us-west-2"}  → Errors in specific region
*/
```

**Documentation Path**:
1. 📖 **Complete reference**: [API_REFERENCE.md → WithLabel()](./API_REFERENCE.md#withlabel-key-value-string-errific)
2. 📖 **Batch labels**: [API_REFERENCE.md → WithLabels()](./API_REFERENCE.md#withlabels-labels-mapstringstring-errific)
3. 📖 **Tags**: [API_REFERENCE.md → WithTags()](./API_REFERENCE.md#withtags-tags-string-errific)

**What you'll learn**:
- How to add labels to errors
- How to export metrics to Prometheus
- How to use labels for filtering
- Best practices for cardinality control

---

### Use Case 10: Performance Tracking & SLA Monitoring

<!-- RAG: Track operation duration and monitor SLA violations for performance analysis -->

**Goal**: Track operation duration and identify slow operations that fail

**Scenario**: You need to monitor query performance and alert on SLA violations

**Quick Example**:
```go
// Use Case: Track database query performance and SLA violations
// Keywords: performance, sla-monitoring, duration-tracking, slow-queries

var ErrSlowQuery errific.Err = "database query exceeded SLA"

func queryWithSLA(ctx context.Context, sql string) ([]Row, error) {
    start := time.Now()
    slaThreshold := 1 * time.Second  // Queries must complete in 1s

    rows, err := db.QueryContext(ctx, sql)
    duration := time.Since(start)

    if err != nil {
        slaViolation := duration > slaThreshold

        queryErr := ErrSlowQuery.New(err).
            WithDuration(duration).  // Track how long it took
            WithContext(errific.Context{
                "query": sql,
                "duration_ms": duration.Milliseconds(),
                "sla_threshold_ms": slaThreshold.Milliseconds(),
                "sla_violation": slaViolation,
                "sla_percentage": (duration.Seconds() / slaThreshold.Seconds()) * 100,
            })

        // Alert if SLA violated
        if slaViolation {
            monitoring.Alert("Query SLA violated", queryErr)
        }

        return nil, queryErr
    }

    // Success but slow (warning)
    if duration > slaThreshold {
        log.Warn("Slow query",
            "duration_ms", duration.Milliseconds(),
            "query", sql,
            "sla_ms", slaThreshold.Milliseconds(),
        )
    }

    return scanRows(rows)
}

/*
Analysis:
- Failed queries: 1500ms average duration
- Successful queries: 200ms average duration
- Insight: Errors are taking 7.5x longer! (timeout issue)
*/
```

**Documentation Path**:
1. 📖 **Complete reference**: [API_REFERENCE.md → WithDuration()](./API_REFERENCE.md#withduration-d-timeduration-errific)
2. 📖 **Timestamp**: [API_REFERENCE.md → WithTimestamp()](./API_REFERENCE.md#withtimestamp-t-timetime-errific)
3. 📖 **Full example**: [API_REFERENCE.md → Performance Comparison Example](./API_REFERENCE.md#withduration-d-timeduration-errific)

**What you'll learn**:
- How to track operation duration
- How to monitor SLA compliance
- How to identify slow operations
- Best practices for performance tracking

---

## 🏷️ Semantic Tags for RAG

<!-- RAG: Keyword tags for semantic search and RAG system retrieval -->

### Error Handling Concepts
`error-handling`, `error-wrapping`, `error-chaining`, `error-types`, `error-codes`, `error-categories`, `error-context`, `structured-errors`, `typed-errors`

### Automation & AI
`ai-agents`, `automated-retry`, `retry-logic`, `machine-readable`, `decision-making`, `error-routing`, `self-healing`, `mcp`, `llm-integration`, `json-rpc`, `ai-readable`

### Observability
`structured-logging`, `json-logging`, `error-monitoring`, `error-tracking`, `debugging`, `stack-traces`, `caller-information`, `distributed-tracing`, `correlation-id`, `request-id`, `opentelemetry`, `datadog`

### Metrics & Monitoring
`prometheus`, `grafana`, `metrics`, `labels`, `tags`, `dashboards`, `alerting`, `sla-monitoring`, `performance-tracking`, `duration-tracking`

### Web & API
`http-errors`, `api-errors`, `status-codes`, `json-responses`, `rest-api`, `error-responses`, `rate-limiting`

### Microservices
`distributed-systems`, `microservices`, `service-mesh`, `tracing`, `correlation`, `multi-tenant`

### Go Ecosystem
`golang`, `go-errors`, `errors-package`, `error-interface`, `errors-is`, `errors-as`

### Operations
`retry-strategies`, `exponential-backoff`, `circuit-breaker`, `resilience`, `fault-tolerance`, `error-recovery`, `idempotency`

---

## ❓ FAQ for RAG Systems

<!-- RAG: Frequently asked questions with complete answers for AI retrieval -->

### Q: What is errific?

**A**: errific is an AI-ready error handling library for Go that adds structured context, machine-readable error codes, automated retry metadata, distributed tracing IDs, MCP integration, and JSON serialization to Go errors. It's designed for systems where AI agents, automation, or monitoring tools need to make decisions based on error metadata.

### Q: How is this different from stdlib errors?

**A**: stdlib `errors` provides basic error wrapping with `fmt.Errorf("%w", err)`. errific adds:
- **Automatic caller information** (file:line.function)
- **Structured context** (Context maps with any data)
- **Error codes and categories** for classification
- **Retry metadata** (retryable, retry_after, max_retries)
- **HTTP status codes** for API errors
- **MCP error codes** for LLM integration
- **Distributed tracing IDs** (correlation_id, request_id)
- **Help/suggestion text** for AI agents and users
- **Labels and tags** for metrics and filtering
- **Performance tracking** (duration, timestamp)
- **JSON serialization** for logging
- Full compatibility with `errors.Is()` and `errors.As()`

### Q: When should I use errific vs stdlib errors?

**A**: Use errific when:
- ✅ Building REST APIs (need HTTP status codes)
- ✅ Implementing retry logic (need retryable metadata)
- ✅ Building for AI/LLMs (need MCP, help text)
- ✅ Microservices (need distributed tracing)
- ✅ Structured logging (need rich context)
- ✅ Error monitoring (need codes, categories, labels)

Use stdlib errors when:
- ❌ Simple scripts or tools
- ❌ No need for metadata
- ❌ Performance is absolutely critical (errific adds ~1-2µs overhead)

### Q: Is errific thread-safe?

**A**: Yes. All operations are thread-safe:
- Error creation is concurrent-safe
- Configuration uses `sync.RWMutex`
- All helper functions safe for concurrent use
- No shared mutable state in error instances

### Q: What's the performance overhead?

**A**: Minimal:
- **Error creation**: ~1-2 microseconds (vs ~500ns for stdlib)
- **Memory**: ~500 bytes per error with full metadata
- **CPU**: Negligible for most applications
- **Allocations**: 1 allocation per error

For 99% of applications, the overhead is negligible compared to the operation that caused the error (network call, database query, etc.).

### Q: Can I migrate incrementally from stdlib/pkg/errors?

**A**: Yes! errific is designed for gradual adoption:
1. All errific helper functions work with any error type (return zero values for non-errific errors)
2. You can wrap stdlib errors: `errific.Err("new error").New(stdlibErr)`
3. errific errors work with `errors.Is()` and `errors.As()`
4. You can mix errific and stdlib errors in the same codebase

See [DECISION_GUIDE.md → Migration](./DECISION_GUIDE.md#migration-decision-guide) for detailed migration patterns.

### Q: How do I test code that uses errific?

**A**: Use standard Go testing with `errors.Is()`:
```go
func TestUserNotFound(t *testing.T) {
    err := getUserByID("invalid-id")

    // Test error type
    assert.True(t, errors.Is(err, ErrUserNotFound))

    // Test metadata
    assert.Equal(t, 404, errific.GetHTTPStatus(err))
    assert.Equal(t, errific.CategoryNotFound, errific.GetCategory(err))
    assert.False(t, errific.IsRetryable(err))
}
```

### Q: Does errific work with OpenTelemetry/Datadog/Sentry?

**A**: Yes! errific integrates well with observability tools:
- **OpenTelemetry**: Use `WithCorrelationID()` with trace IDs, `WithLabels()` for span attributes
- **Datadog**: Labels map to Datadog tags, context maps to JSON logs
- **Sentry**: JSON serialization provides rich error context
- **Prometheus**: Labels map to metric labels
- **Custom tools**: JSON format is compatible with any logging/monitoring system

### Q: Can I use errific for MCP tool servers?

**A**: Absolutely! errific was designed with MCP in mind:
- `WithMCPCode()` for JSON-RPC 2.0 error codes
- `WithHelp()` for LLM-readable explanations
- `WithSuggestion()` for automated recovery
- `WithDocs()` for documentation links
- `ToMCPError()` converts to MCP format

See [Use Case 6: MCP Tool Server](#use-case-6-mcp-tool-server-for-llms) above.

### Q: What Go version is required?

**A**: Go 1.20+ required. No external dependencies (stdlib only).

### Q: How do I serialize custom types in Context?

**A**: Context supports any JSON-serializable type:
```go
type CustomData struct {
    Field string `json:"field"`
}

err := ErrAPI.New().WithContext(errific.Context{
    "custom": CustomData{Field: "value"},  // OK if JSON-serializable
    "map": map[string]int{"count": 42},    // OK
    "slice": []string{"a", "b"},           // OK
})
```

Avoid: channels, functions, unexported fields.

---

## 📄 Document Metadata

### Last Updated
2024-11-27

### Documentation Version
Phase 1 & 2A Complete (v1.0.0)

### Target Audience
- AI Agents & RAG Systems
- LLM Tool Developers (MCP)
- Go Developers (REST APIs, microservices)
- DevOps Engineers (monitoring, observability)
- SRE Teams (error tracking, alerting)

### Related Libraries
- stdlib `errors` - Basic error handling
- `github.com/pkg/errors` - Stack traces
- `github.com/cockroachdb/errors` - Feature-rich errors
- `github.com/rotisserie/eris` - Stack traces with JSON

### Complementary Tools
- **OpenTelemetry** - Distributed tracing
- **Prometheus** - Metrics and monitoring
- **Datadog** - APM and logging
- **Sentry** - Error tracking
- **MCP** - LLM tool protocol

### Version History
- **v1.0.0** (Phase 1 & 2A Complete):
  - Core Types (Err, Context, Category)
  - Retry Metadata (retryable, retry_after, max_retries)
  - HTTP Status Codes
  - MCP Integration (error codes, help, suggestions)
  - Distributed Tracing (correlation_id, request_id, user_id, session_id)
  - Metrics & Labels (tags, labels, prometheus integration)
  - Performance Tracking (duration, timestamp)
  - JSON Serialization
  - Complete API Reference (4,400+ lines)
  - Decision Guide (800+ lines)
  - RAG-optimized documentation
