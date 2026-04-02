package os_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/config"
	coreos "github.com/Kaffyn/Vectora/src/core/os"
)

// TestOSFactory_HappyPath (100%): Testar se a factory retorna um manager compatível com o SO atual.
func TestOSFactory_HappyPath(t *testing.T) {
	paths := config.GetDefaultPaths()
	manager := coreos.NewOSManager(paths)

	if manager == nil {
		t.Fatal("Expected OSManager, got nil")
	}

	// Verificar se o caminho do binário faz sentido para o SO
	bin := manager.GetLlamaServerBinaryPath()

	switch runtime.GOOS {
	case "windows":
		// No Windows, deve terminar em .exe
		if !hasSuffix(bin, ".exe") {
			t.Errorf("Expected .exe suffix on Windows, got: %s", bin)
		}
	case "linux", "darwin":
		// No Unix, não deve ter .exe
		if hasSuffix(bin, ".exe") {
			t.Errorf("Expected no .exe suffix on Unix, got: %s", bin)
		}
	}
}

// TestOSFactory_Negative (200%): Testar se a factory lida com caminhos corrompidos ou inválidos.
func TestOSFactory_Negative(t *testing.T) {
	customPath := "/invalid/path/test"
	paths := config.VectoraPaths{
		Bin: customPath,
	}
	manager := coreos.NewOSManager(paths)

	bin := manager.GetLlamaServerBinaryPath()
	// Normalizar para o SO atual antes de comparar
	expected := filepath.FromSlash(customPath)
	if !contains(bin, expected) {
		t.Errorf("Expected manager to respect custom Bin path, got: %s", bin)
	}
}

func hasSuffix(s, suffix string) bool {
	if len(s) < len(suffix) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)
}
