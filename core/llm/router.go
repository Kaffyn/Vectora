package llm

import (
	"fmt"
	// "vectora/core/config"
)

// Emulação do pacote config
type Config struct {
	Gemini struct {
		APIKey string
		Model  string
	}
}

type Router struct {
	providers       map[string]LLMProvider
	defaultProvider string
}

func NewRouter(cfg *Config) (*Router, error) {
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
