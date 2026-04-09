# Blueprint: Pipeline de Ingestão de Dados

**Status:** Implementado  
**Módulo:** `core/ingestion/`  
**Características:** Indexação On-Demand, Semantic Aware

O motor de ingestão do Vectora é responsável por transformar código fonte bruto e documentação em vetores semânticos pesquisáveis. Diferente de busca por texto simples, a ingestão permite que o agente entenda a _intenção_ por trás do código.

---

## 1. Fluxo de Ingestão (Pipeline)

A ingestão ocorre em quatro fases coordenadas pelo `Engine`:

1.  **Discovery:** O motor escaneia o `TrustFolder`, respeitando o `.gitignore` e as extensões bloqueadas pelo `Guardian`.
2.  **Parsing & Chunking:**
    - Arquivos de código são quebrados em fragmentos (chunks) baseados no tamanho do contexto (Tokens).
    - Utiliza-se uma estratégia de segmentação por funções/classes sempre que possível para manter a unidade lógica.
3.  **Embedding Generation:**
    - Os fragmentos são enviados ao `LLM Gateway` (ex: Gemini Embedding 004).
    - Gera um vetor de alta dimensão que representa o significado do texto.
4.  **Vector Upsert:**
    - Os vetores e metadados (path, linha inicial) são salvos no `Chromem-go`.

---

## 2. Estratégia "On-Demand" vs "Full Scan"

Para evitar consumo excessivo de CPU e bateria em máquinas de desenvolvedores:

- **Initial Scan:** Realizado na primeira vez que um workspace é aberto.
- **Incremental Sync:** O Core monitora mudanças nos arquivos (File System Events) e re-indexa apenas os arquivos modificados quase em tempo real.
- **Lazy Indexing:** Se um arquivo nunca foi acessado ou questionado, o sistema pode optar por indexá-lo apenas quando uma busca semântica falhar em encontrar resultados relevantes nos arquivos ativos.

---

## 3. Gestão de Contexto e Relevância

A ingestão não indexa tudo cegamente. Existe uma lógica de prioridade:

- **Prioridade Alta:** Arquivos de configuração, `main.go`, `package.json`, `README.md`.
- **Prioridade Média:** Arquivos de lógica de negócio e testes.
- **Prioridade Baixa:** Arquivos de utilitários e documentação redundante.

---

## 4. Próximas Implementações

- **AST-Based Chunking:** Uso de Parsers reais (Tree-sitter) para segmentar o código com precisão sintática absoluta.
- **Image Ingestion:** Indexação de diagramas de arquitetura e capturas de tela para suporte multimodal.
