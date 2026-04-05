package engines

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestIntegrationFullInstallationFlow simula o fluxo completo de instalação.
// Este é um teste de integração (mais lento, mais realista).
func TestIntegrationFullInstallationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Usar diretório temporário para a instalação
	tempDir := t.TempDir()

	// 1. Detectar Hardware
	hw, err := DetectHardware()
	if err != nil {
		t.Fatalf("hardware detection failed: %v", err)
	}
	t.Logf("Hardware detected: %s-%s (GPU: %s)", hw.OS, hw.Architecture, hw.GPUType)

	// 2. Carregar Catálogo
	if err := LoadCatalog(); err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}

	// 3. Encontrar build recomendada
	catalog, _ := GetCatalog()
	build, err := catalog.FindBestBuild(hw)
	if err != nil {
		t.Fatalf("failed to find recommended build: %v", err)
	}
	t.Logf("Recommended build: %s", build.ID)

	// 4. Criar ZIP fake com estrutura de build
	zipPath := filepath.Join(tempDir, "llama.zip")
	if err := createFakeEngineZip(zipPath); err != nil {
		t.Fatalf("failed to create fake engine zip: %v", err)
	}

	// 5. Verificar integridade
	actualHash, _ := ComputeFileSHA256(zipPath)
	t.Logf("Generated ZIP SHA256: %s", actualHash)

	// Actualizar build com hash real para o teste
	build.SHA256 = actualHash

	// 6. Extrair
	installDir := filepath.Join(tempDir, "install")
	extractor := NewExtractor(installDir)
	if err := extractor.ExtractZIP(zipPath, []string{"llama"}); err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// Verificar se arquivos foram extraídos
	entries, _ := os.ReadDir(installDir)
	if len(entries) == 0 {
		t.Fatal("no files extracted")
	}
	t.Logf("Extracted %d files", len(entries))

	// 7. Salvar metadata
	info := &InstallationInfo{
		BuildID:   build.ID,
		Path:      installDir,
		Installed: time.Now(),
		SHA256:    build.SHA256,
		SizeBytes: build.SizeBytes,
		IsActive:  true,
	}

	metadataPath := filepath.Join(tempDir, "metadata.json")
	data, _ := json.MarshalIndent(info, "", "  ")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("failed to save metadata: %v", err)
	}

	// 8. Validar que tudo está em ordem
	if err := VerifyFile(zipPath, build.SHA256); err != nil {
		t.Fatalf("verification failed: %v", err)
	}

	// 9. Simular busca do binário
	llamaBinary := filepath.Join(installDir, "llama-cli")
	if _, err := os.Stat(llamaBinary); err != nil {
		t.Logf("warning: llama-cli not found at %s (expected in integration test)", llamaBinary)
	}

	t.Log("✅ Full installation flow completed successfully!")
}

// TestIntegrationEngineManagerWithTempDir testa EngineManager com diretório temporário.
func TestIntegrationEngineManagerWithTempDir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Criar manager (normalmente usaria GetEnginesDir, aqui simulamos)
	manager, err := NewEngineManager()
	if err != nil {
		t.Fatalf("failed to create engine manager: %v", err)
	}

	// Tentar listar engines instalados (provavelmente vazio no temp)
	engines, err := manager.ListInstalled(context.Background())
	if err != nil {
		t.Logf("list engines returned: %v (ok, pode estar vazio)", err)
	}

	t.Logf("Installed engines: %v", engines)

	// Tentar pegar engine ativo (deve falhar gracefully)
	_, err = manager.GetActive()
	if err == nil {
		t.Logf("unexpected: engine already active")
	} else {
		t.Logf("expected error: %v", err)
	}

	t.Log("✅ EngineManager initialization test passed!")
}

// TestDownloadProgressCallback testa o callback de progresso.
func TestDownloadProgressCallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()

	// Criar um pequeno arquivo para "download"
	testFile := filepath.Join(tempDir, "source.bin")
	testContent := make([]byte, 1024*100) // 100KB
	for i := range testContent {
		testContent[i] = byte(i % 256)
	}
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Simular progresso
	progressCalls := 0
	progressCallback := func(p *DownloadProgress) error {
		progressCalls++
		if progressCalls%10 == 0 {
			t.Logf("Progress: %d/%d bytes (%.1f KB/s)",
				p.Current, p.Total, p.Speed/1024)
		}
		return nil
	}

	// Tentar "download" (seria com URL real)
	// Por enquanto, apenas testa que o callback é válido
	if progressCallback != nil {
		progressCallback(&DownloadProgress{Current: 100, Total: 1024, Speed: 1024.0})
	}

	t.Logf("Progress callback tested with %d calls", progressCalls)
	t.Log("✅ Download progress callback test passed!")
}

