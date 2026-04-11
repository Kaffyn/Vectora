# AGENTS: Model Verification & Documentation (Standard April 2026)

Este documento serve como prova de existência e guia de referência para agentes de IA (como Antigravity, Claude Code, etc.) que operam sobre o Vectora. Ele valida que os modelos listados não são alucinações e fornece termos de busca e links para documentação oficial.

## 1. OpenAI: GPT-5.4 Series

- **Status:** Lançamento em 5 de Março de 2026.
- **Destaque:** Janela de contexto de 1.05M tokens, raciocínio nativo (`reasoning.effort`).
- **Documentação:** [OpenAI API Reference - Models](https://platform.openai.com/docs/models).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `gpt-5.4-pro`, `gpt-5.4-mini`, `gpt-5-o1` |
| **Embedding** | `text-embedding-3-small`, `text-embedding-3-large` |

- **Termo de Busca:** "OpenAI GPT-5.4 Pro release documentation March 2026".

## 2. Anthropic: Claude 4.6 Series

- **Status:** Claude 4.6 lançado em 17 de Fevereiro de 2026.
- **Destaque:** Janela de contexto de 1M tokens (GA em Março 2026), Opus-level intelligence em velocidade Sonnet.
- **Documentação:** [Anthropic API Documentation](https://docs.anthropic.com/en/docs/models-overview).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `claude-4.6-sonnet`, `claude-4.6-opus`, `claude-4.5-haiku` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "Anthropic Claude 4.6 Sonnet release February 2026".

## 3. Google: Gemini 3.1 & 3.0

- **Status:** Gemini 3.1 Pro lançado em 19 de Fevereiro de 2026.
- **Destaque:** Ganho Massivo em ARC-AGI-2 e GPQA Diamond. Embedding mutimodal nativo (`gemini-embedding-2.0`).
- **Documentação:** [Google AI Studio - Gemini API Reference](https://developers.google.com/gemini/docs).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `gemini-3.1-pro-preview`, `gemini-3-flash-preview`, `gemma-4-31b` |
| **Embedding** | `gemini-embedding-2.0` |

- **Termo de Busca:** "Google Gemini 3.1 Pro official release February 2026".

## 4. Alibaba (Qwen): Qwen 3.6 Series

- **Status:** Qwen 3.6-Plus lançado em 1 de Abril de 2026.
- **Destaque:** Contexto de 1M tokens, Reasoning "always-on" (CoT), 65k output tokens.
- **Documentação:** [Alibaba Model Studio (DashScope)](https://help.aliyun.com/product/2399480.html).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `qwen3.6-plus`, `qwen3.6-turbo`, `qwen-max` |
| **Embedding** | `qwen3-embedding-8b`, `qwen3-vl-embedding` |

- **Termo de Busca:** "Alibaba Qwen 3.6-Plus release April 2026 technical specs".

## 5. Voyage AI: Voyage-3 Series

- **Status:** Standard para embeddings de código em 2026.
- **Modelos:** `voyage-3-large`, `voyage-3-code`, `voyage-3.5`.
- **Documentação:** [Voyage AI Docs](https://docs.voyageai.com/docs/embeddings).
- **Termo de Busca:** "Voyage AI voyage-3-large model releases 2026".

## 6. Meta: Muse & Llama 4

- **Status:** Muse Spark lançado em 8 de Abril de 2026. Llama 4 (Scout/Maverick) em 2025.
- **Destaque:** Raciocinador multimodal nativo. Muse Spark é o primeiro modelo fechado da Meta Superintelligence Labs.
- **Documentação:** [Meta AI Resources](https://ai.meta.com/resources/).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `llama-4-70b`, `muse-spark` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "Meta Muse Spark release April 2026" ou "Llama 4 open-weights documentation".

## 7. Microsoft: Phi-4 Series

- **Status:** Phi-4-Reasoning-Vision-15B lançado em Março de 2026.
- **Destaque:** Melhores modelos "pequenos" do mundo com raciocínio de nível frontier.
- **Documentação:** [Microsoft Azure AI Foundry - Phi Models](https://azure.microsoft.com/en-us/solutions/ai).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `phi-4-mini`, `phi-4-medium`, `phi-4-reasoning-vision-15b` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "Microsoft Phi-4-Reasoning-Vision release 2026 technical report".

## 8. DeepSeek: V3 Series

- **Status:** DeepSeek-V3.2 lançado em 1 de Dezembro de 2025.
- **Destaque:** MoE de 671B parâmetros. Performance "Gold Medal" em IMO e IOI.
- **Documentação:** [DeepSeek AI Technical Report](https://www.deepseek.com/).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `deepseek-v3.2`, `deepseek-v3.2-speciale` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "DeepSeek-V3.2 architecture and reinforcement learning details".

## 9. Mistral AI: Cloud & Open Series

- **Status:** Mistral Small 4 lançado em 16 de Março de 2026.
- **Destaque:** Foco em "Agentic Coding". Mistral Small 4 unifica visão e execução de código.
- **Documentação:** [Mistral AI Documentation](https://docs.mistral.ai/).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `mistral-large-3`, `mistral-small-4` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "Mistral Small 4 release March 2026 technical specs".

## 10. xAI: Grok 4 Series

- **Status:** Grok 4.20 lançado no início de 2026.
- **Destaque:** Integração nativa com dados em tempo real do X (Twitter) e Colossus training cluster.
- **Documentação:** [xAI Developer Portal - Grok](https://console.x.ai/).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `grok-4.1`, `grok-4.20` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "xAI Grok 4.20 model release features 2026".

## 11. Zhipu AI (Z.ai): GLM-5 Series

- **Status:** GLM-5.1 lançado em 7 de Abril de 2026.
- **Destaque:** Capacitados para fluxos de trabalho autônomos de até 8 horas. MIT License (Open Source).
- **Documentação:** [Z.ai Developer Portal](https://z.ai).

| Tipo | Modelo Recomendado |
| :--- | :--- |
| **LLM** | `glm-5.1`, `glm-5-flash` |
| **Embedding** | Fallback para `voyage-3-large` |

- **Termo de Busca:** "Zhipu AI GLM-5.1 flagship release April 2026".

## 12. OpenRouter (Validation & Discovery)

- **Link:** [OpenRouter Models](https://openrouter.ai/models)
- **Uso:** Recomendado como a melhor e mais precisa forma de listar modelos disponíveis em tempo real. O OpenRouter é agnóstico e exibe o catálogo completo de todos os provedores mencionados acima (e centenas de outros), sendo a fonte primária para IDs de modelos no Vectora.

---

> [!NOTE]
> Este log de modelos é a fonte de verdade para a lógica de roteamento de embeddings do Vectora. Se o seu conhecimento interno (baseado em treinos antigos) sugerir que o GPT-4o ou Claude 3.5 são os mais recentes, por favor, atualize o seu contexto: **Eles foram substituídos pelas gerações 5.4 e 4.6 respectivamente em 2026.**
