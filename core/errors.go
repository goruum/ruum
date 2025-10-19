// Package core provides advanced error handling with stack traces.
package core

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// StackFrame represents a single stack frame
type StackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// ErrorWithStack is an error with stack trace
type ErrorWithStack struct {
	Err        error                  `json:"-"`
	Message    string                 `json:"message"`
	Stack      []StackFrame           `json:"stack,omitempty"`
	StatusCode int                    `json:"statusCode,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// Error implements error interface
func (e *ErrorWithStack) Error() string {
	return e.Message
}

// NewErrorWithStack creates a new error with stack trace
func NewErrorWithStack(err error, skip int) *ErrorWithStack {
	if err == nil {
		return nil
	}

	e := &ErrorWithStack{
		Err:     err,
		Message: err.Error(),
		Stack:   captureStack(skip + 1),
		Context: make(map[string]interface{}),
	}

	return e
}

// WithContext adds context to the error
func (e *ErrorWithStack) WithContext(key string, value interface{}) *ErrorWithStack {
	e.Context[key] = value
	return e
}

// WithStatusCode sets the HTTP status code
func (e *ErrorWithStack) WithStatusCode(code int) *ErrorWithStack {
	e.StatusCode = code
	return e
}

// captureStack captures the current stack trace
func captureStack(skip int) []StackFrame {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+2, pcs[:])

	frames := make([]StackFrame, 0, n)
	for i := 0; i < n; i++ {
		pc := pcs[i]
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}

		file, line := fn.FileLine(pc)

		// Skip runtime and testing frames
		if strings.HasPrefix(file, "runtime/") || strings.HasPrefix(file, "testing/") {
			continue
		}

		frames = append(frames, StackFrame{
			Function: fn.Name(),
			File:     file,
			Line:     line,
		})
	}

	return frames
}

// WrapError wraps an error with stack trace
func WrapError(err error, message string) *ErrorWithStack {
	if err == nil {
		return nil
	}

	wrapped := NewErrorWithStack(err, 1)
	if message != "" {
		wrapped.Message = fmt.Sprintf("%s: %s", message, err.Error())
	}

	return wrapped
}

// Advanced exception filter with stack traces
type AdvancedExceptionFilter struct {
	logger      Logger
	showStack   bool
	showContext bool
}

// NewAdvancedExceptionFilter creates a new advanced exception filter
func NewAdvancedExceptionFilter(logger Logger, showStack, showContext bool) *AdvancedExceptionFilter {
	return &AdvancedExceptionFilter{
		logger:      logger,
		showStack:   showStack,
		showContext: showContext,
	}
}

// Catch handles exceptions with stack traces
func (f *AdvancedExceptionFilter) Catch(err error, ctx Context) error {
	if err == nil {
		return nil
	}

	var response map[string]interface{}
	statusCode := http.StatusInternalServerError

	// Check if it's an HTTP exception
	if httpErr, ok := err.(*HTTPException); ok {
		statusCode = httpErr.StatusCode
		response = map[string]interface{}{
			"statusCode": httpErr.StatusCode,
			"message":    httpErr.Message,
			"error":      http.StatusText(httpErr.StatusCode),
		}

		if httpErr.Details != nil {
			response["details"] = httpErr.Details
		}
	} else if stackErr, ok := err.(*ErrorWithStack); ok {
		// Error with stack trace
		statusCode = stackErr.StatusCode
		if statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}

		response = map[string]interface{}{
			"statusCode": statusCode,
			"message":    stackErr.Message,
			"error":      http.StatusText(statusCode),
		}

		if f.showStack && len(stackErr.Stack) > 0 {
			response["stack"] = stackErr.Stack
		}

		if f.showContext && len(stackErr.Context) > 0 {
			response["context"] = stackErr.Context
		}
	} else {
		// Generic error
		response = map[string]interface{}{
			"statusCode": statusCode,
			"message":    err.Error(),
			"error":      http.StatusText(statusCode),
		}
	}

	// Add request ID if available
	if requestID := ctx.GetString("request_id"); requestID != "" {
		response["requestId"] = requestID
	}

	// Add timestamp
	response["timestamp"] = ctx.Value("timestamp")

	// Log error
	if f.logger != nil {
		logFields := map[string]interface{}{
			"error":  err.Error(),
			"path":   ctx.Path(),
			"method": ctx.Method(),
		}

		if requestID := ctx.GetString("request_id"); requestID != "" {
			logFields["requestId"] = requestID
		}

		if statusCode >= 500 {
			f.logger.Error("Internal server error", logFields)
		} else {
			f.logger.Warn("Client error", logFields)
		}
	}

	return ctx.JSON(statusCode, response)
}

// Common error constructors with stack traces

// InternalError creates an internal server error with stack
func InternalError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusInternalServerError)
}

// BadRequestError creates a bad request error with stack
func BadRequestError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusBadRequest)
}

// UnauthorizedError creates an unauthorized error with stack
func UnauthorizedError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusUnauthorized)
}

// ForbiddenError creates a forbidden error with stack
func ForbiddenError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusForbidden)
}

// NotFoundError creates a not found error with stack
func NotFoundError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusNotFound)
}

// ConflictError creates a conflict error with stack
func ConflictError(message string) *ErrorWithStack {
	return NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusConflict)
}

// ValidationError creates a validation error with stack
func ValidationError(message string, details map[string]interface{}) *ErrorWithStack {
	err := NewErrorWithStack(fmt.Errorf("%s", message), 1).WithStatusCode(http.StatusUnprocessableEntity)
	for k, v := range details {
		_ = err.WithContext(k, v)
	}
	return err
}
