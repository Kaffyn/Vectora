# Arquitetura MCP do Vectora - Clarificação

## ❌ ERRADO: Tools Expostas Diretamente

```
Claude Code
    ├─ chama web_search()
    ├─ chama vector_search()
    ├─ chama file_read()
    └─ chama embedding()
```

**Problema:** Claude Code seria responsável por orquestrar as ferramentas. Não há raciocínio no Vectora.

---

## ✅ CORRETO: Comunicação Agent-to-Agent

```
┌─────────────────────────────────────────────────────────────┐
│                        Claude Code                          │
│              (Host Agent / Agente Principal)                │
└──────────────────────────┬──────────────────────────────────┘
                           │
                    (lê Resources)
                           │
          ┌────────────────┴────────────────┐
          ↓                                  ↓
   Context Resource                 History Resource
   {                               {
     "status": "active",             "messages": [...]
     "summary": "..."               }
   }
          │                                  │
          └────────────┬─────────────────────┘
                       │
            (comunica com Vectora via MCP)
                       │
                       ↓
┌──────────────────────────────────────────────────────────────┐
│                    Vectora MCP Server                        │
│           (Sub-Agente / Agente Colaborativo)                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Comunica: "Preciso saber sobre X"                         │
│                     ↓                                        │
│         LangGraph (Raciocínio Interno)                      │
│         ├─ MAIN_NODE: interpreta requisição                │
│         ├─ TOOL_NODE: orquestra tools                       │
│         │   ├─ web_search()         (interno)              │
│         │   ├─ vector_search()      (interno)              │
│         │   ├─ file_read()          (interno)              │
│         │   ├─ grep()               (interno)              │
│         │   ├─ terminal()           (interno)              │
│         │   ├─ embedding()          (interno)              │
│         │   ├─ list_dir()           (interno)              │
│         │   ├─ file_edit()          (interno)              │
│         │   ├─ call_mcp_tool()      (interno)              │
│         │   ├─ ingest_docs()        (interno)              │
│         │   └─ fetch_url()          (interno)              │
│         ├─ SUMMARIZER_NODE: resume contexto                │
│         └─ SUB_NODE: workflows complexos                   │
│                     ↓                                        │
│  Responde: "Baseado em pesquisa + conhecimento local..."  │
│                                                              │
└──────────────────────────┬───────────────────────────────────┘
                           │
                    (Status Resource)
                           │
                           ↓
                  { "status": "ready",
                    "tools": 11,
                    "capabilities": [...] }
```

---

## 📡 O Que é Exposto via MCP?

### ✅ Recursos (Resources)

```
GET vectora://thread/{id}/context
    → Estado cognitivo: "Já coletei X docs sobre Y"

GET vectora://thread/{id}/history
    → Últimas 5 mensagens da conversa

GET vectora://status
    → Status do servidor (LLM, RAG, uptime)
```

### ❌ Ferramentas NÃO são expostas

```
- web_search()      ← Interno (não exposto)
- vector_search()   ← Interno (não exposto)
- file_read()       ← Interno (não exposto)
- terminal()        ← Interno (não exposto)
- ... (11 tools)    ← Todos internos
```

---

## 🔄 Fluxo de Decisão Correto

### Fase 1: Claude Code Lê Estado

```
Claude Code: "Qual é o contexto do Vectora?"
↓
GET vectora://thread/123/context
↓
Resposta: "Vectora sabe sobre: RAG, vector_search, LanceDB"
```

### Fase 2: Claude Code Decide

```
Pensamento: "Usuário perguntou sobre Redis.
Vectora sabe sobre vector search,
mas não sabe sobre cache distribuído.
Devo comunicar requisição ao Vectora."
```

### Fase 3: Claude Code Comunica COM Vectora

```
Claude Code: "Vectora, pesquise e contextualize:
como usar Redis com vector search?"
↓
(Enviado via MCP JSON-RPC)
```

### Fase 4: Vectora Processa Internamente

```
Vectora recebe requisição
↓
LangGraph executa:
  1. MAIN_NODE analisa: "pesquisa sobre Redis + vector_search"
  2. TOOL_NODE orquestra internamente:
     - web_search("Redis vector search 2026")
     - vector_search("cache patterns") em seu LanceDB
  3. SUMMARIZER_NODE resume achados
  4. Retorna ao Claude Code
```

### Fase 5: Claude Code Sintetiza

```
Claude Code recebe resposta processada
↓
Integra com seu próprio conhecimento
↓
Envia resposta final ao usuário
```

---

## 🎯 Pontos-Chave

| Aspecto                  | Errado ❌          | Correto ✅                  |
| ------------------------ | ------------------ | --------------------------- |
| O que Claude Code chama? | As 11 tools direto | O Vectora (via MCP)         |
| Onde as tools executam?  | Claude Code        | Dentro do Vectora           |
| O que é exposto via MCP? | 11 tools           | 3 Resources                 |
| Protocolo                | HTTP/REST          | JSON-RPC via stdio          |
| Raciocínio               | Em Claude Code     | Em ambos (colaborativo)     |
| Estado                   | Não compartilhado  | Compartilhado via Resources |

---

## 💬 Analogia

**Errado:**

```
Você (Claude Code) diretamente:
- Abre o Google → web_search()
- Abre o LanceDB → vector_search()
- Lê arquivos → file_read()
- Executa comandos → terminal()
```

**Correto:**

```
Você (Claude Code) trabalha com Vectora:

"Vectora, eu preciso entender como
funciona Redis com vector search.
Você tem pesquisa ou conhecimento
local sobre isso?"

Vectora (internamente):
- Pesquisa na web
- Busca em seu banco de conhecimento
- Executa comandos para testar
- Resume e responde

"Aqui está o que encontrei..."
```

---

## 📝 Resumo para Documentação

**O MCP do Vectora expõe:**

- ✅ **3 Resources** (context, history, status)
- ✅ **Comunicação JSON-RPC** (Agent-to-Agent)
- ✅ **Raciocínio LangGraph** (internamente)

**O MCP do Vectora NÃO expõe:**

- ❌ As 11 ferramentas (são internas)
- ❌ Interface HTTP (usa stdio)
- ❌ Chamadas diretas a tools

**Claude Code faz:**

- Lê o estado do Vectora (Resources)
- Comunica requisições de alto nível
- Recebe respostas processadas

**Vectora faz:**

- Processa requisições internamente
- Orquestra as 11 ferramentas
- Executa LangGraph com raciocínio próprio
- Retorna resultado sintetizado
