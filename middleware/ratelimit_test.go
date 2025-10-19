package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goruum/ruum/core"
)

func TestRateLimiter(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.RequestsPerWindow = 2
	config.Window = time.Second

	limiter := RateLimiter(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := limiter(handler)

	// First request should succeed
	req1 := httptest.NewRequest("GET", "/", nil)
	res1 := httptest.NewRecorder()
	ctx1 := core.NewContext(context.Background(), req1, res1, core.NewContainer())

	err := wrappedHandler(ctx1)
	if err != nil {
		t.Errorf("First request failed: %v", err)
	}

	// Second request should succeed
	req2 := httptest.NewRequest("GET", "/", nil)
	res2 := httptest.NewRecorder()
	ctx2 := core.NewContext(context.Background(), req2, res2, core.NewContainer())

	err = wrappedHandler(ctx2)
	if err != nil {
		t.Errorf("Second request failed: %v", err)
	}

	// Third request should be rate limited
	req3 := httptest.NewRequest("GET", "/", nil)
	res3 := httptest.NewRecorder()
	ctx3 := core.NewContext(context.Background(), req3, res3, core.NewContainer())

	err = wrappedHandler(ctx3)
	if err == nil {
		t.Error("Third request should be rate limited")
	}
}

func TestRateLimiter_CustomKeyFunc(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.KeyFunc = func(ctx core.Context) string {
		return "custom-key"
	}

	limiter := RateLimiter(config)
	handler := func(ctx core.Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := limiter(handler)(ctx)
	if err != nil {
		t.Errorf("Request failed: %v", err)
	}
}

func TestRateLimiter_Skip(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.SkipFunc = func(ctx core.Context) bool {
		return true
	}

	limiter := RateLimiter(config)
	handler := func(ctx core.Context) error {
		return nil
	}

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := limiter(handler)(ctx)
	if err != nil {
		t.Errorf("Skipped request failed: %v", err)
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	config := DefaultRateLimiterConfig()
	config.RequestsPerWindow = 10
	config.Window = time.Millisecond * 50
	
	limiter := RateLimiter(config)
	
	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}
	
	// Make a request to create an entry
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())
	
	_ = limiter(handler)(ctx)
	
	// Wait for cleanup to potentially run and clean up expired entries
	time.Sleep(time.Millisecond * 150)
	
	// Make another request - should work as old entry should be cleaned
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	res2 := httptest.NewRecorder()
	ctx2 := core.NewContext(context.Background(), req2, res2, core.NewContainer())
	
	err := limiter(handler)(ctx2)
	if err != nil {
		t.Errorf("Request after cleanup failed: %v", err)
	}
}

