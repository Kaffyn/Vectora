# Plano de Implementação: Motor de Inferência Llama (Qwen3)

Este plano detalha a integração do motor de inferência local `llama.cpp` no Vectora, utilizando a arquitetura **Zero-Port (Pipes)** para máxima segurança e isolamento.

---

## 1. Visão Geral e Objetivos

O objetivo é fornecer inferência de LLM e Embeddings 100% offline, sem abrir portas de rede local (TCP) e sem depender de drivers de terceiros. A comunicação é realizada via subprocessos e pipes de Standard I/O.

- **Modelo Principal:** Família Qwen3 (Qwen3-Coder, Qwen3-Thinking).
- **Protocolo:** JSON-ND (Newline Delimited) sobre Pipes.
- **Isolamento:** Processo sidecar gerenciado pelo Daemon.

---

## 2. Arquitetura Zero-Port (Pipes)

Diferente das implementações convencionais que sobem um servidor HTTP local, o Vectora utiliza execução direta via `os/exec`.

### 2.1 O Ciclo de Vida do Processo (`internal/llm/protocol_llama.go`)

O `LlamaProcess` gerencia o binário `llama-cli`:

1. **Iniciação:** O Daemon invoca o binário com as flags `--interactive`, `--json-nd` e `--model <path>`.
2. **Stdin (Comandos):** O Daemon envia prompts e pedidos de embedding como objetos JSON em uma única linha.
3. **Stdout (Resposta):** O binário responde conforme gera os tokens, enviando objetos JSON contendo o token e o status.
4. **Stderr:** Capturado pelo Logger do Vectora para depuração de integridade do modelo.

### 2.2 Estrutura de Payload (JSON-ND)

**Input (Daemon -> Sidecar):**
```json
{ "prompt": "[INST] Olá, mundo! [/INST]", "temp": 0.7, "max_tokens": 1024, "stop": ["</s>"] }
```

**Output (Sidecar -> Daemon):**
```json
{ "token": "Olá", "done": false }
{ "token": "!", "done": true }
```

---

## 3. Gestão de Modelos (Qwen3 Optimized)

O Vectora é otimizado para os modelos da Kaffyn/Qwen, distribuídos via **Vectora Index**.

| Modelo | Uso Recomendado | RAM (GGUF Q4) |
| --- | --- | --- |
| **Qwen3-Embedding (0.6B)** | Indexação RAG (Obrigatório) | ~500MB |
| **Qwen3-4B-Thinking** | Raciocínio Lógico / Debug | ~3GB |
| **Qwen3-Coder-Next (80B)** | Refatoração de Arquitetura | Requer GPU/32GB+ |

---

## 4. Implementação do `Provider.Embed` (Lacuna Atual)

Atualmente no código (`qwen.go`), a função `Embed` retorna erro. O plano de refatoração para a `internal/llm` é:

1. **Multiplexação:** Expandir o `LlamaProcess` para aceitar um `Mode` (INFERENCE | EMBEDDING).
2. **Buffer de Embedding:** Enviar blocos de texto e capturar o vetor `[]float32` bruto retornado pelo `llama-cli` em modo `-e`.
3. **Cache:** Implementar cache de vetores para evitar re-calculo de strings idênticas no mesmo workspace.

---

## 5. Regras de Segurança e Performance

- **RN-LLM-01:** O processo sidecar deve ser encerrado imediatamente se o Daemon (parent) for finalizado (`kill inheritance`).
- **RN-LLM-02:** O consumo de RAM do sidecar deve ser monitorado e reportado via IPC `daemon.status`.
- **RN-LLM-03:** Em caso de `OOM (Out Of Memory)`, o Daemon deve notificar o Web UI para sugerir um modelo mais leve (ex: 0.6B).

---

## 6. Próximos Passos (Workflow)

1.  [ ] Finalizar a implementação do `Embed(ctx, input)` no `qwen.go`.
2.  [ ] Validar a integração com o `chromem-go` usando vetores locais.
3.  [ ] Criar testes de estresse para concorrência de Pipes no Windows (Named Pipes).

[Fim do Plano Llama - Revisão 2026.04.03]
