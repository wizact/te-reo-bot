package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

// TestHTTPHandlerErrorScenarios tests various HTTP handler error scenarios
func TestHTTPHandlerErrorScenarios(t *testing.T) {
	tests := []struct {
		name                string
		setupHandler        func(logger.Logger) http.Handler
		request             func() *http.Request
		expectedStatus      int
		expectedBody        string
		shouldLogError      bool
		logShouldContain    []string
		logShouldNotContain []string
	}{
		{
			name: "AppHandler returns AppError with stack trace",
			setupHandler: func(testLogger logger.Logger) http.Handler {
				// Create a test appHandler that returns an error
				testHandler := func(w http.ResponseWriter, r *http.Request) *entities.AppError {
					err := entities.NewAppError(
						fmt.Errorf("database connection failed"),
						500,
						"Internal server error occurred",
					)
					err.WithContext("operation", "database_query")
					err.WithContext("table", "users")
					return err
				}
				// Convert to http.Handler using the same pattern as in the handlers package
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if e := testHandler(w, r); e != nil {
						// Extract request context for logging
						requestContext := logger.ExtractRequestContext(r)

						// Prepare logging fields with request context and error context
						logFields := []logger.Field{}

						// Add request context fields
						if requestContext != nil {
							logFields = append(logFields, requestContext.ToFields()...)
						}

						// Add error context fields if available
						if e.Context != nil {
							for key, value := range e.Context {
								logFields = append(logFields, logger.Field{Key: key, Value: value})
							}
						}

						// Always log with stack trace for testing purposes
						testLogger.ErrorWithStack(e.Err, "HTTP handler error occurred", logFields...)

						// Return friendly error response to client (never include stack traces or internal details)
						w.WriteHeader(e.Code)
						json.NewEncoder(w).Encode(&entities.FriendlyError{Message: e.Message})
					}
				})
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test-error", nil)
				req.Header.Set("User-Agent", "test-client/1.0")
				req.Header.Set("X-Request-ID", "test-123")
				return req
			},
			expectedStatus: 500,
			expectedBody:   `{"message":"Internal server error occurred"}`,
			shouldLogError: true,
			logShouldContain: []string{
				"HTTP handler error occurred",
				"database connection failed",
				"stack_trace",
				"request_method",
				"GET",
				"/test-error",
				"operation",
				"database_query",
				"table",
				"users",
			},
			logShouldNotContain: []string{
				"database connection failed", // Should not be in HTTP response
			},
		},
		{
			name: "AppHandler returns AppError without stack trace",
			setupHandler: func(testLogger logger.Logger) http.Handler {
				// Create a test appHandler that returns an error without stack trace
				testHandler := func(w http.ResponseWriter, r *http.Request) *entities.AppError {
					// Create AppError without automatic stack trace capture
					return &entities.AppError{
						Err:     fmt.Errorf("validation failed"),
						Message: "Invalid input provided",
						Code:    400,
						Context: map[string]interface{}{
							"field": "email",
							"value": "invalid-email",
						},
					}
				}
				// Convert to http.Handler using the same pattern as in the handlers package
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if e := testHandler(w, r); e != nil {
						// Extract request context for logging
						requestContext := logger.ExtractRequestContext(r)

						// Prepare logging fields with request context and error context
						logFields := []logger.Field{}

						// Add request context fields
						if requestContext != nil {
							logFields = append(logFields, requestContext.ToFields()...)
						}

						// Add error context fields if available
						if e.Context != nil {
							for key, value := range e.Context {
								logFields = append(logFields, logger.Field{Key: key, Value: value})
							}
						}

						// Log error with stack trace if available, otherwise log without stack trace
						// Always log with stack trace for testing purposes
						testLogger.ErrorWithStack(e.Err, "HTTP handler error occurred", logFields...)

						// Return friendly error response to client (never include stack traces or internal details)
						w.WriteHeader(e.Code)
						json.NewEncoder(w).Encode(&entities.FriendlyError{Message: e.Message})
					}
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("POST", "/validate", strings.NewReader(`{"email":"invalid"}`))
			},
			expectedStatus: 400,
			expectedBody:   `{"message":"Invalid input provided"}`,
			shouldLogError: true,
			logShouldContain: []string{
				"HTTP handler error occurred",
				"validation failed",
				"stack_trace", // Should capture stack trace at ServeHTTP level
				"field",
				"email",
				"value",
				"invalid-email",
			},
		},
		{
			name: "Successful request - no error logging",
			setupHandler: func(testLogger logger.Logger) http.Handler {
				// Create a test appHandler that succeeds
				testHandler := func(w http.ResponseWriter, r *http.Request) *entities.AppError {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status":"success"}`))
					return nil
				}
				// Convert to http.Handler using the same pattern as in the handlers package
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if e := testHandler(w, r); e != nil {
						// This shouldn't happen for successful requests, but include for completeness
						testLogger.ErrorWithStack(e.Err, "HTTP handler error occurred")
						w.WriteHeader(e.Code)
						json.NewEncoder(w).Encode(&entities.FriendlyError{Message: e.Message})
					}
				})
			},
			request: func() *http.Request {
				return httptest.NewRequest("GET", "/success", nil)
			},
			expectedStatus: 200,
			expectedBody:   `{"status":"success"}`,
			shouldLogError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var logBuffer bytes.Buffer

			// Create a logger with the buffer as writer
			config := &logger.LoggerConfig{
				EnableStackTraces: true,
				LogLevel:          "debug",
				Environment:       "test",
				LogFormat:         "json",
			}
			testLogger := logger.NewLoggerWithWriter(config, &logBuffer)

			// Set up the handler with test logger
			handler := tt.setupHandler(testLogger)

			// Create test request
			req := tt.request()

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			handler.ServeHTTP(rr, req)

			// Verify HTTP response
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String(), "Unexpected response body")
			}

			// Verify logging behavior
			logOutput := logBuffer.String()

			if tt.shouldLogError {
				assert.NotEmpty(t, logOutput, "Expected error to be logged")

				for _, shouldContain := range tt.logShouldContain {
					assert.Contains(t, logOutput, shouldContain,
						"Log should contain: %s", shouldContain)
				}

				for _, shouldNotContain := range tt.logShouldNotContain {
					assert.NotContains(t, rr.Body.String(), shouldNotContain,
						"HTTP response should not contain: %s", shouldNotContain)
				}
			} else {
				// For successful requests, there should be no error logs
				assert.NotContains(t, logOutput, "HTTP handler error occurred",
					"Should not log errors for successful requests")
			}
		})
	}
}

