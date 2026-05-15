# 🚀 Release Engineering Roadmap - Vectora v0.1.0 MVP

**Status:** Release Candidate (RC)  
**Target:** Official Launch (PyPI + GitHub MCP Registry + GHCR)  
**Current Phase:** Hardening (100% Coverage + QA)

---

## 📊 Estado Atual

```
✅ Fire-and-Forget Architecture: IMPLEMENTADO
✅ 6 Passos (EmbeddingQueue → Tests): COMPLETO
⏳ Coverage: TBD (Necessário 100%)
❌ QA: Não iniciado (Requer 5 testers)
❌ Auditoria: Não iniciada
❌ Release: Não iniciada
```

---

## 🎯 Critérios de Sucesso para v0.1.0 MVP

| Critério                | Meta                                   | Status          |
| ----------------------- | -------------------------------------- | --------------- |
| **Cobertura de Testes** | 100%                                   | ⏳ Em andamento |
| **Testes Passando**     | Todos (unit + integration + E2E)       | ⏳ Em validação |
| **QA Friends & Family** | 5 testers independentes                | ❌ Não iniciado |
| **Aprovação QA**        | 0 bugs críticos, <5 bugs menores       | ❌ Aguardando   |
| **Auditoria**           | Observabilidade + Performance validada | ❌ Não iniciada |
| **Release Packing**     | PyPI + MCP Registry + GHCR             | ❌ Não iniciado |

---

## 🛠️ FASE 01: Cobertura 100% de Testes

### Objetivo

Identificar e cobrir todos os gaps de cobertura até atingir exatamente 100%.

### Etapas

#### 1.1 Auditoria de Cobertura Atual

```bash
# Execute este comando para gerar relatório completo
python -m pytest --cov=src --cov-report=html --cov-report=term-missing tests/

# Abra o relatório: htmlcov/index.html
```

**Métricas esperadas:**

- `src/` coverage: [TBD]
- Linhas não cobertas: [TBD]
- Branches não testadas: [TBD]

#### 1.2 Identificar Gaps de Cobertura

Procure por (em ordem de prioridade):

1. **Exceções não testadas** - Try/except blocks sem teste de erro
2. **Branches condicionais** - if/else que nunca executa ambos os caminhos
3. **Casos de borda** - Valores None, listas vazias, limites máximos
4. **Concorrência** - Race conditions entre threads/async tasks
5. **Falhas de rede** - Mock de timeouts, desconexões

#### 1.3 Implementar Testes Faltantes

**Prioridade Alta:**

- [ ] Tests de timeout em `background_worker.py`
- [ ] Tests de race condition entre `embedding()` e `background_worker()`
- [ ] Tests de falha de conexão LanceDB
- [ ] Tests de Voyage AI rate limiting (429 errors)
- [ ] Tests de reconciliation com records em "processing" há >2min

**Prioridade Média:**

- [ ] Tests de estado de LangGraph após 50+ interações
- [ ] Tests de memory leaks em loops assíncrono
- [ ] Tests de performance (latência <200ms para embedding())
- [ ] Tests de idempotência com mesmos queue_id

**Prioridade Baixa:**

- [ ] Tests de UI Textual (interações do usuário)
- [ ] Tests de MCP server routes
- [ ] Tests de formatação de output

#### 1.4 Validação Final

```bash
# Confirmar 100% de cobertura
python -m pytest --cov=src --cov-report=term-missing tests/
# Esperado: TOTAL 100%
```

---

## 🧪 FASE 02: QA Real (Friends & Family)

### Objetivo

Validar o Vectora com 5 testers independentes em cenários reais.

### Preparação

#### 2.1 Criar Ambiente de Staging

```bash
# Branch de staging
git checkout -b staging/pre-release

# Versão estável do código
git tag v0.1.0-rc1
```

#### 2.2 Distribuir para Testers

**Arquivo:** `QA_TESTING_GUIDE.md`

````markdown
# Teste do Vectora v0.1.0-RC

## Instalação

1. Clone o repositório
2. `uv sync --group test`
3. `python src/run_chat.py`

## Cenários de Teste

### Teste 1: TUI Responsiva

- Digite: "Pesquise sobre Next.js 16"
- **Esperado:** Resposta em <5 segundos
- **Erro reportado?** Salve o debug dump (veja abaixo)

