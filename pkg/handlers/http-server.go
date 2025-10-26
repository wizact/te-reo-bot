package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

const (
	healthCheckRoute = "/__health-check"
	messagesRoute    = "/messages"
)

// getLogger returns the global logger instance for HTTP handlers
func getLogger() logger.Logger {
	return logger.GetGlobalLogger()
}

// StartServer starts the http server
func StartServer(address, port string, tls bool) {
	serverAddress := fmt.Sprintf("%s:%s", address, port)

	// Prepare server context fields for logging
	serverFields := []logger.Field{
		{Key: "server_address", Value: serverAddress},
		{Key: "tls_enabled", Value: tls},
		{Key: "address", Value: address},
		{Key: "port", Value: port},
	}

	getLogger().Info("Starting server initialization", serverFields...)

	router := mux.NewRouter()
	router.Use(panicRecoveryMiddleware)
	router.Use(commonMiddleware)

	// HealthCheck route setup
	hcr := HealthCheckRoute{}
	hcr.SetupRoutes(healthCheckRoute, router)

	// MessageRoute route setup - handle configuration errors with enhanced logging
	bn, err := (&StorageConfig{}).GetBucketName()
	if err != nil {
		configFields := append(serverFields, logger.Field{Key: "config_type", Value: "storage_bucket"})
		getLogger().Fatal(err, "Cannot get the bucket name from environment variables", configFields...)
		return
	}

	mr := MessagesRoute{bucketName: bn}
	mr.SetupRoutes(messagesRoute, router)

	// Start server with enhanced error logging and contextual information
	if tls {
		tlsFields := append(serverFields,
			logger.Field{Key: "cert_file", Value: "certs/server.crt"},
			logger.Field{Key: "key_file", Value: "certs/server.key"},
		)

		getLogger().Info("Starting HTTPS server", tlsFields...)
		err := http.ListenAndServeTLS(serverAddress,
			"certs/server.crt",
			"certs/server.key",
			router)
		if err != nil {
			getLogger().Fatal(err, "HTTPS server failed to start", tlsFields...)
		}
	} else {
		getLogger().Info("Starting HTTP server", serverFields...)
		err := http.ListenAndServe(serverAddress, router)
		if err != nil {
			getLogger().Fatal(err, "HTTP server failed to start", serverFields...)
		}
	}
}

// commonMiddleware the generic middleware
func commonMiddleware(next http.Handler) http.Handler {
	var s ServerConfig
	err := envconfig.Process("tereobot", &s)

	if err != nil {
		// Use new logger with contextual information for configuration errors
		configFields := []logger.Field{
			{Key: "config_type", Value: "server_config"},
			{Key: "config_prefix", Value: "tereobot"},
		}
		getLogger().ErrorWithStack(err, "Cannot read server configuration", configFields...)

		// Return a middleware that always returns HTTP 500 for configuration errors
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract request context for configuration error logging
			requestContext := logger.ExtractRequestContext(r)
			logFields := configFields
			if requestContext != nil {
				logFields = append(logFields, requestContext.ToFields()...)
			}

			getLogger().ErrorWithStack(err, "Server configuration error - cannot process request", logFields...)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Index(r.RequestURI, healthCheckRoute) != 0 {
			rak, err := findCaseInsensitiveHeader("X-Api-Key", r)

			if err != nil {
				// Extract request context for authentication failure logging
				requestContext := logger.ExtractRequestContext(r)
				logFields := []logger.Field{}
				if requestContext != nil {
					logFields = append(logFields, requestContext.ToFields()...)
				}

				// Log authentication failure with request context
				getLogger().ErrorWithStack(err, "Authentication failed - missing API key", logFields...)

				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}

			if rak != s.ApiKey {
				// Extract request context for authentication failure logging
				requestContext := logger.ExtractRequestContext(r)
				logFields := []logger.Field{
					{Key: "provided_api_key_length", Value: len(rak)},
				}
				if requestContext != nil {
					logFields = append(logFields, requestContext.ToFields()...)
				}

				// Log authentication failure with request context (don't log actual keys for security)
				getLogger().ErrorWithStack(errors.New("invalid API key"), "Authentication failed - invalid API key", logFields...)

				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func findCaseInsensitiveHeader(headerName string, r *http.Request) (string, error) {
	if strings.Trim(headerName, "") == "" {
		return "", errors.New("auth header is missing")
	}

	for s := range r.Header {
		if strings.EqualFold(s, headerName) {
			apiKeyHeader := r.Header[headerName]
			if len(apiKeyHeader) > 0 {
				return apiKeyHeader[0], nil
			}
		}
	}

	return "", errors.New("auth header is missing")

}

// panicRecoveryMiddleware recovers from panics in HTTP handlers and logs them with stack traces
func panicRecoveryMiddleware(next http.Handler) http.Handler {
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
				getLogger().ErrorWithStack(panicErr, "HTTP handler panic recovered", logFields...)

				// Return HTTP 500 status code to client
				// Don't include panic details in the response for security
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

type appHandler func(http.ResponseWriter, *http.Request) *ent.AppError

// ServeHTTP to serve requests but respond with a friendly error message if any
func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.

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

		// Always capture stack trace at the handler level for consistent logging
		// This ensures we always have debugging information, regardless of
		// whether the AppError was created with NewAppError() or manually
		getLogger().ErrorWithStack(e.Err, "HTTP handler error occurred", logFields...)

		// Return friendly error response to client (never include stack traces or internal details)
		w.WriteHeader(e.Code)
		ee := json.NewEncoder(w).Encode(&ent.FriendlyError{Message: e.Message})
		if ee != nil {
			// Use the new logger for encoding errors as well
			getLogger().Fatal(ee, "Failed to encode error response")
		}
	}
}

// ServerConfig to wrap configuration
type ServerConfig struct {
	ApiKey string
}

// StorageConfig stores information required for storage service
type StorageConfig struct {
	BucketName string
}

func (s *StorageConfig) GetBucketName() (string, error) {
	err := envconfig.Process("tereobot", s)
	if err != nil {
		return "", err
	}

	return s.BucketName, nil
}
