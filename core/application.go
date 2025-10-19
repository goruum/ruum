// Package core provides the main application framework components including
// the application factory, lifecycle management, and HTTP server setup.
package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Application is the main application interface
type Application interface {
	Listen(addr string) error
	Close() error
	Use(middleware ...MiddlewareFunc)
	UseGlobalGuards(guards ...Guard)
	UseGlobalInterceptors(interceptors ...Interceptor)
	UseGlobalPipes(pipes ...Pipe)
	UseGlobalFilters(filters ...ExceptionFilter)
	Get(path string, handler HandlerFunc)
	Post(path string, handler HandlerFunc)
	Put(path string, handler HandlerFunc)
	Delete(path string, handler HandlerFunc)
	Patch(path string, handler HandlerFunc)
	GetContainer() Container
	GetLogger() Logger
	SetLogger(logger Logger)
}

// ApplicationConfig contains application configuration
type ApplicationConfig struct {
	GlobalPrefix        string
	CorsEnabled         bool
	CorsOrigins         []string
	Logger              Logger
	ShutdownTimeout     time.Duration
	EnableShutdownHooks bool
}

// DefaultApplication implements Application interface
type DefaultApplication struct {
	container          Container
	router             Router
	server             *http.Server
	logger             Logger
	config             ApplicationConfig
	middleware         []MiddlewareFunc
	globalGuards       []Guard
	globalInterceptors []Interceptor
	globalPipes        []Pipe
	globalFilters      []ExceptionFilter
	shutdownHooks      []func()
}

// ApplicationFactory creates new applications
type ApplicationFactory struct{}

// NewApplicationFactory creates a new application factory
func NewApplicationFactory() *ApplicationFactory {
	return &ApplicationFactory{}
}

// Create creates a new application from a root module
func (f *ApplicationFactory) Create(rootModule Module, config ApplicationConfig) (Application, error) {
	container := NewContainer()

	// Configure the root module
	if err := rootModule.Configure(container); err != nil {
		return nil, fmt.Errorf("failed to configure root module: %w", err)
	}

	// Create router
	router := NewRouter()

	// Register controllers
	if err := f.registerControllers(rootModule, container, router); err != nil {
		return nil, fmt.Errorf("failed to register controllers: %w", err)
	}

	// Set default logger if not provided
	if config.Logger == nil {
		config.Logger = NewDefaultLogger()
	}

	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	app := &DefaultApplication{
		container:          container,
		router:             router,
		logger:             config.Logger,
		config:             config,
		middleware:         make([]MiddlewareFunc, 0),
		globalGuards:       make([]Guard, 0),
		globalInterceptors: make([]Interceptor, 0),
		globalPipes:        make([]Pipe, 0),
		globalFilters:      make([]ExceptionFilter, 0),
		shutdownHooks:      make([]func(), 0),
	}

	// Call lifecycle hooks
	if err := f.callLifecycleHooks(rootModule, container); err != nil {
		return nil, fmt.Errorf("failed to call lifecycle hooks: %w", err)
	}

	return app, nil
}

func (f *ApplicationFactory) registerControllers(module Module, _ Container, router Router) error {
	// Register controllers from this module
	controllers := module.GetControllers()
	for _, ctrl := range controllers {
		if controller, ok := ctrl.(Controller); ok {
			if err := controller.RegisterRoutes(router); err != nil {
				return fmt.Errorf("failed to register controller routes: %w", err)
			}
		}
	}

	// Register controllers from imported modules
	for _, importedModule := range module.GetImports() {
		var nilContainer Container
		if err := f.registerControllers(importedModule, nilContainer, router); err != nil {
			return err
		}
	}

	return nil
}

func (f *ApplicationFactory) callLifecycleHooks(_ Module, cnt Container) error {
	// Call OnModuleInit for providers
	for name := range cnt.GetAll() {
		instance, err := cnt.Resolve(name)
		if err != nil {
			continue
		}

		if hook, ok := instance.(OnModuleInit); ok {
			if err := hook.OnModuleInit(); err != nil {
				return fmt.Errorf("OnModuleInit failed for %s: %w", name, err)
			}
		}
	}

	return nil
}

