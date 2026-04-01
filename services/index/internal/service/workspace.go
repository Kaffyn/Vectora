package service

import (
	"context"
	"fmt"
	"time"
)

// Workspace representa um espaço de trabalho
type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	OwnerID   string    `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateWorkspace cria um novo workspace
func (s *Service) CreateWorkspace(ctx context.Context, name, ownerID string) (*Workspace, error) {
	if name == "" {
		return nil, fmt.Errorf("nome do workspace é obrigatório")
	}
	if ownerID == "" {
		return nil, fmt.Errorf("owner_id é obrigatório")
	}

	workspace := &Workspace{
		Name:      name,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO workspaces (name, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := s.db.GetConnection().QueryRowContext(
		ctx, query,
		workspace.Name, workspace.OwnerID,
		workspace.CreatedAt, workspace.UpdatedAt,
	).Scan(&workspace.ID, &workspace.CreatedAt, &workspace.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspace obtém um workspace por ID
func (s *Service) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace_id é obrigatório")
	}

	workspace := &Workspace{}
	query := `
		SELECT id, name, owner_id, created_at, updated_at
		FROM workspaces
		WHERE id = $1
	`

	err := s.db.GetConnection().QueryRowContext(ctx, query, workspaceID).Scan(
		&workspace.ID, &workspace.Name, &workspace.OwnerID,
		&workspace.CreatedAt, &workspace.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao obter workspace: %w", err)
	}

	return workspace, nil
}

// ListWorkspaces lista workspaces de um owner com paginação
func (s *Service) ListWorkspaces(ctx context.Context, ownerID string, page, pageSize int) ([]Workspace, error) {
	if ownerID == "" {
		return nil, fmt.Errorf("owner_id é obrigatório")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	query := `
		SELECT id, name, owner_id, created_at, updated_at
		FROM workspaces
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := s.db.GetConnection().QueryContext(ctx, query, ownerID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		workspace := Workspace{}
		if err := rows.Scan(
			&workspace.ID, &workspace.Name, &workspace.OwnerID,
			&workspace.CreatedAt, &workspace.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao fazer scan de workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar workspaces: %w", err)
	}

	return workspaces, nil
}

// DeleteWorkspace deleta um workspace
func (s *Service) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	if workspaceID == "" {
		return fmt.Errorf("workspace_id é obrigatório")
	}

	query := `DELETE FROM workspaces WHERE id = $1`
	result, err := s.db.GetConnection().ExecContext(ctx, query, workspaceID)
	if err != nil {
		return fmt.Errorf("erro ao deletar workspace: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar resultado de delete: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workspace não encontrado")
	}

	return nil
}
