//go:build windows

package singleton

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

var (
	modkernel32         = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess     = modkernel32.NewProc("OpenProcess")
	procGetExitCodeProc = modkernel32.NewProc("GetExitCodeProcess")
)

const (
	processQueryLimitedInfo = 0x1000
	stillActive             = 259
)

// TryLock acquires the singleton lock using a two-layer strategy:
//  1. CreateFile with exclusive share flags — kernel-enforced, released on process exit.
//     Automatically released if the process dies. No stale lock risk.
//  2. PID file — written after lock succeeds, for human diagnostics
//     and cross-process status checks (e.g. `vectora status`).
func (i *Instance) TryLock() error {
	lockPath, _ := syscall.UTF16PtrFromString(i.lockFile)

	// Open/create lock file with no sharing (exclusive).
	// No sharing mode — this is what enforces exclusivity.
	h, err := syscall.CreateFile(
		lockPath,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0, // no sharing
		nil,
		syscall.OPEN_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		// ERROR_SHARING_VIOLATION (32) means another process holds the file open.
		return ErrAlreadyRunning
	}

	i.platformState.lockHandle = h

	// Write PID for diagnostics
	pidBytes := []byte(strconv.Itoa(os.Getpid()))
	var written uint32
	_ = syscall.WriteFile(h, pidBytes, &written, nil)

	return nil
}

// Unlock closes the exclusive file handle and removes the lock file.
func (i *Instance) Unlock() error {
	if i.platformState.lockHandle != syscall.InvalidHandle {
		_ = syscall.CloseHandle(i.platformState.lockHandle)
		i.platformState.lockHandle = syscall.InvalidHandle
	}
	return os.Remove(i.lockFile)
}

// isProcessAlive checks if a process with the given PID is alive using
// OpenProcess + GetExitCodeProcess (reliable on Windows; Signal(0) is not).
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	h, _, _ := procOpenProcess.Call(
		uintptr(processQueryLimitedInfo),
		0,
		uintptr(pid),
	)
	if h == 0 {
		return false
	}
	defer func() { _ = syscall.CloseHandle(syscall.Handle(h)) }()

	var exitCode uint32
	ret, _, _ := procGetExitCodeProc.Call(h, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false
	}

	return exitCode == stillActive
}
