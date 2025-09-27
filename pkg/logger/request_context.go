package logger

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

// RequestContext captures HTTP request information for logging
type RequestContext struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	UserAgent  string `json:"user_agent"`
	RemoteAddr string `json:"remote_addr"`
	RequestID  string `json:"request_id"`
}

// ExtractRequestContext extracts context from HTTP request
func ExtractRequestContext(r *http.Request) *RequestContext {
	if r == nil {
		return nil
	}

	// Generate a unique request ID for tracing
	requestID := generateRequestID()

	// Extract remote address, handling X-Forwarded-For header
	remoteAddr := extractRemoteAddr(r)

	return &RequestContext{
		Method:     r.Method,
		Path:       r.URL.Path,
		UserAgent:  r.UserAgent(),
		RemoteAddr: remoteAddr,
		RequestID:  requestID,
	}
}

// generateRequestID generates a unique request ID for tracing
func generateRequestID() string {
	bytes := make([]byte, 8) // 16 character hex string
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter-based approach if crypto/rand fails
		// This is not ideal but ensures we always have some form of ID
		return "req-fallback"
	}
	return "req-" + hex.EncodeToString(bytes)
}

// extractRemoteAddr extracts the real remote address from the request
// considering proxy headers like X-Forwarded-For
func extractRemoteAddr(r *http.Request) string {
	// Check for X-Forwarded-For header (common with proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			firstIP := strings.TrimSpace(ips[0])
			if firstIP != "" {
				return firstIP
			}
		}
	}

	// Check for X-Real-IP header (another common proxy header)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		realIP := strings.TrimSpace(xri)
		if realIP != "" {
			return realIP
		}
	}

	// Fall back to RemoteAddr from the request
	return r.RemoteAddr
}

// ToFields converts RequestContext to logger fields for structured logging
func (rc *RequestContext) ToFields() []Field {
	if rc == nil {
		return nil
	}

	return []Field{
		{Key: "request_method", Value: rc.Method},
		{Key: "request_path", Value: rc.Path},
		{Key: "request_user_agent", Value: rc.UserAgent},
		{Key: "request_remote_addr", Value: rc.RemoteAddr},
		{Key: "request_id", Value: rc.RequestID},
	}
}
