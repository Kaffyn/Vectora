//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func elevateAdmin() error {
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	cmd := exec.Command("powershell", fmt.Sprintf("Start-Process -FilePath '%s' -Verb runas -WorkingDirectory '%s'", exe, cwd))
	if args != "" {
		cmd = exec.Command("powershell", fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList '%s' -Verb runas -WorkingDirectory '%s'", exe, args, cwd))
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Start()
}
