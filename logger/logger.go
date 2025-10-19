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

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
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

func (l *DefaultLogger) Log(level string, message string, fields map[string]interface{}) {
	logLevel := LogLevel(level)
	if !l.shouldLog(logLevel) {
		return
	}

	timestamp := time.Now().Format(l.timeFormat)

	if l.colorize {
		color := l.getColor(logLevel)
		reset := "\033[0m"
		fmt.Fprintf(os.Stdout, "%s[%s]%s %s %s %s\n",
			color, logLevel, reset, timestamp, message, l.formatFields(fields))
	} else {
		fmt.Fprintf(os.Stdout, "[%s] %s %s %s\n",
			logLevel, timestamp, message, l.formatFields(fields))
	}
}

func (l *DefaultLogger) Debug(message string, fields map[string]interface{}) {
	l.Log(string(LevelDebug), message, fields)
}

func (l *DefaultLogger) Info(message string, fields map[string]interface{}) {
	l.Log(string(LevelInfo), message, fields)
}

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

// LoggerModule provides logger as a module
type LoggerModule struct {
	logger core.Logger
}

// NewLoggerModule creates a new logger module
func NewLoggerModule(logger core.Logger) *LoggerModule {
	if logger == nil {
		logger = NewDefaultLogger()
	}
	return &LoggerModule{
		logger: logger,
	}
}

func (m *LoggerModule) Configure(container core.Container) error {
	return container.RegisterValue("logger", m.logger)
}

func (m *LoggerModule) GetControllers() []interface{} {
	return []interface{}{}
}

func (m *LoggerModule) GetProviders() []interface{} {
	return []interface{}{m.logger}
}

func (m *LoggerModule) GetImports() []core.Module {
	return []core.Module{}
}

func (m *LoggerModule) GetExports() []string {
	return []string{"logger"}
}
