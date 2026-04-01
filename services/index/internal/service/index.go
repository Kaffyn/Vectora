package service

import (
	"context"
	"fmt"
	"time"
)

// Index representa um índice de documentos
type Index struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	DocumentCount  int       `json:"document_count"`
	SizeBytes      int64     `json:"size_bytes"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateIndex cria um novo índice
func (s *Service) CreateIndex(ctx context.Context, workspaceID, name, description string) (*Index, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace_id é obrigatório")
	}
	if name == "" {
		return nil, fmt.Errorf("nome do índice é obrigatório")
	}

	// Verificar se workspace existe
	if _, err := s.GetWorkspace(ctx, workspaceID); err != nil {
		return nil, fmt.Errorf("workspace não encontrado")
	}

	index := &Index{
		WorkspaceID:   workspaceID,
		Name:          name,
		Description:   description,
		DocumentCount: 0,
		SizeBytes:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	query := `
		INSERT INTO indexes (workspace_id, name, description, document_count, size_bytes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err := s.db.GetConnection().QueryRowContext(
		ctx, query,
		index.WorkspaceID, index.Name, index.Description,
		index.DocumentCount, index.SizeBytes,
		index.CreatedAt, index.UpdatedAt,
	).Scan(&index.ID, &index.CreatedAt, &index.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar índice: %w", err)
	}

	return index, nil
}

// GetIndex obtém um índice por ID
func (s *Service) GetIndex(ctx context.Context, indexID string) (*Index, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index_id é obrigatório")
	}

	index := &Index{}
	query := `
		SELECT id, workspace_id, name, description, document_count, size_bytes, created_at, updated_at
		FROM indexes
		WHERE id = $1
	`

	err := s.db.GetConnection().QueryRowContext(ctx, query, indexID).Scan(
		&index.ID, &index.WorkspaceID, &index.Name, &index.Description,
		&index.DocumentCount, &index.SizeBytes, &index.CreatedAt, &index.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao obter índice: %w", err)
	}

	return index, nil
}

// ListIndexes lista índices de um workspace
func (s *Service) ListIndexes(ctx context.Context, workspaceID string) ([]Index, error) {
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace_id é obrigatório")
	}

	query := `
		SELECT id, workspace_id, name, description, document_count, size_bytes, created_at, updated_at
		FROM indexes
		WHERE workspace_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.GetConnection().QueryContext(ctx, query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar índices: %w", err)
	}
	defer rows.Close()

	var indexes []Index
	for rows.Next() {
		index := Index{}
		if err := rows.Scan(
			&index.ID, &index.WorkspaceID, &index.Name, &index.Description,
			&index.DocumentCount, &index.SizeBytes, &index.CreatedAt, &index.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao fazer scan de índice: %w", err)
		}
		indexes = append(indexes, index)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar índices: %w", err)
	}

	return indexes, nil
}

// DeleteIndex deleta um índice
func (s *Service) DeleteIndex(ctx context.Context, indexID string) error {
	if indexID == "" {
		return fmt.Errorf("index_id é obrigatório")
	}

	query := `DELETE FROM indexes WHERE id = $1`
	result, err := s.db.GetConnection().ExecContext(ctx, query, indexID)
	if err != nil {
		return fmt.Errorf("erro ao deletar índice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar resultado de delete: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("índice não encontrado")
	}

	return nil
}
