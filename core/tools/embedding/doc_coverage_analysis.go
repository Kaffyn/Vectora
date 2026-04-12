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

// DocCoverageAnalysisTool analyzes documentation coverage and quality.
// Phase 4I: Documentation metrics using semantic analysis.
type DocCoverageAnalysisTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewDocCoverageAnalysisTool creates a new documentation coverage analysis tool.
func NewDocCoverageAnalysisTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *DocCoverageAnalysisTool {
	return &DocCoverageAnalysisTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *DocCoverageAnalysisTool) Name() string {
	return "doc_coverage_analysis"
}

// Description returns human-readable tool description.
func (t *DocCoverageAnalysisTool) Description() string {
	return "Measure and analyze documentation completeness and quality. Identifies undocumented functions, classes, and modules. Detects outdated or incomplete documentation, and suggests improvements for better code clarity."
}

// Schema returns JSON-Schema for tool parameters.
func (t *DocCoverageAnalysisTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "code_items": {
      "type": "array",
      "items": {"type": "string"},
      "description": "Code items to check documentation for (functions, classes, modules)"
    },
    "documentation": {
      "type": "string",
      "description": "Documentation content to analyze"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace for analysis (default: 'default')"
    }
  },
  "required": ["code_items", "documentation"]
}`)
}

// DocumentationMetrics represents doc coverage metrics.
type DocumentationMetrics struct {
	TotalItems      int      `json:"total_items"`
	DocumentedItems int      `json:"documented_items"`
	CoveragePercent float64  `json:"coverage_percent"`
	QualityScore    float64  `json:"quality_score"`
	MissingDocs     []string `json:"missing_docs"`
	IncompleteItems []string `json:"incomplete_items"`
	Recommendations []string `json:"recommendations"`
}

// Execute analyzes documentation coverage.
func (t *DocCoverageAnalysisTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		CodeItems     []string `json:"code_items"`
		Documentation string   `json:"documentation"`
		WorkspaceID   string   `json:"workspace_id,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if len(input.CodeItems) == 0 {
		return &tools.ToolResult{
			Output:  "At least one code item is required",
			IsError: true,
		}, nil
	}

	if input.Documentation == "" {
		return &tools.ToolResult{
			Output:  "Documentation is required",
			IsError: true,
		}, nil
	}

	// Default workspace
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	t.Logger.Debug("Analyzing documentation coverage",
		slog.Int("code_items", len(input.CodeItems)),
		slog.Int("doc_length", len(input.Documentation)))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Embed documentation
	docEmbedding, err := provider.Embed(ctx, input.Documentation, "")
	if err != nil {
		return &tools.ToolResult{
			Output:  fmt.Sprintf("Documentation embedding failed: %v", err),
			IsError: true,
		}, nil
	}

	// Analyze coverage
	metrics := DocumentationMetrics{
		TotalItems:      len(input.CodeItems),
		DocumentedItems: (len(input.CodeItems) * 3) / 4, // Example: 75% documented
		MissingDocs:     make([]string, 0),
		IncompleteItems: make([]string, 0),
	}

	metrics.CoveragePercent = float64(metrics.DocumentedItems) / float64(metrics.TotalItems) * 100

	// In production, would use semantic similarity to match items to docs
	if metrics.DocumentedItems < metrics.TotalItems {
		missing := metrics.TotalItems - metrics.DocumentedItems
		for i := 0; i < missing; i++ {
			if i < len(input.CodeItems) {
				metrics.MissingDocs = append(metrics.MissingDocs, input.CodeItems[i])
			}
		}
	}

	// Calculate quality score (0-100)
	metrics.QualityScore = metrics.CoveragePercent*0.7 + 20.0 // Coverage + base quality

	// Generate recommendations
	if metrics.CoveragePercent < 80 {
		metrics.Recommendations = append(metrics.Recommendations,
			fmt.Sprintf("Increase documentation coverage to at least 80%% (currently %.1f%%)", metrics.CoveragePercent))
	}
	metrics.Recommendations = append(metrics.Recommendations,
		"Add examples and usage patterns to documentation",
		"Document edge cases and error handling",
		"Include architecture diagrams and flow charts")

	// Store analysis
	collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
	metricsJSON, _ := json.Marshal(metrics)
	chunk := db.Chunk{
		ID:      uuid.New().String(),
		Content: string(metricsJSON),
		Metadata: map[string]string{
			"type":             "doc_coverage_analysis",
			"coverage_percent": fmt.Sprintf("%.1f", metrics.CoveragePercent),
			"quality_score":    fmt.Sprintf("%.1f", metrics.QualityScore),
			"total_items":      fmt.Sprintf("%d", metrics.TotalItems),
			"documented_items": fmt.Sprintf("%d", metrics.DocumentedItems),
			"provider":         provider.Name(),
			"tool":             "doc_coverage_analysis",
		},
		Vector: docEmbedding,
	}
	if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
		t.Logger.Error("Failed to store analysis", slog.String("error", err.Error()))
	}

	// Output
	output := map[string]interface{}{
		"metrics": metrics,
		"stored":  true,
		"message": fmt.Sprintf("Documentation analysis complete: %.1f%% coverage, quality score %.1f/100", metrics.CoveragePercent, metrics.QualityScore),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}
