# 🔬 Vectora Research Agent - Pesquisa Autônoma em Larga Escala

## Visão Geral

O **Vectora Research Agent** é um conceito de demonstração que mostra como o Vectora pode ser usado como um **agente autônomo** para executar pesquisas massivas:

- **150 arquivos** sobre 15 tópicos (10 arquivos por tópico)
- **1,500 web_searches** totais (10 buscas por arquivo)
- **Acesso automático a sites** via `fetch_url()` para cada resultado
- **Sumários inteligentes** gerados pelo LLM após compilar todas as informações

---

## Como Funciona (Arquitetura)

### Fluxo de Pesquisa por Arquivo

```
┌─────────────────────────────────────────────────────────┐
│ Arquivo 1: "Next.js 16 - Features"                      │
├─────────────────────────────────────────────────────────┤
│ 1. web_search("Next.js 16 features 2025")              │
│    ↓ Retorna 10+ URLs                                   │
│    ├─ fetch_url(url1) → Extrai conteúdo                │
│    ├─ fetch_url(url2) → Extrai conteúdo                │
│    └─ fetch_url(url3) → Extrai conteúdo                │
│                                                         │
│ 2. web_search("Next.js 16 app router")                 │
│    ↓ Processa 3-5 melhores URLs                        │
│    ├─ fetch_url(...) → Extrai                          │
│    └─ fetch_url(...) → Extrai                          │
│                                                         │
│ 3-10. Executa 8 buscas adicionais similares            │
│                                                         │
│ ✨ LLM Gemini: Compila tudo em sumário de 1000 palavras│
│    "Next.js 16 oferece features como..."               │
│                                                         │
│ 📄 Salva: output/research/next-js-16/01_features.md    │
└─────────────────────────────────────────────────────────┘
```

### Orquestração Completa (15 Tópicos × 10 Arquivos)

```
TOPIC 1: Next.js 16
  ├─ 01_introducao.md (web_search × 10 + fetch_url × 30)
  ├─ 02_historia.md
  ├─ 03_features.md
  ├─ 04_arquitetura.md
  ├─ ...
  └─ 10_roadmap.md

TOPIC 2: Godot 4.7
  ├─ 01_introducao.md
  ├─ 02_historia.md
  └─ ...

TOPIC 3-15: Outros tópicos (React 19, PostgreSQL 17, Rust Web, etc.)

RESULTADO FINAL:
├── output/research/
│   ├── next-js-16/ (10 × ~1000 palavras = 10K palavras)
│   ├── godot-47/ (10K palavras)
│   ├── react-19/ (10K palavras)
│   └── ... × 12 tópicos adicionais
│
└── TOTAL: 150 arquivos, ~150K palavras, completamente pesquisados
```

---

## Tópicos Implementados (15)

1. **Next.js 16** — Framework React moderno com SSR/ISR
2. **Godot 4.7** — Engine open-source para games
3. **Hono** — Web framework ultrarrápido para Edge
4. **FastAPI 0.120** — Framework Python assíncrono
5. **Rust Web 2025** — Desenvolvimento web com Rust
6. **OpenAI API 2025** — LLMs e AI assistants
7. **React 19** — Library de UI JavaScript
8. **Kubernetes 2025** — Orquestração de containers
9. **PostgreSQL 17** — Banco de dados relacional
10. **LLM Fine-tuning** — Treinamento customizado de modelos
11. **WebAssembly** — Compilação no browser
12. **Tailwind CSS 4** — Framework CSS utilitário
13. **Docker 2025** — Containerização
14. **RAG Systems** — Retrieval-Augmented Generation
15. **Prompt Engineering** — Otimização de prompts para LLMs

---

## Mecanismo de Pesquisa Detalhado

### Exemplo: Arquivo sobre "React 19 - Hooks Best Practices"

