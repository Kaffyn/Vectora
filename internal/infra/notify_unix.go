//go:build !windows

package infra

import (
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/assets"
	"github.com/gen2brain/beeep"
)

func NotifyOS(title, message string) error {
	iconPath := filepath.Join(os.TempDir(), "vectora_logo.png")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		os.WriteFile(iconPath, assets.IconData, 0644)
	}
	return beeep.Notify(title, message, iconPath)
}
