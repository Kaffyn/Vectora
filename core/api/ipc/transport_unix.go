//go:build !windows

package ipc

import (
	"net"
	"os"
	"path/filepath"
)

func listenIPC(addr string) (net.Listener, error) {
	_ = os.MkdirAll(filepath.Dir(addr), 0700)
	_ = os.Remove(addr)
	l, err := net.Listen("unix", addr)
	if err == nil {
		// Restrict socket to owner only
		_ = os.Chmod(addr, 0600)
	}
	return l, err
}

func dialIPC(addr string) (net.Conn, error) {
	return net.Dial("unix", addr)
}
