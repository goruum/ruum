package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/goruum/ruum/core"
)

// TimeoutConfig holds timeout configuration
type TimeoutConfig struct {
	// Timeout is the duration after which the request will be canceled
	Timeout time.Duration
	// OnTimeout is called when a request times out
	OnTimeout func(core.Context)
	// SkipFunc determines if timeout should be skipped for a request
	SkipFunc func(core.Context) bool
}

// DefaultTimeoutConfig returns a default configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout: 30 * time.Second,
		OnTimeout: func(ctx core.Context) {
			ctx.SetHeader("X-Timeout", "true")
		},
		SkipFunc: func(_ core.Context) bool {
			return false
		},
	}
}

// ErrTimeout is returned when a request times out
var ErrTimeout = errors.New("request timeout")

// Timeout creates a timeout middleware
func Timeout(config TimeoutConfig) core.MiddlewareFunc {
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.OnTimeout == nil {
		config.OnTimeout = func(ctx core.Context) {
			ctx.SetHeader("X-Timeout", "true")
		}
	}

	if config.SkipFunc == nil {
		config.SkipFunc = func(_ core.Context) bool {
			return false
		}
	}

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			// Skip if configured
			if config.SkipFunc(ctx) {
				return next(ctx)
			}

			// Create a context with timeout
			timeoutCtx, cancel := context.WithTimeout(ctx, config.Timeout)
			defer cancel()

			// Create a channel to receive the result
			done := make(chan error, 1)

			// Run the handler in a goroutine
			go func() {
				done <- next(ctx)
			}()

			// Wait for either completion or timeout
			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				// Timeout occurred
				if config.OnTimeout != nil {
					config.OnTimeout(ctx)
				}
				return core.NewHTTPException(408, "Request Timeout")
			}
		}
	}
}

// TimeoutWithErrorHandler creates a timeout middleware with custom error handler
func TimeoutWithErrorHandler(timeout time.Duration, errorHandler func(core.Context) error) core.MiddlewareFunc {
	config := DefaultTimeoutConfig()
	config.Timeout = timeout
	
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)

			go func() {
				done <- next(ctx)
			}()

			select {
			case err := <-done:
				return err
			case <-timeoutCtx.Done():
				if errorHandler != nil {
					return errorHandler(ctx)
				}
				return core.NewHTTPException(408, "Request Timeout")
			}
		}
	}
}

