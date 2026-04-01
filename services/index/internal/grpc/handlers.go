package grpc

import (
	"context"

	"github.com/Kaffyn/Vectora-Index/internal/service"
)

// IndexServiceServer implementa a interface gerada pelo protobuf
// Este é um placeholder até que o código seja gerado pelo protoc
type IndexServiceServer struct {
	service *service.Service
	// UnimplementedIndexServiceServer será adicionado após geração do protobuf
}

// NewIndexServiceServer cria uma nova instância do servidor gRPC
func NewIndexServiceServer(svc *service.Service) *IndexServiceServer {
	return &IndexServiceServer{
		service: svc,
	}
}

// CreateWorkspace implementa o RPC CreateWorkspace
func (s *IndexServiceServer) CreateWorkspace(ctx context.Context, name, ownerID string) (*service.Workspace, error) {
	return s.service.CreateWorkspace(ctx, name, ownerID)
}

// GetWorkspace implementa o RPC GetWorkspace
func (s *IndexServiceServer) GetWorkspace(ctx context.Context, workspaceID string) (*service.Workspace, error) {
	return s.service.GetWorkspace(ctx, workspaceID)
}

// ListWorkspaces implementa o RPC ListWorkspaces
func (s *IndexServiceServer) ListWorkspaces(ctx context.Context, ownerID string, page, pageSize int) ([]service.Workspace, error) {
	return s.service.ListWorkspaces(ctx, ownerID, page, pageSize)
}

// CreateIndex implementa o RPC CreateIndex
func (s *IndexServiceServer) CreateIndex(ctx context.Context, workspaceID, name, description string) (*service.Index, error) {
	return s.service.CreateIndex(ctx, workspaceID, name, description)
}

// GetIndex implementa o RPC GetIndex
func (s *IndexServiceServer) GetIndex(ctx context.Context, indexID string) (*service.Index, error) {
	return s.service.GetIndex(ctx, indexID)
}

// ListIndexes implementa o RPC ListIndexes
func (s *IndexServiceServer) ListIndexes(ctx context.Context, workspaceID string) ([]service.Index, error) {
	return s.service.ListIndexes(ctx, workspaceID)
}

// DeleteIndex implementa o RPC DeleteIndex
func (s *IndexServiceServer) DeleteIndex(ctx context.Context, indexID string) error {
	return s.service.DeleteIndex(ctx, indexID)
}

// CreateDocument implementa o RPC para criar documento
func (s *IndexServiceServer) CreateDocument(ctx context.Context, indexID, workspaceID, filename string) (*service.Document, error) {
	return s.service.CreateDocument(ctx, indexID, workspaceID, filename)
}

// GetDocument implementa o RPC GetDocument
func (s *IndexServiceServer) GetDocument(ctx context.Context, documentID string) (*service.Document, error) {
	return s.service.GetDocument(ctx, documentID)
}

// ListDocuments implementa o RPC ListDocuments
func (s *IndexServiceServer) ListDocuments(ctx context.Context, indexID string) ([]service.Document, error) {
	return s.service.ListDocuments(ctx, indexID)
}

// DeleteDocument implementa o RPC DeleteDocument
func (s *IndexServiceServer) DeleteDocument(ctx context.Context, documentID string) error {
	return s.service.DeleteDocument(ctx, documentID)
}
