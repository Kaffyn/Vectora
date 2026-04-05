package models

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	vecos "github.com/Kaffyn/Vectora/internal/os"
)

// NewModelManager cria uma nova instância do gerenciador de modelos
func NewModelManager() (*ModelManager, error) {
	// Obter diretório de dados do aplicativo
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OS manager: %w", err)
	}

	baseDir, err := osMgr.GetAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get app data dir: %w", err)
	}

	modelsDir := filepath.Join(baseDir, "models")

	// Criar diretório se não existir
	if err := os.MkdirAll(modelsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create models directory: %w", err)
	}

	// Carregar catálogo
	catalog, err := LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	mm := &ModelManager{
		catalog:   catalog,
		modelsDir: modelsDir,
		metadata:  make(map[string]*InstalledModel),
	}

	// Carregar metadados existentes
	mm.loadMetadata()

	return mm, nil
}

// GetCatalog retorna o catálogo de modelos
func (mm *ModelManager) GetCatalog() *Catalog {
	return mm.catalog
}

// GetModelsDir retorna o diretório onde os modelos são armazenados
func (mm *ModelManager) GetModelsDir() string {
	return mm.modelsDir
}

// ListInstalled lista todos os modelos instalados
func (mm *ModelManager) ListInstalled() []string {
	var ids []string
	for id := range mm.metadata {
		if mm.metadata[id].Installed {
			ids = append(ids, id)
		}
	}
	return ids
}

// IsInstalled verifica se um modelo está instalado
func (mm *ModelManager) IsInstalled(modelID string) bool {
	meta, exists := mm.metadata[modelID]
	return exists && meta.Installed
}

// GetModelPath retorna o caminho completo para um modelo instalado
func (mm *ModelManager) GetModelPath(modelID string) (string, error) {
	_, err := FindModel(mm.catalog, modelID)
	if err != nil {
		return "", err
	}

	path := filepath.Join(mm.modelsDir, modelID, fmt.Sprintf("%s.gguf", modelID))
	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("model file not found: %s", path)
	}

	return path, nil
}

// GetActive retorna o ID do modelo ativo
func (mm *ModelManager) GetActive() string {
	// Ler do arquivo de metadados raiz
	metaPath := filepath.Join(mm.modelsDir, "active.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		return "" // Nenhum modelo ativo
	}

	var meta map[string]string
	if err := json.Unmarshal(data, &meta); err != nil {
		return ""
	}

	return meta["active"]
}

// SetActive define qual modelo é o ativo
func (mm *ModelManager) SetActive(modelID string) error {
	if !mm.IsInstalled(modelID) {
		return fmt.Errorf("model '%s' is not installed", modelID)
	}

	metaPath := filepath.Join(mm.modelsDir, "active.json")
	meta := map[string]string{"active": modelID}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	return nil
}

// RecommendModel usa algoritmo de 4 camadas para recomendar um modelo
func (mm *ModelManager) RecommendModel(hw *Hardware) (*Model, error) {
	if mm.catalog == nil || len(mm.catalog.Models) == 0 {
		return nil, fmt.Errorf("no models available in catalog")
	}

	// Converter RAM de bytes para GB e reservar 2GB para o sistema
	ramGB := float64(hw.RAM) / (1024 * 1024 * 1024)
	availableRAM := ramGB - 2.0
	if availableRAM < 1.0 {
		availableRAM = 1.0
	}

	// Camada 1: Ajuste perfeito (modelo que cabe e tem GPU match)
	var canFit []Model
	for _, model := range mm.catalog.Models {
		if model.RequiredRAMGB <= availableRAM {
			// Preferir modelos de chat/chat genérico
			for _, cap := range model.Capabilities {
				if cap == "chat" {
					canFit = append(canFit, model)
					break
				}
			}
		}
	}

	if len(canFit) > 0 {
		// Retornar o maior que cabe
		largest := canFit[0]
		for _, m := range canFit {
			if m.RequiredRAMGB > largest.RequiredRAMGB {
				largest = m
			}
		}
		return &largest, nil
	}

	// Camada 2: Maior modelo que cabe (ignorar preferência de GPU)
	var bestFit *Model
	for i, model := range mm.catalog.Models {
		if model.RequiredRAMGB <= availableRAM {
			if bestFit == nil || model.RequiredRAMGB > bestFit.RequiredRAMGB {
				bestFit = &mm.catalog.Models[i]
			}
		}
	}

	if bestFit != nil {
		return bestFit, nil
	}

	// Camada 3: Maior modelo no geral (pode disparar swap)
	largest := &mm.catalog.Models[0]
	for i := 1; i < len(mm.catalog.Models); i++ {
		if mm.catalog.Models[i].RequiredRAMGB > largest.RequiredRAMGB {
			largest = &mm.catalog.Models[i]
		}
	}

	// Camada 4: Fallback para o menor (sempre deve caber)
	smallest := GetSmallestModel(mm.catalog)
	if smallest == nil {
		return nil, fmt.Errorf("no models available")
	}

	return smallest, nil
}

// loadMetadata carrega metadados de modelos instalados do disco
func (mm *ModelManager) loadMetadata() {
	metaPath := filepath.Join(mm.modelsDir, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		// Arquivo não existe ainda, isso é normal
		return
	}

	var metadata map[string]*InstalledModel
	if err := json.Unmarshal(data, &metadata); err != nil {
		fmt.Printf("Warning: failed to load metadata: %v\n", err)
		return
	}

	mm.metadata = metadata
}

// Install orquestra o fluxo completo de instalação de um modelo
// Esta é uma função stub - a implementação completa de download/verify
// será adicionada quando integrada com o downloader
func (mm *ModelManager) Install(ctx context.Context, modelID string, onProgress func(*DownloadProgress)) error {
	// Verificar se modelo existe no catálogo
	foundModel, err := FindModel(mm.catalog, modelID)
	if err != nil {
		return fmt.Errorf("model not found: %w", err)
	}

	// Criar diretório para o modelo
	modelDir := filepath.Join(mm.modelsDir, modelID)
	if err := os.MkdirAll(modelDir, 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	// Salvar metadados (implementação real incluiria download)
	mm.metadata[modelID] = &InstalledModel{
		ID:        modelID,
		Path:      filepath.Join(modelDir, fmt.Sprintf("%s.gguf", modelID)),
		Installed: true,
		SHA256:    foundModel.SHA256,
		Size:      foundModel.SizeBytes,
	}

	// Persistir metadados
	metaPath := filepath.Join(mm.modelsDir, "metadata.json")
	data, _ := json.MarshalIndent(mm.metadata, "", "  ")
	os.WriteFile(metaPath, data, 0644)

	return nil
}
