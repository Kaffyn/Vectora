//go:build windows

package infra

import (
	"os/exec"
	"syscall"
)

// HideWindow attaches a SysProcAttr to the command that prevents a terminal window from popping up on Windows.
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= 0x08000000 // CREATE_NO_WINDOW
}
