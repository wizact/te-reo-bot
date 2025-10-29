package logger

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractRequestContext(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		expectedMethod string
		expectedPath   string
		expectedUA     string
		expectedAddr   string
		expectNil      bool
	}{
		{
			name:      "nil request",
			request:   nil,
			expectNil: true,
		},
		{
			name: "basic GET request",
			request: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "/api/messages",
				},
				Header: http.Header{
					"User-Agent": []string{"curl/7.68.0"},
				},
				RemoteAddr: "192.168.1.100:12345",
			},
			expectedMethod: "GET",
			expectedPath:   "/api/messages",
			expectedUA:     "curl/7.68.0",
			expectedAddr:   "192.168.1.100:12345",
		},
		{
			name: "POST request with empty path",
			request: &http.Request{
				Method: "POST",
				URL: &url.URL{
					Path: "",
				},
				Header: http.Header{
					"User-Agent": []string{"Mozilla/5.0"},
				},
				RemoteAddr: "10.0.0.1:54321",
			},
			expectedMethod: "POST",
			expectedPath:   "",
			expectedUA:     "Mozilla/5.0",
			expectedAddr:   "10.0.0.1:54321",
		},
		{
			name: "request with X-Forwarded-For header",
			request: &http.Request{
				Method: "GET",
				URL: &url.URL{
					Path: "/health",
				},
				Header: http.Header{
					"User-Agent":      []string{"test-client"},
					"X-Forwarded-For": []string{"203.0.113.1, 198.51.100.1"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedMethod: "GET",
			expectedPath:   "/health",
			expectedUA:     "test-client",
			expectedAddr:   "203.0.113.1", // Should use first IP from X-Forwarded-For
		},
		{
			name: "request with X-Real-IP header",
			request: &http.Request{
				Method: "PUT",
				URL: &url.URL{
					Path: "/api/update",
				},
				Header: http.Header{
					"User-Agent": []string{"api-client/1.0"},
					"X-Real-Ip":  []string{"203.0.113.5"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedMethod: "PUT",
			expectedPath:   "/api/update",
			expectedUA:     "api-client/1.0",
			expectedAddr:   "203.0.113.5", // Should use X-Real-IP
		},
		{
			name: "request with both X-Forwarded-For and X-Real-IP (X-Forwarded-For takes precedence)",
			request: &http.Request{
				Method: "DELETE",
				URL: &url.URL{
					Path: "/api/delete",
				},
				Header: http.Header{
					"User-Agent":      []string{"admin-client"},
					"X-Forwarded-For": []string{"203.0.113.10"},
					"X-Real-Ip":       []string{"203.0.113.20"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedMethod: "DELETE",
			expectedPath:   "/api/delete",
			expectedUA:     "admin-client",
			expectedAddr:   "203.0.113.10", // X-Forwarded-For should take precedence
		},
		{
			name: "request with no User-Agent header",
			request: &http.Request{
				Method:     "GET",
				URL:        &url.URL{Path: "/test"},
				Header:     http.Header{},
				RemoteAddr: "127.0.0.1:8080",
			},
			expectedMethod: "GET",
			expectedPath:   "/test",
			expectedUA:     "", // Empty user agent
			expectedAddr:   "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractRequestContext(tt.request)

			if tt.expectNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedMethod, result.Method)
			assert.Equal(t, tt.expectedPath, result.Path)
			assert.Equal(t, tt.expectedUA, result.UserAgent)
			assert.Equal(t, tt.expectedAddr, result.RemoteAddr)

			// Verify request ID is generated and has correct format
			assert.NotEmpty(t, result.RequestID)
			assert.True(t, strings.HasPrefix(result.RequestID, "req-"))
			assert.True(t, len(result.RequestID) > 4) // Should be longer than just "req-"
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Test that request IDs are generated
	id1 := generateRequestID()
	id2 := generateRequestID()

	// Should have correct prefix
	assert.True(t, strings.HasPrefix(id1, "req-"))
	assert.True(t, strings.HasPrefix(id2, "req-"))

	// Should be different
	assert.NotEqual(t, id1, id2)

	// Should have expected length (req- + 16 hex chars = 20 total)
	assert.Equal(t, 20, len(id1))
	assert.Equal(t, 20, len(id2))
}

func TestExtractRemoteAddr(t *testing.T) {
	tests := []struct {
		name         string
		request      *http.Request
		expectedAddr string
	}{
		{
			name: "no proxy headers",
			request: &http.Request{
				Header:     http.Header{},
				RemoteAddr: "192.168.1.100:12345",
			},
			expectedAddr: "192.168.1.100:12345",
		},
		{
			name: "X-Forwarded-For with single IP",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.1"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For with multiple IPs",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.1, 198.51.100.1, 10.0.0.1"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.1", // Should return first IP
		},
		{
			name: "X-Forwarded-For with spaces",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"  203.0.113.1  , 198.51.100.1"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.1", // Should trim spaces
		},
		{
			name: "X-Real-IP header",
			request: &http.Request{
				Header: http.Header{
					"X-Real-Ip": []string{"203.0.113.5"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.5",
		},
		{
			name: "X-Real-IP with spaces",
			request: &http.Request{
				Header: http.Header{
					"X-Real-Ip": []string{"  203.0.113.5  "},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.5", // Should trim spaces
		},
		{
			name: "both X-Forwarded-For and X-Real-IP (X-Forwarded-For wins)",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{"203.0.113.10"},
					"X-Real-Ip":       []string{"203.0.113.20"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.10", // X-Forwarded-For should take precedence
		},
		{
			name: "empty X-Forwarded-For falls back to X-Real-IP",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{""},
					"X-Real-Ip":       []string{"203.0.113.20"},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "203.0.113.20",
		},
		{
			name: "empty headers fall back to RemoteAddr",
			request: &http.Request{
				Header: http.Header{
					"X-Forwarded-For": []string{""},
					"X-Real-Ip":       []string{""},
				},
				RemoteAddr: "10.0.0.1:12345",
			},
			expectedAddr: "10.0.0.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRemoteAddr(tt.request)
			assert.Equal(t, tt.expectedAddr, result)
		})
	}
}

func TestRequestContextToFields(t *testing.T) {
	tests := []struct {
		name     string
		context  *RequestContext
		expected []Field
	}{
		{
			name:     "nil context",
			context:  nil,
			expected: nil,
		},
		{
			name: "complete context",
			context: &RequestContext{
				Method:     "GET",
				Path:       "/api/messages",
				UserAgent:  "curl/7.68.0",
				RemoteAddr: "192.168.1.100",
				RequestID:  "req-123456789abcdef0",
			},
			expected: []Field{
				{Key: "request_method", Value: "GET"},
				{Key: "request_path", Value: "/api/messages"},
				{Key: "request_user_agent", Value: "curl/7.68.0"},
				{Key: "request_remote_addr", Value: "192.168.1.100"},
				{Key: "request_id", Value: "req-123456789abcdef0"},
			},
		},
		{
			name: "context with empty values",
			context: &RequestContext{
				Method:     "",
				Path:       "",
				UserAgent:  "",
				RemoteAddr: "",
				RequestID:  "",
			},
			expected: []Field{
				{Key: "request_method", Value: ""},
				{Key: "request_path", Value: ""},
				{Key: "request_user_agent", Value: ""},
				{Key: "request_remote_addr", Value: ""},
				{Key: "request_id", Value: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.context.ToFields()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequestContextIntegration(t *testing.T) {
	// Create a realistic HTTP request
	req := &http.Request{
		Method: "POST",
		URL: &url.URL{
			Path: "/api/messages",
		},
		Header: http.Header{
			"User-Agent":      []string{"Te-Reo-Bot/1.0"},
			"X-Forwarded-For": []string{"203.0.113.1, 10.0.0.1"},
			"Content-Type":    []string{"application/json"},
		},
		RemoteAddr: "10.0.0.1:54321",
	}

	// Extract context
	ctx := ExtractRequestContext(req)

	// Verify all fields are populated correctly
	assert.NotNil(t, ctx)
	assert.Equal(t, "POST", ctx.Method)
	assert.Equal(t, "/api/messages", ctx.Path)
	assert.Equal(t, "Te-Reo-Bot/1.0", ctx.UserAgent)
	assert.Equal(t, "203.0.113.1", ctx.RemoteAddr) // Should use X-Forwarded-For
	assert.True(t, strings.HasPrefix(ctx.RequestID, "req-"))

	// Convert to fields and verify structure
	fields := ctx.ToFields()
	assert.Len(t, fields, 5)

	// Verify field keys and values
	fieldMap := make(map[string]interface{})
	for _, field := range fields {
		fieldMap[field.Key] = field.Value
	}

	assert.Equal(t, "POST", fieldMap["request_method"])
	assert.Equal(t, "/api/messages", fieldMap["request_path"])
	assert.Equal(t, "Te-Reo-Bot/1.0", fieldMap["request_user_agent"])
	assert.Equal(t, "203.0.113.1", fieldMap["request_remote_addr"])
	assert.True(t, strings.HasPrefix(fieldMap["request_id"].(string), "req-"))
}
