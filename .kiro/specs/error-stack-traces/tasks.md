# Implementation Plan

- [x] 1. Create core logger package structure and interfaces
  - Create `pkg/logger` directory and basic package structure
  - Define Logger interface with Error, ErrorWithStack, Fatal, Info, and Debug methods
  - Implement Field struct for structured logging
  - Create Config struct for logger configuration management
  - _Requirements: 2.1, 2.2, 3.1_

- [x] 2. Implement stack trace capture functionality
  - Create StackTrace and Frame structs to represent stack trace data
  - Implement CaptureStackTrace function using Go's runtime package
  - Create Format method for StackTrace to generate readable stack trace strings
  - Add unit tests for stack trace capture and formatting
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 3. Create logger configuration management
  - Implement LoggerConfig struct with environment variable bindings
  - Add configuration loading using envconfig package
  - Set appropriate defaults for development and production environments
  - Create unit tests for configuration loading and validation
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [x] 4. Implement concrete logger with structured output
  - Create concrete Logger implementation with JSON and text formatting options
  - Implement Error, ErrorWithStack, Fatal, Info, and Debug methods
  - Add timestamp, log level, and structured field support
  - Ensure thread-safe logging operations
  - _Requirements: 2.1, 2.2, 2.3_

- [x] 5. Enhance AppError struct with stack trace support
  - Update `pkg/entities/http-entities.go` to include StackTrace field in AppError
  - Add Context field to AppError for additional debugging information
  - Create NewAppError constructor function that captures stack traces
  - Implement WithContext method for adding contextual information
  - _Requirements: 1.1, 1.2, 1.3_

- [x] 6. Create HTTP request context utilities
  - Implement RequestContext struct to capture HTTP request information
  - Create ExtractRequestContext function to extract context from http.Request
  - Add request ID generation for tracing requests across logs
  - Write unit tests for request context extraction
  - _Requirements: 4.1, 4.2_

- [x] 7. Update HTTP handler error logging
  - Modify appHandler.ServeHTTP method in `pkg/handlers/http-server.go` to use new logger
  - Ensure stack traces are logged but never included in HTTP responses
  - Add request context to error logs for better debugging
  - Maintain existing FriendlyError response structure for clients
  - _Requirements: 4.1, 4.4, 4.5_

- [x] 8. Implement panic recovery middleware
  - Create panic recovery middleware for HTTP handlers
  - Log panics with full stack traces using new logger
  - Return appropriate HTTP 500 status codes for recovered panics
  - Ensure application continues running after panic recovery
  - _Requirements: 4.3_

- [x] 9. Update word selector error handling
  - Modify `pkg/wotd/word-selector.go` to use new logger for file reading and parsing errors
  - Add contextual information (file path, word index) to error logs
  - Replace existing error returns with enhanced AppError instances
  - Write unit tests for error scenarios in word selection
  - _Requirements: 5.1_

- [x] 10. Update Twitter client error handling
  - Modify `pkg/wotd/twitter-client.go` to use new logger for API errors
  - Add contextual information (message content, API response) to error logs
  - Update Tweet function to use enhanced error logging
  - Replace log.Println calls with structured logging
  - _Requirements: 5.2_

- [x] 11. Update Mastodon client error handling
  - Modify `pkg/wotd/mastodon-client.go` to use new logger for API and media errors
  - Add contextual information (toot content, media details) to error logs
  - Update Toot and acquireMedia functions to use enhanced error logging
  - Ensure media upload errors include relevant context
  - _Requirements: 5.2_

- [x] 12. Update Google Cloud Storage error handling
  - Modify `pkg/storage/google-cloud-storage.go` to use new logger
  - Add contextual information (bucket name, object name, operation type) to error logs
  - Replace existing log.Printf calls with structured logging
  - Update GetObject method to use enhanced error logging
  - _Requirements: 5.3_

- [x] 13. Update server startup error handling
  - Modify `pkg/handlers/http-server.go` StartServer function to use new logger
  - Replace log.Fatal calls with enhanced fatal logging that includes stack traces
  - Add contextual information (server address, TLS status) to startup errors
  - Ensure configuration errors are logged with appropriate context
  - _Requirements: 1.1, 1.2_

- [x] 14. Update middleware error handling
  - Modify commonMiddleware in `pkg/handlers/http-server.go` to use new logger
  - Add request context to authentication failure logs
  - Replace panic calls with proper error logging and HTTP error responses
  - Ensure middleware errors include request information for debugging
  - _Requirements: 4.1, 4.2_

- [x] 15. Create comprehensive integration tests
  - Write integration tests for HTTP handler error scenarios with mock requests
  - Test panic recovery middleware with intentional panics
  - Verify stack traces are captured and logged but not exposed in HTTP responses
  - Test background service error scenarios (word selection, social media, storage)
  - _Requirements: 1.1, 4.4, 4.5_

- [x] 16. Add configuration integration and environment setup
  - Update application startup to initialize logger with environment configuration
  - Add logger instance to dependency injection or global access pattern
  - Ensure all components can access the configured logger instance
  - Test configuration loading in different environments (dev, prod)
  - _Requirements: 3.1, 3.2, 3.3, 3.4_