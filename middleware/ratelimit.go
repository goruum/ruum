package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/goruum/ruum/core"
)

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerWindow int           // Number of requests allowed per window
	Window            time.Duration // Time window
	KeyFunc           func(core.Context) string
	SkipFunc          func(core.Context) bool
	OnLimitReached    func(core.Context) error
}

// DefaultRateLimiterConfig returns default configuration
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerWindow: 100,
		Window:            time.Minute,
		KeyFunc: func(ctx core.Context) string {
			return ctx.RemoteAddr()
		},
		SkipFunc: func(_ core.Context) bool {
			return false
		},
		OnLimitReached: func(ctx core.Context) error {
			return core.NewHTTPException(http.StatusTooManyRequests, "Too many requests")
		},
	}
}

type client struct {
	requests  int
	resetTime time.Time
	mu        sync.Mutex
}

type rateLimiter struct {
	clients map[string]*client
	mu      sync.RWMutex
	config  RateLimiterConfig
}

// RateLimiter creates a rate limiting middleware
func RateLimiter(config RateLimiterConfig) core.MiddlewareFunc {
	limiter := &rateLimiter{
		clients: make(map[string]*client),
		config:  config,
	}

	// Cleanup goroutine
	go limiter.cleanup()

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			if limiter.config.SkipFunc(ctx) {
				return next(ctx)
			}

			key := limiter.config.KeyFunc(ctx)

			limiter.mu.Lock()
			c, exists := limiter.clients[key]
			if !exists {
				c = &client{
					requests:  0,
					resetTime: time.Now().Add(limiter.config.Window),
				}
				limiter.clients[key] = c
			}
			limiter.mu.Unlock()

			c.mu.Lock()
			defer c.mu.Unlock()

			// Reset if window expired
			if time.Now().After(c.resetTime) {
				c.requests = 0
				c.resetTime = time.Now().Add(limiter.config.Window)
			}

			// Check limit
			if c.requests >= limiter.config.RequestsPerWindow {
				remaining := time.Until(c.resetTime).Seconds()
				ctx.SetHeader("X-RateLimit-Limit", string(rune(limiter.config.RequestsPerWindow)))
				ctx.SetHeader("X-RateLimit-Remaining", "0")
				ctx.SetHeader("X-RateLimit-Reset", string(rune(int(remaining))))
				return limiter.config.OnLimitReached(ctx)
			}

			c.requests++

			// Set rate limit headers
			remaining := limiter.config.RequestsPerWindow - c.requests
			resetIn := time.Until(c.resetTime).Seconds()
			ctx.SetHeader("X-RateLimit-Limit", string(rune(limiter.config.RequestsPerWindow)))
			ctx.SetHeader("X-RateLimit-Remaining", string(rune(remaining)))
			ctx.SetHeader("X-RateLimit-Reset", string(rune(int(resetIn))))

			return next(ctx)
		}
	}
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, client := range rl.clients {
			client.mu.Lock()
			if now.After(client.resetTime.Add(rl.config.Window)) {
				delete(rl.clients, key)
			}
			client.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}