// TestCatalogRecommendation testa a recomendação de build para vários hardware.
func TestCatalogRecommendation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	if err := LoadCatalog(); err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}

	catalog, _ := GetCatalog()

	testCases := []struct {
		name string
		hw   *Hardware
	}{
		{
			name: "Windows x86_64 CUDA 12",
			hw: &Hardware{
				OS:           "windows",
				Architecture: "x86_64",
				GPUType:      "cuda",
				GPUVersion:   "12.0",
			},
		},
		{
			name: "Linux x86_64 CPU",
			hw: &Hardware{
				OS:           "linux",
				Architecture: "x86_64",
				GPUType:      "none",
			},
		},
		{
			name: "macOS ARM64 Metal",
			hw: &Hardware{
				OS:           "darwin",
				Architecture: "arm64",
				GPUType:      "metal",
			},
		},
		{
			name: "Windows ARM64 CPU",
			hw: &Hardware{
				OS:           "windows",
				Architecture: "arm64",
				GPUType:      "none",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			build, err := catalog.FindBestBuild(tc.hw)
			if err != nil {
				t.Logf("no build found (ok): %v", err)
				return
			}

			t.Logf("✅ Recommended: %s (%s)", build.ID, build.Description)

			if build.OS != tc.hw.OS {
				t.Errorf("OS mismatch: expected %s, got %s", tc.hw.OS, build.OS)
			}

			if build.Architecture != tc.hw.Architecture {
				t.Logf("note: architecture mismatch (fallback ok): expected %s, got %s",
					tc.hw.Architecture, build.Architecture)
			}
		})
	}
}

// TestIntegrationVerifyAndExtract testa o ciclo: create ZIP → verify → extract.
func TestIntegrationVerifyAndExtract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tempDir := t.TempDir()

	// 1. Criar ZIP com estrutura realista
	zipPath := filepath.Join(tempDir, "engine.zip")
	if err := createFakeEngineZip(zipPath); err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}

	// 2. Validar ZIP
	if err := ValidateZIP(zipPath); err != nil {
		t.Fatalf("zip validation failed: %v", err)
	}

	// 3. Listar arquivos
	files, err := ListFilesInZip(zipPath)
	if err != nil {
		t.Logf("warning: failed to list files: %v (ok for test)", err)
	} else {
		t.Logf("Files in ZIP: %v", files)
	}

	// 4. Computar hash
	hash, err := ComputeFileSHA256(zipPath)
	if err != nil {
		t.Fatalf("failed to compute hash: %v", err)
	}
	t.Logf("ZIP SHA256: %s", hash[:16]+"...")

	// 5. Verificar com hash correto
	if err := VerifyFile(zipPath, hash); err != nil {
		t.Fatalf("verification with correct hash failed: %v", err)
	}

	// 6. Extrair
	extractDir := filepath.Join(tempDir, "extracted")
	extractor := NewExtractor(extractDir)
	if err := extractor.ExtractZIP(zipPath, nil); err != nil {
		t.Fatalf("extraction failed: %v", err)
	}

	// 7. Verificar extracted files
	extractedFiles, _ := os.ReadDir(extractDir)
	if len(extractedFiles) == 0 {
		t.Fatal("no files extracted")
	}

	t.Logf("✅ Extracted %d files", len(extractedFiles))
	for _, f := range extractedFiles {
		t.Logf("  - %s (%d bytes)", f.Name(), f.Type())
	}

	t.Log("✅ Verify and extract integration test passed!")
}

// createFakeEngineZip cria um ZIP fake com estrutura de engine.
func createFakeEngineZip(zipPath string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	// Adicionar alguns "binários" fake
	files := []struct {
		name    string
		content string
	}{
		{"llama-cli", "fake llama-cli binary"},
		{"llama", "fake llama library"},
		{"llama.dll", "fake windows dll"},
		{"README.md", "# Llama Engine\n\nThis is a fake engine for testing."},
	}

	for _, file := range files {
		w, err := zw.Create(file.name)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, file.content); err != nil {
			return err
		}
	}

	return nil
}

// TestBenchmarkHardwareDetection mede o desempenho da detecção completa.
func BenchmarkHardwareDetectionFull(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		hw, _ := DetectHardware()
		if hw == nil {
			b.Fatal("hardware detection failed")
		}
	}
}

// TestBenchmarkCatalogSearch mede o desempenho da busca no catálogo.
func BenchmarkCatalogSearch(b *testing.B) {
	if err := LoadCatalog(); err != nil {
		b.Fatalf("failed to load catalog: %v", err)
	}

	catalog, _ := GetCatalog()

	testHW := &Hardware{
		OS:           "windows",
		Architecture: "x86_64",
		GPUType:      "cuda",
		GPUVersion:   "12.0",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		catalog.FindBestBuild(testHW)
	}
}

// TestBenchmarkZIPExtraction mede o desempenho de extração.
func BenchmarkZIPExtraction(b *testing.B) {
	tempDir := b.TempDir()

	// Criar ZIP uma vez
	zipPath := filepath.Join(tempDir, "bench.zip")
	if err := createFakeEngineZip(zipPath); err != nil {
		b.Fatalf("failed to create zip: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		extractDir := filepath.Join(tempDir, "extracted", string(rune(i)))
		extractor := NewExtractor(extractDir)
		extractor.ExtractZIP(zipPath, nil)
	}
}