```python
# 1. Executar 10 web_searches em paralelo/sequencial
searches = [
    "React 19 hooks best practices",
    "React 19 custom hooks patterns",
    "React 19 useCallback optimization",
    "React 19 useMemo performance",
    "React 19 hooks state management",
    "React 19 hooks vs alternatives",
    "React 19 hooks testing strategies",
    "React 19 hooks ESLint rules",
    "React 19 hooks memory leaks",
    "React 19 hooks with TypeScript",
]

# 2. Para CADA busca, executar fetch_url() em top 3 resultados
for search in searches:
    results = web_search(search, num_results=10)
    for url in results[:3]:  # Top 3 URLs
        content = fetch_url(url)
        # Armazena em contexto de pesquisa
        research_context += f"[{search}] {url}\n{content}\n\n"

# 3. Compilar com LLM
summary = llm.invoke(f"""
Com base nessa pesquisa sobre React 19 Hooks:

{research_context}

Escreva um guia completo sobre React 19 Hooks Best Practices
com exemplos de código e padrões recomendados.
""")

# 4. Salvar arquivo
save_file("output/research/react-19/03_hooks.md", summary)
```

---

## Ferramentas Vectora Utilizadas

### `web_search(query, num_results=10)`

```python
results = web_search("Next.js 16 new features 2025", num_results=10)
# Retorna: [
#     {"title": "...", "url": "...", "snippet": "..."},
#     ...
# ]
```

### `fetch_url(url)`

```python
content = fetch_url("https://example.com/nextjs-guide")
# Retorna: Full HTML content de uma URL
# Vectora extrai texto automaticamente (remove noise)
```

### `embedding(text)`

```python
embeddings = embedding("React 19 features and benefits")
# Usado para: RAG sobre documentação já indexada
# Opcional: se o Vectora encontrar docs relacionados, injeta como contexto
```

### `file_edit(path, old_text, new_text)`

```python
# Vectora atualiza o arquivo conforme compila informações
file_edit(
    "output/research/react-19/03_hooks.md",
    "# Placeholder",
    "# React 19 Hooks: Best Practices\n\n## Introdução\n..."
)
```

---

## Tempo de Execução Estimado

Com execução **paralela** de `web_search()`:

```
150 arquivos × 10 buscas = 1,500 web_searches
Cada busca: ~2-3 segundos (DuckDuckGo)
Fetch URLs: ~30 URLs processadas = ~1 segundo cada = 30 segundos
LLM summary: ~10 segundos por arquivo

PER FILE (paralelo):
  - web_searches (paralelo): 3 segundos
  - fetch_urls (paralelo): 1 segundo
  - LLM summary: 10 segundos
  = ~14 segundos/arquivo

TOTAL:
  150 arquivos × 14 seg = 2,100 segundos = ~35 minutos

OTIMIZADO (com cache):
  - Segunda busca do mesmo tópico reutiliza cache
  - Fetch URLs em paralelo (até 5 simultâneos)
  = ~20 minutos
```

---

## Modo de Operação

### Opção 1: Modo Manual (Demonstração)

```bash
# Iniciar chat interativo
python src/run_chat.py

# Dentro do chat
user> Pesquise sobre Next.js 16 e crie 10 arquivos
      - Um arquivo por tópico
      - Cada um com 10 web searches
      - Compile em sumários

Vectora> Iniciando pesquisa sobre Next.js 16...
         ✅ web_search("Next.js 16 features") → 10 URLs
         ✅ fetch_url(3 melhores) → conteúdo extraído
         ...
         ✅ Arquivo 1 criado: output/research/next-js-16/01_intro.md
```

### Opção 2: Modo Automático (Script Agente)

```bash
# Executar research agent automaticamente
python vectora_research_agent.py

# Processa 15 tópicos × 10 arquivos = 150 arquivos
# Salva em: output/research/{topic-slug}/
```

### Opção 3: MCP Server (Integração com Claude Code)

```bash
# Iniciar Vectora como servidor MCP
vectora-mcp

# No Claude Code, conecte e invoque:
# > Execute research agent: 150 arquivos sobre 15 tópicos
```

---

## Resultado Final

Após execução completa:

