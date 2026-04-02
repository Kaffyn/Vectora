package tool

import (
	"context"

	"github.com/Kaffyn/Vectora/src/core/domain"
)

// Tool define a interface universal para ferramentas agênticas no Vectora.
type Tool interface {
	Name() string
	Description() string
	Type() domain.ACPActionType // Novo: read, write, execute
	InputSchema() map[string]interface{}
	Execute(ctx context.Context, args map[string]interface{}) (interface{}, error)
}
