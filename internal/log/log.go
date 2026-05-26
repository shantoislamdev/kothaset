// Package log provides structured logging for KothaSet using log/slog.
package log

import (
	"log/slog"
	"os"
)

var level = new(slog.LevelVar)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))

// SetLevel sets the minimum log level.
func SetLevel(l slog.Level) {
	level.Set(l)
}

// Debug logs a debug message.
func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Info logs an informational message.
func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

// Warn logs a warning message.
func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs an error message.
func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}
