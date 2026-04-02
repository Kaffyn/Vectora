# 🌉 Bridge Layer: Adaptadores de Protocolo

A camada **Bridge** permite que as capacidades de **Memória Técnica Profunda** do Vectora sejam consumidas por ecossistemas externos de IA. Ela atua como um tradutor, expondo a inteligência do `src/core` em formatos industriais.

## 📂 Protocolos Suportados

### 1. `mcp/` (Model Context Protocol)
Implementação oficial do protocolo MCP para permitir que agentes de IA corporativos acessem o conhecimento vetorial local.
- **Como Funciona**: IDEs como **Cursor**, **VS Code (Copilot)** ou o agente **Claude Code** conectam-se a este servidor MCP.
- **Ações Expostas**:
    - `query_engine_docs`: Busca na documentação XML trancada da engine.
    - `query_project_context`: Busca no código-fonte local vetorizado.
    - `get_engine_methods`: Retorna assinaturas exatas de API para evitar alucinações.

### 2. `vscode/` (VS Code Integration)
Configurações e extensões auxiliares para melhorar a experiência de uso do Vectora dentro do editor mais popular de gamedev.
- Foca em fornecer snippets e integração visual com o Dashboard.

## 📜 Regras de Arquitetura (Adapter Pattern)
- **Zero Lógica no Bridge**: A ponte não deve conter lógica de busca ou processamento. Ela apenas chama os serviços correspondentes no `src/core/rag`.
- **Isolamento de Tipos**: Tipos de dados específicos de protocolos (como structs do MCP) devem ser mapeados de/para os tipos de domínio do Vectora (`src/core/domain`) para garantir que o core permaneça puro.
- **Opcionalidade**: O sistema deve ser 100% funcional sem a camada Bridge. Ela é um multiplicador de utilidade, não uma dependência.
