package service

import (
	"context"
	"fmt"
	"time"
)

// Document representa um documento no índice
type Document struct {
	ID          string    `json:"id"`
	IndexID     string    `json:"index_id"`
	WorkspaceID string    `json:"workspace_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	FileSize    int64     `json:"file_size"`
	UploadedBy  string    `json:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateDocument cria um novo documento
func (s *Service) CreateDocument(ctx context.Context, indexID, workspaceID, filename string) (*Document, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index_id é obrigatório")
	}
	if workspaceID == "" {
		return nil, fmt.Errorf("workspace_id é obrigatório")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename é obrigatório")
	}

	// Verificar se índice existe
	if _, err := s.GetIndex(ctx, indexID); err != nil {
		return nil, fmt.Errorf("índice não encontrado")
	}

	document := &Document{
		IndexID:     indexID,
		WorkspaceID: workspaceID,
		Filename:    filename,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO documents (index_id, workspace_id, filename, content_type, file_size, uploaded_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := s.db.GetConnection().QueryRowContext(
		ctx, query,
		document.IndexID, document.WorkspaceID, document.Filename,
		document.ContentType, document.FileSize, document.UploadedBy,
		document.CreatedAt, document.UpdatedAt,
	).Scan(&document.ID, &document.CreatedAt, &document.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar documento: %w", err)
	}

	return document, nil
}

// GetDocument obtém um documento por ID
func (s *Service) GetDocument(ctx context.Context, documentID string) (*Document, error) {
	if documentID == "" {
		return nil, fmt.Errorf("document_id é obrigatório")
	}

	document := &Document{}
	query := `
		SELECT id, index_id, workspace_id, filename, content_type, file_size, uploaded_by, created_at, updated_at
		FROM documents
		WHERE id = $1
	`

	err := s.db.GetConnection().QueryRowContext(ctx, query, documentID).Scan(
		&document.ID, &document.IndexID, &document.WorkspaceID, &document.Filename,
		&document.ContentType, &document.FileSize, &document.UploadedBy,
		&document.CreatedAt, &document.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("erro ao obter documento: %w", err)
	}

	return document, nil
}

// ListDocuments lista documentos de um índice
func (s *Service) ListDocuments(ctx context.Context, indexID string) ([]Document, error) {
	if indexID == "" {
		return nil, fmt.Errorf("index_id é obrigatório")
	}

	query := `
		SELECT id, index_id, workspace_id, filename, content_type, file_size, uploaded_by, created_at, updated_at
		FROM documents
		WHERE index_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.GetConnection().QueryContext(ctx, query, indexID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar documentos: %w", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		document := Document{}
		if err := rows.Scan(
			&document.ID, &document.IndexID, &document.WorkspaceID, &document.Filename,
			&document.ContentType, &document.FileSize, &document.UploadedBy,
			&document.CreatedAt, &document.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao fazer scan de documento: %w", err)
		}
		documents = append(documents, document)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar documentos: %w", err)
	}

	return documents, nil
}

// DeleteDocument deleta um documento
func (s *Service) DeleteDocument(ctx context.Context, documentID string) error {
	if documentID == "" {
		return fmt.Errorf("document_id é obrigatório")
	}

	query := `DELETE FROM documents WHERE id = $1`
	result, err := s.db.GetConnection().ExecContext(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("erro ao deletar documento: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erro ao verificar resultado de delete: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("documento não encontrado")
	}

	return nil
}
