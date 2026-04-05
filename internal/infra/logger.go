package infra

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	globalLogger *slog.Logger
	loggerMutex  sync.Mutex
)

// InitLogger initializes the global logger with file and console output
func InitLogger(logPath string, level string) error {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	logLevel := slog.LevelInfo
	switch level {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	}

	// Create log directory
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer for both stderr and file
	multiWriter := io.MultiWriter(os.Stderr, file)

	// JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	handler := slog.NewJSONHandler(multiWriter, opts)

	globalLogger = slog.New(handler)
	return nil
}

// Logger returns the global logger instance
func Logger() *slog.Logger {
	loggerMutex.Lock()
	defer loggerMutex.Unlock()

	if globalLogger == nil {
		return slog.Default()
	}
	return globalLogger
}

// LogWithContext logs with additional context attributes
func LogWithContext(logger *slog.Logger, msg string, attrs ...interface{}) {
	logger.Info(msg, attrs...)
}

// LogError logs an error with context
func LogError(logger *slog.Logger, msg string, err error, attrs ...interface{}) {
	allAttrs := append([]interface{}{"error", err}, attrs...)
	logger.Error(msg, allAttrs...)
}
