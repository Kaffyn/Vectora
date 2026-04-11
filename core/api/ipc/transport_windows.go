//go:build windows

package ipc

import (
	"log"
	"net"

	winio "github.com/Microsoft/go-winio"
)

func listenIPC(addr string) (net.Listener, error) {
	// Use go-winio for proper Windows named pipe support.
	// Security: restrict access to current user only.
	cfg := &winio.PipeConfig{SecurityDescriptor: "D:P(A;;GA;;;CU)"}
	l, err := winio.ListenPipe(addr, cfg)
	if err != nil {
		// Fallback to TCP if named pipe fails (e.g. permissions).
		log.Printf("IPC: named pipe unavailable (%v), falling back to TCP", err)
		l, err = net.Listen("tcp", "127.0.0.1:42781")
	}
	return l, err
}

func dialIPC(addr string) (net.Conn, error) {
	// Use go-winio for proper named pipe dial.
	conn, err := winio.DialPipe(addr, nil)
	if err != nil {
		// Fallback to TCP (matches server fallback).
		conn, err = net.Dial("tcp", "127.0.0.1:42781")
	}
	return conn, err
}
