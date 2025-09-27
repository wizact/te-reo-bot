package entities

import "github.com/wizact/te-reo-bot/pkg/logger"

// AppError as app error container with stack trace support
type AppError struct {
	Err        error                  `json:"-"` // Don't expose internal error in JSON
	Message    string                 `json:"message"`
	Code       int                    `json:"code"`
	StackTrace *logger.StackTrace     `json:"-"` // Don't expose stack trace in JSON
	Context    map[string]interface{} `json:"-"` // Don't expose context in JSON
}

// NewAppError creates a new AppError with stack trace capture
// skip parameter allows skipping stack frames (typically 0 for direct calls)
func NewAppError(err error, code int, message string) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		Code:       code,
		StackTrace: logger.CaptureStackTrace(1), // Skip this function call
		Context:    make(map[string]interface{}),
	}
}

// WithContext adds contextual information to the error for debugging
func (ae *AppError) WithContext(key string, value interface{}) *AppError {
	if ae.Context == nil {
		ae.Context = make(map[string]interface{})
	}
	ae.Context[key] = value
	return ae
}

// GetContext retrieves a context value by key
func (ae *AppError) GetContext(key string) (interface{}, bool) {
	if ae.Context == nil {
		return nil, false
	}
	value, exists := ae.Context[key]
	return value, exists
}

// HasStackTrace returns true if the error has a captured stack trace
func (ae *AppError) HasStackTrace() bool {
	return ae.StackTrace != nil && !ae.StackTrace.IsEmpty()
}

// Error implements the error interface
func (ae *AppError) Error() string {
	if ae.Err != nil {
		return ae.Err.Error()
	}
	return ae.Message
}

// PostResponse is the tweet/mastodon Id after a successful update operation
type PostResponse struct {
	TwitterId string `json:"tweetId"`
	TootId    string `json:"tootId"`
	Message   string `json:"message"`
}

// FriendlyError is sanitised error message sent back to the user
type FriendlyError struct {
	Message string `json:"message"`
}
