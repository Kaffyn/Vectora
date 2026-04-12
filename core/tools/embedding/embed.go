package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/google/uuid"
)

// EmbedTool converts text content into embeddings using Vectora's configured LLM provider.
// Phase 4G: Core embedding tool exposed via MCP and ACP.
type EmbedTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewEmbedTool creates a new embedding tool instance.
func NewEmbedTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *EmbedTool {
	return &EmbedTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *EmbedTool) Name() string {
	return "embed"
}

// Description returns human-readable tool description.
func (t *EmbedTool) Description() string {
	return "Index text content into Vectora's vector database for semantic search. Use this to add files, code snippets, or documentation to the searchable knowledge base. Makes content discoverable via semantic queries without needing exact keywords."
}

// Schema returns JSON-Schema for tool parameters.
func (t *EmbedTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "content": {
      "type": "string",
      "description": "Text content to embed (max 50KB)"
    },
    "metadata": {
      "type": "object",
      "description": "Optional metadata (source, filename, type, etc)",
      "properties": {
        "source": {"type": "string"},
        "filename": {"type": "string"},
        "type": {"type": "string"},
        "language": {"type": "string"}
      }
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to store embedding (default: 'default')"
    }
  },
  "required": ["content"]
}`)
}

// Execute embeds text content and stores it in the vector database.
func (t *EmbedTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		Content     string                 `json:"content"`
		Metadata    map[string]interface{} `json:"metadata,omitempty"`
		WorkspaceID string                 `json:"workspace_id,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Invalid input: %v", err),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.Content == "" {
		return &tools.ToolResult{
			Output:  "Content is required",
			IsError: true,
		}, nil
	}

	if len(input.Content) > 50*1024 {
		return &tools.ToolResult{
			Output:  "Content exceeds maximum size (50KB)",
			IsError: true,
		}, nil
	}

	// Default workspace
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	// Get default LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Generate embedding
	t.Logger.Debug("Embedding content", slog.Int("content_length", len(input.Content)))
	embedding, err := provider.Embed(ctx, input.Content, "")
	if err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Embedding failed: %v", err),
			IsError: true,
		}, nil
	}

	// Generate chunk ID
	chunkID := uuid.New().String()

	// Store in vector database
	// Note: In production, would use tenant-specific collection
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)

	// Convert metadata to string map (required by VectorStore interface)
	strMetadata := make(map[string]string)
	if input.Metadata != nil {
		for k, v := range input.Metadata {
			strMetadata[k] = fmt.Sprintf("%v", v)
		}
	}
	strMetadata["provider"] = provider.Name()
	strMetadata["embedding_dim"] = fmt.Sprintf("%d", len(embedding))

	// Create chunk and store in ChromemDB via UpsertChunk
	chunk := db.Chunk{
		ID:       chunkID,
		Content:  input.Content,
		Metadata: strMetadata,
		Vector:   embedding,
	}
	if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Storage failed: %v", err),
			IsError: true,
		}, nil
	}

	t.Logger.Info("Embedding stored successfully",
		slog.String("chunk_id", chunkID),
		slog.String("provider", provider.Name()),
		slog.Int("embedding_dim", len(embedding)),
	)

	// Return result
	output := map[string]interface{}{
		"chunk_id":      chunkID,
		"embedding_dim": len(embedding),
		"metadata":      input.Metadata,
		"provider":      provider.Name(),
		"stored":        true,
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:   string(result),
		IsError:  false,
		Metadata: map[string]interface{}{"chunk_id": chunkID},
	}, nil
}
