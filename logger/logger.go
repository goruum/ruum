// Package logger provides logging functionality for the application.
package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/goruum/ruum/core"
)

// LogLevel represents the severity of a log message
type LogLevel string

// Log levels
const (
	// LevelDebug is for debug messages
	LevelDebug LogLevel = "DEBUG"
	// LevelInfo is for informational messages
	LevelInfo  LogLevel = "INFO"
	// LevelWarn is for warning messages
	LevelWarn  LogLevel = "WARN"
	// LevelError is for error messages
	LevelError LogLevel = "ERROR"
)

// DefaultLogger implements core.Logger
type DefaultLogger struct {
	level      LogLevel
	timeFormat string
	colorize   bool
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		level:      LevelInfo,
		timeFormat: time.RFC3339,
		colorize:   true,
	}
}

// SetLevel sets the minimum log level
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

// SetColorize enables or disables colored output
func (l *DefaultLogger) SetColorize(colorize bool) {
	l.colorize = colorize
}

// Log logs a message at the specified level with optional fields.
func (l *DefaultLogger) Log(level string, message string, fields map[string]interface{}) {
	logLevel := LogLevel(level)
	if !l.shouldLog(logLevel) {
		return
	}

	timestamp := time.Now().Format(l.timeFormat)

	if l.colorize {
		color := l.getColor(logLevel)
		reset := "\033[0m"
		_, _ = fmt.Fprintf(os.Stdout, "%s[%s]%s %s %s %s\n",
			color, logLevel, reset, timestamp, message, l.formatFields(fields))
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "[%s] %s %s %s\n",
			logLevel, timestamp, message, l.formatFields(fields))
	}
}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(message string, fields map[string]interface{}) {
	l.Log(string(LevelDebug), message, fields)
}

// Info logs an info message.
func (l *DefaultLogger) Info(message string, fields map[string]interface{}) {
	l.Log(string(LevelInfo), message, fields)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(message string, fields map[string]interface{}) {
	l.Log(string(LevelWarn), message, fields)
}

func (l *DefaultLogger) Error(message string, fields map[string]interface{}) {
	l.Log(string(LevelError), message, fields)
}

func (l *DefaultLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	return levels[level] >= levels[l.level]
}

func (l *DefaultLogger) getColor(level LogLevel) string {
	colors := map[LogLevel]string{
		LevelDebug: "\033[36m", // Cyan
		LevelInfo:  "\033[32m", // Green
		LevelWarn:  "\033[33m", // Yellow
		LevelError: "\033[31m", // Red
	}

	return colors[level]
}

func (l *DefaultLogger) formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}

	data, err := json.Marshal(fields)
	if err != nil {
		return ""
	}

	return string(data)
}

// Module provides logging services as a module.
type Module struct {
	logger core.Logger
}

// NewLoggerModule creates a new logger module
func NewLoggerModule(logger core.Logger) *Module {
	if logger == nil {
		logger = NewDefaultLogger()
	}
	return &Module{
		logger: logger,
	}
}

// Configure registers the logger in the container.
func (m *Module) Configure(container core.Container) error {
	return container.RegisterValue("logger", m.logger)
}

// GetControllers returns the module's controllers.
func (m *Module) GetControllers() []interface{} {
	return []interface{}{}
}

// GetProviders returns the module's providers.
func (m *Module) GetProviders() []interface{} {
	return []interface{}{m.logger}
}

// GetImports returns the imported modules.
func (m *Module) GetImports() []core.Module {
	return []core.Module{}
}

// GetExports returns the exported provider names.
func (m *Module) GetExports() []string {
	return []string{"logger"}
}
