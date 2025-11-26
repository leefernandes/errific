// Package errific provides enhanced error handling for Go with caller information,
// clean error wrapping, and helpful formatting methods.
//
// errific simplifies error creation by adding runtime caller metadata (file, line, function)
// to errors, making debugging easier without sacrificing clean error messages. It supports
// error chaining, formatted messages, and configurable output options including stack traces.
//
// Basic usage:
//
//	var ErrProcessThing errific.Err = "error processing thing"
//
//	func process() error {
//	    if err := validate(); err != nil {
//	        return ErrProcessThing.New(err)
//	    }
//	    return nil
//	}
//
// The resulting error includes caller information:
//
//	error processing thing [mypackage/file.go:42.process]
//	validation failed [mypackage/validate.go:15.validate]
//
// Configuration options include caller position (prefix/suffix/disabled),
// layout (newline/inline), stack traces, and path trimming.
package errific

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// Err string type.
//
// To include runtime caller information on the error,
// one of the Err methods, other than Error(), must be called.
//
// For examples see the example tests.  All examples
// demonstrate using exported errors as a recommended best
// practice because exported errors enable unit-tests that assert
// expected errors such as: assert.ErrorIs(t, err, ErrProcessThing).
type Err string

// New returns an error using Err as text with errors joined.
//
//	var ErrProcessThing errific.Err = "error processing a thing"
//
//	return ErrProcessThing.New(err)
func (e Err) New(errs ...error) errific {
	a := make([]any, len(errs))
	for i := range errs {
		a[i] = errs[i]
	}

	caller, stack := callstack(a)
	return errific{
		err:    e,
		errs:   errs,
		caller: caller,
		stack:  stack,
	}
}

// Errorf returns an error using Err formatted as text.
// Use Errorf if your Err string itself contains fmt format specifiers.
//
//	var ErrProcessThing errific.Err = "error processing thing id: '%s'"
//
//	return ErrProcessThing.Errorf("abc")
func (e Err) Errorf(a ...any) errific {
	caller, stack := callstack(a)
	return errific{
		err:    fmt.Errorf(e.Error(), a...),
		caller: caller,
		unwrap: []error{e},
		stack:  stack,
	}
}

// Withf returns an error with a formatted string inline to Err as text.
//
//	var ErrProcessThing errific.Err = "error processing thing"
//
//	return ErrProcessThing.Withf("id: '%s'", "abc")
func (e Err) Withf(format string, a ...any) errific {
	caller, stack := callstack(a)
	format = e.Error() + ": " + format
	return errific{
		err:    fmt.Errorf(format, a...),
		caller: caller,
		unwrap: []error{e},
		stack:  stack,
	}
}

// Wrapf return an error using Err as text and wraps a formatted error.
// Use Wrapf to format an error and wrap it.
//
//	var ErrProcessThing errific.Err = "error processing thing"
//
//	return ErrProcessThing.Wrapf("cause: %w", err)
func (e Err) Wrapf(format string, a ...any) errific {
	caller, stack := callstack(a)
	return errific{
		err:    e,
		errs:   []error{fmt.Errorf(format, a...)},
		caller: caller,
		stack:  stack,
	}
}

func (e Err) Error() string {
	return string(e)
}

// Context is a map of key-value pairs that provides additional context for errors.
// This structured data can be used for debugging, logging, and automated error handling.
type Context map[string]any

// Category represents the category of an error for automated handling.
type Category string

const (
	// CategoryClient represents client-side errors (4xx).
	CategoryClient Category = "client"
	// CategoryServer represents server-side errors (5xx).
	CategoryServer Category = "server"
	// CategoryNetwork represents network connectivity errors.
	CategoryNetwork Category = "network"
	// CategoryValidation represents input validation errors.
	CategoryValidation Category = "validation"
	// CategoryNotFound represents resource not found errors (404).
	CategoryNotFound Category = "not_found"
	// CategoryUnauthorized represents authentication/authorization errors (401/403).
	CategoryUnauthorized Category = "unauthorized"
	// CategoryTimeout represents timeout errors.
	CategoryTimeout Category = "timeout"
)

