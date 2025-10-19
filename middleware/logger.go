package middleware

import (
	"time"

	"github.com/goruum/ruum/core"
)

// Logger returns a logging middleware
func Logger(logger core.Logger) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			start := time.Now()
			req := ctx.Request()
			
			// Log request
			logger.Info("Incoming request", map[string]interface{}{
				"method": req.Method,
				"path":   req.URL.Path,
				"remote": req.RemoteAddr,
			})
			
			// Call next handler
			err := next(ctx)
			
			// Log response
			duration := time.Since(start)
			fields := map[string]interface{}{
				"method":   req.Method,
				"path":     req.URL.Path,
				"duration": duration.String(),
			}
			
			if err != nil {
				fields["error"] = err.Error()
				logger.Error("Request failed", fields)
			} else {
				logger.Info("Request completed", fields)
			}
			
			return err
		}
	}
}

// Recovery returns a panic recovery middleware
func Recovery(logger core.Logger) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Panic recovered", map[string]interface{}{
						"panic": r,
					})
					err = core.InternalServerErrorException("Internal server error")
				}
			}()
			
			return next(ctx)
		}
	}
}

