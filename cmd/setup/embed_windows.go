//go:build windows

package main

import _ "embed"

//go:embed vectora.exe
var vectoraExe []byte

//go:embed vectora-cli.exe
var vectoraCliExe []byte

//go:embed lpm.exe
var lpmExe []byte

//go:embed mpm.exe
var mpmExe []byte

func getInstallerAssets() map[string][]byte {
	return map[string][]byte{
		"vectora.exe":    vectoraExe,
		"vectora-cli.exe": vectoraCliExe,
		"lpm.exe":        lpmExe,
		"mpm.exe":        mpmExe,
	}
}
