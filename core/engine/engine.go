package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"vectora/core/llm"
	"vectora/core/policies"
	"vectora/core/storage"
	"vectora/core/tools"
)

// Engine é o orquestrador central do Vectora Core.
// Ele amarra a inteligência (LLM), a memória (Storage) e a ação (Tools).
type Engine struct {
	Storage  *storage.Engine
	LLM      *llm.Router
	Tools    *tools.Registry
	Guardian *policies.Guardian
	Status   string
}

func NewEngine(s *storage.Engine, l *llm.Router, t *tools.Registry, g *policies.Guardian) *Engine {
	return &Engine{
		Storage:  s,
		LLM:      l,
		Tools:    t,
		Guardian: g,
		Status:   "idle",
	}
}

// ExecuteTool orquestra a chamada de uma ferramenta validando-a via Guardian.
func (e *Engine) ExecuteTool(ctx context.Context, name string, args json.RawMessage) (*tools.ToolResult, error) {
	// A validação de path seguro já acontece dentro de cada ferramenta via Guardian injetado,
	// mas aqui podemos adicionar uma camada de log ou pré-verificação global.

	tool, ok := e.Tools.GetTool(name)
	if !ok {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	return tool.Execute(ctx, args)
}

// QueryRequest simplificado para o MVP.
type QueryRequest struct {
	Prompt      string
	WorkspaceID string
}

// StreamQuery (Skeleton) para ser implementado com a lógica de RAG real.
func (e *Engine) StreamQuery(ctx context.Context, query string, workspaceID string) (chan QueryChunk, error) {
	ch := make(chan QueryChunk)

	go func() {
		defer close(ch)
		// 1. Recuperar Contexto do Storage (RAG)
		// 2. Montar Prompt via LLM Factory
		// 3. Chamar LLM Provider
		// 4. Streamar tokens
		ch <- QueryChunk{Token: "RAG not fully wired yet", IsFinal: true}
	}()

	return ch, nil
}

type QueryChunk struct {
	Token   string
	Sources []string
	IsFinal bool
}

// ProcessQuery é a versão síncrona de StreamQuery para uso via IPC/CLI.
func (e *Engine) ProcessQuery(query string) (string, error) {
	ctx := context.Background()
	ch, err := e.StreamQuery(ctx, query, "default")
	if err != nil {
		return "", err
	}
	var result string
	for chunk := range ch {
		result += chunk.Token
	}
	return result, nil
}

func (e *Engine) StartIndexation() {
	e.Status = "indexing"
	// Pipeline de ingestão seria disparado aqui
}

func (e *Engine) GetStatus() string {
	return e.Status
}
