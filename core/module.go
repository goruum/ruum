package core

import "fmt"

// ModuleMetadata contains module configuration
type ModuleMetadata struct {
	Controllers []interface{}
	Providers   []ProviderMetadata
	Imports     []Module
	Exports     []string
}

// ProviderMetadata contains provider configuration
type ProviderMetadata struct {
	Name     string
	Provider interface{}
	Scope    Scope
	Exports  bool
}

// BaseModule is the base implementation of Module interface
type BaseModule struct {
	metadata ModuleMetadata
}

// NewModule creates a new module
func NewModule(metadata ModuleMetadata) Module {
	return &BaseModule{
		metadata: metadata,
	}
}

// Configure registers providers in the DI container.
func (m *BaseModule) Configure(container Container) error {
	// Register providers
	for _, providerMeta := range m.metadata.Providers {
		err := container.Register(
			providerMeta.Name,
			providerMeta.Provider,
			WithScope(providerMeta.Scope),
		)
		if err != nil {
			return fmt.Errorf("failed to register provider %s: %w", providerMeta.Name, err)
		}
	}

	// Configure imported modules
	for _, importedModule := range m.metadata.Imports {
		if err := importedModule.Configure(container); err != nil {
			return fmt.Errorf("failed to configure imported module: %w", err)
		}
	}

	return nil
}

// GetControllers returns the module's controllers.
func (m *BaseModule) GetControllers() []interface{} {
	return m.metadata.Controllers
}

// GetProviders returns the module's providers.
func (m *BaseModule) GetProviders() []interface{} {
	providers := make([]interface{}, len(m.metadata.Providers))
	for i, p := range m.metadata.Providers {
		providers[i] = p.Provider
	}
	return providers
}

// GetImports returns the imported modules.
func (m *BaseModule) GetImports() []Module {
	return m.metadata.Imports
}

// GetExports returns the exported provider names.
func (m *BaseModule) GetExports() []string {
	return m.metadata.Exports
}

// ModuleBuilder helps build modules fluently
type ModuleBuilder struct {
	metadata ModuleMetadata
}

// NewModuleBuilder creates a new module builder
func NewModuleBuilder() *ModuleBuilder {
	return &ModuleBuilder{
		metadata: ModuleMetadata{
			Controllers: make([]interface{}, 0),
			Providers:   make([]ProviderMetadata, 0),
			Imports:     make([]Module, 0),
			Exports:     make([]string, 0),
		},
	}
}

// Controllers adds controllers to the module
func (b *ModuleBuilder) Controllers(controllers ...interface{}) *ModuleBuilder {
	b.metadata.Controllers = append(b.metadata.Controllers, controllers...)
	return b
}

// Providers adds providers to the module
func (b *ModuleBuilder) Providers(providers ...ProviderMetadata) *ModuleBuilder {
	b.metadata.Providers = append(b.metadata.Providers, providers...)
	return b
}

// Provider adds a single provider
func (b *ModuleBuilder) Provider(name string, provider interface{}, scope Scope, exports bool) *ModuleBuilder {
	b.metadata.Providers = append(b.metadata.Providers, ProviderMetadata{
		Name:     name,
		Provider: provider,
		Scope:    scope,
		Exports:  exports,
	})
	return b
}

// Imports adds imported modules
func (b *ModuleBuilder) Imports(modules ...Module) *ModuleBuilder {
	b.metadata.Imports = append(b.metadata.Imports, modules...)
	return b
}

// Exports adds exported provider names
func (b *ModuleBuilder) Exports(names ...string) *ModuleBuilder {
	b.metadata.Exports = append(b.metadata.Exports, names...)
	return b
}

// Build creates the module
func (b *ModuleBuilder) Build() Module {
	return NewModule(b.metadata)
}
