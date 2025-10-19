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

