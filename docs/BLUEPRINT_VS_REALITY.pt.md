# Blueprint vs. Realidade: O Estado do Vectora

**Última Atualização:** Abril de 2026  
**Fase Atual:** Pós-Implementação de Baseline (Estabilizada)

Este documento compara a visão original do projeto com o que foi efetivamente construído, servindo como uma auditoria técnica de fidelidade arquitetural.

---

## 1. Fidelidade da Arquitetura Core

| Componente    | Visão Inicial (Blueprint) | Realidade (Implementação) | Nota                                                                       |
| :------------ | :------------------------ | :------------------------ | :------------------------------------------------------------------------- |
| **Linguagem** | Go Nativo para o Core     | ✅ Confirmado             | Utilizado `Go 1.22+` em toda a infraestrutura.                             |
| **Protocolo** | MCP/JSON-RPC              | ✅ Confirmado             | Implementado **ACP (Agent Client Protocol)** via JSON-RPC 2.0 sobre Stdio. |
| **Segurança** | Guardian Engine           | ✅ Confirmado             | Motor de interceptação `core/policies` funcional e imutável.               |
| **Database**  | BBolt + VectorDB          | ✅ Confirmado             | Integração de `BBolt` com `Chromem-go` para RAG local.                     |

---

## 2. A Evolução do Protocolo (ACP)

Originalmente, planejamos usar o MCP (Model Context Protocol) como única via. Na realidade, percebemos que o MCP é excelente para ferramentas de terceiros, mas limitado para a integração profunda de UX que uma IDE exige (como streaming de tokens e status detalhado de sub-agentes).

**A Solução Realizada:** Criamos o **ACP**, uma super-estrutura do MCP que fornece as notificações e o gerenciamento de sessões necessários para a extensão VS Code, mantendo a compatibilidade de ferramentas com o padrão MCP.

---

## 3. Extensões e Interfaces

A visão de "Múltiplas Extensões" tornou-se realidade através de um modelo de **Binário Unificado**:

- O mesmo binário `vectora.exe` atua como daemon, CLI e servidor de protocolo.
- A extensão VS Code foi implementada usando um **Cliente Unificado** em TypeScript que simplifica drasticamente a manutenção.

---

## 4. O Sub-Agente: De Ferramenta a Raciocínio

No blueprint original, o sub-agente era descrito apenas como uma ferramenta utilitária. Na implementação real, ele evoluiu para uma camada de raciocínio recursiva capaz de executar loops de correção de bugs de forma autônoma, elevando o Vectora de um assistente de chat para um assistente de engenharia.

---

## 5. Próximos Passos Decididos

Com a baseline concluída com sucesso e alinhada ao blueprint em mais de 90%, o foco agora é a **Escalabilidade Agêntica**:

- Suporte a múltiplos sub-agentes paralelos.
- Integração profunda com ferramentas de build e CI locais.
- Expansão para TUI (Terminal User Interface) robusta.
