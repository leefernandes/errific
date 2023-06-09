package errific

import (
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
func (e Err) New(errs ...error) error {
	return &errible{
		err:    e,
		errs:   errs,
		caller: caller(),
	}
}

// Errorf returns an error using Err formatted as text.
// Use Errorf if your Err string itself contains fmt format specifiers.
//
//	var ErrProcessThing errific.Err = "error processing thing id: '%s'"
//
//	return ErrProcessThing.Errorf("abc")
func (e Err) Errorf(a ...any) error {
	return &errible{
		err:    fmt.Errorf(e.Error(), a...),
		caller: caller(),
		unwrap: []error{e},
	}
}

// Withf returns an error with a formatted string inline to Err as text.
//
//	var ErrProcessThing errific.Err = "error processing thing"
//
//	return ErrProcessThing.Withf("id: '%s'", "abc")
func (e Err) Withf(format string, a ...any) (err error) {
	format = fmt.Sprintf("%s: %s", e.Error(), format)
	return &errible{
		err:    fmt.Errorf(format, a...),
		caller: caller(),
		unwrap: []error{e},
	}
}

// Wrapf return an error using Err as text and wraps a formatted error.
// Use Wrapf to format an error and wrap it.
//
//	var ErrProcessThing errific.Err = "error processing thing"
//
//	return ErrProcessThing.Wrapf("cause: %w", err)
func (e Err) Wrapf(format string, a ...any) error {
	return &errible{
		err:    e,
		errs:   []error{fmt.Errorf(format, a...)},
		caller: caller(),
	}
}

func (e Err) Error() string {
	return string(e)
}

type errible struct {
	err    error   // primary error.
	errs   []error // errors used in string output, and satisfy errors.Is.
	unwrap []error // errors not used in string output, but satisfy errors.Is.
	caller string  // caller information.
}

func (e errible) Error() (msg string) {
	switch c.caller {
	case Disabled:

	case Prefix:
		msg = fmt.Sprintf("[%s] %s", e.caller, e.err.Error())

	default:
		msg = fmt.Sprintf("%s [%s]", e.err.Error(), e.caller)
	}

	switch c.layout {
	case Inline:
		for i := range e.errs {
			msg = fmt.Sprintf("%s ↩ %s", msg, e.errs[i].Error())
		}

	default:
		for i := range e.errs {
			msg = fmt.Sprintf("%s\n%s", msg, e.errs[i].Error())
		}
	}

	return msg
}

func (e errible) Unwrap() []error {
	var errs []error
	if e.err != nil {
		errs = append(errs, e.err)
	}
	errs = append(errs, e.errs...)
	errs = append(errs, e.unwrap...)
	return errs
}

func caller() string {
	pc := make([]uintptr, 1)
	n := runtime.Callers(3, pc)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	funcParts := strings.Split(frame.Function, "/")
	funcParts = strings.Split(funcParts[len(funcParts)-1], ".")
	callFunc := funcParts[len(funcParts)-1]
	callFile := strings.TrimPrefix(frame.File, root)
	callLine := frame.Line
	return fmt.Sprintf("%s:%d.%s", callFile, callLine, callFunc)
}
