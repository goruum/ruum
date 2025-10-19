package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/goruum/ruum/core"
)

// RequestIDConfig holds request ID configuration
type RequestIDConfig struct {
	Generator  func() string
	Header     string
	ContextKey string
}

// DefaultRequestIDConfig returns default configuration
func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Generator: func() string {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return ""
			}
			return hex.EncodeToString(b)
		},
		Header:     "X-Request-ID",
		ContextKey: "request_id",
	}
}

// RequestID creates a request ID middleware
func RequestID(config RequestIDConfig) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			// Check if request ID already exists
			requestID := ctx.Header(config.Header)
			if requestID == "" {
				requestID = config.Generator()
			}

			// Set request ID in context and response header
			ctx.Set(config.ContextKey, requestID)
			ctx.SetHeader(config.Header, requestID)

			return next(ctx)
		}
	}
}
