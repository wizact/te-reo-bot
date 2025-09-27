package logger

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureStackTrace(t *testing.T) {
	t.Run("captures stack trace with correct number of frames", func(t *testing.T) {
		st := captureStackTraceHelper()

		assert.NotNil(t, st)
		assert.Greater(t, len(st.Frames), 0, "should capture at least one frame")

		// Should contain this test function and the helper function
		assert.Greater(t, len(st.Frames), 1, "should capture multiple frames")
	})

	t.Run("captures correct function names", func(t *testing.T) {
		st := captureStackTraceHelper()

		// One of the frames should contain the helper function
		foundHelper := false
		foundTest := false
		for _, frame := range st.Frames {
			if strings.Contains(frame.Function, "captureStackTraceHelper") {
				foundHelper = true
			}
			if strings.Contains(frame.Function, "TestCaptureStackTrace") {
				foundTest = true
			}
		}
		assert.True(t, foundHelper, "should contain the helper function in stack trace")
		assert.True(t, foundTest, "should contain the test function in stack trace")
	})

	t.Run("captures file and line information", func(t *testing.T) {
		st := captureStackTraceHelper()

		for _, frame := range st.Frames {
			assert.NotEmpty(t, frame.File, "file should not be empty")
			assert.Greater(t, frame.Line, 0, "line number should be positive")
		}
	})

	t.Run("skip parameter works correctly", func(t *testing.T) {
		// Capture with skip=0 (should include CaptureStackTrace itself)
		st1 := CaptureStackTrace(0)

		// Capture with skip=1 (should skip CaptureStackTrace)
		st2 := CaptureStackTrace(1)

		// st2 should have one less frame than st1
		assert.Equal(t, len(st1.Frames)-1, len(st2.Frames))
	})
}

func TestStackTraceFormat(t *testing.T) {
	t.Run("formats stack trace correctly", func(t *testing.T) {
		st := &StackTrace{
			Frames: []Frame{
				{
					Function: "github.com/wizact/te-reo-bot/pkg/logger.TestFunction",
					File:     "/app/pkg/logger/test.go",
					Line:     42,
				},
				{
					Function: "github.com/wizact/te-reo-bot/pkg/handlers.ServeHTTP",
					File:     "/app/pkg/handlers/http-server.go",
					Line:     123,
				},
			},
		}

		formatted := st.Format()

		assert.Contains(t, formatted, "Stack trace:")
		assert.Contains(t, formatted, "1. github.com/wizact/te-reo-bot/pkg/logger.TestFunction")
		assert.Contains(t, formatted, "/app/pkg/logger/test.go:42")
		assert.Contains(t, formatted, "2. github.com/wizact/te-reo-bot/pkg/handlers.ServeHTTP")
		assert.Contains(t, formatted, "/app/pkg/handlers/http-server.go:123")
	})

	t.Run("handles empty stack trace", func(t *testing.T) {
		st := &StackTrace{Frames: []Frame{}}

		formatted := st.Format()

		assert.Equal(t, "no stack trace available", formatted)
	})

	t.Run("string method works", func(t *testing.T) {
		st := &StackTrace{
			Frames: []Frame{
				{
					Function: "test.function",
					File:     "/test/file.go",
					Line:     10,
				},
			},
		}

		stringOutput := st.String()
		formatOutput := st.Format()

		assert.Equal(t, formatOutput, stringOutput)
	})
}

func TestStackTraceIsEmpty(t *testing.T) {
	t.Run("returns true for empty stack trace", func(t *testing.T) {
		st := &StackTrace{Frames: []Frame{}}
		assert.True(t, st.IsEmpty())
	})

	t.Run("returns false for non-empty stack trace", func(t *testing.T) {
		st := &StackTrace{
			Frames: []Frame{
				{Function: "test", File: "test.go", Line: 1},
			},
		}
		assert.False(t, st.IsEmpty())
	})
}

func TestFrame(t *testing.T) {
	t.Run("frame contains expected fields", func(t *testing.T) {
		frame := Frame{
			Function: "github.com/wizact/te-reo-bot/pkg/logger.TestFunction",
			File:     "/app/pkg/logger/test.go",
			Line:     42,
		}

		assert.Equal(t, "github.com/wizact/te-reo-bot/pkg/logger.TestFunction", frame.Function)
		assert.Equal(t, "/app/pkg/logger/test.go", frame.File)
		assert.Equal(t, 42, frame.Line)
	})
}

// Helper function to test stack trace capture
func captureStackTraceHelper() *StackTrace {
	return CaptureStackTrace(0)
}

// Integration test that demonstrates real stack trace capture
func TestStackTraceIntegration(t *testing.T) {
	t.Run("captures real stack trace in nested calls", func(t *testing.T) {
		st := nestedFunction1()

		assert.NotNil(t, st)
		assert.Greater(t, len(st.Frames), 2, "should capture multiple nested frames")

		// Verify the stack contains our nested functions
		formatted := st.Format()
		assert.Contains(t, formatted, "nestedFunction")
	})
}

func nestedFunction1() *StackTrace {
	return nestedFunction2()
}

func nestedFunction2() *StackTrace {
	return nestedFunction3()
}

func nestedFunction3() *StackTrace {
	return CaptureStackTrace(0)
}
