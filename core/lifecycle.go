// Package core provides advanced lifecycle management.
package core

import (
	"context"
	"fmt"
	"sync"
)

// Additional lifecycle hooks

// OnApplicationBootstrap is called when the application starts (after all modules are initialized)
type OnApplicationBootstrapHook interface {
	OnApplicationBootstrap(ctx context.Context) error
}

// BeforeApplicationShutdown is called before the application begins shutdown
type BeforeApplicationShutdown interface {
	BeforeApplicationShutdown(signal string) error
}

// OnApplicationShutdownHook is called during application shutdown
type OnApplicationShutdownHook interface {
	OnApplicationShutdown(ctx context.Context) error
}

// OnModuleDestroy is called when a module is being destroyed
type OnModuleDestroyHook interface {
	OnModuleDestroy(ctx context.Context) error
}

// AfterInit is called after initialization is complete
type AfterInit interface {
	AfterInit() error
}

// BeforeDestroy is called before destruction begins
type BeforeDestroy interface {
	BeforeDestroy() error
}

// LifecycleManager manages application lifecycle hooks
type LifecycleManager struct {
	bootstrapHooks []OnApplicationBootstrapHook
	shutdownHooks  []OnApplicationShutdownHook
	beforeShutdown []BeforeApplicationShutdown
	moduleDestroy  []OnModuleDestroyHook
	afterInit      []AfterInit
	beforeDestroy  []BeforeDestroy
	mu             sync.RWMutex
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager() *LifecycleManager {
	return &LifecycleManager{
		bootstrapHooks: make([]OnApplicationBootstrapHook, 0),
		shutdownHooks:  make([]OnApplicationShutdownHook, 0),
		beforeShutdown: make([]BeforeApplicationShutdown, 0),
		moduleDestroy:  make([]OnModuleDestroyHook, 0),
		afterInit:      make([]AfterInit, 0),
		beforeDestroy:  make([]BeforeDestroy, 0),
	}
}

// RegisterBootstrapHook registers a bootstrap hook
func (lm *LifecycleManager) RegisterBootstrapHook(hook OnApplicationBootstrapHook) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.bootstrapHooks = append(lm.bootstrapHooks, hook)
}

// RegisterShutdownHook registers a shutdown hook
func (lm *LifecycleManager) RegisterShutdownHook(hook OnApplicationShutdownHook) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.shutdownHooks = append(lm.shutdownHooks, hook)
}

// RegisterBeforeShutdownHook registers a before-shutdown hook
func (lm *LifecycleManager) RegisterBeforeShutdownHook(hook BeforeApplicationShutdown) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.beforeShutdown = append(lm.beforeShutdown, hook)
}

// RegisterModuleDestroyHook registers a module destroy hook
func (lm *LifecycleManager) RegisterModuleDestroyHook(hook OnModuleDestroyHook) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.moduleDestroy = append(lm.moduleDestroy, hook)
}

// RegisterAfterInitHook registers an after-init hook
func (lm *LifecycleManager) RegisterAfterInitHook(hook AfterInit) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.afterInit = append(lm.afterInit, hook)
}

// RegisterBeforeDestroyHook registers a before-destroy hook
func (lm *LifecycleManager) RegisterBeforeDestroyHook(hook BeforeDestroy) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.beforeDestroy = append(lm.beforeDestroy, hook)
}

// CallBootstrapHooks calls all bootstrap hooks
func (lm *LifecycleManager) CallBootstrapHooks(ctx context.Context) error {
	lm.mu.RLock()
	hooks := make([]OnApplicationBootstrapHook, len(lm.bootstrapHooks))
	copy(hooks, lm.bootstrapHooks)
	lm.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.OnApplicationBootstrap(ctx); err != nil {
			return fmt.Errorf("bootstrap hook failed: %w", err)
		}
	}
	return nil
}

// CallShutdownHooks calls all shutdown hooks
func (lm *LifecycleManager) CallShutdownHooks(ctx context.Context) error {
	lm.mu.RLock()
	hooks := make([]OnApplicationShutdownHook, len(lm.shutdownHooks))
	copy(hooks, lm.shutdownHooks)
	lm.mu.RUnlock()

	var errors []error
	for _, hook := range hooks {
		if err := hook.OnApplicationShutdown(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown hooks failed: %v", errors)
	}
	return nil
}

// CallBeforeShutdownHooks calls all before-shutdown hooks
func (lm *LifecycleManager) CallBeforeShutdownHooks(signal string) error {
	lm.mu.RLock()
	hooks := make([]BeforeApplicationShutdown, len(lm.beforeShutdown))
	copy(hooks, lm.beforeShutdown)
	lm.mu.RUnlock()

	var errors []error
	for _, hook := range hooks {
		if err := hook.BeforeApplicationShutdown(signal); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("before-shutdown hooks failed: %v", errors)
	}
	return nil
}

// CallModuleDestroyHooks calls all module destroy hooks
func (lm *LifecycleManager) CallModuleDestroyHooks(ctx context.Context) error {
	lm.mu.RLock()
	hooks := make([]OnModuleDestroyHook, len(lm.moduleDestroy))
	copy(hooks, lm.moduleDestroy)
	lm.mu.RUnlock()

	var errors []error
	for _, hook := range hooks {
		if err := hook.OnModuleDestroy(ctx); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("module destroy hooks failed: %v", errors)
	}
	return nil
}

// CallAfterInitHooks calls all after-init hooks
func (lm *LifecycleManager) CallAfterInitHooks() error {
	lm.mu.RLock()
	hooks := make([]AfterInit, len(lm.afterInit))
	copy(hooks, lm.afterInit)
	lm.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.AfterInit(); err != nil {
			return fmt.Errorf("after-init hook failed: %w", err)
		}
	}
	return nil
}

