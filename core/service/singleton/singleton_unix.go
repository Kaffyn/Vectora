//go:build linux || darwin

package singleton

import (
	"os"
	"strconv"
	"syscall"
)

var lockFd *os.File

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
		if err == syscall.EWOULDBLOCK {
			// Validate the PID in the file as a diagnostic fallback
			if pid, readErr := readPID(i.lockFile); readErr == nil && isProcessAlive(pid) {
				return ErrAlreadyRunning
			}
			// flock failed but PID is dead — another process crashed while holding flock.
			// This can happen if the OS releases flock but we still see EWOULDBLOCK briefly.
			return ErrAlreadyRunning
		}
		return err
	}

	// Lock acquired — keep the file descriptor open (releasing fd releases flock).
	lockFd = f

	// Write PID for diagnostics (e.g. `vectora status`)
	if err := f.Truncate(0); err == nil {
		f.Seek(0, 0)
		f.WriteString(intToStr(os.Getpid()))
	}

	return nil
}

// Unlock releases the flock and removes the lock file.
func (i *Instance) Unlock() error {
	if lockFd != nil {
		syscall.Flock(int(lockFd.Fd()), syscall.LOCK_UN)
		lockFd.Close()
		lockFd = nil
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

func intToStr(n int) string {
	return strconv.Itoa(n)
}
