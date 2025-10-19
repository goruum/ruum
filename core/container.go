package core

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is the dependency injection container
type Container interface {
	Register(name string, provider interface{}, opts ...ProviderOption) error
	RegisterFactory(name string, factory interface{}, opts ...ProviderOption) error
	RegisterValue(name string, value interface{}) error
	Resolve(name string) (interface{}, error)
	ResolveByType(t reflect.Type) (interface{}, error)
	Has(name string) bool
	GetAll() map[string]interface{}
}

// Scope defines the lifecycle of a provider
type Scope string

// Provider scopes
const (
	// ScopeSingleton creates a single instance shared across the application
	ScopeSingleton Scope = "singleton"
	// ScopeTransient creates a new instance on every request
	ScopeTransient Scope = "transient"
	// ScopeRequest creates a new instance per request
	ScopeRequest   Scope = "request"
)

// ProviderOption configures a provider
type ProviderOption func(*providerConfig)

type providerConfig struct {
	scope      Scope
	tags       map[string]string
	interfaces []reflect.Type
}

// WithScope sets the provider scope
func WithScope(scope Scope) ProviderOption {
	return func(c *providerConfig) {
		c.scope = scope
	}
}

// WithTags adds metadata tags to the provider
func WithTags(tags map[string]string) ProviderOption {
	return func(c *providerConfig) {
		c.tags = tags
	}
}

// WithInterfaces registers the provider for given interfaces
func WithInterfaces(interfaces ...interface{}) ProviderOption {
	return func(c *providerConfig) {
		for _, iface := range interfaces {
			c.interfaces = append(c.interfaces, reflect.TypeOf(iface).Elem())
		}
	}
}

type provider struct {
	factory  interface{}
	instance interface{}
	config   *providerConfig
	isValue  bool
	mu       sync.RWMutex
}

// DefaultContainer implements Container interface
type DefaultContainer struct {
	providers map[string]*provider
	typeMap   map[reflect.Type]string
	mu        sync.RWMutex
}

// NewContainer creates a new DI container
func NewContainer() Container {
	return &DefaultContainer{
		providers: make(map[string]*provider),
		typeMap:   make(map[reflect.Type]string),
	}
}

// Register registers a provider with the given name and options.
func (c *DefaultContainer) Register(name string, providerFunc interface{}, opts ...ProviderOption) error {
	config := &providerConfig{
		scope: ScopeSingleton,
		tags:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(config)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	p := &provider{
		factory: providerFunc,
		config:  config,
		isValue: false,
	}

	c.providers[name] = p

	// Register by type
	providerType := reflect.TypeOf(providerFunc)
	if providerType.Kind() == reflect.Func && providerType.NumOut() > 0 {
		returnType := providerType.Out(0)
		c.typeMap[returnType] = name

		// Register for interfaces
		for _, iface := range config.interfaces {
			c.typeMap[iface] = name
		}
	}

	return nil
}

// RegisterFactory registers a factory function.
func (c *DefaultContainer) RegisterFactory(name string, factory interface{}, opts ...ProviderOption) error {
	config := &providerConfig{
		scope: ScopeTransient,
		tags:  make(map[string]string),
	}

	for _, opt := range opts {
		opt(config)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	p := &provider{
		factory: factory,
		config:  config,
		isValue: false,
	}

	c.providers[name] = p
	return nil
}

// RegisterValue registers a static value.
func (c *DefaultContainer) RegisterValue(name string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	p := &provider{
		instance: value,
		isValue:  true,
		config: &providerConfig{
			scope: ScopeSingleton,
		},
	}

	c.providers[name] = p

	// Register by type
	valueType := reflect.TypeOf(value)
	c.typeMap[valueType] = name

	return nil
}

// Resolve resolves a provider by name.
func (c *DefaultContainer) Resolve(name string) (interface{}, error) {
	c.mu.RLock()
	p, exists := c.providers[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return c.resolveProvider(p)
}

// ResolveByType resolves a provider by type.
func (c *DefaultContainer) ResolveByType(t reflect.Type) (interface{}, error) {
	c.mu.RLock()
	name, exists := c.typeMap[t]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no provider registered for type %s", t.String())
	}

	return c.Resolve(name)
}

func (c *DefaultContainer) resolveProvider(p *provider) (interface{}, error) {
	if p.isValue {
		return p.instance, nil
	}

	// Singleton: return cached instance
	if p.config.scope == ScopeSingleton {
		p.mu.RLock()
		if p.instance != nil {
			instance := p.instance
			p.mu.RUnlock()
			return instance, nil
		}
		p.mu.RUnlock()

		p.mu.Lock()
		defer p.mu.Unlock()

		// Double check
		if p.instance != nil {
			return p.instance, nil
		}

		instance, err := c.invokeFactory(p.factory)
		if err != nil {
			return nil, err
		}

		p.instance = instance
		return instance, nil
	}

	// Transient: create new instance
	return c.invokeFactory(p.factory)
}

func (c *DefaultContainer) invokeFactory(factory interface{}) (interface{}, error) {
	factoryValue := reflect.ValueOf(factory)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		return nil, fmt.Errorf("factory must be a function")
	}

	// Resolve dependencies
	args := make([]reflect.Value, factoryType.NumIn())
	for i := 0; i < factoryType.NumIn(); i++ {
		argType := factoryType.In(i)

		// Try to resolve by type
		dep, err := c.ResolveByType(argType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %s: %w", argType.String(), err)
		}

		args[i] = reflect.ValueOf(dep)
	}

	// Call factory
	results := factoryValue.Call(args)

	if len(results) == 0 {
		return nil, fmt.Errorf("factory must return at least one value")
	}

	// Check for error return
	if len(results) == 2 {
		if err, ok := results[1].Interface().(error); ok && err != nil {
			return nil, err
		}
	}

	return results[0].Interface(), nil
}

// Has checks if a provider exists.
func (c *DefaultContainer) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.providers[name]
	return exists
}

// GetAll returns all registered providers.
func (c *DefaultContainer) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for name := range c.providers {
		if instance, err := c.Resolve(name); err == nil {
			result[name] = instance
		}
	}
	return result
}
