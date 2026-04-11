package singleton

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

var (
	ErrAlreadyRunning = errors.New("another instance is already running")
)

// Instance manages the singleton lock for the application using
// a hybrid strategy: PID file + OS-level file lock (flock on Unix,
// CreateFile exclusive on Windows). This handles stale locks from
// crashes and prevents race conditions on simultaneous startup.
//
// platformState is an embedded struct defined per-platform that holds
// the OS-specific lock handle (lockFd on Unix, lockHandle on Windows).
type Instance struct {
	name     string
	lockFile string
	platformState
}

// New creates a new singleton manager.
// Windows: lock in %APPDATA%\Vectora\.lock  (matches GetAppDataDir)
// Unix:    lock in ~/.Vectora/.lock
func New(appName string) *Instance {
	appDir := appDataDir()
	_ = os.MkdirAll(appDir, 0755)
	return &Instance{
		name:          appName,
		lockFile:      filepath.Join(appDir, ".lock"),
		platformState: newPlatformState(),
	}
}

// LockFilePath returns the lock file path for diagnostics/status commands.
func (i *Instance) LockFilePath() string {
	return i.lockFile
}

// readPID reads the PID from an existing lock file.
func readPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}
