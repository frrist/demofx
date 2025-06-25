package shared

import (
	"fmt"
	"strings"
	"time"
)

// Logger provides basic logging functionality
type Logger struct {
	level    string
	env      string
}

// NewLogger creates a new logger instance with config
func NewLogger(config *Config) *Logger {
	return &Logger{
		level:  config.App.LogLevel,
		env:    config.App.Environment,
	}
}

// Log outputs a timestamped message
func (l *Logger) Log(component, message string) {
	// In production, might suppress debug logs
	if l.env == "production" && l.level == "error" && !strings.Contains(message, "Error") {
		return
	}
	
	envTag := ""
	if l.env != "development" {
		envTag = fmt.Sprintf("[%s]", strings.ToUpper(l.env))
	}
	
	fmt.Printf("[%s]%s %s: %s\n", time.Now().Format("15:04:05"), envTag, component, message)
}