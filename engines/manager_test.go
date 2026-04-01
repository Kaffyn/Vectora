package engines

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadCatalog verifica se o catálogo é carregado corretamente.
func TestLoadCatalog(t *testing.T) {
	if err := LoadCatalog(); err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}

	catalog, err := GetCatalog()
	if err != nil {
		t.Fatalf("failed to get catalog: %v", err)
	}

	if len(catalog.Builds) == 0 {
		t.Fatal("catalog has no builds")
	}

	// Verificar estrutura básica
	for _, build := range catalog.Builds {
		if build.ID == "" {
			t.Error("build has no ID")
		}
		if build.OS == "" {
			t.Error("build has no OS")
		}
		if build.Architecture == "" {
			t.Error("build has no Architecture")
		}
		if build.SHA256 == "" {
			t.Error("build has no SHA256")
		}
	}
}

// TestDetectHardware verifica se a detecção de hardware funciona.
func TestDetectHardware(t *testing.T) {
	hw, err := DetectHardware()
	if err != nil {
		t.Fatalf("failed to detect hardware: %v", err)
	}

	if hw.OS == "" {
		t.Fatal("hardware detection failed: OS is empty")
	}

	if hw.Architecture == "" {
		t.Fatal("hardware detection failed: Architecture is empty")
	}

	if hw.CoreCount == 0 {
		t.Fatal("hardware detection failed: CoreCount is 0")
	}

	if hw.RAM == 0 {
		t.Fatal("hardware detection failed: RAM is 0")
	}

	t.Logf("Detected: OS=%s, Arch=%s, Cores=%d, RAM=%.2f GB, GPU=%s",
		hw.OS, hw.Architecture, hw.CoreCount, float64(hw.RAM)/(1024*1024*1024), hw.GPUType)
}

// TestFindBestBuild verifica se a melhor build é encontrada.
func TestFindBestBuild(t *testing.T) {
	if err := LoadCatalog(); err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}

	hw, err := DetectHardware()
	if err != nil {
		t.Fatalf("failed to detect hardware: %v", err)
	}

	build, err := RecommendedBuild(hw)
	if err != nil {
		t.Fatalf("failed to find recommended build: %v", err)
	}

	if build.ID == "" {
		t.Fatal("recommended build has no ID")
	}

	if build.OS != hw.OS {
		t.Fatalf("recommended build OS %s doesn't match hardware OS %s", build.OS, hw.OS)
	}

	t.Logf("Recommended build: %s (%s)", build.ID, build.Description)
}

// TestVerifyFile verifica a funcionalidade de integridade de arquivo.
func TestVerifyFile(t *testing.T) {
	// Criar arquivo temporário
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "hello world"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Computar hash correto
	correctHash, err := ComputeFileSHA256(testFile)
	if err != nil {
		t.Fatalf("failed to compute hash: %v", err)
	}

	// Verificar com hash correto
	if err := VerifyFile(testFile, correctHash); err != nil {
		t.Fatalf("verification failed with correct hash: %v", err)
	}

	// Verificar com hash incorreto
	if err := VerifyFile(testFile, "wronghash1234567890abcdef"); err == nil {
		t.Fatal("verification should fail with wrong hash")
	}
}

// TestExtractZIP testa a extração de arquivo ZIP.
func TestExtractZIP(t *testing.T) {
	// Este teste requer um arquivo ZIP real
	// Skipar se não disponível
	t.Skip("ZIP extraction test requires real ZIP file")
}

// TestEngineManager testa o gerenciador de engines.
func TestEngineManager(t *testing.T) {
	// Este teste requer download real e espaço em disco
	// Skipar em testes rápidos
	t.Skip("EngineManager test requires download and disk space")
}

// TestPaths verifica resolução de paths.
func TestPaths(t *testing.T) {
	enginesDir, err := GetEnginesDir()
	if err != nil {
		t.Fatalf("failed to get engines dir: %v", err)
	}

	if enginesDir == "" {
		t.Fatal("engines directory is empty")
	}

	// Verificar se é um path absoluto
	if !filepath.IsAbs(enginesDir) {
		t.Fatalf("engines directory is not absolute: %s", enginesDir)
	}

	// Verificar se pode ser criado (ou já existe)
	if err := os.MkdirAll(enginesDir, 0755); err != nil {
		t.Fatalf("failed to create engines directory: %v", err)
	}

	t.Logf("Engines directory: %s", enginesDir)
}

// BenchmarkDetectHardware mede o desempenho da detecção de hardware.
func BenchmarkDetectHardware(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DetectHardware()
	}
}

// BenchmarkComputeSHA256 mede o desempenho do cálculo de SHA256.
func BenchmarkComputeSHA256(b *testing.B) {
	// Criar arquivo temporário
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "test.bin")

	// Escrever 1MB de dados
	data := make([]byte, 1024*1024)
	if err := os.WriteFile(testFile, data, 0644); err != nil {
		b.Fatalf("failed to create test file: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ComputeFileSHA256(testFile)
	}
}

// Fuzzing test (se Go 1.18+)
// func FuzzProcessComplete(f *testing.F) {
//     f.Fuzz(func(t *testing.T, prompt string) {
//         // Testar robustez com inputs aleatórios
//     })
// }
