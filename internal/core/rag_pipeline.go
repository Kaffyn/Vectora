package core

import (
	"context"
	"errors"

	"github.com/Kaffyn/vectora/internal/acp"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/llm"
)

// Pipeline orchestrates data between the LLM (Interface) and the VectorDB (Storage).
// It is instantiated and held by the SystemTray Daemon.
type Pipeline struct {
	LLM      llm.Provider
	VectorDB db.VectorStore
	KVStore  db.KVStore
	Agent    *acp.AgentContext
}

// Motherboard that couples all loose modules to inject queries.
func NewPipeline(provider llm.Provider, vectorStore db.VectorStore, kvStore db.KVStore) *Pipeline {
	return &Pipeline{
		LLM:      provider,
		VectorDB: vectorStore,
		KVStore:  kvStore,
		Agent:    acp.NewAgent(kvStore),
	}
}

type QueryRequest struct {
	WorkspaceID string
	Query       string
}

type QueryResponse struct {
	Answer  string
	Sources []db.ScoredChunk
}

// Validates the canonical RAG (Retrieval-Augmented Generation) of Vectora.
func (p *Pipeline) Query(ctx context.Context, req QueryRequest) (QueryResponse, error) {
	if p.LLM == nil || !p.LLM.IsConfigured() {
		return QueryResponse{}, errors.New("pipeline_err: LLM provider not injected or not configured in the SystemTray")
	}

	// 1. Native Embedding
	vector, err := p.LLM.Embed(ctx, req.Query)
	if err != nil {
		return QueryResponse{}, err
	}

	// 2. Nearest Neighbors in Chromem KNN Database
	// TOP K locked at 5 to avoid LLM context window bloating on low RAM environments.
	resChunks, err := p.VectorDB.Query(ctx, "ws_"+req.WorkspaceID, vector, 5)
	if err != nil {
		// Zero-Shot Fallback. If the collection is empty, it just ignores.
		resChunks = []db.ScoredChunk{}
	}

	// 3. Flatten the textual chunks
	contextText := ""
	for _, doc := range resChunks {
		if filename, ok := doc.Metadata["filename"]; ok {
			contextText += "File: " + filename + "\n"
		}
		contextText += doc.Content + "\n---\n"
	}

	// 4. Inject Local Universal User Memory (If preferences exist)
	rawMem, _ := p.KVStore.Get(ctx, "memories", "user_preferences")
	userPrefs := string(rawMem)

	// 5. Prepare Agentic Completeness
	sysPrompt := "Você é o assistente IA Vectora. Auxilie na inferência e desenvolvimento local usando os Arquivos Abaixo:\n\n" + contextText + "\n[Preferências]\n" + userPrefs

	var toolkit []llm.ToolDefinition
	for _, tMatch := range p.Agent.Registry.GetAll() {
		toolkit = append(toolkit, llm.ToolDefinition{
			Name:        tMatch.Name(),
			Description: tMatch.Description(),
			Schema:      tMatch.Schema(),
		})
	}

	llmTreq := llm.CompletionRequest{
		SystemPrompt: sysPrompt,
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: req.Query},
		},
		MaxTokens:   1500,
		Temperature: 0.1, // Extreme precision
		Tools:       toolkit, // Tool Registry Inteiro Disposto no Pipeline Master
	}

	// Completa c/ Google ou Local Qwen GGUF
	resp, err := p.LLM.Complete(ctx, llmTreq)
	if err != nil {
		return QueryResponse{}, err
	}

	// 6. Agentic Hook: If the LLM outputted Tools
	if len(resp.ToolCalls) > 0 {
		for _, tc := range resp.ToolCalls {
			tr, trErr := p.Agent.Registry.ExecuteStringArgs(ctx, tc.Name, tc.Args)
			_ = tr 
			_ = trErr
			// Futuro: Recursividade ReAct. Retro-injetamos "tr" no LLM pra re-pensar. 
			// (Limit maximum Loop to protect local RAM)
		}
	}

	return QueryResponse{
		Answer:  resp.Content,
		Sources: resChunks, // Envia metadados pro IPC Event Draw
	}, nil
}
