//go:build !windows

package tray

import "github.com/Kaffyn/Vectora/core/llm"

var ActiveProvider llm.Provider

func Setup() {
	// Tray available only on Windows
}
