//go:build !windows

package main

import (
	"fmt"
	"os"
)

func elevateAdmin() error {
	// On Unix systems, we rely on the user running with sudo if needed
	// No automatic elevation
	fmt.Fprintf(os.Stderr, "Warning: Running without admin privileges. Some features may not work correctly.\n")
	return nil
}
