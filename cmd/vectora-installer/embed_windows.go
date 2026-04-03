//go:build windows

package main

import _ "embed"

//go:embed vectora.exe
var vectoraExe []byte

//go:embed llama-installer.exe
var llamaInstallerExe []byte

func getInstallerAssets() map[string][]byte {
	return map[string][]byte{
		"vectora.exe":         vectoraExe,
		"llama-installer.exe": llamaInstallerExe,
	}
}
