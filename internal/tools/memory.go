package tools

import (
	"context"
	"encoding/json"

	"github.com/Kaffyn/vectora/internal/db"
)

// Tool vinculável à memória perene (No BBolt)
type SaveMemoryTool struct {
	KV db.KVStore
}

func (t *SaveMemoryTool) Name() string        { return "save_memory" }
func (t *SaveMemoryTool) Description() string { return "Armazena instrucoes base ineditas sobre o ambiente na memória crônica permanente." }
func (t *SaveMemoryTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"key":{"type":"string"},"value":{"type":"string"}},"required":["key","value"]}`)
}
func (t *SaveMemoryTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	if t.KV == nil {
		return ToolResult{IsError: true, Output: "Banco Bbolt KVStore Indisponível em Runtime."}, nil
	}

	key, _ := args["key"].(string)
	val, _ := args["value"].(string)

	err := t.KV.Set(ctx, "memories", key, []byte(val))
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}

	return ToolResult{Output: "Sínapse guardada! (Isso permacerá comigo para sempre)"}, nil
}

// Tool de Fragmentação Temporal (Agente Meta)
type PlanModeTool struct{}

func (t *PlanModeTool) Name() string        { return "enter_plan_mode" }
func (t *PlanModeTool) Description() string { return "Ativa quebra sub-tarefa. Utilize quando enfrentar um requisito gigantesco de engenharia de software." }
func (t *PlanModeTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"plan_description":{"type":"string"}},"required":["plan_description"]}`)
}
func (t *PlanModeTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	plan, _ := args["plan_description"].(string)
	return ToolResult{Output: "Planejamento inicializado localmente e emitido via IPC p/ a interface: \n" + plan}, nil
}
