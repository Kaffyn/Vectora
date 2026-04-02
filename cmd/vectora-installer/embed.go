package main

import _ "embed"

//go:embed ../../vectora.exe
var vectoraExe []byte

// [FIX] Llama server deve estar na raiz do repo ou na pasta cmd/
//go:embed ../../llama-server.exe
var llamaServerExe []byte
