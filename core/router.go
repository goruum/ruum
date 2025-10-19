package core

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Router handles HTTP routing
type Router interface {
	Get(path string, handler HandlerFunc)
	Post(path string, handler HandlerFunc)
	Put(path string, handler HandlerFunc)
	Delete(path string, handler HandlerFunc)
	Patch(path string, handler HandlerFunc)
	Options(path string, handler HandlerFunc)
	Head(path string, handler HandlerFunc)
	Group(prefix string) Router
	ServeHTTP(ctx Context) error
}

// Controller is the base interface for controllers
type Controller interface {
	RegisterRoutes(router Router) error
}

// DefaultRouter implements Router using gorilla/mux
type DefaultRouter struct {
	mux      *mux.Router
	prefix   string
	handlers map[string]HandlerFunc
}

// NewRouter creates a new router
func NewRouter() Router {
	return &DefaultRouter{
		mux:      mux.NewRouter(),
		prefix:   "",
		handlers: make(map[string]HandlerFunc),
	}
}

// Get registers a GET route.
func (r *DefaultRouter) Get(path string, handler HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

// Post registers a POST route.
func (r *DefaultRouter) Post(path string, handler HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

// Put registers a PUT route.
func (r *DefaultRouter) Put(path string, handler HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

// Delete registers a DELETE route.
func (r *DefaultRouter) Delete(path string, handler HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

// Patch registers a PATCH route.
func (r *DefaultRouter) Patch(path string, handler HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

// Options registers an OPTIONS route.
func (r *DefaultRouter) Options(path string, handler HandlerFunc) {
	r.handle(http.MethodOptions, path, handler)
}

// Head registers a HEAD route.
func (r *DefaultRouter) Head(path string, handler HandlerFunc) {
	r.handle(http.MethodHead, path, handler)
}

func (r *DefaultRouter) handle(method, path string, handler HandlerFunc) {
	fullPath := r.prefix + path
	key := method + ":" + fullPath
	r.handlers[key] = handler

	r.mux.HandleFunc(fullPath, func(_ http.ResponseWriter, _ *http.Request) {
		// Handler will be called by ServeHTTP
	}).Methods(method)
}

// Group creates a new router group with the given prefix.
func (r *DefaultRouter) Group(prefix string) Router {
	return &DefaultRouter{
		mux:      r.mux,
		prefix:   r.prefix + prefix,
		handlers: r.handlers, // Share handlers map
	}
}

func (r *DefaultRouter) ServeHTTP(ctx Context) error {
	// Extract route params
	vars := mux.Vars(ctx.Request())
	if defaultCtx, ok := ctx.(*DefaultContext); ok {
		for key, value := range vars {
			defaultCtx.SetParam(key, value)
		}
	}

	// Find and execute the handler
	key := ctx.Request().Method + ":" + ctx.Request().URL.Path
	if handler, exists := r.handlers[key]; exists {
		return handler(ctx)
	}

	return NewHTTPException(http.StatusNotFound, "Route not found")
}

// GetMux returns the underlying mux router for advanced usage
func (r *DefaultRouter) GetMux() *mux.Router {
	return r.mux
}

// RouteMetadata contains route configuration
type RouteMetadata struct {
	Method       string
	Path         string
	Handler      HandlerFunc
	Guards       []Guard
	Interceptors []Interceptor
	Pipes        []Pipe
}

// ControllerMetadata contains controller configuration
type ControllerMetadata struct {
	Prefix string
	Routes []RouteMetadata
}
