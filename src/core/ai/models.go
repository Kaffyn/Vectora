package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Kaffyn/Vectora/src/core/config"
)

// ModelInfo representa um modelo no registro.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Size    string `json:"size"`
	URL     string `json:"url"`
}

// RegistryResponse representa a resposta do backserver.
type RegistryResponse struct {
	Models []ModelInfo `json:"models"`
}

// ModelManager gerencia a instalação e remoção de modelos GGUF.
type ModelManager struct {
	paths config.VectoraPaths
}

// NewModelManager cria uma nova instância do ModelManager.
func NewModelManager(paths config.VectoraPaths) *ModelManager {
	return &ModelManager{paths: paths}
}

// FetchRegistry busca o catálogo de modelos do backserver.
func (m *ModelManager) FetchRegistry(registryURL string) (*RegistryResponse, error) {
	resp, err := http.Get(registryURL)
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar registro: %w", err)
	}
	defer resp.Body.Close()

	var registry RegistryResponse
	if err := json.NewDecoder(resp.Body).Decode(&registry); err != nil {
		return nil, fmt.Errorf("falha ao decodificar registro: %w", err)
	}

	return &registry, nil
}

// InstallModel baixa um modelo do registro para a pasta local.
func (m *ModelManager) InstallModel(info ModelInfo, progress func(float64)) error {
	destDir := filepath.Join(m.paths.Models, info.ID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("falha ao criar pasta do modelo: %w", err)
	}

	destPath := filepath.Join(destDir, info.ID+".gguf")

	// Inicia o download
	resp, err := http.Get(info.URL)
	if err != nil {
		return fmt.Errorf("falha ao iniciar download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("erro no download: status %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo de destino: %w", err)
	}
	defer out.Close()

	// TODO: Implementar progresso real usando um Wrapper pro io.Copy
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("falha durante o download: %w", err)
	}

	return nil
}

// RemoveModel apaga a pasta de um modelo.
func (m *ModelManager) RemoveModel(id string) error {
	modelDir := filepath.Join(m.paths.Models, id)
	return os.RemoveAll(modelDir)
}

// ListInstalledModels lista os modelos presentes na pasta local.
func (m *ModelManager) ListInstalledModels() ([]string, error) {
	files, err := os.ReadDir(m.paths.Models)
	if err != nil {
		return nil, err
	}

	var models []string
	for _, f := range files {
		if f.IsDir() {
			models = append(models, f.Name())
		}
	}
	return models, nil
}

type DataPackInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Author   string `json:"author"`
	Type     string `json:"type"`
	ZPackURL string `json:"zpack_url"`
}

// LibraryResponse representa a resposta do catálogo de packs.
type LibraryResponse []DataPackInfo

// FetchLibrary busca o catálogo de packs do backserver.
func (m *ModelManager) FetchLibrary(libraryURL string) (*LibraryResponse, error) {
	resp, err := http.Get(libraryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var library LibraryResponse
	if err := json.NewDecoder(resp.Body).Decode(&library); err != nil {
		return nil, err
	}
	return &library, nil
}

// InstallDataPack baixa um arquivo .zpack para a pasta de dados.
func (m *ModelManager) InstallDataPack(info DataPackInfo) error {
	destDir := filepath.Join(m.paths.Data, info.ID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destPath := filepath.Join(destDir, info.ID+".zpack")

	resp, err := http.Get(info.ZPackURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
