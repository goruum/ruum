package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHTTPException(t *testing.T) {
	exc := NewHTTPException(400, "test error")
	if exc.StatusCode != 400 {
		t.Errorf("StatusCode = %v, want 400", exc.StatusCode)
	}
	if exc.Message != "test error" {
		t.Errorf("Message = %v, want 'test error'", exc.Message)
	}
	if exc.Details != nil {
		t.Error("Details should be nil")
	}
}

func TestNewHTTPExceptionWithDetails(t *testing.T) {
	details := map[string]interface{}{"field": "value"}
	exc := NewHTTPExceptionWithDetails(422, "validation failed", details)

	if exc.StatusCode != 422 {
		t.Errorf("StatusCode = %v, want 422", exc.StatusCode)
	}
	if exc.Message != "validation failed" {
		t.Errorf("Message = %v, want 'validation failed'", exc.Message)
	}
	if exc.Details == nil {
		t.Error("Details should not be nil")
	}
}

func TestHTTPException_Error(t *testing.T) {
	tests := []struct {
		name     string
		exc      *HTTPException
		contains string
	}{
		{
			name:     "without details",
			exc:      NewHTTPException(404, "not found"),
			contains: "[404] not found",
		},
		{
			name:     "with details",
			exc:      NewHTTPExceptionWithDetails(400, "bad request", map[string]string{"error": "invalid"}),
			contains: "[400] bad request:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.exc.Error()
			if err == "" {
				t.Error("Error() returned empty string")
			}
		})
	}
}

func TestBadRequestException(t *testing.T) {
	exc := BadRequestException("invalid input")
	if exc.StatusCode != 400 {
		t.Errorf("StatusCode = %v, want 400", exc.StatusCode)
	}
	if exc.Message != "invalid input" {
		t.Errorf("Message = %v, want 'invalid input'", exc.Message)
	}
}

func TestUnauthorizedException(t *testing.T) {
	exc := UnauthorizedException("not authenticated")
	if exc.StatusCode != 401 {
		t.Errorf("StatusCode = %v, want 401", exc.StatusCode)
	}
}

func TestForbiddenException(t *testing.T) {
	exc := ForbiddenException("access denied")
	if exc.StatusCode != 403 {
		t.Errorf("StatusCode = %v, want 403", exc.StatusCode)
	}
}

func TestNotFoundException(t *testing.T) {
	exc := NotFoundException("resource not found")
	if exc.StatusCode != 404 {
		t.Errorf("StatusCode = %v, want 404", exc.StatusCode)
	}
}

func TestConflictException(t *testing.T) {
	exc := ConflictException("resource conflict")
	if exc.StatusCode != 409 {
		t.Errorf("StatusCode = %v, want 409", exc.StatusCode)
	}
}

func TestInternalServerErrorException(t *testing.T) {
	exc := InternalServerErrorException("server error")
	if exc.StatusCode != 500 {
		t.Errorf("StatusCode = %v, want 500", exc.StatusCode)
	}
}

func TestDefaultExceptionFilter_Catch(t *testing.T) {
	filter := NewDefaultExceptionFilter()

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		shouldContain  string
	}{
		{
			name:           "http exception",
			err:            BadRequestException("invalid"),
			expectedStatus: 400,
			shouldContain:  "invalid",
		},
		{
			name:           "regular error",
			err:            http.ErrServerClosed,
			expectedStatus: 500,
			shouldContain:  "http: Server closed",
		},
		{
			name:           "http exception with details",
			err:            NewHTTPExceptionWithDetails(422, "validation error", map[string]string{"field": "invalid"}),
			expectedStatus: 422,
			shouldContain:  "validation error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			res := httptest.NewRecorder()
			ctx := NewContext(req.Context(), req, res, NewContainer())

			err := filter.Catch(tt.err, ctx)
			if err != nil {
				t.Errorf("Catch() returned error: %v", err)
			}

			if res.Code != tt.expectedStatus {
				t.Errorf("Response code = %v, want %v", res.Code, tt.expectedStatus)
			}
		})
	}
}
