//go:build !windows

package infra

import (
	"os/exec"
)

// HideWindow is a no-op on platforms other than Windows.
func HideWindow(cmd *exec.Cmd) {
	// No terminal windows pop up by default on Unix-like systems for background exec.
}
