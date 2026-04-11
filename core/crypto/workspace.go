package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// WorkspaceSalter handles salted hashing of workspace paths.
// Each installation has a unique salt stored in ~/.Vectora/salt, ensuring
// workspace IDs differ across different machines.
type WorkspaceSalter struct {
	salt string
}

// NewWorkspaceSalter initializes the salter, creating the salt file if needed.
func NewWorkspaceSalter(appDataDir string) (*WorkspaceSalter, error) {
	saltPath := filepath.Join(appDataDir, "salt")

	// Read existing salt or create new one
	saltBytes, err := os.ReadFile(saltPath)
	if err == nil {
		return &WorkspaceSalter{salt: string(saltBytes)}, nil
	}

	// Generate new 32-byte salt
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	salt := hex.EncodeToString(buf)

	// Write salt to disk
	if err := os.WriteFile(saltPath, []byte(salt), 0600); err != nil {
		return nil, fmt.Errorf("failed to write salt: %w", err)
	}

	return &WorkspaceSalter{salt: salt}, nil
}

// HashPath returns a deterministic SHA256 hash of the workspace path
// mixed with the installation's unique salt.
// Result is a 64-character hex string suitable for use as a collection ID.
func (s *WorkspaceSalter) HashPath(path string) string {
	h := sha256.New()
	h.Write([]byte(s.salt))
	h.Write([]byte(path))
	return hex.EncodeToString(h.Sum(nil))
}