// Listen starts the HTTP server on the specified address.
func (a *DefaultApplication) Listen(addr string) error {
	handler := a.buildHandler()

	a.server = &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Setup graceful shutdown
	if a.config.EnableShutdownHooks {
		go a.setupGracefulShutdown()
	}

	a.logger.Info(fmt.Sprintf("🚀 Application is running on: http://%s", addr), nil)

	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Close gracefully shuts down the application.
func (a *DefaultApplication) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.config.ShutdownTimeout)
	defer cancel()

	// Call shutdown hooks
	for _, hook := range a.shutdownHooks {
		hook()
	}

	// Call OnApplicationShutdown for providers
	for name := range a.container.GetAll() {
		instance, err := a.container.Resolve(name)
		if err != nil {
			continue
		}

		if hook, ok := instance.(OnApplicationShutdown); ok {
			if err := hook.OnApplicationShutdown(); err != nil {
				a.logger.Error(fmt.Sprintf("OnApplicationShutdown failed for %s", name), map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}

	if a.server != nil {
		return a.server.Shutdown(ctx)
	}

	return nil
}

func (a *DefaultApplication) setupGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	a.logger.Info("Shutting down application...", nil)

	if err := a.Close(); err != nil {
		a.logger.Error("Error during shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (a *DefaultApplication) buildHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(context.Background(), r, w, a.container)

		// Build handler chain with middleware
		handler := a.router.ServeHTTP

		// Apply global filters
		handler = a.applyFilters(handler)

		// Apply global pipes
		handler = a.applyPipes(handler)

		// Apply global interceptors
		handler = a.applyInterceptors(handler)

		// Apply global guards
		handler = a.applyGuards(handler)

		// Apply middleware
		for i := len(a.middleware) - 1; i >= 0; i-- {
			handler = a.middleware[i](handler)
		}

		// Execute handler
		if err := handler(ctx); err != nil {
			a.handleError(err, ctx)
		}
	})
}

func (a *DefaultApplication) applyGuards(next HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		for _, guard := range a.globalGuards {
			allowed, err := guard.CanActivate(ctx)
			if err != nil {
				return err
			}
			if !allowed {
				return NewHTTPException(http.StatusForbidden, "Forbidden")
			}
		}
		return next(ctx)
	}
}

func (a *DefaultApplication) applyInterceptors(next HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		handler := next
		for i := len(a.globalInterceptors) - 1; i >= 0; i-- {
			interceptor := a.globalInterceptors[i]
			currentHandler := handler
			handler = func(ctx Context) error {
				return interceptor.Intercept(ctx, currentHandler)
			}
		}
		return handler(ctx)
	}
}

func (a *DefaultApplication) applyPipes(next HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		// Pipes are applied at the parameter level
		// This is a placeholder for global pipe application
		return next(ctx)
	}
}

func (a *DefaultApplication) applyFilters(next HandlerFunc) HandlerFunc {
	return func(ctx Context) error {
		err := next(ctx)
		if err != nil {
			for _, filter := range a.globalFilters {
				if filterErr := filter.Catch(err, ctx); filterErr == nil {
					return nil
				}
			}
			return err
		}
		return nil
	}
}

func (a *DefaultApplication) handleError(err error, ctx Context) {
	httpErr, ok := err.(*HTTPException)
	if !ok {
		httpErr = NewHTTPException(http.StatusInternalServerError, err.Error())
	}

	_ = ctx.JSON(httpErr.StatusCode, map[string]interface{}{
		"statusCode": httpErr.StatusCode,
		"message":    httpErr.Message,
		"error":      http.StatusText(httpErr.StatusCode),
	})
}

// Use adds global middleware to the application.
func (a *DefaultApplication) Use(middleware ...MiddlewareFunc) {
	a.middleware = append(a.middleware, middleware...)
}

// UseGlobalGuards adds global guards to all routes.
func (a *DefaultApplication) UseGlobalGuards(guards ...Guard) {
	a.globalGuards = append(a.globalGuards, guards...)
}

// UseGlobalInterceptors adds global interceptors to all routes.
func (a *DefaultApplication) UseGlobalInterceptors(interceptors ...Interceptor) {
	a.globalInterceptors = append(a.globalInterceptors, interceptors...)
}

// UseGlobalPipes adds global pipes for validation and transformation.
func (a *DefaultApplication) UseGlobalPipes(pipes ...Pipe) {
	a.globalPipes = append(a.globalPipes, pipes...)
}

// UseGlobalFilters adds global exception filters.
func (a *DefaultApplication) UseGlobalFilters(filters ...ExceptionFilter) {
	a.globalFilters = append(a.globalFilters, filters...)
}

// Get registers a GET route.
func (a *DefaultApplication) Get(path string, handler HandlerFunc) {
	a.router.Get(path, handler)
}

// Post registers a POST route.
func (a *DefaultApplication) Post(path string, handler HandlerFunc) {
	a.router.Post(path, handler)
}

// Put registers a PUT route.
func (a *DefaultApplication) Put(path string, handler HandlerFunc) {
	a.router.Put(path, handler)
}

// Delete registers a DELETE route.
func (a *DefaultApplication) Delete(path string, handler HandlerFunc) {
	a.router.Delete(path, handler)
}

// Patch registers a PATCH route.
func (a *DefaultApplication) Patch(path string, handler HandlerFunc) {
	a.router.Patch(path, handler)
}

// GetContainer returns the dependency injection container.
func (a *DefaultApplication) GetContainer() Container {
	return a.container
}

// GetLogger returns the application logger.
func (a *DefaultApplication) GetLogger() Logger {
	return a.logger
}

// SetLogger sets the application logger.
func (a *DefaultApplication) SetLogger(logger Logger) {
	a.logger = logger
}