### Teste 2: RAG e Vector Search

- Digite: "Pesquise sobre React 19"
- Aguarde 30 segundos
- Digite: "Quais são os principais features do React que você aprendeu?"
- **Esperado:** Resposta baseada em docs indexados (não web_search)

### Teste 3: Tratamento de Erros

- Digite: "Pesquise sobre um tópico que não existe na web"
- **Esperado:** Resposta graceful com sugestões

### Teste 4: Chat Multi-turn

- Digite 20+ mensagens em sequência
- **Esperado:** Sem crashes, histórico mantido

### Teste 5: MCP Integration

- Digite: "Use a ferramenta de clipboard do MCP"
- **Esperado:** Resposta bem-sucedida (se MCP disponível)

## Reportar Bug

Se algo falhar:

1. **Salve o debug dump:**
   ```bash
   vectora --debug-dump bug_report.tar.gz
   ```
````

2. **Envie o arquivo com:**
   - Descrição do que você estava fazendo
   - Mensagem de erro exata
   - Seu arquivo `bug_report.tar.gz`

````

#### 2.3 Debug Dump Command

**Implementar em `src/cli.py`:**

```python
import tarfile
import json
from pathlib import Path
from datetime import datetime

async def debug_dump(output_file: str = None):
    """Cria dump de debug para auditoria de QA."""
    if output_file is None:
        timestamp = datetime.now().isoformat().replace(':', '-')
        output_file = f"vectora_debug_{timestamp}.tar.gz"

    with tarfile.open(output_file, "w:gz") as tar:
        # Banco de dados
        if Path("data").exists():
            tar.add("data", arcname="data")

        # Logs
        if Path("logs").exists():
            tar.add("logs", arcname="logs")

        # Configuração (sem secrets)
        config = {
            "timestamp": timestamp,
            "python_version": sys.version,
            "platform": platform.platform(),
        }

        # Adicionar info.json
        tar.add(
            io.BytesIO(json.dumps(config, indent=2).encode()),
            arcname="info.json"
        )

    print(f"✅ Debug dump salvo em: {output_file}")
````

#### 2.4 Matriz de Testes (5 Testers)

| Tester   | Cenário Primário       | SO      | Feedback |
| -------- | ---------------------- | ------- | -------: |
| Tester 1 | TUI + RAG              | Windows |      [ ] |
| Tester 2 | Web Search + Embedding | macOS   |      [ ] |
| Tester 3 | MCP Integration        | Linux   |      [ ] |
| Tester 4 | Error Handling         | Windows |      [ ] |
| Tester 5 | Long Sessions (stress) | macOS   |      [ ] |

---

## 🔍 FASE 03: Auditoria de Observabilidade

### Objetivo

Validar que todo o fluxo é rastreável e pode ser auditado.

#### 3.1 Implementar Correlation ID

**Em `src/context.py`:**

```python
from uuid import uuid4

class Context(TypedDict):
    user_id: str
    user_type: str
    thread_id: int
    correlation_id: str  # ← NOVO
```

**Em cada nó do graph:**

```python
async def main_node(state: State, context: Context) -> dict:
    """Log com correlation_id para rastreabilidade."""
    logger.info(
        "main_node_executed",
        extra={
            "correlation_id": context.correlation_id,  # ← Adicionar
            "thread_id": context.thread_id,
            "messages_count": len(state.messages),
        }
    )
```

#### 3.2 Rastrear Cadeia RAG Completa

**Estrutura esperada de logs:**

```json
{"timestamp": "2026-05-14T10:00:00Z", "correlation_id": "abc123", "event": "web_search_started", "query": "Next.js 16"}
{"timestamp": "2026-05-14T10:00:02Z", "correlation_id": "abc123", "event": "web_search_completed", "results": 5}
{"timestamp": "2026-05-14T10:00:02Z", "correlation_id": "abc123", "event": "embedding_enqueued", "queue_id": "q123", "count": 5}
{"timestamp": "2026-05-14T10:00:07Z", "correlation_id": "abc123", "event": "embedding_processed", "queue_id": "q123", "status": "success"}
{"timestamp": "2026-05-14T10:00:08Z", "correlation_id": "abc123", "event": "vector_search_executed", "collection": "articles", "results": 3}
{"timestamp": "2026-05-14T10:00:09Z", "correlation_id": "abc123", "event": "response_sent"}
```

#### 3.3 Relatório de Performance

**Métrica:** Calcular latências médias

```python
# Exemplo: Coletar latências de cada nó
async def generate_performance_report():
    """Gera relatório de latência por nó."""
    data = {
        "embedding_enqueue": 0.150,  # <200ms ✓
        "vector_search": 0.080,       # <100ms ✓
        "web_search": 2.500,          # <5s ✓
        "background_worker_batch": 0.500,  # 5 docs/s ✓
    }

    # Gerar relatório
    print("📈 PERFORMANCE AUDIT")
    for metric, latency in data.items():
        status = "✅" if latency < 5 else "⚠️"
        print(f"{status} {metric}: {latency:.3f}s")
