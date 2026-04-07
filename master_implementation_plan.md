# Vectora: Plano Mestre de Implementação (Master Plan)

Este documento coordena a expansão do ecossistema Vectora, integrando a inteligência vetorial remota com a interface desktop nativa. O desenvolvimento é descentralizado em planos específicos para garantir densidade técnica e clareza arquitetural.

## Visão Geral do Sistema

O Vectora é composto por três pilares principais operando em harmonia:
1.  **Daemon Local**: Inteligência local e orquestração de modelos.
2.  **Index Service**: Cérebro vetorial persistente na nuvem (Supabase + pgvector).
3.  **Desktop App (Fyne)**: Interface de controle centralizada com foco em UX premium.

---

## Planos Descentralizados

Para detalhes técnicos profundos, consulte os documentos de implementação específicos:

### 1. [📚 Index Service Plan](file:///C:/Users/bruno/.gemini/antigravity/brain/c916d7f5-cb3a-4952-884a-a7adea46bbaa/index_service_plan.md)
**Foco**: Implementação do backend de RAG (Retrieval-Augmented Generation).
*   Provedores de Embedding (Gemini text-embedding-004).
*   Lógica de processamento e chunking de documentos.
*   Busca vetorial via gRPC no Supabase.
*   Infraestrutura resiliente.

### 2. [🖥️ Desktop App Plan (Fase 2)](file:///C:/Users/bruno/.gemini/antigravity/brain/c916d7f5-cb3a-4952-884a-a7adea46bbaa/desktop_app_plan.md)
**Foco**: Saída do estágio de "Mocks" para integração total de produção.
*   Implementação real dos stubs gRPC no `IndexClient`.
*   Monitoramento de uploads resilientes via navegador.
*   UI dinâmica para gerenciamento de índices remotos.
*   Persistência de histórico de chat local.

---

## Roadmap de Sincronização

| Fase | Objetivo | Dependência |
| :--- | :--- | :--- |
| **Fase A** | Geração de Protobufs Unificada | Makefile compartilhado |
| **Fase B** | Backend Funcional (Search/Upload) | Supabase + Gemini API |
| **Fase C** | Integração Desktop-Remoto | Index Service URL nas Prefs |
| **Fase D** | Validação E2E (Publicar -> Buscar) | Conexão gRPC estável |

## User Review Required

> [!IMPORTANT]
> **Estratégia de Deploy**: Como discutido anteriormente, a Vercel possui limitações para gRPC TCP. A execução dos planos abaixo assumirá que o **Index Service** será hospedado em uma plataforma que suporte processos gRPC de longa duração (ex: Railway), ou que implementaremos o bridge gRPC-web se a Vercel for mandatória.

Aguardando aprovação do Plano Mestre para detalhar os planos específicos.
