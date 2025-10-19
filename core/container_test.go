package core

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewContainer(t *testing.T) {
	container := NewContainer()
	if container == nil {
		t.Fatal("NewContainer() returned nil")
	}
}

func TestContainer_RegisterValue(t *testing.T) {
	container := NewContainer()

	value := "test-value"
	err := container.RegisterValue("testValue", value)
	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	result, err := container.Resolve("testValue")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != value {
		t.Errorf("Resolve() = %v, want %v", result, value)
	}
}

func TestContainer_Register_Singleton(t *testing.T) {
	container := NewContainer()

	// Register a singleton factory
	callCount := 0
	factory := func() string {
		callCount++
		return "singleton-value"
	}

	err := container.Register("singleton", factory, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Resolve multiple times
	result1, err := container.Resolve("singleton")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	result2, err := container.Resolve("singleton")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	// Should call factory only once
	if callCount != 1 {
		t.Errorf("Factory called %d times, want 1", callCount)
	}

	// Should return same instance
	if result1 != result2 {
		t.Error("Singleton should return same instance")
	}
}

func TestContainer_Register_Transient(t *testing.T) {
	container := NewContainer()

	callCount := 0
	factory := func() string {
		callCount++
		return "transient-value"
	}

	err := container.Register("transient", factory, WithScope(ScopeTransient))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Resolve multiple times
	_, err = container.Resolve("transient")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	_, err = container.Resolve("transient")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	// Should call factory multiple times for transient
	if callCount != 2 {
		t.Errorf("Factory called %d times, want 2", callCount)
	}
}

func TestContainer_Register_WithDependencies(t *testing.T) {
	container := NewContainer()

	type Config struct {
		Value string
	}

	type Service struct {
		ConfigValue string
	}

	// Register dependency
	configInstance := &Config{Value: "test-config"}
	err := container.RegisterValue("config", configInstance)
	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	// Register provider with dependency
	serviceFactory := func(config *Config) *Service {
		return &Service{ConfigValue: config.Value}
	}

	err = container.Register("service", serviceFactory, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	result, err := container.Resolve("service")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	service, ok := result.(*Service)
	if !ok {
		t.Fatal("Resolve() returned wrong type")
	}

	if service.ConfigValue != "test-config" {
		t.Errorf("Service.ConfigValue = %v, want 'test-config'", service.ConfigValue)
	}
}

func TestContainer_Register_WithErrorReturn(t *testing.T) {
	container := NewContainer()

	factoryWithError := func() (string, error) {
		return "", errors.New("factory error")
	}

	err := container.Register("errorProvider", factoryWithError, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	_, err = container.Resolve("errorProvider")
	if err == nil {
		t.Error("Expected error from factory, got nil")
	}
}

func TestContainer_ResolveByType(t *testing.T) {
	container := NewContainer()

	type TestService struct {
		Name string
	}

	factory := func() *TestService {
		return &TestService{Name: "test"}
	}

	err := container.Register("testService", factory, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	serviceType := reflect.TypeOf(&TestService{})
	result, err := container.ResolveByType(serviceType)
	if err != nil {
		t.Errorf("ResolveByType() error = %v", err)
	}

	service, ok := result.(*TestService)
	if !ok {
		t.Error("ResolveByType() returned wrong type")
	}

	if service.Name != "test" {
		t.Errorf("Service.Name = %v, want 'test'", service.Name)
	}
}

func TestContainer_ResolveByType_NotFound(t *testing.T) {
	container := NewContainer()

	type UnknownService struct{}

	serviceType := reflect.TypeOf(&UnknownService{})
	_, err := container.ResolveByType(serviceType)
	if err == nil {
		t.Error("Expected error for unknown type, got nil")
	}
}

func TestContainer_Resolve_NotFound(t *testing.T) {
	container := NewContainer()

	_, err := container.Resolve("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent provider, got nil")
	}
}

func TestContainer_Has(t *testing.T) {
	container := NewContainer()

	err := container.RegisterValue("existing", "value")
	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	if !container.Has("existing") {
		t.Error("Has() returned false for existing provider")
	}

	if container.Has("nonexistent") {
		t.Error("Has() returned true for non-existent provider")
	}
}

func TestContainer_GetAll(t *testing.T) {
	container := NewContainer()

	err := container.RegisterValue("value1", "test1")
	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	err = container.RegisterValue("value2", "test2")
	if err != nil {
		t.Errorf("RegisterValue() error = %v", err)
	}

	all := container.GetAll()
	if len(all) != 2 {
		t.Errorf("GetAll() returned %d items, want 2", len(all))
	}

	if all["value1"] != "test1" {
		t.Error("GetAll() missing value1")
	}

	if all["value2"] != "test2" {
		t.Error("GetAll() missing value2")
	}
}

func TestContainer_RegisterFactory(t *testing.T) {
	container := NewContainer()

	factory := func() string {
		return "factory-result"
	}

	err := container.RegisterFactory("factory", factory, WithScope(ScopeTransient))
	if err != nil {
		t.Errorf("RegisterFactory() error = %v", err)
	}

	result, err := container.Resolve("factory")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != "factory-result" {
		t.Errorf("Resolve() = %v, want 'factory-result'", result)
	}
}

func TestContainer_WithTags(t *testing.T) {
	container := NewContainer()

	tags := map[string]string{
		"environment": "test",
		"version":     "1.0",
	}

	factory := func() string {
		return "tagged-service"
	}

	err := container.Register("taggedService", factory, WithScope(ScopeSingleton), WithTags(tags))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	result, err := container.Resolve("taggedService")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != "tagged-service" {
		t.Errorf("Resolve() = %v, want 'tagged-service'", result)
	}
}

func TestContainer_MissingDependency(t *testing.T) {
	container := NewContainer()

	type MissingService struct{}

	// Register a factory that depends on a missing service
	factory := func(missing *MissingService) string {
		return "should-not-reach"
	}

	err := container.Register("dependent", factory, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	_, err = container.Resolve("dependent")
	if err == nil {
		t.Error("Expected error for missing dependency, got nil")
	}
}

func TestContainer_NonFunctionFactory(t *testing.T) {
	container := NewContainer()

	// Try to register a non-function as factory
	err := container.Register("invalid", "not-a-function", WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	_, err = container.Resolve("invalid")
	if err == nil {
		t.Error("Expected error for non-function factory, got nil")
	}
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	container := NewContainer()

	err := container.Register("concurrent", func() string {
		return "value"
	}, WithScope(ScopeSingleton))
	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := container.Resolve("concurrent")
			if err != nil {
				t.Errorf("Concurrent Resolve() error = %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

type TestInterface interface {
	DoSomething() string
}

type TestService struct{}

func (s *TestService) DoSomething() string {
	return "done"
}

func TestContainer_WithInterfaces(t *testing.T) {
	container := NewContainer()

	factory := func() *TestService {
		return &TestService{}
	}

	// Register with interface binding
	err := container.Register(
		"testService",
		factory,
		WithScope(ScopeSingleton),
		WithInterfaces((*TestInterface)(nil)),
	)
	if err != nil {
		t.Errorf("Register() with interfaces error = %v", err)
	}

	// Should be able to resolve by interface type
	result, err := container.ResolveByType(reflect.TypeOf((*TestInterface)(nil)).Elem())
	if err != nil {
		t.Errorf("ResolveByType() error = %v", err)
	}

	if result == nil {
		t.Error("ResolveByType() returned nil")
	}

	// Verify the result implements the interface
	_, ok := result.(TestInterface)
	if !ok {
		t.Error("Resolved instance does not implement TestInterface")
	}
}