```

---

## 📦 FASE 04: Release Packing

### 4.1 PyPI Publishing

**Checklist:**

- [ ] `pyproject.toml` versão correta (`0.1.0`)
- [ ] `CHANGELOG.md` com release notes
- [ ] `README.md` atualizado
- [ ] Tests passando (100% coverage)
- [ ] CI/CD pipeline sucesso
- [ ] GitHub release criada
- [ ] PyPI account configurada

**Comando de publicação:**

```bash
# Build
python -m build

# Publish (requer token PyPI)
twine upload dist/*
```

### 4.2 GitHub MCP Registry

**Arquivo:** `.github/mcp-manifest.json`

```json
{
  "name": "vectora",
  "version": "0.1.0",
  "description": "Vectora - Advanced AI Assistant with RAG and MCP",
  "command": "vectora-mcp",
  "args": [],
  "env": {
    "VOYAGE_API_KEY": "required",
    "GOOGLE_API_KEY": "required"
  }
}
```

**Submit em:** https://github.com/modelcontextprotocol/registry

### 4.3 GHCR (GitHub Container Registry)

**Dockerfile:**

```dockerfile
FROM python:3.13-slim
WORKDIR /app
COPY . .
RUN pip install -e .
ENTRYPOINT ["vectora"]
```

**Build & Push:**

```bash
docker build -t ghcr.io/usuario/vectora:0.1.0 .
docker push ghcr.io/usuario/vectora:0.1.0
```

---

## 📋 Timeline Estimada

| Fase       | Atividade           | Duração  | Data Estimada  |
| ---------- | ------------------- | -------- | -------------- |
| 01         | Cobertura 100%      | 2-3 dias | 2026-05-16     |
| 02         | QA (5 testers)      | 1 semana | 2026-05-23     |
| 03         | Auditoria           | 2-3 dias | 2026-05-26     |
| 04         | Release Packing     | 1-2 dias | 2026-05-28     |
| **LAUNCH** | **v0.1.0 Official** |          | **2026-05-29** |

---

## ✅ Critérios de Aceitação para Launch

- [x] Code: Fire-and-Forget implementado
- [ ] Tests: 100% coverage
- [ ] QA: 5 testers aprovaram (0 critical bugs)
- [ ] Audit: Performance validada, observabilidade completa
- [ ] Release: PyPI, MCP Registry, GHCR publicados
- [ ] Docs: README, CHANGELOG, guia de instalação
- [ ] CI/CD: Pipeline verde, sem warnings

---

## 🚨 Casos de Bloqueio (Go/No-Go)

**Go para QA (Fase 02):**

- ✅ 100% de cobertura testada
- ✅ Todos os testes passando
- ✅ 0 warnings em linting/type checking
- ✅ Fire-and-Forget validado em dev

**Go para Auditoria (Fase 03):**

- ✅ QA completada sem bugs críticos
- ✅ <5 bugs menores reportados e fechados
- ✅ Nenhuma crash observada

**Go para Release (Fase 04):**

- ✅ Auditoria completada
- ✅ Performance dentro dos SLAs
- ✅ Observabilidade completa
- ✅ Documentação finalizada

---

## 📞 Próximos Passos Imediatos

1. **Hoje:** Gerar relatório de cobertura atual
2. **Hoje:** Identificar os 10-20 testes que faltam
3. **Amanhã:** Implementar testes faltantes
4. **Semana que vem:** Distribuir para QA com `debug-dump` command

---

**Versão:** Release Engineering Roadmap v0.1  
**Última atualização:** 2026-05-14  
**Responsável:** Machi + Claude
