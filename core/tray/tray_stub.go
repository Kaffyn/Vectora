//go:build !windows

package tray

import "github.com/Kaffyn/Vectora/core/llm"

var ActiveProvider llm.Provider
var ActiveProviderID string

func Setup() {
	// Tray available only on Windows for now (PE embedded resources)
}

func ReloadActiveProvider() {
	// No-op on non-windows platforms where the tray daemon isn't the primary config consumer
}
