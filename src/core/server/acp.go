package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"sync"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/lib/git"
)

// ACPHandler gerencia as solicitações autônomas da IA e a ponte Git.
type ACPHandler struct {
	gitBridge *git.GitBridge
	level     domain.ACPPermissionLevel
	tasks     map[string]*domain.ACPTask
	mu        sync.RWMutex
}

func NewACPHandler(gitBridge *git.GitBridge) *ACPHandler {
	return &ACPHandler{
		gitBridge: gitBridge,
		level:     domain.ACPLevelGuarded, // Padrão seguro
		tasks:     make(map[string]*domain.ACPTask),
	}
}

// SetPermissionLevel altera o nível de autonomia em tempo real.
func (h *ACPHandler) SetPermissionLevel(level domain.ACPPermissionLevel) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.level = level
}

// Garante que ACPHandler implementa ACPServer
var _ domain.ACPServer = (*ACPHandler)(nil)

// RequestPermission solicita permissão para uma tarefa.
func (h *ACPHandler) RequestPermission(ctx context.Context, task *domain.ACPTask) (bool, error) {
	approved := h.checkAutoApproval(*task)
	if approved {
		task.Status = "approved"
		return true, nil
	}
	task.Status = "pending"

	h.mu.Lock()
	h.tasks[task.ID] = task
	h.mu.Unlock()

	return false, nil
}

// Execute executa uma tarefa aprovada.
func (h *ACPHandler) Execute(ctx context.Context, task *domain.ACPTask) error {
	// Se for escrita ou execução, criamos um Snapshot antes
	if task.Type == domain.ACPActionWrite || task.Type == domain.ACPActionExecute {
		if _, err := h.gitBridge.CreateSnapshot(fmt.Sprintf("Auto-task: %s", task.Command)); err != nil {
			return fmt.Errorf("failed to create snapshot: %w", err)
		}
	}

	cmd := exec.CommandContext(ctx, task.Command, task.Arguments...)
	cmd.Dir = task.WorkDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		task.Status = "failed"
		return fmt.Errorf("command execution failed (output: %s): %w", string(output), err)
	}

	task.Status = "success"
	return nil
}

func (h *ACPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// 1. POST /api/acp/task - Solicitar execução
	if r.URL.Path == "/api/acp/task" && r.Method == http.MethodPost {
		var task domain.ACPTask
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		approved, _ := h.RequestPermission(r.Context(), &task)
		if approved {
			if err := h.Execute(r.Context(), &task); err != nil {
				// Erro já tratado no Execute
			}
		}

		json.NewEncoder(w).Encode(task)
		return
	}

	// 2. GET /api/acp/pending - Listar pendências
	if r.URL.Path == "/api/acp/pending" && r.Method == http.MethodGet {
		pending := make([]*domain.ACPTask, 0)
		h.mu.RLock()
		for _, t := range h.tasks {
			if t.Status == "pending" {
				pending = append(pending, t)
			}
		}
		h.mu.RUnlock()
		json.NewEncoder(w).Encode(pending)
		return
	}

	// 3. POST /api/acp/undo - Rollback total (Snapshot Git)
	if r.URL.Path == "/api/acp/undo" && r.Method == http.MethodPost {
		var req struct {
			Hash string `json:"hash"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		if err := h.gitBridge.Rollback(req.Hash); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Error(w, "ACP Endpoint Not Found", http.StatusNotFound)
}

func (h *ACPHandler) checkAutoApproval(task domain.ACPTask) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.level == domain.ACPLevelYOLO {
		return true
	}
	if h.level == domain.ACPLevelGuarded {
		// Ações de leitura (Guarded) são auto-aprovadas
		if task.Type == domain.ACPActionRead {
			return true
		}
	}
	// Level AskAny nunca auto-aprova nada
	return false
}
