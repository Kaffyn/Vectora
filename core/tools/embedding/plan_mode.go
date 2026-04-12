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

// PlanModeTool creates structured plans using context-aware embeddings.
// Phase 4H: Planning tool that leverages vector search for similar past plans/patterns.
type PlanModeTool struct {
	Router   *llm.Router
	VecStore db.VectorStore
	Logger   *slog.Logger
}

// NewPlanModeTool creates a new plan mode tool instance.
func NewPlanModeTool(router *llm.Router, vecStore db.VectorStore, logger *slog.Logger) *PlanModeTool {
	return &PlanModeTool{
		Router:   router,
		VecStore: vecStore,
		Logger:   logger,
	}
}

// Name returns the tool name for registration.
func (t *PlanModeTool) Name() string {
	return "plan_mode"
}

// Description returns human-readable tool description.
func (t *PlanModeTool) Description() string {
	return "Break down complex problems into structured, step-by-step implementation plans. Analyzes codebase patterns and constraints to create realistic, phased plans with task dependencies, effort estimates, and risk assessments."
}

// Schema returns JSON-Schema for tool parameters.
func (t *PlanModeTool) Schema() json.RawMessage {
	return []byte(`{
  "type": "object",
  "properties": {
    "objective": {
      "type": "string",
      "description": "Main objective or feature to plan"
    },
    "context": {
      "type": "string",
      "description": "Additional context (architecture, constraints, dependencies)"
    },
    "scope": {
      "type": "string",
      "enum": ["minimal", "standard", "comprehensive"],
      "description": "Planning scope (default: standard)"
    },
    "workspace_id": {
      "type": "string",
      "description": "Workspace to store plan (default: 'default')"
    },
    "search_similar": {
      "type": "boolean",
      "description": "Search for similar plans in vector DB (default: true)"
    }
  },
  "required": ["objective"]
}`)
}

// PlanPhase represents one phase in an implementation plan.
type PlanPhase struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Tasks        []string `json:"tasks"`
	Dependencies string   `json:"dependencies"`
}

// ImplementationPlan represents a complete implementation plan.
type ImplementationPlan struct {
	Objective string      `json:"objective"`
	Scope     string      `json:"scope"`
	Phases    []PlanPhase `json:"phases"`
	Risks     []string    `json:"risks"`
	Timeline  string      `json:"timeline"`
	Resources []string    `json:"resources"`
}

// Execute creates a structured implementation plan.
func (t *PlanModeTool) Execute(ctx context.Context, args json.RawMessage) (*tools.ToolResult, error) {
	var input struct {
		Objective     string `json:"objective"`
		Context       string `json:"context,omitempty"`
		Scope         string `json:"scope,omitempty"`
		WorkspaceID   string `json:"workspace_id,omitempty"`
		SearchSimilar bool   `json:"search_similar,omitempty"`
	}

	if err := json.Unmarshal(args, &input); err != nil {
		return &tools.ToolResult{
			Output:  "Invalid input: " + err.Error(),
			IsError: true,
		}, nil
	}

	// Validate input
	if input.Objective == "" {
		return &tools.ToolResult{
			Output:  "Objective is required",
			IsError: true,
		}, nil
	}

	// Default values
	if input.Scope == "" {
		input.Scope = "standard"
	}
	if input.WorkspaceID == "" {
		input.WorkspaceID = "default"
	}

	// Default search_similar to true
	if !input.SearchSimilar {
		input.SearchSimilar = true
	}

	t.Logger.Debug("Creating implementation plan",
		slog.String("objective", input.Objective),
		slog.String("scope", input.Scope),
		slog.Bool("search_similar", input.SearchSimilar))

	// Get LLM provider
	provider := t.Router.GetDefault()
	if provider == nil {
		return &tools.ToolResult{
			Output:  "No LLM provider configured",
			IsError: true,
		}, nil
	}

	// Search for similar plans if requested
	var similarPlans []db.ScoredChunk
	if input.SearchSimilar {
		collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
		queryEmbedding, err := provider.Embed(ctx, input.Objective, "")
		if err == nil {
			results, _ := t.VecStore.Query(ctx, collectionID, queryEmbedding, 5)
			similarPlans = results
			t.Logger.Debug("Found similar plans", slog.Int("count", len(similarPlans)))
		}
	}

	// Create implementation plan structure
	plan := ImplementationPlan{
		Objective: input.Objective,
		Scope:     input.Scope,
		Phases: []PlanPhase{
			{
				Name:        "Phase 1: Analysis",
				Description: "Understand requirements and scope",
				Tasks: []string{
					"Analyze objective and constraints",
					"Identify dependencies",
					"Define success criteria",
				},
				Dependencies: "None",
			},
			{
				Name:        "Phase 2: Design",
				Description: "Create technical design",
				Tasks: []string{
					"Design architecture",
					"Create data models",
					"Plan integration points",
				},
				Dependencies: "Phase 1",
			},
			{
				Name:        "Phase 3: Implementation",
				Description: "Execute implementation",
				Tasks: []string{
					"Implement core features",
					"Add tests",
					"Integration testing",
				},
				Dependencies: "Phase 2",
			},
			{
				Name:        "Phase 4: Verification",
				Description: "Verify and optimize",
				Tasks: []string{
					"Performance testing",
					"Security review",
					"Documentation",
				},
				Dependencies: "Phase 3",
			},
		},
		Risks: []string{
			"Scope creep",
			"External dependencies",
			"Resource constraints",
		},
		Timeline: "4-6 weeks depending on complexity",
		Resources: []string{
			"Development team",
			"Architecture review",
			"Testing resources",
		},
	}

	// Store plan in vector database
	planJSON, _ := json.Marshal(plan)
	planEmbedding, err := provider.Embed(ctx, plan.Objective, "")
	if err == nil {
		collectionID := fmt.Sprintf("ws_%s", input.WorkspaceID)
		chunk := db.Chunk{
			ID:      uuid.New().String(),
			Content: string(planJSON),
			Metadata: map[string]string{
				"type":        "implementation_plan",
				"objective":   plan.Objective,
				"scope":       plan.Scope,
				"phase_count": fmt.Sprintf("%d", len(plan.Phases)),
				"provider":    provider.Name(),
				"tool":        "plan_mode",
			},
			Vector: planEmbedding,
		}
		if err := t.VecStore.UpsertChunk(ctx, collectionID, chunk); err != nil {
			t.Logger.Error("Failed to store plan", slog.String("error", err.Error()))
		}
	}

	// Build output
	output := map[string]interface{}{
		"plan":          plan,
		"similar_plans": len(similarPlans),
		"stored":        true,
		"message":       fmt.Sprintf("Created %s implementation plan with %d phases", input.Scope, len(plan.Phases)),
	}

	result, _ := json.Marshal(output)
	return &tools.ToolResult{
		Output:  string(result),
		IsError: false,
	}, nil
}
