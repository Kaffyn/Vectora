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

	lockHandle syscall.Handle = syscall.InvalidHandle
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
	// FILE_FLAG_DELETE_ON_CLOSE ensures cleanup even on crash.
	h, err := syscall.CreateFile(
		lockPath,
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0, // no sharing — this is what enforces exclusivity
		nil,
		syscall.OPEN_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		// ERROR_SHARING_VIOLATION (32) means another process holds the file open.
		if errno, ok := err.(syscall.Errno); ok && errno == 32 {
			// Double-check: read PID and verify the process is actually alive.
			if pid, readErr := readPID(i.lockFile); readErr == nil && isProcessAlive(pid) {
				return ErrAlreadyRunning
			}
			// Process is dead but lock file is still held (e.g. zombie handle).
			return ErrAlreadyRunning
		}
		return err
	}

	lockHandle = h

	// Write PID for diagnostics
	pid := strconv.Itoa(os.Getpid())
	pidBytes := []byte(pid)
	var written uint32
	_ = syscall.WriteFile(h, pidBytes, &written, nil)

	return nil
}

// Unlock closes the exclusive file handle and removes the lock file.
func (i *Instance) Unlock() error {
	if lockHandle != syscall.InvalidHandle {
		_ = syscall.CloseHandle(lockHandle)
		lockHandle = syscall.InvalidHandle
	}
	return os.Remove(i.lockFile)
}

// isProcessAlive checks if a process with the given PID is alive using
// OpenProcess + GetExitCodeProcess (reliable on Windows; Signal(0) is not).
func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	h, _, err := procOpenProcess.Call(
		uintptr(processQueryLimitedInfo),
		0,
		uintptr(pid),
	)
	if h == 0 {
		// Process doesn't exist or access denied — assume dead.
		_ = err
		return false
	}
	defer func() {
		_ = syscall.CloseHandle(syscall.Handle(h))
	}()

	var exitCode uint32
	ret, _, _ := procGetExitCodeProc.Call(h, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		return false
	}

	return exitCode == stillActive
}
