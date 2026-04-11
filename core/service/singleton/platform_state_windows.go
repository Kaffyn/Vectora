//go:build windows

package singleton

import "syscall"

type platformState struct {
	lockHandle syscall.Handle
}

func newPlatformState() platformState {
	return platformState{lockHandle: syscall.InvalidHandle}
}
