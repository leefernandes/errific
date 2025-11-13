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
	"errors"
	"fmt"
	"runtime"
	"strings"
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

type errific struct {
	err    error   // primary error.
	errs   []error // errors used in string output, and satisfy errors.Is.
	unwrap []error // errors not used in string output, but satisfy errors.Is.
	caller string  // caller information.
	stack  []byte  // optional stack buffer.
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

func (e errific) Unwrap() []error {
	var errs []error
	if e.err != nil {
		errs = append(errs, e.err)
	}
	errs = append(errs, e.errs...)
	errs = append(errs, e.unwrap...)
	return errs
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
		if !strings.HasPrefix(frame.File, goroot) {
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
