package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	vecos "github.com/Kaffyn/vectora/internal/os"
	"github.com/nu7hatch/gouuid"
)

// Snapshots de Arquivos Puros baseados na Root O.S
func backupFileLocally(path string) string {
	osMgr, _ := vecos.NewManager()
	baseDir, _ := osMgr.GetAppDataDir()
	backups := filepath.Join(baseDir, "backups")

	u, _ := uuid.NewV4()
	snapID := u.String()

	srcFile, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer srcFile.Close()

	dstPath := filepath.Join(backups, snapID+"_"+filepath.Base(path))
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return ""
	}
	defer dstFile.Close()

	io.Copy(dstFile, srcFile)
	return snapID
}

// ----------------------------------------------------
// 1. Tool: read_file
// ----------------------------------------------------
type ReadFileTool struct{}

func (t *ReadFileTool) Name() string        { return "read_file" }
func (t *ReadFileTool) Description() string { return "Lê integralmente o conteúdo em texto de um arquivo local." }
func (t *ReadFileTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *ReadFileTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}
	return ToolResult{Output: string(data)}, nil
}

// ----------------------------------------------------
// 2. Tool: write_file
// ----------------------------------------------------
type WriteFileTool struct{}

func (t *WriteFileTool) Name() string        { return "write_file" }
func (t *WriteFileTool) Description() string { return "Cria ou sobrescreve por absoluto o conteúdo de um arquivo existente." }
func (t *WriteFileTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"},"content":{"type":"string"}},"required":["path","content"]}`)
}
func (t *WriteFileTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	content, _ := args["content"].(string)

	snapID := ""
	if _, err := os.Stat(path); err == nil {
		snapID = backupFileLocally(path)
	}

	os.MkdirAll(filepath.Dir(path), 0755)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}
	return ToolResult{Output: "Arquivo escrito com sucesso. (SnapID: " + snapID + ")", SnapshotID: snapID}, nil
}

// ----------------------------------------------------
// 3. Tool: read_folder
// ----------------------------------------------------
type ReadFolderTool struct{}

func (t *ReadFolderTool) Name() string        { return "read_folder" }
func (t *ReadFolderTool) Description() string { return "Lista arquivos e pastas irmãs dentro de um diretório base." }
func (t *ReadFolderTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"}},"required":["path"]}`)
}
func (t *ReadFolderTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	entries, err := os.ReadDir(path)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}
	var dirs []string
	for _, e := range entries {
		kind := "[Arquivo]"
		if e.IsDir() {
			kind = "[Pasta]"
		}
		dirs = append(dirs, fmt.Sprintf("%s %s", kind, e.Name()))
	}
	return ToolResult{Output: strings.Join(dirs, "\n")}, nil
}

// ----------------------------------------------------
// 4. Tool: edit
// ----------------------------------------------------
type EditTool struct{}

func (t *EditTool) Name() string        { return "edit" }
func (t *EditTool) Description() string { return "Substitui um bloco sequencial cirurgico de texto em um arquivo." }
func (t *EditTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"path":{"type":"string"},"target":{"type":"string"},"replacement":{"type":"string"}},"required":["path","target","replacement"]}`)
}
func (t *EditTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	path, _ := args["path"].(string)
	target, _ := args["target"].(string)
	replace, _ := args["replacement"].(string)

	data, err := os.ReadFile(path)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}

	snapID := backupFileLocally(path)

	strData := string(data)
	if !strings.Contains(strData, target) {
		return ToolResult{IsError: true, Output: "Alvo exato não encontrado. Furtividade de edição cancelada."}, nil
	}

	newData := strings.Replace(strData, target, replace, 1)
	err = os.WriteFile(path, []byte(newData), 0644)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}

	return ToolResult{Output: "Múltiplas linhas editadas com sucesso.", SnapshotID: snapID}, nil
}
