//go:build darwin

package main

import _ "embed"

//go:embed ../../internal/os/macos/llama-b8583/llama-server
var llamaServerBin []byte

//go:embed ../../internal/os/macos/llama-b8583/llama-cli
var llamaCliBin []byte

func getLlamaAssets() map[string][]byte {
	return map[string][]byte{
		"llama-server": llamaServerBin,
		"llama-cli":    llamaCliBin,
	}
}
