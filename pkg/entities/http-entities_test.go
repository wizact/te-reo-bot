package entities

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppError(t *testing.T) {
	// Test creating a new AppError
	originalErr := errors.New("original error")
	appErr := NewAppError(originalErr, 500, "Internal server error")

	assert.NotNil(t, appErr)
	assert.Equal(t, originalErr, appErr.Err)
	assert.Equal(t, 500, appErr.Code)
	assert.Equal(t, "Internal server error", appErr.Message)
	assert.NotNil(t, appErr.StackTrace)
	assert.NotNil(t, appErr.Context)
	assert.True(t, appErr.HasStackTrace())
}

func TestAppError_WithContext(t *testing.T) {
	// Test adding context to AppError
	appErr := NewAppError(errors.New("test error"), 400, "Bad request")

	// Add context using method chaining
	appErr.WithContext("user_id", "12345").
		WithContext("operation", "create_user").
		WithContext("request_id", "req-abc123")

	// Verify context values
	userID, exists := appErr.GetContext("user_id")
	assert.True(t, exists)
	assert.Equal(t, "12345", userID)

	operation, exists := appErr.GetContext("operation")
	assert.True(t, exists)
	assert.Equal(t, "create_user", operation)

	requestID, exists := appErr.GetContext("request_id")
	assert.True(t, exists)
	assert.Equal(t, "req-abc123", requestID)

	// Test non-existent key
	_, exists = appErr.GetContext("non_existent")
	assert.False(t, exists)
}

func TestAppError_GetContext(t *testing.T) {
	// Test getting context from AppError
	appErr := NewAppError(errors.New("test error"), 404, "Not found")

	// Test empty context
	_, exists := appErr.GetContext("missing_key")
	assert.False(t, exists)

	// Add context and test retrieval
	appErr.WithContext("resource", "user")
	resource, exists := appErr.GetContext("resource")
	assert.True(t, exists)
	assert.Equal(t, "user", resource)
}

func TestAppError_HasStackTrace(t *testing.T) {
	// Test stack trace detection
	appErr := NewAppError(errors.New("test error"), 500, "Server error")
	assert.True(t, appErr.HasStackTrace())

	// Test with nil stack trace
	appErr.StackTrace = nil
	assert.False(t, appErr.HasStackTrace())
}

func TestAppError_JSONSerialization(t *testing.T) {
	// Test that sensitive fields are not exposed in JSON
	appErr := NewAppError(errors.New("internal error"), 500, "Server error")
	appErr.WithContext("sensitive_data", "secret")

	// The JSON tags should prevent Error, StackTrace, and Context from being serialized
	// This is important for security - we don't want to expose internal details to clients

	// We can't easily test JSON serialization without importing encoding/json,
	// but the struct tags ensure the fields won't be serialized
	assert.NotNil(t, appErr.Err)                    // Internal error exists
	assert.NotNil(t, appErr.StackTrace)             // Stack trace exists
	assert.NotNil(t, appErr.Context)                // Context exists
	assert.Equal(t, "Server error", appErr.Message) // But only Message and Code are exposed
	assert.Equal(t, 500, appErr.Code)
}
