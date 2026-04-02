package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/core/rag"
	"github.com/Kaffyn/Vectora/src/core/tool"
)

type MCPHandler struct {
	searchSvc  *rag.SearchService
	acpHandler *ACPHandler
	registry   *tool.Registry
}

func NewMCPHandler(searchSvc *rag.SearchService, acp *ACPHandler, registry *tool.Registry) *MCPHandler {
	return &MCPHandler{
		searchSvc:  searchSvc,
		acpHandler: acp,
		registry:   registry,
	}
}

func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if (r.URL.Path == "/api/mcp/tools" || r.URL.Path == "/api/mcp/list_tools") && r.Method == http.MethodGet {
		tools := h.registry.List()
		mcpTools := make([]interface{}, 0, len(tools))
		for _, t := range tools {
			mcpTools = append(mcpTools, map[string]interface{}{
				"name":         t.Name(),
				"description":  t.Description(),
				"input_schema": t.InputSchema(),
			})
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": mcpTools,
		})
		return
	}

	if (r.URL.Path == "/api/mcp/tools/call" || r.URL.Path == "/api/mcp/call_tool") && r.Method == http.MethodPost {
		var body struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}

		// Buscar ferramenta no Registry
		t, ok := h.registry.Get(body.Name)
		if !ok {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		// Fluxo ACP: Verificar se a ferramenta exige aprovação
		task := &domain.ACPTask{
			ID:      fmt.Sprintf("mcp_%d", r.Context().Value("request_id")), // Mock ID, implementar real
			Type:    t.Type(),
			Command: t.Name(),
			Status:  "pending",
		}

		approved, err := h.acpHandler.RequestPermission(r.Context(), task)
		if err != nil {
			http.Error(w, fmt.Sprintf("ACP Permission Error: %v", err), http.StatusInternalServerError)
			return
		}

		if !approved {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": "Ação pendente de aprovação manual no Tray do Vectora.",
					},
				},
				"isError": false,
				"status":  "pending",
				"taskId":  task.ID,
			})
			return
		}

		// Snapshot de segurança pré-execução
		if task.Type == domain.ACPActionWrite || task.Type == domain.ACPActionExecute {
			if _, snapErr := h.acpHandler.gitBridge.CreateSnapshot(fmt.Sprintf("Tool: %s", t.Name())); snapErr != nil {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"content": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": fmt.Sprintf("Segurança ACP: Falha ao criar snapshot preventivo antes de executar a ferramenta: %v", snapErr),
						},
					},
					"isError": true,
				})
				return
			}
		}

		// Executar ferramenta se aprovada
		result, err := t.Execute(r.Context(), body.Arguments)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"content": []interface{}{
					map[string]interface{}{
						"type": "text",
						"text": fmt.Sprintf("Erro na execução: %v", err),
					},
				},
				"isError": true,
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []interface{}{
				map[string]interface{}{
					"type": "text",
					"text": fmt.Sprintf("%v", result),
				},
			},
		})
		return
	}

	http.Error(w, `{"error":"mcp endpoint not found"}`, http.StatusNotFound)
}
