//go:build !windows

package infra

import "fmt"

// NotifyOS prints to stdout on non-Windows systems.
func NotifyOS(title, message string) error {
	fmt.Printf("[Vectora] %s: %s\n", title, message)
	return nil
}
