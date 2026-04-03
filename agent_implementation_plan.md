# Plano de Implementação: Agência, ACP e MCP (Model Context Protocol)

Este plano descreve o cérebro agêntico do Vectora, permitindo autonomia no sistema de arquivos e integração com ferramentas externas.

---

## 1. Definição de Agência no Vectora

O Vectora não é apenas um chat; é um executor. Ele opera em dois modos agênticos:

1. **ACP (Autonomous Control Protocol):** Loop interno onde o LLM usa o toolkit em `internal/tools`.
2. **MCP (Model Context Protocol):** O Vectora atua como um servidor para IDEs (Cursor, VSCode), provendo contexto e ferramentas para outros IAs.

---

## 2. O Loop de Raciocínio (ReAct) no `internal/acp`

O Agente ACP utiliza o padrão **Thought -> Action -> Observation**:

1. **Pensamento:** O LLM analisa o objetivo (ex: "Corrija o bug no main.go").
2. **Ação:** Escolhe uma tool (ex: `find_files`).
3. **Execução:** O Daemon valida a segurança (GitBridge) e executa a tool.
4. **Observação:** O resultado volta para o LLM para o próximo ciclo.

---

## 3. Servidor MCP (`internal/mcp/server.go`)

O Vectora implementa a especificação oficial do **Model Context Protocol** da Anthropic:

- **Resources:** Expor coleções do banco vetorial como recursos de texto.
- **Tools:** Expor as mesmas tools do Agente (Read/Write) via MCP.
- **Prompts:** Templates de prompt pré-definidos para auditoria de código.

---

## 4. Toolkit de Segurança (Governança Agêntica)

Nenhuma ação agêntica é "solta" no sistema.

### 4.1 Confirmação de Tool-Use (Human-in-the-loop)

Toda execução de tool reportada via IPC `tool.execute` pode requerer aprovação explícita no Web UI se a flag de segurança estiver ativa.

### 4.2 O Papel do GitBridge

O Agente ACP deve chamar obrigatoriamente a `internal/git.Bridge.Snapshot()` antes de cada `Action` que envolva escrita, garantindo que o usuário possa dar `undo` em qualquer alteração feita pelo IA.

---

## 5. Regras de Negócio (Agente)

- **RN-AG-01:** O Agente nunca deve executar `run_shell_command` com privilégios de root sem aviso visual persistente.
- **RN-AG-02:** O histórico de pensamentos (Thought stream) deve ser persistido no bbolt para auditoria futura.
- **RN-AG-03:** O servidor MCP deve ser acessível via `stdio` e, opcionalmente, via `SSE` (Server-Sent Events) para conexões locais de rede.

---

## 6. Próximos Passos (Agente)

1.  [ ] **Implementar o Servidor MCP:** Mapear os comandos IPC para as structs exigidas pelo protocolo MCP.
2.  [ ] **Interface de Pensamento:** No Web UI (`internal/app`), criar uma aba para visualizar o "raciocínio" do agente durante execuções longas.

[Fim do Plano do Agente - Revisão 2026.04.03]
