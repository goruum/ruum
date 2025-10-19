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

func (r *DefaultRouter) Get(path string, handler HandlerFunc) {
	r.handle(http.MethodGet, path, handler)
}

func (r *DefaultRouter) Post(path string, handler HandlerFunc) {
	r.handle(http.MethodPost, path, handler)
}

func (r *DefaultRouter) Put(path string, handler HandlerFunc) {
	r.handle(http.MethodPut, path, handler)
}

func (r *DefaultRouter) Delete(path string, handler HandlerFunc) {
	r.handle(http.MethodDelete, path, handler)
}

func (r *DefaultRouter) Patch(path string, handler HandlerFunc) {
	r.handle(http.MethodPatch, path, handler)
}

func (r *DefaultRouter) Options(path string, handler HandlerFunc) {
	r.handle(http.MethodOptions, path, handler)
}

func (r *DefaultRouter) Head(path string, handler HandlerFunc) {
	r.handle(http.MethodHead, path, handler)
}

func (r *DefaultRouter) handle(method, path string, handler HandlerFunc) {
	fullPath := r.prefix + path
	key := method + ":" + fullPath
	r.handlers[key] = handler
	
	r.mux.HandleFunc(fullPath, func(w http.ResponseWriter, req *http.Request) {
		// Handler will be called by ServeHTTP
	}).Methods(method)
}

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
	
	return NewHttpException(http.StatusNotFound, "Route not found")
}

// GetMux returns the underlying mux router for advanced usage
func (r *DefaultRouter) GetMux() *mux.Router {
	return r.mux
}

// RouteMetadata contains route configuration
type RouteMetadata struct {
	Method      string
	Path        string
	Handler     HandlerFunc
	Guards      []Guard
	Interceptors []Interceptor
	Pipes       []Pipe
}

// ControllerMetadata contains controller configuration
type ControllerMetadata struct {
	Prefix string
	Routes []RouteMetadata
}

