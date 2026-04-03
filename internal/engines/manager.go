package engines

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EngineInfo contains metadata about an engine binary.
type EngineInfo struct {
	Name     string
	ExpectedSha256 string
}

// EngineManager manages the physical integrity of Vectora assets.
type EngineManager struct {
	baseDir string
}

func NewManager() (*EngineManager, error) {
	dir, err := GetEnginesDir()
	if err != nil {
		return nil, err
	}
	return &EngineManager{baseDir: dir}, nil
}

// Verify checks the integrity of a specific binary.
func (m *EngineManager) Verify(targetPath string, expectedHex string) error {
	f, err := os.Open(targetPath)
	if err != nil {
		return fmt.Errorf("engine_missing: %v", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actualHex := hex.EncodeToString(h.Sum(nil))
	if actualHex != expectedHex {
		return fmt.Errorf("integrity_fail: hash mismatch (exp: %s, got: %s)", expectedHex, actualHex)
	}

	return nil
}

// Install extracts the binaries (passed as embedded bytes) to the data folder.
func (m *EngineManager) Install(ctx context.Context, engineName string, assets map[string][]byte) error {
	dest := filepath.Join(m.baseDir, engineName)
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for filename, data := range assets {
		target := filepath.Join(dest, filename)
		if err := os.WriteFile(target, data, 0755); err != nil {
			return fmt.Errorf("extraction_error %s: %v", filename, err)
		}
	}

	return nil
}

// GetBinaryPath returns the absolute path for executing a binary.
func (m *EngineManager) GetBinaryPath(engineName string, binary string) string {
	return filepath.Join(m.baseDir, engineName, binary)
}