// MCP error codes following JSON-RPC 2.0 specification.
// These codes enable errific errors to be serialized in MCP-compatible format
// for AI tool calling and Model Context Protocol integration.
const (
	// MCPParseError represents invalid JSON was received by the server.
	MCPParseError = -32700
	// MCPInvalidRequest represents the JSON sent is not a valid Request object.
	MCPInvalidRequest = -32600
	// MCPMethodNotFound represents the method does not exist / is not available.
	MCPMethodNotFound = -32601
	// MCPInvalidParams represents invalid method parameter(s).
	MCPInvalidParams = -32602
	// MCPInternalError represents internal JSON-RPC error.
	MCPInternalError = -32603
	// MCPToolError represents a tool execution error (custom range -32000 to -32099).
	MCPToolError = -32000
)

// MCPError represents a Model Context Protocol error in JSON-RPC 2.0 format.
// This format is compatible with MCP server error responses and AI tool calling protocols.
type MCPError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface for MCPError.
func (m MCPError) Error() string {
	return fmt.Sprintf("MCP error %d: %s", m.Code, m.Message)
}

type errific struct {
	err        error         // primary error.
	errs       []error       // errors used in string output, and satisfy errors.Is.
	unwrap     []error       // errors not used in string output, but satisfy errors.Is.
	caller     string        // caller information.
	stack      []byte        // optional stack buffer.
	context    Context       // structured context data.
	code       string        // error code for machine-readable identification.
	category   Category      // error category for automated handling.
	retryable  bool          // whether this error is retryable.
	retryAfter time.Duration // suggested retry delay.
	maxRetries int           // maximum number of retry attempts.
	httpStatus int           // HTTP status code (0 if not applicable).
	mcpCode    int           // MCP error code for JSON-RPC 2.0 compatibility (0 if not applicable).
	// Phase 2A: MCP & RAG features
	correlationID string            // correlation ID for distributed tracing.
	requestID     string            // request ID for this operation.
	userID        string            // user ID associated with the error.
	sessionID     string            // session ID for multi-step operations.
	help          string            // help text for recovery.
	suggestion    string            // suggested action to resolve error.
	docsURL       string            // documentation URL for more info.
	tags          []string          // semantic tags for RAG search and categorization.
	labels        map[string]string // key-value labels for filtering and grouping.
	timestamp     time.Time         // when the error occurred.
	duration      time.Duration     // operation duration before error.
}

func (e errific) Error() (msg string) {
	cMu.RLock()
	caller := c.caller
	layout := c.layout
	withStack := c.withStack
	cMu.RUnlock()

	switch caller {
	case Disabled:

	case Prefix:
		msg = fmt.Sprintf("[%s] %s", e.caller, e.err.Error())

	default:
		msg = fmt.Sprintf("%s [%s]", e.err.Error(), e.caller)
	}

	switch layout {
	case Inline:
		for i := range e.errs {
			msg = fmt.Sprintf("%s ↩ %s", msg, e.errs[i].Error())
		}

	default:
		for i := range e.errs {
			msg = fmt.Sprintf("%s\n%s", msg, e.errs[i].Error())
		}
	}

	if withStack && len(e.stack) > 0 {
		// Remove duplicate stack traces from nested errors before appending
		stackStr := string(e.stack)
		msg = strings.ReplaceAll(msg, stackStr, "")
		msg += stackStr
	}

	return msg
}

func (e errific) Join(errs ...error) error {
	e.errs = append(e.errs, errs...)
	return e
}

func (e errific) Withf(format string, a ...any) errific {
	originalErr := e.err
	format = e.err.Error() + ": " + format
	e.err = fmt.Errorf(format, a...)
	e.unwrap = append(e.unwrap, originalErr)
	return e
}

func (e errific) Wrapf(format string, a ...any) errific {
	e.errs = append(e.errs, fmt.Errorf(format, a...))
	return e
}

// WithContext adds structured context data to the error.
// Context is a map of key-value pairs that can be used for debugging,
// logging, and automated error handling.
//
//	err := ErrDatabaseQuery.New(sqlErr).WithContext(errific.Context{
//	    "query": "SELECT * FROM users",
//	    "duration_ms": 1500,
//	})
func (e errific) WithContext(ctx Context) errific {
	if e.context == nil {
		e.context = make(Context)
	}
	for k, v := range ctx {
		e.context[k] = v
	}
	return e
}

