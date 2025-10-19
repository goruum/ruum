package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

type mockLogger struct {
	infoCount  int
	errorCount int
	lastMsg    string
}

func (m *mockLogger) Log(level string, message string, fields map[string]interface{}) {
	m.lastMsg = message
}

func (m *mockLogger) Debug(message string, fields map[string]interface{}) {
	m.lastMsg = message
}

func (m *mockLogger) Info(message string, fields map[string]interface{}) {
	m.infoCount++
	m.lastMsg = message
}

func (m *mockLogger) Warn(message string, fields map[string]interface{}) {
	m.lastMsg = message
}

func (m *mockLogger) Error(message string, fields map[string]interface{}) {
	m.errorCount++
	m.lastMsg = message
}

func TestLogger(t *testing.T) {
	logger := &mockLogger{}
	middleware := Logger(logger)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	// Should log incoming and completed requests
	if logger.infoCount < 2 {
		t.Errorf("Info called %d times, want at least 2", logger.infoCount)
	}
}

func TestLogger_WithError(t *testing.T) {
	logger := &mockLogger{}
	middleware := Logger(logger)

	expectedErr := errors.New("test error")
	handler := func(ctx core.Context) error {
		return expectedErr
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != expectedErr {
		t.Errorf("Handler returned error = %v, want %v", err, expectedErr)
	}

	// Should log error
	if logger.errorCount == 0 {
		t.Error("Error was not logged")
	}

	if logger.lastMsg != "Request failed" {
		t.Errorf("Last message = %v, want 'Request failed'", logger.lastMsg)
	}
}

func TestLogger_LogsRequestDetails(t *testing.T) {
	logger := &mockLogger{}
	middleware := Logger(logger)

	handler := func(ctx core.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("POST", "/api/users", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	// Logger should have been called
	if logger.infoCount == 0 {
		t.Error("Logger was not called")
	}
}

func TestRecovery(t *testing.T) {
	logger := &mockLogger{}
	middleware := Recovery(logger)

	handler := func(ctx core.Context) error {
		panic("test panic")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)

	// Should return an error instead of panicking
	if err == nil {
		t.Error("Expected error after panic recovery, got nil")
	}

	httpErr, ok := err.(*core.HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", httpErr.StatusCode)
	}

	// Should log the panic
	if logger.errorCount == 0 {
		t.Error("Panic was not logged")
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	logger := &mockLogger{}
	middleware := Recovery(logger)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Handler error = %v", err)
	}

	// Should not log error if no panic
	if logger.errorCount != 0 {
		t.Errorf("Error count = %v, want 0", logger.errorCount)
	}
}

func TestRecovery_WithHandlerError(t *testing.T) {
	logger := &mockLogger{}
	middleware := Recovery(logger)

	expectedErr := core.BadRequestException("bad request")
	handler := func(ctx core.Context) error {
		return expectedErr
	}

	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	err := wrappedHandler(ctx)

	// Should return the handler error, not a panic error
	if err != expectedErr {
		t.Errorf("Handler returned error = %v, want %v", err, expectedErr)
	}
}
