//go:build !windows

package tray

import (
	"github.com/Kaffyn/Vectora/internal/llm"
)

// ActiveProvider is a stub for non-Windows platforms
var ActiveProvider llm.Provider

// Setup is a no-op on non-Windows platforms
func Setup() {
	// System tray is only available on Windows
	// On Linux/macOS, the daemon runs without a tray icon
}
