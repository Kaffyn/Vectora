package tool

import (
	"context"
	"fmt"
	"os"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// ReadFolderTool implementa a interface Tool para listagem de arquivos e diretórios.
type ReadFolderTool struct{}

func NewReadFolderTool() *ReadFolderTool {
	return &ReadFolderTool{}
}

func (t *ReadFolderTool) Name() string {
	return "read_folder"
}

func (t *ReadFolderTool) Description() string {
	return "Lista todos os arquivos e subdiretórios dentro de um caminho fornecido. Use para explorar a estrutura do projeto."
}

func (t *ReadFolderTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *ReadFolderTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Caminho do diretório (ex: 'src/', '.').",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadFolderTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	path, ok := args["path"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'path' é obrigatório e deve ser string")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler diretório: %w", err)
	}

	results := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, _ := entry.Info()
		results = append(results, map[string]interface{}{
			"name":  entry.Name(),
			"isDir": entry.IsDir(),
			"size":  info.Size(),
		})
	}

	return map[string]interface{}{
		"path":    path,
		"entries": results,
	}, nil
}
