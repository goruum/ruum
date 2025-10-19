package core

import (
	"fmt"
	"time"
)

// SimpleLogger is a basic logger implementation
type SimpleLogger struct{}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger() Logger {
	return &SimpleLogger{}
}

func (l *SimpleLogger) Log(level string, message string, fields map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s %s", level, timestamp, message)
	if len(fields) > 0 {
		fmt.Printf(" %v", fields)
	}
	fmt.Println()
}

func (l *SimpleLogger) Debug(message string, fields map[string]interface{}) {
	l.Log("DEBUG", message, fields)
}

func (l *SimpleLogger) Info(message string, fields map[string]interface{}) {
	l.Log("INFO", message, fields)
}

func (l *SimpleLogger) Warn(message string, fields map[string]interface{}) {
	l.Log("WARN", message, fields)
}

func (l *SimpleLogger) Error(message string, fields map[string]interface{}) {
	l.Log("ERROR", message, fields)
}