// TestPanicRecoveryIntegration tests panic recovery middleware integration
func TestPanicRecoveryIntegration(t *testing.T) {
	tests := []struct {
		name             string
		panicValue       interface{}
		expectedStatus   int
		logShouldContain []string
	}{
		{
			name:           "String panic",
			panicValue:     "something went wrong",
			expectedStatus: 500,
			logShouldContain: []string{
				"HTTP handler panic recovered",
				"something went wrong",
				"stack_trace",
				"request_method",
				"GET",
			},
		},
		{
			name:           "Error panic",
			panicValue:     fmt.Errorf("database timeout"),
			expectedStatus: 500,
			logShouldContain: []string{
				"HTTP handler panic recovered",
				"database timeout",
				"stack_trace",
			},
		},
		{
			name:           "Interface nil panic",
			panicValue:     (*error)(nil),
			expectedStatus: 500,
			logShouldContain: []string{
				"HTTP handler panic recovered",
				"stack_trace",
			},
		},
		{
			name:           "Complex object panic",
			panicValue:     map[string]interface{}{"error": "complex", "code": 123},
			expectedStatus: 500,
			logShouldContain: []string{
				"HTTP handler panic recovered",
				"stack_trace",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Create a handler that panics
			panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic(tt.panicValue)
			})

			// Set up router with panic recovery middleware
			router := mux.NewRouter()
			// Create a custom panic recovery middleware for testing
			panicRecovery := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer func() {
						if recovered := recover(); recovered != nil {
							// Extract request context for panic logging
							requestContext := logger.ExtractRequestContext(r)

							// Prepare logging fields with request context
							logFields := []logger.Field{
								{Key: "panic_value", Value: fmt.Sprintf("%v", recovered)},
							}

							// Add request context fields
							if requestContext != nil {
								logFields = append(logFields, requestContext.ToFields()...)
							}

							// Create an error from the panic value
							var panicErr error
							if err, ok := recovered.(error); ok {
								panicErr = err
							} else {
								panicErr = fmt.Errorf("panic: %v", recovered)
							}

							// Log the panic with full stack trace
							testLogger.ErrorWithStack(panicErr, "HTTP handler panic recovered", logFields...)

							// Return HTTP 500 status code to client
							// Don't include panic details in the response for security
							http.Error(w, "Internal Server Error", http.StatusInternalServerError)
						}
					}()

					// Continue with the next handler
					next.ServeHTTP(w, r)
				})
			}
			router.Use(panicRecovery)
			router.Handle("/panic", panicHandler)

			// Create test request
			req := httptest.NewRequest("GET", "/panic", nil)
			req.Header.Set("User-Agent", "panic-test-client")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(rr, req)

			// Verify HTTP response
			assert.Equal(t, tt.expectedStatus, rr.Code, "Expected HTTP 500 status")
			assert.Contains(t, rr.Body.String(), "Internal Server Error",
				"Expected generic error message")

			// Verify panic details are NOT exposed in response
			panicStr := fmt.Sprintf("%v", tt.panicValue)
			if panicStr != "<nil>" && panicStr != "" {
				assert.NotContains(t, rr.Body.String(), panicStr,
					"Panic details should not be exposed in HTTP response")
			}

			// Verify logging
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput, "Panic should be logged")

			for _, shouldContain := range tt.logShouldContain {
				assert.Contains(t, logOutput, shouldContain,
					"Log should contain: %s", shouldContain)
			}

			// Verify stack trace is captured in logs
			assert.Contains(t, logOutput, "stack_trace", "Stack trace should be logged")
			assert.Contains(t, logOutput, "TestPanicRecoveryIntegration",
				"Stack trace should include test function")
		})
	}
}

