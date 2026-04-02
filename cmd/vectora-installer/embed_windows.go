//go:build windows

package main

import _ "embed"

//go:embed vectora.exe
var vectoraExe []byte

//go:embed llama/llama-server.exe
var llamaServerExe []byte

//go:embed llama/rpc-server.exe
var rpcServerExe []byte

//go:embed llama/llama-cli.exe
var llamaCliExe []byte

//go:embed llama/ggml-vulkan.dll
var vulkanDll []byte

func getInstallerAssets() map[string][]byte {
	return map[string][]byte{
		"vectora.exe":      vectoraExe,
		"llama-server.exe": llamaServerExe,
		"rpc-server.exe":   rpcServerExe,
		"llama-cli.exe":    llamaCliExe,
		"ggml-vulkan.dll":  vulkanDll,
	}
}
