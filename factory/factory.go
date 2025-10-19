package factory

import (
	"github.com/goruum/ruum/core"
)

// CreateApplication creates a new Ruum application
func CreateApplication(rootModule core.Module, opts ...ApplicationOption) (core.Application, error) {
	config := core.ApplicationConfig{
		GlobalPrefix:        "",
		CorsEnabled:         false,
		CorsOrigins:         []string{"*"},
		Logger:              nil,
		ShutdownTimeout:     0,
		EnableShutdownHooks: true,
	}

	for _, opt := range opts {
		opt(&config)
	}

	factory := core.NewApplicationFactory()
	return factory.Create(rootModule, config)
}

// ApplicationOption configures the application
type ApplicationOption func(*core.ApplicationConfig)

// WithGlobalPrefix sets the global route prefix
func WithGlobalPrefix(prefix string) ApplicationOption {
	return func(c *core.ApplicationConfig) {
		c.GlobalPrefix = prefix
	}
}

// WithCORS enables CORS with specified origins
func WithCORS(origins ...string) ApplicationOption {
	return func(c *core.ApplicationConfig) {
		c.CorsEnabled = true
		if len(origins) > 0 {
			c.CorsOrigins = origins
		}
	}
}

// WithLogger sets a custom logger
func WithLogger(logger core.Logger) ApplicationOption {
	return func(c *core.ApplicationConfig) {
		c.Logger = logger
	}
}

// WithShutdownHooks enables graceful shutdown hooks
func WithShutdownHooks(enabled bool) ApplicationOption {
	return func(c *core.ApplicationConfig) {
		c.EnableShutdownHooks = enabled
	}
}
