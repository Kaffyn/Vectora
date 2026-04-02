package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// VectoraPaths defines the standard directory structure for Vectora.
type VectoraPaths struct {
	Root   string
	Bin    string
	DB     string
	Models string // models/qwen
	Data   string // models/data (indices)
	Docs   string
}

// GetDefaultPaths calculates the paths based on the environment and home directory.
func GetDefaultPaths() VectoraPaths {
	isDev := os.Getenv("VECTORA_DEV_MODE") == "true"

	var root string
	if isDev {
		// No modo dev, usamos o diretório atual do projeto
		root, _ = os.Getwd()
	} else {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".Vectora")
	}

	return VectoraPaths{
		Root:   root,
		Bin:    filepath.Join(root, "bin"),
		DB:     filepath.Join(root, "db"),
		Models: filepath.Join(root, "models", "qwen"),
		Data:   filepath.Join(root, "models", "data"),
		Docs:   filepath.Join(root, "docs"),
	}
}

// GetBinaryPath returns the absolute path for a specific sidecar binary.
func (p VectoraPaths) GetBinaryPath(name string) string {
	binName := name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return filepath.Join(p.Bin, binName)
}

// EnsureDirectories creates the necessary folder structure if it doesn't exist.
func (p VectoraPaths) EnsureDirectories() error {
	dirs := []string{p.Bin, p.DB, p.Models, p.Data, p.Docs}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
