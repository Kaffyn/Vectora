//go:build linux || darwin

package singleton

import (
	"os"
	"strconv"
	"syscall"
)

// TryLock acquires the singleton lock using a two-layer strategy:
//  1. syscall.Flock (LOCK_EX | LOCK_NB) — atomic, kernel-enforced.
//     Automatically released if the process dies. No stale lock risk.
//  2. PID file — written after flock succeeds, for human diagnostics
//     and cross-process status checks (e.g. `vectora status`).
func (i *Instance) TryLock() error {
	f, err := os.OpenFile(i.lockFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	// LOCK_EX | LOCK_NB: exclusive, non-blocking.
	// Returns EWOULDBLOCK immediately if another process holds the lock.
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return ErrAlreadyRunning
	}

	// Lock acquired — keep the file descriptor open (releasing fd releases flock).
	i.platformState.lockFd = f

	// Write PID for diagnostics (e.g. `vectora status`)
	if err := f.Truncate(0); err == nil {
		f.Seek(0, 0)
		f.WriteString(strconv.Itoa(os.Getpid()))
	}

	return nil
}

// Unlock releases the flock and removes the lock file.
func (i *Instance) Unlock() error {
	if i.platformState.lockFd != nil {
		syscall.Flock(int(i.platformState.lockFd.Fd()), syscall.LOCK_UN)
		i.platformState.lockFd.Close()
		i.platformState.lockFd = nil
	}
	return os.Remove(i.lockFile)
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Signal(0) on Unix: no signal sent, just checks if process exists.
	return process.Signal(syscall.Signal(0)) == nil
}
