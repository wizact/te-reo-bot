package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	logger := NewLogger(config)
	assert.NotNil(t, logger)

	// Verify it's a ConcreteLogger
	concreteLogger, ok := logger.(*ConcreteLogger)
	assert.True(t, ok)
	assert.Equal(t, config, concreteLogger.config)
}

func TestNewLoggerWithWriter(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	assert.NotNil(t, logger)

	// Test that it writes to the custom writer
	logger.Info("test message")
	assert.Contains(t, buffer.String(), "test message")
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{FatalLevel, "fatal"},
		{LogLevel(999), "unknown"},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, test.level.String())
	}
}

func TestConcreteLogger_Info_JSON(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	logger.Info("test info message", String("key1", "value1"), Int("key2", 42))

	output := buffer.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "info")

	// Parse JSON to verify structure
	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	assert.Equal(t, "info", entry.Level)
	assert.Equal(t, "test info message", entry.Message)
	assert.Empty(t, entry.Error)
	assert.Nil(t, entry.StackTrace)
	assert.Equal(t, "value1", entry.Fields["key1"])
	assert.Equal(t, float64(42), entry.Fields["key2"]) // JSON unmarshals numbers as float64
}

func TestConcreteLogger_Info_Text(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "text",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	logger.Info("test info message", String("key1", "value1"), Int("key2", 42))

	output := buffer.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "key1=value1")
	assert.Contains(t, output, "key2=42")
}

func TestConcreteLogger_Debug_Enabled(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	logger.Debug("debug message")

	output := buffer.String()
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "debug")
}

func TestConcreteLogger_Debug_Disabled(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info", // Debug disabled
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	logger.Debug("debug message")

	output := buffer.String()
	assert.Empty(t, output) // Should not log debug when level is info
}

func TestConcreteLogger_Error_WithoutStackTrace(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	testErr := errors.New("test error")
	logger.Error(testErr, "error occurred", String("context", "test"))

	output := buffer.String()
	assert.Contains(t, output, "error occurred")
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "error")

	// Parse JSON to verify structure
	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	assert.Equal(t, "error", entry.Level)
	assert.Equal(t, "error occurred", entry.Message)
	assert.Equal(t, "test error", entry.Error)
	assert.Nil(t, entry.StackTrace) // No stack trace when disabled
	assert.Equal(t, "test", entry.Fields["context"])
}

func TestConcreteLogger_ErrorWithStack(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	testErr := errors.New("test error with stack")
	logger.ErrorWithStack(testErr, "error with stack occurred")

	output := buffer.String()
	assert.Contains(t, output, "error with stack occurred")
	assert.Contains(t, output, "test error with stack")

	// Parse JSON to verify structure
	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	assert.Equal(t, "error", entry.Level)
	assert.Equal(t, "error with stack occurred", entry.Message)
	assert.Equal(t, "test error with stack", entry.Error)
	assert.NotNil(t, entry.StackTrace) // Should have stack trace
	assert.Greater(t, len(entry.StackTrace), 0)

	// Verify stack trace structure
	frame := entry.StackTrace[0]
	assert.Contains(t, frame["function"], "TestConcreteLogger_ErrorWithStack")
	assert.Contains(t, frame["file"], "logger_test.go")
	// JSON unmarshaling converts numbers to float64
	lineNum, ok := frame["line"].(float64)
	assert.True(t, ok, "line should be a number")
	assert.Greater(t, lineNum, float64(0))
}

func TestConcreteLogger_ErrorWithStack_Disabled(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false, // Stack traces disabled
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	testErr := errors.New("test error")
	logger.ErrorWithStack(testErr, "error occurred")

	output := buffer.String()

	// Parse JSON to verify structure
	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	assert.Nil(t, entry.StackTrace) // No stack trace when disabled
}

func TestConcreteLogger_TextFormat_WithStackTrace(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "text",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	testErr := errors.New("test error")
	logger.ErrorWithStack(testErr, "error with stack")

	output := buffer.String()
	assert.Contains(t, output, "error with stack")
	assert.Contains(t, output, "test error")
	assert.Contains(t, output, "Stack trace:")
	assert.Contains(t, output, "TestConcreteLogger_TextFormat_WithStackTrace")
}

func TestConcreteLogger_ThreadSafety(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	// Run multiple goroutines logging simultaneously
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent log", Int("goroutine", id))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buffer.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 10, len(lines)) // Should have exactly 10 log lines

	// Verify each line is valid JSON
	for _, line := range lines {
		var entry LogEntry
		err := json.Unmarshal([]byte(line), &entry)
		assert.NoError(t, err, "Line should be valid JSON: %s", line)
		assert.Equal(t, "concurrent log", entry.Message)
	}
}

func TestConcreteLogger_JSONMarshalError(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	// Create a field with a value that can't be marshaled to JSON
	unmarshalableValue := make(chan int) // channels can't be marshaled to JSON
	logger.Info("test message", Any("bad_field", unmarshalableValue))

	output := buffer.String()
	assert.Contains(t, output, "JSON marshal error")
}

func TestConcreteLogger_EmptyFields(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	logger.Info("message without fields")

	output := buffer.String()

	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	assert.Nil(t, entry.Fields) // Should be nil when no fields provided
}

func TestConcreteLogger_TimestampFormat(t *testing.T) {
	config := &LoggerConfig{
		EnableStackTraces: false,
		LogLevel:          "info",
		Environment:       "test",
		LogFormat:         "json",
	}

	var buffer bytes.Buffer
	logger := NewLoggerWithWriter(config, &buffer)

	beforeLog := time.Now().UTC()
	logger.Info("timestamp test")
	afterLog := time.Now().UTC()

	output := buffer.String()

	var entry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &entry)
	require.NoError(t, err)

	// Verify timestamp is within expected range and in UTC
	assert.True(t, entry.Timestamp.After(beforeLog.Add(-time.Second)))
	assert.True(t, entry.Timestamp.Before(afterLog.Add(time.Second)))
	assert.Equal(t, time.UTC, entry.Timestamp.Location())
}
