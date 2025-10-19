package interceptors

import (
	"time"

	"github.com/goruum/ruum/core"
)

// LoggingInterceptor logs request and response
type LoggingInterceptor struct {
	logger core.Logger
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(logger core.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

func (i *LoggingInterceptor) Intercept(ctx core.Context, next core.HandlerFunc) error {
	start := time.Now()

	i.logger.Debug("Before handler", map[string]interface{}{
		"method": ctx.Request().Method,
		"path":   ctx.Request().URL.Path,
	})

	err := next(ctx)

	duration := time.Since(start)
	i.logger.Debug("After handler", map[string]interface{}{
		"method":   ctx.Request().Method,
		"path":     ctx.Request().URL.Path,
		"duration": duration.String(),
		"error":    err,
	})

	return err
}

// TransformInterceptor transforms response data
type TransformInterceptor struct {
	transformFunc func(interface{}) interface{}
}

// NewTransformInterceptor creates a new transform interceptor
func NewTransformInterceptor(fn func(interface{}) interface{}) *TransformInterceptor {
	return &TransformInterceptor{
		transformFunc: fn,
	}
}

func (i *TransformInterceptor) Intercept(ctx core.Context, next core.HandlerFunc) error {
	// Execute handler
	err := next(ctx)
	if err != nil {
		return err
	}

	// Transform response (if needed)
	// This is a simplified version - in a real implementation,
	// you'd need to intercept the response writer

	return nil
}

// CacheInterceptor caches responses
type CacheInterceptor struct {
	cache map[string]interface{}
	ttl   time.Duration
}

// NewCacheInterceptor creates a new cache interceptor
func NewCacheInterceptor(ttl time.Duration) *CacheInterceptor {
	return &CacheInterceptor{
		cache: make(map[string]interface{}),
		ttl:   ttl,
	}
}

func (i *CacheInterceptor) Intercept(ctx core.Context, next core.HandlerFunc) error {
	// Generate cache key
	key := ctx.Request().Method + ":" + ctx.Request().URL.Path

	// Check cache
	if cached, exists := i.cache[key]; exists {
		// Return cached response
		return ctx.JSON(200, cached)
	}

	// Execute handler
	err := next(ctx)
	if err != nil {
		return err
	}

	// Store in cache (simplified version)
	// In a real implementation, you'd need to capture the response

	return nil
}
