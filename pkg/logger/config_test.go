package logger

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear environment variables
	clearLoggerEnvVars()

	config, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Check default values for development environment
	assert.True(t, config.EnableStackTraces)
	assert.Equal(t, "debug", config.LogLevel) // Should be debug in dev environment
	assert.Equal(t, "dev", config.Environment)
	assert.Equal(t, "json", config.LogFormat)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Clear environment variables first
	clearLoggerEnvVars()

	// Set environment variables
	os.Setenv("ENABLE_STACK_TRACES", "false")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("LOG_FORMAT", "text")
	defer clearLoggerEnvVars()

	config, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	assert.False(t, config.EnableStackTraces)
	assert.Equal(t, "error", config.LogLevel)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "text", config.LogFormat)
}

func TestLoadConfig_EnvironmentDefaults(t *testing.T) {
	tests := []struct {
		name                string
		environment         string
		expectedStackTraces bool
		expectedLogLevel    string
	}{
		{
			name:                "development_environment",
			environment:         "development",
			expectedStackTraces: true,
			expectedLogLevel:    "debug",
		},
		{
			name:                "dev_environment",
			environment:         "dev",
			expectedStackTraces: true,
			expectedLogLevel:    "debug",
		},
		{
			name:                "production_environment",
			environment:         "production",
			expectedStackTraces: true, // Default from struct tag
			expectedLogLevel:    "info",
		},
		{
			name:                "prod_environment",
			environment:         "prod",
			expectedStackTraces: true, // Default from struct tag
			expectedLogLevel:    "info",
		},
		{
			name:                "test_environment",
			environment:         "test",
			expectedStackTraces: false,
			expectedLogLevel:    "info",
		},
		{
			name:                "testing_environment",
			environment:         "testing",
			expectedStackTraces: false,
			expectedLogLevel:    "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearLoggerEnvVars()
			os.Setenv("ENVIRONMENT", tt.environment)
			defer clearLoggerEnvVars()

			config, err := LoadConfig()
			require.NoError(t, err)
			require.NotNil(t, config)

			assert.Equal(t, tt.expectedStackTraces, config.EnableStackTraces)
			assert.Equal(t, tt.expectedLogLevel, config.LogLevel)
			assert.Equal(t, strings.ToLower(tt.environment), config.Environment)
		})
	}
}

func TestLoggerConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      LoggerConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_config",
			config: LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "info",
				Environment:       "dev",
				LogFormat:         "json",
			},
			expectError: false,
		},
		{
			name: "invalid_log_level",
			config: LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "invalid",
				Environment:       "dev",
				LogFormat:         "json",
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "invalid_log_format",
			config: LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "info",
				Environment:       "dev",
				LogFormat:         "invalid",
			},
			expectError: true,
			errorMsg:    "invalid log format",
		},
		{
			name: "invalid_environment",
			config: LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "info",
				Environment:       "invalid",
				LogFormat:         "json",
			},
			expectError: true,
			errorMsg:    "invalid environment",
		},
		{
			name: "case_insensitive_validation",
			config: LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "INFO",
				Environment:       "PROD",
				LogFormat:         "JSON",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				// Check that values are normalized to lowercase
				assert.Equal(t, strings.ToLower(tt.config.LogLevel), tt.config.LogLevel)
				assert.Equal(t, strings.ToLower(tt.config.Environment), tt.config.Environment)
				assert.Equal(t, strings.ToLower(tt.config.LogFormat), tt.config.LogFormat)
			}
		})
	}
}

func TestLoggerConfig_IsDebugEnabled(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected bool
	}{
		{"debug_level", "debug", true},
		{"info_level", "info", false},
		{"error_level", "error", false},
		{"fatal_level", "fatal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{LogLevel: tt.logLevel}
			assert.Equal(t, tt.expected, config.IsDebugEnabled())
		})
	}
}

func TestLoggerConfig_IsStackTraceEnabled(t *testing.T) {
	tests := []struct {
		name              string
		enableStackTraces bool
		expected          bool
	}{
		{"enabled", true, true},
		{"disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &LoggerConfig{EnableStackTraces: tt.enableStackTraces}
			assert.Equal(t, tt.expected, config.IsStackTraceEnabled())
		})
	}
}

func TestLoadConfig_ProductionStackTraceOverride(t *testing.T) {
	clearLoggerEnvVars()

	// Test that production environment doesn't override explicitly set stack traces
	os.Setenv("ENVIRONMENT", "production")
	os.Setenv("ENABLE_STACK_TRACES", "false")
	defer clearLoggerEnvVars()

	config, err := LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Should respect the explicit setting
	assert.False(t, config.EnableStackTraces)
	assert.Equal(t, "production", config.Environment)
}

// Helper function to clear logger-related environment variables
func clearLoggerEnvVars() {
	os.Unsetenv("ENABLE_STACK_TRACES")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("ENVIRONMENT")
	os.Unsetenv("LOG_FORMAT")
}
