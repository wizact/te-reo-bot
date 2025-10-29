package logger

import (
	"sync"
)

var (
	// globalLogger holds the global logger instance
	globalLogger Logger
	// globalLoggerOnce ensures the global logger is initialized only once
	globalLoggerOnce sync.Once
	// globalLoggerMutex protects access to the global logger
	globalLoggerMutex sync.RWMutex
)

// InitializeGlobalLogger initializes the global logger with the provided configuration
// This should be called once during application startup
func InitializeGlobalLogger(config *LoggerConfig) error {
	var initErr error

	globalLoggerOnce.Do(func() {
		if config == nil {
			// Load configuration from environment if not provided
			config, initErr = LoadConfig()
			if initErr != nil {
				return
			}
		}

		globalLoggerMutex.Lock()
		defer globalLoggerMutex.Unlock()

		globalLogger = NewLogger(config)
	})

	return initErr
}

// GetGlobalLogger returns the global logger instance
// If the global logger hasn't been initialized, it will initialize with default configuration
func GetGlobalLogger() Logger {
	globalLoggerMutex.RLock()
	if globalLogger != nil {
		defer globalLoggerMutex.RUnlock()
		return globalLogger
	}
	globalLoggerMutex.RUnlock()

	// Initialize with default configuration if not already initialized
	config, err := LoadConfig()
	if err != nil {
		// Fallback to hardcoded defaults if environment loading fails
		config = &LoggerConfig{
			EnableStackTraces: true,
			LogLevel:          "info",
			Environment:       "dev",
			LogFormat:         "json",
		}
	}

	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()

	// Double-check after acquiring write lock
	if globalLogger == nil {
		globalLogger = NewLogger(config)
	}

	return globalLogger
}

// SetGlobalLogger sets the global logger instance (primarily for testing)
func SetGlobalLogger(logger Logger) {
	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()
	globalLogger = logger
}

// ResetGlobalLogger resets the global logger (primarily for testing)
func ResetGlobalLogger() {
	globalLoggerMutex.Lock()
	defer globalLoggerMutex.Unlock()
	globalLogger = nil
	globalLoggerOnce = sync.Once{}
}
