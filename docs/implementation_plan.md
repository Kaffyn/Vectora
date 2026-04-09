# Blueprint: Arquitetura Vectora Core (v1.5)

**Status:** Implementado (Baseline Consolidada)  
**Descrição:** Este documento define a arquitetura oficial do Vectora Core, servindo como a "Single Source of Truth" para o funcionamento do motor agêntico RAG.

---

## 1. Visão Geral do Sistema

O Vectora é um motor local de produtividade agêntica que unifica busca semântica (RAG), execução de ferramentas e interfaces de IDE através de um núcleo comum escrito em Go.

### Componentes Chave:

- **Core Daemon:** Processo de fundo que gerencia conexões, persistência e execução de ferramentas.
- **ACP (Agent Client Protocol):** Protocolo proprietário baseado em JSON-RPC 2.0 para comunicação agêntica de alta performance.
- **Guardian Engine:** Camada de segurança imutável que protege o sistema de arquivos e segredos do usuário.

---

## 2. Status dos Módulos Integrados

| Módulo                   | Blueprint                                | Status          | Descrição                                             |
| :----------------------- | :--------------------------------------- | :-------------- | :---------------------------------------------------- |
| **Proteção & Escopo**    | [POLICIES.md](POLICIES.md)               | ✅ Implementado | Motor Guardian com Trust Folder Enforcement.          |
| **Persistência**         | [Storage.md](Storage.md)                 | ✅ Implementado | BBolt para metadados e Chromem-go para vetores.       |
| **Executores agênticos** | [Tools_Executors.md](Tools_Executors.md) | ✅ Implementado | Suite de ferramentas Go nativas com segurança nativa. |
| **Gateway LLM**          | [LLM_Gateway.md](LLM_Gateway.md)         | ✅ Implementado | Suporte a Gemini 1.5/3.1 e Claude 3.5.                |
| **Comunicação (ACP)**    | [API.md](API.md)                         | ✅ Implementado | Servidor JSON-RPC 2.0 sobre Stdio/IPC.                |
| **Configuração**         | [Config_Manager.md](Config_Manager.md)   | ✅ Implementado | Gerenciamento de segredos e workspaces isolados.      |

---

## 3. Dinâmica Agêntica (ACP vs MCP)

O Vectora opera como uma ponte inteligente entre diferentes padrões de inteligência artificial:

1.  **Vectora como Servidor ACP:** Fornece capacidades agênticas avançadas para a extensão VS Code e CLI, incluindo gestão de sessões e streaming de planos de raciocínio.
2.  **Vectora como Cliente MCP:** O sistema é capaz de se conectar a servidores externos que sigam o Model Context Protocol (MCP), expandindo seu toolkit sob demanda.

---

## 4. Segurança Sistêmica (O Modelo Guardian)

A segurança no Vectora não é opcional e não depende de System Prompts. Ela é implementada via **Interceptação de Chamadas**:

- **Scope Checking:** Antes de qualquer operação de arquivo, o Core valida se o path absoluto pertence ao `TrustFolder`.
- **Sensitive Masking:** O `Guardian` escaneia outputs de ferramentas em busca de API Keys ou senhas conhecidas antes de enviá-los ao LLM.

---

## 5. Próximas Evoluções (Roadmap v2.0)

- **Local Inference:** Integração com Ollama/Llama.cpp para RAG 100% offline.
- **Visual Context (Multimodal):** Suporte para análise de screenshots da IDE e diagramas via Gemini Pro Vision.
- **Multi-Sub-Agents:** Orquestração paralela de múltiplos sub-agentes para refatorações de larga escala.
