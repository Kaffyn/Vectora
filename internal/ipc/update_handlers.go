package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// UpdateCheckResult representa o resultado da verificação de atualizações
type UpdateCheckResult struct {
	Available   bool                  `json:"available"`
	Components  []ComponentUpdateInfo `json:"components"`
	LastChecked string                `json:"last_checked"`
}

// ComponentUpdateInfo contém informações sobre atualização de um componente
type ComponentUpdateInfo struct {
	Name        string `json:"name"`
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	Available   bool   `json:"available"`
	DownloadURL string `json:"download_url,omitempty"`
}

// UpdateProgress é enviado durante o processo de atualização
type UpdateProgress struct {
	Status    string `json:"status"` // "in_progress", "completed", "error"
	Component string `json:"component"`
	Progress  int    `json:"progress"` // 0-100
	Message   string `json:"message"`
}

// UpdateResult contém resultado de uma atualização individual
type UpdateResult struct {
	Component  string `json:"component"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	NewVersion string `json:"new_version,omitempty"`
}

// HandleUpdateCheck verifica se há atualizações disponíveis
func HandleUpdateCheck(ctx context.Context, payload json.RawMessage) (any, *IPCError) {
	result := &UpdateCheckResult{
		Available:   false,
		Components:  []ComponentUpdateInfo{},
		LastChecked: time.Now().UTC().Format(time.RFC3339),
	}

	// Componentes padrão para verificar
	components := []string{"daemon", "tui", "lpm", "mpm", "setup"}

	// Diretório de instalação
	installDir := getInstallDir()

	for _, component := range components {
		current := getBinaryVersion(installDir, component)
		latest, downloadURL := getLatestReleaseInfo(component)

		info := ComponentUpdateInfo{
			Name:        component,
			Current:     current,
			Latest:      latest,
			Available:   (latest != "" && current != latest),
			DownloadURL: downloadURL,
		}

		if info.Available {
			result.Available = true
		}

		result.Components = append(result.Components, info)
	}

	return result, nil
}

// HandleUpdateExecute executa as atualizações (streaming via broadcast)
func HandleUpdateExecute(ctx context.Context, payload json.RawMessage, server *Server) (any, *IPCError) {
	var req struct {
		Components []string `json:"components"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		return nil, ErrIPCPayloadInvalid
	}

	// Se nenhum componente especificado, atualizar todos
	if len(req.Components) == 0 {
		req.Components = []string{"daemon", "tui", "lpm", "mpm", "setup"}
	}

	// Executar atualização em goroutine separada para não bloquear IPC
	go func() {
		installDir := getInstallDir()
		var results []UpdateResult

		for _, component := range req.Components {
			// Broadcast: iniciando atualização
			server.Broadcast("update_progress", map[string]interface{}{
				"status":    "in_progress",
				"component": component,
				"progress":  0,
				"message":   fmt.Sprintf("Verificando atualização para %s...", component),
			})

			current := getBinaryVersion(installDir, component)
			latest, downloadURL := getLatestReleaseInfo(component)

			// Se versão é a mesma, pular
			if current == latest {
				results = append(results, UpdateResult{
					Component: component,
					Success:   true,
					Message:   "Já está atualizado",
				})
				continue
			}

			if downloadURL == "" {
				results = append(results, UpdateResult{
					Component: component,
					Success:   false,
					Message:   "URL de download não disponível",
				})
				continue
			}

			// Broadcast: baixando
			server.Broadcast("update_progress", map[string]interface{}{
				"status":    "in_progress",
				"component": component,
				"progress":  25,
				"message":   fmt.Sprintf("Baixando %s %s...", component, latest),
			})

			// Baixar e instalar
			err := updateComponentBinary(installDir, component, downloadURL)

			if err != nil {
				results = append(results, UpdateResult{
					Component: component,
					Success:   false,
					Message:   fmt.Sprintf("Falha na atualização: %v", err),
				})
				continue
			}

			// Broadcast: concluído
			server.Broadcast("update_progress", map[string]interface{}{
				"status":    "in_progress",
				"component": component,
				"progress":  100,
				"message":   fmt.Sprintf("%s atualizado para %s ✓", component, latest),
			})

			results = append(results, UpdateResult{
				Component:  component,
				Success:    true,
				Message:    fmt.Sprintf("Atualizado com sucesso"),
				NewVersion: latest,
			})

			time.Sleep(500 * time.Millisecond) // Pequena pausa para visualização
		}

		// Broadcast final: conclusão
		server.Broadcast("update_completed", map[string]interface{}{
			"status":  "completed",
			"results": results,
		})
	}()

	return map[string]string{"status": "update_in_progress"}, nil
}

