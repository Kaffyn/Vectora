//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawnDetached launches the vectora process as a fully detached background process
// using setsid to create a new session, surviving terminal close.
func spawnDetached() error {
	exe, _ := os.Executable()
	args := []string{"start", "--detached"}
	if startPort != 42780 {
		args = append(args, "--port", fmt.Sprintf("%d", startPort))
	}

	cmd := exec.Command(exe, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}
