package http

import (
	"github.com/goruum/ruum/core"
)

// BaseController provides base functionality for controllers
type BaseController struct {
	prefix       string
	routes       []RouteDefinition
	guards       []core.Guard
	interceptors []core.Interceptor
}

// RouteDefinition defines a route
type RouteDefinition struct {
	Method       string
	Path         string
	Handler      core.HandlerFunc
	Guards       []core.Guard
	Interceptors []core.Interceptor
	Pipes        []core.Pipe
}

// NewBaseController creates a new base controller
func NewBaseController(prefix string) *BaseController {
	return &BaseController{
		prefix:       prefix,
		routes:       make([]RouteDefinition, 0),
		guards:       make([]core.Guard, 0),
		interceptors: make([]core.Interceptor, 0),
	}
}

// SetPrefix sets the controller prefix
func (c *BaseController) SetPrefix(prefix string) {
	c.prefix = prefix
}

// UseGuards adds guards to the controller
func (c *BaseController) UseGuards(guards ...core.Guard) {
	c.guards = append(c.guards, guards...)
}

// UseInterceptors adds interceptors to the controller
func (c *BaseController) UseInterceptors(interceptors ...core.Interceptor) {
	c.interceptors = append(c.interceptors, interceptors...)
}

// Get adds a GET route
func (c *BaseController) Get(path string, handler core.HandlerFunc, opts ...RouteOption) {
	c.addRoute("GET", path, handler, opts...)
}

// Post adds a POST route
func (c *BaseController) Post(path string, handler core.HandlerFunc, opts ...RouteOption) {
	c.addRoute("POST", path, handler, opts...)
}

// Put adds a PUT route
func (c *BaseController) Put(path string, handler core.HandlerFunc, opts ...RouteOption) {
	c.addRoute("PUT", path, handler, opts...)
}

// Delete adds a DELETE route
func (c *BaseController) Delete(path string, handler core.HandlerFunc, opts ...RouteOption) {
	c.addRoute("DELETE", path, handler, opts...)
}

// Patch adds a PATCH route
func (c *BaseController) Patch(path string, handler core.HandlerFunc, opts ...RouteOption) {
	c.addRoute("PATCH", path, handler, opts...)
}

func (c *BaseController) addRoute(method, path string, handler core.HandlerFunc, opts ...RouteOption) {
	route := RouteDefinition{
		Method:       method,
		Path:         path,
		Handler:      handler,
		Guards:       make([]core.Guard, 0),
		Interceptors: make([]core.Interceptor, 0),
		Pipes:        make([]core.Pipe, 0),
	}

	for _, opt := range opts {
		opt(&route)
	}

	c.routes = append(c.routes, route)
}

// RegisterRoutes registers all routes with the router
func (c *BaseController) RegisterRoutes(router core.Router) error {
	group := router.Group(c.prefix)

	for _, route := range c.routes {
		// Wrap handler with route-specific middleware
		handler := route.Handler

		// Apply route pipes
		handler = c.applyPipes(handler, route.Pipes)

		// Apply route interceptors
		handler = c.applyInterceptors(handler, route.Interceptors)

		// Apply controller interceptors
		handler = c.applyInterceptors(handler, c.interceptors)

		// Apply route guards
		handler = c.applyGuards(handler, route.Guards)

		// Apply controller guards
		handler = c.applyGuards(handler, c.guards)

		// Register route
		switch route.Method {
		case "GET":
			group.Get(route.Path, handler)
		case "POST":
			group.Post(route.Path, handler)
		case "PUT":
			group.Put(route.Path, handler)
		case "DELETE":
			group.Delete(route.Path, handler)
		case "PATCH":
			group.Patch(route.Path, handler)
		}
	}

	return nil
}

func (c *BaseController) applyGuards(handler core.HandlerFunc, guards []core.Guard) core.HandlerFunc {
	return func(ctx core.Context) error {
		for _, guard := range guards {
			allowed, err := guard.CanActivate(ctx)
			if err != nil {
				return err
			}
			if !allowed {
				return core.ForbiddenException("Access denied")
			}
		}
		return handler(ctx)
	}
}

func (c *BaseController) applyInterceptors(handler core.HandlerFunc, interceptors []core.Interceptor) core.HandlerFunc {
	result := handler
	for i := len(interceptors) - 1; i >= 0; i-- {
		interceptor := interceptors[i]
		currentHandler := result
		result = func(ctx core.Context) error {
			return interceptor.Intercept(ctx, currentHandler)
		}
	}
	return result
}

func (c *BaseController) applyPipes(handler core.HandlerFunc, pipes []core.Pipe) core.HandlerFunc {
	return func(ctx core.Context) error {
		// Pipes are typically applied at parameter level
		// This is a placeholder for route-level pipe application
		return handler(ctx)
	}
}

// RouteOption configures a route
type RouteOption func(*RouteDefinition)

// WithGuards adds guards to a route
func WithGuards(guards ...core.Guard) RouteOption {
	return func(r *RouteDefinition) {
		r.Guards = append(r.Guards, guards...)
	}
}

// WithInterceptors adds interceptors to a route
func WithInterceptors(interceptors ...core.Interceptor) RouteOption {
	return func(r *RouteDefinition) {
		r.Interceptors = append(r.Interceptors, interceptors...)
	}
}

// WithPipes adds pipes to a route
func WithPipes(pipes ...core.Pipe) RouteOption {
	return func(r *RouteDefinition) {
		r.Pipes = append(r.Pipes, pipes...)
	}
}
