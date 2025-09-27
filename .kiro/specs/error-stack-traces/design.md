# Design Document

## Overview

This design enhances the Te Reo Bot's error handling and logging capabilities by adding comprehensive stack trace logging throughout the application. The current codebase uses basic error logging with `log.Println()` and `log.Fatal()` calls, but lacks detailed stack trace information that would help with debugging and troubleshooting.

The solution will introduce a centralized error logging package that captures and formats stack traces consistently across all components while maintaining security by ensuring stack traces are never exposed in HTTP responses to clients.

## Architecture

### Current Error Handling Analysis

The current codebase has several error handling patterns:

1. **HTTP Handlers**: Use custom `AppError` struct with error wrapping in `appHandler` type
2. **Basic Logging**: Simple `log.Println()` calls without stack traces
3. **Fatal Errors**: `log.Fatal()` calls for critical startup errors
4. **Error Responses**: Generic error messages sent to clients via `FriendlyError` struct

### Proposed Architecture

The new architecture will introduce:

1. **Enhanced Logger Package**: A new `pkg/logger` package for centralized error logging with stack traces
2. **Stack Trace Capture**: Runtime stack trace capture using Go's `runtime` package
3. **Configuration Management**: Environment-based configuration for stack trace verbosity
4. **Middleware Enhancement**: Updated HTTP middleware to log errors with stack traces
5. **Error Wrapper Enhancement**: Enhanced `AppError` to include stack trace information

## Components and Interfaces

### 1. Logger Package (`pkg/logger`)

```go
// Logger interface for structured logging with stack traces
type Logger interface {
    Error(err error, message string, fields ...Field)
    ErrorWithStack(err error, message string, fields ...Field)
    Fatal(err error, message string, fields ...Field)
    Info(message string, fields ...Field)
    Debug(message string, fields ...Field)
}

// Config for logger configuration
type Config struct {
    EnableStackTraces bool
    LogLevel         string
    Environment      string // dev, prod, etc.
}

// Field for structured logging
type Field struct {
    Key   string
    Value interface{}
}
```

### 2. Stack Trace Utilities

```go
// StackTrace captures and formats stack trace information
type StackTrace struct {
    Frames []Frame
}

// Frame represents a single stack frame
type Frame struct {
    Function string
    File     string
    Line     int
}

// CaptureStackTrace captures the current stack trace
func CaptureStackTrace(skip int) *StackTrace

// Format formats the stack trace for logging
func (st *StackTrace) Format() string
```

### 3. Enhanced Error Types

```go
// Enhanced AppError with stack trace support
type AppError struct {
    Error      error      `json:"-"`
    Message    string     `json:"message"`
    Code       int        `json:"code"`
    StackTrace *StackTrace `json:"-"`
    Context    map[string]interface{} `json:"-"`
}

// NewAppError creates a new AppError with stack trace
func NewAppError(err error, code int, message string) *AppError

// WithContext adds context information to the error
func (ae *AppError) WithContext(key string, value interface{}) *AppError
```

### 4. HTTP Context Enhancement

```go
// RequestContext captures HTTP request information for logging
type RequestContext struct {
    Method     string
    Path       string
    UserAgent  string
    RemoteAddr string
    RequestID  string
}

// ExtractRequestContext extracts context from HTTP request
func ExtractRequestContext(r *http.Request) *RequestContext
```

## Data Models

### Logger Configuration

```go
type LoggerConfig struct {
    EnableStackTraces bool   `envconfig:"ENABLE_STACK_TRACES" default:"true"`
    LogLevel         string `envconfig:"LOG_LEVEL" default:"info"`
    Environment      string `envconfig:"ENVIRONMENT" default:"dev"`
    LogFormat        string `envconfig:"LOG_FORMAT" default:"json"`
}
```

### Log Entry Structure

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "error",
  "message": "Failed to acquire image",
  "error": "storage: object doesn't exist",
  "stack_trace": [
    {
      "function": "github.com/wizact/te-reo-bot/pkg/storage.(*GoogleCloudStorageClientWrapper).GetObject",
      "file": "/app/pkg/storage/google-cloud-storage.go",
      "line": 25
    },
    {
      "function": "github.com/wizact/te-reo-bot/pkg/wotd.acquireMedia",
      "file": "/app/pkg/wotd/mastodon-client.go",
      "line": 85
    }
  ],
  "context": {
    "bucket_name": "te-reo-images",
    "object_name": "kiwi.jpg",
    "operation": "get_object"
  },
  "request_context": {
    "method": "GET",
    "path": "/messages",
    "user_agent": "curl/7.68.0",
    "remote_addr": "192.168.1.100",
    "request_id": "req-123456"
  }
}
```

## Error Handling

### 1. HTTP Handler Error Handling

- Enhance the existing `appHandler` to log errors with stack traces before returning responses
- Ensure stack traces are never included in HTTP responses to clients
- Maintain the existing `FriendlyError` structure for client responses
- Add request context to error logs for better debugging

### 2. Background Service Error Handling

- Update word selection, social media posting, and storage operations to use the new logger
- Include relevant context (word index, platform, bucket/object names) in error logs
- Capture stack traces for all error conditions

### 3. Panic Recovery

- Implement panic recovery middleware for HTTP handlers
- Log panics with full stack traces
- Return appropriate HTTP status codes (500) for recovered panics
- Ensure application continues running after panic recovery

### 4. Configuration-Based Logging

- Allow disabling stack traces in production if needed for performance
- Support different log levels (debug, info, error, fatal)
- Environment-specific defaults (verbose in dev, configurable in prod)

## Testing Strategy

### 1. Unit Tests

- Test stack trace capture functionality
- Test log formatting and output
- Test configuration loading and validation
- Test error wrapping and context addition

### 2. Integration Tests

- Test HTTP handler error logging with mock requests
- Test background service error scenarios
- Test panic recovery in HTTP handlers
- Verify stack traces are not leaked in HTTP responses

### 3. Performance Tests

- Measure performance impact of stack trace capture
- Test with stack traces enabled and disabled
- Ensure minimal impact on normal operation paths

### 4. Security Tests

- Verify stack traces are never exposed in HTTP responses
- Test that sensitive information is not logged
- Validate that error messages to clients remain generic

## Implementation Phases

### Phase 1: Core Logger Package
- Create `pkg/logger` package with basic logging functionality
- Implement stack trace capture utilities
- Add configuration management

### Phase 2: Error Type Enhancement
- Enhance `AppError` with stack trace support
- Update error creation patterns throughout codebase
- Add context support for errors

### Phase 3: HTTP Handler Integration
- Update `appHandler` to use new logging
- Enhance middleware for request context capture
- Implement panic recovery middleware

### Phase 4: Background Service Integration
- Update word selection error handling
- Update social media client error handling
- Update storage client error handling

### Phase 5: Configuration and Testing
- Add environment variable configuration
- Implement comprehensive test suite
- Performance testing and optimization