// WithCode sets an error code for machine-readable identification.
// Error codes enable automated error handling and routing.
//
//	err := ErrDatabaseConnection.New().WithCode("DB_CONN_TIMEOUT")
func (e errific) WithCode(code string) errific {
	e.code = code
	return e
}

// WithCategory sets the error category for automated handling.
// Categories help AI agents and automation systems decide how to respond.
//
//	err := ErrDatabaseConnection.New().WithCategory(errific.CategoryNetwork)
func (e errific) WithCategory(category Category) errific {
	e.category = category
	return e
}

// WithRetryable marks whether the error is retryable.
// This enables automated retry logic in AI agents and resilience systems.
//
//	err := ErrAPICall.New(httpErr).WithRetryable(true)
func (e errific) WithRetryable(retryable bool) errific {
	e.retryable = retryable
	return e
}

// WithRetryAfter sets the suggested retry delay duration.
// This guides automated retry strategies with appropriate backoff.
//
//	err := ErrRateLimit.New().WithRetryAfter(5 * time.Second)
func (e errific) WithRetryAfter(duration time.Duration) errific {
	e.retryAfter = duration
	return e
}

// WithMaxRetries sets the maximum number of retry attempts.
// This prevents infinite retry loops in automated systems.
//
//	err := ErrAPICall.New().WithRetryable(true).WithMaxRetries(3)
func (e errific) WithMaxRetries(max int) errific {
	e.maxRetries = max
	return e
}

// WithHTTPStatus sets the HTTP status code for this error.
// This enables automatic HTTP response handling in web services.
//
//	err := ErrValidation.New().WithHTTPStatus(400)
func (e errific) WithHTTPStatus(status int) errific {
	e.httpStatus = status
	return e
}

// WithMCPCode sets an MCP error code following JSON-RPC 2.0 specification.
// Use the predefined MCP constants (MCPInternalError, MCPInvalidParams, etc.)
// or custom codes in the range -32000 to -32099 for application-specific errors.
//
//	err := ErrToolExecution.New().WithMCPCode(MCPToolError)
func (e errific) WithMCPCode(code int) errific {
	e.mcpCode = code
	return e
}

// WithCorrelationID sets a correlation ID for distributed tracing.
// This enables tracking errors across MCP tool calls and distributed systems.
//
//	err := ErrMCPTool.New().WithCorrelationID(correlationID)
func (e errific) WithCorrelationID(id string) errific {
	e.correlationID = id
	return e
}

// WithRequestID sets a request ID for this specific operation.
// This enables tracking individual requests in logging and monitoring.
//
//	err := ErrAPI.New().WithRequestID(uuid.New().String())
func (e errific) WithRequestID(id string) errific {
	e.requestID = id
	return e
}

// WithUserID sets the user ID associated with this error.
// This enables user-specific error tracking and analysis.
//
//	err := ErrPermission.New().WithUserID(userID)
func (e errific) WithUserID(id string) errific {
	e.userID = id
	return e
}

// WithSessionID sets a session ID for multi-step operations.
// This enables tracking errors across agent conversation sessions.
//
//	err := ErrAgent.New().WithSessionID(sessionID)
func (e errific) WithSessionID(id string) errific {
	e.sessionID = id
	return e
}

// WithHelp adds recovery help text to the error.
// This enables AI agents to display actionable guidance to users.
//
//	err := ErrPermission.New().WithHelp("Run 'kubectl get roles' to check permissions")
func (e errific) WithHelp(text string) errific {
	e.help = text
	return e
}

// WithSuggestion adds a suggested action to resolve the error.
// This enables AI agents to attempt automatic recovery.
//
//	err := ErrRateLimit.New().WithSuggestion("Reduce request frequency or upgrade plan")
func (e errific) WithSuggestion(text string) errific {
	e.suggestion = text
	return e
}

// WithDocs adds a documentation URL for more information.
// This enables AI agents to provide users with detailed documentation.
//
//	err := ErrConfig.New().WithDocs("https://docs.example.com/configuration")
func (e errific) WithDocs(url string) errific {
	e.docsURL = url
	return e
}

