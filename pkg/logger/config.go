package logger

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// LoggerConfig holds logger configuration settings with environment variable bindings
type LoggerConfig struct {
	EnableStackTraces bool   `envconfig:"ENABLE_STACK_TRACES" default:"true"`
	LogLevel          string `envconfig:"LOG_LEVEL" default:"info"`
	Environment       string `envconfig:"ENVIRONMENT" default:"dev"`
	LogFormat         string `envconfig:"LOG_FORMAT" default:"json"`
}

// LoadConfig loads logger configuration from environment variables
func LoadConfig() (*LoggerConfig, error) {
	var config LoggerConfig

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load logger configuration: %w", err)
	}

	// Apply environment-specific defaults
	config = applyEnvironmentDefaults(config)

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger configuration: %w", err)
	}

	return &config, nil
}

// applyEnvironmentDefaults applies environment-specific default values
func applyEnvironmentDefaults(config LoggerConfig) LoggerConfig {
	switch strings.ToLower(config.Environment) {
	case "dev", "development":
		// Development defaults: enable stack traces, debug level
		if config.LogLevel == "info" {
			config.LogLevel = "debug"
		}
		config.EnableStackTraces = true
	case "prod", "production":
		// Production defaults: configurable stack traces, info level
		if config.LogLevel == "debug" {
			config.LogLevel = "info"
		}
		// Keep EnableStackTraces as configured (don't override)
	case "test", "testing":
		// Test defaults: disable stack traces for cleaner test output
		config.EnableStackTraces = false
		if config.LogLevel == "debug" {
			config.LogLevel = "info"
		}
	}

	return config
}

// Validate validates the logger configuration
func (c *LoggerConfig) Validate() error {
	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error", "fatal"}
	levelValid := false
	normalizedLevel := strings.ToLower(c.LogLevel)

	for _, level := range validLevels {
		if normalizedLevel == level {
			levelValid = true
			c.LogLevel = normalizedLevel // Normalize to lowercase
			break
		}
	}

	if !levelValid {
		return fmt.Errorf("invalid log level '%s', must be one of: %s", c.LogLevel, strings.Join(validLevels, ", "))
	}

	// Validate log format
	validFormats := []string{"json", "text"}
	formatValid := false
	normalizedFormat := strings.ToLower(c.LogFormat)

	for _, format := range validFormats {
		if normalizedFormat == format {
			formatValid = true
			c.LogFormat = normalizedFormat // Normalize to lowercase
			break
		}
	}

	if !formatValid {
		return fmt.Errorf("invalid log format '%s', must be one of: %s", c.LogFormat, strings.Join(validFormats, ", "))
	}

	// Validate environment
	validEnvironments := []string{"dev", "development", "prod", "production", "test", "testing"}
	envValid := false
	normalizedEnv := strings.ToLower(c.Environment)

	for _, env := range validEnvironments {
		if normalizedEnv == env {
			envValid = true
			c.Environment = normalizedEnv // Normalize to lowercase
			break
		}
	}

	if !envValid {
		return fmt.Errorf("invalid environment '%s', must be one of: %s", c.Environment, strings.Join(validEnvironments, ", "))
	}

	return nil
}

// IsDebugEnabled returns true if debug logging is enabled
func (c *LoggerConfig) IsDebugEnabled() bool {
	return c.LogLevel == "debug"
}

// IsStackTraceEnabled returns true if stack traces should be included in logs
func (c *LoggerConfig) IsStackTraceEnabled() bool {
	return c.EnableStackTraces
}
