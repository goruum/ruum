package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	if len(config.AllowOrigins) == 0 {
		t.Error("AllowOrigins should not be empty")
	}

	if config.AllowOrigins[0] != "*" {
		t.Errorf("AllowOrigins[0] = %v, want '*'", config.AllowOrigins[0])
	}

	if len(config.AllowMethods) == 0 {
		t.Error("AllowMethods should not be empty")
	}

	if len(config.AllowHeaders) == 0 {
		t.Error("AllowHeaders should not be empty")
	}

	if config.MaxAge != 3600 {
		t.Errorf("MaxAge = %v, want 3600", config.MaxAge)
	}
}

func TestCORS_AllowAll(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"*"},
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %v, want '*'", allowOrigin)
	}
}

func TestCORS_SpecificOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"https://example.com", "https://test.com"},
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "https://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %v, want 'https://example.com'", allowOrigin)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"https://example.com"},
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "" {
		t.Errorf("Access-Control-Allow-Origin should be empty, got %v", allowOrigin)
	}
}

func TestCORS_AllowCredentials(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowCredentials: true,
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	allowCredentials := res.Header().Get("Access-Control-Allow-Credentials")
	if allowCredentials != "true" {
		t.Errorf("Access-Control-Allow-Credentials = %v, want 'true'", allowCredentials)
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
		MaxAge:       3600,
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	if res.Code != http.StatusNoContent {
		t.Errorf("Status code = %v, want %v", res.Code, http.StatusNoContent)
	}

	allowMethods := res.Header().Get("Access-Control-Allow-Methods")
	if allowMethods == "" {
		t.Error("Access-Control-Allow-Methods should not be empty")
	}

	allowHeaders := res.Header().Get("Access-Control-Allow-Headers")
	if allowHeaders == "" {
		t.Error("Access-Control-Allow-Headers should not be empty")
	}
}

func TestCORS_ExposeHeaders(t *testing.T) {
	config := CORSConfig{
		AllowOrigins:  []string{"*"},
		ExposeHeaders: []string{"X-Custom-Header", "X-Another-Header"},
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	exposeHeaders := res.Header().Get("Access-Control-Expose-Headers")
	if exposeHeaders == "" {
		t.Error("Access-Control-Expose-Headers should not be empty")
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	config := CORSConfig{
		AllowOrigins: []string{"*"},
	}

	middleware := CORS(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	// No Origin header set
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	// Should still set CORS headers for wildcard
	allowOrigin := res.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("Access-Control-Allow-Origin = %v, want '*'", allowOrigin)
	}
}
