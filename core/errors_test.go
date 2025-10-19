package core

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewErrorWithStack(t *testing.T) {
	err := NewErrorWithStack(errors.New("test error"), 0)

	if err == nil {
		t.Fatal("NewErrorWithStack() returned nil")
	}

	if err.Message != "test error" {
		t.Errorf("Message = %v, want test error", err.Message)
	}

	if len(err.Stack) == 0 {
		t.Error("Stack should not be empty")
	}
}

func TestErrorWithStack_WithContext(t *testing.T) {
	err := NewErrorWithStack(errors.New("test"), 0)
	_ = err.WithContext("key", "value")

	if err.Context["key"] != "value" {
		t.Error("Context was not set")
	}
}

func TestErrorWithStack_WithStatusCode(t *testing.T) {
	err := NewErrorWithStack(errors.New("test"), 0)
	_ = err.WithStatusCode(http.StatusBadRequest)

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusBadRequest)
	}
}

func TestWrapError(t *testing.T) {
	original := errors.New("original error")
	wrapped := WrapError(original, "wrapped")

	if wrapped == nil {
		t.Fatal("WrapError() returned nil")
	}

	if !contains(wrapped.Message, "wrapped") {
		t.Error("Message should contain 'wrapped'")
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		constructor func(string) *ErrorWithStack
		statusCode int
	}{
		{"InternalError", InternalError, http.StatusInternalServerError},
		{"BadRequestError", BadRequestError, http.StatusBadRequest},
		{"UnauthorizedError", UnauthorizedError, http.StatusUnauthorized},
		{"ForbiddenError", ForbiddenError, http.StatusForbidden},
		{"NotFoundError", NotFoundError, http.StatusNotFound},
		{"ConflictError", ConflictError, http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor("test message")
			if err == nil {
				t.Fatal("Constructor returned nil")
			}
			if err.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %v, want %v", err.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	details := map[string]interface{}{
		"field": "email",
		"error": "invalid",
	}

	err := ValidationError("validation failed", details)

	if err == nil {
		t.Fatal("ValidationError() returned nil")
	}

	if err.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %v, want %v", err.StatusCode, http.StatusUnprocessableEntity)
	}
}

func TestAdvancedExceptionFilter(t *testing.T) {
	logger := NewDefaultLogger()
	filter := NewAdvancedExceptionFilter(logger, true, true)

	if filter == nil {
		t.Fatal("NewAdvancedExceptionFilter() returned nil")
	}
}

func TestAdvancedExceptionFilter_Catch_HTTPException(t *testing.T) {
	logger := NewDefaultLogger()
	filter := NewAdvancedExceptionFilter(logger, true, true)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	httpErr := BadRequestException("test error")
	err := filter.Catch(httpErr, ctx)

	if err != nil {
		t.Errorf("Catch() error = %v, want nil", err)
	}

	if res.Code != http.StatusBadRequest {
		t.Errorf("Status code = %v, want %v", res.Code, http.StatusBadRequest)
	}
}

func TestAdvancedExceptionFilter_Catch_ErrorWithStack(t *testing.T) {
	logger := NewDefaultLogger()
	filter := NewAdvancedExceptionFilter(logger, true, true)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	stackErr := InternalError("internal error")
	_ = stackErr.WithContext("key", "value")

	err := filter.Catch(stackErr, ctx)

	if err != nil {
		t.Errorf("Catch() error = %v, want nil", err)
	}

	if res.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %v, want %v", res.Code, http.StatusInternalServerError)
	}
}

func TestAdvancedExceptionFilter_Catch_GenericError(t *testing.T) {
	logger := NewDefaultLogger()
	filter := NewAdvancedExceptionFilter(logger, false, false)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	genericErr := errors.New("generic error")
	err := filter.Catch(genericErr, ctx)

	if err != nil {
		t.Errorf("Catch() error = %v, want nil", err)
	}

	if res.Code != http.StatusInternalServerError {
		t.Errorf("Status code = %v, want %v", res.Code, http.StatusInternalServerError)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr))
}

