package llm

import "context"

type Provider interface {
	Complete(ctx context.Context, prompt string) (string, error)
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	IsConfigured() bool
	Name() string
}

type SimpleProvider struct {
	name string
}

func NewSimpleProvider(name string) *SimpleProvider {
	return &SimpleProvider{name: name}
}

func (p *SimpleProvider) Complete(ctx context.Context, prompt string) (string, error) {
	// Simula resposta LLM
	return "Resposta simulada do " + p.name, nil
}

func (p *SimpleProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	// Simula embeddings (384-dimensional vectors)
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = make([]float32, 384)
		for j := range embeddings[i] {
			embeddings[i][j] = 0.1
		}
	}
	return embeddings, nil
}

func (p *SimpleProvider) IsConfigured() bool {
	return true
}

func (p *SimpleProvider) Name() string {
	return p.name
}
