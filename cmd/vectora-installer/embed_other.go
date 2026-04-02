//go:build linux || darwin

package main

func getInstallerAssets() map[string][]byte {
	// Fallback para outros sistemas se os binários não estiverem disponíveis
	return map[string][]byte{}
}
