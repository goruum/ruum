package middleware

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestHealthCheckMiddleware(t *testing.T) {
	config := DefaultHealthCheckConfig()
	config.Path = "/health"
	config.Checks = map[string]func() bool{
		"database": func() bool { return true },
	}

	middleware := HealthCheckMiddleware(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	t.Run("health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if res.Code != 200 {
			t.Errorf("Status = %d, want 200", res.Code)
		}

		var result map[string]interface{}
		_ = json.Unmarshal(res.Body.Bytes(), &result)
		if result["status"] != "healthy" {
			t.Error("Status should be healthy")
		}
	})

	t.Run("liveness endpoint", func(t *testing.T) {
		config.LivePath = "/health/live"
		middleware := HealthCheckMiddleware(config)

		req := httptest.NewRequest("GET", "/health/live", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if res.Code != 200 {
			t.Errorf("Status = %d, want 200", res.Code)
		}
	})

	t.Run("readiness endpoint", func(t *testing.T) {
		config.ReadyPath = "/health/ready"
		middleware := HealthCheckMiddleware(config)

		req := httptest.NewRequest("GET", "/health/ready", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}

		if res.Code != 200 {
			t.Errorf("Status = %d, want 200", res.Code)
		}
	})

	t.Run("unhealthy check", func(t *testing.T) {
		config := DefaultHealthCheckConfig()
		config.Path = "/health"
		config.Checks = map[string]func() bool{
			"database": func() bool { return false },
		}

		middleware := HealthCheckMiddleware(config)

		req := httptest.NewRequest("GET", "/health", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		_ = middleware(handler)(ctx)

		if res.Code != 503 {
			t.Errorf("Status = %d, want 503", res.Code)
		}
	})

	t.Run("non-health endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/other", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}
	})
}

