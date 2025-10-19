package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/goruum/ruum/core"
)

// CacheEntry represents a cached response
type CacheEntry struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Expiration time.Time
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	TTL         time.Duration
	KeyFunc     func(core.Context) string
	SkipFunc    func(core.Context) bool
	Methods     []string
	StatusCodes []int
}

// DefaultCacheConfig returns default configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		TTL: 5 * time.Minute,
		KeyFunc: func(ctx core.Context) string {
			hash := sha256.Sum256([]byte(ctx.Path() + ctx.Request().URL.RawQuery))
			return hex.EncodeToString(hash[:])
		},
		SkipFunc: func(_ core.Context) bool {
			return false
		},
		Methods:     []string{"GET", "HEAD"},
		StatusCodes: []int{http.StatusOK},
	}
}

type cacheStore struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

func newCacheStore() *cacheStore {
	store := &cacheStore{
		entries: make(map[string]*CacheEntry),
	}
	go store.cleanup()
	return store
}

func (cs *cacheStore) Get(key string) (*CacheEntry, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	entry, exists := cs.entries[key]
	if !exists || time.Now().After(entry.Expiration) {
		return nil, false
	}

	return entry, true
}

func (cs *cacheStore) Set(key string, entry *CacheEntry) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.entries[key] = entry
}

func (cs *cacheStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cs.mu.Lock()
		now := time.Now()
		for key, entry := range cs.entries {
			if now.After(entry.Expiration) {
				delete(cs.entries, key)
			}
		}
		cs.mu.Unlock()
	}
}

type responseCapture struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

// Cache creates a caching middleware
func Cache(config CacheConfig) core.MiddlewareFunc {
	store := newCacheStore()

	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			// Skip if configured
			if config.SkipFunc(ctx) {
				return next(ctx)
			}

			// Check if method should be cached
			methodAllowed := false
			for _, method := range config.Methods {
				if ctx.Method() == method {
					methodAllowed = true
					break
				}
			}
			if !methodAllowed {
				return next(ctx)
			}

			// Generate cache key
			key := config.KeyFunc(ctx)

			// Try to get from cache
			if entry, found := store.Get(key); found {
				// Set cached headers
				for k, v := range entry.Headers {
					for _, val := range v {
						ctx.AddHeader(k, val)
					}
				}
				ctx.SetHeader("X-Cache", "HIT")

				ctx.Response().WriteHeader(entry.StatusCode)
				_, err := ctx.Response().Write(entry.Body)
				return err
			}

			// Capture response
			capture := &responseCapture{
				ResponseWriter: ctx.Response(),
				statusCode:     http.StatusOK,
				body:           new(bytes.Buffer),
			}

			// Create new context with capture writer
			wrappedCtx := core.NewContext(
				ctx,
				ctx.Request(),
				capture,
				ctx.Container(),
			)

			// Call next handler
			err := next(wrappedCtx)
			if err != nil {
				return err
			}

			// Check if status code should be cached
			shouldCache := false
			for _, code := range config.StatusCodes {
				if capture.statusCode == code {
					shouldCache = true
					break
				}
			}

			// Store in cache
			if shouldCache {
				entry := &CacheEntry{
					StatusCode: capture.statusCode,
					Headers:    capture.Header(),
					Body:       capture.body.Bytes(),
					Expiration: time.Now().Add(config.TTL),
				}
				store.Set(key, entry)
			}

			ctx.SetHeader("X-Cache", "MISS")
			return nil
		}
	}
}
