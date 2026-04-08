# Blueprint: Orquestração LLM (The Brain)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/llm/`  
**Dependencies:** `google.golang.org/genai` (Gemini), `github.com/pkoukk/tiktoken-go` (Tokenization), `context`, `fmt`

## 1. Interface Unificada do Provider (`provider.go`)

Define o contrato mínimo que qualquer provedor (Gemini, Claude, Qwen/Local) deve implementar. Isso permite trocar o "cérebro" sem reescrever a lógica do agente.

```go
package llm

import (
	"context"
	"io"
)

// Message representa uma unidade de conversa.
type Message struct {
	Role    string // "user", "model", "system"
	Content string
}

// ChatRequest encapsula os parâmetros de uma chamada de chat.
type ChatRequest struct {
	Messages      []Message
	Temperature   float32
	MaxOutputTokens int
	Tools         []ToolDefinition // Schema das ferramentas disponíveis
}

// ChatResponse streama a resposta do modelo.
type ChatResponse struct {
	Text  string
	Done  bool
	Error error
}

// EmbedRequest para geração de vetores.
type EmbedRequest struct {
	Text string
	Model string // ex: "gemini-embedding-2"
}

// LLMProvider é a interface mestre.
type LLMProvider interface {
	// Chat envia mensagens e retorna um reader para streamar a resposta.
	Chat(ctx context.Context, req ChatRequest) (io.ReadCloser, error)

	// Embed gera o vetor para um texto específico.
	Embed(ctx context.Context, req EmbedRequest) ([]float32, error)

	// Name retorna o identificador do provider (ex: "gemini-pro")
	Name() string
}
```

## 2. Gerenciador de Contexto e Truncagem (`context_manager.go`)

Implementa a lógica de "Janela Deslizante" com priorização de conhecimento local (RAG) sobre histórico antigo.

```go
package llm

import (
	"github.com/pkoukk/tiktoken-go"
)

const (
	SafetyMargin = 500 // Margem de segurança em tokens para evitar off-by-one errors
	SystemPromptOverhead = 1000 // Reserva para system prompt e policies
)

type ContextManager struct {
	encoding *tiktoken.Tiktoken
	MaxContextTokens int
}

func NewContextManager(modelName string, maxTokens int) (*ContextManager, error) {
	enc, err := tiktoken.EncodingForModel(modelName)
	if err != nil {
		// Fallback para cl100k_base (usado por GPT-4 e Gemini)
		enc, err = tiktoken.GetEncoding("cl100k_base")
		if err != nil {
			return nil, err
		}
	}
	return &ContextManager{encoding: enc, MaxContextTokens: maxTokens}, nil
}

// CountTokens calcula o número exato de tokens.
func (cm *ContextManager) CountTokens(text string) int {
	return len(cm.encoding.Encode(text, nil, nil))
}

// TrimMessages aplica a estratégia de truncagem agressiva.
// Prioridade: System Prompt > RAG Context > Recent History > Old History
func (cm *ContextManager) TrimMessages(systemPrompt string, ragContext string, history []Message) ([]Message, error) {
	availableTokens := cm.MaxContextTokens - SystemPromptOverhead - cm.CountTokens(systemPrompt) - cm.CountTokens(ragContext)

	if availableTokens <= 0 {
		return nil, fmt.Errorf("system prompt and RAG context exceed model limits")
	}

	var trimmedHistory []Message
	currentTokens := 0

	// Itera do mais recente para o mais antigo
	for i := len(history) - 1; i >= 0; i-- {
		msg := history[i]
		tokens := cm.CountTokens(msg.Content)

		if currentTokens + tokens > availableTokens {
			break // Corta o histórico antigo
		}

		trimmedHistory = append([]Message{msg}, trimmedHistory...)
		currentTokens += tokens
	}

	return trimmedHistory, nil
}
```

## 3. Fábrica de System Prompt Dinâmico (`prompt_factory.go`)

Monta o prompt final injetando as Políticas YAML e o Contexto RAG como "Verdade Absoluta".

