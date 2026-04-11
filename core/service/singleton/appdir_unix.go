//go:build linux || darwin

package singleton

import (
	"os"
	"path/filepath"
)

// appDataDir returns ~/.Vectora on Unix.
func appDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".Vectora")
}