```
output/research/
├── next-js-16/
│   ├── 01_introducao.md (950 palavras)
│   ├── 02_historia_evolucao.md (1100 palavras)
│   ├── 03_recursos_principais.md (1200 palavras)
│   ├── 04_arquitetura.md (980 palavras)
│   ├── 05_instalacao_setup.md (850 palavras)
│   ├── 06_exemplos_praticos.md (1300 palavras)
│   ├── 07_performance.md (1050 palavras)
│   ├── 08_seguranca.md (900 palavras)
│   ├── 09_comparacao_alternativas.md (1100 palavras)
│   └── 10_roadmap_futuro.md (800 palavras)
│
├── godot-47/
│   ├── 01_introducao.md
│   ├── ...
│   └── 10_roadmap_futuro.md
│
├── hono-js/
├── fastapi-0120/
├── rust-web-2025/
├── openai-api-2025/
├── react-19/
├── kubernetes-2025/
├── postgres-17/
├── llm-fine-tuning/
├── webassembly-2025/
├── tailwind-css-4/
├── docker-container/
├── rag-systems-2025/
└── prompt-engineering/

ESTATÍSTICAS:
✅ 150 arquivos criados
✅ ~150,000 palavras de conteúdo pesquisado
✅ ~1,500 web searches executadas
✅ ~450 URLs acessadas e processadas
✅ Cada tópico tem 10 ângulos diferentes
✅ Todos com referências aos sites consultados
```

---

## Casos de Uso

### 1. **Documentação Técnica de Projetos**

```
Gere 150 arquivos sobre a arquitetura do nosso projeto:
- 10 arquivos sobre API REST
- 10 arquivos sobre frontend React
- 10 arquivos sobre backend Node.js
- ... etc
```

### 2. **Análise Competitiva**

```
Pesquise 5 concorrentes (30 arquivos):
- 10 arquivos: análise de features
- 10 arquivos: análise de pricing
- 10 arquivos: análise de marketing
```

### 3. **Currículo/Treinamento**

```
Crie 150 unidades de um curso online:
- 15 módulos (tópicos)
- 10 lições por módulo
- Cada lição com pesquisa profunda
```

### 4. **Market Research**

```
Analise 15 mercados diferentes:
- 10 arquivos por mercado
- Web search para dados atuais
- Sumários e insights compilados
```

---

## Limitações Atuais (MVP v0.1.0)

⚠️ **Implementação Parcial:**

- Context passing em LangGraph ainda em refinamento
- Ollama model binding não suporta `bind_tools()` completo
- Parallelização de web_search precisa de otimização
- Rate limiting em `fetch_url()` pode desacelerar

✅ **Soluções Planejadas (v0.2.0):**

- Auto-retry mais robusto em fetch_url()
- Connection pooling para múltiplas URLs
- Cache de web_search por query + date
- Batch processing de arquivos

---

## Scripts Fornecidos

### `vectora_research_agent.py` — Versão Full (15 tópicos)

- 150 arquivos completos
- Toda a infraestrutura
- ~35-40 minutos para completar

### `test_vectora_agent.py` — Versão Reduzida (2 tópicos)

- Apenas Next.js 16 e Godot 4.7
- Demonstração rápida
- ~2 minutos para completar

---

## Como Executar

```bash
# Versão demonstração (2 tópicos)
python test_vectora_agent.py

# Versão completa (15 tópicos × 10 arquivos)
python vectora_research_agent.py

# Dentro do chat interativo
python src/run_chat.py
# > Pesquise sobre os 15 tópicos e crie 150 arquivos
```

---

##️ Conclusão

O **Vectora Research Agent** demonstra o potencial do Vectora como um agente de pesquisa autônomo em larga escala. Combinando:

- **150 pesquisas web** profundas
- **Acesso automático** a múltiplos sites
- **Compilação inteligente** com LLM
- **Organização estruturada** em 150 arquivos

Isso abre possibilidades para **documentação automática**, **análise competitiva**, **market research** e muito mais.

---

**Status:** ✅ Conceito validado, implementação em andamento  
**Próximo:** Estabilizar context passing em LangGraph (v0.2.0)
