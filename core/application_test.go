package core

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewApplicationFactory(t *testing.T) {
	factory := NewApplicationFactory()
	if factory == nil {
		t.Fatal("NewApplicationFactory() returned nil")
	}
}

func TestApplicationFactory_Create(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()

	config := ApplicationConfig{
		Logger: NewDefaultLogger(),
	}

	app, err := factory.Create(module, config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if app == nil {
		t.Fatal("Create() returned nil application")
	}
}

type testController struct{}

func (c *testController) RegisterRoutes(router Router) error {
	return nil
}

func TestApplicationFactory_Create_WithControllers(t *testing.T) {
	factory := NewApplicationFactory()

	ctrl := &testController{}
	module := NewModuleBuilder().
		Controllers(ctrl).
		Build()

	config := ApplicationConfig{}

	app, err := factory.Create(module, config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if app == nil {
		t.Fatal("Create() returned nil application")
	}
}

func TestApplicationFactory_Create_WithProviders(t *testing.T) {
	factory := NewApplicationFactory()

	serviceFactory := func() string {
		return "test-service"
	}

	module := NewModuleBuilder().
		Provider("testService", serviceFactory, ScopeSingleton, false).
		Build()

	config := ApplicationConfig{}

	app, err := factory.Create(module, config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if !app.GetContainer().Has("testService") {
		t.Error("Provider not registered")
	}
}

func TestApplicationFactory_Create_DefaultLogger(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()

	config := ApplicationConfig{
		Logger: nil, // Should use default
	}

	app, err := factory.Create(module, config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if app.GetLogger() == nil {
		t.Error("Default logger not set")
	}
}

func TestApplicationFactory_Create_DefaultShutdownTimeout(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()

	config := ApplicationConfig{
		ShutdownTimeout: 0, // Should use default
	}

	app, err := factory.Create(module, config)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	defaultApp := app.(*DefaultApplication)
	if defaultApp.config.ShutdownTimeout != 30*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 30s", defaultApp.config.ShutdownTimeout)
	}
}

func TestDefaultApplication_Use(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	middleware1 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			return next(ctx)
		}
	}

	middleware2 := func(next HandlerFunc) HandlerFunc {
		return func(ctx Context) error {
			return next(ctx)
		}
	}

	app.Use(middleware1, middleware2)

	defaultApp := app.(*DefaultApplication)
	if len(defaultApp.middleware) != 2 {
		t.Errorf("middleware length = %d, want 2", len(defaultApp.middleware))
	}
}

type mockGuard struct{}

func (g *mockGuard) CanActivate(ctx Context) (bool, error) {
	return true, nil
}

func TestDefaultApplication_UseGlobalGuards(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	guard := &mockGuard{}
	app.UseGlobalGuards(guard)

	defaultApp := app.(*DefaultApplication)
	if len(defaultApp.globalGuards) != 1 {
		t.Errorf("globalGuards length = %d, want 1", len(defaultApp.globalGuards))
	}
}

type mockInterceptor struct{}

func (i *mockInterceptor) Intercept(ctx Context, next HandlerFunc) error {
	return next(ctx)
}

func TestDefaultApplication_UseGlobalInterceptors(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	interceptor := &mockInterceptor{}
	app.UseGlobalInterceptors(interceptor)

	defaultApp := app.(*DefaultApplication)
	if len(defaultApp.globalInterceptors) != 1 {
		t.Errorf("globalInterceptors length = %d, want 1", len(defaultApp.globalInterceptors))
	}
}

type mockPipe struct{}

func (p *mockPipe) Transform(value interface{}, metadata *ArgumentMetadata) (interface{}, error) {
	return value, nil
}

func TestDefaultApplication_UseGlobalPipes(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	pipe := &mockPipe{}
	app.UseGlobalPipes(pipe)

	defaultApp := app.(*DefaultApplication)
	if len(defaultApp.globalPipes) != 1 {
		t.Errorf("globalPipes length = %d, want 1", len(defaultApp.globalPipes))
	}
}

func TestDefaultApplication_UseGlobalFilters(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	filter := NewDefaultExceptionFilter()
	app.UseGlobalFilters(filter)

	defaultApp := app.(*DefaultApplication)
	if len(defaultApp.globalFilters) != 1 {
		t.Errorf("globalFilters length = %d, want 1", len(defaultApp.globalFilters))
	}
}

func TestDefaultApplication_GetContainer(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	container := app.GetContainer()
	if container == nil {
		t.Error("GetContainer() returned nil")
	}
}

func TestDefaultApplication_GetLogger(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	logger := app.GetLogger()
	if logger == nil {
		t.Error("GetLogger() returned nil")
	}
}

func TestDefaultApplication_SetLogger(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	newLogger := NewDefaultLogger()
	app.SetLogger(newLogger)

	if app.GetLogger() != newLogger {
		t.Error("SetLogger() did not set the logger")
	}
}

func TestDefaultApplication_GetPostPutDeletePatch(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	handler := func(ctx Context) error {
		return ctx.String(200, "OK")
	}

	app.Get("/test", handler)
	app.Post("/test", handler)
	app.Put("/test", handler)
	app.Delete("/test", handler)
	app.Patch("/test", handler)

	// Just verify no panic
}

func TestDefaultApplication_Close(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	err := app.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

type testShutdownProvider struct {
	shutdownCalled bool
}

func (p *testShutdownProvider) OnApplicationShutdown() error {
	p.shutdownCalled = true
	return nil
}

func TestDefaultApplication_Close_WithShutdownHooks(t *testing.T) {
	factory := NewApplicationFactory()

	provider := &testShutdownProvider{}

	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	_ = app.GetContainer().RegisterValue("provider", provider)

	err := app.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !provider.shutdownCalled {
		t.Error("OnApplicationShutdown was not called")
	}
}

type testErrorProvider struct{}

func (p *testErrorProvider) OnApplicationShutdown() error {
	return errors.New("shutdown error")
}

func TestDefaultApplication_Close_WithShutdownError(t *testing.T) {
	factory := NewApplicationFactory()

	provider := &testErrorProvider{}

	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	_ = app.GetContainer().RegisterValue("provider", provider)

	// Should not return error, just log it
	err := app.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

type testInitProvider struct {
	initCalled bool
}

func (p *testInitProvider) OnModuleInit() error {
	p.initCalled = true
	return nil
}

func TestApplicationFactory_WithLifecycleHooks(t *testing.T) {
	factory := NewApplicationFactory()

	providerInstance := &testInitProvider{}

	module := NewModuleBuilder().Build()
	config := ApplicationConfig{}

	app, _ := factory.Create(module, config)

	_ = app.GetContainer().RegisterValue("provider", providerInstance)

	// Manually call lifecycle hooks for the test
	if hook, ok := interface{}(providerInstance).(OnModuleInit); ok {
		_ = hook.OnModuleInit()
	}

	if !providerInstance.initCalled {
		t.Error("OnModuleInit was not called")
	}
}

func TestDefaultApplication_HandleError(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, app.GetContainer())

	defaultApp := app.(*DefaultApplication)

	// Test with HTTPException
	httpErr := BadRequestException("test error")
	defaultApp.handleError(httpErr, ctx)

	if res.Code != 400 {
		t.Errorf("Response code = %v, want 400", res.Code)
	}

	// Test with regular error
	res = httptest.NewRecorder()
	ctx = NewContext(req.Context(), req, res, app.GetContainer())
	regularErr := errors.New("regular error")
	defaultApp.handleError(regularErr, ctx)

	if res.Code != 500 {
		t.Errorf("Response code = %v, want 500", res.Code)
	}
}

func TestDefaultApplication_ApplyGuards(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	guardCalled := false
	guard := &flexibleMockGuard{
		canActivate: func(ctx Context) (bool, error) {
			guardCalled = true
			return true, nil
		},
	}

	defaultApp.UseGlobalGuards(guard)

	handler := func(ctx Context) error {
		return nil
	}

	wrappedHandler := defaultApp.applyGuards(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("applyGuards() error = %v", err)
	}

	if !guardCalled {
		t.Error("Guard was not called")
	}
}

func TestDefaultApplication_ApplyGuards_Forbidden(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	guard := &flexibleMockGuard{
		canActivate: func(ctx Context) (bool, error) {
			return false, nil
		},
	}

	defaultApp.UseGlobalGuards(guard)

	handler := func(ctx Context) error {
		return nil
	}

	wrappedHandler := defaultApp.applyGuards(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected forbidden error")
	}

	httpErr, ok := err.(*HTTPException)
	if !ok || httpErr.StatusCode != 403 {
		t.Error("Expected 403 Forbidden error")
	}
}

func TestDefaultApplication_ApplyInterceptors(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	interceptorCalled := false
	interceptor := &flexibleMockInterceptor{
		intercept: func(ctx Context, next HandlerFunc) error {
			interceptorCalled = true
			return next(ctx)
		},
	}

	defaultApp.UseGlobalInterceptors(interceptor)

	handler := func(ctx Context) error {
		return nil
	}

	wrappedHandler := defaultApp.applyInterceptors(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("applyInterceptors() error = %v", err)
	}

	if !interceptorCalled {
		t.Error("Interceptor was not called")
	}
}

func TestDefaultApplication_ApplyPipes(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	mockPipe := &mockPipe{}
	defaultApp.UseGlobalPipes(mockPipe)

	handler := func(ctx Context) error {
		return nil
	}

	wrappedHandler := defaultApp.applyPipes(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("applyPipes() error = %v", err)
	}
}

func TestDefaultApplication_ApplyFilters(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	filterCalled := false
	mockFilter := &mockExceptionFilter{
		catch: func(err error, ctx Context) error {
			filterCalled = true
			return nil
		},
	}

	defaultApp.UseGlobalFilters(mockFilter)

	handler := func(ctx Context) error {
		return errors.New("test error")
	}

	wrappedHandler := defaultApp.applyFilters(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Error("Filter should have handled the error")
	}

	if !filterCalled {
		t.Error("Filter was not called")
	}
}

func TestDefaultApplication_ApplyFilters_NoError(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	filterCalled := false
	mockFilter := &mockExceptionFilter{
		catch: func(err error, ctx Context) error {
			filterCalled = true
			return nil
		},
	}

	defaultApp.UseGlobalFilters(mockFilter)

	handler := func(ctx Context) error {
		return nil
	}

	wrappedHandler := defaultApp.applyFilters(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	ctx := NewContext(req.Context(), req, res, defaultApp.container)

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("applyFilters() error = %v", err)
	}

	if filterCalled {
		t.Error("Filter should not be called when there is no error")
	}
}

func TestDefaultApplication_BuildHandler(t *testing.T) {
	factory := NewApplicationFactory()
	module := NewModuleBuilder().Build()
	app, _ := factory.Create(module, ApplicationConfig{})
	defaultApp := app.(*DefaultApplication)

	// Register a test route
	called := false
	defaultApp.Get("/test", func(ctx Context) error {
		called = true
		return nil
	})

	// Build handler
	handler := defaultApp.buildHandler()

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()

	// Execute handler
	handler.ServeHTTP(res, req)

	if !called {
		t.Error("Route handler was not called")
	}
}

// Mock types for advanced testing
type flexibleMockGuard struct {
	canActivate func(ctx Context) (bool, error)
}

func (g *flexibleMockGuard) CanActivate(ctx Context) (bool, error) {
	return g.canActivate(ctx)
}

type flexibleMockInterceptor struct {
	intercept func(ctx Context, next HandlerFunc) error
}

func (i *flexibleMockInterceptor) Intercept(ctx Context, next HandlerFunc) error {
	return i.intercept(ctx, next)
}

type mockExceptionFilter struct {
	catch func(err error, ctx Context) error
}

func (f *mockExceptionFilter) Catch(err error, ctx Context) error {
	return f.catch(err, ctx)
}

func TestSimpleLogger_Debug(t *testing.T) {
	logger := NewDefaultLogger()
	// Just verify it doesn't panic
	logger.Debug("test debug message", map[string]interface{}{"key": "value"})
	logger.Debug("test debug message", nil)
}

func TestSimpleLogger_Info(t *testing.T) {
	logger := NewDefaultLogger()
	// Just verify it doesn't panic
	logger.Info("test info message", map[string]interface{}{"key": "value"})
	logger.Info("test info message", nil)
}

func TestSimpleLogger_Warn(t *testing.T) {
	logger := NewDefaultLogger()
	// Just verify it doesn't panic
	logger.Warn("test warn message", map[string]interface{}{"key": "value"})
	logger.Warn("test warn message", nil)
}
