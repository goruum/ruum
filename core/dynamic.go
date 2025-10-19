// Package core provides dynamic module support.
package core

import (
	"fmt"
	"reflect"
	"sync"
)

// DynamicModule represents a module that can be configured dynamically
type DynamicModule interface {
	Module
	Configure(container Container) error
}

// DynamicModuleBuilder helps build dynamic modules
type DynamicModuleBuilder struct {
	metadata ModuleMetadata
	config   interface{}
}

// NewDynamicModuleBuilder creates a new dynamic module builder
func NewDynamicModuleBuilder() *DynamicModuleBuilder {
	return &DynamicModuleBuilder{
		metadata: ModuleMetadata{
			Controllers: make([]interface{}, 0),
			Providers:   make([]ProviderMetadata, 0),
			Imports:     make([]Module, 0),
			Exports:     make([]string, 0),
		},
	}
}

// WithConfig sets the module configuration
func (b *DynamicModuleBuilder) WithConfig(config interface{}) *DynamicModuleBuilder {
	b.config = config
	return b
}

// Controllers adds controllers to the module
func (b *DynamicModuleBuilder) Controllers(controllers ...interface{}) *DynamicModuleBuilder {
	b.metadata.Controllers = append(b.metadata.Controllers, controllers...)
	return b
}

// Providers adds providers to the module
func (b *DynamicModuleBuilder) Providers(providers ...ProviderMetadata) *DynamicModuleBuilder {
	b.metadata.Providers = append(b.metadata.Providers, providers...)
	return b
}

// Provider adds a single provider
func (b *DynamicModuleBuilder) Provider(name string, provider interface{}, scope Scope, exports bool) *DynamicModuleBuilder {
	b.metadata.Providers = append(b.metadata.Providers, ProviderMetadata{
		Name:     name,
		Provider: provider,
		Scope:    scope,
		Exports:  exports,
	})
	return b
}

// Imports adds imported modules
func (b *DynamicModuleBuilder) Imports(modules ...Module) *DynamicModuleBuilder {
	b.metadata.Imports = append(b.metadata.Imports, modules...)
	return b
}

// Exports adds exported provider names
func (b *DynamicModuleBuilder) Exports(names ...string) *DynamicModuleBuilder {
	b.metadata.Exports = append(b.metadata.Exports, names...)
	return b
}

// Build creates the dynamic module
func (b *DynamicModuleBuilder) Build() Module {
	return NewModule(b.metadata)
}

// ModuleFactory is a factory function that creates a module
type ModuleFactory func(config interface{}) (Module, error)

// DynamicModuleOptions contains options for creating a dynamic module
type DynamicModuleOptions struct {
	Module      Module
	Controllers []interface{}
	Providers   []ProviderMetadata
	Imports     []Module
	Exports     []string
	Global      bool
}

// ForRoot creates a root dynamic module (singleton)
type ForRoot interface {
	ForRoot(config interface{}) (Module, error)
}

// ForFeature creates a feature dynamic module (per-import)
type ForFeature interface {
	ForFeature(config interface{}) (Module, error)
}

// ForRootAsync creates a root dynamic module asynchronously
type ForRootAsync interface {
	ForRootAsync(config interface{}) (Module, error)
}

// AsyncModuleOptions contains options for async modules
type AsyncModuleOptions struct {
	UseFactory func() (interface{}, error)
	Inject     []string
	Imports    []Module
}

// ConfigurableModuleBuilder helps build configurable modules
type ConfigurableModuleBuilder struct {
	moduleName string
}

// NewConfigurableModuleBuilder creates a new configurable module builder
func NewConfigurableModuleBuilder(moduleName string) *ConfigurableModuleBuilder {
	return &ConfigurableModuleBuilder{
		moduleName: moduleName,
	}
}

// Build returns a module factory
func (cmb *ConfigurableModuleBuilder) Build() ModuleFactory {
	return func(config interface{}) (Module, error) {
		return NewDynamicModuleBuilder().
			WithConfig(config).
			Build(), nil
	}
}

// GlobalModule marks a module as global
type GlobalModule struct {
	Module
	isGlobal bool
}

// NewGlobalModule creates a new global module
func NewGlobalModule(module Module) *GlobalModule {
	return &GlobalModule{
		Module:   module,
		isGlobal: true,
	}
}

// IsGlobal returns whether the module is global
func (gm *GlobalModule) IsGlobal() bool {
	return gm.isGlobal
}

