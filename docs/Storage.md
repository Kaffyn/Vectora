# Blueprint: Arquitetura de Persistência Dual-Store

**Status:** Implementado
**Módulo:** `core/storage/`
**Tecnologias:** BBolt (Key-Value), Chromem-go (Vector Store)

O sistema de armazenamento do Vectora foi projetado para ser 100% local, autocontido e otimizado para a performance de pesquisa semântica em máquinas de desenvolvimento.

---

## 1. Estratégia Dual-Store

Diferente de sistemas cloud que usam um único banco de dados gigantesco, o Vectora separa os dados em duas camadas lógicas e físicas:

### A. Meta-Store (BBolt)

- **O que armazena:** Sessões de chat, configurações de workspace, memória agêntica (fatos aprendidos) e estados de indexação.
- **Por que BBolt?** É um banco KV nativo do Go, extremante estável, que utiliza um único arquivo no disco e oferece transações ACID.
- **Localização:** `%LOCALAPPDATA%/Vectora/data/meta.db`

### B. Vector-Index (Chromem-go)

- **O que armazena:** Embeddings de código e documentação para busca semântica (RAG).
- **Integridade:** Utiliza o motor local `chromem-go` para gerenciar as coleções de vetores. A persistência é feita via arquivos `.json` e `.bin` otimizados para leitura sequencial rápida.
- **Localização:** `%LOCALAPPDATA%/Vectora/vectors/<workspace-hash>/`

---

## 2. Ingestão e Chunking

Para maximizar a precisão da busca RAG, o Vectora aplica uma lógica de fragmentação (chunking) inteligente:

1.  **Parser Aware:** Tenta identificar funções e classes para não quebrar blocos lógicos de código.
2.  **Overlap Estratégico:** Aplica um overlap de 10-15% entre chunks para garantir que o contexto semântico não se perca nas bordas dos fragmentos.
3.  **Filtragem Hierárquica:** O motor de indexação respeita as regras do `Guardian` e arquivos `.gitignore`, nunca transformando dados sensíveis em vetores.

---

## 3. Isolamento de Workspace

A segurança e a relevância são mantidas através do isolamento físico por workspace:

- Cada projeto aberto gera um hash único baseado no seu path absoluto.
- O Vectora carrega apenas os índices correspondentes ao workspace da sessão ativa, evitando "contaminação" de contexto entre projetos diferentes (ex: código de trabalho vs projeto pessoal).

---

## 4. Próximas Implementações

- **Auto-Compaction:** Rotinas de fundo para desfragmentar o BBolt e otimizar o índice vetorial.
- **Encryption-at-Rest:** Opção para criptografar o banco de metadados com uma chave mestre do usuário.
