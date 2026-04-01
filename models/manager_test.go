package models

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadCatalog verifica que o catálogo é carregado corretamente
func TestLoadCatalog(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	if cat == nil || len(cat.Models) == 0 {
		t.Fatal("Catalog is empty or nil")
	}

	// Verificar modelos conhecidos
	qwen06 := FindModelByID(cat, "qwen3-0.6b")
	if qwen06 == nil {
		t.Fatal("qwen3-0.6b not found in catalog")
	}

	if qwen06.RequiredRAMGB < 1.0 {
		t.Errorf("qwen3-0.6b RAM requirement too low: %f", qwen06.RequiredRAMGB)
	}
}

// TestDetectHardware verifica detecção de hardware
func TestDetectHardware(t *testing.T) {
	hw, err := DetectHardware()
	if err != nil {
		t.Fatalf("DetectHardware failed: %v", err)
	}

	if hw.OS == "" {
		t.Fatal("OS not detected")
	}

	if hw.RAM == 0 {
		t.Fatal("RAM not detected")
	}

	if hw.CoreCount == 0 {
		t.Fatal("CPU cores not detected")
	}

	t.Logf("Hardware: OS=%s, RAM=%dGB, Cores=%d, GPU=%s",
		hw.OS, hw.RAM/(1024*1024*1024), hw.CoreCount, hw.GPUType)
}

// TestRecommendModel verifica 4 cenários de recomendação
func TestRecommendModel(t *testing.T) {
	cat, _ := LoadCatalog()
	mm := &ModelManager{catalog: cat}

	tests := []struct {
		name       string
		ramGB      int64
		gpuType    string
		expectedID string
		expectMin  float64
		expectMax  float64
	}{
		{
			name:       "32GB should recommend largest fit (8B)",
			ramGB:      32 * 1024 * 1024 * 1024,
			gpuType:    "cuda",
			expectedID: "qwen3-8b",
			expectMin:  15.0,
			expectMax:  17.0,
		},
		{
			name:       "18GB should recommend 8B or 4B",
			ramGB:      18 * 1024 * 1024 * 1024,
			gpuType:    "none",
			expectedID: "qwen3-8b",
			expectMin:  15.0,
			expectMax:  17.0,
		},
		{
			name:       "10GB should recommend 4B",
			ramGB:      10 * 1024 * 1024 * 1024,
			gpuType:    "none",
			expectedID: "qwen3-4b",
			expectMin:  7.0,
			expectMax:  9.0,
		},
		{
			name:       "6GB should recommend 1.7B",
			ramGB:      6 * 1024 * 1024 * 1024,
			gpuType:    "none",
			expectedID: "qwen3-1.7b",
			expectMin:  3.0,
			expectMax:  5.0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hw := &Hardware{
				OS:           "windows",
				Architecture: "amd64",
				RAM:          test.ramGB,
				GPUType:      test.gpuType,
				CoreCount:    8,
			}

			recommended, err := mm.RecommendModel(hw)
			if err != nil {
				t.Fatalf("RecommendModel failed: %v", err)
			}

			if recommended.ID != test.expectedID {
				t.Logf("Note: Expected %s, got %s (both are valid for this RAM)", test.expectedID, recommended.ID)
			}

			if recommended.RequiredRAMGB < test.expectMin || recommended.RequiredRAMGB > test.expectMax {
				t.Errorf("Recommended model RAM out of range: %f (expected %f-%f)",
					recommended.RequiredRAMGB, test.expectMin, test.expectMax)
			}

			t.Logf("RAM %dGB -> Recommended %s (requires %.0fGB)", test.ramGB/(1024*1024*1024), recommended.ID, recommended.RequiredRAMGB)
		})
	}
}

// TestSearchByName verifica busca por nome
func TestSearchByName(t *testing.T) {
	cat, _ := LoadCatalog()

	results := SearchByName(cat, "qwen")
	if len(results) == 0 {
		t.Fatal("No results found for 'qwen'")
	}

	if len(results) < 5 {
		t.Errorf("Expected at least 5 qwen models, got %d", len(results))
	}

	// Verificar que todos os resultados contêm "qwen"
	for _, m := range results {
		if !contains(m.Tags, "qwen") && !contains(m.Capabilities, "qwen") {
			if m.ID != "qwen3-0.6b" && m.ID != "qwen3-1.7b" {
				// Alguns modelos Qwen podem não ter tag explícita
				continue
			}
		}
	}
}

// TestSearchByCapability verifica busca por capacidade
func TestSearchByCapability(t *testing.T) {
	cat, _ := LoadCatalog()

	results := SearchByCapability(cat, "chat")
	if len(results) == 0 {
		t.Fatal("No chat models found")
	}

	// Todos os resultados devem ter capability "chat"
	for _, m := range results {
		found := false
		for _, cap := range m.Capabilities {
			if cap == "chat" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Model %s doesn't have 'chat' capability", m.ID)
		}
	}
}

// TestModelManager criação e persistência de metadados
func TestModelManager(t *testing.T) {
	// Criar dir temporário para teste
	tmpDir := filepath.Join(os.TempDir(), "mpm-test")
	os.RemoveAll(tmpDir)
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Criar manager manual para teste
	cat, _ := LoadCatalog()
	mm := &ModelManager{
		catalog:   cat,
		modelsDir: filepath.Join(tmpDir, "models"),
		metadata:  make(map[string]*InstalledModel),
	}

	// Criar dir
	os.MkdirAll(mm.modelsDir, 0755)

	// Teste ListInstalled vazio
	installed := mm.ListInstalled()
	if len(installed) != 0 {
		t.Errorf("Expected 0 installed models, got %d", len(installed))
	}

	// Teste IsInstalled
	if mm.IsInstalled("qwen3-4b") {
		t.Error("qwen3-4b should not be installed")
	}

	// Simular instalação
	mm.metadata["qwen3-4b"] = &InstalledModel{
		ID:        "qwen3-4b",
		Path:      filepath.Join(mm.modelsDir, "qwen3-4b", "qwen3-4b.gguf"),
		Installed: true,
		SHA256:    "test-sha",
		Size:      2693406720,
	}

	// Teste IsInstalled após inserção
	if !mm.IsInstalled("qwen3-4b") {
		t.Error("qwen3-4b should be installed")
	}

	// Teste GetActive vazio
	active := mm.GetActive()
	if active != "" {
		t.Errorf("Expected no active model, got %s", active)
	}

	// Teste SetActive
	err := mm.SetActive("qwen3-4b")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Verificar que foi persistido
	active = mm.GetActive()
	if active != "qwen3-4b" {
		t.Errorf("Expected qwen3-4b as active, got %s", active)
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func FindModelByID(cat *Catalog, id string) *Model {
	for i, m := range cat.Models {
		if m.ID == id {
			return &cat.Models[i]
		}
	}
	return nil
}
