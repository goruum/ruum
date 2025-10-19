package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewBaseController(t *testing.T) {
	ctrl := NewBaseController("/api")

	if ctrl == nil {
		t.Fatal("NewBaseController() returned nil")
	}

	if ctrl.prefix != "/api" {
		t.Errorf("prefix = %v, want '/api'", ctrl.prefix)
	}
}

func TestBaseController_SetPrefix(t *testing.T) {
	ctrl := NewBaseController("/api")
	ctrl.SetPrefix("/v1")

	if ctrl.prefix != "/v1" {
		t.Errorf("prefix = %v, want '/v1'", ctrl.prefix)
	}
}

type mockGuard struct{}

func (g *mockGuard) CanActivate(ctx core.Context) (bool, error) {
	return true, nil
}

func TestBaseController_UseGuards(t *testing.T) {
	ctrl := NewBaseController("/api")

	guard := &mockGuard{}
	ctrl.UseGuards(guard)

	if len(ctrl.guards) != 1 {
		t.Errorf("guards length = %d, want 1", len(ctrl.guards))
	}
}

type mockInterceptor struct{}

func (i *mockInterceptor) Intercept(ctx core.Context, next core.HandlerFunc) error {
	return next(ctx)
}

func TestBaseController_UseInterceptors(t *testing.T) {
	ctrl := NewBaseController("/api")

	interceptor := &mockInterceptor{}
	ctrl.UseInterceptors(interceptor)

	if len(ctrl.interceptors) != 1 {
		t.Errorf("interceptors length = %d, want 1", len(ctrl.interceptors))
	}
}

func TestBaseController_Get(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Get("/users", handler)

	if len(ctrl.routes) != 1 {
		t.Errorf("routes length = %d, want 1", len(ctrl.routes))
	}

	if ctrl.routes[0].Method != "GET" {
		t.Errorf("Method = %v, want 'GET'", ctrl.routes[0].Method)
	}

	if ctrl.routes[0].Path != "/users" {
		t.Errorf("Path = %v, want '/users'", ctrl.routes[0].Path)
	}
}

func TestBaseController_Post(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Post("/users", handler)

	if len(ctrl.routes) != 1 {
		t.Errorf("routes length = %d, want 1", len(ctrl.routes))
	}

	if ctrl.routes[0].Method != "POST" {
		t.Errorf("Method = %v, want 'POST'", ctrl.routes[0].Method)
	}
}

func TestBaseController_Put(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Put("/users/{id}", handler)

	if len(ctrl.routes) != 1 {
		t.Errorf("routes length = %d, want 1", len(ctrl.routes))
	}

	if ctrl.routes[0].Method != "PUT" {
		t.Errorf("Method = %v, want 'PUT'", ctrl.routes[0].Method)
	}
}

func TestBaseController_Delete(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Delete("/users/{id}", handler)

	if len(ctrl.routes) != 1 {
		t.Errorf("routes length = %d, want 1", len(ctrl.routes))
	}

	if ctrl.routes[0].Method != "DELETE" {
		t.Errorf("Method = %v, want 'DELETE'", ctrl.routes[0].Method)
	}
}

func TestBaseController_Patch(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Patch("/users/{id}", handler)

	if len(ctrl.routes) != 1 {
		t.Errorf("routes length = %d, want 1", len(ctrl.routes))
	}

	if ctrl.routes[0].Method != "PATCH" {
		t.Errorf("Method = %v, want 'PATCH'", ctrl.routes[0].Method)
	}
}

func TestBaseController_RegisterRoutes(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Get("/users", handler)
	ctrl.Post("/users", handler)

	router := core.NewRouter()
	err := ctrl.RegisterRoutes(router)

	if err != nil {
		t.Errorf("RegisterRoutes() error = %v", err)
	}
}

func TestWithGuards(t *testing.T) {
	guard := &mockGuard{}
	route := RouteDefinition{}

	opt := WithGuards(guard)
	opt(&route)

	if len(route.Guards) != 1 {
		t.Errorf("Guards length = %d, want 1", len(route.Guards))
	}
}

type mockPipe struct{}

func (p *mockPipe) Transform(value interface{}, metadata *core.ArgumentMetadata) (interface{}, error) {
	return value, nil
}

func TestWithInterceptors(t *testing.T) {
	interceptor := &mockInterceptor{}
	route := RouteDefinition{}

	opt := WithInterceptors(interceptor)
	opt(&route)

	if len(route.Interceptors) != 1 {
		t.Errorf("Interceptors length = %d, want 1", len(route.Interceptors))
	}
}

func TestWithPipes(t *testing.T) {
	pipe := &mockPipe{}
	route := RouteDefinition{}

	opt := WithPipes(pipe)
	opt(&route)

	if len(route.Pipes) != 1 {
		t.Errorf("Pipes length = %d, want 1", len(route.Pipes))
	}
}

func TestBaseController_RouteWithOptions(t *testing.T) {
	ctrl := NewBaseController("/api")
	guard := &mockGuard{}

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Get("/protected", handler, WithGuards(guard))

	if len(ctrl.routes) != 1 {
		t.Fatalf("routes length = %d, want 1", len(ctrl.routes))
	}

	if len(ctrl.routes[0].Guards) != 1 {
		t.Errorf("Route guards length = %d, want 1", len(ctrl.routes[0].Guards))
	}
}

func TestBaseController_MultipleRoutes(t *testing.T) {
	ctrl := NewBaseController("/api")

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Get("/users", handler)
	ctrl.Post("/users", handler)
	ctrl.Put("/users/{id}", handler)
	ctrl.Delete("/users/{id}", handler)
	ctrl.Patch("/users/{id}", handler)

	if len(ctrl.routes) != 5 {
		t.Errorf("routes length = %d, want 5", len(ctrl.routes))
	}
}

func TestBaseController_GuardsExecution(t *testing.T) {
	ctrl := NewBaseController("/api")
	guard := &mockGuard{}
	ctrl.UseGuards(guard)

	handler := func(ctx core.Context) error {
		return ctx.String(200, "OK")
	}

	ctrl.Get("/test", handler)

	router := core.NewRouter()
	err := ctrl.RegisterRoutes(router)

	if err != nil {
		t.Errorf("RegisterRoutes() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	res := httptest.NewRecorder()
	ctx := core.NewContext(context.Background(), req, res, core.NewContainer())

	_ = router.ServeHTTP(ctx)
}
