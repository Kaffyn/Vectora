//go:build linux || darwin

package singleton

import "os"

type platformState struct {
	lockFd *os.File
}

func newPlatformState() platformState {
	return platformState{}
}