// WithTags adds semantic tags for RAG search and categorization.
// Tags enable semantic search, error clustering, and pattern recognition.
//
//	err := ErrMCPTool.New().WithTags("mcp", "tool", "search", "timeout")
func (e errific) WithTags(tags ...string) errific {
	e.tags = append(e.tags, tags...)
	return e
}

// WithLabels adds key-value labels for filtering and grouping.
// Labels enable precise error filtering in monitoring and analytics.
//
//	err := ErrAPI.New().WithLabels(map[string]string{
//	    "environment": "production",
//	    "region": "us-east-1",
//	})
func (e errific) WithLabels(labels map[string]string) errific {
	if e.labels == nil {
		e.labels = make(map[string]string)
	}
	for k, v := range labels {
		e.labels[k] = v
	}
	return e
}

// WithLabel adds a single key-value label.
// Convenience method for adding individual labels.
//
//	err := ErrAPI.New().WithLabel("environment", "production")
func (e errific) WithLabel(key, value string) errific {
	if e.labels == nil {
		e.labels = make(map[string]string)
	}
	e.labels[key] = value
	return e
}

// WithTimestamp sets when the error occurred.
// If not set, defaults to time of error creation.
//
//	err := ErrOperation.New().WithTimestamp(time.Now())
func (e errific) WithTimestamp(t time.Time) errific {
	e.timestamp = t
	return e
}

// WithDuration sets the operation duration before the error occurred.
// This enables performance analysis and SLA monitoring.
//
//	err := ErrSlowQuery.New().WithDuration(elapsed)
func (e errific) WithDuration(d time.Duration) errific {
	e.duration = d
	return e
}

func (e errific) Unwrap() []error {
	var errs []error
	if e.err != nil {
		errs = append(errs, e.err)
	}
	errs = append(errs, e.errs...)
	errs = append(errs, e.unwrap...)
	return errs
}

// MarshalJSON implements json.Marshaler for structured error output.
// This enables errific errors to be serialized to JSON for logging,
// API responses, and integration with monitoring systems.
func (e errific) MarshalJSON() ([]byte, error) {
	type jsonError struct {
		Error         string            `json:"error"`
		Code          string            `json:"code,omitempty"`
		Category      Category          `json:"category,omitempty"`
		Caller        string            `json:"caller,omitempty"`
		Context       Context           `json:"context,omitempty"`
		Retryable     bool              `json:"retryable,omitempty"`
		RetryAfter    string            `json:"retry_after,omitempty"`
		MaxRetries    int               `json:"max_retries,omitempty"`
		HTTPStatus    int               `json:"http_status,omitempty"`
		MCPCode       int               `json:"mcp_code,omitempty"`
		Stack         []string          `json:"stack,omitempty"`
		Wrapped       []string          `json:"wrapped,omitempty"`
		CorrelationID string            `json:"correlation_id,omitempty"`
		RequestID     string            `json:"request_id,omitempty"`
		UserID        string            `json:"user_id,omitempty"`
		SessionID     string            `json:"session_id,omitempty"`
		Help          string            `json:"help,omitempty"`
		Suggestion    string            `json:"suggestion,omitempty"`
		Docs          string            `json:"docs,omitempty"`
		Tags          []string          `json:"tags,omitempty"`
		Labels        map[string]string `json:"labels,omitempty"`
		Timestamp     string            `json:"timestamp,omitempty"`
		Duration      string            `json:"duration,omitempty"`
	}

	je := jsonError{
		Error:         e.err.Error(),
		Code:          e.code,
		Category:      e.category,
		Caller:        e.caller,
		Context:       e.context,
		Retryable:     e.retryable,
		MaxRetries:    e.maxRetries,
		HTTPStatus:    e.httpStatus,
		MCPCode:       e.mcpCode,
		CorrelationID: e.correlationID,
		RequestID:     e.requestID,
		UserID:        e.userID,
		SessionID:     e.sessionID,
		Help:          e.help,
		Suggestion:    e.suggestion,
		Docs:          e.docsURL,
		Tags:          e.tags,
		Labels:        e.labels,
	}

	if e.retryAfter > 0 {
		je.RetryAfter = e.retryAfter.String()
	}

	if !e.timestamp.IsZero() {
		je.Timestamp = e.timestamp.Format(time.RFC3339)
	}

	if e.duration > 0 {
		je.Duration = e.duration.String()
	}

	// Parse stack trace into lines
	if len(e.stack) > 0 {
		stackLines := strings.Split(strings.TrimSpace(string(e.stack)), "\n")
		je.Stack = stackLines
	}

	// Add wrapped errors
	for _, err := range e.errs {
		je.Wrapped = append(je.Wrapped, err.Error())
	}

	return json.Marshal(je)
}

