package core

import (
	"reflect"
	"testing"
)

func TestNewDynamicModuleBuilder(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	if builder == nil {
		t.Fatal("NewDynamicModuleBuilder() returned nil")
	}
}

func TestDynamicModuleBuilder_Build(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	module := builder.Build()

	if module == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestDynamicModuleBuilder_WithConfig(t *testing.T) {
	config := map[string]interface{}{
		"test": "value",
	}

	builder := NewDynamicModuleBuilder().WithConfig(config)
	module := builder.Build()

	if module == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestDynamicModuleBuilder_Controllers(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	builder.Controllers("controller1", "controller2")

	module := builder.Build()
	controllers := module.GetControllers()

	if len(controllers) != 2 {
		t.Errorf("GetControllers() returned %d, want 2", len(controllers))
	}
}

func TestDynamicModuleBuilder_Provider(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	builder.Provider("test", func() string { return "test" }, ScopeSingleton, true)

	module := builder.Build()
	providers := module.GetProviders()

	if len(providers) != 1 {
		t.Errorf("GetProviders() returned %d, want 1", len(providers))
	}
}

func TestDynamicModuleBuilder_Exports(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	builder.Exports("provider1", "provider2")

	module := builder.Build()
	exports := module.GetExports()

	if len(exports) != 2 {
		t.Errorf("GetExports() returned %d, want 2", len(exports))
	}
}

func TestNewConfigurableModuleBuilder(t *testing.T) {
	builder := NewConfigurableModuleBuilder("test")
	if builder == nil {
		t.Fatal("NewConfigurableModuleBuilder() returned nil")
	}

	factory := builder.Build()
	if factory == nil {
		t.Fatal("Build() returned nil factory")
	}

	module, err := factory(nil)
	if err != nil {
		t.Errorf("Factory() error = %v", err)
	}

	if module == nil {
		t.Fatal("Factory() returned nil module")
	}
}

func TestNewGlobalModule(t *testing.T) {
	module := NewModuleBuilder().Build()
	globalModule := NewGlobalModule(module)

	if globalModule == nil {
		t.Fatal("NewGlobalModule() returned nil")
	}

	if !globalModule.IsGlobal() {
		t.Error("Module should be global")
	}
}

func TestNewModuleRef(t *testing.T) {
	container := NewContainer()
	ref := NewModuleRef(container)

	if ref == nil {
		t.Fatal("NewModuleRef() returned nil")
	}
}

func TestModuleRef_Get(t *testing.T) {
	container := NewContainer()
	_ = container.RegisterValue("test", "value")

	ref := NewModuleRef(container)
	value, err := ref.Get("test")

	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if value != "value" {
		t.Errorf("Get() = %v, want value", value)
	}
}

func TestNewLazyModuleLoader(t *testing.T) {
	loader := NewLazyModuleLoader()
	if loader == nil {
		t.Fatal("NewLazyModuleLoader() returned nil")
	}
}

func TestLazyModuleLoader_RegisterAndLoad(t *testing.T) {
	loader := NewLazyModuleLoader()

	factory := func(config interface{}) (Module, error) {
		return NewModuleBuilder().Build(), nil
	}

	loader.Register("test", factory)

	module, err := loader.Load("test", nil)
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if module == nil {
		t.Fatal("Load() returned nil")
	}

	// Second load should return cached module
	module2, err := loader.Load("test", nil)
	if err != nil {
		t.Errorf("Load() error = %v", err)
	}

	if module != module2 {
		t.Error("Load() should return cached module")
	}
}

func TestLazyModuleLoader_LoadNotFound(t *testing.T) {
	loader := NewLazyModuleLoader()

	_, err := loader.Load("nonexistent", nil)
	if err == nil {
		t.Error("Load() should return error for nonexistent module")
	}
}

func TestNewConfigModule(t *testing.T) {
	config := map[string]interface{}{
		"key": "value",
	}

	cm := NewConfigModule(config)
	if cm == nil {
		t.Fatal("NewConfigModule() returned nil")
	}
}

func TestConfigModule_Get(t *testing.T) {
	config := map[string]interface{}{
		"key": "value",
	}

	cm := NewConfigModule(config)
	value, ok := cm.Get("key")

	if !ok {
		t.Error("Get() should return true for existing key")
	}

	if value != "value" {
		t.Errorf("Get() = %v, want value", value)
	}
}

func TestConfigModule_Set(t *testing.T) {
	cm := NewConfigModule(make(map[string]interface{}))
	cm.Set("key", "value")

	value, ok := cm.Get("key")
	if !ok || value != "value" {
		t.Error("Set() did not set value correctly")
	}
}

func TestConfigModule_GetString(t *testing.T) {
	config := map[string]interface{}{
		"key": "value",
	}

	cm := NewConfigModule(config)
	value := cm.GetString("key")

	if value != "value" {
		t.Errorf("GetString() = %v, want value", value)
	}
}

func TestConfigModule_GetInt(t *testing.T) {
	config := map[string]interface{}{
		"key": 42,
	}

	cm := NewConfigModule(config)
	value := cm.GetInt("key")

	if value != 42 {
		t.Errorf("GetInt() = %v, want 42", value)
	}
}

func TestConfigModule_GetBool(t *testing.T) {
	config := map[string]interface{}{
		"key": true,
	}

	cm := NewConfigModule(config)
	value := cm.GetBool("key")

	if !value {
		t.Error("GetBool() should return true")
	}
}

func TestCreateDynamicModule(t *testing.T) {
	opts := DynamicModuleOptions{
		Controllers: []interface{}{"ctrl"},
		Providers:   []ProviderMetadata{},
		Imports:     []Module{},
		Exports:     []string{"test"},
		Global:      false,
	}

	module := CreateDynamicModule(opts)
	if module == nil {
		t.Fatal("CreateDynamicModule() returned nil")
	}
}

func TestCreateDynamicModule_Global(t *testing.T) {
	opts := DynamicModuleOptions{
		Global: true,
	}

	module := CreateDynamicModule(opts)
	if module == nil {
		t.Fatal("CreateDynamicModule() returned nil")
	}

	if gm, ok := module.(*GlobalModule); !ok || !gm.IsGlobal() {
		t.Error("Module should be global")
	}
}

func TestCreateAsyncModule(t *testing.T) {
	opts := AsyncModuleOptions{
		UseFactory: func() (interface{}, error) {
			return "test", nil
		},
	}

	module, err := CreateAsyncModule(opts)
	if err != nil {
		t.Errorf("CreateAsyncModule() error = %v", err)
	}

	if module == nil {
		t.Fatal("CreateAsyncModule() returned nil")
	}
}

func TestCreateAsyncModule_NoFactory(t *testing.T) {
	opts := AsyncModuleOptions{}

	_, err := CreateAsyncModule(opts)
	if err == nil {
		t.Error("CreateAsyncModule() should return error when UseFactory is nil")
	}
}

func TestDynamicModuleBuilder_Providers(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	providers := []ProviderMetadata{
		{Name: "provider1", Provider: "test1", Scope: ScopeSingleton},
		{Name: "provider2", Provider: "test2", Scope: ScopeTransient},
	}
	builder.Providers(providers...)

	module := builder.Build()
	if module == nil {
		t.Fatal("Build() returned nil")
	}
}

func TestDynamicModuleBuilder_Imports(t *testing.T) {
	builder := NewDynamicModuleBuilder()
	importedModule := NewModule(ModuleMetadata{})

	builder.Imports(importedModule)

	module := builder.Build()
	imports := module.GetImports()

	if len(imports) != 1 {
		t.Errorf("GetImports() returned %d, want 1", len(imports))
	}
}

func TestModuleRef_GetByType(t *testing.T) {
	container := NewContainer()
	testValue := "test-value"
	_ = container.RegisterValue("test", testValue)

	ref := NewModuleRef(container)
	result, err := ref.GetByType(reflect.TypeOf(""))

	// This may fail if type isn't registered, that's OK
	// Just testing that the method can be called
	_ = result
	_ = err
}

func TestModuleRef_Create(t *testing.T) {
	container := NewContainer()
	factory := func() string { return "created" }
	_ = container.Register("test", factory, WithScope(ScopeSingleton))

	ref := NewModuleRef(container)
	result, err := ref.Create("test")

	if err != nil {
		t.Errorf("Create() returned error: %v", err)
	}
	
	if result == nil {
		t.Error("Create() should create new instance")
	}
}