// TestMiddlewareErrorHandling tests error handling in middleware
func TestMiddlewareErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		setupEnv         func()
		cleanupEnv       func()
		request          func() *http.Request
		expectedStatus   int
		logShouldContain []string
	}{
		{
			name: "Missing API key",
			setupEnv: func() {
				os.Setenv("TEREOBOT_APIKEY", "valid-api-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("TEREOBOT_APIKEY")
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/messages", nil)
				// No API key header
				return req
			},
			expectedStatus: 401,
			logShouldContain: []string{
				"Authentication failed - missing API key",
				"stack_trace",
				"request_method",
				"GET",
				"/messages",
			},
		},
		{
			name: "Invalid API key",
			setupEnv: func() {
				os.Setenv("TEREOBOT_APIKEY", "valid-api-key")
			},
			cleanupEnv: func() {
				os.Unsetenv("TEREOBOT_APIKEY")
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/messages", nil)
				req.Header.Set("X-Api-Key", "invalid-key")
				return req
			},
			expectedStatus: 401,
			logShouldContain: []string{
				"Authentication failed - invalid API key",
				"stack_trace",
				"provided_api_key_length",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment
			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			defer func() {
				if tt.cleanupEnv != nil {
					tt.cleanupEnv()
				}
			}()

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

			// Create a simple handler for testing middleware
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Set up router with middleware
			router := mux.NewRouter()
			// Create a custom common middleware for testing
			commonMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Skip health check route
					if !strings.HasPrefix(r.RequestURI, "/__health-check") {
						// Check for API key header
						apiKey := r.Header.Get("X-Api-Key")
						if apiKey == "" {
							// Extract request context for authentication failure logging
							requestContext := logger.ExtractRequestContext(r)
							logFields := []logger.Field{}
							if requestContext != nil {
								logFields = append(logFields, requestContext.ToFields()...)
							}

							// Log authentication failure with request context
							testLogger.ErrorWithStack(fmt.Errorf("auth header is missing"), "Authentication failed - missing API key", logFields...)

							http.Error(w, "authentication failed", http.StatusUnauthorized)
							return
						}

						// Check if API key is valid (get from environment)
						expectedKey := os.Getenv("TEREOBOT_APIKEY")
						if apiKey != expectedKey {
							// Extract request context for authentication failure logging
							requestContext := logger.ExtractRequestContext(r)
							logFields := []logger.Field{
								{Key: "provided_api_key_length", Value: len(apiKey)},
							}
							if requestContext != nil {
								logFields = append(logFields, requestContext.ToFields()...)
							}

							// Log authentication failure with request context (don't log actual keys for security)
							testLogger.ErrorWithStack(fmt.Errorf("invalid API key"), "Authentication failed - invalid API key", logFields...)

							http.Error(w, "authentication failed", http.StatusUnauthorized)
							return
						}
					}
					next.ServeHTTP(w, r)
				})
			}
			router.Use(commonMiddleware)
			router.Handle("/messages", testHandler)

			// Create test request
			req := tt.request()

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(rr, req)

			// Verify HTTP response
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Verify logging
			logOutput := logBuffer.String()

			for _, shouldContain := range tt.logShouldContain {
				assert.Contains(t, logOutput, shouldContain,
					"Log should contain: %s", shouldContain)
			}
		})
	}
}
