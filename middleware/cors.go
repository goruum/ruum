// Package middleware provides HTTP middleware functions.
package middleware

import (
	"net/http"

	"github.com/goruum/ruum/core"
)

// CORSConfig contains CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// CORS returns a CORS middleware
func CORS(config CORSConfig) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			req := ctx.Request()
			origin := req.Header.Get("Origin")

			// Check if origin is allowed
			allowOrigin := ""
			if len(config.AllowOrigins) > 0 {
				if config.AllowOrigins[0] == "*" {
					allowOrigin = "*"
				} else {
					for _, o := range config.AllowOrigins {
						if o == origin {
							allowOrigin = origin
							break
						}
					}
				}
			}

			if allowOrigin != "" {
				ctx.SetHeader("Access-Control-Allow-Origin", allowOrigin)
			}

			if config.AllowCredentials {
				ctx.SetHeader("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight request
			if req.Method == http.MethodOptions {
				ctx.SetHeader("Access-Control-Allow-Methods", joinStrings(config.AllowMethods, ", "))
				ctx.SetHeader("Access-Control-Allow-Headers", joinStrings(config.AllowHeaders, ", "))

				if len(config.ExposeHeaders) > 0 {
					ctx.SetHeader("Access-Control-Expose-Headers", joinStrings(config.ExposeHeaders, ", "))
				}

				if config.MaxAge > 0 {
					ctx.SetHeader("Access-Control-Max-Age", string(rune(config.MaxAge)))
				}

				ctx.Status(http.StatusNoContent)
				return nil
			}

			return next(ctx)
		}
	}
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
