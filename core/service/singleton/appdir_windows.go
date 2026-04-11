//go:build windows

package singleton

import (
	"os"
	"path/filepath"
)

// appDataDir returns %APPDATA%\Vectora on Windows —
// consistent with WindowsManager.GetAppDataDir().
func appDataDir() string {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "AppData", "Roaming", "Vectora")
	}
	return filepath.Join(appData, "Vectora")
}
