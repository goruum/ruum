package core

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	container := NewContainer()

	ctx := NewContext(context.Background(), req, res, container)

	if ctx == nil {
		t.Fatal("NewContext() returned nil")
	}

	if ctx.Request() != req {
		t.Error("Request() does not match")
	}

	if ctx.Response() != res {
		t.Error("Response() does not match")
	}

	if ctx.Container() != container {
		t.Error("Container() does not match")
	}
}

func TestDefaultContext_Request(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.Request() != req {
		t.Error("Request() returned unexpected value")
	}
}

func TestDefaultContext_Response(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	if ctx.Response() != res {
		t.Error("Response() returned unexpected value")
	}
}

func TestDefaultContext_Param(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer()).(*DefaultContext)

	ctx.SetParam("id", "123")

	param := ctx.Param("id")
	if param != "123" {
		t.Errorf("Param() = %v, want '123'", param)
	}

	nonExistent := ctx.Param("nonexistent")
	if nonExistent != "" {
		t.Errorf("Param() for non-existent key = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?page=1&limit=10", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	page := ctx.Query("page")
	if page != "1" {
		t.Errorf("Query('page') = %v, want '1'", page)
	}

	limit := ctx.Query("limit")
	if limit != "10" {
		t.Errorf("Query('limit') = %v, want '10'", limit)
	}

	nonExistent := ctx.Query("nonexistent")
	if nonExistent != "" {
		t.Errorf("Query('nonexistent') = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_Body(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid json",
			body:    `{"name":"test","age":25}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			body:    `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			ctx := NewContext(context.Background(), req, res, NewContainer())

			var result map[string]interface{}
			err := ctx.Body(&result)

			if (err != nil) != tt.wantErr {
				t.Errorf("Body() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultContext_Header(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	auth := ctx.Header("Authorization")
	if auth != "Bearer token123" {
		t.Errorf("Header('Authorization') = %v, want 'Bearer token123'", auth)
	}

	contentType := ctx.Header("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Header('Content-Type') = %v, want 'application/json'", contentType)
	}

	nonExistent := ctx.Header("NonExistent")
	if nonExistent != "" {
		t.Errorf("Header('NonExistent') = %v, want ''", nonExistent)
	}
}

func TestDefaultContext_JSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	data := map[string]interface{}{
		"message": "success",
		"code":    200,
	}

	err := ctx.JSON(200, data)
	if err != nil {
		t.Errorf("JSON() error = %v", err)
	}

	if res.Code != 200 {
		t.Errorf("Response code = %v, want 200", res.Code)
	}

	contentType := res.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %v, want 'application/json'", contentType)
	}

	body := res.Body.String()
	if !strings.Contains(body, "success") {
		t.Error("Response body does not contain expected data")
	}
}

func TestDefaultContext_String(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := ctx.String(200, "Hello, World!")
	if err != nil {
		t.Errorf("String() error = %v", err)
	}

	if res.Code != 200 {
		t.Errorf("Response code = %v, want 200", res.Code)
	}

	contentType := res.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Content-Type = %v, want 'text/plain'", contentType)
	}

	body := res.Body.String()
	if body != "Hello, World!" {
		t.Errorf("Response body = %v, want 'Hello, World!'", body)
	}
}

func TestDefaultContext_Status(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.Status(204)

	if res.Code != 204 {
		t.Errorf("Response code = %v, want 204", res.Code)
	}
}

func TestDefaultContext_SetHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	ctx.SetHeader("X-Custom-Header", "test-value")

	header := res.Header().Get("X-Custom-Header")
	if header != "test-value" {
		t.Errorf("Header = %v, want 'test-value'", header)
	}
}

func TestDefaultContext_GetSet(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	// Test Set and Get
	ctx.Set("user", "john")
	ctx.Set("role", "admin")

	user := ctx.Get("user")
	if user != "john" {
		t.Errorf("Get('user') = %v, want 'john'", user)
	}

	role := ctx.Get("role")
	if role != "admin" {
		t.Errorf("Get('role') = %v, want 'admin'", role)
	}

	// Test non-existent key
	nonExistent := ctx.Get("nonexistent")
	if nonExistent != nil {
		t.Errorf("Get('nonexistent') = %v, want nil", nonExistent)
	}
}

func TestDefaultContext_Container(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	container := NewContainer()
	ctx := NewContext(context.Background(), req, res, container)

	if ctx.Container() != container {
		t.Error("Container() does not match the provided container")
	}
}

