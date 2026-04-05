Este é um plano estratégico para o **Vectora Web**, desenhado como um produto SaaS independente, mas conectado ao ecossistema Vectora (mesmo Index, mesma lógica de roteamento).

A decisão de usar o **Open WebUI** como base inicial é extremamente inteligente: economiza meses de desenvolvimento de frontend, UX e fluxo de chat, permitindo que a equipe foque no **diferencial do Vectora**: o **Index Curado**, os **Workspaces Isolados** e a **Integração com LLMs**.

---

# 🚀 Plano de Desenvolvimento: Vectora Web (SaaS)

## 1. Visão Geral e Arquitetura

**Objetivo:** Criar uma interface web moderna, colaborativa e baseada em navegador para o Vectora, hospedada na infraestrutura da Kaffyn.
**Diferencial:** Acesso ao **Vectora Index** global, workspaces isolados via backend, e roteamento inteligente de LLMs, sem a necessidade de hardware local potente.

### Stack Tecnológica (Novo Projeto)

| Camada             | Tecnologia                            | Justificativa                                                                                                    |
| :----------------- | :------------------------------------ | :--------------------------------------------------------------------------------------------------------------- |
| **Frontend**       | **Open WebUI** (Fork/Customização)    | Base sólida, suporte a Ollama/LLMs, UI rica, modo "Chat", temas, histórico.                                      |
| **Backend**        | **Python (FastAPI)**                  | Compatibilidade nativa com Open WebUI, fácil integração com bibliotecas de RAG (`langchain`, `llama-index`).     |
| **Vector DB**      | **Bass API** (ou Milvus/Qdrant Cloud) | Escalável, multi-tenant, separado do banco local (`bbolt`), otimizado para busca vetorial massiva.               |
| **LLM Routing**    | **Custom Gateway (Go ou Python)**     | Gerencia roteamento entre Qwen (via API interna/K8s) e Gemini. Substitui `langchaingo` por um gateway escalável. |
| **Auth & Tenancy** | **PostgreSQL + Supabase/Auth0**       | Gestão de usuários, assinaturas, isolamento de dados por tenant.                                                 |
| **Storage**        | **S3 Compatible (MinIO ou AWS S3)**   | Armazenamento de uploads, datasets do Index, logs.                                                               |
| **Infra**          | **Kubernetes / Docker Compose**       | Orquestração na VPS da Kaffyn.                                                                                   |

> **Nota sobre o Fork:** Não usaremos o Open WebUI "como está". Faremos um **fork limpo**, removendo dependências de Ollama locais e substituindo-as pelo nosso **Gateway de Roteamento de LLMs**.

---

## 2. Estratégia de Migração e Customização

O objetivo não é reinventar a roda, mas adaptar o Open WebUI para ser o "motor" do Vectora Web.

### Passo 1: Limpeza do Open WebUI

- Remover módulos específicos de Ollama local.
- Desacoplar a camada de conexão de modelos.
- Padronizar a estrutura de dados para suportar **Workspaces Isolados** (namespace tagging).

### Passo 2: Implementação do Core Vectora

- **Roteador de Modelos:** Substituir a chamada direta para Ollama por um serviço que decide se usa Qwen (local/cloud) ou Gemini (API).
- **Motor de RAG Híbrido:** Integrar o `chromem-go` (via gRPC) ou reescrever a busca vetorial usando o **Bass API** para buscar no Index Global + Workspaces do Usuário.
- **Sistema de Workspaces:** Adaptar o conceito de "Collections" do Open WebUI para **Workspaces Vetoriais Isolados** (como definido no Vectora App).

### Passo 3: Integração com o Vectora Index

- Criar endpoint `/api/index/browse` que consome a API pública do Index.
- Permitir que o usuário "baixe" um dataset do Index diretamente para seu workspace na nuvem (com criptografia).

---

## 3. Funcionalidades Específicas (Diferenciais)

