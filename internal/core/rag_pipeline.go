package core

import (
	"context"
	"errors"

	"github.com/Kaffyn/vectora/internal/acp"
	"github.com/Kaffyn/vectora/internal/db"
	"github.com/Kaffyn/vectora/internal/llm"
)

// Pipeline orquestra dados entre o LLM (Interface) e o VectorDB (Armazenamento).
// Ela é instanciada e retida pelo Daemon SystemTray.
type Pipeline struct {
	LLM      llm.Provider
	VectorDB db.VectorStore
	KVStore  db.KVStore
	Agent    *acp.AgentContext
}

// Placa Mãe que acopla todos os módulos soltos para injetar queries.
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

// O RAG (Retrieval-Augmented Generation) Canônico do Vectora.
func (p *Pipeline) Query(ctx context.Context, req QueryRequest) (QueryResponse, error) {
	if p.LLM == nil || !p.LLM.IsConfigured() {
		return QueryResponse{}, errors.New("pipeline_err: LLM provider não injetado ou não configurado na Bandeja (Tray)")
	}

	// 1. Embendamento Nativo
	vector, err := p.LLM.Embed(ctx, req.Query)
	if err != nil {
		return QueryResponse{}, err
	}

	// 2. Nearest Neighbors no Banco KNN Chromem
	// TOP K travado em 5 para evitar poluição de LLM window em baixa RAM.
	resChunks, err := p.VectorDB.Query(ctx, "ws_"+req.WorkspaceID, vector, 5)
	if err != nil {
		// Zero-Shot Fallback. Se coleção é vazia ele apenas ignora.
		resChunks = []db.ScoredChunk{}
	}

	// 3. Empena os Bloquinhos Textuais
	contextText := ""
	for _, doc := range resChunks {
		if filename, ok := doc.Metadata["filename"]; ok {
			contextText += "File: " + filename + "\n"
		}
		contextText += doc.Content + "\n---\n"
	}

	// 4. Injeta Memória Universal Local do User (Se existir preferência)
	rawMem, _ := p.KVStore.Get(ctx, "memories", "user_preferences")
	userPrefs := string(rawMem)

	// 5. Prepara Completude Agêntica
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
		Temperature: 0.1, // Extrema exatidão
		Tools:       toolkit, // Tool Registry Inteiro Disposto no Pipeline Master
	}

	// Completa c/ Google ou Local Qwen GGUF
	resp, err := p.LLM.Complete(ctx, llmTreq)
	if err != nil {
		return QueryResponse{}, err
	}

	// 6. Hook Agêntico: Se o LLM Cuspiu Ferramentas
	if len(resp.ToolCalls) > 0 {
		for _, tc := range resp.ToolCalls {
			tr, trErr := p.Agent.Registry.ExecuteStringArgs(ctx, tc.Name, tc.Args)
			_ = tr 
			_ = trErr
			// Futuro: Recursividade ReAct. Retro-injetamos "tr" no LLM pra re-pensar. 
			// (Limitaremos o Loop máximo p/ proteger RAM local)
		}
	}

	return QueryResponse{
		Answer:  resp.Content,
		Sources: resChunks, // Envia metadados pro IPC Event Draw
	}, nil
}
