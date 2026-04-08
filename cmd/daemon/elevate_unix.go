//go:build !windows

package main

import (
	"fmt"
	"os"
)

func elevateAdmin() error {
	fmt.Fprintf(os.Stderr, "Warning: Running without admin privileges. Some features may not work correctly.\n")
	return nil
}
