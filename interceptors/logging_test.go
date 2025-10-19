package interceptors

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goruum/ruum/core"
)

type mockLogger struct {
	debugCount int
	lastMsg    string
	lastFields map[string]interface{}
}

func (m *mockLogger) Log(level string, message string, fields map[string]interface{}) {
	m.lastMsg = message
	m.lastFields = fields
}

func (m *mockLogger) Debug(message string, fields map[string]interface{}) {
	m.debugCount++
	m.lastMsg = message
	m.lastFields = fields
}

func (m *mockLogger) Info(message string, fields map[string]interface{}) {
	m.lastMsg = message
	m.lastFields = fields
}

func (m *mockLogger) Warn(message string, fields map[string]interface{}) {
	m.lastMsg = message
	m.lastFields = fields
}

func (m *mockLogger) Error(message string, fields map[string]interface{}) {
	m.lastMsg = message
	m.lastFields = fields
}

func TestNewLoggingInterceptor(t *testing.T) {
	logger := &mockLogger{}
	interceptor := NewLoggingInterceptor(logger)

	if interceptor == nil {
		t.Fatal("NewLoggingInterceptor() returned nil")
	}

	if interceptor.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestLoggingInterceptor_Intercept(t *testing.T) {
	logger := &mockLogger{}
	interceptor := NewLoggingInterceptor(logger)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}

	// Should log before and after
	if logger.debugCount < 2 {
		t.Errorf("Debug called %d times, want at least 2", logger.debugCount)
	}

	// Last message should be "After handler"
	if logger.lastMsg != "After handler" {
		t.Errorf("Last message = %v, want 'After handler'", logger.lastMsg)
	}

	// Should log duration
	if logger.lastFields["duration"] == nil {
		t.Error("Duration not logged")
	}
}

func TestLoggingInterceptor_Intercept_WithError(t *testing.T) {
	logger := &mockLogger{}
	interceptor := NewLoggingInterceptor(logger)

	expectedErr := errors.New("test error")
	handler := func(ctx core.Context) error {
		return expectedErr
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != expectedErr {
		t.Errorf("Intercept() error = %v, want %v", err, expectedErr)
	}

	// Should still log
	if logger.debugCount < 2 {
		t.Errorf("Debug called %d times, want at least 2", logger.debugCount)
	}

	// Should log the error
	if logger.lastFields["error"] == nil {
		t.Error("Error not logged")
	}
}

func TestLoggingInterceptor_LogsRequestDetails(t *testing.T) {
	logger := &mockLogger{}
	interceptor := NewLoggingInterceptor(logger)

	handler := func(ctx core.Context) error {
		return nil
	}

	req := httptest.NewRequest("POST", "/api/users", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}

	// Check if method and path were logged
	if logger.lastFields["method"] != "POST" {
		t.Errorf("Method = %v, want 'POST'", logger.lastFields["method"])
	}

	if logger.lastFields["path"] != "/api/users" {
		t.Errorf("Path = %v, want '/api/users'", logger.lastFields["path"])
	}
}

func TestNewTransformInterceptor(t *testing.T) {
	transformFunc := func(v interface{}) interface{} {
		return v
	}

	interceptor := NewTransformInterceptor(transformFunc)

	if interceptor == nil {
		t.Fatal("NewTransformInterceptor() returned nil")
	}

	if interceptor.transformFunc == nil {
		t.Error("Transform function not set")
	}
}

func TestTransformInterceptor_Intercept(t *testing.T) {
	transformFunc := func(v interface{}) interface{} {
		return "transformed"
	}

	interceptor := NewTransformInterceptor(transformFunc)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}
}

func TestTransformInterceptor_Intercept_WithError(t *testing.T) {
	transformFunc := func(v interface{}) interface{} {
		return v
	}

	interceptor := NewTransformInterceptor(transformFunc)

	expectedErr := errors.New("test error")
	handler := func(ctx core.Context) error {
		return expectedErr
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != expectedErr {
		t.Errorf("Intercept() error = %v, want %v", err, expectedErr)
	}
}

func TestNewCacheInterceptor(t *testing.T) {
	ttl := 5 * time.Minute
	interceptor := NewCacheInterceptor(ttl)

	if interceptor == nil {
		t.Fatal("NewCacheInterceptor() returned nil")
	}

	if interceptor.ttl != ttl {
		t.Errorf("TTL = %v, want %v", interceptor.ttl, ttl)
	}

	if interceptor.cache == nil {
		t.Error("Cache map not initialized")
	}
}

func TestCacheInterceptor_Intercept_CacheMiss(t *testing.T) {
	interceptor := NewCacheInterceptor(5 * time.Minute)

	handlerCalled := false
	handler := func(ctx core.Context) error {
		handlerCalled = true
		return ctx.String(200, "OK")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}

	if !handlerCalled {
		t.Error("Handler should be called on cache miss")
	}
}

func TestCacheInterceptor_Intercept_CacheHit(t *testing.T) {
	interceptor := NewCacheInterceptor(5 * time.Minute)

	// Pre-populate cache
	interceptor.cache["GET:/test"] = map[string]interface{}{
		"message": "cached response",
	}

	handlerCalled := false
	handler := func(ctx core.Context) error {
		handlerCalled = true
		return ctx.String(200, "OK")
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}

	if handlerCalled {
		t.Error("Handler should not be called on cache hit")
	}
}

func TestCacheInterceptor_Intercept_WithError(t *testing.T) {
	interceptor := NewCacheInterceptor(5 * time.Minute)

	expectedErr := errors.New("test error")
	handler := func(ctx core.Context) error {
		return expectedErr
	}

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != expectedErr {
		t.Errorf("Intercept() error = %v, want %v", err, expectedErr)
	}
}

func TestCacheInterceptor_DifferentPaths(t *testing.T) {
	interceptor := NewCacheInterceptor(5 * time.Minute)

	// Pre-populate cache for one path
	interceptor.cache["GET:/path1"] = map[string]interface{}{
		"message": "cached",
	}

	handlerCalled := false
	handler := func(ctx core.Context) error {
		handlerCalled = true
		return ctx.String(200, "OK")
	}

	// Request different path
	req := httptest.NewRequest("GET", "/path2", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := interceptor.Intercept(ctx, handler)
	if err != nil {
		t.Errorf("Intercept() error = %v", err)
	}

	// Handler should be called for different path
	if !handlerCalled {
		t.Error("Handler should be called for different path")
	}
}

