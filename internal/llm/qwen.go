package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai" // Llama.cpp é OpenAI Server-Compatible nativamente
	vecos "github.com/Kaffyn/vectora/internal/os"
)

type QwenProvider struct {
	client    *openai.LLM
	port      int
	osManager vecos.OSManager
}

// Inicializa o Motor Físico do Qwen chamando o OS Manager para alocar llama.cpp.
func NewQwenProvider(ctx context.Context, modelPath string) (*QwenProvider, error) {
	osMgr, err := vecos.NewManager()
	if err != nil {
		return nil, err
	}

	// Determinação dinâmica de Porta TCP
	dynamicPort := 42781 // Poderiamos scanear, por agora é hardcoded alto.
	
	err = osMgr.StartLlamaEngine(modelPath, dynamicPort)
	if err != nil {
		return nil, fmt.Errorf("qwen_boot_failed: falha ao acoplar sidecar do llama.cpp (%v)", err)
	}

	// Espera inicial p/ binding TCP. Llama-server demora alguns MSp/ boot.
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/v1", dynamicPort)
	
	// Usa cliente nativo OpenAI que aponta p/ o host local do LLaMa
	client, err := openai.New(
		openai.WithBaseURL(baseURL),
		openai.WithToken("local-qwen-offline"), // Mock key
	)
	if err != nil {
		osMgr.StopLlamaEngine()
		return nil, err
	}

	return &QwenProvider{
		client:    client,
		port:      dynamicPort,
		osManager: osMgr,
	}, nil
}

func (p *QwenProvider) Name() string {
	return "qwen"
}

func (p *QwenProvider) IsConfigured() bool {
	return p.client != nil && p.osManager.GetEngineState() == string(vecos.EngineRunning)
}

func (p *QwenProvider) Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error) {
	// Checa Saúde do Motor
	if p.osManager.GetEngineState() != string(vecos.EngineRunning) {
		return CompletionResponse{}, errors.New("llama_engine_offline: the qwen backend crashed or is stopped")
	}

	var content []llms.MessageContent
	if req.SystemPrompt != "" {
		content = append(content, llms.TextParts(llms.ChatMessageTypeSystem, req.SystemPrompt))
	}

	for _, msg := range req.Messages {
		var role llms.ChatMessageType
		switch msg.Role {
		case RoleUser:
			role = llms.ChatMessageTypeHuman
		case RoleAssistant:
			role = llms.ChatMessageTypeAI
		default:
			role = llms.ChatMessageTypeHuman 
		}
		content = append(content, llms.TextParts(role, msg.Content))
	}

	var langchanTools []llms.Tool
	for _, t := range req.Tools {
		langchanTools = append(langchanTools, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  json.RawMessage(t.Schema),
			},
		})
	}
	
	opts := []llms.CallOption{
		llms.WithMaxTokens(req.MaxTokens),
		llms.WithTemperature(float64(req.Temperature)),
	}

	if len(langchanTools) > 0 {
		opts = append(opts, llms.WithTools(langchanTools))
	}

	resp, err := p.client.GenerateContent(ctx, content, opts...)
	if err != nil {
		return CompletionResponse{}, err
	}

	if len(resp.Choices) == 0 {
		return CompletionResponse{}, errors.New("llm_empty_response")
	}

	choice := resp.Choices[0]

	var tCalls []ToolCall
	for _, lcTc := range choice.ToolCalls {
		tCalls = append(tCalls, ToolCall{
			ID:   lcTc.ID,
			Name: lcTc.FunctionCall.Name,
			Args: lcTc.FunctionCall.Arguments,
		})
	}

	return CompletionResponse{
		Content: choice.Content,
		ToolCalls: tCalls,
	}, nil
}

func (p *QwenProvider) Embed(ctx context.Context, input string) ([]float32, error) {
	embClient, err := p.client.CreateEmbedding(ctx, []string{input})
	if err != nil {
		return nil, err
	}
	if len(embClient) > 0 {
		return embClient[0], nil
	}
	return nil, errors.New("qwen_embedding_failed: sem vetores do llama.cpp")
}

func (p *QwenProvider) Shutdown() {
	if p.osManager != nil {
		p.osManager.StopLlamaEngine()
	}
}
