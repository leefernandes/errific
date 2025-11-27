package errific

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Configure errific options.
func Configure(opts ...Option) {
	cMu.Lock()
	defer cMu.Unlock()

	// defaults
	c.caller = Suffix
	c.layout = Newline
	c.withStack = false
	c.trimPrefixes = nil
	c.trimCWD = false
	c.outputFormat = OutputJSON
	c.verbosity = VerbosityFull

	// Default field visibility (used when verbosity is VerbosityFull or VerbosityCustom)
	c.showCode = true
	c.showCategory = true
	c.showContext = true
	c.showHTTPStatus = true
	c.showRetryMetadata = true
	c.showMCPData = true
	c.showTags = true
	c.showLabels = true
	c.showTimestamps = true

	for _, opt := range opts {
		switch o := opt.(type) {
		case callerOption:
			c.caller = o

		case layoutOption:
			c.layout = o

		case withStackTraceOption:
			c.withStack = o

		case trimPrefixesOption:
			c.trimPrefixes = o.Prefixes()

		case trimCWDOption:
			c.trimCWD = o

		case outputFormatOption:
			c.outputFormat = o

		case verbosityOption:
			c.verbosity = o
			// Set field visibility based on verbosity level
			switch o {
			case VerbosityMinimal:
				c.showCode = false
				c.showCategory = false
				c.showContext = false
				c.showHTTPStatus = false
				c.showRetryMetadata = false
				c.showMCPData = false
				c.showTags = false
				c.showLabels = false
				c.showTimestamps = false

			case VerbosityStandard:
				c.showCode = true
				c.showCategory = true
				c.showContext = true
				c.showHTTPStatus = false
				c.showRetryMetadata = false
				c.showMCPData = false
				c.showTags = false
				c.showLabels = false
				c.showTimestamps = false

			case VerbosityFull:
				c.showCode = true
				c.showCategory = true
				c.showContext = true
				c.showHTTPStatus = true
				c.showRetryMetadata = true
				c.showMCPData = true
				c.showTags = true
				c.showLabels = true
				c.showTimestamps = true
			}

		case fieldVisibilityOption:
			// When using field visibility options, automatically switch to VerbosityCustom
			if c.verbosity != VerbosityCustom {
				c.verbosity = VerbosityCustom
			}
			// Apply the specific field visibility setting
			switch o.field {
			case "code":
				c.showCode = o.show
			case "category":
				c.showCategory = o.show
			case "context":
				c.showContext = o.show
			case "http_status":
				c.showHTTPStatus = o.show
			case "retry_metadata":
				c.showRetryMetadata = o.show
			case "mcp_data":
				c.showMCPData = o.show
			case "tags":
				c.showTags = o.show
			case "labels":
				c.showLabels = o.show
			case "timestamps":
				c.showTimestamps = o.show
			}
		}
	}

	if c.trimCWD {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to not trimming CWD if we can't get it
			c.trimCWD = false
			return
		}

		// Trim the current working directory itself, not its parent
		c.trimPrefixes = append([]string{cwd + "/"}, c.trimPrefixes...)
	}
}

var (
	c struct {
		// Caller will configure the caller: Suffix|Prefix|Disabled.
		// Default is Suffix.
		caller callerOption
		// Layout will configure the layout of wrapped errors: Newline|Inline.
		// Default is Newline.
		layout layoutOption
		// WithStack will append stacktrace to end of message.
		// Default is not including the stack.
		withStack withStackTraceOption
		// TrimPrefixes will trim prefixes from caller frame filenames.
		trimPrefixes []string
		// TrimCWD will trim the current working directory from filenames.
		// Default is false.
		trimCWD trimCWDOption
		// Output format: Pretty, JSON, or Compact.
		// Default is Pretty.
		outputFormat outputFormatOption
		// Verbosity controls which fields are shown in Error() output.
		// Default is VerbosityFull (show all non-empty fields).
		verbosity verbosityOption
		// Field visibility flags (used when verbosity is VerbosityCustom)
		showCode          bool
		showCategory      bool
		showContext       bool
		showHTTPStatus    bool
		showRetryMetadata bool
		showMCPData       bool
		showTags          bool
		showLabels        bool
		showTimestamps    bool
	}
	cMu sync.RWMutex
)

type callerOption int

func (callerOption) ErrificOption() {}

const (
	// Suffix adds caller information at the end of the error message.
	// This is default.
	Suffix callerOption = iota
	// Prefix adds caller information at the beginning of the error message.
	Prefix
	// Disabled does not include caller information in the error message.
	Disabled
)

type layoutOption int

func (layoutOption) ErrificOption() {}

const (
	// Newline joins errors with \n.
	// This is default.
	Newline layoutOption = iota
	// Inline wraps errors with ↩.
	Inline
)

type withStackTraceOption bool

func (withStackTraceOption) ErrificOption() {}

const (
	// Include stacktrace in error message.
	WithStack withStackTraceOption = true
)

type trimPrefixesOption struct {
	prefixes []string
}

func (trimPrefixesOption) ErrificOption() {}

func (t trimPrefixesOption) Prefixes() []string {
	return t.prefixes
}

var (
	// TrimPrefixes from caller frame filenames.
	TrimPrefixes = func(prefixes ...string) trimPrefixesOption {
		return trimPrefixesOption{prefixes: prefixes}
	}
)

type trimCWDOption bool

func (trimCWDOption) ErrificOption() {}

const (
	// Trim current working directory from filenames.
	TrimCWD trimCWDOption = true
)

type Option interface {
	ErrificOption()
}

