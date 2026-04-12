package mcp

import (
	"context"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
)

// VectoraMCPServer implementa o Vectora como um servidor MCP (Model Context Protocol).
// Quando invocado como sub-agente, oferece ferramentas e contexto via MCP.
//
// Phase 4G: Core Embedding Tools para MCP exposure
// Phase 7D: MCP Server em Core para funcionar como sub-agent
// Exemplo: quando Claude Code invoca Vectora Core como sub-agent
type VectoraMCPServer struct {
	name           string
	version        string
	kvStore        db.KVStore
	vecStore       db.VectorStore
	llmRouter      *llm.Router
	msgService     *llm.MessageService
	logger         *slog.Logger
	embeddingTools map[string]interface{} // Phase 4G: Registered embedding tools
}

// NewVectoraMCPServer cria um novo servidor MCP do Vectora.
func NewVectoraMCPServer(
	name, version string,
	kvStore db.KVStore,
	vecStore db.VectorStore,
	llmRouter *llm.Router,
	msgService *llm.MessageService,
	logger *slog.Logger,
) *VectoraMCPServer {
	return &VectoraMCPServer{
		name:       name,
		version:    version,
		kvStore:    kvStore,
		vecStore:   vecStore,
		llmRouter:  llmRouter,
		msgService: msgService,
		logger:     logger,
	}
}

// GetTools retorna as ferramentas expostas pelo servidor MCP.
func (s *VectoraMCPServer) GetTools() []Tool {
	return []Tool{
		{
			Name:        "workspace.query",
			Description: "Consultar e buscar em um workspace usando busca semântica",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Consulta de busca",
					},
					"workspace": map[string]interface{}{
						"type":        "string",
						"description": "ID do workspace",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "workspace.embed",
			Description: "Indexar um arquivo ou diretório em um workspace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Caminho do arquivo ou diretório",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "chat.send",
			Description: "Enviar mensagem em uma conversa",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"conversation_id": map[string]interface{}{
						"type":        "string",
						"description": "ID da conversa",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Mensagem a enviar",
					},
				},
				"required": []string{"conversation_id", "message"},
			},
		},
	}
}

// Tool representa uma ferramenta MCP.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// CallTool executa uma ferramenta MCP.
func (s *VectoraMCPServer) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (interface{}, error) {
	s.logger.Debug("Chamando ferramenta MCP",
		slog.String("tool", name),
	)

	// TODO: Implementar lógica de execução de ferramentas
	// Por enquanto, retorna uma resposta stub
	return map[string]interface{}{
		"status": "success",
		"tool":   name,
	}, nil
}
