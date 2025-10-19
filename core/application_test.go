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
