//go:build ignore

package main

// Note: llama binaries are not embedded in the binary.
// They are downloaded and managed by the engines package.
// This file is kept for future binary embedding if needed.

func getLlamaAssets() map[string][]byte {
	// Return empty map - binaries are managed externally
	return make(map[string][]byte)
}
