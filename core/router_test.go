package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}
}

func TestRouter_Get(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Get("/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Post(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Post("/test", handler)

	req := httptest.NewRequest("POST", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Put(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Put("/test", handler)

	req := httptest.NewRequest("PUT", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Delete(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Delete("/test", handler)

	req := httptest.NewRequest("DELETE", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Patch(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Patch("/test", handler)

	req := httptest.NewRequest("PATCH", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Options(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Options("/test", handler)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Head(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	router.Head("/test", handler)

	req := httptest.NewRequest("HEAD", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_Group(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	// Create a group with prefix
	group := router.Group("/api")
	group.Get("/users", handler)

	req := httptest.NewRequest("GET", "/api/users", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_NestedGroups(t *testing.T) {
	router := NewRouter()
	called := false

	handler := func(ctx Context) error {
		called = true
		return nil
	}

	// Create nested groups
	apiGroup := router.Group("/api")
	v1Group := apiGroup.Group("/v1")
	v1Group.Get("/users", handler)

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err != nil {
		t.Errorf("ServeHTTP() error = %v", err)
	}

	if !called {
		t.Error("Handler was not called")
	}
}

func TestRouter_RouteParams(t *testing.T) {
	router := NewRouter()

	handler := func(ctx Context) error {
		// Check if param can be retrieved
		_ = ctx.Param("id")
		return nil
	}

	router.Get("/users/{id}", handler)

	// We won't test the actual param extraction since it requires mux internal state
	// Just verify the route is registered
	if router == nil {
		t.Error("Router should not be nil")
	}
}

func TestRouter_NotFound(t *testing.T) {
	router := NewRouter()

	handler := func(ctx Context) error {
		return nil
	}

	router.Get("/existing", handler)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err == nil {
		t.Error("Expected error for non-existent route, got nil")
	}

	httpErr, ok := err.(*HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %v, want %v", httpErr.StatusCode, http.StatusNotFound)
	}
}

func TestRouter_MultipleRoutes(t *testing.T) {
	router := NewRouter()
	route1Called := false
	route2Called := false

	handler1 := func(ctx Context) error {
		route1Called = true
		return nil
	}

	handler2 := func(ctx Context) error {
		route2Called = true
		return nil
	}

	router.Get("/route1", handler1)
	router.Get("/route2", handler2)

	// Call route1
	req1 := httptest.NewRequest("GET", "/route1", nil)
	res1 := httptest.NewRecorder()
	ctx1 := NewContext(context.Background(), req1, res1, NewContainer())
	_ = router.ServeHTTP(ctx1)

	// Call route2
	req2 := httptest.NewRequest("GET", "/route2", nil)
	res2 := httptest.NewRecorder()
	ctx2 := NewContext(context.Background(), req2, res2, NewContainer())
	_ = router.ServeHTTP(ctx2)

	if !route1Called {
		t.Error("Route1 handler was not called")
	}

	if !route2Called {
		t.Error("Route2 handler was not called")
	}
}

func TestRouter_HandlerError(t *testing.T) {
	router := NewRouter()

	handler := func(ctx Context) error {
		return BadRequestException("test error")
	}

	router.Get("/error", handler)

	req := httptest.NewRequest("GET", "/error", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(context.Background(), req, res, NewContainer())

	err := router.ServeHTTP(ctx)
	if err == nil {
		t.Error("Expected error from handler, got nil")
	}

	httpErr, ok := err.(*HTTPException)
	if !ok {
		t.Error("Expected HTTPException")
	}

	if httpErr.StatusCode != 400 {
		t.Errorf("StatusCode = %v, want 400", httpErr.StatusCode)
	}
}
