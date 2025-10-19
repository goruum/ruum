package factory

import (
	"testing"

	"github.com/goruum/ruum/core"
	"github.com/goruum/ruum/logger"
)

func TestCreateApplication(t *testing.T) {
	module := core.NewModuleBuilder().Build()

	app, err := CreateApplication(module)
	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}
}

func TestCreateApplication_WithOptions(t *testing.T) {
	module := core.NewModuleBuilder().Build()
	testLogger := logger.NewDefaultLogger()

	app, err := CreateApplication(
		module,
		WithGlobalPrefix("/api/v1"),
		WithLogger(testLogger),
		WithShutdownHooks(true),
	)

	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}

	if app.GetLogger() != testLogger {
		t.Error("Logger was not set correctly")
	}
}

func TestWithGlobalPrefix(t *testing.T) {
	config := core.ApplicationConfig{}
	opt := WithGlobalPrefix("/api/v1")
	opt(&config)

	if config.GlobalPrefix != "/api/v1" {
		t.Errorf("GlobalPrefix = %v, want '/api/v1'", config.GlobalPrefix)
	}
}

func TestWithCORS(t *testing.T) {
	config := core.ApplicationConfig{}
	opt := WithCORS("https://example.com", "https://test.com")
	opt(&config)

	if !config.CorsEnabled {
		t.Error("CORS should be enabled")
	}

	if len(config.CorsOrigins) != 2 {
		t.Errorf("CorsOrigins length = %d, want 2", len(config.CorsOrigins))
	}

	if config.CorsOrigins[0] != "https://example.com" {
		t.Errorf("CorsOrigins[0] = %v, want 'https://example.com'", config.CorsOrigins[0])
	}
}

func TestWithCORS_NoOrigins(t *testing.T) {
	config := core.ApplicationConfig{
		CorsOrigins: []string{"*"},
	}
	opt := WithCORS()
	opt(&config)

	if !config.CorsEnabled {
		t.Error("CORS should be enabled")
	}

	// When no origins provided, it should not override existing ones
	if len(config.CorsOrigins) < 1 {
		t.Errorf("CorsOrigins length = %d, want at least 1", len(config.CorsOrigins))
	}
}

func TestWithLogger(t *testing.T) {
	config := core.ApplicationConfig{}
	testLogger := logger.NewDefaultLogger()
	opt := WithLogger(testLogger)
	opt(&config)

	if config.Logger != testLogger {
		t.Error("Logger was not set correctly")
	}
}

func TestWithShutdownHooks(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := core.ApplicationConfig{}
			opt := WithShutdownHooks(tt.enabled)
			opt(&config)

			if config.EnableShutdownHooks != tt.enabled {
				t.Errorf("EnableShutdownHooks = %v, want %v", config.EnableShutdownHooks, tt.enabled)
			}
		})
	}
}

func TestCreateApplication_WithController(t *testing.T) {
	type TestController struct {
		*core.BaseModule
	}

	ctrl := &TestController{}

	module := core.NewModuleBuilder().
		Controllers(ctrl).
		Build()

	app, err := CreateApplication(module)
	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}
}

func TestCreateApplication_WithProviders(t *testing.T) {
	factory := func() string {
		return "test-service"
	}

	module := core.NewModuleBuilder().
		Provider("testService", factory, core.ScopeSingleton, false).
		Build()

	app, err := CreateApplication(module)
	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}

	container := app.GetContainer()
	if !container.Has("testService") {
		t.Error("Provider was not registered")
	}
}

func TestCreateApplication_WithImports(t *testing.T) {
	childModule := core.NewModuleBuilder().Build()

	parentModule := core.NewModuleBuilder().
		Imports(childModule).
		Build()

	app, err := CreateApplication(parentModule)
	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}
}

func TestCreateApplication_MultipleOptions(t *testing.T) {
	module := core.NewModuleBuilder().Build()
	testLogger := logger.NewDefaultLogger()

	app, err := CreateApplication(
		module,
		WithGlobalPrefix("/api"),
		WithCORS("*"),
		WithLogger(testLogger),
		WithShutdownHooks(false),
	)

	if err != nil {
		t.Errorf("CreateApplication() error = %v", err)
	}

	if app == nil {
		t.Fatal("CreateApplication() returned nil")
	}
}