// CallBeforeDestroyHooks calls all before-destroy hooks
func (lm *LifecycleManager) CallBeforeDestroyHooks() error {
	lm.mu.RLock()
	hooks := make([]BeforeDestroy, len(lm.beforeDestroy))
	copy(hooks, lm.beforeDestroy)
	lm.mu.RUnlock()

	var errors []error
	for _, hook := range hooks {
		if err := hook.BeforeDestroy(); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("before-destroy hooks failed: %v", errors)
	}
	return nil
}

// DiscoverAndRegisterHooks discovers and registers lifecycle hooks from container
func (lm *LifecycleManager) DiscoverAndRegisterHooks(container Container) {
	for name := range container.GetAll() {
		instance, err := container.Resolve(name)
		if err != nil {
			continue
		}

		// Check and register each hook type
		if hook, ok := instance.(OnApplicationBootstrapHook); ok {
			lm.RegisterBootstrapHook(hook)
		}

		if hook, ok := instance.(OnApplicationShutdownHook); ok {
			lm.RegisterShutdownHook(hook)
		}

		if hook, ok := instance.(BeforeApplicationShutdown); ok {
			lm.RegisterBeforeShutdownHook(hook)
		}

		if hook, ok := instance.(OnModuleDestroyHook); ok {
			lm.RegisterModuleDestroyHook(hook)
		}

		if hook, ok := instance.(AfterInit); ok {
			lm.RegisterAfterInitHook(hook)
		}

		if hook, ok := instance.(BeforeDestroy); ok {
			lm.RegisterBeforeDestroyHook(hook)
		}

		// Support legacy interfaces for backward compatibility
		if hook, ok := instance.(OnModuleInit); ok {
			lm.RegisterAfterInitHook(legacyInitWrapper{hook})
		}

		if hook, ok := instance.(OnApplicationShutdown); ok {
			lm.RegisterShutdownHook(legacyShutdownWrapper{hook})
		}
	}
}

// Wrappers for legacy interfaces

type legacyInitWrapper struct {
	hook OnModuleInit
}

func (w legacyInitWrapper) AfterInit() error {
	return w.hook.OnModuleInit()
}

type legacyShutdownWrapper struct {
	hook OnApplicationShutdown
}

func (w legacyShutdownWrapper) OnApplicationShutdown(ctx context.Context) error {
	return w.hook.OnApplicationShutdown()
}

// HealthIndicator provides health status for health checks
type HealthIndicator interface {
	GetHealth() (HealthStatus, error)
}

// HealthStatus represents a health status
type HealthStatus struct {
	Status  string                 `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthCheckRegistry manages health indicators
type HealthCheckRegistry struct {
	indicators map[string]HealthIndicator
	mu         sync.RWMutex
}

// NewHealthCheckRegistry creates a new health check registry
func NewHealthCheckRegistry() *HealthCheckRegistry {
	return &HealthCheckRegistry{
		indicators: make(map[string]HealthIndicator),
	}
}

// Register registers a health indicator
func (hcr *HealthCheckRegistry) Register(name string, indicator HealthIndicator) {
	hcr.mu.Lock()
	defer hcr.mu.Unlock()
	hcr.indicators[name] = indicator
}

// Unregister removes a health indicator
func (hcr *HealthCheckRegistry) Unregister(name string) {
	hcr.mu.Lock()
	defer hcr.mu.Unlock()
	delete(hcr.indicators, name)
}

// Check runs all health checks
func (hcr *HealthCheckRegistry) Check() map[string]HealthStatus {
	hcr.mu.RLock()
	defer hcr.mu.RUnlock()

	results := make(map[string]HealthStatus)
	for name, indicator := range hcr.indicators {
		status, err := indicator.GetHealth()
		if err != nil {
			results[name] = HealthStatus{
				Status: "unhealthy",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			}
		} else {
			results[name] = status
		}
	}

	return results
}

// TerminationSignal represents a termination signal
type TerminationSignal struct {
	Signal string
	Reason string
}

// GracefulShutdownManager manages graceful shutdown
type GracefulShutdownManager struct {
	signals   chan TerminationSignal
	listeners []func(TerminationSignal)
	mu        sync.RWMutex
}

// NewGracefulShutdownManager creates a new graceful shutdown manager
func NewGracefulShutdownManager() *GracefulShutdownManager {
	return &GracefulShutdownManager{
		signals:   make(chan TerminationSignal, 10),
		listeners: make([]func(TerminationSignal), 0),
	}
}

// OnShutdown registers a shutdown listener
func (gsm *GracefulShutdownManager) OnShutdown(listener func(TerminationSignal)) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	gsm.listeners = append(gsm.listeners, listener)
}

// Shutdown initiates graceful shutdown
func (gsm *GracefulShutdownManager) Shutdown(signal, reason string) {
	termSignal := TerminationSignal{
		Signal: signal,
		Reason: reason,
	}

	gsm.signals <- termSignal

	gsm.mu.RLock()
	listeners := make([]func(TerminationSignal), len(gsm.listeners))
	copy(listeners, gsm.listeners)
	gsm.mu.RUnlock()

	for _, listener := range listeners {
		listener(termSignal)
	}
}

// Wait waits for a shutdown signal
func (gsm *GracefulShutdownManager) Wait() TerminationSignal {
	return <-gsm.signals
}
