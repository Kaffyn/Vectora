//go:build !windows

package tray

import "vectora/core/llm"

// Setup é no-op em Linux/macOS.
func Setup(router *llm.Router) {
	// Tray só disponível no Windows
}
