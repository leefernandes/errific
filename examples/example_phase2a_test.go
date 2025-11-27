package examples

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/leefernandes/errific"
)

var (
	ErrMCPToolExecution errific.Err = "MCP tool execution failed"
	ErrAPICall          errific.Err = "API call failed"
)

// Example_mcpToolError demonstrates MCP tool error handling with correlation tracking,
// recovery suggestions, and semantic tags for RAG systems.
func Example_mcpToolError() {
	errific.Configure()

	// MCP tool error with full metadata for AI agents
	err := ErrMCPToolExecution.New().
		WithMCPCode(errific.MCPToolError).
		WithCorrelationID("corr-abc-123").
		WithRequestID("req-xyz-456").
		WithUserID("user-789").
		WithSessionID("sess-def-012").
		WithHelp("The search_database tool encountered an error while querying the users table").
		WithSuggestion("Check database connection and retry with exponential backoff").
		WithDocs("https://docs.mcp.ai/tools/search-database").
		WithTags("mcp", "tool-error", "database", "retryable").
		WithLabel("tool_name", "search_database").
		WithLabel("severity", "medium").
		WithTimestamp(time.Now()).
		WithDuration(2500 * time.Millisecond).
		WithCategory(errific.CategoryServer).
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3)

	fmt.Println(err)
	// Output: MCP tool execution failed [errific/examples/example_phase2a_test.go:22.Example_mcpToolError]
}

// Example_mcpErrorFormat demonstrates converting an errific error to MCP JSON-RPC 2.0 format
// for use in MCP server error responses.
func Example_mcpErrorFormat() {
	errific.Configure()

	err := ErrMCPToolExecution.New().
		WithMCPCode(errific.MCPInvalidParams).
		WithContext(errific.Context{
			"param":    "query",
			"expected": "string",
			"received": "number",
		}).
		WithHelp("The 'query' parameter must be a string value")

	// Convert to MCP format
	mcpErr := errific.ToMCPError(err)
	jsonBytes, _ := json.MarshalIndent(mcpErr, "", "  ")
	fmt.Println(string(jsonBytes))
	// Output will be MCP JSON-RPC 2.0 format (line numbers may vary)
}

// Example_correlationTracking demonstrates using correlation IDs to track errors
// across distributed MCP tool calls.
func Example_correlationTracking() {
	errific.Configure()

	correlationID := "trace-12345"

	// First tool call
	err1 := ErrAPICall.New().
		WithCorrelationID(correlationID).
		WithRequestID("req-001").
		WithContext(errific.Context{
			"service":  "user-api",
			"endpoint": "/users/123",
		})

	// Second tool call (same correlation ID)
	err2 := ErrAPICall.New().
		WithCorrelationID(correlationID).
		WithRequestID("req-002").
		WithContext(errific.Context{
			"service":  "order-api",
			"endpoint": "/orders/456",
		})

	// AI can correlate these errors as part of the same operation
	fmt.Printf("Both errors share correlation ID: %s\n", errific.GetCorrelationID(err1))
	fmt.Printf("Request 1 ID: %s\n", errific.GetRequestID(err1))
	fmt.Printf("Request 2 ID: %s\n", errific.GetRequestID(err2))
	// Output:
	// Both errors share correlation ID: trace-12345
	// Request 1 ID: req-001
	// Request 2 ID: req-002
}

