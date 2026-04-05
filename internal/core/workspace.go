package core

import (
	"context"
	"fmt"
	"time"
)

type WorkspaceStatus string

const (
	StatusIdle     WorkspaceStatus = "idle"
	StatusIndexing WorkspaceStatus = "indexing"
	StatusDone     WorkspaceStatus = "done"
	StatusError    WorkspaceStatus = "error"
)

type Workspace struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"created_at"`
	IndexedAt   *time.Time        `json:"indexed_at,omitempty"`
	ChunkCount  int               `json:"chunk_count"`
	Status      WorkspaceStatus   `json:"status"`
}

type WorkspaceManager struct {
	workspaces map[string]*Workspace
}

func NewWorkspaceManager() *WorkspaceManager {
	return &WorkspaceManager{
		workspaces: make(map[string]*Workspace),
	}
}

func (wm *WorkspaceManager) Create(ctx context.Context, name, description string) (*Workspace, error) {
	id := fmt.Sprintf("ws_%d", time.Now().Unix())
	now := time.Now()

	ws := &Workspace{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		Status:      StatusIdle,
	}

	wm.workspaces[id] = ws
	return ws, nil
}

func (wm *WorkspaceManager) Get(ctx context.Context, id string) (*Workspace, error) {
	ws, ok := wm.workspaces[id]
	if !ok {
		return nil, fmt.Errorf("workspace not found: %s", id)
	}
	return ws, nil
}

func (wm *WorkspaceManager) List(ctx context.Context) ([]*Workspace, error) {
	result := make([]*Workspace, 0, len(wm.workspaces))
	for _, ws := range wm.workspaces {
		result = append(result, ws)
	}
	return result, nil
}

func (wm *WorkspaceManager) Delete(ctx context.Context, id string) error {
	delete(wm.workspaces, id)
	return nil
}

func (wm *WorkspaceManager) SetIndexing(ctx context.Context, id string) error {
	ws, err := wm.Get(ctx, id)
	if err != nil {
		return err
	}
	ws.Status = StatusIndexing
	return nil
}

func (wm *WorkspaceManager) SetIndexed(ctx context.Context, id string, chunkCount int) error {
	ws, err := wm.Get(ctx, id)
	if err != nil {
		return err
	}
	ws.Status = StatusDone
	ws.ChunkCount = chunkCount
	now := time.Now()
	ws.IndexedAt = &now
	return nil
}
