package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestRequestID(t *testing.T) {
	config := DefaultRequestIDConfig()
	middleware := RequestID(config)

	handler := func(ctx core.Context) error {
		requestID := ctx.GetString("request_id")
		if requestID == "" {
			t.Error("Request ID not set in context")
		}
		return nil
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler failed: %v", err)
	}

	if res.Header().Get("X-Request-ID") == "" {
		t.Error("X-Request-ID header not set")
	}
}

func TestRequestID_ExistingID(t *testing.T) {
	config := DefaultRequestIDConfig()
	middleware := RequestID(config)

	handler := func(ctx core.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "existing-id")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler failed: %v", err)
	}

	if res.Header().Get("X-Request-ID") != "existing-id" {
		t.Error("Existing request ID not preserved")
	}
}

