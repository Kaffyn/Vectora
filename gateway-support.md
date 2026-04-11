# Suporte a Gateways & Agregadores

O Vectora AI foi projetado para ser agnóstico à plataforma. Além de suportar SDKs oficiais dos principais provedores, ele inclui integração nativa para os gateways e agregadores de IA mais populares. Isso permite usar uma única chave de API para acessar centenas de modelos ou apontar o Vectora para mecanismos de inferência locais.

## 🚀 Gateways Suportados

### 1. OpenRouter

O [OpenRouter](https://openrouter.ai/) fornece uma interface unificada para quase todos os LLMs (GPT-4, Claude 3, Llama 3, Gemini, etc.) com preços competitivos e modelos gratuitos.

**Configuração:**

```bash
vectora config set OPENROUTER_API_KEY sk-or-v1-sua-chave-aqui
```

**Uso:**

- **ID do Provider**: `openrouter`
- **IDs de Modelo**: Use o formato com barra, ex: `anthropic/claude-3.5-sonnet` ou `google/gemini-pro-1.5`.

---

### 2. Anannas

O [Anannas](https://anannas.ai/) é um agregador focado em acessibilidade e performance, frequentemente usado para testes e acesso a APIs gratuitas.

**Configuração:**

```bash
vectora config set ANANNAS_API_KEY sua-chave-anannas
```

**Uso:**

- **ID do Provider**: `anannas`
- **IDs de Modelo**: Ex: `openai/gpt-4o` ou `meta-llama/llama-3-70b`.

---

### 3. Endpoints Customizados compatíveis com OpenAI

Você pode apontar o Vectora para qualquer serviço que implemente a especificação da API da OpenAI (ex: **LocalAI, vLLM, LiteLLM, Ollama**).

**Configuração:**

```bash
vectora config set OPENAI_API_KEY chave-local-ficticia
vectora config set OPENAI_BASE_URL http://localhost:8080/v1
```

**Uso:**

- **ID do Provider**: `openai` (configurado com a URL customizada)
- **ID do Modelo**: O ID definido no seu mecanismo local.

---

## 🛠️ Detalhes Técnicos

Os Gateways no Vectora são alimentados pelo **SDK Oficial da OpenAI para Go** (`github.com/openai/openai-go`). Isso garante que recursos avançados como:

- ✅ **Respostas via Streaming**
- ✅ **Rastreamento de Uso de Tokens**
- ✅ **Tratamento de Erros Padronizado**
- ✅ **Chamadas de Ferramenta (Function Calling)**

...funcionem de forma confiável, independentemente de qual agregador você esteja usando.

## 💡 Dica de Teste

Se você quiser testar o suporte aos modelos mais recentes do Vectora (como GPT-5.4 ou Qwen 3) sem ter uma assinatura direta com esses provedores, o **OpenRouter** é a forma recomendada para começar.

> [!IMPORTANT]
> Ao usar um gateway, o `DEFAULT_PROVIDER` deve corresponder ao ID do provedor configurado (`openrouter`, `anannas` ou `openai`).
>
> ```bash
> vectora config set DEFAULT_PROVIDER openrouter
> vectora config set DEFAULT_MODEL anthropic/claude-3.5-sonnet
> ```
