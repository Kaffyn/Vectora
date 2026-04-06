package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// EngineManager gerencia o ciclo de vida completo de engines (llama.cpp).
type EngineManager struct {
	baseDir         string
	installDir      string
	activeInfo      *InstallationInfo
	metadataPath    string
}

// NewEngineManager cria um novo manager de engines.
func NewEngineManager() (*EngineManager, error) {
	baseDir, err := GetEnginesDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get engines directory: %w", err)
	}

	installDir := filepath.Join(baseDir, "llama")
	metadataPath := filepath.Join(baseDir, "metadata.json")

	em := &EngineManager{
		baseDir:      baseDir,
		installDir:   installDir,
		metadataPath: metadataPath,
	}

	// Carregar informações de engine ativo se existir
	em.loadMetadata()

	return em, nil
}

// Install baixa, verifica e instala uma build específica.
func (m *EngineManager) Install(
	ctx context.Context,
	buildID string,
	onProgress func(*DownloadProgress) error,
) error {
	// Obter build do catálogo
	build, err := GetBuildByID(buildID)
	if err != nil {
		return fmt.Errorf("build not found: %w", err)
	}

	// Verificar se já instalado
	installed, err := m.IsInstalled(buildID)
	if err == nil && installed {
		return fmt.Errorf("build %s already installed", buildID)
	}

	// Criar diretório temporário para download
	tempDir := filepath.Join(m.baseDir, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Download
	downloadPath := filepath.Join(tempDir, buildID+".zip")
	downloader := NewDownloader()

	fmt.Printf("Downloading %s...\n", buildID)
	if err := downloader.Download(ctx, build.DownloadURL, downloadPath, onProgress); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Verificar integridade
	fmt.Printf("Verifying integrity...\n")
	if err := VerifyFile(downloadPath, build.SHA256); err != nil {
		return fmt.Errorf("integrity check failed: %w", err)
	}

	// Extrair
	fmt.Printf("Extracting files...\n")
	extractor := NewExtractor(m.installDir)
	// Filtros: extrair apenas binários e DLLs/SOs
	filters := []string{"llama", ".exe", ".dll", ".so"}
	if err := extractor.ExtractZIP(downloadPath, filters); err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}

	// Salvar metadados
	info := &InstallationInfo{
		BuildID:    buildID,
		Path:       m.installDir,
		Installed:  time.Now(),
		SHA256:     build.SHA256,
		SizeBytes:  build.SizeBytes,
		IsActive:   true,
	}

	if err := m.saveMetadata(info); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	m.activeInfo = info
	fmt.Printf("Successfully installed %s\n", buildID)
	return nil
}

// IsInstalled verifica se uma build está instalada.
func (m *EngineManager) IsInstalled(buildID string) (bool, error) {
	// Verificar se diretório existe e tem binários
	installDir := filepath.Join(m.baseDir, "llama")
	info, err := os.Stat(installDir)
	if err != nil {
		return false, nil
	}
	if !info.IsDir() {
		return false, nil
	}

	// Verificar metadados
	if err := m.loadMetadata(); err == nil && m.activeInfo != nil {
		return m.activeInfo.BuildID == buildID, nil
	}

	return false, nil
}

// GetActive retorna a build atualmente ativa.
func (m *EngineManager) GetActive() (*InstallationInfo, error) {
	if m.activeInfo == nil {
		return nil, fmt.Errorf("no active engine")
	}
	return m.activeInfo, nil
}

// SetActive marca uma build como ativa.
func (m *EngineManager) SetActive(buildID string) error {
	installed, _ := m.IsInstalled(buildID)
	if !installed {
		return fmt.Errorf("build %s not installed", buildID)
	}

	m.activeInfo = &InstallationInfo{
		BuildID:   buildID,
		Path:      m.installDir,
		IsActive:  true,
	}

	return m.saveMetadata(m.activeInfo)
}

// Start inicia um processo llama.cpp com a build ativa.
func (m *EngineManager) Start(
	ctx context.Context,
	ctxTokens int,
	nThreads int,
	gpuLayers int,
) (*ProcessManager, error) {
	if m.activeInfo == nil {
		return nil, fmt.Errorf("no active engine")
	}

	// Localizar binário llama-cli
	llamaBinary := m.findLlamaBinary()
	if llamaBinary == "" {
		return nil, fmt.Errorf("llama-cli binary not found")
	}

	// TODO: Localizar modelo GGUF (por enquanto, usar path hardcoded)
	modelPath := filepath.Join(m.baseDir, "models", "qwen.gguf")

	// Verificar se modelo existe
	if _, err := os.Stat(modelPath); err != nil {
		return nil, fmt.Errorf("model not found at %s: %w", modelPath, err)
	}

	return StartProcess(ctx, llamaBinary, modelPath, ctxTokens, nThreads, gpuLayers)
}

// ListInstalled lista todas as builds instaladas.
func (m *EngineManager) ListInstalled(ctx context.Context) ([]string, error) {
	var builds []string

	entries, err := os.ReadDir(m.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "temp" && entry.Name() != "models" {
			builds = append(builds, entry.Name())
		}
	}

	return builds, nil
}

// Uninstall remove uma build instalada.
func (m *EngineManager) Uninstall(ctx context.Context, buildID string) error {
	// Não permitir desinstalar a engine ativa
	if m.activeInfo != nil && m.activeInfo.BuildID == buildID {
		return fmt.Errorf("cannot uninstall active engine")
	}

	buildDir := filepath.Join(m.baseDir, buildID)
	return os.RemoveAll(buildDir)
}

// GetBinaryPath retorna o caminho absoluto do binário llama-cli.
func (m *EngineManager) GetBinaryPath() string {
	return m.findLlamaBinary()
}

// Private helpers

func (m *EngineManager) findLlamaBinary() string {
	// Procurar em locais comuns
	locations := []string{
		filepath.Join(m.installDir, "llama-cli"),
		filepath.Join(m.installDir, "llama-cli.exe"),
		filepath.Join(m.installDir, "bin", "llama-cli"),
		filepath.Join(m.installDir, "bin", "llama-cli.exe"),
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc
		}
	}

	return ""
}

func (m *EngineManager) loadMetadata() error {
	data, err := os.ReadFile(m.metadataPath)
	if err != nil {
		return err
	}

	m.activeInfo = &InstallationInfo{}
	return json.Unmarshal(data, m.activeInfo)
}

func (m *EngineManager) saveMetadata(info *InstallationInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.metadataPath, data, 0644)
}