// outputFormatOption controls the format of error string output.
type outputFormatOption int

func (outputFormatOption) ErrificOption() {}

const (
	// OutputPretty formats errors as human-readable multi-line text with all metadata.
	//
	// Example:
	//   user not found [main.go:20.GetUser]
	//     code: USER_404
	//     context: {user_id: user-123, source: database}
	//     http_status: 400
	OutputPretty outputFormatOption = iota

	// OutputJSON formats errors as compact JSON.
	// This is the default.
	// Useful for structured logging and machine processing.
	//
	// Example:
	//   {"error":"user not found","caller":"main.go:20","code":"USER_404",...}
	OutputJSON

	// OutputJSONPretty formats errors as indented JSON.
	// Useful for documentation, debugging, and human-readable JSON output.
	//
	// Example:
	//   {
	//     "error": "user not found",
	//     "code": "USER_404",
	//     "caller": "main.go:20"
	//   }
	OutputJSONPretty

	// OutputCompact formats errors as single-line text with key=value pairs.
	// Useful for log aggregation systems.
	//
	// Example:
	//   user not found [main.go:20] code=USER_404 user_id=user-123 http_status=400
	OutputCompact
)

// verbosityOption controls which fields are included in Error() output.
type verbosityOption int

func (verbosityOption) ErrificOption() {}

const (
	// VerbosityMinimal shows only the error message and caller.
	//
	// Example:
	//   user not found [main.go:20.GetUser]
	VerbosityMinimal verbosityOption = iota

	// VerbosityStandard shows message, caller, code, category, and context.
	// Good balance for most applications.
	//
	// Example:
	//   user not found [main.go:20.GetUser]
	//     code: USER_404
	//     category: validation
	//     context: {user_id: user-123}
	VerbosityStandard

	// VerbosityFull shows all non-empty fields (default).
	// Recommended for debugging and development.
	//
	// Example:
	//   user not found [main.go:20.GetUser]
	//     code: USER_404
	//     category: validation
	//     context: {user_id: user-123, source: database}
	//     http_status: 400
	//     retryable: true
	//     correlation_id: trace-123
	//     help: Check if user exists
	VerbosityFull

	// VerbosityCustom allows fine-grained control via individual field flags.
	// Use with Show* and Hide* options.
	VerbosityCustom
)

// Field visibility options for VerbosityCustom.
type fieldVisibilityOption struct {
	field string
	show  bool
}

func (fieldVisibilityOption) ErrificOption() {}

var (
	// ShowCode includes error code in output.
	ShowCode = fieldVisibilityOption{field: "code", show: true}
	// HideCode excludes error code from output.
	HideCode = fieldVisibilityOption{field: "code", show: false}

	// ShowCategory includes error category in output.
	ShowCategory = fieldVisibilityOption{field: "category", show: true}
	// HideCategory excludes error category from output.
	HideCategory = fieldVisibilityOption{field: "category", show: false}

	// ShowContext includes structured context in output.
	ShowContext = fieldVisibilityOption{field: "context", show: true}
	// HideContext excludes structured context from output.
	HideContext = fieldVisibilityOption{field: "context", show: false}

	// ShowHTTPStatus includes HTTP status code in output.
	ShowHTTPStatus = fieldVisibilityOption{field: "http_status", show: true}
	// HideHTTPStatus excludes HTTP status code from output.
	HideHTTPStatus = fieldVisibilityOption{field: "http_status", show: false}

	// ShowRetryMetadata includes retry information (retryable, retry_after, max_retries) in output.
	ShowRetryMetadata = fieldVisibilityOption{field: "retry_metadata", show: true}
	// HideRetryMetadata excludes retry information from output.
	HideRetryMetadata = fieldVisibilityOption{field: "retry_metadata", show: false}

	// ShowMCPData includes MCP-related fields (correlation_id, help, suggestion, etc.) in output.
	ShowMCPData = fieldVisibilityOption{field: "mcp_data", show: true}
	// HideMCPData excludes MCP-related fields from output.
	HideMCPData = fieldVisibilityOption{field: "mcp_data", show: false}

	// ShowTags includes semantic tags in output.
	ShowTags = fieldVisibilityOption{field: "tags", show: true}
	// HideTags excludes semantic tags from output.
	HideTags = fieldVisibilityOption{field: "tags", show: false}

	// ShowLabels includes key-value labels in output.
	ShowLabels = fieldVisibilityOption{field: "labels", show: true}
	// HideLabels excludes key-value labels from output.
	HideLabels = fieldVisibilityOption{field: "labels", show: false}

	// ShowTimestamps includes timestamp and duration in output.
	ShowTimestamps = fieldVisibilityOption{field: "timestamps", show: true}
	// HideTimestamps excludes timestamp and duration from output.
	HideTimestamps = fieldVisibilityOption{field: "timestamps", show: false}
)

var root string
var goroot string

func init() {
	_, file, _, _ := runtime.Caller(0)
	root = fmt.Sprintf("%s/", filepath.Join(filepath.Dir(file), ".."))

	// Try to get GOROOT using "go env GOROOT" first (preferred method)
	if cmd := exec.Command("go", "env", "GOROOT"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			trimmed := strings.TrimSpace(string(output))
			if trimmed != "" {
				goroot = trimmed
				return
			}
		}
	}

	// Fallback to runtime.GOROOT() if command failed
	// Note: runtime.GOROOT() is deprecated but still works as a fallback
	if fallback := runtime.GOROOT(); fallback != "" {
		goroot = fallback
	}
}
