package testutils

import (
	"bytes"

	"github.com/wizact/te-reo-bot/pkg/logger"
)

// SetupTestLogger creates a test logger with a buffer for capturing output
// Returns the logger instance and a buffer to check log output
func SetupTestLogger() (logger.Logger, *bytes.Buffer) {
	var logBuffer bytes.Buffer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
	return testLogger, &logBuffer
}

// SetupGlobalTestLogger sets up a global test logger and returns a cleanup function
// Use this when you need to set the global logger for tests
func SetupGlobalTestLogger() func() {
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "debug",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLogger(config)
	logger.SetGlobalLogger(testLogger)

	return func() {
		logger.ResetGlobalLogger()
	}
}

// SetupTestLoggerWithLevel creates a test logger with a specific log level
func SetupTestLoggerWithLevel(level string) (logger.Logger, *bytes.Buffer) {
	var logBuffer bytes.Buffer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          level,
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)
	return testLogger, &logBuffer
}