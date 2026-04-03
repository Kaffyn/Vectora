//go:build windows

package main

import "embed"

//go:embed all:assets
var llamaAssets embed.FS

func getLlamaAssets() map[string][]byte {
	assets := make(map[string][]byte)
	
	// List of vital files for the engine to function on Windows
	files := []string{
		"llama-cli.exe",
		"llama-server.exe",
		"rpc-server.exe",
		"ggml-vulkan.dll",
		"llama.dll",
		"ggml.dll",
		"ggml-base.dll",
		"ggml-rpc.dll",
	}

	for _, name := range files {
		data, err := llamaAssets.ReadFile("assets/" + name)
		if err == nil {
			assets[name] = data
		}
	}

	return assets
}
