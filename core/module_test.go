package core

import (
	"testing"
)

func TestNewModule(t *testing.T) {
	metadata := ModuleMetadata{
		Controllers: []interface{}{},
		Providers:   []ProviderMetadata{},
		Imports:     []Module{},
		Exports:     []string{},
	}

	module := NewModule(metadata)
	if module == nil {
		t.Fatal("NewModule() returned nil")
	}
}

func TestModuleBuilder_Build(t *testing.T) {
	module := NewModuleBuilder().Build()
	if module == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestModuleBuilder_Controllers(t *testing.T) {
	type TestController struct{}
	ctrl := &TestController{}

	module := NewModuleBuilder().
		Controllers(ctrl).
		Build()

	controllers := module.GetControllers()
	if len(controllers) != 1 {
		t.Errorf("GetControllers() length = %d, want 1", len(controllers))
	}
}

func TestModuleBuilder_Provider(t *testing.T) {
	factory := func() string {
		return "test-service"
	}

	module := NewModuleBuilder().
		Provider("testService", factory, ScopeSingleton, true).
		Build()

	providers := module.GetProviders()
	if len(providers) != 1 {
		t.Errorf("GetProviders() length = %d, want 1", len(providers))
	}
}

func TestModuleBuilder_Providers(t *testing.T) {
	factory1 := func() string { return "service1" }
	factory2 := func() string { return "service2" }

	providerMeta1 := ProviderMetadata{
		Name:     "service1",
		Provider: factory1,
		Scope:    ScopeSingleton,
		Exports:  true,
	}

	providerMeta2 := ProviderMetadata{
		Name:     "service2",
		Provider: factory2,
		Scope:    ScopeTransient,
		Exports:  false,
	}

	module := NewModuleBuilder().
		Providers(providerMeta1, providerMeta2).
		Build()

	providers := module.GetProviders()
	if len(providers) != 2 {
		t.Errorf("GetProviders() length = %d, want 2", len(providers))
	}
}

func TestModuleBuilder_Imports(t *testing.T) {
	childModule := NewModuleBuilder().Build()

	parentModule := NewModuleBuilder().
		Imports(childModule).
		Build()

	imports := parentModule.GetImports()
	if len(imports) != 1 {
		t.Errorf("GetImports() length = %d, want 1", len(imports))
	}
}

func TestModuleBuilder_Exports(t *testing.T) {
	module := NewModuleBuilder().
		Exports("service1", "service2").
		Build()

	exports := module.GetExports()
	if len(exports) != 2 {
		t.Errorf("GetExports() length = %d, want 2", len(exports))
	}

	if exports[0] != "service1" || exports[1] != "service2" {
		t.Error("GetExports() returned unexpected values")
	}
}

func TestBaseModule_Configure(t *testing.T) {
	container := NewContainer()

	factory := func() string {
		return "test-value"
	}

	module := NewModuleBuilder().
		Provider("testService", factory, ScopeSingleton, false).
		Build()

	err := module.Configure(container)
	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}

	// Verify provider was registered
	if !container.Has("testService") {
		t.Error("Provider was not registered in container")
	}

	result, err := container.Resolve("testService")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != "test-value" {
		t.Errorf("Resolve() = %v, want 'test-value'", result)
	}
}

func TestBaseModule_Configure_WithImports(t *testing.T) {
	container := NewContainer()

	// Create child module with a provider
	childFactory := func() string {
		return "child-service"
	}
	childModule := NewModuleBuilder().
		Provider("childService", childFactory, ScopeSingleton, true).
		Build()

	// Create parent module that imports child
	parentFactory := func() string {
		return "parent-service"
	}
	parentModule := NewModuleBuilder().
		Provider("parentService", parentFactory, ScopeSingleton, false).
		Imports(childModule).
		Build()

	err := parentModule.Configure(container)
	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}

	// Verify both providers were registered
	if !container.Has("parentService") {
		t.Error("Parent provider was not registered")
	}

	if !container.Has("childService") {
		t.Error("Child provider was not registered")
	}
}

func TestBaseModule_GetControllers(t *testing.T) {
	type TestController1 struct{}
	type TestController2 struct{}

	ctrl1 := &TestController1{}
	ctrl2 := &TestController2{}

	module := NewModuleBuilder().
		Controllers(ctrl1, ctrl2).
		Build()

	controllers := module.GetControllers()
	if len(controllers) != 2 {
		t.Errorf("GetControllers() length = %d, want 2", len(controllers))
	}
}

func TestBaseModule_GetProviders(t *testing.T) {
	factory1 := func() string { return "service1" }
	factory2 := func() string { return "service2" }

	module := NewModuleBuilder().
		Provider("service1", factory1, ScopeSingleton, false).
		Provider("service2", factory2, ScopeTransient, true).
		Build()

	providers := module.GetProviders()
	if len(providers) != 2 {
		t.Errorf("GetProviders() length = %d, want 2", len(providers))
	}
}

func TestBaseModule_GetImports(t *testing.T) {
	child1 := NewModuleBuilder().Build()
	child2 := NewModuleBuilder().Build()

	parent := NewModuleBuilder().
		Imports(child1, child2).
		Build()

	imports := parent.GetImports()
	if len(imports) != 2 {
		t.Errorf("GetImports() length = %d, want 2", len(imports))
	}
}

func TestBaseModule_GetExports(t *testing.T) {
	module := NewModuleBuilder().
		Exports("service1", "service2", "service3").
		Build()

	exports := module.GetExports()
	if len(exports) != 3 {
		t.Errorf("GetExports() length = %d, want 3", len(exports))
	}
}

func TestModuleBuilder_Chaining(t *testing.T) {
	type TestController struct{}
	ctrl := &TestController{}

	childModule := NewModuleBuilder().Build()

	factory := func() string { return "service" }

	// Test method chaining
	module := NewModuleBuilder().
		Controllers(ctrl).
		Provider("service", factory, ScopeSingleton, true).
		Imports(childModule).
		Exports("service").
		Build()

	if module == nil {
		t.Fatal("Build() returned nil")
	}

	if len(module.GetControllers()) != 1 {
		t.Error("Controllers not properly set")
	}

	if len(module.GetProviders()) != 1 {
		t.Error("Providers not properly set")
	}

	if len(module.GetImports()) != 1 {
		t.Error("Imports not properly set")
	}

	if len(module.GetExports()) != 1 {
		t.Error("Exports not properly set")
	}
}

func TestBaseModule_Configure_ErrorHandling(t *testing.T) {
	container := NewContainer()

	// Provider with non-function factory should work at registration
	// but fail at resolution
	module := NewModuleBuilder().
		Provider("invalid", "not-a-function", ScopeSingleton, false).
		Build()

	err := module.Configure(container)
	// Configure should succeed, error happens at Resolve time
	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}
}
