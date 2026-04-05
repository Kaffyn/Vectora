package infra

import (
	"log/slog"
	"os"
	"path/filepath"
)

var Logger *slog.Logger

// SetupLogger initializes the global structured logger.
func SetupLogger() error {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(userProfile, ".Vectora", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	handler := slog.NewJSONHandler(logFile, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
	return nil
}
