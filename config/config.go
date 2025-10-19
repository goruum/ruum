package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/goruum/ruum/core"
)

// DefaultConfigService implements core.ConfigService
type DefaultConfigService struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewConfigService creates a new config service
func NewConfigService() *DefaultConfigService {
	return &DefaultConfigService{
		data: make(map[string]interface{}),
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *DefaultConfigService) LoadFromEnv(prefix string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for range os.Environ() {
		// Parse environment variable
		// In a real implementation, you'd parse key=value pairs
		// and filter by prefix
	}

	return nil
}

// LoadFromFile loads configuration from a JSON file
func (c *DefaultConfigService) LoadFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for key, value := range config {
		c.data[key] = value
	}

	return nil
}

func (c *DefaultConfigService) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.data[key]
	if !exists {
		// Try environment variable
		if envValue := os.Getenv(key); envValue != "" {
			return envValue
		}
		return nil
	}

	return value
}

func (c *DefaultConfigService) GetString(key string) string {
	value := c.Get(key)
	if value == nil {
		return ""
	}

	if str, ok := value.(string); ok {
		return str
	}

	return fmt.Sprintf("%v", value)
}

func (c *DefaultConfigService) GetInt(key string) int {
	value := c.Get(key)
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}

	return 0
}

func (c *DefaultConfigService) GetBool(key string) bool {
	value := c.Get(key)
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes"
	case int:
		return v != 0
	}

	return false
}

func (c *DefaultConfigService) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = value
}

func (c *DefaultConfigService) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.data[key]
	if exists {
		return true
	}

	// Check environment variable
	return os.Getenv(key) != ""
}

// ConfigModule provides configuration as a module
type ConfigModule struct {
	config core.ConfigService
}

// NewConfigModule creates a new config module
func NewConfigModule(config core.ConfigService) *ConfigModule {
	if config == nil {
		config = NewConfigService()
	}
	return &ConfigModule{
		config: config,
	}
}

func (m *ConfigModule) Configure(container core.Container) error {
	return container.RegisterValue("config", m.config)
}

func (m *ConfigModule) GetControllers() []interface{} {
	return []interface{}{}
}

func (m *ConfigModule) GetProviders() []interface{} {
	return []interface{}{m.config}
}

func (m *ConfigModule) GetImports() []core.Module {
	return []core.Module{}
}

func (m *ConfigModule) GetExports() []string {
	return []string{"config"}
}
