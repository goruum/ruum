// Package middleware provides HTTP middleware functions.
package middleware

import (
	"net/http"
	"strconv"
	"strings"

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

			// Set CORS headers
			allowOrigin := getAllowedOrigin(config.AllowOrigins, origin)
			setCORSHeaders(ctx, config, allowOrigin)

			// Handle preflight request
			if req.Method == http.MethodOptions {
				handlePreflightRequest(ctx, config)
				return nil
			}

			return next(ctx)
		}
	}
}

// getAllowedOrigin checks if the origin is allowed and returns it
func getAllowedOrigin(allowOrigins []string, origin string) string {
	if len(allowOrigins) == 0 {
		return ""
	}

	if allowOrigins[0] == "*" {
		return "*"
	}

	for _, o := range allowOrigins {
		if o == origin {
			return origin
		}
	}

	return ""
}

// setCORSHeaders sets the basic CORS headers
func setCORSHeaders(ctx core.Context, config CORSConfig, allowOrigin string) {
	if allowOrigin != "" {
		ctx.SetHeader("Access-Control-Allow-Origin", allowOrigin)
	}

	if config.AllowCredentials {
		ctx.SetHeader("Access-Control-Allow-Credentials", "true")
	}
}

// handlePreflightRequest handles OPTIONS preflight requests
func handlePreflightRequest(ctx core.Context, config CORSConfig) {
	ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
	ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

	if len(config.ExposeHeaders) > 0 {
		ctx.SetHeader("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
	}

	if config.MaxAge > 0 {
		ctx.SetHeader("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
	}

	ctx.Status(http.StatusNoContent)
}
