package mcp

import (
	"encoding/json"
	"net/http"

	"github.com/Kaffyn/Vectora/src/core/rag"
)

type MCPServer struct {
	searchService *rag.SearchService
}

func NewServer(search *rag.SearchService) *MCPServer {
	return &MCPServer{searchService: search}
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Method string          `json:"method"`
		Params json.RawMessage `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case "tools/call":
		var params toolCallParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			http.Error(w, "Invalid Params", http.StatusBadRequest)
			return
		}
		s.handleToolCall(w, r, params)
	default:
		http.Error(w, "Method Not Found", http.StatusNotFound)
	}
}

func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request, params toolCallParams) {
	ctx := r.Context()

	switch params.Name {
	case "query_engine_docs":
		var args struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(params.Arguments, &args); err != nil {
			json.NewEncoder(w).Encode(map[string]any{"error": "invalid arguments"})
			return
		}

		results, err := s.searchService.Search(ctx, args.Query)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"results": results})

	default:
		json.NewEncoder(w).Encode(map[string]any{"error": "tool not found"})
	}
}