func unwrapStack(errs []any) []byte {
	for _, err := range errs {
		if err == nil {
			return nil
		}
		if e, ok := err.(errific); ok {
			return e.stack
		}

		if err, ok := err.(error); ok {
			return unwrapStack([]any{errors.Unwrap(err)})
		}
	}
	return nil
}

func callstack(errs []any) (caller string, stack []byte) {
	pc := make([]uintptr, 32)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return "", stack
	}

	frames := runtime.CallersFrames(pc)
	frame, more := frames.Next()
	caller = parseFrame(frame)

	cMu.RLock()
	withStack := c.withStack
	cMu.RUnlock()

	if !withStack {
		return caller, stack
	}

	stack = unwrapStack(errs)

	if len(stack) > 0 {
		return caller, stack
	}

	if !more {
		return caller, stack
	}

	for {
		frame, more := frames.Next()
		// Skip frames from GOROOT and _testmain.go (generated test runner)
		if !strings.HasPrefix(frame.File, goroot) && !strings.HasSuffix(frame.File, "_testmain.go") {
			frameStr := fmt.Sprintf("\n  %s", parseFrame(frame))
			stack = append(stack, frameStr...)
		}
		if !more {
			break
		}
	}

	return caller, stack
}

func parseFrame(frame runtime.Frame) string {
	funcParts := strings.Split(frame.Function, "/")
	funcParts = strings.Split(funcParts[len(funcParts)-1], ".")
	callFunc := funcParts[len(funcParts)-1]
	callFile := frame.File

	cMu.RLock()
	trimPrefixes := c.trimPrefixes
	cMu.RUnlock()

	for _, trimPrefix := range trimPrefixes {
		callFile = strings.TrimPrefix(callFile, trimPrefix)
	}
	callFile = strings.TrimPrefix(callFile, goroot)
	callFile = strings.TrimPrefix(callFile, root)
	callLine := frame.Line

	return fmt.Sprintf("%s:%d.%s", callFile, callLine, callFunc)
}

// GetContext extracts structured context from an error.
// Returns nil if the error doesn't have context data.
// This function works with any error type but only extracts
// context from errific errors.
func GetContext(err error) Context {
	if err == nil {
		return nil
	}

	// Check if it's an errific error
	var e errific
	if errors.As(err, &e) {
		return e.context
	}

	return nil
}

// GetCode extracts the error code from an error.
// Returns an empty string if the error doesn't have a code.
func GetCode(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.code
	}

	return ""
}

// GetCategory extracts the error category from an error.
// Returns an empty category if the error doesn't have one.
func GetCategory(err error) Category {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.category
	}

	return ""
}

// IsRetryable checks if an error is marked as retryable.
// Returns false if the error is not retryable or not an errific error.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	var e errific
	if errors.As(err, &e) {
		return e.retryable
	}

	return false
}

// GetRetryAfter extracts the suggested retry delay from an error.
// Returns 0 if no retry delay is set.
func GetRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}

	var e errific
	if errors.As(err, &e) {
		return e.retryAfter
	}

	return 0
}

// GetMaxRetries extracts the maximum retry count from an error.
// Returns 0 if no max retries is set.
func GetMaxRetries(err error) int {
	if err == nil {
		return 0
	}

	var e errific
	if errors.As(err, &e) {
		return e.maxRetries
	}

	return 0
}

// GetHTTPStatus extracts the HTTP status code from an error.
// Returns 0 if no HTTP status is set.
func GetHTTPStatus(err error) int {
	if err == nil {
		return 0
	}

	var e errific
	if errors.As(err, &e) {
		return e.httpStatus
	}

	return 0
}