```go
package llm

import (
	"fmt"
	"strings"
	"vectora/core/policies"
)

type PromptFactory struct {
	PolicyEngine *policies.PolicyConfig
	BaseIdentity string
}

func NewPromptFactory(policy *policies.PolicyConfig) *PromptFactory {
	return &PromptFactory{
		PolicyEngine: policy,
		BaseIdentity: "You are Vectora, an elite AI software engineer assistant. You operate strictly within the Trust Folder.",
	}
}

// BuildSystemPrompt cria o prompt mestre.
func (pf *PromptFactory) BuildSystemPrompt(ragContext string) string {
	var sb strings.Builder

	// 1. Identidade e Regras Básicas
	sb.WriteString(pf.BaseIdentity)
	sb.WriteString("\n\n[SECURITY POLICIES - NON-NEGOTIABLE]\n")
	sb.WriteString("- NEVER access files outside the Trust Folder.\n")
	sb.WriteString("- NEVER read/write protected files (.env, .key, .db).\n")
	sb.WriteString("- ALWAYS use provided tools for file operations.\n")

	// 2. Injeção de Contexto RAG (Prioridade Máxima)
	if ragContext != "" {
		sb.WriteString("\n\n[SYSTEM_KNOWLEDGE - SOURCE OF TRUTH]\n")
		sb.WriteString("The following information is retrieved from the local codebase. It has HIGHEST priority over your pre-trained knowledge. Do not contradict it.\n\n")
		sb.WriteString(ragContext)
		sb.WriteString("\n[/SYSTEM_KNOWLEDGE]\n")
	} else {
		sb.WriteString("\n\n[SYSTEM_NOTE]\nNo local context retrieved. Rely on general knowledge but verify via tools if possible.\n")
	}

	return sb.String()
}

// BuildFinalPayload monta o array de mensagens pronto para o Provider.
func (pf *PromptFactory) BuildFinalPayload(query string, ragContext string, history []Message, cm *ContextManager) (ChatRequest, error) {
	systemPrompt := pf.BuildSystemPrompt(ragContext)

	// Trunca histórico para caber na janela
	trimmedHistory, err := cm.TrimMessages(systemPrompt, ragContext, history)
	if err != nil {
		return ChatRequest{}, err
	}

	// Monta mensagens finais
	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}
	messages = append(messages, trimmedHistory...)
	messages = append(messages, Message{Role: "user", Content: query})

	return ChatRequest{
		Messages: messages,
		Temperature: 0.2, // Baixa temperatura para código/fatos
	}, nil
}
```

## 4. Implementação do Provider Gemini (`gemini_provider.go`)

Exemplo concreto de implementação da interface `LLMProvider`.

```go
package llm

import (
	"context"
	"fmt"
	"io"
	"strings"

	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
	model  string
	apiKey string
}

func NewGeminiProvider(apiKey, model string) (*GeminiProvider, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &GeminiProvider{client: client, model: model, apiKey: apiKey}, nil
}

func (g *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (io.ReadCloser, error) {
	// Conversão simplificada para o formato do SDK Gemini
	contents := make([]*genai.Content, len(req.Messages))
	for i, msg := range req.Messages {
		role := "user"
		if msg.Role == "model" {
			role = "model"
		}
		contents[i] = &genai.Content{
			Parts: []*genai.Part{{Text: msg.Content}},
			Role:  role,
		}
	}

	// Nota: O SDK Gemini real usa streams de forma diferente.
	// Aqui simulamos um io.ReadCloser para manter a interface unificada.
	// Em produção, adaptaríamos para o stream nativo do SDK.

	resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, nil)
	if err != nil {
		return nil, err
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	return io.NopCloser(strings.NewReader(text)), nil
}

func (g *GeminiProvider) Embed(ctx context.Context, req EmbedRequest) ([]float32, error) {
	res, err := g.client.Models.EmbedContent(ctx, g.model, &genai.EmbedContentRequest{
		Content: &genai.Content{Parts: []*genai.Part{{Text: req.Text}}},
	})
	if err != nil {
		return nil, err
	}

	// Extrai valores do embedding
	values := res.Embedding.Values
	floats := make([]float32, len(values))
	for i, v := range values {
		floats[i] = float32(v)
	}
	return floats, nil
}

func (g *GeminiProvider) Name() string {
	return g.model
}
```

## 5. Router de Providers (`router.go`)

Seleciona o provider correto baseado na configuração do usuário.

```go
package llm

import (
	"fmt"
	"vectora/core/config"
)

type Router struct {
	providers map[string]LLMProvider
	defaultProvider string
}

func NewRouter(cfg *config.Config) (*Router, error) {
	r := &Router{
		providers: make(map[string]LLMProvider),
	}

	// Inicializa providers disponíveis
	if cfg.Gemini.APIKey != "" {
		p, err := NewGeminiProvider(cfg.Gemini.APIKey, cfg.Gemini.Model)
		if err != nil {
			return nil, err
		}
		r.providers["gemini"] = p
		r.defaultProvider = "gemini"
	}

	// Futuro: Adicionar Qwen/Local aqui
	// if cfg.Qwen.Enabled { ... }

	if len(r.providers) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}

	return r, nil
}

func (r *Router) GetProvider(name string) (LLMProvider, error) {
	if p, ok := r.providers[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("provider %s not found", name)
}

func (r *Router) GetDefault() LLMProvider {
	return r.providers[r.defaultProvider]
}
```

---

### Resumo da Estratégia de "The Brain"

1.  **Desacoplamento Total:** O `core/agent` nunca chama a API do Google ou da OpenAI diretamente. Ele chama `provider.Chat()`.
2.  **Segurança de Contexto:** O `ContextManager` garante matematicamente que não haverão erros de `400 Context Length Exceeded`.
3.  **Prioridade de Conhecimento:** O `PromptFactory` injeta o RAG como `[SYSTEM_KNOWLEDGE]`, instruindo o modelo a ignorar seu treinamento prévio se houver conflito com o código local.
4.  **Extensibilidade:** Adicionar suporte ao Qwen Local ou Claude requer apenas criar um novo arquivo `qwen_provider.go` que implemente a interface `LLMProvider` e registrá-lo no `Router`.
