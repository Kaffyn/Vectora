//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawnDetached launches the vectora process as a fully detached background process
// that survives terminal close. No console window is created.
func spawnDetached() error {
	exe, _ := os.Executable()
	args := []string{"start", "--detached"}
	if startPort != 42780 {
		args = append(args, "--port", fmt.Sprintf("%d", startPort))
	}

	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		// DETACHED_PROCESS (0x8) | CREATE_NEW_PROCESS_GROUP (0x200)
		// This detaches from the parent console entirely.
		CreationFlags: 0x00000008 | 0x00000200,
	}
	return cmd.Start()
}
