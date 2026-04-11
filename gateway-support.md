# Suporte a Gateways & Agregadores

O Vectora AI foi projetado para ser agnóstico à plataforma. Além de suportar SDKs oficiais dos principais provedores, ele inclui integração nativa para os gateways e agregadores de IA mais populares. Isso permite usar uma única chave de API para acessar centenas de modelos ou apontar o Vectora para mecanismos de inferência locais.

## 🚀 Gateways Suportados

### 1. OpenRouter

O [OpenRouter](https://openrouter.ai/) fornece uma interface unificada para quase todos os LLMs modernos (GPT-5, Claude 4, Gemini 3, Qwen 3, etc.). Ele é considerado a melhor forma de descobrir novos modelos, já que a página [openrouter.ai/models](https://openrouter.ai/models) lista cada modelo disponível em tempo real com seus respectivos IDs.

**Configuração:**

```bash
vectora config set OPENROUTER_API_KEY sk-or-v1-sua-chave-aqui
```

**Uso:**

- **ID do Provider**: `openrouter`
- **IDs de Modelo**: Use o formato com barra, ex: `anthropic/claude-4.6-sonnet` ou `google/gemini-3.1-pro`.

---

### 2. Anannas

O [Anannas](https://anannas.ai/) é um agregador focado em acessibilidade e performance, frequentemente usado para testes e acesso a APIs gratuitas.

**Configuração:**

```bash
vectora config set ANANNAS_API_KEY sua-chave-anannas
```

**Uso:**

- **ID do Provider**: `anannas`
- **IDs de Modelo**: Ex: `openai/gpt-5.4-mini` ou `alibaba/qwen3.6-plus`.

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

## 🔍 Descoberta Dinâmica de Modelos (Model Discovery)

Diferente de sistemas estáticos, o Vectora consegue listar em tempo real quais modelos o seu gateway ou provedor configurado oferece.

```bash
vectora models list
```

Este comando consulta o endpoint do provedor ativo e retorna todos os IDs de modelos que você pode utilizar na configuração.

---

## 🧠 Roteamento de Embedding Inteligente (Family Detection)

Para garantir a melhor qualidade de recuperação (RAG) sem configuração manual complexa, o Vectora implementa a **Detecção de Família**.

Ao selecionar um modelo LLM via Gateway, o Vectora analisa o nome do modelo:

1. **Família Qwen**: Se o modelo for da família Qwen (ex: `qwen3.6`), o sistema tenta usar automaticamente o endpoint de embedding do Qwen (`qwen3-embedding-8b`).
2. **Família OpenAI/Gemini**: O sistema utiliza os modelos de embedding nativos desses provedores (`text-embedding-3-large` ou `gemini-embedding-2.0`).
3. **Fallback Global (Voyage AI)**: Para qualquer outro modelo (ex: Phi-4, Llama 4, Mistral) que não possua um embedding nativo configurado, o Vectora utilizará o **Voyage AI (voyage-3-large)** como fallback de alta performance.

> [!TIP]
> Para garantir que o fallback funcione corretamente, certifique-se de configurar a sua `VOYAGE_API_KEY`.

---

## 🛠️ Detalhes Técnicos

Os Gateways no Vectora são alimentados por um provedor especializado (`GatewayProvider`) baseado no **SDK Oficial da OpenAI para Go**. Isso garante que recursos avançados como:

- ✅ **Respostas via Streaming**
- ✅ **Rastreamento de Uso de Tokens**
- ✅ **Tratamento de Erros Padronizado**
- ✅ **Chamadas de Ferramenta (Function Calling)**

...funcionem de forma confiável, independentemente de qual agregador você esteja usando.

---

## 💡 Dica de Teste

Se você quiser testar o suporte aos modelos mais recentes do Vectora (como GPT-5.4 ou Qwen 3.6) sem ter uma assinatura direta com esses provedores, o **OpenRouter** é a forma recomendada para começar.

> [!IMPORTANT]
> Ao usar um gateway, o `DEFAULT_PROVIDER` deve corresponder ao ID do provedor configurado (`openrouter`, `anannas` ou `openai`).
