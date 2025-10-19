package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestCompression(t *testing.T) {
	config := DefaultCompressionConfig()
	middleware := Compression(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "This is a test response that should be compressed")
	}

	t.Run("with gzip accept", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}
	})

	t.Run("without gzip accept", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}
	})

	t.Run("skip function", func(t *testing.T) {
		config := DefaultCompressionConfig()
		config.SkipFunc = func(ctx core.Context) bool {
			return true
		}
		middleware := Compression(config)

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		res := httptest.NewRecorder()
		ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

		err := middleware(handler)(ctx)
		if err != nil {
			t.Errorf("Handler failed: %v", err)
		}
	})
}

