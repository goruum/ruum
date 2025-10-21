package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/goruum/ruum/core"
)

func TestDefaultTimeoutConfig(t *testing.T) {
	config := DefaultTimeoutConfig()

	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", config.Timeout)
	}

	if config.OnTimeout == nil {
		t.Error("OnTimeout should not be nil")
	}

	if config.SkipFunc == nil {
		t.Error("SkipFunc should not be nil")
	}

	// Test OnTimeout default behavior
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())
	
	config.OnTimeout(ctx)
	
	header := res.Header().Get("X-Timeout")
	if header != "true" {
		t.Errorf("X-Timeout header = %v, want 'true'", header)
	}

	// Test SkipFunc default behavior
	if config.SkipFunc(ctx) {
		t.Error("SkipFunc should return false by default")
	}
}

func TestTimeout_Success(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Timeout = 100 * time.Millisecond
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Should succeed, got error: %v", err)
	}
}

func TestTimeout_TimesOut(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Timeout = 50 * time.Millisecond
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Should return timeout error")
	}

	// Check if it's an HTTP exception with 408 status
	if httpErr, ok := err.(*core.HTTPException); ok {
		if httpErr.StatusCode != 408 {
			t.Errorf("Status code = %d, want 408", httpErr.StatusCode)
		}
	} else {
		t.Errorf("Expected HTTPException, got %T", err)
	}

	header := res.Header().Get("X-Timeout")
	if header != "true" {
		t.Errorf("X-Timeout header = %v, want 'true'", header)
	}
}

func TestTimeout_CustomOnTimeout(t *testing.T) {
	timeoutCalled := false
	config := DefaultTimeoutConfig()
	config.Timeout = 50 * time.Millisecond
	config.OnTimeout = func(ctx core.Context) {
		timeoutCalled = true
		ctx.SetHeader("X-Custom-Timeout", "yes")
	}
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	_ = wrappedHandler(ctx)

	if !timeoutCalled {
		t.Error("OnTimeout callback was not called")
	}

	header := res.Header().Get("X-Custom-Timeout")
	if header != "yes" {
		t.Errorf("X-Custom-Timeout header = %v, want 'yes'", header)
	}
}

func TestTimeout_SkipFunc(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Timeout = 50 * time.Millisecond
	config.SkipFunc = func(ctx core.Context) bool {
		return ctx.Path() == "/skip"
	}
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	// Test with skipped path
	req := httptest.NewRequest("GET", "/skip", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Should succeed for skipped path, got error: %v", err)
	}
}

func TestTimeout_DefaultValues(t *testing.T) {
	config := TimeoutConfig{
		Timeout: 0, // Should default to 30s
	}
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Should succeed, got error: %v", err)
	}
}

func TestTimeout_HandlerError(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Timeout = 100 * time.Millisecond
	middleware := Timeout(config)

	testErr := errors.New("handler error")
	handler := func(ctx core.Context) error {
		return testErr
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != testErr {
		t.Errorf("Should return handler error, got: %v", err)
	}
}

func TestTimeout_NilOnTimeout(t *testing.T) {
	config := TimeoutConfig{
		Timeout:   50 * time.Millisecond,
		OnTimeout: nil, // Should use default
	}
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Should return timeout error")
	}
}

func TestTimeout_NilSkipFunc(t *testing.T) {
	config := TimeoutConfig{
		Timeout:  100 * time.Millisecond,
		SkipFunc: nil, // Should use default
	}
	middleware := Timeout(config)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Should succeed, got error: %v", err)
	}
}

func TestTimeoutWithErrorHandler_Success(t *testing.T) {
	errorHandler := func(ctx core.Context) error {
		return core.NewHTTPException(503, "Service Unavailable")
	}

	middleware := TimeoutWithErrorHandler(100*time.Millisecond, errorHandler)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Should succeed, got error: %v", err)
	}
}

func TestTimeoutWithErrorHandler_TimesOut(t *testing.T) {
	errorHandlerCalled := false
	errorHandler := func(ctx core.Context) error {
		errorHandlerCalled = true
		ctx.SetHeader("X-Custom-Error", "timeout")
		return core.NewHTTPException(503, "Service Unavailable")
	}

	middleware := TimeoutWithErrorHandler(50*time.Millisecond, errorHandler)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Should return timeout error")
	}

	if !errorHandlerCalled {
		t.Error("Error handler was not called")
	}

	if httpErr, ok := err.(*core.HTTPException); ok {
		if httpErr.StatusCode != 503 {
			t.Errorf("Status code = %d, want 503", httpErr.StatusCode)
		}
	} else {
		t.Errorf("Expected HTTPException, got %T", err)
	}

	header := res.Header().Get("X-Custom-Error")
	if header != "timeout" {
		t.Errorf("X-Custom-Error header = %v, want 'timeout'", header)
	}
}

func TestTimeoutWithErrorHandler_NilErrorHandler(t *testing.T) {
	middleware := TimeoutWithErrorHandler(50*time.Millisecond, nil)

	handler := func(ctx core.Context) error {
		time.Sleep(100 * time.Millisecond)
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Should return timeout error")
	}

	// Should use default 408 error
	if httpErr, ok := err.(*core.HTTPException); ok {
		if httpErr.StatusCode != 408 {
			t.Errorf("Status code = %d, want 408", httpErr.StatusCode)
		}
	}
}

func TestTimeout_HandlerReturnsError(t *testing.T) {
	config := DefaultTimeoutConfig()
	config.Timeout = 100 * time.Millisecond
	middleware := Timeout(config)

	testErr := errors.New("test error")
	handler := func(ctx core.Context) error {
		time.Sleep(20 * time.Millisecond)
		return testErr
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != testErr {
		t.Errorf("Should return handler error, got: %v", err)
	}
}

func TestErrTimeout_Variable(t *testing.T) {
	if ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}

	if ErrTimeout.Error() != "request timeout" {
		t.Errorf("ErrTimeout message = %v, want 'request timeout'", ErrTimeout.Error())
	}
}