// Example_recoverySuggestions demonstrates providing recovery guidance for AI agents
// to automatically resolve errors.
func Example_recoverySuggestions() {
	errific.Configure()

	var ErrDatabaseTimeout errific.Err = "database query timeout"

	err := ErrDatabaseTimeout.New().
		WithHelp("The database query exceeded the 30 second timeout").
		WithSuggestion("Increase query timeout to 60 seconds or optimize the query with an index").
		WithDocs("https://docs.example.com/database/timeouts").
		WithContext(errific.Context{
			"query":       "SELECT * FROM large_table WHERE complex_condition",
			"timeout_sec": 30,
			"table_size":  "1.2TB",
		})

	// AI agent can extract recovery information
	fmt.Printf("Help: %s\n", errific.GetHelp(err))
	fmt.Printf("Suggestion: %s\n", errific.GetSuggestion(err))
	fmt.Printf("Documentation: %s\n", errific.GetDocs(err))
	// Output:
	// Help: The database query exceeded the 30 second timeout
	// Suggestion: Increase query timeout to 60 seconds or optimize the query with an index
	// Documentation: https://docs.example.com/database/timeouts
}

// Example_semanticTags demonstrates using semantic tags for RAG systems
// to categorize and search errors.
func Example_semanticTags() {
	errific.Configure()

	var ErrNetworkTimeout errific.Err = "network timeout"

	err := ErrNetworkTimeout.New().
		WithTags("network", "timeout", "retryable", "transient", "connectivity").
		WithCategory(errific.CategoryTimeout).
		WithRetryable(true)

	// RAG system can search errors by tags
	tags := errific.GetTags(err)
	fmt.Printf("Semantic tags: %v\n", tags)
	fmt.Printf("Number of tags: %d\n", len(tags))
	// Output:
	// Semantic tags: [network timeout retryable transient connectivity]
	// Number of tags: 5
}

// Example_labelsForFiltering demonstrates using key-value labels
// to filter and group errors for monitoring and alerting.
func Example_labelsForFiltering() {
	errific.Configure()

	var ErrServiceDegraded errific.Err = "service degraded"

	err := ErrServiceDegraded.New().
		WithLabel("severity", "high").
		WithLabel("team", "backend").
		WithLabel("region", "us-east-1").
		WithLabel("environment", "production").
		WithLabel("alert_oncall", "true")

	// Monitoring system can filter errors by labels
	labels := errific.GetLabels(err)
	fmt.Printf("Severity: %s\n", labels["severity"])
	fmt.Printf("Team: %s\n", labels["team"])
	fmt.Printf("Should alert on-call: %s\n", errific.GetLabel(err, "alert_oncall"))
	// Output:
	// Severity: high
	// Team: backend
	// Should alert on-call: true
}

// Example_timestampAndDuration demonstrates tracking when an error occurred
// and how long the operation took before failing.
func Example_timestampAndDuration() {
	errific.Configure()

	var ErrSlowQuery errific.Err = "slow database query"

	start := time.Now()
	// Simulate slow operation
	time.Sleep(100 * time.Millisecond)

	err := ErrSlowQuery.New().
		WithTimestamp(start).
		WithDuration(time.Since(start)).
		WithContext(errific.Context{
			"query": "SELECT * FROM users WHERE complex_condition",
		})

	// Monitoring can track error timing
	ts := errific.GetTimestamp(err)
	duration := errific.GetDuration(err)
	fmt.Printf("Error occurred at: %s\n", ts.Format(time.RFC3339))
	fmt.Printf("Operation duration: %s\n", duration)
	// Output will vary based on actual timing
}

// Example_phase2aJSONSerialization demonstrates JSON serialization of all Phase 2A fields
// for structured logging and monitoring systems.
func Example_phase2aJSONSerialization() {
	errific.Configure()

	var ErrCompleteExample errific.Err = "complete Phase 2A example"

	err := ErrCompleteExample.New().
		WithCode("PHASE2A_001").
		WithCategory(errific.CategoryServer).
		WithMCPCode(errific.MCPToolError).
		WithCorrelationID("corr-123").
		WithRequestID("req-456").
		WithUserID("user-789").
		WithSessionID("sess-abc").
		WithHelp("Example error demonstrating all Phase 2A fields").
		WithSuggestion("Review the documentation for Phase 2A features").
		WithDocs("https://docs.example.com/phase2a").
		WithTags("example", "phase2a", "complete").
		WithLabel("version", "2.0").
		WithTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)).
		WithDuration(5 * time.Second).
		WithRetryable(true).
		WithRetryAfter(10 * time.Second).
		WithMaxRetries(3).
		WithHTTPStatus(500)

	jsonBytes, _ := json.MarshalIndent(err, "", "  ")
	fmt.Println(string(jsonBytes))
	// Output will be JSON with all Phase 2A fields (line numbers and timestamps may vary)
}

