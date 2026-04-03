package engines

import (
	"os"
	"path/filepath"
	"runtime"
)

// GetEnginesDir resolves the absolute path for storing inference binaries (backends).
// No Windows: %LOCALAPPDATA%\Vectora\engines
// No Unix: ~/.local/share/vectora/engines
func GetEnginesDir() (string, error) {
	var base string
	var err error

	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		base = filepath.Join(base, "Vectora", "engines")
	case "darwin":
		base, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(base, "Library", "Application Support", "Vectora", "engines")
	default: // Linux e outros
		base, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(base, ".local", "share", "vectora", "engines")
	}

	// Ensures the base folder exists
	if err := os.MkdirAll(base, 0755); err != nil {
		return "", err
	}

	return base, nil
}

// GetLlamaPath returns the specific path for the Llama-CPP engine.
func GetLlamaPath() (string, error) {
	base, err := GetEnginesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "llama"), nil
}
