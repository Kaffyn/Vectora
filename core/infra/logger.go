package infra

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/core/telemetry"
)

// SetupLogger initializes the global structured logger with rotating file output.
// Delegates to core/telemetry for RotatingWriter (10MB rotation, 1 backup).
func SetupLogger() error {
	userProfile, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logDir := filepath.Join(userProfile, ".Vectora", "logs")
	return telemetry.InitLogger(logDir)
}

// Logger returns the global structured logger.
// After SetupLogger() is called, this returns the telemetry.GlobalLogger.
// Before SetupLogger(), it returns a fallback stderr logger.
func Logger() *slog.Logger {
	return telemetry.GetLogger()
}
