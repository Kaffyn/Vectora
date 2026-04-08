package policies

import (
	"path/filepath"
	"testing"
)

func TestGuardianBlocksProtectedFiles(t *testing.T) {
	g := NewGuardian("/trust")

	protectedFiles := []string{".env", "secrets.yml", "id_rsa", "test.key", "cert.pem", "db.sqlite"}
	for _, f := range protectedFiles {
		if !g.IsProtected(f) {
			t.Errorf("expected %s to be protected", f)
		}
	}
}

func TestGuardianBlocksProtectedExtensions(t *testing.T) {
	g := NewGuardian("/trust")

	protectedExts := []string{"file.db", "file.sqlite", "file.exe", "file.dll", "file.key", "file.pem", "file.log"}
	for _, f := range protectedExts {
		if !g.IsProtected(f) {
			t.Errorf("expected %s to be protected by extension", f)
		}
	}
}

func TestGuardianPathSafe(t *testing.T) {
	g := NewGuardian("/trust")

	if !g.IsPathSafe("/trust/file.txt") {
		t.Error("expected /trust/file.txt to be path-safe")
	}
	if !g.IsPathSafe("/trust/sub/file.txt") {
		t.Error("expected nested path to be safe")
	}
}

func TestGuardianExcludedDirs(t *testing.T) {
	g := NewGuardian("/trust")

	excluded := []string{".git", "node_modules", "vendor", "dist", "build"}
	for _, d := range excluded {
		if !g.IsExcludedDir(d) {
			t.Errorf("expected %s to be excluded", d)
		}
	}
}

func TestGuardianSanitizeOutput(t *testing.T) {
	g := NewGuardian("/trust")

	input := "normal text AKIAIOSFODNN7EXAMPLE more text"
	output := g.SanitizeOutput(input)
	if output == input {
		t.Error("expected secret to be redacted")
	}
}

func TestGuardianBlocksPathTraversal(t *testing.T) {
	g := NewGuardian("/trust")

	traversal := filepath.Join("/trust", "..", "escape.txt")
	if g.IsPathSafe(traversal) {
		t.Error("expected path traversal to be blocked")
	}
}
