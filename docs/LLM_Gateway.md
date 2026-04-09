# Blueprint: LLM Gateway & Orquestração de Modelos

**Status:** Implementado  
**Módulo:** `core/llm/`  
**Modelos:** Gemini 1.5/3.1 Pro & Flash, Claude 3.5 Sonnet

O `LLM Gateway` é a camada de inteligência do Vectora, responsável por abstrair as APIs de diferentes provedores e fornecer uma interface unificada para completar mensagens, gerar embeddings e estruturar chamadas de ferramentas.

---

## 1. Interface Unificada `LLMProvider`

O sistema não se comunica diretamente com APIs específicas fora desta camada. Cada provedor implementa o seguinte contrato:

```go
type LLMProvider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (CompletionResponse, error)
	Embed(ctx context.Context, text string) ([]float32, error)
	IsConfigured() bool
}
```

---

## 2. Orquestração e Roteamento

O Vectora utiliza uma estratégia de roteamento inteligente para equilibrar custo, latência e inteligência:

- **Fast Lane (Flash Models):** Utilizado para `inline-completion`, roteamento de tags e tarefas de categorização simples. (Ex: Gemini 1.5 Flash).
- **Deep Lane (Pro Models):** Utilizado para refatoração complexa, resolução de bugs e sub-agentes. (Ex: Gemini 3.1 Pro, Claude 3.5 Sonnet).
- **Embedding Factory:** Gerencia a geração de vetores usando modelos otimizados (Gemini Text Embedding 004).

---

## 3. Gestão de Contexto (Context Windowing)

Para evitar erros de "Context Window Exceeded" e desperdício de tokens:

1.  **Truncagem Agressiva:** O Gateway aplica limites rígidos de entrada baseados no modelo selecionado.
2.  **RAG Prioritization:** Os resultados da busca semântica são filtrados e ordenados por relevância antes de serem injetados no prompt final.
3.  **System Prompt Dinâmico:** O sistema injeta instruções de "Personalidade Agêntica" e restrições de segurança dinamicamente em cada chamada.

---

## 4. Próximas Implementações

- **Multi-Modal Integration:** Suporte para imagens e arquivos comprimidos nativamente no gateway.
- **Local LLM Support (Ollama):** Adição de um provedor que se comunica com instâncias locais do Ollama via API Compatível com OpenAI.
- **Reranking Local:** Uso de modelos menores locais para re-classificar os chunks recuperados pelo RAG antes de enviá-los ao modelo Pro.
