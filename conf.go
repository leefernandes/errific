package errific

import (
	"fmt"
	"path/filepath"
	"runtime"
)

// Configure errific options.
func Configure(opts ...Option) {
	for _, opt := range opts {
		switch o := opt.(type) {
		case callerOption:
			c.caller = o

		case layoutOption:
			c.layout = o
		}
	}
}

var c struct {
	// Caller will configure the caller: Suffix|Prefix|Disabled.
	// Default is Suffix.
	caller callerOption
	// Layout will configure the layout of wrapped errors: Newline|Inline.
	// Default is Newline.
	layout layoutOption
}

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

type Option interface {
	ErrificOption()
}

var root string

func init() {
	// TODO find root of project, to strip from frame file.
	_, file, _, _ := runtime.Caller(0)
	root = fmt.Sprintf("%s/", filepath.Join(filepath.Dir(file), ".."))
}
