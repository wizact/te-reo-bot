package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeGlobalLogger(t *testing.T) {
	// Reset global logger before each test
	defer ResetGlobalLogger()

	tests := []struct {
		name           string
		config         *LoggerConfig
		expectError    bool
		expectedLevel  string
		expectedFormat string
	}{
		{
			name: "Initialize with provided config",
			config: &LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "debug",
				Environment:       "dev",
				LogFormat:         "json",
			},
			expectError:    false,
			expectedLevel:  "debug",
			expectedFormat: "json",
		},
		{
			name:           "Initialize with nil config (load from environment)",
			config:         nil,
			expectError:    false,
			expectedLevel:  "info", // default
			expectedFormat: "json", // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger for each test
			ResetGlobalLogger()

			err := InitializeGlobalLogger(tt.config)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify global logger was initialized
				logger := GetGlobalLogger()
				assert.NotNil(t, logger)

				// Test that the logger works
				var logBuffer bytes.Buffer
				if concreteLogger, ok := logger.(*ConcreteLogger); ok {
					concreteLogger.writer = &logBuffer
					concreteLogger.Info("test message")

					logOutput := logBuffer.String()
					assert.Contains(t, logOutput, "test message")
					assert.Contains(t, logOutput, "info")
				}
			}
		})
	}
}

func TestGetGlobalLogger(t *testing.T) {
	// Reset global logger before test
	defer ResetGlobalLogger()

	t.Run("Get logger when not initialized", func(t *testing.T) {
		ResetGlobalLogger()

		logger := GetGlobalLogger()
		assert.NotNil(t, logger)

		// Test that the logger works
		var logBuffer bytes.Buffer
		if concreteLogger, ok := logger.(*ConcreteLogger); ok {
			concreteLogger.writer = &logBuffer
			concreteLogger.Info("test message")

			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, "test message")
		}
	})

	t.Run("Get logger when already initialized", func(t *testing.T) {
		ResetGlobalLogger()

		// Initialize first
		config := &LoggerConfig{
			EnableStackTraces: false,
			LogLevel:          "error",
			Environment:       "prod",
			LogFormat:         "text",
		}
		err := InitializeGlobalLogger(config)
		assert.NoError(t, err)

		// Get logger
		logger1 := GetGlobalLogger()
		logger2 := GetGlobalLogger()

		// Should be the same instance
		assert.Equal(t, logger1, logger2)
	})
}

func TestSetGlobalLogger(t *testing.T) {
	// Reset global logger before test
	defer ResetGlobalLogger()

	t.Run("Set custom logger", func(t *testing.T) {
		ResetGlobalLogger()

		// Create custom logger
		config := &LoggerConfig{
			EnableStackTraces: true,
			LogLevel:          "debug",
			Environment:       "test",
			LogFormat:         "json",
		}
		customLogger := NewLogger(config)

		// Set as global logger
		SetGlobalLogger(customLogger)

		// Verify it's set
		globalLogger := GetGlobalLogger()
		assert.Equal(t, customLogger, globalLogger)
	})
}

func TestEnvironmentSpecificConfiguration(t *testing.T) {
	// Reset global logger before test
	defer ResetGlobalLogger()

	tests := []struct {
		name        string
		envVars     map[string]string
		expectedEnv string
		expectedLog string
	}{
		{
			name: "Development environment",
			envVars: map[string]string{
				"ENVIRONMENT": "dev",
				"LOG_LEVEL":   "debug",
			},
			expectedEnv: "dev",
			expectedLog: "debug",
		},
		{
			name: "Production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "prod",
				"LOG_LEVEL":   "info",
			},
			expectedEnv: "prod",
			expectedLog: "info",
		},
		{
			name: "Test environment",
			envVars: map[string]string{
				"ENVIRONMENT": "test",
				"LOG_LEVEL":   "error",
			},
			expectedEnv: "test",
			expectedLog: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger for each test
			ResetGlobalLogger()

			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
				defer os.Unsetenv(key)
			}

			// Initialize global logger (will load from environment)
			err := InitializeGlobalLogger(nil)
			assert.NoError(t, err)

			// Get the logger and test it
			logger := GetGlobalLogger()
			assert.NotNil(t, logger)

			// Test logging at different levels
			var logBuffer bytes.Buffer
			if concreteLogger, ok := logger.(*ConcreteLogger); ok {
				concreteLogger.writer = &logBuffer

				// Test info logging
				concreteLogger.Info("test info message")
				logOutput := logBuffer.String()

				// Verify the log contains expected information
				assert.Contains(t, logOutput, "test info message")
				assert.Contains(t, logOutput, "info")

				// Test debug logging (should only appear if debug level is enabled)
				logBuffer.Reset()
				concreteLogger.Debug("test debug message")
				debugOutput := logBuffer.String()

				if tt.expectedLog == "debug" {
					assert.Contains(t, debugOutput, "test debug message")
				} else {
					// Debug should be filtered out for non-debug levels
					assert.Empty(t, debugOutput)
				}
			}
		})
	}
}

func TestGlobalLoggerConcurrency(t *testing.T) {
	// Reset global logger before test
	defer ResetGlobalLogger()

	t.Run("Concurrent initialization", func(t *testing.T) {
		ResetGlobalLogger()

		// Start multiple goroutines trying to initialize
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				logger := GetGlobalLogger()
				assert.NotNil(t, logger)
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify we still have a valid logger
		logger := GetGlobalLogger()
		assert.NotNil(t, logger)
	})
}

func TestGlobalLoggerIntegration(t *testing.T) {
	// Reset global logger before test
	defer ResetGlobalLogger()

	t.Run("Full integration test", func(t *testing.T) {
		ResetGlobalLogger()

		// Set environment variables for a complete test
		os.Setenv("ENVIRONMENT", "dev")
		os.Setenv("LOG_LEVEL", "debug")
		os.Setenv("LOG_FORMAT", "json")
		os.Setenv("ENABLE_STACK_TRACES", "true")
		defer func() {
			os.Unsetenv("ENVIRONMENT")
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("LOG_FORMAT")
			os.Unsetenv("ENABLE_STACK_TRACES")
		}()

		// Initialize global logger
		err := InitializeGlobalLogger(nil)
		assert.NoError(t, err)

		// Get logger and test various operations
		logger := GetGlobalLogger()
		assert.NotNil(t, logger)

		var logBuffer bytes.Buffer
		if concreteLogger, ok := logger.(*ConcreteLogger); ok {
			concreteLogger.writer = &logBuffer

			// Test different log levels
			logger.Debug("debug message", String("key", "value"))
			logger.Info("info message", Int("number", 42))
			logger.Error(nil, "error message", Bool("flag", true))

			logOutput := logBuffer.String()

			// Verify messages are present (debug should be included since level is debug)
			assert.Contains(t, logOutput, "debug message")
			assert.Contains(t, logOutput, "info message")
			assert.Contains(t, logOutput, "error message")

			// Verify structured fields are included
			assert.Contains(t, logOutput, "\"key\"")
			assert.Contains(t, logOutput, "\"value\"")
			assert.Contains(t, logOutput, "\"number\"")
			assert.Contains(t, logOutput, "42")
			assert.Contains(t, logOutput, "\"flag\"")
			assert.Contains(t, logOutput, "true")

			// Verify JSON format
			lines := strings.Split(strings.TrimSpace(logOutput), "\n")
			for _, line := range lines {
				if line != "" {
					assert.True(t, strings.HasPrefix(line, "{"), "Log line should be JSON: %s", line)
					assert.True(t, strings.HasSuffix(line, "}"), "Log line should be JSON: %s", line)
				}
			}
		}
	})
}
