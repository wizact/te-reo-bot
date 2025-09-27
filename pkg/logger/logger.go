package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Logger interface for structured logging with stack traces
type Logger interface {
	Error(err error, message string, fields ...Field)
	ErrorWithStack(err error, message string, fields ...Field)
	Fatal(err error, message string, fields ...Field)
	Info(message string, fields ...Field)
	Debug(message string, fields ...Field)
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp  time.Time                `json:"timestamp"`
	Level      string                   `json:"level"`
	Message    string                   `json:"message"`
	Error      string                   `json:"error,omitempty"`
	StackTrace []map[string]interface{} `json:"stack_trace,omitempty"`
	Fields     map[string]interface{}   `json:"fields,omitempty"`
}

// ConcreteLogger is a concrete implementation of the Logger interface
type ConcreteLogger struct {
	config *LoggerConfig
	writer io.Writer
	mutex  sync.Mutex
}

// NewLogger creates a new logger instance with the provided configuration
func NewLogger(config *LoggerConfig) Logger {
	return &ConcreteLogger{
		config: config,
		writer: os.Stdout,
	}
}

// NewLoggerWithWriter creates a new logger instance with custom writer
func NewLoggerWithWriter(config *LoggerConfig, writer io.Writer) Logger {
	return &ConcreteLogger{
		config: config,
		writer: writer,
	}
}

// Error logs an error message with optional structured fields
func (l *ConcreteLogger) Error(err error, message string, fields ...Field) {
	l.log(ErrorLevel, err, message, nil, fields...)
}

// ErrorWithStack logs an error message with stack trace and optional structured fields
func (l *ConcreteLogger) ErrorWithStack(err error, message string, fields ...Field) {
	var stackTrace *StackTrace
	if l.config.IsStackTraceEnabled() {
		stackTrace = CaptureStackTrace(2) // Skip this function call and the log method call
	}
	l.log(ErrorLevel, err, message, stackTrace, fields...)
}

// Fatal logs a fatal error message with stack trace and exits the program
func (l *ConcreteLogger) Fatal(err error, message string, fields ...Field) {
	var stackTrace *StackTrace
	if l.config.IsStackTraceEnabled() {
		stackTrace = CaptureStackTrace(2) // Skip this function call and the log method call
	}
	l.log(FatalLevel, err, message, stackTrace, fields...)
	os.Exit(1)
}

// Info logs an informational message with optional structured fields
func (l *ConcreteLogger) Info(message string, fields ...Field) {
	l.log(InfoLevel, nil, message, nil, fields...)
}

// Debug logs a debug message with optional structured fields
func (l *ConcreteLogger) Debug(message string, fields ...Field) {
	if !l.config.IsDebugEnabled() {
		return // Skip debug logs if not enabled
	}
	l.log(DebugLevel, nil, message, nil, fields...)
}

// log is the internal logging method that handles all log levels
func (l *ConcreteLogger) log(level LogLevel, err error, message string, stackTrace *StackTrace, fields ...Field) {
	// Thread-safe logging
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Create log entry
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     level.String(),
		Message:   message,
	}

	// Add error if provided
	if err != nil {
		entry.Error = err.Error()
	}

	// Add stack trace if provided
	if stackTrace != nil && !stackTrace.IsEmpty() {
		entry.StackTrace = l.formatStackTraceForJSON(stackTrace)
	}

	// Add structured fields if provided
	if len(fields) > 0 {
		entry.Fields = make(map[string]interface{})
		for _, field := range fields {
			entry.Fields[field.Key] = field.Value
		}
	}

	// Format and write the log entry
	var output string
	if l.config.LogFormat == "json" {
		output = l.formatJSON(entry)
	} else {
		output = l.formatText(entry, stackTrace)
	}

	fmt.Fprintln(l.writer, output)
}

// formatJSON formats the log entry as JSON
func (l *ConcreteLogger) formatJSON(entry LogEntry) string {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		return fmt.Sprintf(`{"timestamp":"%s","level":"%s","message":"JSON marshal error: %s","error":"%s"}`,
			entry.Timestamp.Format(time.RFC3339),
			entry.Level,
			err.Error(),
			entry.Message)
	}
	return string(jsonBytes)
}

// formatText formats the log entry as human-readable text
func (l *ConcreteLogger) formatText(entry LogEntry, stackTrace *StackTrace) string {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")

	// Start with basic log line
	output := fmt.Sprintf("[%s] %s: %s", timestamp, entry.Level, entry.Message)

	// Add error if present
	if entry.Error != "" {
		output += fmt.Sprintf(" | Error: %s", entry.Error)
	}

	// Add fields if present
	if len(entry.Fields) > 0 {
		output += " | Fields: "
		first := true
		for key, value := range entry.Fields {
			if !first {
				output += ", "
			}
			output += fmt.Sprintf("%s=%v", key, value)
			first = false
		}
	}

	// Add stack trace if present (for text format, use the formatted string)
	if stackTrace != nil && !stackTrace.IsEmpty() {
		output += "\n" + stackTrace.Format()
	}

	return output
}

// formatStackTraceForJSON converts a StackTrace to JSON-friendly format
func (l *ConcreteLogger) formatStackTraceForJSON(stackTrace *StackTrace) []map[string]interface{} {
	if stackTrace == nil || stackTrace.IsEmpty() {
		return nil
	}

	frames := make([]map[string]interface{}, len(stackTrace.Frames))
	for i, frame := range stackTrace.Frames {
		frames[i] = map[string]interface{}{
			"function": frame.Function,
			"file":     frame.File,
			"line":     frame.Line,
		}
	}
	return frames
}