| Feature                      | Implementação no Vectora Web                                                                                                                                                          |
| :--------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Vectora Index**            | Catálogo integrado. Busca por tags (ex: "Godot", "Physics"). Download instantâneo de datasets `.lance` para o workspace ativo.                                                        |
| **Workspaces Isolados**      | Cada workspace tem namespace único. Contexto não vaza entre projetos. Configuração de quais workspaces estão ativos na sessão.                                                        |
| **Modo Multimodal (Gemini)** | Upload de PDFs, Imagens, Áudio → Processamento via Gemini Multimodal → Embedding no Bass API.                                                                                         |
| **Modo Code (Limitado)**     | **Não terá acesso ao FS local.**<br>- Foco em análise de código colado.<br>- Visualização de diffs de arquivos enviados.<br>- Integração via MCP (apenas leitura de contexto remoto). |
| **Colaboração**              | Compartilhamento de workspaces (leitura/escrita) entre membros da equipe (RBAC).                                                                                                      |
| **Privacidade**              | Dados criptografados em repouso (AES-256). Chaves geradas no cliente (opcional para planos Enterprise).                                                                               |

---

## 4. Roadmap de Desenvolvimento (8 Semanas)

### Fase 1: Fundação e Fork (Semanas 1-2)

- [ ] **Fork do Open WebUI**: Criar repositório `vectora-web`.
- [ ] **Configurar Backend**: Setup de FastAPI + PostgreSQL + Bass API.
- [ ] **Remover Dependências Locais**: Garantir que nada tente rodar LLMs localmente.
- [ ] **Autenticação**: Implementar login (Email/SSO) e gestão de tenants.

### Fase 2: Motor de RAG e Index (Semanas 3-4)

- [ ] **Integração Bass API**: Migrar a busca vetorial do SQLite local para o Bass API.
- [ ] **API do Index**: Consumir o catálogo público do Vectora Index.
- [ ] **Upload de Datasets**: Permitir upload de arquivos e processamento automático (embedding).
- [ ] **Isolamento de Contexto**: Garantir que queries só busquem nos workspaces selecionados.

### Fase 3: Roteamento de LLMs (Semanas 5-6)

- [ ] **Gateway Qwen/Gemini**: Implementar lógica de roteamento (Qwen via API interna, Gemini via chave do usuário).
- [ ] **Streaming de Respostas**: Adaptar o streaming do Open WebUI para o novo gateway.
- [ ] **Suporte a Multimodal**: Habilitar upload de imagens/PDFs para o Gemini.
- [ ] **Testes de Carga**: Validar performance com múltiplos workspaces simultâneos.

### Fase 4: Polimento e Lançamento (Semanas 7-8)

- [ ] **UI Branding**: Aplicar identidade visual da Kaffyn (cores, logo, tema dark).
- [ ] **Documentação**: Guia de uso, FAQ, preços.
- [ ] **Beta Fechado**: Teste com grupo seleto de usuários.
- [ ] **Lançamento Público**: Release do Vectora Web.

---

## 5. Comparativo Técnico: Vectora Web vs. Vectora App

| Característica        | Vectora App (Desktop/Fyne)           | Vectora Web (SaaS/OpenWebUI)                               |
| :-------------------- | :----------------------------------- | :--------------------------------------------------------- |
| **Stack**             | Go + Fyne + Chromem-go + BBolt       | Python (FastAPI) + OpenWebUI + Bass API + PG               |
| **Hospedagem**        | Local (User's Machine)               | VPS da Kaffyn (Cloud)                                      |
| **Acesso a Arquivos** | ✅ Full (FS, Shell, Git)             | ❌ Limitado (Upload/Download apenas)                       |
| **Modo Agente (ACP)** | ✅ Sim (Undo, Edit, Run)             | ⚠️ Parcial (Leitura de contexto, execução remota simulada) |
| **Performance**       | Depende do Hardware Local            | Depende da Infraestrutura da Kaffyn (Escalável)            |
| **Custo**             | Gratuito (Open Source)               | Freemium / Assinatura (SaaS)                               |
| **Foco**              | Privacidade Total, Offline, Controle | Colaboração, Acessibilidade, Potência de GPU               |

---

## 6. Próximos Passos Imediatos

1.  **Criar Repositório `vectora-web`**: Fork oficial do `open-webui/open-webui`.
2.  **Definir Schema de Banco de Dados**: Como mapear `Workspace`, `Dataset`, `Query` no PostgreSQL/Bass.
3.  **Configurar Ambiente de Dev**: Docker Compose para subir o Open WebUI modificado + Bass API.
4.  **Esboçar a API de Roteamento**: Especificar como o Frontend chamará o backend para escolher entre Qwen e Gemini.

Este plano permite que você lance o **Vectora Web** rapidamente, aproveitando o trabalho da comunidade Open WebUI, enquanto entrega o valor exclusivo da Kaffyn: **o Index Curado e a arquitetura de Workspaces**.
