package tool

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// FindFilesTool implementa a interface Tool para localização de arquivos via padrão glob.
type FindFilesTool struct{}

func NewFindFilesTool() *FindFilesTool {
	return &FindFilesTool{}
}

func (t *FindFilesTool) Name() string {
	return "find_files"
}

func (t *FindFilesTool) Description() string {
	return "Localiza arquivos no sistema que correspondam a um padrão glob (ex: '**/*.go', 'src/core/*.ts'). Útil para encontrar arquivos específicos em projetos grandes."
}

func (t *FindFilesTool) Type() domain.ACPActionType {
	return domain.ACPActionRead
}

func (t *FindFilesTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "O padrão glob de busca (ex: 'cmd/**/*.go').",
			},
		},
		"required": []string{"pattern"},
	}
}

func (t *FindFilesTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	pattern, ok := args["pattern"].(string)
	if !ok {
		return nil, fmt.Errorf("argumento 'pattern' é obrigatório e deve ser string")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("falha ao executar glob: %w", err)
	}

	return map[string]interface{}{
		"pattern": pattern,
		"matches": matches,
		"count":   len(matches),
	}, nil
}
