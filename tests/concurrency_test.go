package errific

import (
	"errors"
	"strings"
	"sync"
	"testing"

	. "github.com/leefernandes/errific"
)

// ============================================================================
// Concurrent Getter Tests
// ============================================================================

func TestConcurrent_Getters(t *testing.T) {
	Configure()
	var ErrTest Err = "concurrent test"

	err := ErrTest.New().
		WithCode("ERR_001").
		WithCategory(CategoryServer).
		WithCorrelationID("corr-123").
		WithTags("tag1", "tag2").
		WithLabels(map[string]string{"k1": "v1"})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Read operations should be safe
			_ = GetCode(err)
			_ = GetCategory(err)
			_ = GetCorrelationID(err)
			_ = GetTags(err)
			_ = GetLabels(err)
			_ = err.Error()
		}()
	}
	wg.Wait()
}

// ============================================================================
// Concurrent Configure Tests
// ============================================================================

func TestConcurrent_ConfigureAndCreate(t *testing.T) {
	var ErrTest Err = "concurrent test"
	var wg sync.WaitGroup

	// Multiple goroutines configuring and creating errors
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				Configure(Suffix, Newline)
			} else {
				Configure(Prefix, Inline)
			}
			_ = ErrTest.New()
		}(i)
	}
	wg.Wait()

	Configure() // Reset
}

// ============================================================================
// Race Condition Tests
// ============================================================================

func TestRaceCondition_ConfigurationSnapshot(t *testing.T) {
	var ErrTest Err = "test error"

	t.Run("error formatting consistent after Configure", func(t *testing.T) {
		// Create error with Suffix config
		Configure(Suffix)
		err := ErrTest.New()

		// Change configuration
		Configure(Prefix)

		// Error should still use Suffix (snapshot at creation time)
		msg := err.Error()
		if !strings.HasSuffix(msg, "]") {
			t.Errorf("Expected suffix format (ends with ]), got: %s", msg)
		}
		if strings.HasPrefix(msg, "[") {
			t.Errorf("Expected suffix format (not prefix), got: %s", msg)
		}
	})

	t.Run("concurrent Configure and Error calls", func(t *testing.T) {
		// This test should pass race detector
		Configure(Suffix, Newline)

		var wg sync.WaitGroup
		errors := make([]error, 100)

		// Create errors concurrently
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				errors[idx] = ErrTest.New()
			}(i)
		}

		// Concurrently change configuration
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				if n%2 == 0 {
					Configure(Prefix, Inline)
				} else {
					Configure(Suffix, Newline)
				}
			}(i)
		}

		wg.Wait()

		// All errors should format successfully without panics
		for i, err := range errors {
			if err == nil {
				continue
			}
			msg := err.Error()
			if msg == "" {
				t.Errorf("Error %d has empty message", i)
			}
		}
	})

	t.Run("stack config snapshot works", func(t *testing.T) {
		// Create error without stack
		Configure()
		err1 := ErrTest.New()

		// Enable stack
		Configure(WithStack)
		err2 := ErrTest.New()

		// Disable stack again
		Configure()
		err3 := ErrTest.New()

		// Each error should use its creation-time config
		msg1 := err1.Error()
		msg2 := err2.Error()
		msg3 := err3.Error()

		// err1 should not have stack (created without WithStack)
		if strings.Contains(msg1, "\n  ") {
			t.Error("err1 should not have stack trace")
		}

		// err2 should have stack (created with WithStack)
		if !strings.Contains(msg2, "concurrency_test.go") {
			t.Error("err2 should have stack trace")
		}

		// err3 should not have stack (created without WithStack again)
		if strings.Contains(msg3, "\n  ") {
			t.Error("err3 should not have stack trace")
		}
	})

	t.Run("layout config snapshot works", func(t *testing.T) {
		// Create error with Newline layout
		Configure(Newline)
		err1 := ErrTest.New(errors.New("wrapped1"), errors.New("wrapped2"))

		// Change to Inline
		Configure(Inline)
		err2 := ErrTest.New(errors.New("wrapped1"), errors.New("wrapped2"))

		// err1 should use newlines
		msg1 := err1.Error()
		if !strings.Contains(msg1, "\n") {
			t.Error("err1 should use newline layout")
		}
		if strings.Contains(msg1, "↩") {
			t.Error("err1 should not use inline symbol")
		}

		// err2 should use inline symbol
		msg2 := err2.Error()
		if !strings.Contains(msg2, "↩") {
			t.Error("err2 should use inline layout symbol")
		}
	})
}

// ============================================================================
// Immutability Tests
// ============================================================================

func TestImmutability_NoMutation(t *testing.T) {
	Configure()
	var ErrTest Err = "test"

	err1 := ErrTest.New().WithCode("CODE1")
	err2 := err1.WithCode("CODE2") // Should create new error, not mutate

	// Original should be unchanged
	if GetCode(err1) != "CODE1" {
		t.Errorf("Original error should be unchanged, got %s", GetCode(err1))
	}
	if GetCode(err2) != "CODE2" {
		t.Errorf("New error should have CODE2, got %s", GetCode(err2))
	}
}

func TestImmutability_MultipleConfigureCalls(t *testing.T) {
	// Test that errors capture config at creation time
	Configure(Suffix, Newline)
	var ErrTest Err = "test"

	err1 := ErrTest.New()

	Configure(Prefix, Inline)
	err2 := ErrTest.New()

	Configure(Disabled)
	err3 := ErrTest.New()

	// Each error should use its creation-time config
	msg1 := err1.Error()
	msg2 := err2.Error()
	msg3 := err3.Error()

	// err1: Suffix format (ends with ])
	if !strings.HasSuffix(msg1, "]") {
		t.Errorf("err1 should use Suffix format, got: %s", msg1)
	}

	// err2: Prefix format (starts with [)
	if !strings.HasPrefix(msg2, "[") {
		t.Errorf("err2 should use Prefix format, got: %s", msg2)
	}

	// err3: Disabled (no brackets)
	if strings.Contains(msg3, "[") || strings.Contains(msg3, "]") {
		t.Errorf("err3 should have no caller info, got: %s", msg3)
	}
}
