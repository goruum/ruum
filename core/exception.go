package core

import "fmt"

// HttpException represents an HTTP error
type HttpException struct {
	StatusCode int
	Message    string
	Details    interface{}
}

// NewHttpException creates a new HTTP exception
func NewHttpException(statusCode int, message string) *HttpException {
	return &HttpException{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewHttpExceptionWithDetails creates a new HTTP exception with details
func NewHttpExceptionWithDetails(statusCode int, message string, details interface{}) *HttpException {
	return &HttpException{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

func (e *HttpException) Error() string {
	if e.Details != nil {
		return fmt.Sprintf("[%d] %s: %v", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// Common HTTP exceptions
func BadRequestException(message string) *HttpException {
	return NewHttpException(400, message)
}

func UnauthorizedException(message string) *HttpException {
	return NewHttpException(401, message)
}

func ForbiddenException(message string) *HttpException {
	return NewHttpException(403, message)
}

func NotFoundException(message string) *HttpException {
	return NewHttpException(404, message)
}

func ConflictException(message string) *HttpException {
	return NewHttpException(409, message)
}

func InternalServerErrorException(message string) *HttpException {
	return NewHttpException(500, message)
}

// DefaultExceptionFilter is the default exception filter
type DefaultExceptionFilter struct{}

func NewDefaultExceptionFilter() *DefaultExceptionFilter {
	return &DefaultExceptionFilter{}
}

func (f *DefaultExceptionFilter) Catch(err error, ctx Context) error {
	httpErr, ok := err.(*HttpException)
	if !ok {
		httpErr = NewHttpException(500, err.Error())
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
