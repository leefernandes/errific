package errific

import (
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestErrNew(t *testing.T) {
	Configure()

	t.Run("basic error", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New()

		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}

		if !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected error message to contain 'test error', got: %s", err.Error())
		}

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New(io.EOF)

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !strings.Contains(err.Error(), "EOF") {
			t.Errorf("expected error message to contain 'EOF', got: %s", err.Error())
		}
	})

	t.Run("with multiple wrapped errors", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Error("expected errors.Is to match io.ErrUnexpectedEOF")
		}
	})
}

func TestErrErrorf(t *testing.T) {
	Configure()

	t.Run("formatted error", func(t *testing.T) {
		var ErrTest Err = "test error: %s %d"
		err := ErrTest.Errorf("hello", 42)

		if !strings.Contains(err.Error(), "hello") {
			t.Errorf("expected error message to contain 'hello', got: %s", err.Error())
		}

		if !strings.Contains(err.Error(), "42") {
			t.Errorf("expected error message to contain '42', got: %s", err.Error())
		}

		if !errors.Is(err, ErrTest) {
			t.Error("expected errors.Is to match ErrTest")
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		var ErrTest Err = "test error: %w"
		err := ErrTest.Errorf(io.EOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}
	})
}

func TestErrWithf(t *testing.T) {
	Configure()

	t.Run("basic withf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Withf("detail: %s", "info")

		if !strings.Contains(err.Error(), "test error") {
			t.Errorf("expected error message to contain 'test error', got: %s", err.Error())
		}

		if !strings.Contains(err.Error(), "detail: info") {
			t.Errorf("expected error message to contain 'detail: info', got: %s", err.Error())
		}
	})

	t.Run("chained withf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Withf("first %d", 1).Withf("second %d", 2)

		msg := err.Error()
		if !strings.Contains(msg, "first 1") {
			t.Errorf("expected error message to contain 'first 1', got: %s", msg)
		}

		if !strings.Contains(msg, "second 2") {
			t.Errorf("expected error message to contain 'second 2', got: %s", msg)
		}
	})
}

func TestErrWrapf(t *testing.T) {
	Configure()

	t.Run("basic wrapf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Wrapf("wrapped: %w", io.EOF)

		if !errors.Is(err, io.EOF) {
			t.Error("expected errors.Is to match io.EOF")
		}

		if !strings.Contains(err.Error(), "wrapped") {
			t.Errorf("expected error message to contain 'wrapped', got: %s", err.Error())
		}
	})

	t.Run("chained wrapf", func(t *testing.T) {
		var ErrTest Err = "test error"
		err := ErrTest.Wrapf("first %d", 1).Wrapf("second %d", 2)

		msg := err.Error()
		if !strings.Contains(msg, "first 1") {
			t.Errorf("expected error message to contain 'first 1', got: %s", msg)
		}

		if !strings.Contains(msg, "second 2") {
			t.Errorf("expected error message to contain 'second 2', got: %s", msg)
		}
	})
}

func TestErrificJoin(t *testing.T) {
	Configure()

	var ErrTest Err = "test error"
	err := ErrTest.New().Join(io.EOF, io.ErrUnexpectedEOF)

	if !errors.Is(err, io.EOF) {
		t.Error("expected errors.Is to match io.EOF")
	}

	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Error("expected errors.Is to match io.ErrUnexpectedEOF")
	}
}

func TestConfigureCallerOption(t *testing.T) {
	t.Run("suffix", func(t *testing.T) {
		Configure(Suffix)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should end with [location]
		if !strings.Contains(msg, "[") || !strings.HasSuffix(msg, "]") {
			t.Errorf("expected suffix format, got: %s", msg)
		}

		if strings.HasPrefix(msg, "[") {
			t.Errorf("expected suffix not prefix, got: %s", msg)
		}
	})

	t.Run("prefix", func(t *testing.T) {
		Configure(Prefix)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should start with [location]
		if !strings.HasPrefix(msg, "[") {
			t.Errorf("expected prefix format, got: %s", msg)
		}
	})

	t.Run("disabled", func(t *testing.T) {
		Configure(Disabled)

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should not contain brackets
		if strings.Contains(msg, "[") || strings.Contains(msg, "]") {
			t.Errorf("expected no caller info, got: %s", msg)
		}
	})
}

