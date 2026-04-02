package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

// Logger is a structured logger for Vectora
type Logger struct {
	*slog.Logger
}

func New(logDir string, debug bool) *Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}

	// Console Handler
	consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})

	// File Handler (Optional)
	if logDir != "" {
		_ = os.MkdirAll(logDir, 0755)
		logFile, err := os.OpenFile(filepath.Join(logDir, "vectora.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			// Multi-handler logic could be added here
			_ = logFile
		}
	}

	return &Logger{slog.New(consoleHandler)}
}
