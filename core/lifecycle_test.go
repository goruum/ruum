package core

import (
	"context"
	"testing"
)

type mockBootstrapHook struct {
	called bool
}

func (m *mockBootstrapHook) OnApplicationBootstrap(ctx context.Context) error {
	m.called = true
	return nil
}

type mockShutdownHook struct {
	called bool
}

func (m *mockShutdownHook) OnApplicationShutdown(ctx context.Context) error {
	m.called = true
	return nil
}

type mockBeforeShutdown struct {
	called bool
}

func (m *mockBeforeShutdown) BeforeApplicationShutdown(signal string) error {
	m.called = true
	return nil
}

type mockModuleDestroyHook struct {
	called bool
}

func (m *mockModuleDestroyHook) OnModuleDestroy(ctx context.Context) error {
	m.called = true
	return nil
}

type mockAfterInit struct {
	called bool
}

func (m *mockAfterInit) AfterInit() error {
	m.called = true
	return nil
}

type mockBeforeDestroy struct {
	called bool
}

func (m *mockBeforeDestroy) BeforeDestroy() error {
	m.called = true
	return nil
}

func TestNewLifecycleManager(t *testing.T) {
	lm := NewLifecycleManager()
	if lm == nil {
		t.Fatal("NewLifecycleManager() returned nil")
	}
}

func TestLifecycleManager_RegisterBootstrapHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockBootstrapHook{}

	lm.RegisterBootstrapHook(hook)

	err := lm.CallBootstrapHooks(context.Background())
	if err != nil {
		t.Errorf("CallBootstrapHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Bootstrap hook was not called")
	}
}

func TestLifecycleManager_RegisterShutdownHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockShutdownHook{}

	lm.RegisterShutdownHook(hook)

	err := lm.CallShutdownHooks(context.Background())
	if err != nil {
		t.Errorf("CallShutdownHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Shutdown hook was not called")
	}
}

func TestLifecycleManager_RegisterBeforeShutdownHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockBeforeShutdown{}

	lm.RegisterBeforeShutdownHook(hook)

	err := lm.CallBeforeShutdownHooks("SIGTERM")
	if err != nil {
		t.Errorf("CallBeforeShutdownHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Before shutdown hook was not called")
	}
}

func TestLifecycleManager_RegisterModuleDestroyHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockModuleDestroyHook{}

	lm.RegisterModuleDestroyHook(hook)

	err := lm.CallModuleDestroyHooks(context.Background())
	if err != nil {
		t.Errorf("CallModuleDestroyHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Module destroy hook was not called")
	}
}

func TestLifecycleManager_RegisterAfterInitHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockAfterInit{}

	lm.RegisterAfterInitHook(hook)

	err := lm.CallAfterInitHooks()
	if err != nil {
		t.Errorf("CallAfterInitHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("After init hook was not called")
	}
}

func TestLifecycleManager_RegisterBeforeDestroyHook(t *testing.T) {
	lm := NewLifecycleManager()
	hook := &mockBeforeDestroy{}

	lm.RegisterBeforeDestroyHook(hook)

	err := lm.CallBeforeDestroyHooks()
	if err != nil {
		t.Errorf("CallBeforeDestroyHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Before destroy hook was not called")
	}
}

func TestLifecycleManager_DiscoverAndRegisterHooks(t *testing.T) {
	lm := NewLifecycleManager()
	container := NewContainer()

	hook := &mockBootstrapHook{}
	_ = container.RegisterValue("hook", hook)

	lm.DiscoverAndRegisterHooks(container)

	err := lm.CallBootstrapHooks(context.Background())
	if err != nil {
		t.Errorf("CallBootstrapHooks() error = %v", err)
	}

	if !hook.called {
		t.Error("Discovered hook was not called")
	}
}

func TestNewHealthCheckRegistry(t *testing.T) {
	registry := NewHealthCheckRegistry()
	if registry == nil {
		t.Fatal("NewHealthCheckRegistry() returned nil")
	}
}

type mockHealthIndicator struct{}

func (m *mockHealthIndicator) GetHealth() (HealthStatus, error) {
	return HealthStatus{
		Status: "healthy",
		Details: map[string]interface{}{
			"test": "ok",
		},
	}, nil
}

func TestHealthCheckRegistry_Register(t *testing.T) {
	registry := NewHealthCheckRegistry()
	indicator := &mockHealthIndicator{}

	registry.Register("test", indicator)

	results := registry.Check()
	if len(results) != 1 {
		t.Errorf("Check() returned %d results, want 1", len(results))
	}

	if results["test"].Status != "healthy" {
		t.Error("Status should be healthy")
	}
}

func TestHealthCheckRegistry_Unregister(t *testing.T) {
	registry := NewHealthCheckRegistry()
	indicator := &mockHealthIndicator{}

	registry.Register("test", indicator)
	registry.Unregister("test")

	results := registry.Check()
	if len(results) != 0 {
		t.Errorf("Check() returned %d results, want 0", len(results))
	}
}

func TestNewGracefulShutdownManager(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	if gsm == nil {
		t.Fatal("NewGracefulShutdownManager() returned nil")
	}
}

func TestGracefulShutdownManager_OnShutdown(t *testing.T) {
	gsm := NewGracefulShutdownManager()
	called := false

	gsm.OnShutdown(func(signal TerminationSignal) {
		called = true
	})

	go gsm.Shutdown("SIGTERM", "test shutdown")

	signal := gsm.Wait()
	if signal.Signal != "SIGTERM" {
		t.Errorf("Signal = %v, want SIGTERM", signal.Signal)
	}

	if !called {
		t.Error("Shutdown listener was not called")
	}
}

