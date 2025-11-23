package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

func TestPanicRecoveryMiddleware(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a logger with the buffer as writer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "error",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

	// Set the global logger for testing
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()

	// Create a handler that panics
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic message")
	})

	// Wrap the panic handler with our panic recovery middleware
	recoveredHandler := panicRecoveryMiddleware(panicHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	recoveredHandler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected HTTP 500 status code")
	assert.Contains(t, rr.Body.String(), "Internal Server Error", "Expected generic error message in response")
	assert.NotContains(t, rr.Body.String(), "test panic message", "Panic details should not be exposed in response")

	// Verify the log output
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "HTTP handler panic recovered", "Expected panic recovery log message")
	assert.Contains(t, logOutput, "test panic message", "Expected panic value in log")
	assert.Contains(t, logOutput, "stack_trace", "Expected stack trace in log")
	assert.Contains(t, logOutput, "request_method", "Expected request context in log")
	assert.Contains(t, logOutput, "GET", "Expected request method in log")
	assert.Contains(t, logOutput, "/test", "Expected request path in log")
}

func TestPanicRecoveryMiddleware_WithErrorPanic(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a logger with the buffer as writer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "error",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

	// Set the global logger for testing
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()

	// Create a handler that panics with an error
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	// Wrap the panic handler with our panic recovery middleware
	recoveredHandler := panicRecoveryMiddleware(panicHandler)

	// Create a test request
	req := httptest.NewRequest("POST", "/api/test", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	recoveredHandler.ServeHTTP(rr, req)

	// Verify the response
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Expected HTTP 500 status code")
	assert.Contains(t, rr.Body.String(), "Internal Server Error", "Expected generic error message in response")

	// Verify the log output contains the error
	logOutput := logBuffer.String()
	assert.Contains(t, logOutput, "HTTP handler panic recovered", "Expected panic recovery log message")
	assert.Contains(t, logOutput, "http: abort Handler", "Expected error message in log")
}

func TestPanicRecoveryMiddleware_NormalFlow(t *testing.T) {
	// Create a buffer to capture log output
	var logBuffer bytes.Buffer

	// Create a logger with the buffer as writer
	config := &logger.LoggerConfig{
		EnableStackTraces: true,
		LogLevel:          "error",
		Environment:       "test",
		LogFormat:         "json",
	}
	testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

	// Set the global logger for testing
	logger.SetGlobalLogger(testLogger)
	defer logger.ResetGlobalLogger()

	// Create a normal handler that doesn't panic
	normalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap the normal handler with our panic recovery middleware
	recoveredHandler := panicRecoveryMiddleware(normalHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/normal", nil)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Execute the request
	recoveredHandler.ServeHTTP(rr, req)

	// Verify the response (should be normal)
	assert.Equal(t, http.StatusOK, rr.Code, "Expected HTTP 200 status code")
	assert.Equal(t, "success", rr.Body.String(), "Expected normal response body")

	// Verify no panic logs were generated
	logOutput := logBuffer.String()
	assert.NotContains(t, logOutput, "HTTP handler panic recovered", "Should not have panic recovery logs for normal flow")
}
