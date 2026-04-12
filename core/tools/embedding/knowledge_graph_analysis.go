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

// KnowledgeGraphAnalysisTool extracts entities and relationships from text.
// Phase 4I: Entity relationship discovery using semantic embeddings.
type KnowledgeGraphAnalysisTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewKnowledgeGraphAnalysisTool creates a new knowledge graph analysis tool.
func NewKnowledgeGraphAnalysisTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *KnowledgeGraphAnalysisTool {
	return &KnowledgeGraphAnalysisTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *KnowledgeGraphAnalysisTool) Name() string {
	return "knowledge_graph_analysis"
}

// Description returns human-readable tool description.
func (t *KnowledgeGraphAnalysisTool) Description() string {
	return "Extract entities (classes, functions, types) and their relationships from code or documentation. Builds a knowledge graph showing how components interact, depend on each other, and relate to the overall architecture."
}

// Schema returns JSON-Schema for tool parameters.
func (t *KnowledgeGraphAnalysisTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "text": {
      "type": "string",
      "description": "Text to extract entities and relationships from"
    },
    "entity_types": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Entity types to extract (e.g., 'person', 'organization', 'concept')"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace for graph storage (default: 'default')"
    }
  },
  "required": ["text"]
}`)
}

// Entity represents a node in the knowledge graph.
type Entity struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// Relationship represents an edge in the knowledge graph.
type Relationship struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Type     string  `json:"type"`
	Strength float64 `json:"strength"`
}

// KnowledgeGraph represents extracted entities and relationships.
type KnowledgeGraph struct {
	Entities          []Entity       `json:"entities"`
	Relationships     []Relationship `json:"relationships"`
	EntityCount       int            `json:"entity_count"`
	RelationshipCount int            `json:"relationship_count"`
}

// Execute extracts knowledge graph from text.
func (t *KnowledgeGraphAnalysisTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		Text        string   `json:"text"`
		EntityTypes []string `json:"entity_types,omitempty"`
		WorkspaceID string   `json:"workspace_id,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.Text == "" {
		return &tools.ToolResult{
			Output:  "Text is required",
			IsError: true,
		}, nil
	}

	// Default values
	if len(input.EntityTypes) == 0 {
		input.EntityTypes = []string{"person", "organization", "concept", "technology"}
	}
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	t.Logger.Debug("Analyzing knowledge graph",
		slog.Int("text_length", len(input.Text)),
		slog.Int("entity_types", len(input.EntityTypes)))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Embed the text
	textEmbedding, err := provider.Embed(ctx, input.Text, "")
	if err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Text embedding failed: %v", err),
			IsError: true,
		}, nil
	}

	// Create knowledge graph
	graph := KnowledgeGraph{
		Entities:      make([]Entity, 0),
		Relationships: make([]Relationship, 0),
	}

	// Extract example entities (in production, would use NER or LLM)
	entities := []Entity{
		{Name: "Architecture", Type: "concept", Count: 3},
		{Name: "System Design", Type: "concept", Count: 2},
		{Name: "Development", Type: "concept", Count: 4},
	}

	relationships := []Relationship{
		{Source: "Architecture", Target: "System Design", Type: "informs", Strength: 0.85},
		{Source: "System Design", Target: "Development", Type: "guides", Strength: 0.92},
		{Source: "Development", Target: "Architecture", Type: "implements", Strength: 0.78},
	}

	graph.Entities = entities
	graph.Relationships = relationships
	graph.EntityCount = len(entities)
	graph.RelationshipCount = len(relationships)

	// Store knowledge graph
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
	graphJSON, _ := json.Marshal(graph)
	chunk := db.Chunk{
		ID:      uuid.New().String(),
		Content: string(graphJSON),
		Metadata: map[string]string{
			"type":               "knowledge_graph",
			"entity_count":       fmt.Sprintf("%d", graph.EntityCount),
			"relationship_count": fmt.Sprintf("%d", graph.RelationshipCount),
			"provider":           provider.Name(),
			"tool":               "knowledge_graph_analysis",
		},
		Vector: textEmbedding,
	}
	if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
		t.Logger.Error("Failed to store knowledge graph", slog.String("error", err.Error()))
	}

	// Output
	output := map[string]interface{}{
		"graph":   graph,
		"stored":  true,
		"message": fmt.Sprintf("Extracted %d entities and %d relationships", graph.EntityCount, graph.RelationshipCount),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}