func TestConfigureLayoutOption(t *testing.T) {
	t.Run("newline", func(t *testing.T) {
		Configure(Newline)

		var ErrTest Err = "test"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)
		msg := err.Error()

		// Should contain newlines
		if !strings.Contains(msg, "\n") {
			t.Errorf("expected newline layout, got: %s", msg)
		}
	})

	t.Run("inline", func(t *testing.T) {
		Configure(Inline)

		var ErrTest Err = "test"
		err := ErrTest.New(io.EOF, io.ErrUnexpectedEOF)
		msg := err.Error()

		// Should contain ↩ symbol
		if !strings.Contains(msg, "↩") {
			t.Errorf("expected inline layout with ↩, got: %s", msg)
		}

		// Should not contain newlines (except maybe in caller/stack)
		lines := strings.Split(msg, "\n")
		if len(lines) > 2 { // Allow for potential stack traces
			t.Errorf("expected inline layout with minimal newlines, got: %s", msg)
		}
	})
}

func TestConfigureWithStack(t *testing.T) {
	t.Run("with stack", func(t *testing.T) {
		Configure(WithStack)

		var ErrTest Err = "test"

		// Create error in a helper function to ensure stack has frames
		err := helperFunctionForStack(ErrTest)
		msg := err.Error()

		// The error message should still be valid
		if !strings.Contains(msg, "test") {
			t.Errorf("expected error message to contain 'test', got: %s", msg)
		}

		// WithStack configuration should not cause errors
		if msg == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("without stack", func(t *testing.T) {
		Configure() // Default is without stack

		var ErrTest Err = "test"
		err := ErrTest.New()
		msg := err.Error()

		// Should be a simple error message
		if msg == "" {
			t.Error("expected non-empty error message")
		}

		if !strings.Contains(msg, "test") {
			t.Errorf("expected error message to contain 'test', got: %s", msg)
		}
	})
}

// Helper function to create errors with a deeper stack
func helperFunctionForStack(e Err) errific {
	return e.New(io.EOF)
}

func TestConfigureTrimPrefixes(t *testing.T) {
	Configure(TrimPrefixes("/usr/local/go/", "/home/user/"))

	var ErrTest Err = "test"
	err := ErrTest.New()
	msg := err.Error()

	// Should not contain the trimmed prefixes
	if strings.Contains(msg, "/usr/local/go/") {
		t.Errorf("expected trimmed prefix, got: %s", msg)
	}

	if strings.Contains(msg, "/home/user/") {
		t.Errorf("expected trimmed prefix, got: %s", msg)
	}
}

func TestConfigureTrimCWD(t *testing.T) {
	Configure(TrimCWD)

	var ErrTest Err = "test"
	err := ErrTest.New()
	msg := err.Error()

	// Should have relative paths
	if !strings.Contains(msg, "errific") {
		t.Errorf("expected relative path, got: %s", msg)
	}
}

func TestConcurrentConfigure(t *testing.T) {
	// Test that concurrent Configure calls don't cause races
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			switch n % 4 {
			case 0:
				Configure(Suffix)
			case 1:
				Configure(Prefix)
			case 2:
				Configure(Disabled)
			case 3:
				Configure(Newline)
			}
		}(i)
	}

	wg.Wait()
}

func TestConcurrentErrorCreation(t *testing.T) {
	Configure()

	var ErrTest Err = "test"
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := ErrTest.New(io.EOF)
			if !errors.Is(err, ErrTest) {
				t.Error("expected errors.Is to match ErrTest")
			}

			_ = err.Error()
		}()
	}

	wg.Wait()
}

func TestUnwrap(t *testing.T) {
	Configure()

	var (
		Err1 Err = "error 1"
		Err2 Err = "error 2"
	)

	err1 := Err1.New(io.EOF)
	err2 := Err2.New(err1)

	// Test that unwrap chain works
	if !errors.Is(err2, Err2) {
		t.Error("expected errors.Is to match Err2")
	}

	if !errors.Is(err2, Err1) {
		t.Error("expected errors.Is to match Err1")
	}

	if !errors.Is(err2, io.EOF) {
		t.Error("expected errors.Is to match io.EOF")
	}
}

func TestCircularReferenceFixed(t *testing.T) {
	Configure()

	var ErrTest Err = "test"
	err := ErrTest.Withf("detail %d", 1)

	// This should not cause infinite loop
	msg := err.Error()

	if msg == "" {
		t.Error("expected non-empty error message")
	}

	// Make sure the error chain is valid
	// errific.Unwrap() returns []error, so we can't use errors.Unwrap
	// Instead, verify that errors.Is works properly
	if !errors.Is(err, ErrTest) {
		t.Error("expected errors.Is to match ErrTest")
	}
}

func BenchmarkErrNew(b *testing.B) {
	Configure()
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New()
	}
}

func BenchmarkErrNewWithWrap(b *testing.B) {
	Configure()
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New(io.EOF)
	}
}

func BenchmarkErrError(b *testing.B) {
	Configure()
	var ErrTest Err = "test error"
	err := ErrTest.New(io.EOF)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrWithStack(b *testing.B) {
	Configure(WithStack)
	var ErrTest Err = "test error"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ErrTest.New()
	}
}