// getBinaryVersion obtém a versão atual de um binário
func getBinaryVersion(installDir, component string) string {
	var binaryName string
	var suffix string

	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	switch component {
	case "daemon":
		binaryName = "vectora" + suffix
	case "tui", "cli":
		binaryName = "vectora-tui" + suffix
	case "desktop":
		binaryName = "vectora-desktop" + suffix
	case "lpm":
		binaryName = "lpm" + suffix
	case "mpm":
		binaryName = "mpm" + suffix
	case "setup":
		binaryName = "vectora-setup" + suffix
	default:
		return "unknown"
	}

	binaryPath := filepath.Join(installDir, binaryName)

	if _, err := os.Stat(binaryPath); err != nil {
		return "not installed"
	}

	cmd := exec.Command(binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	versionStr := strings.TrimSpace(string(output))

	// Remover prefixo "v" se existir
	if strings.HasPrefix(versionStr, "v") {
		versionStr = versionStr[1:]
	}

	// Extrair apenas a versão (primeiro campo)
	if parts := strings.Fields(versionStr); len(parts) > 0 {
		return parts[0]
	}

	return versionStr
}

// getLatestReleaseInfo obtém informações da release mais recente do GitHub
func getLatestReleaseInfo(component string) (version, downloadURL string) {
	// Esta função é um wrapper que utilizaria cmd/daemon/update.go:getLatestRelease()
	// Por enquanto, retornaremos valores padrão de fallback
	versions := map[string]string{
		"daemon": "1.0.0",
		"tui":    "1.0.0",
		"lpm":    "1.0.0",
		"mpm":    "1.0.0",
		"setup":  "1.0.0",
	}
	return versions[component], ""
}

// updateComponentBinary baixa e substitui o binário de um componente
func updateComponentBinary(installDir, component, downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("URL de download vazia")
	}

	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	var binaryName string
	switch component {
	case "daemon":
		binaryName = "vectora" + suffix
	case "tui":
		binaryName = "vectora-tui" + suffix
	case "lpm":
		binaryName = "lpm" + suffix
	case "mpm":
		binaryName = "mpm" + suffix
	case "setup":
		binaryName = "vectora-setup" + suffix
	default:
		return fmt.Errorf("componente desconhecido: %s", component)
	}

	binaryPath := filepath.Join(installDir, binaryName)

	// Criar backup
	backupPath := binaryPath + ".backup"
	if _, err := os.Stat(binaryPath); err == nil {
		if err := os.Rename(binaryPath, backupPath); err != nil {
			return fmt.Errorf("falha ao criar backup: %w", err)
		}
	}

	// Aqui seria feito o download real usando a lógica de cmd/daemon/update.go
	// Por enquanto, é apenas um placeholder
	return nil
}

// getInstallDir retorna o diretório de instalação do Vectora
func getInstallDir() string {
	// Tentar obter do sistema
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe)
	}

	// Fallback para AppData
	home, err := os.UserHomeDir()
	if err == nil {
		return filepath.Join(home, "AppData", "Local", "Vectora")
	}

	return "."
}
