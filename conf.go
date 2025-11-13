package errific

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		}
	}

	if c.trimCWD {
		cwd, err := os.Getwd()
		if err != nil {
			// Fallback to not trimming CWD if we can't get it
			c.trimCWD = false
			return
		}

		c.trimPrefixes = append([]string{filepath.Dir(cwd) + "/"}, c.trimPrefixes...)
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

var root string

func init() {
	_, file, _, _ := runtime.Caller(0)
	root = fmt.Sprintf("%s/", filepath.Join(filepath.Dir(file), ".."))
}
