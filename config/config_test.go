package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewConfigService(t *testing.T) {
	config := NewConfigService()
	if config == nil {
		t.Fatal("NewConfigService() returned nil")
	}
}

func TestConfigService_Get(t *testing.T) {
	config := NewConfigService()
	config.Set("test-key", "test-value")

	value := config.Get("test-key")
	if value != "test-value" {
		t.Errorf("Get() = %v, want 'test-value'", value)
	}
}

func TestConfigService_Get_NonExistent(t *testing.T) {
	config := NewConfigService()

	value := config.Get("nonexistent")
	if value != nil {
		t.Errorf("Get() = %v, want nil", value)
	}
}

func TestConfigService_Get_FromEnv(t *testing.T) {
	config := NewConfigService()

	// Set environment variable
	err := os.Setenv("TEST_ENV_VAR", "env-value")
	if err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_ENV_VAR")
	}()

	value := config.Get("TEST_ENV_VAR")
	if value != "env-value" {
		t.Errorf("Get() = %v, want 'env-value'", value)
	}
}

func TestConfigService_GetString(t *testing.T) {
	config := NewConfigService()

	tests := []struct {
		name     string
		key      string
		setValue interface{}
		expected string
	}{
		{
			name:     "string value",
			key:      "string",
			setValue: "test",
			expected: "test",
		},
		{
			name:     "int value",
			key:      "int",
			setValue: 42,
			expected: "42",
		},
		{
			name:     "nonexistent",
			key:      "nonexistent",
			setValue: nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != nil {
				config.Set(tt.key, tt.setValue)
			}

			result := config.GetString(tt.key)
			if result != tt.expected {
				t.Errorf("GetString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigService_GetInt(t *testing.T) {
	config := NewConfigService()

	tests := []struct {
		name     string
		key      string
		setValue interface{}
		expected int
	}{
		{
			name:     "int value",
			key:      "int",
			setValue: 42,
			expected: 42,
		},
		{
			name:     "float64 value",
			key:      "float",
			setValue: 42.5,
			expected: 42,
		},
		{
			name:     "string value",
			key:      "string",
			setValue: "123",
			expected: 123,
		},
		{
			name:     "nonexistent",
			key:      "nonexistent",
			setValue: nil,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != nil {
				config.Set(tt.key, tt.setValue)
			}

			result := config.GetInt(tt.key)
			if result != tt.expected {
				t.Errorf("GetInt() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigService_GetBool(t *testing.T) {
	config := NewConfigService()

	tests := []struct {
		name     string
		key      string
		setValue interface{}
		expected bool
	}{
		{
			name:     "bool true",
			key:      "bool",
			setValue: true,
			expected: true,
		},
		{
			name:     "bool false",
			key:      "bool",
			setValue: false,
			expected: false,
		},
		{
			name:     "string true",
			key:      "string",
			setValue: "true",
			expected: true,
		},
		{
			name:     "string 1",
			key:      "string",
			setValue: "1",
			expected: true,
		},
		{
			name:     "string yes",
			key:      "string",
			setValue: "yes",
			expected: true,
		},
		{
			name:     "int non-zero",
			key:      "int",
			setValue: 42,
			expected: true,
		},
		{
			name:     "int zero",
			key:      "int",
			setValue: 0,
			expected: false,
		},
		{
			name:     "nonexistent",
			key:      "nonexistent",
			setValue: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != nil {
				config.Set(tt.key, tt.setValue)
			}

			result := config.GetBool(tt.key)
			if result != tt.expected {
				t.Errorf("GetBool() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConfigService_Set(t *testing.T) {
	config := NewConfigService()

	config.Set("key1", "value1")
	config.Set("key2", 123)
	config.Set("key3", true)

	if config.Get("key1") != "value1" {
		t.Error("Set() failed for string")
	}

	if config.Get("key2") != 123 {
		t.Error("Set() failed for int")
	}

	if config.Get("key3") != true {
		t.Error("Set() failed for bool")
	}
}

func TestConfigService_Has(t *testing.T) {
	config := NewConfigService()

	config.Set("existing", "value")

	if !config.Has("existing") {
		t.Error("Has() returned false for existing key")
	}

	if config.Has("nonexistent") {
		t.Error("Has() returned true for non-existent key")
	}
}

func TestConfigService_Has_FromEnv(t *testing.T) {
	config := NewConfigService()

	err := os.Setenv("TEST_HAS_ENV", "value")
	if err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	defer func() {
		_ = os.Unsetenv("TEST_HAS_ENV")
	}()

	if !config.Has("TEST_HAS_ENV") {
		t.Error("Has() returned false for environment variable")
	}
}

func TestConfigService_LoadFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")

	configData := `{
		"app_name": "test-app",
		"port": 3000,
		"debug": true
	}`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigService()
	err = config.LoadFromFile(configFile)
	if err != nil {
		t.Errorf("LoadFromFile() error = %v", err)
	}

	if config.GetString("app_name") != "test-app" {
		t.Error("LoadFromFile() failed to load app_name")
	}

	if config.GetInt("port") != 3000 {
		t.Error("LoadFromFile() failed to load port")
	}

	if !config.GetBool("debug") {
		t.Error("LoadFromFile() failed to load debug")
	}
}

func TestConfigService_LoadFromFile_NonExistent(t *testing.T) {
	config := NewConfigService()

	err := config.LoadFromFile("/nonexistent/config.json")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestConfigService_LoadFromFile_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(configFile, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := NewConfigService()
	err = config.LoadFromFile(configFile)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestConfigService_LoadFromEnv(t *testing.T) {
	config := NewConfigService()

	// LoadFromEnv is a no-op in the current implementation
	err := config.LoadFromEnv("TEST_")
	if err != nil {
		t.Errorf("LoadFromEnv() error = %v", err)
	}
}

func TestNewConfigModule(t *testing.T) {
	config := NewConfigService()
	module := NewConfigModule(config)

	if module == nil {
		t.Fatal("NewConfigModule() returned nil")
	}

	if module.config != config {
		t.Error("Config not set correctly")
	}
}

func TestNewConfigModule_NilConfig(t *testing.T) {
	module := NewConfigModule(nil)

	if module == nil {
		t.Fatal("NewConfigModule() returned nil")
	}

	if module.config == nil {
		t.Error("Config should be created automatically")
	}
}

func TestConfigModule_Configure(t *testing.T) {
	container := core.NewContainer()
	config := NewConfigService()
	module := NewConfigModule(config)

	err := module.Configure(container)
	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}

	if !container.Has("config") {
		t.Error("Config not registered in container")
	}

	result, err := container.Resolve("config")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != config {
		t.Error("Resolved config does not match")
	}
}

func TestConfigModule_GetControllers(t *testing.T) {
	module := NewConfigModule(nil)

	controllers := module.GetControllers()
	if len(controllers) != 0 {
		t.Errorf("GetControllers() length = %d, want 0", len(controllers))
	}
}

func TestConfigModule_GetProviders(t *testing.T) {
	config := NewConfigService()
	module := NewConfigModule(config)

	providers := module.GetProviders()
	if len(providers) != 1 {
		t.Errorf("GetProviders() length = %d, want 1", len(providers))
	}
}

func TestConfigModule_GetImports(t *testing.T) {
	module := NewConfigModule(nil)

	imports := module.GetImports()
	if len(imports) != 0 {
		t.Errorf("GetImports() length = %d, want 0", len(imports))
	}
}

func TestConfigModule_GetExports(t *testing.T) {
	module := NewConfigModule(nil)

	exports := module.GetExports()
	if len(exports) != 1 {
		t.Errorf("GetExports() length = %d, want 1", len(exports))
	}

	if exports[0] != "config" {
		t.Errorf("GetExports()[0] = %v, want 'config'", exports[0])
	}
}

