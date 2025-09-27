package logger

import (
	"fmt"
	"runtime"
	"strings"
)

// StackTrace represents a captured stack trace with multiple frames
type StackTrace struct {
	Frames []Frame
}

// Frame represents a single stack frame with function, file, and line information
type Frame struct {
	Function string
	File     string
	Line     int
}

// CaptureStackTrace captures the current stack trace, skipping the specified number of frames
// skip parameter allows skipping frames (e.g., skip=1 to skip the CaptureStackTrace call itself)
func CaptureStackTrace(skip int) *StackTrace {
	// Add 1 to skip to account for this function call
	skip++

	// Get program counters for the stack trace
	pcs := make([]uintptr, 32) // Capture up to 32 frames
	n := runtime.Callers(skip, pcs)

	if n == 0 {
		return &StackTrace{Frames: []Frame{}}
	}

	// Get frames from program counters
	frames := runtime.CallersFrames(pcs[:n])

	var stackFrames []Frame
	for {
		frame, more := frames.Next()

		stackFrames = append(stackFrames, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})

		if !more {
			break
		}
	}

	return &StackTrace{Frames: stackFrames}
}

// Format formats the stack trace into a readable string representation
func (st *StackTrace) Format() string {
	if len(st.Frames) == 0 {
		return "no stack trace available"
	}

	var builder strings.Builder
	builder.WriteString("Stack trace:\n")

	for i, frame := range st.Frames {
		builder.WriteString(fmt.Sprintf("  %d. %s\n", i+1, frame.Function))
		builder.WriteString(fmt.Sprintf("     %s:%d\n", frame.File, frame.Line))
	}

	return builder.String()
}

// String implements the Stringer interface for StackTrace
func (st *StackTrace) String() string {
	return st.Format()
}

// IsEmpty returns true if the stack trace has no frames
func (st *StackTrace) IsEmpty() bool {
	return len(st.Frames) == 0
}