// Example_mcpInvalidParams demonstrates handling MCP invalid parameter errors
// with detailed validation context.
func Example_mcpInvalidParams() {
	errific.Configure()

	var ErrInvalidToolParams errific.Err = "invalid tool parameters"

	err := ErrInvalidToolParams.New().
		WithMCPCode(errific.MCPInvalidParams).
		WithContext(errific.Context{
			"tool":     "search_database",
			"param":    "limit",
			"expected": "number (1-100)",
			"received": "string",
			"value":    "all",
		}).
		WithHelp("The 'limit' parameter must be a number between 1 and 100").
		WithSuggestion("Provide a numeric limit value, e.g., 10 or 50").
		WithDocs("https://docs.mcp.ai/tools/search-database#parameters").
		WithRetryable(false)

	fmt.Printf("MCP Code: %d\n", errific.GetMCPCode(err))
	fmt.Printf("Help: %s\n", errific.GetHelp(err))
	fmt.Printf("Retryable: %v\n", errific.IsRetryable(err))
	// Output:
	// MCP Code: -32602
	// Help: The 'limit' parameter must be a number between 1 and 100
	// Retryable: false
}

// Example_aiAgentWorkflow demonstrates a complete AI agent error handling workflow
// using Phase 2A features for automated decision-making.
func Example_aiAgentWorkflow() {
	errific.Configure()

	var ErrToolFailed errific.Err = "tool execution failed"

	// AI encounters an error during tool execution
	err := ErrToolFailed.New().
		WithMCPCode(errific.MCPToolError).
		WithCorrelationID("workflow-123").
		WithRequestID("step-1").
		WithHelp("Database connection pool exhausted").
		WithSuggestion("Wait 5 seconds and retry, or increase pool size").
		WithDocs("https://docs.ai/error-recovery/db-pool").
		WithTags("database", "connection-pool", "resource-exhaustion", "retryable").
		WithLabel("criticality", "medium").
		WithLabel("auto_recover", "true").
		WithTimestamp(time.Now()).
		WithRetryable(true).
		WithRetryAfter(5 * time.Second).
		WithMaxRetries(3)

	// AI agent decision-making process
	fmt.Println("=== AI Agent Error Analysis ===")

	// 1. Check if error is retryable
	if errific.IsRetryable(err) {
		fmt.Printf("✓ Error is retryable\n")
	}

	// 2. Get retry parameters
	retryAfter := errific.GetRetryAfter(err)
	maxRetries := errific.GetMaxRetries(err)
	fmt.Printf("✓ Wait %v before retry\n", retryAfter)
	fmt.Printf("✓ Maximum %d retry attempts\n", maxRetries)

	// 3. Check if auto-recovery is enabled
	if errific.GetLabel(err, "auto_recover") == "true" {
		fmt.Printf("✓ Auto-recovery enabled\n")
	}

	// 4. Extract recovery guidance
	fmt.Printf("✓ Recovery help: %s\n", errific.GetHelp(err))

	// 5. Track correlation for distributed tracing
	fmt.Printf("✓ Correlation ID: %s\n", errific.GetCorrelationID(err))

	// Output:
	// === AI Agent Error Analysis ===
	// ✓ Error is retryable
	// ✓ Wait 5s before retry
	// ✓ Maximum 3 retry attempts
	// ✓ Auto-recovery enabled
	// ✓ Recovery help: Database connection pool exhausted
	// ✓ Correlation ID: workflow-123
}
