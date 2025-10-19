package logger

import (
	"testing"

	"github.com/goruum/ruum/core"
)

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger()
	if logger == nil {
		t.Fatal("NewDefaultLogger() returned nil")
	}
}

func TestDefaultLogger_SetLevel(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetLevel(LevelError)

	// Level should be set (we can't easily test output suppression)
	if logger.level != LevelError {
		t.Errorf("Level = %v, want %v", logger.level, LevelError)
	}
}

func TestDefaultLogger_SetColorize(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetColorize(false)

	if logger.colorize != false {
		t.Error("Colorize should be false")
	}

	logger.SetColorize(true)

	if logger.colorize != true {
		t.Error("Colorize should be true")
	}
}

func TestDefaultLogger_Log(t *testing.T) {
	logger := NewDefaultLogger()

	// Test that logging doesn't panic
	logger.Log("INFO", "test message", nil)
	logger.Log("DEBUG", "debug message", map[string]interface{}{
		"key": "value",
	})
}

func TestDefaultLogger_Debug(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetLevel(LevelDebug)

	// Should not panic
	logger.Debug("debug message", nil)
	logger.Debug("debug with fields", map[string]interface{}{
		"field1": "value1",
	})
}

func TestDefaultLogger_Info(t *testing.T) {
	logger := NewDefaultLogger()

	// Should not panic
	logger.Info("info message", nil)
	logger.Info("info with fields", map[string]interface{}{
		"field1": "value1",
	})
}

func TestDefaultLogger_Warn(t *testing.T) {
	logger := NewDefaultLogger()

	// Should not panic
	logger.Warn("warning message", nil)
	logger.Warn("warning with fields", map[string]interface{}{
		"field1": "value1",
	})
}

func TestDefaultLogger_Error(t *testing.T) {
	logger := NewDefaultLogger()

	// Should not panic
	logger.Error("error message", nil)
	logger.Error("error with fields", map[string]interface{}{
		"field1": "value1",
		"error":  "something went wrong",
	})
}

func TestDefaultLogger_ShouldLog(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetLevel(LevelWarn)

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{LevelDebug, false},
		{LevelInfo, false},
		{LevelWarn, true},
		{LevelError, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := logger.shouldLog(tt.level)
			if result != tt.expected {
				t.Errorf("shouldLog(%v) = %v, want %v", tt.level, result, tt.expected)
			}
		})
	}
}

func TestDefaultLogger_GetColor(t *testing.T) {
	logger := NewDefaultLogger()

	tests := []struct {
		level LogLevel
	}{
		{LevelDebug},
		{LevelInfo},
		{LevelWarn},
		{LevelError},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			color := logger.getColor(tt.level)
			if color == "" {
				t.Errorf("getColor(%v) returned empty string", tt.level)
			}
		})
	}
}

func TestDefaultLogger_FormatFields(t *testing.T) {
	logger := NewDefaultLogger()

	tests := []struct {
		name   string
		fields map[string]interface{}
		empty  bool
	}{
		{
			name:   "nil fields",
			fields: nil,
			empty:  true,
		},
		{
			name:   "empty fields",
			fields: map[string]interface{}{},
			empty:  true,
		},
		{
			name: "with fields",
			fields: map[string]interface{}{
				"key": "value",
			},
			empty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := logger.formatFields(tt.fields)
			if tt.empty && result != "" {
				t.Errorf("formatFields() = %v, want empty string", result)
			}
			if !tt.empty && result == "" {
				t.Error("formatFields() returned empty string for non-empty fields")
			}
		})
	}
}

func TestNewLoggerModule(t *testing.T) {
	logger := NewDefaultLogger()
	module := NewLoggerModule(logger)

	if module == nil {
		t.Fatal("NewLoggerModule() returned nil")
	}

	if module.logger != logger {
		t.Error("Logger not set correctly")
	}
}

func TestNewLoggerModule_NilLogger(t *testing.T) {
	module := NewLoggerModule(nil)

	if module == nil {
		t.Fatal("NewLoggerModule() returned nil")
	}

	if module.logger == nil {
		t.Error("Logger should be created automatically")
	}
}

func TestLoggerModule_Configure(t *testing.T) {
	container := core.NewContainer()
	logger := NewDefaultLogger()
	module := NewLoggerModule(logger)

	err := module.Configure(container)
	if err != nil {
		t.Errorf("Configure() error = %v", err)
	}

	if !container.Has("logger") {
		t.Error("Logger not registered in container")
	}

	result, err := container.Resolve("logger")
	if err != nil {
		t.Errorf("Resolve() error = %v", err)
	}

	if result != logger {
		t.Error("Resolved logger does not match")
	}
}

func TestLoggerModule_GetControllers(t *testing.T) {
	module := NewLoggerModule(nil)

	controllers := module.GetControllers()
	if len(controllers) != 0 {
		t.Errorf("GetControllers() length = %d, want 0", len(controllers))
	}
}

func TestLoggerModule_GetProviders(t *testing.T) {
	logger := NewDefaultLogger()
	module := NewLoggerModule(logger)

	providers := module.GetProviders()
	if len(providers) != 1 {
		t.Errorf("GetProviders() length = %d, want 1", len(providers))
	}
}

func TestLoggerModule_GetImports(t *testing.T) {
	module := NewLoggerModule(nil)

	imports := module.GetImports()
	if len(imports) != 0 {
		t.Errorf("GetImports() length = %d, want 0", len(imports))
	}
}

func TestLoggerModule_GetExports(t *testing.T) {
	module := NewLoggerModule(nil)

	exports := module.GetExports()
	if len(exports) != 1 {
		t.Errorf("GetExports() length = %d, want 1", len(exports))
	}

	if exports[0] != "logger" {
		t.Errorf("GetExports()[0] = %v, want 'logger'", exports[0])
	}
}

func TestDefaultLogger_ColorizedOutput(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetColorize(true)

	// Should not panic with colorized output
	logger.Debug("colorized debug", nil)
	logger.Info("colorized info", nil)
	logger.Warn("colorized warning", nil)
	logger.Error("colorized error", nil)
}

func TestDefaultLogger_NonColorizedOutput(t *testing.T) {
	logger := NewDefaultLogger()
	logger.SetColorize(false)

	// Should not panic without colorized output
	logger.Debug("plain debug", nil)
	logger.Info("plain info", nil)
	logger.Warn("plain warning", nil)
	logger.Error("plain error", nil)
}

