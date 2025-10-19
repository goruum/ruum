package guards

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewAuthGuard(t *testing.T) {
	guard := NewAuthGuard()
	if guard == nil {
		t.Fatal("NewAuthGuard() returned nil")
	}

	if guard.headerKey != "Authorization" {
		t.Errorf("headerKey = %v, want 'Authorization'", guard.headerKey)
	}

	if guard.tokenType != "Bearer" {
		t.Errorf("tokenType = %v, want 'Bearer'", guard.tokenType)
	}
}

func TestAuthGuard_WithHeaderKey(t *testing.T) {
	guard := NewAuthGuard().WithHeaderKey("X-Auth-Token")

	if guard.headerKey != "X-Auth-Token" {
		t.Errorf("headerKey = %v, want 'X-Auth-Token'", guard.headerKey)
	}
}

func TestAuthGuard_WithTokenType(t *testing.T) {
	guard := NewAuthGuard().WithTokenType("Token")

	if guard.tokenType != "Token" {
		t.Errorf("tokenType = %v, want 'Token'", guard.tokenType)
	}
}

func TestAuthGuard_CanActivate_Success(t *testing.T) {
	guard := NewAuthGuard()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err != nil {
		t.Errorf("CanActivate() error = %v", err)
	}

	if !allowed {
		t.Error("CanActivate() returned false, want true")
	}

	// Check if token was stored in context
	token := ctx.Get("auth_token")
	if token != "valid-token" {
		t.Errorf("Token in context = %v, want 'valid-token'", token)
	}
}

func TestAuthGuard_CanActivate_MissingHeader(t *testing.T) {
	guard := NewAuthGuard()

	req := httptest.NewRequest("GET", "/protected", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for missing header, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}

	httpErr, ok := err.(*core.HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != 401 {
		t.Errorf("StatusCode = %v, want 401", httpErr.StatusCode)
	}
}

func TestAuthGuard_CanActivate_InvalidFormat_NoSpace(t *testing.T) {
	guard := NewAuthGuard()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}
}

func TestAuthGuard_CanActivate_InvalidFormat_WrongType(t *testing.T) {
	guard := NewAuthGuard()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Basic token123")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for wrong token type, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}
}

func TestAuthGuard_CanActivate_EmptyToken(t *testing.T) {
	guard := NewAuthGuard()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}
}

func TestAuthGuard_CustomHeaderKey(t *testing.T) {
	guard := NewAuthGuard().WithHeaderKey("X-API-Key")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("X-API-Key", "Bearer mytoken")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err != nil {
		t.Errorf("CanActivate() error = %v", err)
	}

	if !allowed {
		t.Error("CanActivate() returned false, want true")
	}
}

func TestAuthGuard_CustomTokenType(t *testing.T) {
	guard := NewAuthGuard().WithTokenType("Token")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Token mytoken")
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err != nil {
		t.Errorf("CanActivate() error = %v", err)
	}

	if !allowed {
		t.Error("CanActivate() returned false, want true")
	}
}

func TestNewRolesGuard(t *testing.T) {
	guard := NewRolesGuard("admin", "moderator")

	if guard == nil {
		t.Fatal("NewRolesGuard() returned nil")
	}

	if len(guard.requiredRoles) != 2 {
		t.Errorf("requiredRoles length = %d, want 2", len(guard.requiredRoles))
	}
}

func TestRolesGuard_CanActivate_Success(t *testing.T) {
	guard := NewRolesGuard("admin", "moderator")

	req := httptest.NewRequest("GET", "/admin", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	// Set user roles in context
	ctx.Set("user_roles", []string{"admin", "user"})

	allowed, err := guard.CanActivate(ctx)
	if err != nil {
		t.Errorf("CanActivate() error = %v", err)
	}

	if !allowed {
		t.Error("CanActivate() returned false, want true")
	}
}

func TestRolesGuard_CanActivate_NoRolesInContext(t *testing.T) {
	guard := NewRolesGuard("admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for missing roles, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}

	httpErr, ok := err.(*core.HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != 403 {
		t.Errorf("StatusCode = %v, want 403", httpErr.StatusCode)
	}
}

func TestRolesGuard_CanActivate_InvalidRolesType(t *testing.T) {
	guard := NewRolesGuard("admin")

	req := httptest.NewRequest("GET", "/admin", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	// Set invalid type for user_roles
	ctx.Set("user_roles", "not-a-slice")

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for invalid roles type, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}
}

func TestRolesGuard_CanActivate_InsufficientPermissions(t *testing.T) {
	guard := NewRolesGuard("admin", "moderator")

	req := httptest.NewRequest("GET", "/admin", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	// User has roles but not the required ones
	ctx.Set("user_roles", []string{"user", "viewer"})

	allowed, err := guard.CanActivate(ctx)
	if err == nil {
		t.Error("Expected error for insufficient permissions, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}

	httpErr, ok := err.(*core.HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != 403 {
		t.Errorf("StatusCode = %v, want 403", httpErr.StatusCode)
	}
}

func TestRolesGuard_CanActivate_MultipleRequiredRoles(t *testing.T) {
	guard := NewRolesGuard("admin", "moderator", "superuser")

	req := httptest.NewRequest("GET", "/admin", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	// User has one of the required roles
	ctx.Set("user_roles", []string{"moderator", "user"})

	allowed, err := guard.CanActivate(ctx)
	if err != nil {
		t.Errorf("CanActivate() error = %v", err)
	}

	if !allowed {
		t.Error("CanActivate() returned false, want true")
	}
}

func TestRolesGuard_EmptyRequiredRoles(t *testing.T) {
	guard := NewRolesGuard()

	req := httptest.NewRequest("GET", "/public", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	ctx.Set("user_roles", []string{"user"})

	allowed, err := guard.CanActivate(ctx)
	// With no required roles, should not allow
	if err == nil {
		t.Error("Expected error for empty required roles, got nil")
	}

	if allowed {
		t.Error("CanActivate() returned true, want false")
	}
}
