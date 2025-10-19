package core

import "fmt"

// HTTPException represents an HTTP error with status code and details.
type HTTPException struct {
	StatusCode int
	Message    string
	Details    interface{}
}

// NewHTTPException creates a new HTTP exception.
func NewHTTPException(statusCode int, message string) *HTTPException {
	return &HTTPException{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewHTTPExceptionWithDetails creates a new HTTP exception with details.
func NewHTTPExceptionWithDetails(statusCode int, message string, details interface{}) *HTTPException {
	return &HTTPException{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

func (e *HTTPException) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%d] %s: %v", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// BadRequestException creates a 400 Bad Request exception.
func BadRequestException(message string) *HTTPException {
	return NewHTTPException(400, message)
}

// UnauthorizedException creates a 401 Unauthorized exception.
func UnauthorizedException(message string) *HTTPException {
	return NewHTTPException(401, message)
}

// ForbiddenException creates a 403 Forbidden exception.
func ForbiddenException(message string) *HTTPException {
	return NewHTTPException(403, message)
}

// NotFoundException creates a 404 Not Found exception.
func NotFoundException(message string) *HTTPException {
	return NewHTTPException(404, message)
}

// ConflictException creates a 409 Conflict exception.
func ConflictException(message string) *HTTPException {
	return NewHTTPException(409, message)
}

// InternalServerErrorException creates a 500 Internal Server Error exception.
func InternalServerErrorException(message string) *HTTPException {
	return NewHTTPException(500, message)
}

// DefaultExceptionFilter is the default exception filter
type DefaultExceptionFilter struct{}

// NewDefaultExceptionFilter creates a new default exception filter.
func NewDefaultExceptionFilter() *DefaultExceptionFilter {
	return &DefaultExceptionFilter{}
}

// Catch handles exceptions and returns appropriate HTTP responses.
func (f *DefaultExceptionFilter) Catch(err error, ctx Context) error {
	httpErr, ok := err.(*HTTPException)
	if !ok {
		httpErr = NewHTTPException(500, err.Error())
	}

	response := map[string]interface{}{
		"statusCode": httpErr.StatusCode,
		"message":    httpErr.Message,
	}

	if httpErr.Details != nil {
		response["details"] = httpErr.Details
	}

	return ctx.JSON(httpErr.StatusCode, response)
}
