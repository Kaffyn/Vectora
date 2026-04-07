package desktop

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// IndexClient gerencia comunicação com o Index Service remoto (gRPC)
type IndexClient struct {
	conn   *grpc.ClientConn
	addr   string
	client interface{} // Será IndexServiceClient após geração protobuf
}

// NewIndexClient cria um novo cliente para o Index Service
func NewIndexClient(addr string) (*IndexClient, error) {
	// Conectar ao Index Service via gRPC
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Index Service: %w", err)
	}

	client := &IndexClient{
		conn:   conn,
		addr:   addr,
		client: nil, // Será inicializado após protobuf generation
	}

	return client, nil
}

// CreateWorkspace cria um novo workspace remoto
func (c *IndexClient) CreateWorkspace(ctx context.Context, name, ownerID string) (map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado após protobuf
	// Por enquanto, retornar estrutura mock
	return map[string]interface{}{
		"id":        fmt.Sprintf("ws_%d", time.Now().Unix()),
		"name":      name,
		"owner_id":  ownerID,
		"created_at": time.Now().String(),
	}, nil
}

// GetWorkspace obtém workspace remoto por ID
func (c *IndexClient) GetWorkspace(ctx context.Context, workspaceID string) (map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return map[string]interface{}{
		"id": workspaceID,
	}, nil
}

// CreateIndex cria um novo índice remoto
func (c *IndexClient) CreateIndex(ctx context.Context, workspaceID, name, description string) (map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return map[string]interface{}{
		"id":             fmt.Sprintf("idx_%d", time.Now().Unix()),
		"workspace_id":   workspaceID,
		"name":           name,
		"description":    description,
		"document_count": 0,
		"size_bytes":     0,
		"created_at":     time.Now().String(),
	}, nil
}

// ListIndexes lista índices remotos
func (c *IndexClient) ListIndexes(ctx context.Context, workspaceID string) ([]map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return []map[string]interface{}{}, nil
}

// GetIndex obtém índice remoto por ID
func (c *IndexClient) GetIndex(ctx context.Context, indexID string) (map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return map[string]interface{}{
		"id": indexID,
	}, nil
}

// CreateUploadSession cria uma sessão de upload via navegador
// Retorna uma URL que pode ser aberta no navegador para upload
func (c *IndexClient) CreateUploadSession(ctx context.Context, indexID, workspaceID string, documentCount int) (string, error) {
	if c.conn == nil {
		return "", fmt.Errorf("Index Service desconectado")
	}

	// Esta função seria chamada no RPC CreateUploadSession
	// Por enquanto, gerar URL mock
	sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
	uploadURL := fmt.Sprintf("https://index.vectora.com/upload/%s", sessionID)

	return uploadURL, nil
}

// GetUploadStatus obtém status de uma sessão de upload
func (c *IndexClient) GetUploadStatus(ctx context.Context, sessionID string) (map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return map[string]interface{}{
		"session_id": sessionID,
		"status":     "pending", // pending, uploading, processing, completed, failed
		"progress":   0,
	}, nil
}

// SearchDocuments busca documentos no índice remoto
func (c *IndexClient) SearchDocuments(ctx context.Context, indexID, workspaceID, query string, topK int) ([]map[string]interface{}, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("Index Service desconectado")
	}

	// TODO: Usar gRPC stub gerado
	return []map[string]interface{}{}, nil
}

// Close fecha a conexão gRPC
func (c *IndexClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected verifica se está conectado
func (c *IndexClient) IsConnected() bool {
	return c.conn != nil
}