// GetMCPCode extracts the MCP error code from an error.
// Returns 0 if the error is nil or doesn't have an MCP code.
func GetMCPCode(err error) int {
	if err == nil {
		return 0
	}

	var e errific
	if errors.As(err, &e) {
		return e.mcpCode
	}

	return 0
}

// ToMCPError converts any error to MCP JSON-RPC 2.0 format.
// If the error is an errific error with an MCP code set, it uses that code.
// Otherwise, it defaults to MCPInternalError.
// Returns a zero MCPError if err is nil.
//
//	mcpErr := ToMCPError(err)
//	json.NewEncoder(w).Encode(mcpErr)
func ToMCPError(err error) MCPError {
	if err == nil {
		return MCPError{}
	}

	var e errific
	if errors.As(err, &e) {
		return e.ToMCPError()
	}

	// Non-errific errors default to internal error
	return MCPError{
		Code:    MCPInternalError,
		Message: err.Error(),
	}
}

// ToMCPError converts an errific error to MCP JSON-RPC 2.0 format.
// If the error has an MCP code set, it uses that code. Otherwise, it defaults to MCPInternalError.
// The error's JSON serialization is included in the Data field for rich context.
//
//	mcpErr := err.(errific).ToMCPError()
//	json.NewEncoder(w).Encode(mcpErr)
func (e errific) ToMCPError() MCPError {
	code := e.mcpCode
	if code == 0 {
		code = MCPInternalError
	}

	// Serialize the full errific error as data
	data, _ := json.Marshal(e)

	return MCPError{
		Code:    code,
		Message: e.err.Error(),
		Data:    data,
	}
}

// GetCorrelationID extracts the correlation ID from an error.
// Returns an empty string if no correlation ID is set.
func GetCorrelationID(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.correlationID
	}

	return ""
}

// GetRequestID extracts the request ID from an error.
// Returns an empty string if no request ID is set.
func GetRequestID(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.requestID
	}

	return ""
}

// GetUserID extracts the user ID from an error.
// Returns an empty string if no user ID is set.
func GetUserID(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.userID
	}

	return ""
}

// GetSessionID extracts the session ID from an error.
// Returns an empty string if no session ID is set.
func GetSessionID(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.sessionID
	}

	return ""
}

// GetHelp extracts the help text from an error.
// Returns an empty string if no help text is set.
func GetHelp(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.help
	}

	return ""
}

// GetSuggestion extracts the suggestion text from an error.
// Returns an empty string if no suggestion is set.
func GetSuggestion(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.suggestion
	}

	return ""
}

// GetDocs extracts the documentation URL from an error.
// Returns an empty string if no docs URL is set.
func GetDocs(err error) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		return e.docsURL
	}

	return ""
}

// GetTags extracts the semantic tags from an error.
// Returns nil if no tags are set.
func GetTags(err error) []string {
	if err == nil {
		return nil
	}

	var e errific
	if errors.As(err, &e) {
		return e.tags
	}

	return nil
}

// GetLabels extracts the labels from an error.
// Returns nil if no labels are set.
func GetLabels(err error) map[string]string {
	if err == nil {
		return nil
	}

	var e errific
	if errors.As(err, &e) {
		return e.labels
	}

	return nil
}

// GetLabel extracts a specific label value from an error.
// Returns an empty string if the label doesn't exist.
func GetLabel(err error, key string) string {
	if err == nil {
		return ""
	}

	var e errific
	if errors.As(err, &e) {
		if e.labels != nil {
			return e.labels[key]
		}
	}

	return ""
}

// GetTimestamp extracts the timestamp from an error.
// Returns zero time if no timestamp is set.
func GetTimestamp(err error) time.Time {
	if err == nil {
		return time.Time{}
	}

	var e errific
	if errors.As(err, &e) {
		return e.timestamp
	}

	return time.Time{}
}

// GetDuration extracts the operation duration from an error.
// Returns 0 if no duration is set.
func GetDuration(err error) time.Duration {
	if err == nil {
		return 0
	}

	var e errific
	if errors.As(err, &e) {
		return e.duration
	}

	return 0
}
