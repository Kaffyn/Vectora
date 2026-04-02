package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Kaffyn/Vectora/src/core/domain"
	"github.com/Kaffyn/Vectora/src/core/server"
	"github.com/Kaffyn/Vectora/src/lib/git"
)

// setupACP cria um environment de teste para o ACPHandler.
func setupACP(t *testing.T) (*server.ACPHandler, string) {
	dir, _ := os.MkdirTemp("", "acp-test-*")
	// Mock git init
	gitInit := git.NewGitBridge(dir)
	// (Simular init git se necessário ou usar o real se possível no ambiente)

	h := server.NewACPHandler(gitInit)
	return h, dir
}

// TestACP_HappyPath_AutoApproval (100%): Testar auto-aprovação de tarefas Read no level Guarded.
func TestACP_HappyPath_AutoApproval(t *testing.T) {
	h, dir := setupACP(t)
	defer os.RemoveAll(dir)

	task := domain.ACPTask{
		ID:      "t1",
		Type:    domain.ACPActionRead,
		Command: "ls",
		Status:  "pending",
	}
	body, _ := json.Marshal(task)
	req := httptest.NewRequest(http.MethodPost, "/api/acp/task", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	var resp domain.ACPTask
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "approved" && resp.Status != "success" {
		t.Errorf("Expected task to be auto-approved, got %s", resp.Status)
	}
}

// TestACP_Negative_ManualApproval (200%): Testar que tarefas Write NO level Guarded ficam Pendentes.
func TestACP_Negative_ManualApproval(t *testing.T) {
	h, dir := setupACP(t)
	defer os.RemoveAll(dir)

	task := domain.ACPTask{
		ID:      "t2",
		Type:    domain.ACPActionWrite,
		Command: "touch",
		Status:  "pending",
	}
	body, _ := json.Marshal(task)
	req := httptest.NewRequest(http.MethodPost, "/api/acp/task", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	var resp domain.ACPTask
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "pending" {
		t.Errorf("Expected write task to be 'pending' in Guarded mode, got %s", resp.Status)
	}
}

// TestACP_EdgeCase_YOLO (300%): Testar autonomia absoluta no modo YOLO.
func TestACP_EdgeCase_YOLO(t *testing.T) {
	h, dir := setupACP(t)
	defer os.RemoveAll(dir)

	// Simular mudança de nível para YOLO (Geralmente via settings, aqui simulado internamente)
	// Como o level é privado no handler e não tem setter exposto no teste (exceto via interface),
	// faremos o teste assumindo o comportamento esperado ou se tivermos o setter.
	// (Adicionarei o setter se necessário).
	h.SetPermissionLevel(domain.ACPLevelYOLO)

	task := domain.ACPTask{
		ID:        "t3",
		Type:      domain.ACPActionWrite,
		Command:   "echo",
		Arguments: []string{"hello"},
		Status:    "pending",
	}
	body, _ := json.Marshal(task)
	req := httptest.NewRequest(http.MethodPost, "/api/acp/task", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	h.ServeHTTP(w, req)

	var resp domain.ACPTask
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Status != "approved" && resp.Status != "success" {
		t.Errorf("Expected Write task to be auto-approved in YOLO mode, got %s", resp.Status)
	}
}