// ModuleRef provides access to module providers
type ModuleRef struct {
	container Container
	mu        *sync.RWMutex
}

// NewModuleRef creates a new module reference
func NewModuleRef(container Container) *ModuleRef {
	return &ModuleRef{
		container: container,
		mu:        &sync.RWMutex{},
	}
}

// Get retrieves a provider by name
func (mr *ModuleRef) Get(name string) (interface{}, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.container.Resolve(name)
}

// GetByType retrieves a provider by type
func (mr *ModuleRef) GetByType(t reflect.Type) (interface{}, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	return mr.container.ResolveByType(t)
}

// Create creates a new instance (for transient providers)
func (mr *ModuleRef) Create(name string) (interface{}, error) {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	// This would create a new instance for transient scopes
	return mr.container.Resolve(name)
}

// LazyModuleLoader loads modules lazily
type LazyModuleLoader struct {
	modules map[string]ModuleFactory
	loaded  map[string]Module
	mu      sync.RWMutex
}

// NewLazyModuleLoader creates a new lazy module loader
func NewLazyModuleLoader() *LazyModuleLoader {
	return &LazyModuleLoader{
		modules: make(map[string]ModuleFactory),
		loaded:  make(map[string]Module),
	}
}

// Register registers a module factory
func (lml *LazyModuleLoader) Register(name string, factory ModuleFactory) {
	lml.mu.Lock()
	defer lml.mu.Unlock()
	lml.modules[name] = factory
}

// Load loads a module by name
func (lml *LazyModuleLoader) Load(name string, config interface{}) (Module, error) {
	lml.mu.RLock()
	if module, ok := lml.loaded[name]; ok {
		lml.mu.RUnlock()
		return module, nil
	}
	lml.mu.RUnlock()

	lml.mu.Lock()
	defer lml.mu.Unlock()

	// Double-check after acquiring write lock
	if module, ok := lml.loaded[name]; ok {
		return module, nil
	}

	factory, ok := lml.modules[name]
	if !ok {
		return nil, fmt.Errorf("module %s not registered", name)
	}

	module, err := factory(config)
	if err != nil {
		return nil, err
	}

	lml.loaded[name] = module
	return module, nil
}

// ConfigModule provides configuration for dynamic modules
type ConfigModule struct {
	config map[string]interface{}
	mu     sync.RWMutex
}

// NewConfigModule creates a new config module
func NewConfigModule(config map[string]interface{}) *ConfigModule {
	return &ConfigModule{
		config: config,
	}
}

// Get retrieves a configuration value
func (cm *ConfigModule) Get(key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	val, ok := cm.config[key]
	return val, ok
}

// Set sets a configuration value
func (cm *ConfigModule) Set(key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.config[key] = value
}

// GetString retrieves a string configuration value
func (cm *ConfigModule) GetString(key string) string {
	if val, ok := cm.Get(key); ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt retrieves an int configuration value
func (cm *ConfigModule) GetInt(key string) int {
	if val, ok := cm.Get(key); ok {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool retrieves a bool configuration value
func (cm *ConfigModule) GetBool(key string) bool {
	if val, ok := cm.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// Helper functions for creating dynamic modules

// CreateDynamicModule creates a dynamic module from options
func CreateDynamicModule(opts DynamicModuleOptions) Module {
	builder := NewDynamicModuleBuilder()

	if len(opts.Controllers) > 0 {
		builder.Controllers(opts.Controllers...)
	}

	if len(opts.Providers) > 0 {
		builder.Providers(opts.Providers...)
	}

	if len(opts.Imports) > 0 {
		builder.Imports(opts.Imports...)
	}

	if len(opts.Exports) > 0 {
		builder.Exports(opts.Exports...)
	}

	module := builder.Build()

	if opts.Global {
		return NewGlobalModule(module)
	}

	return module
}

// CreateAsyncModule creates an async module from options
func CreateAsyncModule(opts AsyncModuleOptions) (Module, error) {
	if opts.UseFactory == nil {
		return nil, fmt.Errorf("useFactory is required for async modules")
	}

	// Execute factory
	result, err := opts.UseFactory()
	if err != nil {
		return nil, err
	}

	// Create module with result
	builder := NewDynamicModuleBuilder()

	if len(opts.Imports) > 0 {
		builder.Imports(opts.Imports...)
	}

	// Register the factory result as a provider
	if result != nil {
		builder.Provider("asyncResult", func() interface{} { return result }, ScopeSingleton, true)
	}

	return builder.Build(), nil
}
