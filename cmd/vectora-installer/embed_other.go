//go:build linux || darwin

package main

func getInstallerAssets() map[string][]byte {
	// Fallback for other systems if binaries are not available
	return map[string][]byte{}
}
