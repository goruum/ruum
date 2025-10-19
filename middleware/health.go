package middleware

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/goruum/ruum/core"
)

// HealthCheck holds health check data
type HealthCheck struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    float64                `json:"uptime"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Path       string
	LivePath   string
	ReadyPath  string
	Checks     map[string]func() bool
	ShowSystem bool
}

var startTime = time.Now()

// DefaultHealthCheckConfig returns default configuration
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Path:       "/health",
		LivePath:   "/health/live",
		ReadyPath:  "/health/ready",
		Checks:     make(map[string]func() bool),
		ShowSystem: true,
	}
}

// HealthCheckMiddleware creates a health check middleware
func HealthCheckMiddleware(config HealthCheckConfig) core.MiddlewareFunc {
	return func(next core.HandlerFunc) core.HandlerFunc {
		return func(ctx core.Context) error {
			path := ctx.Path()

			// Liveness probe
			if path == config.LivePath {
				return handleLiveness(ctx, config)
			}

			// Readiness probe
			if path == config.ReadyPath {
				return handleReadiness(ctx, config)
			}

			// General health check
			if path == config.Path {
				return handleHealth(ctx, config)
			}

			return next(ctx)
		}
	}
}

func handleLiveness(ctx core.Context, _ HealthCheckConfig) error {
	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(http.StatusOK)
	return json.NewEncoder(ctx.Response()).Encode(map[string]interface{}{
		"status":    "alive",
		"timestamp": time.Now(),
	})
}

func handleReadiness(ctx core.Context, config HealthCheckConfig) error {
	checks := make(map[string]interface{})
	allHealthy := true

	for name, check := range config.Checks {
		healthy := check()
		checks[name] = map[string]interface{}{
			"status": map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
		}
		if !healthy {
			allHealthy = false
		}
	}

	status := http.StatusOK
	statusStr := "ready"
	if !allHealthy {
		status = http.StatusServiceUnavailable
		statusStr = "not ready"
	}

	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(status)
	return json.NewEncoder(ctx.Response()).Encode(map[string]interface{}{
		"status":    statusStr,
		"timestamp": time.Now(),
		"checks":    checks,
	})
}

func handleHealth(ctx core.Context, config HealthCheckConfig) error {
	health := HealthCheck{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).Seconds(),
		Checks:    make(map[string]interface{}),
	}

	// Run custom checks
	allHealthy := true
	for name, check := range config.Checks {
		healthy := check()
		health.Checks[name] = map[string]interface{}{
			"status": map[bool]string{true: "healthy", false: "unhealthy"}[healthy],
		}
		if !healthy {
			allHealthy = false
		}
	}

	if !allHealthy {
		health.Status = "unhealthy"
	}

	// Add system info if enabled
	if config.ShowSystem {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		health.Checks["system"] = map[string]interface{}{
			"goroutines": runtime.NumGoroutine(),
			"memory": map[string]interface{}{
				"alloc":      mem.Alloc,
				"totalAlloc": mem.TotalAlloc,
				"sys":        mem.Sys,
				"numGC":      mem.NumGC,
			},
		}
	}

	status := http.StatusOK
	if !allHealthy {
		status = http.StatusServiceUnavailable
	}

	ctx.Response().Header().Set("Content-Type", "application/json")
	ctx.Response().WriteHeader(status)
	return json.NewEncoder(ctx.Response()).Encode(health)
}
